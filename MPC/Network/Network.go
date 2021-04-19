package Network

import (
	bundle "MPC/Bundle"
	numberbundle "MPC/Bundle/Number-bundle"
	"bufio"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type Receiver interface {
	Receive(bundle bundle.Bundle)
}

type Packet struct {
	ID string
	Type string
	Connections []string
	Bundle bundle.Bundle
}

//List of IPs
var peers []string
//List of connections
var connections []net.Conn

var connMutex = &sync.Mutex{}
var readyMutex = &sync.Mutex{}
var peersMutex = &sync.Mutex{}
var partiesMutex = &sync.Mutex{}

var encoders = make(map[net.Conn]*gob.Encoder)
var decoders = make(map[net.Conn]*gob.Decoder)
var parties = make(map[string]net.Conn)
var finalNetworkSize int
var readyParties = make(map[net.Conn]bool)
var readySent = false
var isHost bool
var receiver []Receiver
var myIP string

var debug = true


func GetPartyNumber() int {
	for i, p := range peers {
		if p == myIP {
			return i + 1
		}
	}
	panic("Could not find miself :(")
}

func Init(networkSize int) bool {

	finalNetworkSize = networkSize

	gob.Register(numberbundle.NumberBundle{})

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Ip and port of a peer on the network >")
	ipPort, _ := reader.ReadString('\n')
	ipPort = strings.TrimSpace(ipPort)


	ln, err := net.Listen("tcp", ":40404")
	if debug {
		ln, err = net.Listen("tcp", ":")
	}
	_, port, _ := net.SplitHostPort(ln.Addr().String())

	myIP = getPublicIP() + ":" + port

	if debug {
		myIP = getLocalIP() + ":" + port
	}

	fmt.Println("Listening on following connection: ", myIP)
	peersMutex.Lock()
	peers = append(peers, myIP)
	peersMutex.Unlock()

	if err != nil {
		fmt.Println("Could not listen for incoming connections:", err.Error())
		panic(err.Error())
	}

	//Connect returnere false hvis man failer et connect - hermed er du den første
	isHost = !connect(ipPort)

	go listen(ln)
	return isHost
}

func RegisterReceiver(r Receiver) {
	receiver = append(receiver, r)
}

func GetParties() int {
	connMutex.Lock()
	connLen := len(connections)
	connMutex.Unlock()
	return connLen + 1
}

func IsReady() bool {
	readyMutex.Lock()
	readyLen := len(readyParties)
	readyMutex.Unlock()
	return readyLen + 1 == finalNetworkSize
}

func sendReady() {
	readyMutex.Lock()
	defer readyMutex.Unlock()
	if readySent {
		return
	}

	packet := Packet{
		ID: uuid.Must(uuid.NewRandom()).String(),
		Type: "ready",
	}

	if isHost {
		peersMutex.Lock()
		packet.Connections = peers
		peersMutex.Unlock()
	}

	fmt.Println(packet)

	connMutex.Lock()
	for _, conn := range connections {
		encoder := encoders[conn]
		err := encoder.Encode(packet)

		if err != nil {
			fmt.Println("Failed to send ready", err.Error())
		}
	}
	connMutex.Unlock()

	fmt.Println("I am ready!")
	readySent = true
}

func Send(bundle bundle.Bundle, party int) {
	//TODO make party int consistent (-1?)
	peersMutex.Lock()
	peer := peers[party - 1 ]
	peersMutex.Unlock()
	partiesMutex.Lock()
	partyToSend, found := parties[peer]//connections[party]
	partiesMutex.Unlock()
	if !found {
		fmt.Println("Party could not be found in parties :(")
	}
	packet := Packet{
		ID: uuid.Must(uuid.NewRandom()).String(),
		Type: "bundle",
		Bundle: bundle,
	}

	encoder := encoders[partyToSend]
	//encoder := gob.NewEncoder(partyToSend)

	err := encoder.Encode(packet)

	if err != nil {
		fmt.Println("Failed to gob bundle:", err.Error())
	}
}

// Listen for incoming connections
func listen(ln net.Listener) {

	//Accept incoming connections
	for {
		conn, err := ln.Accept()

		if err != nil {
			fmt.Println("Failed to accept incoming connection:", err.Error())
			return
		}

		fmt.Println("Accepted connection from:", conn.RemoteAddr())

		go sendPeers(conn)
		go handleConnection(conn)

		connMutex.Lock()
		connections = append(connections, conn)
		connMutex.Unlock()

		peersMutex.Lock()
		peersLength := len(peers)
		peersMutex.Unlock()
		if peersLength == finalNetworkSize {
			sendReady()
		}

		//fmt.Println("I have the following connections:", peers)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := gob.NewDecoder(conn)
	decoders[conn] = decoder //Add decoders to map

	for {
		packet := Packet{}
		err := decoder.Decode(&packet)

		if err != nil {
			fmt.Println("Connection error:", err.Error())
			return
		}

		if packet.Type == "peerlist" {
			go getPeers(packet.Connections, conn)
		}

		if packet.Type == "bundle" {
			if len(receiver) == 0 {
				fmt.Println("No receiver registered")
				return
			}
			//fmt.Println("Sending packet to receivers ", packet.Bundle)
			for _, r := range receiver {
				r.Receive(packet.Bundle)
			}
			//receiver.Receive(packet.Bundle)
		}

		if packet.Type == "ready" {
			if len(packet.Connections) > 0 {
				peersMutex.Lock()
				peers = packet.Connections
				peersMutex.Unlock()
			}
			readyMutex.Lock()
			readyParties[conn] = true
			readyMutex.Unlock()
		}
	}
}

func getPeers(conns []string, sender net.Conn) {
	connMutex.Lock()
	senderIP := conns[0]
	connMutex.Unlock()
	partiesMutex.Lock()
	parties[senderIP] = sender
	partiesMutex.Unlock()
	for i, ip := range conns {
		if newIP(ip) {
			peersMutex.Lock()
			peers = append(peers, ip)
			peersMutex.Unlock()

			//Do not connect to connected peers own ip - we already have a connection
			if i != 0 {
				connect(ip)
			}
		}
	}
	connMutex.Lock()
	connLen := len(connections)
	connMutex.Unlock()
	if connLen + 1 == finalNetworkSize {
		sendReady()
	}

	//fmt.Println("Received peers. I now have the following connections:", peers)
}

//Check if we already have a connection to this ip or if it is our own ip
func newIP(ip string) bool {
	peersMutex.Lock()
	defer peersMutex.Unlock()
	for _, peer := range peers {
		if peer == ip {
			return false
		}
	}

	//Check if own ip
	if ip == peers[0] {
		return false
	}

	return true
}

func sendPeers(conn net.Conn) {
	peersMutex.Lock()
	peersList := peers
	peersMutex.Unlock()
	encoder := gob.NewEncoder(conn)
	encoders[conn] = encoder //Add encoder to map
	packet := Packet{
		ID: uuid.Must(uuid.NewRandom()).String(),
		Type: "peerlist",
		Connections: peersList,
	}
	err := encoder.Encode(packet)

	//fmt.Println("Sent peerlist to peer:", conn.RemoteAddr().String())

	if err != nil {
		fmt.Println("Failed to gob peer packet:", err.Error())
	}
}

func connect(ipPort string) bool {
	conn, err := net.Dial("tcp", ipPort)

	if err != nil {
		fmt.Println("Failed to connect to peer:", err.Error())
		return false
	}

	fmt.Println("Connected to peer", ipPort)
	sendPeers(conn) //Send ip I'm listening on to connected peer
	go handleConnection(conn)

	connMutex.Lock()
	connections = append(connections, conn)
	connMutex.Unlock()

	partiesMutex.Lock()
	parties[ipPort] = conn
	partiesMutex.Unlock()
	return true
}

// Inspired by https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()

	if err != nil {
		log.Fatal(err)
	}

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}


func getPublicIP() string {
	url := "https://api.ipify.org?format=text"	// we are using a public IP API, we're using ipify here, below are some others
	// https://www.ipify.org
	// http://myexternalip.com
	// http://api.ident.me
	// http://whatismyipaddress.com/api
	fmt.Printf("Getting IP address from  ipify ...\n")
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(ip)
}






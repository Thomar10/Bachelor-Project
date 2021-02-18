package Network

import (
	bundle "MPC/Bundle"
	Prime_bundle "MPC/Bundle/Prime-bundle"
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"net"
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

var receiver Receiver

func Init() bool {

	gob.Register(Prime_bundle.PrimeBundle{})

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Ip and port of a peer on the network >")
	ipPort, _ := reader.ReadString('\n')
	ipPort = strings.TrimSpace(ipPort)

	ln, err := net.Listen("tcp", ":")
	_, port, _ := net.SplitHostPort(ln.Addr().String())

	ownIP := getPublicIP() + ":" + port
	fmt.Println("Listening on following connection: ", ownIP)
	peers = append(peers, ownIP)

	if err != nil {
		fmt.Println("Could not listen for incoming connections:", err.Error())
		panic(err.Error())
	}

	connected := connect(ipPort)

	go listen(ln)
	return !connected
}

func RegisterReceiver(r Receiver) {
	receiver = r
}

func GetParties() int {
	return len(connections)
}

func Send(bundle bundle.Bundle, party int) {
	partyToSend := connections[party]

	packet := Packet{
		ID: uuid.Must(uuid.NewRandom()).String(),
		Type: "bundle",
		Bundle: bundle,
	}

	encoder := gob.NewEncoder(partyToSend)

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

		//fmt.Println("I have the following connections:", peers)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := gob.NewDecoder(conn)

	for {
		packet := Packet{}
		err := decoder.Decode(&packet)

		if err != nil {
			fmt.Println("Connection error:", err.Error())
			return
		}

		if packet.Type == "peerlist" {
			go getPeers(packet.Connections)
		}

		if packet.Type == "bundle" {
			if receiver == nil {
				fmt.Println("No receiver registered")
				return
			}

			receiver.Receive(packet.Bundle)
		}
	}
}

func getPeers(conns []string) {

	for i, ip := range conns {
		if newIP(ip) {
			peers = append(peers, ip)

			//Do not connect to connected peers own ip - we already have a connection
			if i != 0 {
				connect(ip)
			}
		}
	}

	//fmt.Println("Received peers. I now have the following connections:", peers)
}

//Check if we already have a connection to this ip or if it is our own ip
func newIP(ip string) bool {
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
	encoder := gob.NewEncoder(conn)

	packet := Packet{
		ID: uuid.Must(uuid.NewRandom()).String(),
		Type: "peerlist",
		Connections: peers,
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
	return true
}

// Inspired by https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go

func getPublicIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()

	if err != nil {
		log.Fatal(err)
	}

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

/*
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

 */




package Network

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type Packet struct {
	ID string
	Type string
	Connections []string
}

//List of IPs
var peers []string
//List of connections
var connections []net.Conn
var connMutex = &sync.Mutex{}

func Init() {

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Ip and port of a peer on the network >")
	ipPort, _ := reader.ReadString('\n')
	ipPort = strings.TrimSpace(ipPort)

	ln, err := net.Listen("tcp", ":40404")
	_, port, _ := net.SplitHostPort(ln.Addr().String())

	ownIP := getPublicIP() + ":" + port
	fmt.Println("Listening on following connection: ", ownIP)
	peers = append(peers, ownIP)

	if err != nil {
		fmt.Println("Could not listen for incoming connections:", err.Error())
		return
	}

	connect(ipPort)

	go listen(ln)
}

func Send(message string, party int) {

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

	for {
		packet := Packet{}
		decoder := gob.NewDecoder(conn)
		err := decoder.Decode(&packet)

		if err != nil {
			fmt.Println("Connection error:", err.Error())
			return
		}

		if packet.Type == "peerlist" {
			go getPeers(packet.Connections)
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

func connect(ipPort string) {
	conn, err := net.Dial("tcp", ipPort)

	if err != nil {
		fmt.Println("Failed to connect to peer:", err.Error())
		return
	}

	fmt.Println("Connected to peer", ipPort)
	sendPeers(conn) //Send ip I'm listening on to connected peer
	go handleConnection(conn)

	connMutex.Lock()
	connections = append(connections, conn)
	connMutex.Unlock()
}

// Inspired by https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
/*
func getPublicIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()

	if err != nil {
		log.Fatal(err)
	}

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
*/

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




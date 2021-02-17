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

//List of connections
var connections []net.Conn
var connMutex = &sync.Mutex{}

func Init() {

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Ip and port of a peer on the network >")
	ipPort, _ := reader.ReadString('\n')
	ipPort = strings.TrimSpace(ipPort)

	connect(ipPort)

	go listen()
}

func Send(message string, party int) {

}

// Listen for incoming connections
func listen() {
	ln, err := net.Listen("tcp", ":40404")
	_, port, _ := net.SplitHostPort(ln.Addr().String())

	ipPort := getPublicIP() + ":" + port
	fmt.Println("Listening on following connection: ", ipPort)

	if err != nil {
		fmt.Println("Could not listen for incoming connections:", err.Error())
		return
	}

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

		fmt.Println("I have the following connections:", connections)
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
	for _, ip := range conns {
		if newIP(ip) {
			connect(ip)
		}
	}
}

//TODO filter properly
//Check if we already have a connection to this ip or if it is our own ip
func newIP(ip string) bool {
	connMutex.Lock()
	for _, c := range connections {
		if c.RemoteAddr().String() == ip {
			return false
		}
	}
	connMutex.Unlock()
	if ip == getPublicIP() {
		return false
	}

	return true
}

func sendPeers(conn net.Conn) {
	encoder := gob.NewEncoder(conn)
	connMutex.Lock()

	var conns []string

	for _, c := range connections {
		if c == conn {
			continue
		}
		conns = append(conns, c.RemoteAddr().String())
	}

	packet := Packet{
		ID: uuid.Must(uuid.NewRandom()).String(),
		Type: "peerlist",
		Connections: conns,
	}
	connMutex.Unlock()
	err := encoder.Encode(packet)

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



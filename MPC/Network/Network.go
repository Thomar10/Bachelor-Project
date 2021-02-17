package Network

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
)

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

		connMutex.Lock()
		connections = append(connections, conn)
		connMutex.Unlock()
	}
}

func connect(ipPort string) {
	conn, err := net.Dial("tcp", ipPort)

	if err != nil {
		fmt.Println("Failed to connect to peer:", err.Error())
		return
	}

	fmt.Println("Connected to peer", ipPort)

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

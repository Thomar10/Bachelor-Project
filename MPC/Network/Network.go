package Network

import (
	"bufio"
	"fmt"
	"log"
	"net"
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
	ln, err := net.Listen("tcp", ":")
	_, port, _ := net.SplitHostPort(ln.Addr().String())

	ipPort := getOutboundIP() + ":" + port
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

	connMutex.Lock()
	connections = append(connections, conn)
	connMutex.Unlock()
}

// Inspired by https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()

	if err != nil {
		log.Fatal(err)
	}

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:42069")
	if err != nil {
		log.Printf("Failed to create a listener with error: %s\n", err)
		os.Exit(1)
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Printf("Failed to create a UDP connection with the address %s\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("> ")
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Failed to get user input: %s\n", err)
		}
		_, err = conn.Write([]byte(message))
		if err != nil {
			log.Printf("Failed to write message to UDP: %s\n", err)
		}

	}
}

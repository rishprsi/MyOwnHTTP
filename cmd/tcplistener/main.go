package main

import (
	"fmt"
	"log"
	"net"

	"MyOwnHTTP/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:42069")
	if err != nil {
		log.Printf("Failed to create a network listener with error: %s\n", err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept the incoming connection with the error: %s", err)
		}
		log.Println("Conenction has been accepted")
		request, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Printf("Failed to read message from network: %s\n", err)
		}
		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", request.RequestLine.Method)
		fmt.Printf("- Target: %s\n", request.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", request.RequestLine.HttpVersion)
		fmt.Printf("Headers:\n")
		for key, value := range request.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}
		fmt.Println("Body:")
		fmt.Println(string(request.Body))
		log.Println("The channel has been closed")
	}
}

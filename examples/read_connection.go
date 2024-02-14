package examples

import (
	"fmt"
	"io"
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Read error: %v", err)
			}
			break
		}

		data := buf[:n]
		fmt.Printf("Received: %s\n", string(data))
	}
}

func runReadConnectionExample() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Error on listening: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error on accept: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

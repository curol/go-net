package mock

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("New connection from " + conn.RemoteAddr().String())

	// 1. Read Reqquest
	lines := make([]string, 0)
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	lines = append(lines, line)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error:", err)
			return
		}
		lines = append(lines, line)
		if line == "\r\n" {
			break
		}
	}
	fmt.Println("Server Request:")
	fmt.Println(lines)

	// 2. Write Response
	fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\n\r\n")
}

func MockServer() {
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

package message

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"
)

func mockRequest() []byte {
	CRLF := "\r\n"
	json := `{"name":"John","age":30,"car":null}`
	contLen := "Content-Length: " + strconv.Itoa(len(json)) + CRLF
	data := "GET / HTTP/1.1\r\nHost: localhost:8080\r\nUser-Agent: curl/7.43.0\r\nAccept: */*\r\n" + contLen + CRLF + json
	fmt.Println("- Input: (string):\n", data)
	fmt.Println("- Input (bytes):\n", []byte(data))
	fmt.Println("- Input length:\n", len(data))
	fmt.Println("******************************")
	return []byte(data)
}

func TestMessage(t *testing.T) {
	input := mockRequest()

	// Create a pair of connected net.Conn objects
	server, client := net.Pipe()

	// Server
	go func() {
		req, err := NewRequestMessageFromConnection(server)
		if err != nil {
			fmt.Println("Server error:")
			panic(err)
		}
		fmt.Println("Server received", len(req.ToBytes()), "bytes")

		// Test server output matches expected client input
		output := req.ToBytes()
		if !bytes.Equal(input, output) {
			t.Error("Expected:", input, "Got:", output)
		}

		server.Close()
	}()

	time.Sleep(2 * time.Second)

	// Client
	fmt.Print("\n\n")
	n, err := client.Write(input) // Blocks until server reads all bytes
	fmt.Println("Client sent", n, "bytes")
	if err != nil {
		t.Log("Client error:", err)
	}
	client.Close() // Sends EOF to server
	time.Sleep(15 * time.Second)
}

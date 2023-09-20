package message

import (
	"fmt"
	"testing"
	"time"
	"util/mock"
)

func TestMessageFromIOReader(t *testing.T) {
	input := mock.PostJSONRequest()
	reader := mock.ReaderFromBytes(input)
	message, err := NewMessage(reader)
	if err != nil {
		t.Error(err)
	}
	message.ToFile("test-message.txt")
	fmt.Println(message)
}

func TestMessageFromConn(t *testing.T) {
	input := mock.PostJSONRequest()

	// Create a pair of connected net.Conn objects
	server, client := mock.Connection()

	// Server
	go func() {
		mes, err := NewMessage(server)
		if err != nil {
			fmt.Println("Server error:")
			panic(err)
		}
		fmt.Println("Bytes:", mes.ToBytes())
		fmt.Println("Bytes:", string(mes.ToBytes()))

		// fmt.Println("Server received", len(req.ToBytes()), "bytes")

		// Test server output matches expected client input
		// output := req.ToBytes()
		// if !bytes.Equal(input, output) {
		// 	t.Error("Expected:", input, "Got:", output)
		// }

		fmt.Println(mes)

		// Close connection
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

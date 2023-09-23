package message

import (
	"bufio"
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

	reader := bufio.NewReader(client)
	writer := bufio.NewWriter(server)

	// Server
	go func() {
		mes := NewMessageReader(reader)
		fmt.Println("Bytes:", mes.ToBytes())
		fmt.Println("String:", mes)

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
	// Write bytes from input to writer's buffer
	n, err := writer.Write(input) // Blocks until server reads all bytes
	if err != nil {
		t.Log("Client error:", err)
	}
	// Write bytes from writer's buffer to client
	err = writer.Flush()
	if err != nil {
		t.Log("Client error:", err)
	}
	fmt.Println("Client sent", n, "bytes")
	client.Close() // Sends EOF to server
	time.Sleep(15 * time.Second)
}

func TestMessageReader(t *testing.T) {
	input := mock.PostJSONRequest()
	reader := mock.ReaderFromBytes(input)
	// message, err := NewMessage(reader)
	// if err != nil {
	// 	t.Error(err)
	// }
	message, err := ParseReaderToMessage(reader)
	if err != nil {
		t.Error(err)
	}
	// message.String()
	fmt.Println(message)
}

func TestParsedBytesToMessage(t *testing.T) {
	input := mock.PostJSONRequest()
	// reader := mock.ReaderFromBytes(input)
	// message, err := NewMessage(reader)
	// if err != nil {
	// 	t.Error(err)
	// }
	message, err := ParseBytesToMessage(input)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(message)
}

package message

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"testing"
	"time"
)

func TestNewMessageFromReader(t *testing.T) {
	// Arrange
	expected, expectedReader, _ := mockRequest()
	got := NewRequest(expectedReader)

	// Compare expected and got
	err := got.Equals(expected)
	if err != nil {
		t.Error(err)
	}
}

func TestNewMessageFromConn(t *testing.T) {
	// Mock server and client connections
	server, client := net.Pipe()

	// Run server
	go testRequestEqualsExpected(t, server)

	// Sleep to allow server to start
	time.Sleep(2 * time.Second)

	// Simulate client
	mockClientRequestAndResponse(t, client)

	fmt.Println("\nFinished.")
}

func TestNewMessageFromBytes(t *testing.T) {
	expected, _, input := mockRequest()

	got := NewRequestFromBytes(input)

	err := got.Equals(expected)
	if err != nil {
		t.Error(err)
	}
}

func TestToFile(t *testing.T) {
	expected, reader, expectedInput := mockRequest()
	expectedSize := int64(len(expectedInput))

	message := NewRequest(reader)

	if err := message.Equals(expected); err != nil {
		t.Error(err)
	}

	// Test ToFile
	gotN, err := message.ToFile("test-message.txt")
	if err != nil {
		t.Error(err)
	}

	// Test bytes written
	if gotN != expectedSize {
		t.Errorf("Expected %d bytes, got %d bytes.", expectedSize, gotN)
	}

	// Read file
	gotFile, err := os.ReadFile("test-message.txt")
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(gotFile, expectedInput) {
		t.Errorf("Expected %s, got %s.", expectedInput, gotFile)
	}
}

func TestReset(t *testing.T) {
	// Arrange
	expected := &Request{}
	got := NewRequestFromBytes(nil)

	// Act
	got.Reset()

	// Assert
	err := got.Equals(expected)
	if err != nil {
		t.Error(err)
	}
}

func TestCopy(t *testing.T) {
	// Arrange
	expected, expectedReader, expectedInput := mockRequest()
	got := NewRequestFromBytes(expectedInput)

	// Act
	got.Copy(expectedReader)

	// Assert
	err := got.Equals(expected)
	if err != nil {
		t.Error(err)
	}
}

//**********************************************************************************************************************
// Helpers
//**********************************************************************************************************************

func testRequestEqualsExpected(t *testing.T, server net.Conn) {
	fmt.Println("Server connection started.")

	// Read request
	got := NewRequest(server)
	fmt.Println("Server received", got.Len(), "bytes.")

	// Test output equals expected output
	expected, _, _ := mockRequest() // mock data
	err := got.Equals(expected)
	if err != nil {
		t.Error(err)
	}

	// Write response
	n, err := server.Write(got.ToBytes()) // write response
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Server wrote response of ", n, "bytes to client.")

	// Close connection
	server.Close()
	fmt.Println("Server closed connection.")
}
func mockRequest() (*Request, *bytes.Buffer, []byte) {
	// input
	reqLine := "POST / HTTP/1.1\r\n"
	headers := "Content-Length: 15\r\nContent-Type: application/json\r\n\r\n"
	body := "{\"name\":\"John\"}"
	input := reqLine + headers + body

	// reader
	reader := bytes.NewBuffer([]byte(input))

	// instance
	m := &Request{
		method:        "POST",
		path:          "/",
		protocol:      "HTTP/1.1",
		headersMap:    map[string]string{"Content-Type": "application/json", "Content-Length": "15"},
		contentType:   "application/json",
		contentLength: 15,
		reqLineBuf:    []byte(reqLine),
		headersBuf:    []byte(headers),
		bodyBuf:       []byte(body),
		len:           len(input),
	}

	return m, reader, []byte(input)
}

func mockServerRequestEqualsExpectedRequest(t *testing.T) {
	// Create a pair of connected network connections
	serverConn, clientConn := net.Pipe()

	// Start the server in a goroutine
	go func() {
		// Read the request from the client
		request := make([]byte, 1024)
		n, err := serverConn.Read(request)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println("Server received", n, "bytes.")

		// Process the request
		response := []byte("Hello, client!")

		// Write the response to the client
		_, err = serverConn.Write(response)
		if err != nil {
			t.Error(err)
			return
		}
	}()

	// Send a request from the client
	request := []byte("Hello, server!")
	_, err := clientConn.Write(request)
	if err != nil {
		t.Error(err)
		return
	}

	// Read the response from the server
	response := make([]byte, 1024)
	n, err := clientConn.Read(response)
	if err != nil {
		t.Error(err)
		return
	}

	// Check that the response matches the expected value
	expected := []byte("Hello, client!")
	if !bytes.Equal(response[:n], expected) {
		t.Errorf("got %q, expected %q", response[:n], expected)
	}
}

func mockClientRequestAndResponse(t *testing.T, client net.Conn) {
	// Mock data
	_, _, input := mockRequest()

	// Write request to buffer
	writer := bufio.NewWriter(client)
	n, err := writer.Write(input) // blocks until server reads all bytes
	if err != nil {
		t.Errorf("Error writing to buffer: %v", err)
	}
	fmt.Printf("\nClient wrote %d bytes to buffer.\n", n)
	time.Sleep(3 * time.Second)

	// Write request to server
	fmt.Println("Client flushed buffer and sent", n, "bytes to server.")
	err = writer.Flush()
	if err != nil {
		t.Errorf("Error flushing buffer: %v", err)
	}
	time.Sleep(3 * time.Second)

	// Read response
	clientReader := bufio.NewReader(client)   //
	clientRequest := NewRequest(clientReader) //
	time.Sleep(2 * time.Second)
	fmt.Printf("\nClient received:\n%s\n", clientRequest)

	// Close connection
	client.Close() // sends EOF to server
	fmt.Println("Client closed connection.")
}

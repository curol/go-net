package message

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

// TODO: Test multiple forms of the request Body (e.g. io.Reader, io.ReadCloser, etc.)
// TODO: Test multiple forms of input and output

func TestNewMessageFromReader(t *testing.T) {
	// Arrange
	expected, expectedReader, _ := mockRequest()

	// Act
	got := NewRequest(expectedReader)

	// Test if request `got` equals request `expected`
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
	fn := "test/test-message.txt"

	message := NewRequest(reader)

	if err := message.Equals(expected); err != nil {
		t.Error(err)
	}

	// Test ToFile
	gotN, err := message.ToFile(fn)
	if err != nil {
		t.Error(err)
	}

	// Test bytes written
	if gotN != expectedSize {
		t.Errorf("Expected %d bytes, got %d bytes.", expectedSize, gotN)
	}

	// Read file
	gotFile, err := os.ReadFile(fn)
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

func TestResponse(t *testing.T) {
	// Arrange
	// expected, expectedReader, expectedInput := mockResponse()
	// got := NewRequestFromBytes(expectedInput)
	response, responseOutput := mockResponse()

	// Act
	// got.Copy(expectedReader)
	response.Text("Hello World!")

	// Assert
	// err := got.Equals(expected)
	// if err != nil {
	// 	t.Error(err)
	// }
	// buffer := bytes.NewBuffer(nil) // client
	// response.WriteTo(buffer)
	// fmt.Println("Got output:", buffer.Bytes())
	fmt.Println("Expected output: ", responseOutput)

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

// MockRequest returns a mock request.
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

func mockResponeOutput() (output string, headers map[string]string, body *bytes.Buffer, lines []string) {
	lines = []string{
		"HTTP/1.1 200 OK\r\n", // status
		"Content-Length: 15\r\nContent-Type: application/json\r\n\r\n", // headers
		"{\"name\":\"John\"}", // body
	}
	output = strings.Join(lines, "")
	headers = map[string]string{"Content-Type": "application/json", "Content-Length": "15"}
	body = bytes.NewBuffer([]byte(lines[2]))

	// TODO: bufio.NewReader(buffer) ?

	// create a reader for the body
	// r := strings.NewReader("")
	// b := bytes.NewReader([]byte(""))
	// body := bytes.NewBuffer([]byte(bodyLine))
	// bodyReader := bufio.NewReader(body)

	return output, headers, body, lines
}

// MockResponse returns a mock response.
func mockResponse() (*Response, []byte) {
	output, headers, _, _ := mockResponeOutput()

	// Convert string to int for content length
	// cl, err := strconv.Atoi(headers["Content-Length"])
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Content-Length:", cl)

	// instance
	m := &Response{
		status:     "200 OK",
		statusCode: 200,
		protocol:   "HTTP/1.1",
		header:     headers,
		body:       NewBody(),
		size:       len(output),
	}

	return m, []byte(output)
}

// MockServerRequestAndResponse mocks a server request and response.
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

// MockClientRequestAndResponse mocks a client request and response.
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

// MockClientRequestEqualsExpectedResponse mocks a client request and response.
func mockClientRequestEqualsExpectedResponse(t *testing.T, client net.Conn) {
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

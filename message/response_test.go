package message

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestResponse(t *testing.T) {
	serverConn, clientConn := net.Pipe()

	printR := func(title string, v []byte) {
		hs := "\n-------------------------------------------------------------\n"
		fmt.Printf("\n%s%s%s%s\n", title, hs, v, hs)
	}

	// Start the server in a goroutine
	go func() {
		// Close connection when finished, which sends EOF to client
		defer serverConn.Close()

		fmt.Println("Server started.")

		// Read Request
		request := NewRequest(serverConn)
		printR("1.) Server request:", request.ToBytes())

		// Send Response
		response := NewResponse(serverConn)
		response.Text("Hello, client!") // write response body
		n, err := serverConn.Write(response.ToBytes())
		if err != nil {
			t.Error(err)
		}
		// err := response.Flush()         // flush buffer to connection
		// if err != nil {
		// 	t.Error(err)
		// }
		fmt.Println("Server sent", n, "bytes.")
	}()

	time.Sleep(2 * time.Second) // wait for server

	// Client
	// 1.) Send request
	req, _ := mockTextRequest()
	n, err := req.WriteTo(clientConn)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Client sent", n, "bytes.")
	// 2.) Read response
	time.Sleep(2 * time.Second)
	response := make([]byte, 1024)
	n2, err := clientConn.Read(response) // read response from server
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println("Client received", n2, "bytes.")
	printR("2.) Client response:", response[:n2])
}

func TestPlainResponse(t *testing.T) {
	// Arrange
	expected, _ := mockTextResponse()
	server, _ := net.Pipe()
	got := NewResponse(server)

	// Act
	body := "Hello, World!"
	got.Text(body)

	// Assert
	err := expected.Equals(got)
	if err != nil {
		t.Error(err)
	}
}

func TestJSONResponse(t *testing.T) {
	// Arrange
	expected, _ := mockJSONResponse()
	server, _ := net.Pipe()
	got := NewResponse(server)

	// Act
	m := map[string]string{"name": "John"}
	err := got.JSON(m)
	if err != nil {
		t.Error(err)
	}

	// Assert
	err = expected.Equals(got)
	if err != nil {
		t.Error(err)
	}
}

func mockJSONResponse() (*Response, []byte) {
	lines := []string{
		"HTTP/1.1 200 OK\r\n", // status
		"Content-Length: 15\r\nContent-Type: application/json\r\n\r\n", // headers
		"{\"name\":\"John\"}", // body
	}
	output := strings.Join(lines, "")
	headers := map[string]string{"Content-Type": "application/json", "Content-Length": "15"}
	body := bytes.NewBuffer([]byte(lines[2]))

	r := &Response{
		protocol:   "HTTP/1.1",
		statusCode: 200,
		statusText: "OK",
		header:     headers,
		body:       body.Bytes(),
		size:       len(output),
	}

	return r, []byte(output)
}

func mockTextResponse() (*Response, []byte) {
	bodyText := "Hello, World!"
	lines := []string{
		"HTTP/1.1 200 OK\r\n", // status
		"Content-Length: 13\r\nContent-Type: text/plain\r\n\r\n", // headers
		bodyText, // body
	}
	output := strings.Join(lines, "")
	headers := map[string]string{"Content-Type": "text/plain", "Content-Length": "13"}
	body := bytes.NewBuffer([]byte(lines[2]))

	r := &Response{
		protocol:   "HTTP/1.1",
		statusCode: 200,
		statusText: "OK",
		header:     headers,
		body:       body.Bytes(),
		size:       len(output),
	}

	return r, []byte(output)
}

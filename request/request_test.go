package request

import (
	"bytes"
	"fmt"
	"net"
	"reader"
	"strconv"
	"testing"
	"time"
)

func TesRequest1(t *testing.T) {
	mockOKResponse := func() []byte {
		// Response
		return []byte("HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\nHello")
	}

	mockRequest := func() []byte {
		CRLF := "\r\n"
		json := `{"name":"John","age":30,"car":null}`
		contLen := "Content-Length: " + strconv.Itoa(len(json)) + CRLF
		data := "GET / HTTP/1.1\r\nHost: localhost:8080\r\nUser-Agent: curl/7.43.0\r\nAccept: */*\r\n" + contLen + CRLF + json
		return []byte(data)
	}

	input := mockRequest()
	//******************************************************************************************

	// Create a pair of connected net.Conn objects
	server, client := net.Pipe()

	// Server
	go func() {
		// Read Request
		req := NewRequest(server)

		// Test request matches expected request
		output := req.ToBytes()
		if !bytes.Equal(input, output) {
			t.Error("Expected:", input, "Got:", output)
		}

		// Write Response
		server.Write([]byte(mockOKResponse()))

		server.Close()
	}()

	time.Sleep(2 * time.Second)

	// Client
	// Write request
	_, err := client.Write(input) // Blocks until server reads all bytes
	if err != nil {
		t.Log("Client error:")
		panic(err)
	}

	response, err := reader.ReadHTTPResponse(client)
	if err != nil {
		t.Log("Client error")
		panic(err)
	}
	if response != nil {
		for name, values := range response.Header {
			// Loop over all values for the name.
			for _, value := range values {
				fmt.Println(name, value)
			}
		}
	}

	client.Close() // Sends EOF to server
	time.Sleep(15 * time.Second)
}

func TestRequest2(t *testing.T) {
	mockResponse := func() []byte {
		// Define your expected response here
		return []byte("HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\nHello")
	}

	mockRequest := func() *Request {
		// Define your request here
		CRLF := "\r\n"
		json := `{"name":"John","age":30,"car":null}`
		contLen := "Content-Length: " + strconv.Itoa(len(json)) + CRLF
		data := "GET / HTTP/1.1\r\nHost: localhost:8080\r\nUser-Agent: curl/7.43.0\r\nAccept: */*\r\n" + contLen + CRLF + json
		return NewRequestFromBytes([]byte(data))
	}

	// testStruct := func(expected interface{}, actual interface{}) {
	// 	if !reflect.DeepEqual(expected, actual) {
	// 		t.Errorf("Struct does not match expected. Expected: %+v, got: %+v", expected, actual)
	// 	}
	// }

	testBytes := func(expected []byte, actual []byte) {
		if !bytes.Equal(expected, actual) {
			t.Errorf("Bytes do not match expected. Expected: %+v, got: %+v", expected, actual)
		}
	}

	mockReq := mockRequest()
	mockRes := mockResponse()

	//****************************************************************************************//
	// Test 1: Test that the response matches the expected response
	//****************************************************************************************//

	// Connect
	server, client := net.Pipe()

	// Server
	//****************************************************************************************//
	go func() {
		// Read request
		req := NewRequest(server)

		// Test server output matches expected client input
		testBytes(mockReq.ToBytes(), req.ToBytes()) // Expected, Actual

		// Write response
		server.Write(mockRes)

		// Close connection
		server.Close()
	}()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Client
	//****************************************************************************************//
	_, err := client.Write(mockReq.ToBytes()) // Blocks until server reads all bytes
	if err != nil {
		t.Log("Client error:")
		panic(err)
	}

	response, err := reader.ReadHTTPResponse(client)
	if err != nil {
		t.Log("Client error")
		panic(err)
	}

	// Convert response to bytes
	respBytes := reader.ConvertResponseToBytes(response)
	t.Log(respBytes)

	client.Close() // Sends EOF to server
	time.Sleep(15 * time.Second)
}

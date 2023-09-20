package mock

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	CRLF = "\r\n"
)

type mock struct{}

func PostJSONRequest() []byte {
	json := `{"name":"John","age":30,"car":null}`
	contLen := "Content-Length: " + strconv.Itoa(len(json)) + CRLF
	data := "POST / HTTP/1.1\r\nHost: localhost:8080\r\nUser-Agent: curl/7.43.0\r\nAccept: */*\r\n" + contLen + CRLF + json
	return []byte(data)
}

func GetRequest() []byte {
	data := "GET / HTTP/1.1\r\nHost: localhost:8080\r\nUser-Agent: curl/7.43.0\r\nAccept: */*\r\n" + CRLF
	return []byte(data)
}

func FileGetRequest() []byte {
	data := "GET /index.html HTTP/1.1\r\nHost: localhost:8080\r\nUser-Agent: curl/7.43.0\r\nAccept: */*\r\n" + CRLF
	return []byte(data)
}

func GetRequestFromFile() io.Reader {
	// Open file
	f, err := os.Open("test/mock/get-request.txt")
	if err != nil {
		panic(err)
	}
	return f
}

// MockConnection returns a network connection using net.Pipe.
func Connection() (net.Conn, net.Conn) {
	// Pipe creates a synchronous, in-memory, full duplex network connection; both ends implement the Conn interface.
	// Reads on one end are matched with writes on the other, copying data directly between the two; there is no internal buffering.
	return net.Pipe()
}

// MockServer reads and prints first line of message and headers of message.
func Server() {
	handleConnection := func(conn net.Conn) {
		defer conn.Close()

		// Create reader from connection
		reader := bufio.NewReader(conn)

		// Get first line
		requestLine, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(strings.TrimSpace(requestLine))

		// Print all header lines
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				return
			}

			// HTTP headers are terminated by a blank line
			if line == "\r\n" {
				break
			}

			fmt.Println(strings.TrimSpace(line))
		}
	}

	start := func() {
		listener, err := net.Listen("tcp", ":8080")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer listener.Close()

		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println(err)
				return
			}
			go handleConnection(conn)
		}
	}

	start()
}

// MockNewStringReader returns a new Reader reading from s.
func ReaderFromString(s string) io.Reader {
	return strings.NewReader(s)
}

// MockNewBytesReader returns a new Reader reading from b.
func ReaderFromBytes(b []byte) io.Reader {
	return bytes.NewReader(b)
}

func WriterToBuffer() {
	// Create a bytes.Buffer
	var b bytes.Buffer

	// Create a bufio.Writer that writes to the bytes.Buffer
	w := bufio.NewWriter(&b)

	// Write some data to the bufio.Writer
	w.WriteString("Hello, World!")

	// Ensure all data has been written to the underlying buffer
	w.Flush()

	// Get the []byte from the bytes.Buffer
	data := b.Bytes()

	fmt.Println(string(data)) // Outputs: Hello, World!
}

// TODO: Clean

/*
Mock connection example:

```
func ServerClientConn(input []byte, f func(net.Conn)(interface{},error)) {
	input := MockPostJSONRequest()

	// Create a pair of connected net.Conn objects
	server, client := net.Pipe()

	// Server
	go func() {
		req, err := NewMessageFromConnection(server)
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
```

*/

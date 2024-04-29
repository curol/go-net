package textproto_test

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/curol/network"
	"github.com/curol/network/net/textproto"
)

func TestExampleTextMessage(t *testing.T) {

	// Server
	handlerFunc := func(conn net.Conn) {
		//
		defer conn.Close()

		// Read the request
		tm, err := textproto.ReadTextMessage(bufio.NewReader(conn))
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Request:", tm)

		fmt.Println("Request string:", string(tm.Bytes()))

		// Write response
		rawRes := []byte("HTTP/1.0 200 OK\r\nUser-Agent: textproto example\r\nType: Response\r\nAccept: */*\r\n\r\n")
		fmt.Fprint(conn, string(rawRes))
	}
	go mockServer(handlerFunc)
	time.Sleep(2 * time.Second)

	// Client
	rawReq := []byte("GET / HTTP/1.0\r\nUser-Agent: textproto example\r\nAccept: */*\r\n\r\n")
	mockClientReq(rawReq)

	// 	// Write req
	// 	fmt.Fprintf(conn, string(rawReq))

	// 	// Read the request
	// 	req := textproto.NewRequest(bufio.NewReader(conn))
	// 	err = req.Read()
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		return
	// 	}
	// 	fmt.Println("Request:", req)

	// 	// Write the response
	// 	res := textproto.NewResponse(bufio.NewWriter(conn))
	// 	// res.StatusLine = "HTTP/1.0 200 OK"
	// 	// res.Headers = map[string][]string{
	// 	// 	"User-Agent": {"textproto example"},
	// 	// 	"Accept":     {"*/*"},
	// 	// }
	// 	rawData := []byte("HTTP/1.0 200 OK\r\nUser-Agent: textproto example\r\nAccept: */*\r\n\r\n")
	// 	_, err = res.Write()
	// 	fmt.Println("Response:", res)
	// }

	// exampleTextMessage()
}

func TestExampleTextMessageWithBody(t *testing.T) {

	// Server
	handlerFunc := func(conn net.Conn) {
		//
		defer conn.Close()

		// Read the request
		tm, err := textproto.ReadTextMessage(bufio.NewReader(conn))
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Request:", tm)

		// Cant use this cause body would already be read
		// fmt.Println("Request string:", string(tm.Bytes()))

		tm.File("example-test-go.txt")

		// Write response
		rawRes := []byte("HTTP/1.0 200 OK\r\nUser-Agent: textproto example\r\nType: Response\r\nAccept: */*\r\n\r\n")
		fmt.Fprint(conn, string(rawRes))
	}
	go mockServer(handlerFunc)
	time.Sleep(2 * time.Second)

	// Client
	lines := []string{
		"GET / HTTP/1.0",
		"User-Agent: textproto example",
		"Accept: */*",
		"Content-Length: 5",
		"",
		"Hello",
	}
	rawReq := []byte(strings.Join(lines, "\r\n"))
	mockClientReq(rawReq)

}

func mockServer(handlerFunc func(conn net.Conn)) {
	// Listen for connections
	ln, err := network.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	// When server finished
	defer ln.Close() // close listener when finished
	defer fmt.Println("Server Finished.")

	fmt.Println("Server started.")

	// Without continious loop, accept single connection
	conn, err := ln.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Connection accepted from:", conn.RemoteAddr().String())

	// Handle the connection
	handlerFunc(conn)
}

func mockClientReq(rawReq []byte) {
	// Connect to the server
	conn, err := network.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Write request
	fmt.Fprint(conn, string(rawReq))

	// Read the response
	tm, err := textproto.ReadTextMessage(bufio.NewReader(conn))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Response:", tm)
}

// func TestExampleReq(t *testing.T) {
// 	exampleReq()
// }

// func TestExampleRes(t *testing.T) {
// 	exampleRes()
// }

// func TestExampleServer(t *testing.T) {
// 	// Server
// 	go exampleServer()

// 	// Wait for server to start
// 	time.Sleep(2 * time.Second)

// 	// Client
// 	conn, _ := net.Dial("tcp", "localhost:8080")
// 	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
// 	time.Sleep(5 * time.Second)
// }

// func exampleReq() {
// 	// Connect to the server
// 	conn, err := network.Dial("tcp", "golang.org:80")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	defer conn.Close()

// 	// Mock request
// 	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")

// 	// Read request
// 	req := textproto.NewRequest(bufio.NewReader(conn)) // Read from connection
// 	req.Read()

// 	fmt.Println(req)
// }

// func exampleRes() {
// 	// Connect to the server
// 	conn, err := network.Dial("tcp", "golang.org:80")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer conn.Close()

// 	// Write response
// 	res := textproto.NewResponse(bufio.NewWriter(conn)) // Write to connection
// 	defer res.Close()
// 	_, err = res.Write([]byte("HTTP/1.0 200 OK"))
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// }

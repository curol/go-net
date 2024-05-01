package textproto_test

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/curol/network/net/textproto"
)

func TestExampleTextMessage(t *testing.T) {
	// 1. Server
	handlerFunc := func(conn net.Conn) {
		//
		defer conn.Close()

		// Read the request
		tm, err := textproto.ReadTextMessage(bufio.NewReader(conn))
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Server Request:", tm)
		fmt.Println("Server Request string:", string(tm.Bytes()))

		// Write response
		rawRes := []byte("HTTP/1.0 200 OK\r\nUser-Agent: textproto example\r\nType: Response\r\nAccept: */*\r\n\r\n")
		fmt.Fprint(conn, string(rawRes))
	}
	go mockServer(handlerFunc)
	time.Sleep(2 * time.Second)

	// 2. Client
	rawReq := []byte("GET / HTTP/1.0\r\nUser-Agent: textproto example\r\nAccept: */*\r\n\r\n")
	mockClientReq(rawReq)
}

func TestExampleTextMessageWithBody(t *testing.T) {
	// 1. Server
	go mockServer(mockStdoutHandlerFunc)
	time.Sleep(2 * time.Second)
	// 2. Client
	lines := []string{
		"GET / HTTP/1.0",
		"User-Agent: textproto example",
		"Accept: */*",
		"Content-Length: 5",
		"",
		"Hello",
	}
	rawReq := []byte(strings.Join(lines, "\r\n"))
	fmt.Println("Client request size:", len(rawReq))
	mockClientReq(rawReq)
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

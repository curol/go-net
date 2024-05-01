package textproto_test

import (
	"bufio"
	"fmt"
	"net"

	"github.com/curol/network"
	"github.com/curol/network/net/textproto"
)

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

func mockStdoutHandlerFunc(conn net.Conn) {
	// 1. Close connection when done
	defer conn.Close()
	// 2. Read the request text message
	tm, err := textproto.ReadTextMessage(bufio.NewReader(conn))
	if err != nil {
		fmt.Println(err)
		return
	}
	// 3. Handle the request
	fmt.Println("\n------")
	fmt.Println("\n\n\nHandler Request:\n", tm)
	fmt.Println("- Content Length:", tm.ContentLength())
	fmt.Println("- Content Type:", tm.ContentType())
	fmt.Println("- Size:", tm.Size())
	tm.StdOut()
	fmt.Println("\n------")
	// 4. Write the response
	rawRes := []byte("HTTP/1.0 200 OK\r\nUser-Agent: textproto example\r\nType: Response\r\n\r\n")
	fmt.Fprint(conn, string(rawRes))
}

func mockFilehandlerFunc(conn net.Conn) {
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

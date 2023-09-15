package server

import (
	"fmt"
	"log"
	"net"
	"request"
	"testing"
	"time"
	"writer"
)

func mockData() []string {
	const CRLF = "\r\n"
	jsonString := `{"key":"test","value":"1234"}`

	inputs := []string{
		"PING" + CRLF + CRLF,
		"PING" + CRLF + CRLF,
		"PING" + CRLF + CRLF,
		"PING" + CRLF + CRLF,
		"GET / HTTP/1.1" + CRLF + CRLF,
		"GET / HTTP/1.1\r\nHost: localhost:8080\r\nUser-Agent: curl/7.43.0\r\nAccept: */*" + CRLF + CRLF,
		"GET / HTTP/1.1\r\nHost: localhost:8080\r\nUser-Agent: curl/7.43.0\r\nAccept: */*" + CRLF + CRLF,
		"POST / HTTP/1.1\r\nHost: localhost:8080\r\nUser-Agent: curl/7.43.0\r\nAccept: */*" + CRLF + CRLF,
		"POST / HTTP/1.1\r\nHost: localhost:8080\r\nUser-Agent: curl/7.43.0\r\nAccept: */*\r\nContent-Type: application/json" + CRLF + CRLF + jsonString,
	}
	return inputs
}

func TestServer(t *testing.T) {
	network := "tcp4"
	address := "localhost:8080"

	// Handlers
	handler := func(r *request.Request, w *writer.ResponseWriter) {
		w.Write([]byte("Hello World"))
	}

	// Server
	server := NewServer(network, address, nil)
	server.GET("/", handler)

	// Run server
	go server.Run()

	// Wait for server to start
	time.Sleep(1 * time.Second)

	// Cllient
	for i, input := range mockData() {
		fmt.Printf("%d.) *****************************************************************************\n", i+1)

		// Connect
		log.Println("Test: Client Connecting to server")
		conn, err := net.Dial("tcp4", "localhost:8080")
		if err != nil {
			log.Fatal(err)
		}

		// Write
		fmt.Println("Test: First Write...")
		conn.Write([]byte(input))

		// Read
		log.Println("Test: Client Reading connection")
		buffer := make([]byte, 1024)
		conn.Read(buffer)
		log.Println("\nTest: Client - Message from server: " + string(buffer))

		// Close
		log.Println("Test: Client Closing connection")
		conn.Close()
		fmt.Print("\n\n")
	}
}

// func addHandlers(router *router.Router) {
// 	router.PING("/", pingHandler)
// 	router.GET("/", getHandler)
// 	router.POST("/", postHandler)
// }

// func testClient(network string, address string) {
// 	const CRLF = "\r\n"
// 	jsonString := `{"key":"test","value":"1234"}`

// 	// Data from client
// 	var (
// 		TestMessageInputPING = "PING" + CRLF + CRLF
// 		TestMessageInputGet  = "GET / HTTP/1.1\r\nHost: localhost:8080\r\nUser-Agent: curl/7.43.0\r\nAccept: */*" + CRLF + CRLF
// 		TestMessageInputPost = "POST / HTTP/1.1\r\nHost: localhost:8080\r\nUser-Agent: curl/7.43.0\r\nAccept: */*" + CRLF + CRLF
// 		TestMessageInputJSON = "POST / HTTP/1.1\r\nHost: localhost:8080\r\nUser-Agent: curl/7.43.0\r\nAccept: */*\r\nContent-Type: application/json" + CRLF + CRLF + jsonString
// 	)

// 	// Data to send to server
// 	inputs := []string{TestMessageInputPING, TestMessageInputGet, TestMessageInputPost, TestMessageInputJSON}

// 	// Range over inputs
// 	for i, input := range inputs {
// 		// Print test number
// 		fmt.Printf("%d.) *****************************************************************************\n", i+1)
// 		fmt.Println("- TEST INPUT", i+1, "=", []byte(input))
// 		fmt.Println()

// 		// Connect to server
// 		conn, err := net.Dial("tcp4", "localhost:8080")
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		// Write
// 		fmt.Fprintf(conn, input)

// 		// Read
// 		buffer := make([]byte, 1024)
// 		conn.Read(buffer)
// 		fmt.Println("\nMessage from server: " + string(buffer))

// 		// Close
// 		conn.Close()

// 		fmt.Print("\n\n")
// 	}
// }

// func TestServer(t *testing.T) {
// 	server := NewServer()
// 	// Run server
// 	go server.Run()
// 	// Wait for server to start
// 	time.Sleep(1 * time.Second)
// 	// Connect to server
// 	conn, err := net.Dial("tcp4", "localhost:8080")
// 	// Handle error
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	// Send message
// 	fmt.Fprintf(conn, "PING\r\n")
// 	// Read response
// 	message, err := bufio.NewReader(conn).ReadString('\n')
// 	// Handle error
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	// Print response
// 	fmt.Print("Message from server: " + message)
// 	// Close connection
// 	conn.Close()
// }

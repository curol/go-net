package server

import (
	"fmt"
	"log"
	"message"
	"net"
)

// Request is a structure that represents an HTTP request received by a server or to be sent by a client.
type Request = message.Request

type Response = message.Response

func Run(address string) {
	// Config
	config := NewConfig(address)

	// Server
	server := NewServer(config)

	// Listen for connections and serve connection using config.handler
	server.listenAndServe()
}

type Server struct {
	// Config
	config *Config
	// Connection
	listener net.Listener
	// Misc
	log     Log
	handler Handler
}

func NewServer(config *Config) *Server {
	server := &Server{
		listener: nil,
		config:   config,
		// Misc
		log:     config.Log,
		handler: config.Handler,
	}
	return server
}

func (s *Server) listenAndServe() {
	// ListenAndServe listens on the TCP network and accepts incoming connections concurrently.
	// The handler handles a request and response for the client.

	network := s.config.Network
	address := s.config.Address

	// Listen for connections
	listener, err := net.Listen(network, address)
	if err != nil {
		panic(err)
	} else {
		s.listener = listener
		fmt.Println("Server listening on " + address)
	}

	// Clean up when finished
	defer s.clean()

	// Run listener forever
	for {
		conn, err := listener.Accept() // wait for next connection and accept
		if err != nil {
			log.Fatal(err) // log status
		}
		go s.serve(conn) // serve and handle connection
	}
}

func (s *Server) serve(conn net.Conn) {
	// Clean
	defer conn.Close() // close connection when this is finished

	// Init
	s.initConnectionProps(conn)             // set connection properties
	req := message.NewRequestFromConn(conn) // create new request
	res := message.NewResponse(conn)        // create new response

	// Log
	s.log.Status(req)

	// Serve connection
	s.handler.ServeConn(res, req)

	// TODO: After handler finishes, serialize response?
	// TODO: After handler finishes, flush ResponseWriter?
	err := res.WriteOutput()
	if err != nil {
		panic(err)
	}
}

func (s *Server) initConnectionProps(conn net.Conn) {
	// Set read and write deadlines
	err := conn.SetDeadline(s.config.Deadline)
	if err != nil {
		log.Fatal(err)
	}
}

// Clean cleans up the server when listener is stopped
func (s *Server) clean() {
	s.listener.Close() // close listener
	fmt.Println("Server closed")
}

/*
## ServeHTTP Example
```
package main

import (
    "fmt"
    "net/http"
)

type MyHandler struct{}

func (h MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, client!")
}

func main() {
    // Create a new instance of MyHandler
    handler := MyHandler{}

    // Start the server and listen for requests
    http.ListenAndServe(":8080", handler)
}
```

## HTTP Server Client example
```
package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "strings"
)

func main() {
    // Start the server
    go func() {
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            fmt.Fprintf(w, "Hello, client!")
        })
        http.ListenAndServe(":8080", nil)
    }()

    // Send a request to the server
    resp, err := http.Get("http://localhost:8080/")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    // Read the response body
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }

    // Print the response body
    fmt.Println(strings.TrimSpace(string(body)))
}
```
*/
package message

import (
	"fmt"
	"log"
	"net"
)

// ******************************************************
// ResponseWriter
// ******************************************************

// A ResponseWriter interface is used by an HTTP handler to construct an HTTP response.
//
// Note, a ResponseWriter may not be used after [Handler.ServeHTTP] has returned.
type ResponseWriter interface {
	Write(b []byte) (int, error)
	WriteHeader(string, string)
	Header() Header
}

// ******************************************************
// Handlers
//
// Handlers handle the client's request and response
// ******************************************************

// A Handler responds to an HTTP request.
type Handler interface {

	// ServeHTTP should write reply headers and data to the [ResponseWriter]
	// and then return. Returning signals that the request is finished; it
	// is not valid to use the [ResponseWriter] or read from the
	// [Request.Body] after or concurrently with the completion of the
	// ServeHTTP call.
	ServeConn(ResponseWriter, *Request)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(ResponseWriter, *Request)

// ServeConn calls f(w, r).
func (f HandlerFunc) ServeConn(w ResponseWriter, r *Request) {
	f(w, r)
}

// HandleFunc registers the handler function for the given pattern.
func HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {}

// ******************************************************
// ListenAndServe
//
// ListenAndServe listens on the TCP network and accepts incoming connections concurrently.
// The handler handles a request and response for the client.
// ******************************************************

// Listen for connections and handle client's request
func ListenAndServe(network string, address string, handler Handler) {
	// Listen for connections
	listener, err := net.Listen(network, address)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Server listening on " + address)
	}

	// Close the listener when the application closes.
	defer listener.Close()
	defer fmt.Println("Server closed")

	// Run listener forever
	for {
		conn, err := listener.Accept() // wait for next connection and accept
		if err != nil {
			log.Fatal(err) // log status
		}
		go handleServingConn(conn, handler) // serve and handle connection
	}
}

func handleServingConn(conn net.Conn, handler Handler) {
	req := NewRequest(conn)
	res := NewResponse(conn)

	// Serve connection
	handler.ServeConn(res, req)
}

// ******************************************************
// Server
//
// Servers handles the client connections
// ******************************************************

// // Server handles connections to clients
// type Server struct {
// 	address string
// 	network string
// }

// // Create a new server
// func NewServer(network string, address string) Server {
// 	// Create new server
// 	server := Server{
// 		network: network,
// 		address: address,
// 	}
// 	return server
// }

// func (s *Server) ServeConn(w ResponseWriter, r *Request) {
// 	// Get request
// 	req := NewRequest(conn)

// 	// Write response
// 	res := NewResponse(conn)
// 	res.Write(req.ToBytes())
// }

/*
// Serve handles connection to client
func Serve(conn net.Conn, server *Server) {
	// Clean up connection when done
	defer clean(conn)

	// Log status
	log.Println("Accepted connection from " + conn.RemoteAddr().String())

	// Set connection properties for server
	conn.SetDeadline(server.config.DeadLine)

	// Read request
	req := request.NewRequest(conn)

	//  Write response
	conn.Write(req.ToBytes())
}

// // Serve handles connection to client
// func Serve(conn net.Conn, server *Server) {
// 	// Clean up connection when done
// 	defer clean(conn)

// 	// Log status
// 	log.Println("Accepted connection from " + conn.RemoteAddr().String())

// 	// Set connection properties for server
// 	// conn.SetReadDeadline(server.config.ReadDeadLine)

// 	// Get request
// 	fmt.Println("Serve: Getting request")
// 	request := request.NewRequest(conn)
// 	conn.Write(request.Data())

// 	fmt.Println("Serve: Request: ", request)
// 	// Route request to handler
// 	fmt.Println("Serve: Routing request")
// 	// server.Route(request, request.Writer())
// }

// Cleanup connection to client
func clean(conn net.Conn) {
	// Get address
	address := conn.RemoteAddr().String()

	// Close connection to client
	conn.Close()

	// Log status
	log.Println("Closed connection to " + address)
}


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

package server

import (
	"fmt"
	"gonet"
	"log"
	"net"
)

// Request is a structure that represents an HTTP request received by a server or to be sent by a client.
type Request = gonet.Request

type Response = gonet.Response

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
	s.setConnectionProps(conn) // set connection properties

	// Read Request
	req := gonet.NewRequestFromConn(conn) // read request

	// Log
	s.log.Status(req.Path(), req.Method(), conn.RemoteAddr().String())

	// Response
	res := gonet.NewResponse(conn) // write respone

	// Handler
	s.handler.ServeConn(res, req)

	// Write response
	_, err := res.WriteOutput() // write response output to its writer `w`
	if err != nil {
		panic(err)
	}
	// TODO: After handler finishes, serialize response?
	// TODO: After handler finishes, flush ResponseWriter?

}

func (s *Server) setConnectionProps(conn net.Conn) {
	// Set read and write deadlines
	err := conn.SetDeadline(s.config.Deadline)
	if err != nil {
		panic(err)
	}
}

// Clean cleans up the server when listener is stopped
func (s *Server) clean() {
	err := s.listener.Close() // close listener
	if err != nil {
		panic(err)
	}
	fmt.Println("Server closed")
}

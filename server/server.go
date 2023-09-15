package server

import (
	"fmt"
	"log"
	"net"
	"router"
)

// Server handles connections to clients
type Server struct {
	address string
	network string
	config  Config
	*router.Router
}

type ServerOptions struct {
	MaxConnections int
	MaxReadSize    int
}

// Create a new server
func NewServer(network string, address string, options *Config) Server {
	// Create new server
	server := Server{
		network: network,
		address: address,
		config:  NewConfig(options),
		Router:  router.NewRouter(),
	}

	// TODO: Add handlers to router
	// Add handlers to router
	// addHandlers(server.Router)

	return server
}

// Listen for connections and handle client's request
func (s *Server) Run() {
	// Get listener
	listener, err := net.Listen(s.network, s.address)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Server listening on " + s.address)
	}

	// Close the listener when the application closes.
	defer listener.Close()
	defer fmt.Println("Server closed")

	// Run listener forever
	for {
		// Wait for next connection and return connection to client
		conn, err := listener.Accept()
		// Log status
		if err != nil {
			log.Fatal(err)
		}
		// Serve connection
		go Serve(conn, s)
	}
}

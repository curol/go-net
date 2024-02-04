package tests

import (
	"fmt"
	"testing"

	http "github.com/curol/network/http"
)

func TestServer(t *testing.T) {
	// var s *Server
	// var conn net.Conn
	// var err error

	// Test listenAndServe

	server := http.NewServer("tcp", "localhost:8080")
	fmt.Println(server)
}

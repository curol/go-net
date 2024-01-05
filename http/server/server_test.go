package server

import (
	"testing"
)

func TestServer(t *testing.T) {
	// network := "tcp"
	address := "localhost:8080"
	Run(address)
}

// type mockRouterHandler struct{}

// func (h *mockRouterHandler) ServeConn(w ResponseWriter, r *Request) {
// 	// Arrange
// 	fmt.Println(w)
// 	fmt.Fprintf(w, "Hello, client!")

// 	// Router

// 	// Flush
// }

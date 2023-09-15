package server

import (
	"log"
	"net"
	"request"
)

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

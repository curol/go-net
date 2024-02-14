package examples

import (
	"io"
	"log"
	"net"
)

func server(conn net.Conn) {
	io.Copy(conn, conn) // Echo all incoming data back to the client
	conn.Close()
}

func client(conn net.Conn) {
	conn.Write([]byte("Hello from client!"))
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	log.Println("Received from server: ", string(buf[:n]))
	conn.Close()
}

func runPipeExample() {
	cConn, sConn := net.Pipe()

	go server(sConn)
	go client(cConn)
}

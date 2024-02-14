package tests

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	http "github.com/curol/network/http"
)

func TestServerShutdown(t *testing.T) {
	// Arrange
	server := http.NewServer("tcp", "localhost:8080") // create server
	// Act
	go server.Run()             // serve
	time.Sleep(2 * time.Second) // wait for server to start
	err := server.Shutdown()
	// Assert
	if err != nil || server.IsShutdown() != true {
		t.Fatal(err)
	}
}

func TestServerRun(t *testing.T) {
	// Arrange
	server := http.NewServer("tcp", "localhost:8080") // create server
	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		server.Logger.Info("Test Handler called...")
		w.Write([]byte("Hello world from handler..."))
	})

	// Server
	go func() {
		err := server.Run()
		if err != nil {
			server.Logger.Warn("Test server failed to run...")
			panic(err)
		}
	}()

	defer server.Shutdown()

	// Client
	time.Sleep(2 * time.Second)
	client := http.NewClient("GET", "localhost:8080", nil, nil)
	resp := client.Do()
	buf := bytes.NewBuffer(nil)
	resp.WriteTo(buf)
	fmt.Println(buf.String())
}

func TestServerRunWithPipe(t *testing.T) {
	// Server
	server := http.NewServer("tcp", "localhost:8080") // create server
	server.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		server.Logger.Info("Test Handler called...")
		rw.Write([]byte("Hello world from handler..."))
	})
	go func() {
		err := server.Run()
		if err != nil {
			server.Logger.Warn("Test server failed to run...")
			panic(err)
		}
	}()
	defer server.Shutdown()

	// Client
	time.Sleep(2 * time.Second)
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	// Write request
	go func() {
		_, err = conn.Write([]byte("GET / HTTP/1.1\r\nHost: localhost:8080\r\n\r\n"))
		if err != nil {
			panic(err)
		}
	}()
	// Read response
	go func() {
		buf := make([]byte, 1024)
		conn.Read(buf)
		fmt.Println(string(buf))
		// Close connection
		conn.Close()
	}()
	time.Sleep(2 * time.Second)
	fmt.Println("Finished...")

	// server.conn.Read(buf)
	// client := http.NewClient("GET", "localhost:8080", nil, nil)
	// resp := client.Do()
	// // Read response
	// buf := bytes.NewBuffer(nil)
	// resp.WriteTo(buf)
	// fmt.Println(buf.String())
}

func TestServerClientPipe(t *testing.T) {
	// Client
	time.Sleep(2 * time.Second)
	cConn, sConn := net.Pipe()
	// Server
	go func() {
		io.Copy(sConn, sConn)
		sConn.Close()
	}()
	// Client
	go func() {
		// Write
		cConn.Write([]byte("GET / HTTP/1.1\r\nHost: localhost:8080\r\n\r\n"))
		buf := make([]byte, 1024)
		// Read
		n, _ := cConn.Read(buf)
		fmt.Println("Received from server: ", string(buf[:n]))
		cConn.Close()
	}()

	time.Sleep(2 * time.Second)
}

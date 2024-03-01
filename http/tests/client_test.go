package tests

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	http "github.com/curol/network/http"
	"github.com/curol/network/http/tests/mock"
)

func TestClient(t *testing.T) {
	// Arrange
	client := http.NewClient("GET", "www.google.com:80", nil, nil)
	// Act
	resp := client.Do()
	// Assert
	if resp == nil {
		t.Fatal("Response is nil")
	}
}

func TestClientGet(t *testing.T) {
	// Arrange
	// client := http.NewClient("GET", "www.google.com:80", nil, nil)
	// Act
	resp := http.Get("www.google.com:80", nil, nil)
	// Assert
	if resp == nil {
		t.Fatal("Response is nil")
	}
}

func TestMockClient(t *testing.T) {
	// Arrange
	rawget := "GET / HTTP/1.1\r\nHost: localhost:8080\r\n\r\n"
	go mock.MockServer()
	time.Sleep(2 * time.Second)

	conn, err := net.Dial("tcp", "localhost:8080")
	defer conn.Close()
	if err != nil {
		t.Fatal(err)
	}
	// Write request
	writer := bufio.NewWriter(conn)
	n, err := writer.Write([]byte(rawget))
	fmt.Println("n: ", n)
	err = writer.Flush()

	// Read response
	r := bufio.NewReader(conn)
	lines := make([]string, 0)
	line, _ := r.ReadString('\n')
	lines = append(lines, line)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error:", err)
			return
		}
		lines = append(lines, line)
		if line == "\r\n" {
			break
		}
	}

	fmt.Println("Client Response:")
	fmt.Println(lines)
}

func TestNetClient(t *testing.T) {
	conn, err := net.Dial("tcp", "www.google.com:80")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer conn.Close()

	fmt.Fprintf(conn, "GET / HTTP/1.1\r\n")
	fmt.Fprintf(conn, "Host: www.google.com\r\n")
	fmt.Fprintf(conn, "\r\n")

	// First line
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// Headers
	for {
		line, err := reader.ReadString('\n')
		fmt.Println("Line:", line)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if line == "\r\n" {
			break
		}
	}
	fmt.Println(response)
}

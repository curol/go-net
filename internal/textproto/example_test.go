package textproto_test

import (
	"bufio"
	"fmt"

	"github.com/curol/network"
	"github.com/curol/network/internal/textproto"
)

func ExampleNewReader() {
	// Connect to the server
	conn, err := network.Dial("tcp", "golang.org:80")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Write an HTTP request
	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")

	// Create a textproto.Reader
	reader := textproto.NewReader(bufio.NewReader(conn))

	// Read the status line
	status, err := reader.ReadLine()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(status)

	// Read the headers
	headers, err := reader.ReadMIMEHeader()
	if err != nil {
		fmt.Println(err)
		return
	}
	for k, v := range headers {
		fmt.Println(k+":", v)
	}
}

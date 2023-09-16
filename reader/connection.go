package reader

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
)

// Read reader and return slice of lines seperated by the delimeter
// In order for this to work, the connection must be closed or else it will block forever.
func ReadConnectionIntoBuffer(conn net.Conn) []byte {
	buffer := make([]byte, 1024)
	data := make([]byte, 0)

	// 2 ways to break from reading stream.
	// 1. The client or server needs to close the connection.
	// 2. A protocol needs to be implemented to know when to stop reading from stream.
	for {
		// Read blocks until client writes data to connection.
		n, err := conn.Read(buffer)

		if err != nil {
			// if err == io.EOF { }
			fmt.Println("Connection closed:", err)
			break
		}
		data = append(data, buffer[:n]...)
	}

	return data
}

// Copy TCP connection to buffer
func CopyConnectionToBuffer(conn net.Conn) *bytes.Buffer {
	var buffer bytes.Buffer
	_, err := io.CopyBuffer(&buffer, conn, nil)
	if err != nil {
		fmt.Println(err)
	}
	return &buffer
}

// Read reader and return slice of lines seperated by the delimeter
func ReadReader(reader *bufio.Reader, delim byte) []byte {
	lines := make([]byte, 0)

	// Read data from the connection line by line
	for {
		// ReadBytes reads until the first occurrence of delim in the input, returning a slice containing the data up to and including the delimiter.
		// ReadBytes returns err != nil if and only if the returned data does not end in delim.
		// For simple uses, a Scanner may be more convenient.
		line, err := reader.ReadBytes(delim)

		// When returned err != nil, returned data does not end in delimeter
		// Therefore, break from loop because done reading.
		if err != nil {
			fmt.Println("Error reading:", err)
			break
		}

		lines = append(lines, line...)
	}

	return lines
}

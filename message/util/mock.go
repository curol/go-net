package util

import (
	"bufio"
	"bytes"
	"io"
	"net"
)

// MockConnection returns a pair of connected network connections (e.g., a server and client connection).
func mockConnection() (net.Conn, net.Conn) {
	return net.Pipe()
}

// MockReader returns a mock io.Reader which streams data.
//
// Note, an io.Reader is the interface that wraps the basic Read method.
// Read reads up to len(p) bytes into p.
// It returns the number of bytes read (0 <= n <= len(p))
func mockReader(data []byte) io.Reader {
	return bytes.NewReader(data)
}

// MockReadCloser returns a mock io.ReadCloser which implements a close method that does nothing.
func mockReadCloser(data []byte) io.ReadCloser {
	return io.NopCloser(bytes.NewReader(data))
}

// MockWriter returns a mock io.Writer.
func mockWriter() io.Writer {
	return new(bytes.Buffer)
}

// MockBufioWriter returns a mock *bufio.Writer.
func mockBufioWriter(w io.Writer) *bufio.Writer {
	return bufio.NewWriter(w)
}

// MockBufioReader returns a mock *bufio.Reader.
func mockBufioReader(r io.Reader) *bufio.Reader {
	return bufio.NewReader(r)
}

// MockBuffer returns a mock bytes.Buffer with data as the underlying buffer.
//
// Note, a Buffer is a variable-sized buffer of bytes with Read and Write methods.
// The zero value for Buffer is an empty buffer ready to use.
func mockBuffer(data []byte) *bytes.Buffer {
	return bytes.NewBuffer(data)
}

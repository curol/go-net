package gonet

// Returns error if reader is nil
import (
	"bufio"
	"message"
	"net"
)

// RequestReaderInterface wraps a connection and provides methods for reading data from the connection.
type RequestReaderInterface interface {
	Read() (*message.Message, error)
}

// RequestReader wraps a connection and provides methods for reading data from the connection.
type RequestReader struct {
	conn net.Conn
	r    *bufio.Reader
}

func NewRequestReader(conn net.Conn) *RequestReader {
	return &RequestReader{
		conn: conn,
		r:    bufio.NewReader(conn),
	}
}

func (rr *RequestReader) Read() (*message.Message, error) {
	return message.NewMessage(rr.r)
}

func (rr *RequestReader) Reader() *bufio.Reader {
	return rr.r
}

func (rr *RequestReader) Conn(data []byte) net.Conn {
	return rr.conn
}

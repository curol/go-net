package reader

// Returns error if reader is nil
import (
	"bufio"
	"bytes"
	"message"
	"net"
	"net/http"
)

// RequestReaderInterface wraps a connection and provides methods for reading data from the connection.
type RequestReaderInterface interface {
	Read() (*message.RequestMessage, error)
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

func (rr *RequestReader) Read() (*message.RequestMessage, error) {
	return message.NewRequestMessage(rr.r)
}

func (rr *RequestReader) Reader() *bufio.Reader {
	return rr.r
}

func (rr *RequestReader) Conn(data []byte) net.Conn {
	return rr.conn
}

func ReadHTTPResponse(conn net.Conn) (*http.Response, error) {
	reader := bufio.NewReader(conn)
	return http.ReadResponse(reader, nil)
}

func ReadHTTPRequest(conn net.Conn) (*http.Request, error) {
	reader := bufio.NewReader(conn)
	return http.ReadRequest(reader)
}

func ConvertResponseToBytes(response *http.Response) []byte {
	// Convert response to bytes
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	return buf.Bytes()
}

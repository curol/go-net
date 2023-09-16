package reader

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
)

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

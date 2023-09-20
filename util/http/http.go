package http

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httputil"
)

func DumpRequest(req *http.Request, body bool) {
	httputil.DumpRequest(req, body)
}

func DumpResponse(res *http.Response, body bool) {
	httputil.DumpResponse(res, body)
}

func ReadRequestFromConnection(conn net.Conn) (*http.Request, error) {
	return ReadRequest(bufio.NewReader(conn))
}

func ReadResponseFromConnection(conn net.Conn) (*http.Response, error) {
	return ReadResponse(bufio.NewReader(conn), nil)
}

func ReadRequest(reader *bufio.Reader) (*http.Request, error) {
	return http.ReadRequest(reader)
}

func ReadResponse(reader *bufio.Reader, req *http.Request) (*http.Response, error) {
	return http.ReadResponse(reader, req)
}

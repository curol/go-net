package httputil

import (
	"bufio"
	"bytes"
	"fmt"
	net "net/http"
	"net/http/httputil"
)

func ReadRequest(raw []byte) *net.Request {
	// Read
	hreq, err := net.ReadRequest(bufio.NewReader(bytes.NewReader(raw))) // net/http.Request
	if err != nil {
		panic(err)
	}
	return hreq
}

func WriteRequest(req *net.Request) *bytes.Buffer {
	// Write
	buf := bytes.NewBuffer(nil)
	req.Write(buf)
	fmt.Println("net/http.Request")
	fmt.Println(buf.String())
	return buf
}

func DebugRequest(req *net.Request) {
	buf, err := httputil.DumpRequest(req, true)
	if err != nil {
		panic(err)
	}
	fmt.Println("net/http.Request")
	fmt.Println(string(buf))
}

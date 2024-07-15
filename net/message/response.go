package message

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/curol/network/net/header"
)

// Res is a wrapper for textproto.Writer which implements convenience methods for reading responses.
type Response struct {
	Status string // 200 OK
	// Code   int    // 200
	Proto  string // TextProto/1.0
	Header header.Header
	Body   *bufio.Reader
	// *TextMessage
	isHeaderWritten bool
	isBodyWritten   bool
	w               io.Writer // Underlying writer
}

func NewResponse(w io.Writer) *Response {
	res := &Response{w: w}
	res.Status = "200 OK"
	res.Proto = "TM/1.0"
	res.Header = header.NewHeader(nil)
	res.Body = nil
	res.w = w

	return res
}

func (res *Response) Write(p []byte) (int64, error) {
	res.Body = bufio.NewReader(bytes.NewReader(p))
	return res.serialize(bufio.NewWriter(res.w))
}

func (res *Response) OK() {
	res.Status = "200 OK"
}

func (res *Response) NotFound() {
	res.Status = "404 Not Found"
}

func (res *Response) BadRequest() {
	res.Status = "400 Bad Request"
}

func (res *Response) Unauthorized() {
	res.Status = "401 Unauthorized"
}

func (res *Response) WriteStatus(s string) {
	res.Status = s
}

func (res *Response) WriteHeader(k string, v []string) {
	res.Header[k] = v
}

func (res *Response) WriteJSON(p []byte) (int64, error) {
	res.Header["Content-Type"] = []string{"application/json"}
	return res.Write(p)
}

func (res *Response) WriteText(p []byte) (int64, error) {
	res.Header["Content-Type"] = []string{"text/plain"}
	return res.Write(p)
}

func (res *Response) WriteHTML(p []byte) (int64, error) {
	res.Header["Content-Type"] = []string{"text/html"}
	return res.Write(p)
}

func (res *Response) serialize(w *bufio.Writer) (int64, error) {
	n := int64(0)

	if !res.isHeaderWritten {
		// Write the status line
		nn, err := fmt.Fprintf(w, "%s\r\n", res.Status)
		if err != nil {
			return n, err
		}
		n += int64(nn)
		// Content len
		contLen := []string{fmt.Sprintf("%d", res.Body.Size())} // Set the content length
		if contLen[0] != "0" {
			res.Header["Content-Length"] = contLen
		}
		// Write the headers
		for k, v := range res.Header {
			for _, vv := range v {
				nn, err := fmt.Fprintf(w, "%s: %s\r\n", k, vv)
				if err != nil {
					return n, err
				}
				n += int64(nn)
			}
		}
		// Write blank line
		nn, err = fmt.Fprintf(w, "\r\n")
		if err != nil {
			return n, err
		}
		res.isHeaderWritten = true
	}

	// Body
	// Write the body
	if res.Body != nil {
		// io.CopyN(w, res.Body, resp.)
		bn, err := io.Copy(w, res.Body)
		n += int64(bn)
		if err != nil {
			return n, err
		}
		res.isBodyWritten = true
	}
	return n, w.Flush()
}

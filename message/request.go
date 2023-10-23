// **********************************************************************************************************************
// Request
//
// Request is a reader, buffer, and parser for a client requests.
// It handles reading a client request.
//
// For brevity, the protocol for a message request follows a stripped down, bare bones HTTP request protocol.
// Therefore, a message request consists of a request line, headers, and a body.
// **********************************************************************************************************************
package message

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"message/hashmap"
	"message/util"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Request is a structure for parsed data and buffers from reading a client request.
// It implements streaming, buffering, parsing, decoding, and the interface WriteTo.
//
// More specifically, it will (1) read from a reader or from a byte slice and (2) parse the request line, headers, and body.
// Then, it will (3) buffer the request line, headers, and body.
//
// Note, for brevity, the body is buffered in memory. This is not ideal for large requests.
type Request struct {
	// Client
	r io.Reader // source reader provided by the client

	// Request line
	method   string // parsed method
	path     string // parsed path
	protocol string // parsed protocol

	// Headers
	header Header // contains parsed headers

	// Body
	// TODO: body - For small requests, use a buffer in memory. For large requests, use a stream.
	body []byte // buffer for body contents

	// Misc
	// TODO
	url           *url.URL // parsed url
	len           int      // size of message (request line + headers + body)
	size          int      // size of message (request line + headers + body)
	remoteAddress string
}

// NewRequest returns a new Request from a reader or byte slice.
func NewRequest(r io.Reader) *Request {
	// wb := bufio.NewWriter(body)
	// src := io.NopCloser(r) // TODO: Check if this is needed.
	m, err := ReadRequest(r)
	if err != nil {
		panic(err)
	}
	return m
}

func NewRequestFromConn(conn net.Conn) *Request {
	// Read request from conn
	req, err := ReadRequest(conn)
	if err != nil {
		if err != io.EOF {
			panic(err)
		}
	}
	return req
}

// NewRequestFromBytes parses a byte slice into a Request.
func NewRequestFromBytes(data []byte) *Request {
	if len(data) == 0 || data == nil {
		return newRequest() // return empty request
	}
	newBuffer := bytes.NewBuffer(data)   // wrap data in buffer
	reader := bufio.NewReader(newBuffer) // wrap buffer in reader
	return NewRequest(reader)
}

func NewRequestFromClient(method string, url *url.URL, header Header, body []byte) *Request {
	if body == nil {
		body = make([]byte, 0)
	}
	if header == nil {
		header = NewHeader()
	}
	req := newRequest()
	req.method = method
	req.url = url
	req.path = url.Path
	req.header = header
	req.body = body
	return req
}

func newRequest() *Request {
	return &Request{
		// default values
		body:     make([]byte, 0),
		header:   Header(hashmap.New()),
		protocol: "HTTP/1.1",
		method:   "",
		path:     "",
		len:      0,
		url:      nil,
		r:        nil,
	}
}

//######################################################################################################################
// Read/Write
//######################################################################################################################

// WriteTo writes the buffers to w.
func (p *Request) WriteTo(w io.Writer) (int64, error) {
	// Head
	n, err := w.Write(p.Head())
	if err != nil {
		return int64(n), err
	}
	// Body
	if p.body != nil && len(p.body) > 0 {
		n2, err := w.Write(p.body)
		return int64(n2), err
	}
	return int64(n), err
}

// WriteTo writes the buffers to w.
func (p *Request) writeTo(w io.Writer) (int64, error) {
	// Write request line to w
	n, err := w.Write(p.RequestLine())
	if err != nil {
		return int64(n), err
	}
	// Write headers to w
	n2, err := w.Write(p.Headers())
	if err != nil {
		return int64(n + n2), err
	}
	// Write blank line to w to separate headers from body
	n3, err := w.Write([]byte("\r\n"))
	if err != nil {
		return int64(n + n2 + n3), err
	}
	// Write body to w
	n4, err := w.Write(p.body)
	return int64(n + n2 + n3 + n4), err
}

// ToBytes returns the buffers as a byte slice.
func (p *Request) ToBytes() []byte {
	b := bytes.NewBuffer(nil)
	_, err := p.WriteTo(b)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

// ToFile writes the buffers to a file.
func (p *Request) ToFile(path string) (int64, error) {
	// File stream
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	return p.WriteTo(f)
}

// Clone returns a copy of this Request.
func (p *Request) Clone() *Request {
	// Copy
	m := newRequest()
	m.body = p.body
	m.len = p.len
	m.method = p.method
	m.path = p.path
	m.protocol = p.protocol
	m.header = p.header
	return m
}

//######################################################################################################################
// Mutate
//######################################################################################################################

// Reset resets the Request.
func (p *Request) Reset() {
	p = newRequest()
}

// Copy copies a reader to this Request.
func (p *Request) Copy(src io.Reader) {
	// Reset
	p.Reset()
	// Parse
	m, err := parseReaderToRequest(src)
	if err != nil {
		panic(err)
	}
	// Copy
	p.body = m.body
	p.len = m.len
	p.method = m.method
	p.path = m.path
	p.protocol = m.protocol
	p.header = m.header
}

// Merge merges the other Request into this Request.
func (r *Request) Merge(other *Request) {
	// Copy other into this
	if other.body != nil {
		r.body = other.body
	}
	if other.len != 0 {
		r.len = other.len
	}
	if other.method != "" {
		r.method = other.method
	}
	if other.path != "" {
		r.path = other.path
	}
	if other.protocol != "" {
		r.protocol = other.protocol
	}
	if other.header != nil {
		r.header = other.header
	}
}

func (p *Request) SetURL(url *url.URL) {
	p.url = url
}

func (p *Request) SetRemoteAddress(conn net.Conn) {
	p.remoteAddress = conn.RemoteAddr().String()
}

//######################################################################################################################
// Logic
//######################################################################################################################

// Equals returns true if the other Request is equal to this Request.
func (p *Request) Equals(other *Request) error {
	// Check size
	if p.Len() != other.Len() {
		return fmt.Errorf("size mismatch (%d != %d)", p.Len(), other.Len())
	}

	// Request line
	if !bytes.Equal(p.RequestLine(), other.RequestLine()) {
		return fmt.Errorf("request line mismatch (%s != %s)", p.RequestLine(), other.RequestLine())
	}

	// Headers
	// Don't compare the buffers because order doesn't matter.
	// Instead, check if the other map contains the same key-value pairs and size.
	if len(p.header) != len(other.header) {
		return fmt.Errorf("header's size mismatch (%d != %d)", len(p.header), len(other.header))
	}
	for k, v := range p.header {
		if v != other.header[k] {
			return fmt.Errorf("header mismatch for key %s (%s != %s)", k, v, other.header[k])
		}
	}

	if p.method != other.method {
		return fmt.Errorf("method mismatch (%s != %s)", p.method, other.method)
	}

	if p.path != other.path {
		return fmt.Errorf("path mismatch (%s != %s)", p.path, other.path)
	}

	if p.protocol != other.protocol {
		return fmt.Errorf("protocol mismatch (%s != %s)", p.protocol, other.protocol)
	}

	// Check body
	if !bytes.Equal(p.body, other.body) {
		return fmt.Errorf("body mismatch (%s != %s)", p.body, other.body)
	}

	return nil
}

//######################################################################################################################
// Getters
//######################################################################################################################

// String returns a string representation of the Request.
func (p *Request) String() string {
	// TODO: Format Request as a string?
	// lines := []string{
	// // Request line
	// // Headers
	// // Body
	// }
	b := p.ToBytes()
	return string(b)
}

// RequestLine returns the request line of the Request as a byte slice.
func (p *Request) RequestLine() []byte {
	return []byte(fmt.Sprintf("%s %s %s\r\n", p.method, p.path, p.protocol))
}

// Head
func (p *Request) Head() []byte {
	buf := bytes.NewBuffer(p.RequestLine())
	buf.Write(p.Headers())
	buf.WriteString("\r\n")
	buf.WriteString("\r\n")
	return buf.Bytes()
}

// Body returns the body of the Request as a byte slice.
func (p *Request) Body() []byte { return p.body }

// Headers returns the headers of the Request as a byte slice.
func (p *Request) Headers() []byte { return []byte(p.header.ToBytes()) }

// Header returns the headers as a map of the Request.
func (p *Request) Header() map[string]string { return p.header }

// Method returns the method of the Request.
func (p *Request) Method() string { return p.method }

// Path returns the path of the Request.
func (p *Request) Path() string { return p.path }

// Protocol returns the protocol of the Request.
func (p *Request) Protocol() string { return p.protocol }

// Len returns the size of the Request.
func (p *Request) Len() int { return p.len }

func (p *Request) ContentLengthString() string { return strconv.Itoa(p.ContentLength()) }

func (p *Request) ContentLength() int {
	cl, ok := p.header.Get("Content-Length")
	if !ok || cl == "" {
		return 0
	}
	v, err := strconv.Atoi(cl)
	if err != nil {
		return 0
	}
	return v
}

// ContentType returns the header Content-Type of the Request.
func (p *Request) ContentType() string {
	ct, _ := p.header.Get("Content-Type")
	return ct
}

func (p *Request) URL() *url.URL { return p.url }

// ######################################################################################################################
// Helpers
// ######################################################################################################################

// ReadRequest reads a request from a reader.
func ReadRequest(r io.Reader) (*Request, error) {
	return parseReaderToRequest(r)
}

// parseReaderToMessage parses a reader into a Request.
func parseReaderToRequest(r io.Reader) (*Request, error) {
	reader := bufio.NewReader(r) // wrap src reader in bufio.Reader
	pm := newRequest()
	pm.r = r // set src reader

	// TODO: Finish switch type
	switch v := r.(type) {
	case net.Conn:
		pm.SetRemoteAddress(v)
	default:
		//
	}

	// 1.) Request line
	// Note: First line is the request line.
	rl, err := reader.ReadString('\n') // read first line
	if err != nil && err != io.EOF {
		return nil, err
	}
	parts := strings.SplitN(rl, " ", 3) // split first line into method, path, and protocol
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line")
	}
	pm.method = strings.TrimSpace(parts[0])
	pm.path = strings.TrimSpace(parts[1])
	pm.protocol = strings.TrimSpace(parts[2])
	pm.size = len(rl)

	// 2.) Headers
	// Read each new line until a blank line ("\r\n") is reached.
	pm.header = NewHeader()
	for {
		// Read line
		line, err := reader.ReadString('\n') // read line
		// Check error
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
		}
		pm.size += len(line) // add read size
		// Break if blank line of EOF is reached
		if line == "\r\n" || err == io.EOF { // headers are terminated by a blank line "\r\n"
			break
		}
		// parse line
		parts := strings.SplitN(line, ":", 2) // split line into key and value
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid header line")
		}
		// Set header
		pm.header.Set(parts[0], parts[1])
	}
	cl := pm.ContentLength()
	pm.len = len(pm.RequestLine()) + len(pm.Headers()) + cl // set size

	// 3.) Body
	// One more read call to get body contents
	//
	// TODO: Check if size is too big for MaxReadSize and MaxWriteSize
	// Write body to w
	// if p.contentLength > MaxReadSize {
	// 	return int64(n + n2), fmt.Errorf("content length too big")
	// }
	buf := bytes.NewBuffer(make([]byte, 0, cl))
	_, err = util.CopyReaderToWriterN(buf, reader, int64(cl)) // copy reader to writer
	if err != nil {
		if err != io.EOF {
			panic(err)
		}
	}
	pm.size += cl         // add size of body
	pm.body = buf.Bytes() // set body buf
	return pm, nil
}

func parseReqLine(protocol string, method string, path string) (string, error) {
	method = strings.ToUpper(strings.TrimSpace(method))
	path = strings.TrimSpace(path)
	protocol = strings.TrimSpace(protocol)

	// Validate method
	switch method {
	case "GET":
		//
		break
	case "POST":
		//
		break
	case "PUT":
		//
		break
	case "DELETE":
		//
		break
	case "HEAD":
		//
		break
	case "OPTIONS":
		//
		break
	case "TRACE":
		//
		break
	case "CONNECT":
		//
		break
	default:
		return "", fmt.Errorf("Invalid method: %s", method)
	}

	// <method> <path> HTTP/1.1\r\n
	s := fmt.Sprintf("%s %s %s\r\n", method, path, protocol)

	return s, nil
}

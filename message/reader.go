// **********************************************************************************************************************
// Request
// **********************************************************************************************************************
package message

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"message/hashmap"
	"os"
	"strconv"
	"strings"
)

// Request structures implementation of parsing, decoding, and buffering the message from a client.
// The protocol follows a stripped down, bare bones HTTP request.
//
// More specifically, it will (1) read from a reader or from a byte slice and (2) parse the request line, headers, and body.
// Then, it will (3) buffer the request line, headers, and body.
//
// Note, for brevity, the body is buffered in memory. This is not ideal for large requests.
type Request struct {
	// Client
	r io.Reader // source reader provided by the client
	// Parsed data
	method        string // parsed method
	path          string // parsed path
	protocol      string // parsed protocol
	headersMap    Header // contains parsed headers
	contentType   string // parsed header Content-Type
	contentLength int    // parsed header Content-Length
	len           int    // size of message (request line + headers + body)
	// Buffers
	reqLineBuf []byte // buffer for request line
	headersBuf []byte // buffer for headers
	bodyBuf    []byte
}

// NewRequest returns a new Request from a reader or byte slice.
func NewRequest(r io.Reader) *Request {
	// wb := bufio.NewWriter(body)
	// src := io.NopCloser(r)
	m, err := parseReaderToMessage(r)
	if err != nil {
		panic(err)
	}
	return m
}

// parseBytesToMessage parses a byte slice into a Request.
func NewRequestFromBytes(data []byte) *Request {
	if len(data) == 0 || data == nil {
		return newRequest()
	}

	r := bufio.NewReader(bytes.NewBuffer(data))
	return NewRequest(r)
}

func newRequest() *Request {
	return &Request{
		reqLineBuf: make([]byte, 0),
		headersBuf: make([]byte, 0),
		bodyBuf:    make([]byte, 0),
		headersMap: Header(hashmap.New()),
	}
}

// WriteTo writes the buffers to w.
func (p *Request) WriteTo(w io.Writer) (int64, error) {
	// Write request line to w
	n, err := w.Write(p.reqLineBuf)
	if err != nil {
		return int64(n), err
	}
	// Write headers to w
	n2, err := w.Write(p.headersBuf)
	if err != nil {
		return int64(n + n2), err
	}
	// Write body to w
	n3, err := w.Write(p.bodyBuf)
	return int64(n + n2 + n3), err
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

// Equals returns true if the other Request is equal to this Request.
func (p *Request) Equals(other *Request) error {
	// Check size
	if p.Len() != other.Len() {
		return fmt.Errorf("size mismatch (%d != %d)", p.Len(), other.Len())
	}

	// Request line
	if !bytes.Equal(p.reqLineBuf, other.reqLineBuf) {
		return fmt.Errorf("request line mismatch (%s != %s)", p.reqLineBuf, other.reqLineBuf)
	}

	// Headers
	// Don't compare the buffers because order doesn't matter.
	// Instead, check if the other map contains the same key-value pairs and size.
	if len(p.headersMap) != len(other.headersMap) {
		return fmt.Errorf("header's size mismatch (%d != %d)", len(p.headersMap), len(other.headersMap))
	}
	for k, v := range p.headersMap {
		if v != other.headersMap[k] {
			return fmt.Errorf("header mismatch for key %s (%s != %s)", k, v, other.headersMap[k])
		}
	}

	if p.contentType != other.contentType {
		return fmt.Errorf("content type mismatch (%s != %s)", p.contentType, other.contentType)
	}

	if p.contentLength != other.contentLength {
		return fmt.Errorf("content length mismatch (%d != %d)", p.contentLength, other.contentLength)
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
	if !bytes.Equal(p.bodyBuf, other.bodyBuf) {
		return fmt.Errorf("body mismatch (%s != %s)", p.bodyBuf, other.bodyBuf)
	}

	return nil
}

func (p *Request) Reset() {
	p = newRequest()
}

// Copy copies the source to this Request.
func (p *Request) Copy(src io.Reader) {
	// Reset
	p.Reset()
	// Parse
	m, err := parseReaderToMessage(src)
	if err != nil {
		panic(err)
	}
	// Copy
	p.reqLineBuf = m.reqLineBuf
	p.headersBuf = m.headersBuf
	p.bodyBuf = m.bodyBuf
	p.len = m.len
	p.method = m.method
	p.path = m.path
	p.protocol = m.protocol
	p.headersMap = m.headersMap
	p.contentType = m.contentType
	p.contentLength = m.contentLength
}

func (p *Request) Clone() *Request {
	// Copy
	m := newRequest()
	m.reqLineBuf = p.reqLineBuf
	m.headersBuf = p.headersBuf
	m.bodyBuf = p.bodyBuf
	m.len = p.len
	m.method = p.method
	m.path = p.path
	m.protocol = p.protocol
	m.headersMap = p.headersMap
	m.contentType = p.contentType
	m.contentLength = p.contentLength
	return m
}

//######################################################################################################################
// Getters
//######################################################################################################################

// String returns a string representation of the Request.
func (p *Request) String() string {
	// lines := []string{
	// 	fmt.Sprintf("Request"),
	// 	fmt.Sprintf("\t- Method: %s", p.method),
	// 	fmt.Sprintf("\t- Path: %s", p.path),
	// 	fmt.Sprintf("\t- Protocol: %s", p.protocol),
	// 	fmt.Sprintf("\t- RequestLine: %d", p.reqLineBuf),
	// 	fmt.Sprintf("\t- Headers: %d", p.headers),
	// 	fmt.Sprintf("\t- HeadersMap: %s", p.headersMap),
	// 	fmt.Sprintf("\t- Body: %p", p.body),
	// 	fmt.Sprintf("\t- ContentLength: %d", p.contentLength),
	// 	fmt.Sprintf("\t- ContentType: %s", p.contentType),
	// }
	b := p.ToBytes()
	return string(b)
}

// Len returns the size of the Request.
func (p *Request) Len() int { return p.len }

// RequestLine returns the request line of the Request.
func (p *Request) RequestLine() []byte { return p.reqLineBuf }

// Headers returns the headers of the Request.
func (p *Request) Headers() []byte { return p.headersBuf }

// HeadersMap returns the headers as a map of the Request.
func (p *Request) HeadersMap() map[string]string { return p.headersMap }

// Body returns the body of the Request.
func (p *Request) Body() []byte { return p.bodyBuf }

// Method returns the method of the Request.
func (p *Request) Method() string { return p.method }

// Path returns the path of the Request.
func (p *Request) Path() string { return p.path }

// Protocol returns the protocol of the Request.
func (p *Request) Protocol() string { return p.protocol }

// ContentType returns the header Content-Type of the Request.
func (p *Request) ContentType() string { return p.contentType }

// ContentLength returns the header Content-Length of the Request.
func (p *Request) ContentLength() int { return p.contentLength }

// ######################################################################################################################
// Helpers
// ######################################################################################################################

// ReadRequest reads a request from a reader.
func ReadRequest(r io.Reader) (*Request, error) {
	return parseReaderToMessage(r)
}

// parseReaderToMessage parses a reader into a Request.
func parseReaderToMessage(r io.Reader) (*Request, error) {
	reader := bufio.NewReader(r) // wrap src reader in bufio.Reader

	pm := new(Request)
	pm.r = r // set src reader

	// 1.) Request line
	rl, err := reader.ReadString('\n') // parse first line from reader as the request line
	if err != nil && err != io.EOF {
		return nil, err
	}
	parts := strings.SplitN(rl, " ", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line")
	}
	pm.reqLineBuf = []byte(rl)
	pm.method = strings.TrimSpace(parts[0])
	pm.path = strings.TrimSpace(parts[1])
	pm.protocol = strings.TrimSpace(parts[2])

	// 2.) Headers
	headersBytes := bytes.NewBuffer(nil)
	m := Header(hashmap.New())
	for {
		line, err := reader.ReadString('\n') // read line
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		headersBytes.Write([]byte(line)) // write line to buffer
		if line == "\r\n" {              // headers are terminated by a blank line "\r\n"
			break
		}
		parts := strings.SplitN(line, ":", 2) // split line into key and value
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid header line")
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		m.Set(key, value)
	}
	pm.headersMap = m
	pm.headersBuf = headersBytes.Bytes()
	cl, ok := pm.headersMap["Content-Length"]
	if !ok {
		cl = "0"
	}
	length, err := strconv.Atoi(cl) // convert to int
	if err != nil {
		length = 0
	}
	pm.contentLength = length // set Content-Length
	ct, ok := pm.headersMap["Content-Type"]
	if !ok {
		ct = ""
	}
	pm.contentType = ct // set Content-Type

	pm.len = len(pm.reqLineBuf) + len(pm.headersBuf) + pm.contentLength // set size

	// 3.) Body
	// One more read call to get body contents
	//
	// TODO: Check if size is too big for MaxReadSize and MaxWriteSize
	// Write body to w
	// if p.contentLength > MaxReadSize {
	// 	return int64(n + n2), fmt.Errorf("content length too big")
	// }
	buf := bytes.NewBuffer(make([]byte, 0, pm.contentLength))
	_, err = copyN(buf, reader, int64(pm.contentLength)) // copy reader to writer
	if err != nil {
		panic(err)
	}
	pm.bodyBuf = buf.Bytes()

	return pm, nil
}

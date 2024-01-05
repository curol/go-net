// **********************************************************************************************************************
// Request
//
// Request is a reader, buffer, and parser for a client requests.
// It handles reading a client request.
//
// For brevity, the protocol for a message request follows a stripped down, bare bones HTTP request protocol.
// Therefore, a message request consists of a request line, headers, and a body.
// **********************************************************************************************************************
package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	gonet "net"
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
	// Request line
	method   string   // parsed method
	path     string   // parsed path
	protocol string   // parsed protocol
	url      *url.URL // parsed url

	// Headers
	header Header // contains parsed headers

	// Body
	body          io.ReadCloser // stream for body contents which allows reading and closing connection
	contentLength int64
	contentType   string

	// Misc
	remoteAddress string     // address of client
	host          string     // host address
	form          url.Values // parsed form
	multipartForm *multipart.Form
	bytesRead     int64 // total bytes read
}

// NewRequest
func NewRequest(method string, address string, headers map[string]string, body io.Reader) *Request {
	// Create
	req := newRequest(
		method,
		address,
		headers,
		io.NopCloser(body),
	)

	return req
}

func newRequest(method string, address string, header map[string]string, body io.ReadCloser) *Request {
	// Default values for request instance
	req := newDefaultRequest()

	// Set Request line
	req.SetMethod(method)
	err := req.setURLFromAddress(address)
	if err != nil {
		panic(err)
	}

	// Set headers
	req.header.FromMap(header)
	req.SetContentLength(getContentLength(req.header))
	req.SetContentType(getContentType(req.header))

	// Set body
	req.SetBody(body)

	return req
}

func newDefaultRequest() *Request {
	return &Request{
		header:        NewHeader(),
		protocol:      "HTTP/1.1",
		url:           nil,
		body:          nil,
		method:        "",
		path:          "",
		remoteAddress: "",
		host:          "",
	}
}

// Reset resets the Request.
func (p *Request) Reset() {
	p = newDefaultRequest()
}

// Copy copies a reader to this Request.
func (p *Request) Copy(src io.Reader) {
	// Reset
	p.Reset()
	// Parse
	// other := newRequestParser(src).req
	other := ReadRequest(src)
	// Copy
	p.body = other.body
	p.method = other.method
	p.path = other.path
	p.protocol = other.protocol
	p.header = other.header
	p.contentLength = other.contentLength
	p.contentType = other.contentType
	p.form = other.form
	p.bytesRead = other.bytesRead
	p.header = other.header
	p.host = other.host
	p.multipartForm = other.multipartForm
	p.remoteAddress = other.remoteAddress
	p.url = other.url
}

// Merge merges the other Request into this Request.
func (r *Request) Merge(other *Request) {
	// Copy other into this
	if other.body != nil {
		r.body = other.body
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
	if other.contentLength != 0 {
		r.contentLength = other.contentLength
	}
	if other.contentType != "" {
		r.contentType = other.contentType
	}
	if other.form != nil {
		r.form = other.form
	}
	if other.bytesRead != 0 {
		r.bytesRead = other.bytesRead
	}
	if other.header != nil {
		r.header = other.header
	}
	if other.host != "" {
		r.host = other.host
	}
	if other.multipartForm != nil {
		r.multipartForm = other.multipartForm
	}
	if other.remoteAddress != "" {
		r.remoteAddress = other.remoteAddress
	}
	if other.url != nil {
		r.url = other.url
	}
}

// Equals returns true if the other Request is equal to this Request.
func (p *Request) Equals(other *Request) error {
	if p.header.Len() != other.header.Len() {
		return fmt.Errorf("header's size mismatch (%d != %d)", len(p.header), len(other.header))
	}

	if !p.header.Equals(other.header) {
		return fmt.Errorf("header mismatch (%s != %s)", p.header, other.header)
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

	if p.contentLength != other.contentLength {
		return fmt.Errorf("content length mismatch (%d != %d)", p.contentLength, other.contentLength)
	}

	if p.contentType != other.contentType {
		return fmt.Errorf("content type mismatch (%s != %s)", p.contentType, other.contentType)
	}

	if p.host != other.host {
		return fmt.Errorf("host mismatch (%s != %s)", p.host, other.host)
	}

	if p.remoteAddress != other.remoteAddress {
		return fmt.Errorf("remote address mismatch (%s != %s)", p.remoteAddress, other.remoteAddress)
	}

	if p.url != nil && other.url != nil {
		if p.url.String() != other.url.String() {
			return fmt.Errorf("url mismatch (%s != %s)", p.url, other.url)
		}
	}

	if p.url == nil && other.url != nil || p.url != nil && other.url == nil {
		return fmt.Errorf("url mismatch (%s != %s)", p.url, other.url)
	}

	if p.form == nil && other.form != nil || p.form != nil && other.form == nil {
		return fmt.Errorf("form mismatch (%s != %s)", p.form, other.form)
	}

	if p.form.Encode() != other.form.Encode() {
		return fmt.Errorf("form mismatch (%s != %s)", p.form.Encode(), other.form.Encode())
	}

	if p.body == nil && other.body != nil || p.body != nil && other.body == nil {
		return fmt.Errorf("body mismatch (%s != %s)", p.body, other.body)
	}

	// TODO: Check body
	// if !bytes.Equal(p.body, other.body) {
	// 	return fmt.Errorf("body mismatch (%s != %s)", p.body, other.body)
	// }

	return nil
}

// Clone returns a copy of this Request.
func (p *Request) Clone() *Request {
	// Copy
	r := newRequest(p.method, p.url.String(), p.header.Clone(), p.body)
	r.host = p.host
	r.remoteAddress = p.remoteAddress
	r.form = p.form
	r.multipartForm = p.multipartForm
	return r
}

//######################################################################################################################
// Serialize
//######################################################################################################################

// Serialize serializes the request.
//
// Note:
//   - if head is nil, then only the body will be serialized.
//   - if body is nil, then only the head will be serialized.
func (p *Request) serialize(head io.Writer, body io.Writer) (int64, error) {
	count := int64(0)

	// Head
	if head != nil {
		// Write request line
		n, err := serializeRequestLine(head, p.method, p.path, p.protocol)
		if err != nil {
			return int64(n), err
		}
		count += int64(n)

		//  Write headers
		n, err = serializeHeader(head, p.header)
		if err != nil {
			return count, err
		}
		count += int64(n)
	}

	// Body
	if body != nil {
		n, err := p.ReadBody(body)
		if err != nil {
			return count, err
		}
		count += n
		return count, err
	}

	return count, nil
}

//######################################################################################################################
// Write
//######################################################################################################################

// WriteTo writes the buffers to w.
func (p *Request) WriteTo(w io.Writer) (int64, error) {
	return p.serialize(w, w)
}

// ReadBody reads the body of the Request.
func (r *Request) ReadBody(w io.Writer) (int64, error) {
	return transportBodyToWriter(r.body, w, r.contentLength)
}

//######################################################################################################################
// Encode
//######################################################################################################################

// ToBuffer encodes the entire request into a *bytes.Buffer.
func (p *Request) ToBuffer() (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	_, err := p.WriteTo(buf)
	return buf, err
}

// ToBytes returns the buffers as a byte slice.
func (p *Request) ToBytes() ([]byte, error) {
	buf, err := p.ToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *Request) ToString() (string, error) {
	buf, err := p.ToBuffer()
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ToFile writes the buffers to a file.
func (p *Request) ToFile(fn string) (int64, error) {
	// File stream
	f, err := os.Create(fn)
	defer f.Close()
	if err != nil {
		return 0, err
	}
	return p.WriteTo(f)
}

// BodyBuffer returns the body of the Request as a *bytes.Buffer.
func (r *Request) BodyBuffer() (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	_, err := r.ReadBody(buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// BodyString returns the body of the Request as a string.
func (r *Request) BodyString() (string, error) {
	buf, err := r.BodyBuffer()
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

//######################################################################################################################
// Mutate
//######################################################################################################################

/*
	Body
*/

// SetBody sets the body of the Request.
func (r *Request) SetBody(body io.ReadCloser) {
	// TODO: switch case for body
	// switch v := body.(type) {
	// case io.ReadCloser:
	// 	req.body = v
	// 	req.closer = true
	// default:
	// 	req.body = newCloser(v)
	// 	req.closer = false
	// }
	r.body = body
}

func (r *Request) SetContentLength(length int64) {
	r.contentLength = length
}

func (r *Request) SetContentType(contentType string) {
	r.contentType = contentType
}

func (r *Request) SetForm(form url.Values) {
	r.form = form
}

func (r *Request) SetMultipartForm(form *multipart.Form) {
	r.multipartForm = form
}

/*
	Request Line
*/

func (r *Request) SetRequestLine(method string, path string) {
	r.SetMethod(method)
	r.SetPath(path)
}

func (r *Request) SetMethod(method string) {
	r.method = strings.ToUpper(strings.TrimSpace(method))
}

func (r *Request) SetPath(path string) {
	r.path = strings.TrimSpace(path)
}

func (r *Request) SetProtocol(protocol string) {
	r.protocol = protocol
}

func (r *Request) SetHost(host string) {
	r.host = host
}

/*
	Header
*/

func (r *Request) SetHeader(header Header) {
	r.header = header
}

func (r *Request) SetHeaderFromMap(header map[string]string) {
	if r.header == nil || len(r.header) == 0 {
		return
	}
	// Mutate header from map of strings
	r.header.FromMap(header)
}

/*
	URL
*/

func (r *Request) SetURL(url *url.URL) {
	r.url = url
}

func (r *Request) setURLFromAddress(address string) error {
	url, err := url.Parse(address)
	if err != nil {
		return err
	}
	r.url = url
	r.path = url.Path
	return nil
}

/*
	Connection
*/

func (p *Request) SetRemoteAddress(s string) {
	p.remoteAddress = s
}

func (r *Request) SetBytesRead(n int64) {
	r.bytesRead = n
}

// ######################################################################################################################
// Getters
// ######################################################################################################################
func (r *Request) Form() url.Values {
	return r.form
}

func (r *Request) MultipartForm() *multipart.Form {
	return r.multipartForm
}

// Body returns the body of the Request as a byte slice.
func (r *Request) Body() io.ReadCloser {
	return r.body
}

// ContentType returns the content type of the Request.
func (p *Request) ContentType() string {
	return p.contentType
}

// ContentLength returns the content length of the Request.
func (p *Request) ContentLength() int64 { return p.contentLength }

// ContentLengthString returns the content length as a string.
func (p *Request) ContentLengthString() string { return strconv.Itoa(int(p.contentLength)) }

func (r *Request) RequestLine() string {
	buf := bytes.NewBuffer(nil)
	serializeRequestLine(buf, r.method, r.path, r.protocol)
	return buf.String()
}

// String returns a string representation of the Request.
func (p *Request) String() string {
	s, _ := p.ToString()
	return s
}

// Header returns the headers as a map of the Request.
func (p *Request) Header() map[string]string { return p.header }

// Method returns the method of the Request.
func (p *Request) Method() string { return p.method }

// Path returns the path of the Request.
func (p *Request) Path() string { return p.path }

// Protocol returns the protocol of the Request.
func (p *Request) Protocol() string { return p.protocol }

func (p *Request) URL() *url.URL { return p.url }

// ######################################################################################################################
// Helpers
// ######################################################################################################################

// ReadRequest reads a request from a reader.
//
// Examples:
//
//	1.) Raw request bytes:
//	 	```
//		data := []byte("GET / HTTP/1.1\r\nHost: localhost:8080\r\n\r\n")
//	 	newBuffer := bytes.NewBuffer(data)   // wrap data in buffer
//		reader := bufio.NewReader(newBuffer) // wrap buffer in reader
//		req, _ := ReadRequest(reader)
//	 	```
func ReadRequest(r io.Reader) *Request {
	req, n, err := parseRequest(r)
	if err != nil && err != io.EOF {
		panic(err)
	}
	req.SetBytesRead(n)
	return req
}

// parseReaderToMessage parses a reader into a Request.
func parseRequest(r io.Reader) (*Request, int64, error) {
	if r == nil {
		return newDefaultRequest(), 0, nil
	}
	reader := bufio.NewReader(r)
	req := newDefaultRequest()

	// 1.) Request line
	rl, err := parseRequestLine(reader)
	if err != nil {
		return nil, 0, fmt.Errorf("Error parsing request line: %s", err)
	}
	req.SetMethod(rl.method)
	req.SetPath(rl.path)

	// 2.) Headers
	ph, err := parseHeaders(reader)
	if err != nil {
		return nil, 0, fmt.Errorf("Error parsing request headers: %s", err)
	}
	req.SetHeader(ph.header)

	/// 3.) Body
	pb, err := parseBody(req.header, reader)
	req.SetBody(pb.body)
	req.SetContentLength(pb.len)
	req.SetContentType(pb.typ)

	// 4.) Check type of r and arrange accordingly
	switch v := r.(type) {
	case gonet.Conn:
		pc, err := parseConnection(v)
		if err != nil {
			return nil, 0, fmt.Errorf("Error parsing request connection: %s", err)
		}
		req.SetRemoteAddress(pc.remoteAddress)
		req.SetURL(pc.url)
		req.SetHost(pc.host)
	default:
		// TODO: finished default case
	}

	// 5. Total bytes read
	n := (int64(rl.len + ph.len))

	return req, n, nil
}

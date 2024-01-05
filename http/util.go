package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func newBuffer(b []byte) *bytes.Buffer {
	return bytes.NewBuffer(b)
}

func newWriter(w io.Writer) *bufio.Writer {
	return bufio.NewWriter(w)
}

func newReader(r io.Reader) *bufio.Reader {
	// return bytes.NewReader(nil)
	return bufio.NewReader(r)
}

func newNopCloser(r io.Reader) io.ReadCloser {
	return io.NopCloser(r)
}

func copyReaderToWriter(r io.Reader, w io.Writer, len int64) (int64, error) {
	if len == 0 {
		return 0, fmt.Errorf("can't copy reader of size 0")
	}
	if w == nil {
		return 0, fmt.Errorf("can't copy writer of type nil")
	}
	if r == nil {
		return 0, fmt.Errorf("can't copy reader of type nil")
	}

	// Copy reader to w
	return io.CopyN(w, r, len) // copy reader to writer of size cl
}

func getContentLength(header Header) int64 {
	// v := getContentLength(r.header)
	// // Content length
	// cl, ok := header.Get("Content-Length")
	// if !ok {
	// 	return 0, "", fmt.Errorf("Content-Length header not found")
	// }
	// if cl == "" {
	// 	return 0, "", fmt.Errorf("Content-Length header is empty")
	// }
	// contentLen, err := strconv.ParseInt(cl, 10, 64) // convert to int64
	// if err != nil {
	// 	return 0, "", err
	// }
	if header == nil {
		return 0
	}
	cl, ok := header.Get("Content-Length")
	if !ok || cl == "" {
		return 0
	}
	v, err := strconv.Atoi(cl)
	if err != nil {
		return 0
	}
	return int64(v)
}

func getContentType(header Header) string {
	ct, ok := header.Get("Content-Type")
	if !ok || ct == "" {
		return ""
	}
	return ct
}

//**********************************************************************************************************************
// Serialize
//**********************************************************************************************************************

// SerializeRequestLine writes the request line to w.
func serializeRequestLine(w io.Writer, method, path, protocol string) (int, error) {
	// <method> <path> <protocol>\r\n
	b := []byte(fmt.Sprintf("%s %s %s\r\n", method, path, protocol))
	return w.Write(b)
}

// SerializeResponseLine writes the response line to w.
func serializeResponseLine(w io.Writer, protocol, status, message string) (int, error) {
	// <protocol> <status> <message>\r\n
	b := []byte(fmt.Sprintf("%s %s %s\r\n", protocol, status, message))
	return w.Write(b)
}

// SerializeHeader writes the headers to w.
func serializeHeader(w io.Writer, header Header) (int, error) {
	n, err := w.Write(header.ToBytes("\r\n"))
	if err != nil {
		return n, err
	}
	// Add blank line to header to signal end of header
	n2, err := w.Write([]byte("\r\n")) // last line in header is a blank line to separate head from body
	return n + n2, err
}

// SerializeBody writes the body to w.
func serializeBody(w io.Writer, body io.Reader, contentLen int64) (int64, error) {
	if body == nil {
		return 0, fmt.Errorf("body can't be nil")
	}
	if w == nil {
		return 0, fmt.Errorf("writer can't be nil")
	}
	if contentLen <= 0 {
		return 0, fmt.Errorf("Request.contentLength can't be <= 0")
	}
	return io.CopyN(w, body, contentLen)
}

//**********************************************************************************************************************
// Parse
//**********************************************************************************************************************

type parsedHeaders struct {
	header Header
	len    int // bytes read
}

// ParseHeaders parses the reader for the headers.
// It reads each new line from the reader until there is a blank line or io.EOF.
func parseHeaders(r *bufio.Reader) (*parsedHeaders, error) {
	n := 0                // counter for bytes read
	header := NewHeader() // header

	// Read each new line until a blank line ("\r\n") is reached.
	for {
		// Read line
		line, err := r.ReadString('\n') // read line
		// Check error
		if err != nil && err != io.EOF {
			return nil, err
		}
		// Set # of bytes read
		n += len(line)
		// Finish if blank line of EOF is reached
		if line == "\r\n" || err == io.EOF {
			// Headers are terminated by a blank line "\r\n"
			break
		}
		// Parse line
		parts := strings.SplitN(line, ":", 2) // split line into key and value
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid header line")
		}
		// Set header
		header.Set(parts[0], parts[1])
	}

	return &parsedHeaders{header: header, len: n}, nil
}

type parsedRequestLine struct {
	method   string
	path     string
	protocol string
	len      int // bytes read
}

// ParseRequestLine parses the first line from a reader.
// It reads the first line from the reader and parses it into a request line.
func parseRequestLine(r *bufio.Reader) (*parsedRequestLine, error) {
	// Read
	rl, err := r.ReadString('\n') // read first line
	if err != nil && err != io.EOF {
		return nil, err
	}
	// Parse
	parts := strings.SplitN(rl, " ", 3) // split first line into method, path, and protocol
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line")
	}
	// Set
	method := strings.ToUpper(strings.TrimSpace(parts[0]))
	path := strings.TrimSpace(parts[1])
	protocol := strings.TrimSpace(parts[2])

	reqL := &parsedRequestLine{method: method, path: path, protocol: protocol, len: len(rl)}

	return reqL, nil
}

type parsedResponseLine struct {
	Version      string
	StatusCode   int
	ReasonPhrase string
	len          int
}

func parseResponseLine(r *bufio.Reader) (*parsedResponseLine, error) {
	// Read
	line, err := r.ReadString('\n') // read first line
	if err != nil && err != io.EOF {
		return nil, err
	}

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid response line: %s", line)
	}

	statusCode, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid status code: %s", parts[1])
	}

	return &parsedResponseLine{
		Version:      parts[0],
		StatusCode:   statusCode,
		ReasonPhrase: parts[2],
		len:          len(line),
	}, nil
}

type parsedBody struct {
	len  int64
	typ  string
	body io.ReadCloser
}

// ParseBody parses the body from the reader.
//
// Note:
//   - If bodyWriter is nil, then r is not written to the bodyWriter.
func parseBody(header Header, r io.Reader) (*parsedBody, error) {
	if header == nil {
		return nil, fmt.Errorf("header can't be nil")
	}

	pb := new(parsedBody)
	pb.len = getContentLength(header) // Content length
	pb.typ = getContentType(header)   // Content type
	pb.body = io.NopCloser(r)

	return pb, nil
}

type parsedConnection struct {
	remoteAddress string
	localAddress  string
	url           *url.URL
	host          string
	hostname      string
	path          string
}

func parseConnection(conn net.Conn) (*parsedConnection, error) {
	pc := new(parsedConnection)

	pc.remoteAddress = conn.RemoteAddr().String()
	pc.localAddress = conn.LocalAddr().String()
	u, err := url.Parse(pc.remoteAddress)
	if err != nil {
		return nil, err
	}
	pc.url = u
	pc.hostname = u.Hostname()
	pc.host = u.Host
	return pc, nil
}

//**********************************************************************************************************************
// Transport
//**********************************************************************************************************************

func transportBodyToWriter(bodyReader io.Reader, bodyWriter io.Writer, contentLen int64) (int64, error) {
	if bodyReader == nil {
		return 0, fmt.Errorf("transport body can't be nil")
	}
	if bodyWriter == nil {
		return 0, fmt.Errorf("transport writer can't be nil")
	}
	if contentLen <= 0 {
		return 0, fmt.Errorf("transport contentLen can't be <= 0")
	}
	return io.CopyN(bodyWriter, bodyReader, contentLen)
}

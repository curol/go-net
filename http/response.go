package http

//********************************************************************************************************************//
// Response
//
// Response represents an HTTP response received by a client,
// It structures writing a client response.
//
// The response is seperated into three parts: (1) response line, (2) header, and (3) body.
// Also, you can seperate it into two parts: (1) head and (2) body.
// Where the head equals response line + headers, and the body equals the contents of the writer.
// The size of the response equals: len(responseLine) + len(header.S) + len(body)
//
//  - 1. Response Line = Lines[0]
// 		- First line which contains the protocol, status code, and status text.
// 		- Format: "<protocol> <status code> <status text>"
// 	- 2. Header = Lines[1,...,N,N+1]
// 		- Lines[1,...,N] = Lines 1...N are sequential for the header's "<key>: <value>" pairs.
// 		- `N` = Last line for the header's "<key>: <value>" pairs.
// 		- `N+1` = One line after line `N` that is a blank line ("\r\n") that seperates the head and body.
// 		- Format: "<key>: <value>".
// 		- Size: header.size
// 		- Order doesn't matter.
// 	- 3. Body = Lines[N+2:header["Content-Length"]]
// 		- Optional, if header [Content-Length] exists, then there is a body for the response.
// 		- `N+2` = Two lines after line `N`, which is the starting line for the body.
// 		- Size = header["Content-Length"]
// 		- Format = header["Content-Type"]
// 		- Can contain multiple lines of data, and the number of lines depends on the type of content (i.g., JSON, Text, HTML, XML, ...) of the response.
//
// Note:
//  - A line is a sequence of zero or more characters followed by a line feed ("\n"), a carriage return ("\r"), or a carriage return followed immediately by a line feed ("\r\n").
//  - A blank line is a line that contains only a line feed, a carriage return, or a carriage return followed immediately by a line feed.
// 	- Size of response equals the len(head) + len(body).
//  - Response = Head + Body
//  - Head = Response line + Header lines + Blank line
//  - [N+2:size] = Body of response
//  - Head = Response line + Header
//  - You can also break the response into two parts: head + body
//
// For example:
// ```
// <response line> <header lines> <blank line> <response body>
// ```
//
// ## Response line
//
// The response line in an HTTP response contains the status code, status text, and HTTP version.
// Here's an example of a response line:
//
// ```
// HTTP/1.1 200 OK
// ```
//
// In this example, the HTTP version is "HTTP/1.1", the status code is "200", and the status text is "OK".
// The status code indicates the result of the request, and the status text provides a human-readable description of the status code.
//
// Note, that the response line is the first line of an HTTP response.
//
// ## Headers
//
// The headers in an HTTP response are similar to the headers in an HTTP request.
//
// ## Body
//
// The body of an HTTP response is the data that the server sends back to the client
//********************************************************************************************************************//

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Response represents a bare bones HTTP response.
//
// Note, there are 3 general parts of a response:
// (1) Response line, (2) Headers, and (3) Body.
// Also, you can seperate it into two parts: (1) head and (2) body.
type Response struct {
	// 1.) Response line is the first line of a response.
	// Format: <protocol> <statusCode> <statusText>
	//
	// E.g.,
	// 	"HTTP/1.0 200 OK"
	//
	// 	Status     string // e.g. "200 OK"
	// 	StatusCode int    // e.g. 200
	// 	Proto      string // e.g. "HTTP/1.0"
	// 	ProtoMajor int    // e.g. 1
	// 	ProtoMinor int    // e.g. 0
	protocol   string // e.g., "HTTP/1.1"
	statusCode int    // e.g., 200
	statusText string // e.g., "OK"

	// 2.) Headers come after the response line and each header is a new line of format "<key>: <value>".
	header        Header
	contentLength int
	contentType   string

	// 3.) Body is the payload or contents of the response.
	// For brevity, only using []byte for body.
	body io.ReadCloser

	// Request is the request that was sent to obtain this Response.
	// Request's Body is nil (having already been consumed).
	// This is only populated for Client requests.
	Request *Request

	// Close records whether the header directed that the connection be
	// closed after reading Body. The value is advice for clients: neither
	// ReadResponse nor Response.Write ever closes a connection.
	close bool
}

func newDefaultResponse() *Response {
	return &Response{
		protocol:   "HTTP/1.1", // default protocol
		statusCode: 200,        // default status code
		statusText: "OK",       // default status text
		header:     NewHeader(),
		body:       nil,
	}
}

// Close closes the connection and writes io.EOF to the connection.
func (r *Response) Close() error {
	if r.close {
		return fmt.Errorf("response already closed")
	}
	r.close = true
	return nil
}

// WriteHeader writes the header `k` and `v` to the response.
func (r *Response) WriteHeader(k string, v string) {
	r.header.Set(k, v)
}

func (r *Response) Write(w io.Writer) (int64, error) {
	bw := bufio.NewWriter(w)
	// TODO: Close body?
	return r.write(bw)
}

// Write writes response output to w.
func (r *Response) write(w *bufio.Writer) (int64, error) {
	// TODO: Transfer in chunks?
	err := r.serialize(w) // serialize response
	if err != nil {
		return 0, err
	}
	// n, err := w.Write(buf.Bytes()) // write output to w
	return int64(w.Size()), err
}

// Serialize serializes the response to w.
func (r *Response) serialize(w *bufio.Writer) error {
	// 1. Response line
	rl := fmt.Sprintf("%s %s %s\r\n", r.protocol, strconv.Itoa(r.statusCode), r.statusText)
	_, err := w.Write([]byte(rl))
	if err != nil {
		return err
	}

	// 2. Header
	r.header.Write(w)
	if err != nil {
		return err
	}

	// 3. End of head
	w.Write([]byte("\r\n")) // 3. end of head

	// 4. Flush head
	w.Flush()

	// 5. Body
	if r.body != nil {
		contLen := int64(r.contentLength)
		_, err = io.CopyN(w, r.body, contLen) // copy body to writer
		if err != nil {
			return err
		}
		err = w.Flush() // flush the writer to the client connection
		if err != nil {
			return err
		}
	}

	return nil
}

// ToBuffer encodes the request into a *bytes.Buffer.
func (r *Response) ToBuffer() (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	_, err := r.Write(buf)
	return buf, err
}

// Equals returns true if the other Request is equal to this Request.
// func (p *Response) Equals(other *Response) error {
// 	if p.protocol != other.protocol {
// 		return fmt.Errorf("protocol mismatch (%s != %s)", p.protocol, other.protocol)
// 	}
// 	if p.statusCode != other.statusCode {
// 		return fmt.Errorf("status code mismatch (%d != %d)", p.statusCode, other.statusCode)
// 	}
// 	if p.statusText != other.statusText {
// 		return fmt.Errorf("status text mismatch (%s != %s)", p.statusText, other.statusText)
// 	}
// 	if p.contentLength != other.contentLength {
// 		return fmt.Errorf("content length mismatch (%d != %d)", p.contentLength, other.contentLength)
// 	}
// 	if p.contentType != other.contentType {
// 		return fmt.Errorf("content type mismatch (%d != %d)", p.contentLength, other.contentLength)
// 	}
// 	if !p.close && !other.close {
// 		return fmt.Errorf("close mismatch (%v != %v)", p.close, other.close)
// 	}
// 	if p.Request.Equals(other.Request) != nil {
// 		return fmt.Errorf("request mismatch (%v != %v)", p.Request, other.Request)
// 	}
// 	return nil
// }

//********************************************************************************************************************//
// Output
//********************************************************************************************************************//

// Text writes the string `s` to the response body and sets the content to `text/plain`.
func (r *Response) Text(s string) {
	ct := "text/plain"
	cl := strconv.Itoa(len(s))
	r.header.Set("Content-Type", ct)
	r.header.Set("Content-Length", cl)

	r.body = io.NopCloser(bytes.NewBufferString(s))
}

// JSON writes the JSON `v` to the response body and sets the content to `application/json`.
func (r *Response) JSON(v any) error {
	result, err := json.Marshal(v)
	if err != nil {
		return err
	}
	r.header.Set("Content-Type", "application/json")
	r.header.Set("Content-Length", strconv.Itoa(len(result)))

	r.body = io.NopCloser(bytes.NewBuffer(result))

	return nil
}

func (r *Response) HTML(s string) {
	ct := "text/html"
	cl := strconv.Itoa(len(s))
	r.header.Set("Content-Type", ct)
	r.header.Set("Content-Length", cl)

	r.body = io.NopCloser(bytes.NewBufferString(s))
}

func (r *Response) XML(s string) {
	ct := "text/xml"
	cl := strconv.Itoa(len(s))
	r.header.Set("Content-Type", ct)
	r.header.Set("Content-Length", cl)

	r.body = io.NopCloser(bytes.NewBufferString(s))
}

func (r *Response) JSONP(s string) {
	ct := "application/javascript"
	cl := strconv.Itoa(len(s))
	r.header.Set("Content-Type", ct)
	r.header.Set("Content-Length", cl)

	r.body = io.NopCloser(bytes.NewBufferString(s))
}

func (r *Response) File(s string) {
	ct := "application/octet-stream"
	f, err := os.Open(s)
	if err != nil {
		panic(err)
	}
	stat, err := f.Stat()
	if err != nil {
		panic(err)
	}
	cl := strconv.FormatInt(stat.Size(), 10) // Convert cl to a string
	r.header.Set("Content-Type", ct)
	r.header.Set("Content-Length", cl)

	// TODO: Use io.NopCloser()?
	// r.body = io.NopCloser(f)
	r.body = f
}

//######################################################################################################################
// Status
//######################################################################################################################

// Ok indicates that the request is successful.
func (r *Response) Ok() {
	r.statusCode = 200
	r.statusText = "OK"
}

// BadRequest indicates that the request could not be understood by the server due to malformed syntax.
func (r *Response) BadRequest() {
	r.statusCode = 400
	r.statusText = "Bad Request"
}

// NotFound indicates that the server has not found anything matching the Request-URI.
func (r *Response) NotFound() {
	r.statusCode = 404
	r.statusText = "Not Found"
}

// Unauthorized indicates that the request has not been applied because it lacks valid authentication credentials for the target resource.
func (r *Response) Unauthorized() {
	r.statusCode = 401
	r.statusText = "Unauthorized"
}

// Forbidden indicates that the server understood the request but refuses to authorize it.
func (r *Response) Forbidden() {
	r.statusCode = 403
	r.statusText = "Forbidden"
}

// InternalServerError indicates that the server encountered an unexpected condition which prevented it from fulfilling the request.
func (r *Response) InternalServerError() {
	r.statusCode = 500
	r.statusText = "Internal Server Error"
}

//********************************************************************************************************************
// Getters
//********************************************************************************************************************

// StatusCode returns the response status code.
func (r *Response) StatusCode() int { return r.statusCode }

// Protocol returns the response protocol.
func (r *Response) Protocol() string { return r.protocol }

func (r *Response) StatusText() string { return r.statusText }

// Header returns the response header.
func (r *Response) Header() Header { return r.header }

// Body returns the response body.
func (r *Response) Body() io.ReadCloser { return r.body }

//********************************************************************************************************************//
// Helpers
//********************************************************************************************************************//

// ReadResponse reads and returns an HTTP response from r.
// The req parameter optionally specifies the Request that corresponds
// to this Response. If nil, a GET request is assumed.
// Clients must call resp.Body.Close when finished reading resp.Body.
// After that call, clients can inspect resp.Trailer to find key/value
// pairs included in the response trailer.
func ReadResponse(r io.Reader) (*Response, error) {
	// Read the response
	resp, err := readResponse(r)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// readResponse parses the response from the reader and return a Response.
func readResponse(r io.Reader) (*Response, error) {
	resp := &Response{}
	reader := bufio.NewReader(r) // reader to read response
	n := 0                       // number of bytes read

	// 1.) Response line
	statusLine, err := reader.ReadBytes('\n')
	n = len(statusLine)
	if err != nil {
		return nil, err
	}
	status := strings.TrimSpace(string(statusLine))
	statusLines := strings.SplitN(status, " ", 3)
	if len(statusLines) != 3 {
		return nil, fmt.Errorf("invalid response status line: %s", status)
	}
	resp.protocol = strings.TrimSpace(statusLines[0])
	resp.statusCode, err = strconv.Atoi(strings.TrimSpace(statusLines[1]))
	if err != nil {
		return nil, err
	}
	resp.statusText = strings.TrimSpace(statusLines[2])

	// 2.) Headers
	resp.header = NewHeader()
	for {
		line, err := reader.ReadString('\n') // read line
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
		}
		n += len(line)
		// Headers are terminated by a blank line "\r\n"
		if line == "\r\n" || err == io.EOF {
			break
		}
		// Parse line
		parts := strings.SplitN(line, ":", 2) // split line into key and value
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid header line")
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		resp.header.Set(key, value)
	}

	// TODO: 3.) Body
	cl := resp.header.Get("Content-Length")
	if cl != "" {
		// if err != nil {
		// 	return resp, fmt.Errorf("Error parsing 'Content-Length': %s", err)
		// }
		// buf := make([]byte, l)
		// n2, err := reader.Read(buf)
		// if err != nil {
		// 	if err != io.EOF {
		// 		return resp, fmt.Errorf("Error reading response: %s", err)
		// 	}
		// 	if n2 != l {
		// 		return resp, fmt.Errorf("read %d bytes does not match 'Content-Length: %d'", n2, l)
		// 	}
		// }
		// resp.body = buf
		// resp.size = n + n2
		resp.contentLength, err = strconv.Atoi(cl)
		if err != nil {
			return resp, fmt.Errorf("Error parsing 'Content-Length': %s", err)
		}
		resp.body = io.NopCloser(reader)
	}
	return resp, nil
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

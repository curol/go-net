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

package message

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
)

// Response represents a bare bones HTTP response.
//
// Note, there are 3 general parts of a response:
// (1) Response line, (2) Headers, and (3) Body.
// Also, you can seperate it into two parts: (1) head and (2) body.
type Response struct {
	// Connection
	conn io.ReadWriteCloser // interface for connection
	w    *bufio.Writer      // writer for response

	// 1.) Response line is the first line of a response.
	// Format: <protocol> <statusCode> <statusText>
	// E.g., "HTTP/1.1 200 OK"
	protocol   string // e.g., "HTTP/1.1"
	statusCode int    // e.g., 200
	statusText string // e.g., "OK"

	// 2.) Headers come after the response line and each header is a new line of format "<key>: <value>".
	header Header

	// 3.) Body is the payload or contents of the response.
	// For brevity, only using []byte for body.
	body []byte

	// Misc
	size int // size of response (head+body)
}

func NewResponse(conn net.Conn) *Response {
	r := &Response{
		// Connection
		conn: conn,
		w:    bufio.NewWriter(conn), // writer for response
		// 1.) Status
		protocol:   "HTTP/1.1", // default protocol
		statusCode: 200,        // default status code
		statusText: "OK",       // default status text
		// 2.) Headers
		header: NewHeader(), // empty header for server to write to
		// 3.) Body
		body: make([]byte, 0), // empty body for server to write to
	}
	return r
}

//######################################################################################################################
// Encode
//######################################################################################################################

// ToBytes converts the response to a byte slice.
func (r *Response) ToBytes() []byte {
	b := bytes.NewBuffer(nil)
	_, err := r.WriteTo(b) // write to bytes buffer
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

// ToString converts the response to a string.
func (r *Response) ToString() string {
	return string(r.ToBytes())
}

// Serialize serializes the response to a byte slice.
func (r *Response) serialize() []byte {
	head := r.serializeHead() // 1.) serialize head
	body := r.body            // 2.) payload/contents
	return append(head, body...)
}

// SerializeHead serializes the head of a response (status line + headers) to a byte slice.
func (r *Response) serializeHead() []byte {
	delm := "\r\n" // seperator
	// Format the status line
	statusLine := r.Status() + delm
	// Get headers
	headersLine := r.header.ToString() + delm
	// Add the last delm for (1) seperating the head and body and (2) the start of the body.
	endOfHeadLine := delm
	// Join lines
	s := statusLine + headersLine + endOfHeadLine
	return []byte(s)
}

//######################################################################################################################
// Writer
//######################################################################################################################

// WriteTo writes the response `r` to the writer `w`.
func (r *Response) WriteTo(w io.Writer) (int64, error) {
	// Serialize the head and body of the response
	// to transfer this to the writer `w`
	// TODO: Transfer in chunks?

	// // 1.) First, write the head
	// head := r.Head()
	// n, err := w.Write([]byte(head))
	// if err != nil {
	// 	return 0, err
	// }
	// // 2.) Second, write the body
	// n2, err := w.Write(r.body)
	// if err != nil {
	// 	return 0, err
	// }

	output := r.serialize()
	n, err := w.Write(output)
	if err != nil {
		return 0, err
	}

	return int64(n), err
}

// Write writes the data to the writer.
func (r *Response) Write(b []byte) (int, error) {
	return r.w.Write(b)
}

// Flush flushes the writer.
func (r *Response) Flush() error {
	return r.w.Flush()
}

// WriteOutput writes the serialized response to the writer and flushes the writer to the
// connection.
func (r *Response) WriteOutput() error {
	output := r.serialize()   // encode the response to a byte slice
	_, err := r.Write(output) // write the response output to the writer
	if err != nil {
		return err
	}
	return r.Flush() // flush the writer to the client connection
}

//######################################################################################################################
// Connection
//######################################################################################################################

// Close closes the connection and sends an EOF to the connection.
func (r *Response) Close() error {
	return r.conn.Close()
}

//######################################################################################################################
// Headers
//######################################################################################################################

// WriteHeader writes the header `k` and `v` to the response.
func (r *Response) WriteHeader(k string, v string) {
	r.header.Set(k, v)
}

//######################################################################################################################
// Body
//######################################################################################################################

// Text writes the string `s` to the response body and sets the content to `text/plain`.
func (r *Response) Text(s string) {
	ct := "text/plain"
	cl := strconv.Itoa(len(s))
	r.header.Set("Content-Type", ct)
	r.header.Set("Content-Length", cl)
	r.body = []byte(s)
}

// JSON writes the JSON `v` to the response body and sets the content to `application/json`.
func (r *Response) JSON(v any) error {
	result, err := json.Marshal(v)
	if err != nil {
		return err
	}
	r.header.Set("Content-Type", "application/json")
	r.header.Set("Content-Length", strconv.Itoa(len(result)))
	r.body = result
	return nil
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

// Head returns the response head as a string.
//
// Format:
//
// ```
// Line 1 = Response line
// Lines 2...N = Header{key: value, ...}
// Line N+1 = "\r\n"
// ```
func (r *Response) Head() []byte {
	return r.serializeHead()
}

// Status returns the respnonse line.
func (r *Response) Status() string {
	// return r.protocol + " " + strconv.Itoa(r.statusCode) + " " + r.statusText
	return fmt.Sprintf("%s %s %s", r.protocol, strconv.Itoa(r.statusCode), r.statusText)
}

// StatusCode returns the response status code.
func (r *Response) StatusCode() int { return r.statusCode }

// Protocol returns the response protocol.
func (r *Response) Protocol() string { return r.protocol }

func (r *Response) StatusText() string { return r.statusText }

// Header returns the response header.
func (r *Response) Header() Header { return r.header }

// Body returns the response body.
func (r *Response) Body() []byte { return r.body }

// Size returns the size of the response.
func (r *Response) Size() int { return r.size }

//######################################################################################################################
// Logic
//######################################################################################################################

// Equals returns true if the other Request is equal to this Request.
func (p *Response) Equals(other *Response) error {
	// Check size
	// if p.Len() != other.Len() {
	// 	return fmt.Errorf("size mismatch (%d != %d)", p.Len(), other.Len())
	// }

	// Response line
	if p.Status() != other.Status() {
		return fmt.Errorf("respone line mismatch (%s != %s)", p.Status(), other.Status())
	}
	if p.protocol != other.protocol {
		return fmt.Errorf("protocol mismatch (%s != %s)", p.protocol, other.protocol)
	}
	if p.statusCode != other.statusCode {
		return fmt.Errorf("status code mismatch (%d != %d)", p.statusCode, other.statusCode)
	}
	if p.statusText != other.statusText {
		return fmt.Errorf("status text mismatch (%s != %s)", p.statusText, other.statusText)
	}

	// Headers
	if len(p.header) != len(other.header) { // check length
		return fmt.Errorf("header's size mismatch (%d != %d)", len(p.header), len(other.header))
	}
	for k, v := range p.header { // order dont matter, check if the other map contains the same key-value pairs and size
		if v != other.header[k] {
			return fmt.Errorf("header mismatch for key %s (%s != %s)", k, v, other.header[k])
		}
	}

	// Check body
	if !bytes.Equal(p.body, other.body) {
		return fmt.Errorf("body mismatch (%s != %s)", p.body, other.body)
	}

	return nil
}

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
	return nil, nil
}

// SendChunks writes data to connection in chunks.
func sendChunks(conn net.Conn, data []byte, chunkSize int) error {
	// Send data in chunks
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := data[i:end]
		_, err := conn.Write(chunk)
		if err != nil {
			return err
		}
	}

	return nil
}

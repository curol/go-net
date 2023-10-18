//********************************************************************************************************************//
// Response
//
// Response represents an HTTP response received by a client,
// It handles writing a client response.
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
	"io"
	"message/util"
	"net"
	"strconv"
)

// Response represents an HTTP response received by a client.
// There are 3 general parts of a response:
//
//		(1.) Response line
//		(2.) Headers
//	 (3.) Body
type Response struct {
	// Connection
	w    *bufio.Writer // writer for response
	conn net.Conn      // connection for response

	// 1.) Response line is the first line of a response.
	// Format: <protocol> <statusCode> <statusText>
	// E.g., "HTTP/1.1 200 OK"
	status     string // response line
	protocol   string // e.g., "HTTP/1.1"
	statusCode int    // e.g., 200
	statusText string // e.g., "OK"

	// 2.) Headers come after the response line and each header is a new line of format "<key>: <value>".
	header Header

	// 3.) Body is the payload or contents of the response.
	body *Body

	// Misc
	size int // size of response (head+body)
}

func NewResponse(conn net.Conn) *Response {
	return &Response{
		w:      bufio.NewWriter(conn), // writer for response
		header: NewHeader(),           // empty header for server to write to
		body:   NewBody(),             // empty body for server to write to
		conn:   conn,                  // connection for response
	}
}

// Write writes r to w in the HTTP/1.x server response format,
// including the status line, headers, body, and optional trailer.
//
// This method consults the following fields of the response r:
//
//	StatusCode
//	ProtoMajor
//	ProtoMinor
//	Request.Method
//	TransferEncoding
//	Trailer
//	Body
//	ContentLength
//	Header, values for non-canonical keys will have unpredictable behavior
//
// The Response Body is closed after it is sent.
//
//	func (r *Response) Write(w io.Writer) (int64, error) {
//		// r.WriteTo(w)
//		// TODO: Write response to w
//		w.Write()
//	}
func (r *Response) WriteHeader(k string, v string) {
	r.header.Set(k, v)
}

// Write writes the data to the writer.
func (r *Response) Write(b []byte) (int, error) {
	return r.w.Write(b)
}

// Flush flushes the response writer.
func (r *Response) Flush() error {
	return r.w.Flush()
}

// // WriteString writes the contents of the string `s` to the writer `w`, which accepts a slice of bytes.
// func (r *Response) WriteString(s string) (int, error) {
// 	return r.w.WriteString(s)
// }

func (r *Response) Text(s string) (int, error) {
	ct := "text/plain"
	cl := strconv.Itoa(len(s))
	r.header.Set("Content-Type", ct)
	r.header.Set("Content-Length", cl)
	return r.body.Text(s)
}

// WriteTo writes the writer `w` to the response.
func (r *Response) WriteTo(w io.Writer) (int64, error) {
	// TODO: Transfer in chunks?

	// Serialize head and body to transfer to the writer `w`
	// 1.) Head
	head := r.Head()
	n, err := r.w.Write([]byte(head))
	if err != nil {
		return 0, err
	}
	// 2.) Body
	defer r.body.Close()
	if r.body.IsContents() {
		n2, err := util.CopyReaderToWriter(w, r.body) // write reader `body` to the writer `w`
		return int64(n) + n2, err
	}

	// Return bytes written
	return int64(n), err
}

// Close closes the connection.
func (r *Response) Close() error {
	return r.conn.Close()
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
func (r *Response) Head() string {
	delm := "\r\n"                     // seperator
	statusLine := r.status + delm      // first line
	headersLine := r.header.Join(delm) // headers to string
	endOfHeadLine := delm              // seperate head and body

	return statusLine + headersLine + endOfHeadLine // add the last delm for (1) seperating the head and body and (2) the start of the body.
}

// Body returns the response body.
func (r *Response) Body() *Body { return r.body }

// Header returns the response header.
func (r *Response) Header() Header { return r.header }

// Status returns the response status.
func (r *Response) Status() string { return r.status }

// StatusCode returns the response status code.
func (r *Response) StatusCode() int { return r.statusCode }

// Protocol returns the response protocol.
func (r *Response) Protocol() string { return r.protocol }

// Size returns the size of the response.
func (r *Response) Size() int { return r.size }

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

// // Serialize converts the Response into a format that can be stored or transmitted.
// func (r *Response) Serialize(){
// 	// TODO: Transfer in chunks?

// 	// 1.) Head
// 	head := r.Head()
// 	n, err := r.w.Write([]byte(head))
// 	if err != nil {
// 		return 0, err
// 	}

// 	// 2.) Body
// 	// TODO: Close body?
// 	// defer r.body.Close()
// 	n2, err := util.CopyReaderToWriterN(w, r.body, int64(r.contentLength)) // write reader `body` to the writer `w`

// 	// Return bytes written
// 	return int64(n) + n2, err
// }

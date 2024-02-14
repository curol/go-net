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
	"net"
	"os"
	"strconv"
	"strings"
)

var respExcludeHeader = map[string]bool{
	"Content-Length":    true,
	"Transfer-Encoding": true,
	"Trailer":           true,
}

// Response represents a bare bones HTTP response.
//
// Note, there are 3 general parts of a response:
// (1) Response line, (2) Headers, and (3) Body.
// Also, you can seperate it into two parts: (1) head and (2) body.
type Response struct {
	// Response-Line format: <protocol> <statusCode> <statusText>
	// E.g.,
	// 	"HTTP/1.0 200 OK"
	Proto      string // e.g. "HTTP/1.0"
	ProtoMajor int    // e.g. 1
	ProtoMinor int    // e.g. 0
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g., 200
	StatusText string // e.g., "OK"

	// 2.) Headers come after the response line and each header is a new line of format "<key>: <value>".
	Header        Header
	ContentLength int
	ContentType   string

	// 3.) Body is the payload or contents of the response.
	// For brevity, only using []byte for body.
	Body io.ReadCloser

	// Request is the request that was sent to obtain this Response.
	// Request's Body is nil (having already been consumed).
	// This is only populated for Client requests.
	Request *Request

	// IsClose records whether the header directed that the connection be
	// closed after reading Body. The value is advice for clients: neither
	// ReadResponse nor Response.Write ever closes a connection.
	IsClose bool

	conn net.Conn

	wroteHeader bool

	code int
}

func NewResponse(conn net.Conn) *Response {
	return &Response{
		Proto:         protocol, // default protocol
		StatusCode:    200,      // default status code
		StatusText:    "OK",     // default status text
		Header:        NewHeader(),
		Body:          nil,
		ContentLength: 0,
		conn:          conn,
	}
}

// Close closes the connection and writes io.EOF to the connection.
func (r *Response) Close() error {
	if r.IsClose {
		return fmt.Errorf("response already closed")
	}
	r.IsClose = true
	return nil
}

// WriteHeader writes the header `k` and `v` to the response.
func (r *Response) WriteHeader(statusCode int) {
	req := r.Request
	if req == nil {
		fmt.Println("WriteHeader(): Request is nil")
		return
	}
	if req.URL == nil {
		fmt.Println("WriteHeader(): Request.URL is nil")
		return
	}
	if r.wroteHeader {
		// Note: explicitly using Stderr, as Stdout is our HTTP output.
		fmt.Fprintf(os.Stderr, "WriteHeader attempted to write header twice on request for %s", req.URL)
		return
	}
	r.wroteHeader = true
	r.code = statusCode
}

// Write writes r to w.
func (r *Response) Write(b []byte) (int, error) {
	return r.conn.Write(b)
}

func (r *Response) WriteTo(w io.Writer) (int64, error) {
	// Type switch writer
	switch v := w.(type) {
	case *bufio.Writer:
		return r.write(v)
	case *bytes.Buffer:
		bw := bufio.NewWriter(v)
		return r.write(bw)
	default:
		// TOOD: Add other types
	}
	return 0, fmt.Errorf("invalid type")
}

// write serializes the response to the writer.
func (r *Response) write(w *bufio.Writer) (int64, error) {
	if r == nil {
		return 0, fmt.Errorf("response is nil")
	}
	// 1. Response line
	rl := fmt.Sprintf("%s %s %s\r\n", r.Proto, strconv.Itoa(r.StatusCode), r.StatusText)
	_, err := w.Write([]byte(rl))
	if err != nil {
		return 0, err
	}

	err = w.Flush() // flush request line
	if err != nil {
		return 0, err
	}

	// 2. Header
	err = r.Header.Write(w)
	if err != nil {
		return 0, err
	}
	err = w.Flush() // flush header
	if err != nil {
		return 0, err
	}

	// 3. End of head
	fmt.Fprintf(w, "\r\n") // automatically flushes

	// 5. Body
	if r.Body != nil {
		contLen := int64(r.ContentLength)
		_, err = io.CopyN(w, r.Body, contLen) // copy Body to writer
		if err != nil {
			return 0, err
		}
		// TODO: Flush?
		// err = w.Flush() // flush the writer to the client connection
		// if err != nil {
		// 	return 0, err
		// }
	}

	if err != nil {
		return 0, err
	}
	return int64(w.Size()), err
}

// Cookies parses and returns the cookies set in the Set-Cookie headers.
func (r *Response) Cookies() []*Cookie {
	return readSetCookies(r.Header)
}

// // ToBuffer encodes the request into a *bytes.Buffer.
// func (r *Response) ToBuffer() (*bytes.Buffer, error) {
// 	buf := bytes.NewBuffer(nil)
// 	_, err := r.Write(buf)
// 	return buf, err
// }

//********************************************************************************************************************//
// Output
//********************************************************************************************************************//

// Text writes the string `s` to the response body and sets the content to `text/plain`.
func (r *Response) Text(s string) {
	ct := "text/plain"
	cl := strconv.Itoa(len(s))
	r.Header.Set("Content-Type", ct)
	r.Header.Set("Content-Length", cl)

	r.Body = io.NopCloser(bytes.NewBufferString(s))
}

// JSON writes the JSON `v` to the response body and sets the content to `application/json`.
func (r *Response) JSON(v any) error {
	result, err := json.Marshal(v)
	if err != nil {
		return err
	}
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Content-Length", strconv.Itoa(len(result)))

	r.Body = io.NopCloser(bytes.NewBuffer(result))

	return nil
}

func (r *Response) HTML(s string) {
	ct := "text/html"
	cl := strconv.Itoa(len(s))
	r.Header.Set("Content-Type", ct)
	r.Header.Set("Content-Length", cl)

	r.Body = io.NopCloser(bytes.NewBufferString(s))
}

func (r *Response) XML(s string) {
	ct := "text/xml"
	cl := strconv.Itoa(len(s))
	r.Header.Set("Content-Type", ct)
	r.Header.Set("Content-Length", cl)

	r.Body = io.NopCloser(bytes.NewBufferString(s))
}

func (r *Response) JSONP(s string) {
	ct := "application/javascript"
	cl := strconv.Itoa(len(s))
	r.Header.Set("Content-Type", ct)
	r.Header.Set("Content-Length", cl)

	r.Body = io.NopCloser(bytes.NewBufferString(s))
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
	r.Header.Set("Content-Type", ct)
	r.Header.Set("Content-Length", cl)

	// TODO: Use io.NopCloser()?
	// r.body = io.NopCloser(f)
	r.Body = f
}

//######################################################################################################################
// Status
//######################################################################################################################

// Ok indicates that the request is successful.
func (r *Response) Ok() {
	r.StatusCode = 200
	r.StatusText = "OK"
}

// BadRequest indicates that the request could not be understood by the server due to malformed syntax.
func (r *Response) BadRequest() {
	r.StatusCode = 400
	r.StatusText = "Bad Request"
}

// NotFound indicates that the server has not found anything matching the Request-URI.
func (r *Response) NotFound() {
	r.StatusCode = 404
	r.StatusText = "Not Found"
}

// Unauthorized indicates that the request has not been applied because it lacks valid authentication credentials for the target resource.
func (r *Response) Unauthorized() {
	r.StatusCode = 401
	r.StatusText = "Unauthorized"
}

// Forbidden indicates that the server understood the request but refuses to authorize it.
func (r *Response) Forbidden() {
	r.StatusCode = 403
	r.StatusText = "Forbidden"
}

// InternalServerError indicates that the server encountered an unexpected condition which prevented it from fulfilling the request.
func (r *Response) InternalServerError() {
	r.StatusCode = 500
	r.StatusText = "Internal Server Error"
}

//********************************************************************************************************************
// Getters
//********************************************************************************************************************

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
	resp.Proto = strings.TrimSpace(statusLines[0])
	major, minor, ok := ParseHTTPVersion(resp.Proto)
	if !ok {
		return nil, fmt.Errorf("invalid response protocol version: %s", resp.Proto)
	}
	resp.ProtoMajor = major
	resp.ProtoMinor = minor
	resp.StatusCode, err = strconv.Atoi(strings.TrimSpace(statusLines[1]))
	if err != nil {
		return nil, err
	}
	resp.StatusText = strings.TrimSpace(statusLines[2])

	// 2.) Headers
	resp.Header = NewHeader()
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
		resp.Header.Set(key, value)
	}

	// TODO: 3.) Body
	cl := resp.Header.Get("Content-Length")
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
		resp.ContentLength, err = strconv.Atoi(cl)
		if err != nil {
			return resp, fmt.Errorf("Error parsing 'Content-Length': %s", err)
		}
		resp.Body = io.NopCloser(reader)
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

// A response represents the server side of an HTTP response.
type response struct {
	conn             net.Conn
	req              *Request // request for this response
	reqBody          io.ReadCloser
	wroteHeader      bool // a non-1xx header has been (logically) written
	wroteContinue    bool // 100 Continue response was written
	wants10KeepAlive bool // HTTP/1.0 w/ Connection "keep-alive"
	wantsClose       bool // HTTP request has Connection "close"
}

// The Life Of A Write is like this:
//
// Handler starts. No header has been sent. The handler can either
// write a header, or just start writing. Writing before sending a header
// sends an implicitly empty 200 OK header.
//
// If the handler didn't declare a Content-Length up front, we either
// go into chunking mode or, if the handler finishes running before
// the chunking buffer size, we compute a Content-Length and send that
// in the header instead.
//
// Likewise, if the handler didn't set a Content-Type, we sniff that
// from the initial chunk of output.
//
// The Writers are wired together like:
//
//  1. *response (the ResponseWriter) ->
//  2. (*response).w, a [*bufio.Writer] of bufferBeforeChunkingSize bytes ->
//  3. chunkWriter.Writer (whose writeHeader finalizes Content-Length/Type)
//     and which writes the chunk headers, if needed ->
//  4. conn.bufw, a *bufio.Writer of default (4kB) bytes, writing to ->
//  5. checkConnErrorWriter{c}, which notes any non-nil error on Write
//     and populates c.werr with it if so, but otherwise writes to ->
//  6. the rwc, the [net.Conn].
//
// TODO(bradfitz): short-circuit some of the buffering when the
// initial header contains both a Content-Type and Content-Length.
// Also short-circuit in (1) when the header's been sent and not in
// chunking mode, writing directly to (4) instead, if (2) has no
// buffered data. More generally, we could short-circuit from (1) to
// (3) even in chunking mode if the write size from (1) is over some
// threshold and nothing is in (2).  The answer might be mostly making
// bufferBeforeChunkingSize smaller and having bufio's fast-paths deal
// with this instead.
func (r *response) Write(data []byte) {
	// /
}

// either dataB or dataS is non-zero.
func (w *response) write(lenData int, dataB []byte, dataS string) (n int, err error) {
	// if w.conn.hijacked() {
	// 	if lenData > 0 {
	// 		caller := relevantCaller()
	// 		w.conn.server.logf("http: response.Write on hijacked connection from %s (%s:%d)", caller.Function, path.Base(caller.File), caller.Line)
	// 	}
	// 	return 0, ErrHijacked
	// }

	// if w.canWriteContinue.Load() {
	// 	// Body reader wants to write 100 Continue but hasn't yet.
	// 	// Tell it not to. The store must be done while holding the lock
	// 	// because the lock makes sure that there is not an active write
	// 	// this very moment.
	// 	w.writeContinueMu.Lock()
	// 	w.canWriteContinue.Store(false)
	// 	w.writeContinueMu.Unlock()
	// }

	// if !w.wroteHeader {
	// 	w.WriteHeader(StatusOK)
	// }
	// if lenData == 0 {
	// 	return 0, nil
	// }
	// if !w.bodyAllowed() {
	// 	return 0, ErrBodyNotAllowed
	// }

	// w.written += int64(lenData) // ignoring errors, for errorKludge
	// if w.contentLength != -1 && w.written > w.contentLength {
	// 	return 0, ErrContentLength
	// }
	// if dataB != nil {
	// 	return w.w.Write(dataB)
	// } else {
	// 	return w.w.WriteString(dataS)
	// }
	return 0, nil // TODO: omit this
}

func (w *response) WriteHeader(code int) {
	// if w.conn.hijacked() {
	// 	caller := relevantCaller()
	// 	w.conn.server.logf("http: response.WriteHeader on hijacked connection from %s (%s:%d)", caller.Function, path.Base(caller.File), caller.Line)
	// 	return
	// }
	// if w.wroteHeader {
	// 	caller := relevantCaller()
	// 	w.conn.server.logf("http: superfluous response.WriteHeader call from %s (%s:%d)", caller.Function, path.Base(caller.File), caller.Line)
	// 	return
	// }
	// checkWriteHeaderCode(code)

	// // Handle informational headers.
	// //
	// // We shouldn't send any further headers after 101 Switching Protocols,
	// // so it takes the non-informational path.
	// if code >= 100 && code <= 199 && code != StatusSwitchingProtocols {
	// 	// Prevent a potential race with an automatically-sent 100 Continue triggered by Request.Body.Read()
	// 	if code == 100 && w.canWriteContinue.Load() {
	// 		w.writeContinueMu.Lock()
	// 		w.canWriteContinue.Store(false)
	// 		w.writeContinueMu.Unlock()
	// 	}

	// 	writeStatusLine(w.conn.bufw, w.req.ProtoAtLeast(1, 1), code, w.statusBuf[:])

	// 	// Per RFC 8297 we must not clear the current header map
	// 	w.handlerHeader.WriteSubset(w.conn.bufw, excludedHeadersNoBody)
	// 	w.conn.bufw.Write(crlf)
	// 	w.conn.bufw.Flush()

	// 	return
	// }

	// w.wroteHeader = true
	// w.status = code

	// if w.calledHeader && w.cw.header == nil {
	// 	w.cw.header = w.handlerHeader.Clone()
	// }

	// if cl := w.handlerHeader.get("Content-Length"); cl != "" {
	// 	v, err := strconv.ParseInt(cl, 10, 64)
	// 	if err == nil && v >= 0 {
	// 		w.contentLength = v
	// 	} else {
	// 		w.conn.server.logf("http: invalid Content-Length of %q", cl)
	// 		w.handlerHeader.Del("Content-Length")
	// 	}
	// }
}

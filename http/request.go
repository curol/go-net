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
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/curol/network/http/internal/timeformat"
	url "github.com/curol/network/url"
)

// Request is a structure for parsed data and buffers from reading a client request.
// It implements streaming, buffering, parsing, decoding, and the interface writeTo.
//
// More specifically, it will (1) read from a reader or from a byte slice and (2) parse the request line, headers, and body.
// Then, it will (3) buffer the request line, headers, and body.
//
// Note, for brevity, the body is buffered in memory. This is not ideal for large requests.
type Request struct {
	// RequestURI is the unmodified request-target of the
	// Request-Line (RFC 7230, Section 3.1.1) as sent by the client
	// to a server. Usually the URL field should be used instead.
	// It is an error to set this field in an HTTP client request.
	RequestURI string

	// Method specifies the HTTP method (GET, POST, PUT, etc.).
	// For client requests, an empty string means GET.
	Method string

	// The protocol version for incoming server requests.
	//
	// For client requests, these fields are ignored. The HTTP
	// client code always uses either HTTP/1.1 or HTTP/2.
	// See the docs on Transport for details.
	Protocol string

	// URL specifies either the URI being requested (for server
	// requests) or the URL to access (for client requests).
	//
	// For server requests, the URL is parsed from the URI
	// supplied on the Request-Line as stored in RequestURI.  For
	// most requests, fields other than Path and RawQuery will be
	// empty. (See RFC 7230, Section 5.3)
	//
	// For client requests, the URL's Host specifies the server to
	// connect to, while the Request's Host field optionally
	// specifies the Host header value to send in the HTTP
	// request.
	URL *url.URL

	// HTTP defines that header names are case-insensitive. The
	// request parser implements this by using CanonicalHeaderKey,
	// making the first character and any characters following a
	// hyphen uppercase and the rest lowercase.
	//
	// For client requests, certain headers such as Content-Length
	// and Connection are automatically written when needed and
	// values in Header may be ignored. See the documentation
	// for the Request.Write method.
	Header Header

	// Body is the request's body.
	//
	// For client requests, a nil body means the request has no
	// body, such as a GET request. The HTTP Client's Transport
	// is responsible for calling the Close method.
	//
	// For server requests, the Request Body is always non-nil
	// but will return EOF immediately when no body is present.
	// The Server will close the request body. The ServeHTTP
	// Handler does not need to.
	//
	// Body must allow Read to be called concurrently with Close.
	// In particular, calling Close should unblock a Read waiting
	// for input.
	Body io.ReadCloser // stream for body contents which allows reading and closing connection

	// GetBody defines an optional func to return a new copy of
	// Body. It is used for client requests when a redirect requires
	// reading the body more than once. Use of GetBody still
	// requires setting Body.
	//
	// For server requests, it is unused.
	GetBody func() (io.ReadCloser, error)

	// ContentLength records the length of the associated content.
	// The value -1 indicates that the length is unknown.
	// Values >= 0 indicate that the given number of bytes may
	// be read from Body.
	//
	// For client requests, a value of 0 with a non-nil Body is
	// also treated as unknown.
	ContentLength int64

	// ContentType specifies the request body's MIME type.
	ContentType string

	// RemoteAddr allows HTTP servers and other software to record
	// the network address that sent the request, usually for
	// logging. This field is not filled in by ReadRequest and
	// has no defined format. The HTTP server in this package
	// sets RemoteAddr to an "IP:port" address before invoking a
	// handler.
	// This field is ignored by the HTTP client.
	RemoteAddress string // address of client

	// For server requests, Host specifies the host on which the
	// URL is sought. For HTTP/1 (per RFC 7230, section 5.4), this
	// is either the value of the "Host" header or the host name
	// given in the URL itself. For HTTP/2, it is the value of the
	// ":authority" pseudo-header field.
	// It may be of the form "host:port". For international domain
	// names, Host may be in Punycode or Unicode form. Use
	// golang.org/x/net/idna to convert it to either format if
	// needed.
	// To prevent DNS rebinding attacks, server Handlers should
	// validate that the Host header has a value for which the
	// Handler considers itself authoritative. The included
	// ServeMux supports patterns registered to particular host
	// names and thus protects its registered Handlers.
	Host string // host address

	// Form contains the parsed form data, including both the URL
	// field's query parameters and the PATCH, POST, or PUT form data.
	// This field is only available after ParseForm is called.
	// The HTTP client ignores Form and uses Body instead.
	Form url.Values // parsed form

	// MultipartForm is the parsed multipart form, including file uploads.
	// This field is only available after ParseMultipartForm is called.
	// The HTTP client ignores MultipartForm and uses Body instead.
	MultipartForm *multipart.Form

	Cookies []*Cookie // parsed cookies

	// Close indicates whether to close the connection after
	// replying to this request (for servers) or after sending this
	// request and reading its response (for clients).
	//
	// For server requests, the HTTP server handles this automatically
	// and this field is not needed by Handlers.
	//
	// For client requests, setting this field prevents re-use of
	// TCP connections between requests to the same hosts, as if
	// Transport.DisableKeepAlives were set.
	Close bool

	TransferEncoding []string
}

// NewRequest is for client requests.
// It creates a new request with the given method, address, headers, and body.
func NewRequest(method string, address string, headers map[string][]string, body io.Reader) (*Request, error) {
	return newRequest(
		method,
		address,
		protocol,
		headers,
		body,
	)
}

func newRequest(method string, address string, prot string, header map[string][]string, body io.Reader) (*Request, error) {
	// Arrange
	method = strings.TrimSpace(strings.ToUpper(method))
	if method == "" {
		method = "GET"
	}
	address = strings.TrimSpace(address)
	u, err := url.Parse(address)
	if err != nil {
		panic(err)
	}
	h := Header(header)

	// TODO: Validate fields
	if prot != protocol {
		return nil, fmt.Errorf("invalid protocol")
	}

	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = io.NopCloser(body)
	}

	// The host's colon:port should be normalized. See Issue 14836.
	u.Host = removeEmptyPort(u.Host)

	// Set Request
	req := &Request{
		// Request line
		Method:     method,
		Protocol:   protocol,
		RequestURI: u.RequestURI(),
		URL:        u,
		Host:       u.Host, // TODO: Or use host header? Ex. host: getHost(h)
		// Headers
		Header: h,
		// Body
		Body:          rc,
		Form:          nil,
		MultipartForm: nil,
		// Connection
		RemoteAddress: "",
	}

	// Set Body
	if body != nil {
		switch v := body.(type) {
		case *bytes.Buffer:
			req.ContentLength = int64(v.Len())
			buf := v.Bytes()
			req.GetBody = func() (io.ReadCloser, error) {
				r := bytes.NewReader(buf)
				return io.NopCloser(r), nil
			}
		case *bytes.Reader:
			req.ContentLength = int64(v.Len())
			snapshot := *v
			req.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return io.NopCloser(&r), nil
			}
		case *strings.Reader:
			req.ContentLength = int64(v.Len())
			snapshot := *v
			req.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return io.NopCloser(&r), nil
			}
		default:
			// This is where we'd set it to -1 (at least
			// if body != NoBody) to mean unknown, but
			// that broke people during the Go 1.8 testing
			// period. People depend on it being 0 I
			// guess. Maybe retry later. See Issue 18117.
		}
		// For client requests, Request.ContentLength of 0
		// means either actually 0, or unknown. The only way
		// to explicitly say that the ContentLength is zero is
		// to set the Body to nil. But turns out too much code
		// depends on NewRequest returning a non-nil Body,
		// so we use a well-known ReadCloser variable instead
		// and have the http package also treat that sentinel
		// variable to mean explicitly zero.
		if req.GetBody != nil && req.ContentLength == int64(0) {
			req.Body = NoBody
			req.GetBody = func() (io.ReadCloser, error) { return NoBody, nil }
		}
	}

	return req, nil
}

// Write writes an HTTP/1.1 request, which is the header and body, in wire format.
// This method consults the following fields of the request:
//
//	Host
//	URL
//	Method (defaults to "GET")
//	Header
//	ContentLength
//	TransferEncoding
//	Body
//
// If Body is present, Content-Length is <= 0 and TransferEncoding
// hasn't been set to "identity", Write adds "Transfer-Encoding:
// chunked" to the header. Body is closed after it is sent.
//
// Write is used after the request has been parsed and validated.
func (r *Request) Write(w io.Writer) error {
	return r.write(w)
}

// write serializes r to w.
func (r *Request) write(w io.Writer) error {
	// Note:
	// - 'fmt.Fprintf' writes the string and automaitcally flushes the buffer
	// - 'w.Write' writes the string but does not flush the buffer
	// - 'w.Flush' flushes the buffer

	// TODO: defer closeBody

	// 1.) Wrap the writer in a bufio Writer if it's not already buffered.
	// Don't always call NewWriter, as that forces a bytes.Buffer
	// and other small bufio Writers to have a minimum 4k buffer
	// size.
	var bw *bufio.Writer
	if _, ok := w.(io.ByteWriter); !ok { // check if writer is bufferd
		bw = bufio.NewWriter(w)
		w = bw
	}

	// 2. Serialize and write the request line
	_, err := fmt.Fprintf(w, "%s %s %s\r\n", r.Method, r.URL.RequestURI(), r.Protocol)
	if err != nil {
		return err
	}

	// errMissingHost is returned by Write when there is no Host or URL present in
	// the Request.
	var errMissingHost = errors.New("http: Request.Write on Request with no Host or URL set")

	// 3. Serialize and write the headers
	// TODO: Write Transfer-Encoding, Trailer, and other headers

	// Find the target host. Prefer the Host: header, but if that
	// is not given, use the host from the request URL.
	host := r.Host
	if host == "" {
		if r.URL == nil {
			return errMissingHost
		}
		host = r.URL.Host
	}
	host = timeformat.RemoveZone(host)
	fmt.Fprintf(w, "Host: %s\r\n", host) // write host

	// Use the defaultUserAgent unless the Header contains one, which
	// may be blank to not send the header.
	userAgent := defaultUserAgent
	if r.Header.Get("User-Agent") != "" {
		userAgent = r.Header.Get("User-Agent")
	}
	if userAgent != "" {
		fmt.Fprintf(w, "User-Agent: %s\r\n", userAgent) // write user agent
	}

	cl := r.ContentLength
	if cl < 0 && r.Body != nil {
		fmt.Fprintf(w, "Transfer-Encoding: chunked\r\n") // write transfer encoding
	} else if cl > 0 {
		fmt.Fprintf(w, "Content-Length: %d\r\n", cl) // write content length
	} else {
		// TODO: Handle error
	}

	var reqWriteExcludeHeader = map[string]bool{
		"Host":              true,
		"User-Agent":        true,
		"Content-Length":    true,
		"Transfer-Encoding": true,
		"Trailer":           true,
	}
	err = r.Header.WriteSubset(w, reqWriteExcludeHeader) // write headers
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Validate fields

	// 4. Write blank line that ends with CRLF for end of head
	_, err = io.WriteString(w, "\r\n")
	if err != nil {
		return err
	}

	// 5. Flush
	if bw, ok := w.(*bufio.Writer); ok {
		err = bw.Flush()
		if err != nil {
			return err
		}
	}

	// 5. Write body
	if r.Body != nil {
		// TODO: Check content length not larger than max memory
		_, err := io.CopyN(w, r.Body, r.ContentLength) // write body to w
		if err != nil {
			return err
		}
	}
	if bw != nil {
		return bw.Flush()
	}
	return nil
}

// Reset resets the Request.
func (p *Request) Reset() {
	p = new(Request)
}

// UserAgent returns the client's User-Agent, if sent in the request.
func (r *Request) UserAgent() string {
	return r.Header.Get("User-Agent")
}

// ReadRequest reads and parses a request from a reader.
//
// Note: ReadRequest should only be used for servers.
func ReadRequest(r *bufio.Reader) (*Request, error) {
	req, err := readRequest(r)
	if err != nil {
		return nil, err
	}
	delete(req.Header, "Host")
	return req, nil
}

// readRequest reads and parses a request from a reader.
// A request represents an HTTP request received by a server or to be sent by a client.
// Each line in a raw request is terminated by CRLF ("\r\n").
//
// ## Request line
// - The request line is the first line of a request message with the following format: <method> <path> <protocol>
// - The request line is followed by a sequence of zero or more header fields, followed by an empty line, indicating the end of the header fields.
// - The method specifies the HTTP method to be performed on the resource identified by the request URI.
// - The path specifies the path to the resource on the server.
// - The protocol specifies the protocol version of the request.
//
// ### URI vs URL
// - URI stands for Uniform Resource Identifier, while URL stands for Uniform Resource Locator.
// - A URI is a string of characters that identifies a name or a resource on the Internet. It is a more general term that encompasses URLs.
// - A URL is a type of URI that includes the location of the resource on the Internet and the protocol used to access it. It specifies where an identified resource is available and the mechanism for retrieving it. In other words, a URL is a URI that, in addition to identifying a resource, provides a means of locating the resource by describing its primary access mechanism or network "location".
// - For example, `https://www.example.com/path/to/resource` is a URL. It's also a URI, because every URL is a URI, but not every URI is a URL. A URI like `mailto:example@example.com` is not a URL, because it doesn't specify a location on the Internet and a protocol for retrieving a resource.
//
// ### Request-URI
// - The Request-URI, or Request Uniform Resource Identifier, is a part of the first line (the request line) of an HTTP request message. It identifies the resource upon which to apply the request.
// - There are four forms of Request-URI: absoluteURI, abs_path, authority, and asterisk.
// - The path is the RFC 7230 "request-target": it may be either a  path or an absolute URL.
// - If target is an absolute URL, the host name from the URL is used.  // Otherwise, "example.com" is used.
//
// ## Header
// - The header fields are transmitted after the request line (in case of a request HTTP message) or the response line (in case of a response HTTP message), which is the first line of a message.
// - Header fields are colon-separated key-value pairs in clear-text string format, terminated by a carriage return (CR) and line feed (LF) character sequence.
// - The end of the header fields is indicated by an empty field, resulting in the transmission of two consecutive CR-LF pairs.
//
// ## Body
// - The body of a message (human-readable information to be transmitted) follows the header fields.
// - The body of a request (such as in a POST request) is used to send data to the server.
// - The body of a response (such as in a GET response) is used to send data to the client.
// - The body of a request or response is optional.
// - The body of a request or response is terminated by the first empty line after the header fields.
// - The body of a request or response can be any type of data, such as a file, an image, a video, or a text string.
// - The body of a request or response can be encoded in various formats, such as JSON or XML.
// - The body of a request or response can be compressed using various compression algorithms, such as gzip or deflate
func readRequest(r *bufio.Reader) (*Request, error) {
	if r == nil {
		return nil, fmt.Errorf("reader nil")
	}

	// 1. Read and parse first Line
	// TODO: Validate method, path, and protocol and parse HTTP Version
	line, err := r.ReadString('\n') // read first line
	defer func() {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
	}()
	if err != nil && err != io.EOF {
		return nil, err
	}

	method, requestURI, prot, ok := parseRequestLine(line) // parse first line
	if !ok {
		return nil, fmt.Errorf("invalid request line")
	}

	rawurl := requestURI
	u, err := url.ParseRequestURI(rawurl) // parse uri
	if err != nil {
		return nil, err
	}
	if prot != protocol {
		return nil, fmt.Errorf("invalid protocol")
	}

	// 2. Read and parse headers
	header := NewHeader()
	for { // read each new line until a blank line ("\r\n") is reached.
		line, err := r.ReadString('\n') // read line
		if err != nil && err != io.EOF {
			return nil, err
		}
		if line == "\r\n" || err == io.EOF { // headers are terminated by a blank line "\r\n"
			break
		}
		parts := strings.SplitN(line, ":", 2) // parse line by splitting line into key and value
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid header line")
		}
		// remove leading and trailing whitespace from key and value
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		header.Set(k, v)
	}

	req := &Request{
		Method:        method,
		Protocol:      prot,
		RequestURI:    requestURI,
		URL:           u,
		Host:          u.Host,
		Header:        header,
		ContentLength: getContentLength(header),
		ContentType:   header.Get("Content-Type"),
		Cookies:       readCookies(header, ""),
		Body:          io.NopCloser(r),
		Form:          nil,
		MultipartForm: nil,
		RemoteAddress: "",
	}

	// RFC 7230, section 5.3: Must treat
	//	GET /index.html HTTP/1.1
	//	Host: www.google.com
	// and
	//	GET http://www.google.com/index.html HTTP/1.1
	//	Host: doesntmatter
	// the same. In the second case, any Host line is ignored.
	if req.Host == "" {
		req.Host = req.Header.Get("Host")
	}

	return req, nil
}

// ErrNoCookie is returned by Request's Cookie method when a cookie is not found.
var ErrNoCookie = errors.New("http: named cookie not present")

// Cookie returns the named cookie provided in the request or
// [ErrNoCookie] if not found.
// If multiple cookies match the given name, only one cookie will
// be returned.
func (r *Request) Cookie(name string) (*Cookie, error) {
	if name == "" {
		return nil, ErrNoCookie
	}
	for _, c := range readCookies(r.Header, name) {
		return c, nil
	}
	return nil, ErrNoCookie
}

// AddCookie adds a cookie to the request. Per RFC 6265 section 5.4,
// AddCookie does not attach more than one [Cookie] header field. That
// means all cookies, if any, are written into the same line,
// separated by semicolon.
// AddCookie only sanitizes c's name and value, and does not sanitize
// a Cookie header already present in the request.
func (r *Request) AddCookie(c *Cookie) {
	s := fmt.Sprintf("%s=%s", sanitizeCookieName(c.Name), sanitizeCookieValue(c.Value))
	if c := r.Header.Get("Cookie"); c != "" {
		r.Header.Set("Cookie", c+"; "+s)
	} else {
		r.Header.Set("Cookie", s)
	}
}

// ParseHTTPVersion parses an HTTP version string according to RFC 7230, section 2.6.
// "HTTP/1.0" returns (1, 0, true). Note that strings without
// a minor version, such as "HTTP/2", are not valid.
func ParseHTTPVersion(vers string) (major, minor int, ok bool) {
	switch vers {
	case "HTTP/1.1":
		return 1, 1, true
	case "HTTP/1.0":
		return 1, 0, true
	}
	if !strings.HasPrefix(vers, "HTTP/") {
		return 0, 0, false
	}
	if len(vers) != len("HTTP/X.Y") {
		return 0, 0, false
	}
	if vers[6] != '.' {
		return 0, 0, false
	}
	maj, err := strconv.ParseUint(vers[5:6], 10, 0)
	if err != nil {
		return 0, 0, false
	}
	min, err := strconv.ParseUint(vers[7:8], 10, 0)
	if err != nil {
		return 0, 0, false
	}
	return int(maj), int(min), true
}

// parseRequestLine parses "GET /foo HTTP/1.1" into its three parts.
func parseRequestLine(line string) (method, requestURI, prot string, ok bool) {
	// Alternative
	// method, rest, ok1 := strings.Cut(line, " ")
	// requestURI, proto, ok2 := strings.Cut(rest, " ")
	// if !ok1 || !ok2 {
	// 	return "", "", "", false
	// }
	// return method, requestURI, proto, true
	// Original
	parts := strings.SplitN(line, " ", 3) // parse first line by splitting into method, path, and protocol
	if len(parts) != 3 {
		return "", "", "", false
	}
	method = strings.ToUpper(strings.TrimSpace(parts[0]))
	requestURI = strings.TrimSpace(parts[1])
	prot = strings.TrimSpace(parts[2])
	return method, requestURI, prot, true
}

func getRequestLine(method string, address string, protocol string) string {
	return fmt.Sprintf("%s %s %s\r\n", method, address, protocol)
}

func (r *Request) wantsClose() bool {
	if r.Close {
		return true
	}
	return hasToken(r.Header.Get("Connection"), "close")
}

func (r *Request) closeBody() error {
	if r.Body == nil {
		return nil
	}
	return r.Body.Close()
}

func writeChunks(chunks [][]byte, w io.Writer) {
	// Write the chunks
	for _, chunk := range chunks {
		// Write the chunk size in hexadecimal
		fmt.Fprintf(w, "%x\r\n", len(chunk))

		// Write the chunk data
		fmt.Fprint(w, chunk)

		// Write the chunk end
		fmt.Fprint(w, "\r\n")
	}

	// Write the last chunk (size 0)
	fmt.Fprint(w, "0\r\n\r\n")
}

func readChunks(r *bufio.Reader) (string, error) {
	var body strings.Builder

	for {
		// Read the chunk size
		line, err := r.ReadString('\n')
		if err != nil {
			return "", err
		}

		size, err := strconv.ParseInt(strings.TrimSpace(line), 16, 64)
		if err != nil {
			return "", err
		}

		// End of the message
		if size == 0 {
			break
		}

		// Read the chunk data
		chunk := make([]byte, size)
		if _, err := io.ReadFull(r, chunk); err != nil {
			return "", err
		}

		body.Write(chunk)

		// Read the chunk end
		if _, err := r.ReadString('\n'); err != nil {
			return "", err
		}
	}

	return body.String(), nil
}

// TODO: Implement parseForm
// // ParseForm populates r.Form and r.PostForm.
// //
// // For all requests, ParseForm parses the raw query from the URL and updates
// // r.Form.
// //
// // For POST, PUT, and PATCH requests, it also reads the request body, parses it
// // as a form and puts the results into both r.PostForm and r.Form. Request body
// // parameters take precedence over URL query string values in r.Form.
// //
// // If the request Body's size has not already been limited by MaxBytesReader,
// // the size is capped at 10MB.
// //
// // For other HTTP methods, or when the Content-Type is not
// // application/x-www-form-urlencoded, the request Body is not read, and
// // r.PostForm is initialized to a non-nil, empty value.
// //
// // ParseMultipartForm calls ParseForm automatically.
// // ParseForm is idempotent.
// func (r *Request) ParseForm() error {
// 	var err error
// 	if r.PostForm == nil {
// 		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
// 			r.PostForm, err = parsePostForm(r)
// 		}
// 		if r.PostForm == nil {
// 			r.PostForm = make(url.Values)
// 		}
// 	}
// 	if r.Form == nil {
// 		if len(r.PostForm) > 0 {
// 			r.Form = make(url.Values)
// 			copyValues(r.Form, r.PostForm)
// 		}
// 		var newValues url.Values
// 		if r.URL != nil {
// 			var e error
// 			newValues, e = url.ParseQuery(r.URL.RawQuery)
// 			if err == nil {
// 				err = e
// 			}
// 		}
// 		if newValues == nil {
// 			newValues = make(url.Values)
// 		}
// 		if r.Form == nil {
// 			r.Form = newValues
// 		} else {
// 			copyValues(r.Form, newValues)
// 		}
// 	}
// 	return err
// }

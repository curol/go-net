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
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
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
	Proto      string
	ProtoMinor int
	ProtoMajor int

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

	// Header contains the request header fields either received
	// by the server or to be sent by the client.
	//
	// If a server received a request with header lines,
	//
	//	Host: example.com
	//	accept-encoding: gzip, deflate
	//	Accept-Language: en-us
	//	fOO: Bar
	//	foo: two
	//
	// then
	//
	//	Header = map[string][]string{
	//		"Accept-Encoding": {"gzip, deflate"},
	//		"Accept-Language": {"en-us"},
	//		"Foo": {"Bar", "two"},
	//	}
	//
	// For incoming requests, the Host header is promoted to the
	// Request.Host field and removed from the Header map.
	//
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

	// PostForm contains the parsed form data from PATCH, POST
	// or PUT body parameters.
	//
	// This field is only available after ParseForm is called.
	// The HTTP client ignores PostForm and uses Body instead.
	PostForm url.Values

	// TLS is for servers not clients that records the TLS information from a TLS connection
	TLS *tls.ConnectionState

	// Additional headers sent after the request body.
	Trailer Header

	// TODO: Add misc fields?
	// The following fields are for requests matched by ServeMux.
	// pat         *pattern          // the pattern that matched
	matches     []string          // values for the matching wildcards in pat
	otherValues map[string]string // for calls to SetPathValue that don't match a wildcard
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
	if method == "" {
		method = "GET"
	}
	// 1. Parse request Line
	method = strings.TrimSpace(strings.ToUpper(method))
	address = strings.TrimSpace(address)
	prot = strings.TrimSpace(prot)
	major, minor, ok := ParseHTTPVersion(prot) // parse protocol
	if !ok {
		return nil, fmt.Errorf("invalid protocol")
	}

	// 2. Parse URL
	rawurl := address
	if !strings.Contains(rawurl, "://") { // add scheme if missing
		rawurl = "http://" + rawurl
	}
	u, err := parseURL(rawurl) // parse url
	if err != nil {            // TODO: Handle error
		panic(err)
	}
	u.Host = removeEmptyPort(u.Host) // the host's colon:port should be normalized. See Issue 14836.

	// 3. Set Headers
	// TODO: Which headers should be set by default for client request?
	if header == nil {
		header = make(map[string][]string)
	}
	h := Header(header)
	h.Set("User-Agent", defaultUserAgent)
	h.Set("Host", u.Host)

	// 4. Set Body
	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = io.NopCloser(body)
	}

	// 5. Validate
	// TODO: Validate fields

	// 6. Set new Request
	req := &Request{
		// Request line
		Method:     method,
		Proto:      protocol,
		ProtoMajor: major,
		ProtoMinor: minor,
		// RequestURI: "", // Don't set RequestURI for client requests
		URL:  u,
		Host: u.Host, // TODO: Or use host header? Ex. host: getHost(h)
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
	if w == nil {
		return errors.New("http: nil Writer")
	}
	// Note:
	// - 'fmt.Fprintf' writes the string and automaitcally flushes the buffer
	// - 'w.Write' writes the string but does not flush the buffer
	// - 'w.Flush' flushes the buffer

	// 1.) Wrap the writer in a bufio Writer if it's not already buffered.
	// Don't always call NewWriter, as that forces a bytes.Buffer
	// and other small bufio Writers to have a minimum 4k buffer
	// size.

	// var bw *bufio.Writer
	// if _, ok := w.(*bufio.Writer); !ok {
	// 	bw = bufio.NewWriter(w)
	// } else {
	// 	bw = w.(*bufio.Writer)
	// var bw *bufio.Writer
	// if _, ok := w.(io.ByteWriter); !ok { // check if writer is bufferd
	// 	bw = bufio.NewWriter(w)
	// 	w = bw
	// }
	switch v := w.(type) {
	case *bufio.Writer:
		return r.write(v)
	default:
		return r.write(bufio.NewWriter(w))
	}
}

// write serializes r to w.
func (r *Request) write(w *bufio.Writer) error {
	// 1. Serialize and write the request line
	_, err := fmt.Fprintf(w, "%s %s %s\r\n", r.Method, r.URL.RequestURI(), r.Proto)
	if err != nil {
		return err
	}

	// 2. Serialize and write the headers
	// TODO: Write Transfer-Encoding, Trailer, and other headers
	// Find the target host. Prefer the Host: header, but if that
	// is not given, use the host from the request URL.
	var errMissingHost = errors.New("http: Request.Write on Request with no Host or URL set")
	host := r.Host
	if host == "" {
		if r.URL == nil {
			return errMissingHost
		}
		host = r.URL.Host
	}
	host = timeformat.RemoveZone(host)
	// Validate host
	if !ValidHostHeader(host) {
		// Validate that the Host header is a valid header in general,
		// but don't validate the host itself. This is sufficient to avoid
		// header or request smuggling via the Host field.
		// The server can (and will, if it's a net/http server) reject
		// the request if it doesn't consider the host valid.
		//
		// Historically, we would truncate the Host header after '/' or ' '.
		// Some users have relied on this truncation to convert a network
		// address such as Unix domain socket path into a valid, ignored
		// Host header (see https://go.dev/issue/61431).
		//
		// We don't preserve the truncation, because sending an altered
		// header field opens a smuggling vector. Instead, zero out the
		// Host header entirely if it isn't valid. (An empty Host is valid;
		// see RFC 9112 Section 3.2.)
		//
		// Return an error if we're sending to a proxy, since the proxy
		// probably can't do anything useful with an empty Host header.
		host = ""
	}
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
	_, err = fmt.Fprintf(w, "\r\n")
	if err != nil {
		return err
	}

	// 5. Flush first line and headers
	// err = w.Flush()
	// if err != nil {
	// 	return err
	// }

	// 5. Write body
	if r.Body != nil {
		// TODO: Check content length not larger than max memory
		_, err := io.CopyN(w, r.Body, r.ContentLength) // write body to w
		if err != nil {
			return err
		}
	}

	// // 6. Flush body
	// if bw != nil {
	// 	return bw.Flush()
	// }
	return nil
}

func (r *Request) Clone() *Request {
	clone := new(Request)
	clone.Method = r.Method
	clone.Proto = r.Proto
	clone.ProtoMajor = r.ProtoMajor
	clone.ProtoMinor = r.ProtoMinor
	clone.RequestURI = r.RequestURI
	clone.URL = r.URL
	clone.Header = r.Header.Clone()
	clone.Body = r.Body
	clone.ContentLength = r.ContentLength
	clone.ContentType = r.ContentType
	clone.RemoteAddress = r.RemoteAddress
	clone.GetBody = r.GetBody
	clone.Form = r.Form
	clone.MultipartForm = r.MultipartForm
	clone.TLS = r.TLS
	clone.Trailer = r.Trailer.Clone()
	clone.TransferEncoding = r.TransferEncoding
	clone.Close = r.Close
	return clone
}

// Reset resets the Request.
func (p *Request) Reset() {
	p = new(Request)
}

// UserAgent returns the client's User-Agent, if sent in the request.
func (r *Request) UserAgent() string {
	return r.Header.Get("User-Agent")
}

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

// Cookies parses and returns the HTTP cookies sent with the request.
func (r *Request) Cookies() []*Cookie {
	return readCookies(r.Header, "")
}

// AddCookie adds a cookie to the request. Per RFC 6265 section 5.4,
// AddCookie does not attach more than one [Cookie] header field. That
// means all cookies, if any, are written into the same line,
// separated by semicolon.
// AddCookie only sanitizes c's name and value, and does not sanitize
// a Cookie header already present in the request.
func (r *Request) AddCookie(c *Cookie) {
	if c == nil || (c.Name == "" && c.Value == "") {
		return
	}
	s := fmt.Sprintf("%s=%s", sanitizeCookieName(c.Name), sanitizeCookieValue(c.Value))
	if c := r.Header.Get("Cookie"); c != "" {
		r.Header.Set("Cookie", c+"; "+s)
	} else {
		r.Header.Set("Cookie", s)
	}
}

func (r *Request) FormValue(key string) string {
	if r.Form == nil {
		// TODO: Implement parseForm
		r.ParseForm()
	}
	return r.Form.Get(key)
}

// ToMap returns a map of the request's fields
func (r *Request) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Method":           r.Method,
		"URL":              r.URL.String(),
		"Proto":            r.Proto,
		"ProtoMajor":       r.ProtoMajor,
		"ProtoMinor":       r.ProtoMinor,
		"RequestURI":       r.RequestURI,
		"Host":             r.Host,
		"Header":           r.Header,
		"ContentLength":    r.ContentLength,
		"ContentType":      r.ContentType,
		"RemoteAddress":    r.RemoteAddress,
		"Form":             r.Form,
		"MultipartForm":    r.MultipartForm,
		"TLS":              r.TLS,
		"Trailer":          r.Trailer,
		"TransferEncoding": r.TransferEncoding,
		"Body":             r.Body,
	}
}

// Serialize the request line
func (r *Request) RequestLine() string {
	return r.Method + " " + r.URL.RequestURI() + " " + r.Proto
}

// Serialize the headers
func (r *Request) Headers() (string, error) {
	b := new(bytes.Buffer)
	err := r.Header.Write(b)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// Head serializes r without body in wire format
func (r *Request) Head() ([]byte, error) {
	rl := r.RequestLine()
	h, err := r.Headers()
	if err != nil {
		return nil, err
	}
	s := strings.Join([]string{rl, h, ""}, "\r\n")
	return []byte(s), nil
}

// Merge r with other
func (r *Request) Merge(other *Request) {
	if r.Method == "" {
		r.Method = other.Method
	}
	if r.Proto == "" {
		r.Proto = other.Proto
	}
	if r.ProtoMinor == 0 {
		r.ProtoMinor = other.ProtoMinor
	}
	if r.ProtoMajor == 0 {
		r.ProtoMajor = other.ProtoMajor
	}
	if r.RequestURI == "" {
		r.RequestURI = other.RequestURI
	}
	if r.URL == nil {
		r.URL = other.URL
	}
	if r.Host == "" {
		r.Host = other.Host
	}
	if r.Header == nil {
		r.Header = other.Header
	}
	if r.Body == nil {
		r.Body = other.Body
	}
	if r.ContentLength == 0 {
		r.ContentLength = other.ContentLength
	}
	if r.ContentType == "" {
		r.ContentType = other.ContentType
	}
	if r.RemoteAddress == "" {
		r.RemoteAddress = other.RemoteAddress
	}
	if r.GetBody == nil {
		r.GetBody = other.GetBody
	}
	if r.Form == nil {
		r.Form = other.Form
	}
	if r.MultipartForm == nil {
		r.MultipartForm = other.MultipartForm
	}
	if r.TLS == nil {
		r.TLS = other.TLS
	}
	if r.Trailer == nil {
		r.Trailer = other.Trailer
	}
}

// Debug prints friendly representation of the request for debugging
func (r *Request) Debug() string {
	// Fields
	lines := []string{
		"Method: " + r.Method,
		"URL: " + r.URL.String(),
		"Proto: " + r.Proto,
		"ProtoMajor: " + fmt.Sprint(r.ProtoMajor),
		"ProtoMinor: " + fmt.Sprint(r.ProtoMinor),
		"RequestURI: " + r.RequestURI,
		"Host: " + r.Host,
		"Header: " + fmt.Sprint(r.Header),
		"ContentLength: " + fmt.Sprint(r.ContentLength),
		"ContentType: " + r.ContentType,
		"RemoteAddress: " + r.RemoteAddress,
		"Form: " + fmt.Sprint(r.Form),
		"MultipartForm: " + fmt.Sprint(r.MultipartForm),
		"TLS: " + fmt.Sprint(r.TLS),
		"Trailer: " + fmt.Sprint(r.Trailer),
		"TransferEncoding: " + fmt.Sprint(r.TransferEncoding),
		"Body: " + fmt.Sprint(r.Body),
	}
	// If field is a string and empty, replace with escaped quotes
	for i, line := range lines {
		sp := strings.Split(line, ": ")
		if len(sp) != 2 {
			panic(fmt.Sprintf("invalid line %d: %s", i, line))
		}
		if sp[1] == "" {
			replacedStr := strings.Replace(sp[1], "", "\"\"", -1)
			lines[i] = sp[0] + ": " + replacedStr
		}
	}
	// Prefix each line with a tab
	for i, line := range lines {
		lines[i] = "\t" + line
	}
	// Join
	s := strings.Join(lines, "\n")
	startLine := fmt.Sprintf("\n&Request(%p){\n", &r)
	endLine := "\n}\n"
	// Return string representation of the request
	return startLine + s + endLine
}

// Dump serializes the request to wire format (raw http request) for debugging
func (r *Request) Dump() string {
	head, _ := r.Head()
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, r.Body)
	if err != nil {
		if err != io.EOF {
			return ""
		}
	}
	return string(append(head, buf.Bytes()...))
}

// TODO: Implement parseForm
// ParseForm populates r.Form and r.PostForm.
//
// For all requests, ParseForm parses the raw query from the URL and updates
// r.Form.
//
// For POST, PUT, and PATCH requests, it also reads the request body, parses it
// as a form and puts the results into both r.PostForm and r.Form. Request body
// parameters take precedence over URL query string values in r.Form.
//
// If the request Body's size has not already been limited by MaxBytesReader,
// the size is capped at 10MB.
//
// For other HTTP methods, or when the Content-Type is not
// application/x-www-form-urlencoded, the request Body is not read, and
// r.PostForm is initialized to a non-nil, empty value.
//
// ParseMultipartForm calls ParseForm automatically.
// ParseForm is idempotent.
func (r *Request) ParseForm() error {
	var err error
	if r.PostForm == nil {
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			r.PostForm, err = parsePostForm(r)
		}
		if r.PostForm == nil {
			r.PostForm = make(url.Values)
		}
	}
	if r.Form == nil {
		if len(r.PostForm) > 0 {
			r.Form = make(url.Values)
			copyValues(r.Form, r.PostForm)
		}
		var newValues url.Values
		if r.URL != nil {
			var e error
			newValues, e = url.ParseQuery(r.URL.RawQuery)
			if err == nil {
				err = e
			}
		}
		if newValues == nil {
			newValues = make(url.Values)
		}
		if r.Form == nil {
			r.Form = newValues
		} else {
			copyValues(r.Form, newValues)
		}
	}
	return err
}

// PostFormValue returns the first value for the named component of the POST,
// PUT, or PATCH request body. URL query parameters are ignored.
// PostFormValue calls [Request.ParseMultipartForm] and [Request.ParseForm] if necessary and ignores
// any errors returned by these functions.
// If key is not present, PostFormValue returns the empty string.
func (r *Request) PostFormValue(key string) string {
	if r.PostForm == nil {
		r.ParseMultipartForm(defaultMaxMemory)
	}
	if vs := r.PostForm[key]; len(vs) > 0 {
		return vs[0]
	}
	return ""
}

// multipartByReader is a sentinel value.
// Its presence in Request.MultipartForm indicates that parsing of the request
// body has been handed off to a MultipartReader instead of ParseMultipartForm.
var multipartByReader = &multipart.Form{
	Value: make(map[string][]string),
	File:  make(map[string][]*multipart.FileHeader),
}

// ParseMultipartForm parses a request body as multipart/form-data.
// The whole request body is parsed and up to a total of maxMemory bytes of
// its file parts are stored in memory, with the remainder stored on
// disk in temporary files.
// ParseMultipartForm calls [Request.ParseForm] if necessary.
// If ParseForm returns an error, ParseMultipartForm returns it but also
// continues parsing the request body.
// After one call to ParseMultipartForm, subsequent calls have no effect.
func (r *Request) ParseMultipartForm(maxMemory int64) error {
	if r.MultipartForm == multipartByReader {
		return errors.New("http: multipart handled by MultipartReader")
	}
	var parseFormErr error
	if r.Form == nil {
		// Let errors in ParseForm fall through, and just
		// return it at the end.
		parseFormErr = r.ParseForm()
	}
	if r.MultipartForm != nil {
		return nil
	}

	mr, err := r.multipartReader(false)
	if err != nil {
		return err
	}

	f, err := mr.ReadForm(maxMemory)
	if err != nil {
		return err
	}

	if r.PostForm == nil {
		r.PostForm = make(url.Values)
	}
	for k, v := range f.Value {
		r.Form[k] = append(r.Form[k], v...)
		// r.PostForm should also be populated. See Issue 9305.
		r.PostForm[k] = append(r.PostForm[k], v...)
	}

	r.MultipartForm = f

	return parseFormErr
}

// MultipartReader returns a MIME multipart reader if this is a
// multipart/form-data or a multipart/mixed POST request, else returns nil and an error.
// Use this function instead of [Request.ParseMultipartForm] to
// process the request body as a stream.
func (r *Request) MultipartReader() (*multipart.Reader, error) {
	if r.MultipartForm == multipartByReader {
		return nil, errors.New("http: MultipartReader called twice")
	}
	if r.MultipartForm != nil {
		return nil, errors.New("http: multipart handled by ParseMultipartForm")
	}
	r.MultipartForm = multipartByReader
	return r.multipartReader(true)
}

func (r *Request) multipartReader(allowMixed bool) (*multipart.Reader, error) {
	v := r.Header.Get("Content-Type")
	if v == "" {
		return nil, ErrNotMultipart
	}
	if r.Body == nil {
		return nil, errors.New("missing form body")
	}
	d, params, err := mime.ParseMediaType(v)
	if err != nil || !(d == "multipart/form-data" || allowMixed && d == "multipart/mixed") {
		return nil, ErrNotMultipart
	}
	boundary, ok := params["boundary"]
	if !ok {
		return nil, ErrMissingBoundary
	}
	return multipart.NewReader(r.Body, boundary), nil
}

// FormFile returns the first file for the provided form key.
// FormFile calls [Request.ParseMultipartForm] and [Request.ParseForm] if necessary.
func (r *Request) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	if r.MultipartForm == multipartByReader {
		return nil, nil, errors.New("http: multipart handled by MultipartReader")
	}
	if r.MultipartForm == nil {
		err := r.ParseMultipartForm(defaultMaxMemory)
		if err != nil {
			return nil, nil, err
		}
	}
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if fhs := r.MultipartForm.File[key]; len(fhs) > 0 {
			f, err := fhs[0].Open()
			return f, fhs[0], err
		}
	}
	return nil, nil, ErrMissingFile
}

// parseContentType detects the content type in the first 512 bytes of data for the MIME type.
func (r *Request) parseContentType(b []byte) {
	ct := SniffContentType(b)
	// TODO: Finish parsing content type
	r.Header.Set("Content-Type", ct)
	r.ContentType = ct
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
	defer func() {                  // clean up function
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
	if !strings.Contains(rawurl, "://") { // add scheme if missing
		rawurl = "http://" + rawurl
	}
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

	// 3. Set Request
	req := &Request{
		Method:        method,
		Proto:         prot,
		RequestURI:    requestURI,
		URL:           u,
		Host:          u.Host,
		Header:        header,
		ContentLength: getContentLength(header),
		ContentType:   header.Get("Content-Type"),
		Body:          io.NopCloser(r),
		Form:          nil,
		MultipartForm: nil,
		RemoteAddress: "",
	}

	// TODO: Sniff the content type (MIME type) from first 512 bytes of body?

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

// func (r *Request) Cookie(name string) (*Cookie, error) {
// 	if name == "" {
// 		return nil, ErrNoCookie
// 	}
// 	for _, c := range readCookies(r.Header, name) {
// 		return c, nil
// 	}
// 	return nil, ErrNoCookie
// }

// func (r *Request) Cookies() []*Cookie {
// 	return readCookies(r.Header, "")
// }

// func (r *Request) AddCookie(c *Cookie) {
// 	if c == nil || (c.Name == "" && c.Value == "") {
// 		return
// 	}
// 	s := fmt.Sprintf("%s=%s", sanitizeCookieName(c.Name), sanitizeCookieValue(c.Value))
// 	if c := r.Header.Get("Cookie"); c != "" {
// 		r.Header.Set("Cookie", c+"; "+s)
// 	} else {
// 		r.Header.Set("Cookie", s)
// 	}
// }

// func (r *Request) FormValue(key string) string {
// 	if r.Form == nil {
// 		r.ParseForm()
// 	}
// 	return r.Form.Get(key)
// }

// func (r *Request) PostFormValue(key string) string {
// 	if r.PostForm == nil {
// 		r.ParseForm()
// 	}
// 	return r.PostForm.Get(key)
// }

// func (r *Request) MultipartReader() (*multipart.Reader, error) {
// 	if r.MultipartForm == multipartByReader {
// 		return nil, errors.New("http: MultipartReader called twice")
// 	}
// 	if r.MultipartForm != nil {
// 		return nil, errors.New("http: multipart handled by ParseMultipartForm")
// 	}
// 	r.MultipartForm = multipartByReader
// 	return r.multipartReader(true)
// }

// func (r *Request) ParseMultipartForm(maxMemory int64) error {
// 	if r.MultipartForm == multipartByReader {
// 		return errors.New("http: multipart handled by MultipartReader")
// 	}
// 	var parseFormErr error
// 	if r.Form == nil {
// 		parseFormErr = r.ParseForm()
// 	}
// 	if r.MultipartForm != nil {
// 		return nil
// 	}
// 	mr, err := r.multipartReader(false)
// 	if err != nil {
// 		return err
// 	}
// 	f, err := mr.ReadForm(maxMemory)
// 	if err != nil {
// 		return err
// 	}
// 	if r.PostForm == nil {
// 		r.PostForm = make(url.Values)
// 	}
// 	for k, v := range f.Value {
// 		r.Form[k] = append(r.Form[k], v...)
// 		r.PostForm[k] = append(r.PostForm[k], v...)
// 	}
// 	r.MultipartForm = f
// 	return parseFormErr
// }

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
// 			newValues, err = url.ParseQuery(r.URL.RawQuery)
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

// func (r *Request) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
// 	if r.MultipartForm == multipartByReader {
// 		return nil, nil, errors.New("http: multipart handled by MultipartReader")
// 	}
// 	if r.MultipartForm == nil {
// 		err := r.ParseMultipartForm(defaultMaxMemory)
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 	}
// 	if r.MultipartForm != nil && r.MultipartForm.File != nil {
// 		if fhs := r.MultipartForm.File[key]; len(fhs) > 0 {
// 			f, err := fhs[0].Open()
// 			return f, fhs[0], err
// 		}
// 	}
// 	return nil, nil, ErrMissingFile
// }

package http

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"path"

	"github.com/curol/network/http/internal/ascii"
	"github.com/curol/network/http/internal/token"
	"github.com/curol/network/url"
	"github.com/duke-git/lancet/v2/netutil"
)

func validMethod(method string) bool {
	/*
	     Method         = "OPTIONS"                ; Section 9.2
	                    | "GET"                    ; Section 9.3
	                    | "HEAD"                   ; Section 9.4
	                    | "POST"                   ; Section 9.5
	                    | "PUT"                    ; Section 9.6
	                    | "DELETE"                 ; Section 9.7
	                    | "TRACE"                  ; Section 9.8
	                    | "CONNECT"                ; Section 9.9
	                    | extension-method
	   extension-method = token
	     token          = 1*<any CHAR except CTLs or separators>
	*/
	return len(method) > 0 && strings.IndexFunc(method, isNotToken) == -1
}

// hasToken reports whether token appears with v, ASCII
// case-insensitive, with space or comma boundaries.
// token must be all lowercase.
// v may contain mixed cased.
func hasToken(v, token string) bool {
	if len(token) > len(v) || token == "" {
		return false
	}
	if v == token {
		return true
	}
	for sp := 0; sp <= len(v)-len(token); sp++ {
		// Check that first character is good.
		// The token is ASCII, so checking only a single byte
		// is sufficient. We skip this potential starting
		// position if both the first byte and its potential
		// ASCII uppercase equivalent (b|0x20) don't match.
		// False positives ('^' => '~') are caught by EqualFold.
		if b := v[sp]; b != token[0] && b|0x20 != token[0] {
			continue
		}
		// Check that start pos is on a valid token boundary.
		if sp > 0 && !isTokenBoundary(v[sp-1]) {
			continue
		}
		// Check that end pos is on a valid token boundary.
		if endPos := sp + len(token); endPos != len(v) && !isTokenBoundary(v[endPos]) {
			continue
		}
		if ascii.EqualFold(v[sp:sp+len(token)], token) {
			return true
		}
	}
	return false
}

func isTokenBoundary(b byte) bool {
	return b == ' ' || b == ',' || b == '\t'
}

func getSetCookies(h Header) []*Cookie {
	// Original code:
	// cookies := make([]*Cookie, 0)
	// for k, v := range h {
	// 	// Response header 'Set-Cookie'
	// 	if k == "Set-Cookie" {
	// 		cookie := parseCookie(v)
	// 		cookies = append(cookies, cookie)
	// 	}
	// }
	// return cookies
	return readSetCookies(h)
}

func isNotToken(r rune) bool {
	return !token.IsTokenRune(r)
}

// removeEmptyPort strips the empty port in ":port" to ""
// as mandated by RFC 3986 Section 6.2.3.
func removeEmptyPort(host string) string {
	if hasPort(host) {
		return strings.TrimSuffix(host, ":")
	}
	return host
}

// Given a string of the form "host", "host:port", or "[ipv6::address]:port",
// return true if the string includes a port.
func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

func getContentLength(header Header) int64 {
	if header == nil {
		return 0
	}
	cl := header.Get("Content-Length")
	if cl == "" {
		return 0
	}
	v, err := strconv.Atoi(cl)
	if err != nil {
		return 0
	}
	return int64(v)
}

func addSchemeIfMissing(rawurl string) (string, error) {
	// Add a scheme if it's missing
	if !strings.Contains(rawurl, "://") {
		rawurl = "http://" + rawurl
	}
	return rawurl, nil
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

func parseURL(rawurl string) (*url.URL, error) {
	parsedURL, err := url.Parse(rawurl) // parse url
	if err != nil {
		return nil, err
	}
	return parsedURL, nil
}

func parsePostForm(r *Request) (vs url.Values, err error) {
	if r.Body == nil {
		err = errors.New("missing form body")
		return
	}
	ct := r.Header.Get("Content-Type")
	// RFC 7231, section 3.1.1.5 - empty type
	//   MAY be treated as application/octet-stream
	if ct == "" {
		ct = "application/octet-stream"
	}
	ct, _, err = mime.ParseMediaType(ct)
	switch {
	case ct == "application/x-www-form-urlencoded":
		var reader io.Reader = r.Body
		maxFormSize := int64(1<<63 - 1)
		if _, ok := r.Body.(*maxBytesReader); !ok {
			maxFormSize = int64(10 << 20) // 10 MB is a lot of text.
			reader = io.LimitReader(r.Body, maxFormSize+1)
		}
		b, e := io.ReadAll(reader)
		if e != nil {
			if err == nil {
				err = e
			}
			break
		}
		if int64(len(b)) > maxFormSize {
			err = errors.New("http: POST too large")
			return
		}
		vs, e = url.ParseQuery(string(b))
		if err == nil {
			err = e
		}
	case ct == "multipart/form-data":
		// handled by ParseMultipartForm (which is calling us, or should be)
		// TODO(bradfitz): there are too many possible
		// orders to call too many functions here.
		// Clean this up and write more tests.
		// request_test.go contains the start of this,
		// in TestParseMultipartFormOrder and others.
	}
	return
}

func copyValues(dst, src url.Values) {
	for k, vs := range src {
		dst[k] = append(dst[k], vs...)
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

// RequestFromMap creates an [http.Request] from CGI variables.
// The returned Request's Body field is not populated.
func requestFromMap(params map[string]string) (*Request, error) {
	r := new(Request)
	r.Method = params["REQUEST_METHOD"]
	if r.Method == "" {
		return nil, errors.New("cgi: no REQUEST_METHOD in environment")
	}

	r.Proto = params["SERVER_PROTOCOL"]
	var ok bool
	r.ProtoMajor, r.ProtoMinor, ok = ParseHTTPVersion(r.Proto)
	if !ok {
		return nil, errors.New("cgi: invalid SERVER_PROTOCOL version")
	}

	r.Close = true
	r.Trailer = http.Header{}
	r.Header = http.Header{}

	r.Host = params["HTTP_HOST"]

	if lenstr := params["CONTENT_LENGTH"]; lenstr != "" {
		clen, err := strconv.ParseInt(lenstr, 10, 64)
		if err != nil {
			return nil, errors.New("cgi: bad CONTENT_LENGTH in environment: " + lenstr)
		}
		r.ContentLength = clen
	}

	if ct := params["CONTENT_TYPE"]; ct != "" {
		r.Header.Set("Content-Type", ct)
	}

	// Copy "HTTP_FOO_BAR" variables to "Foo-Bar" Headers
	for k, v := range params {
		if k == "HTTP_HOST" {
			continue
		}
		if after, found := strings.CutPrefix(k, "HTTP_"); found {
			r.Header.Add(strings.ReplaceAll(after, "_", "-"), v)
		}
	}

	uriStr := params["REQUEST_URI"]
	if uriStr == "" {
		// Fallback to SCRIPT_NAME, PATH_INFO and QUERY_STRING.
		uriStr = params["SCRIPT_NAME"] + params["PATH_INFO"]
		s := params["QUERY_STRING"]
		if s != "" {
			uriStr += "?" + s
		}
	}

	// There's apparently a de-facto standard for this.
	// https://web.archive.org/web/20170105004655/http://docstore.mik.ua/orelly/linux/cgi/ch03_02.htm#ch03-35636
	if s := params["HTTPS"]; s == "on" || s == "ON" || s == "1" {
		r.TLS = &tls.ConnectionState{HandshakeComplete: true}
	}

	if r.Host != "" {
		// Hostname is provided, so we can reasonably construct a URL.
		rawurl := r.Host + uriStr
		if r.TLS == nil {
			rawurl = "http://" + rawurl
		} else {
			rawurl = "https://" + rawurl
		}
		url, err := url.Parse(rawurl)
		if err != nil {
			return nil, errors.New("cgi: failed to parse host and REQUEST_URI into a URL: " + rawurl)
		}
		r.URL = url
	}
	// Fallback logic if we don't have a Host header or the URL
	// failed to parse
	if r.URL == nil {
		url, err := url.Parse(uriStr)
		if err != nil {
			return nil, errors.New("cgi: failed to parse REQUEST_URI into a URL: " + uriStr)
		}
		r.URL = url
	}

	// Request.RemoteAddr has its port set by Go's standard http
	// server, so we do here too.
	remotePort, _ := strconv.Atoi(params["REMOTE_PORT"]) // zero if unset or invalid
	r.RemoteAddress = net.JoinHostPort(params["REMOTE_ADDR"], strconv.Itoa(remotePort))

	return r, nil
}

// cleanPath returns the canonical path for p, eliminating . and .. elements.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		// Fast path for common case of p being the string we want:
		if len(p) == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}
	return np
}

var errInvalidPath = errors.New("invalid path")

// FromFS converts a slash-separated path into an operating-system path.
//
// FromFS returns an error if the path cannot be represented by the operating
// system. For example, paths containing '\' and ':' characters are rejected
// on Windows.
func fromFS(path string) (string, error) {
	if path == "" {
		return "", errInvalidPath
	}
	// TODO: check for invalid characters?
	// if !filepath.IsAbs(s) {       // reject relative paths
	// 	return "", errInvalidPath
	// }
	s := filepath.FromSlash(path) // clean and convert slashes
	return s, nil
}

func isAbsolutePath(path string) bool {
	// return path != "" && path[0] == '/'
	return filepath.IsAbs(path)
}

func isRelativePath(path string) bool {
	return !isAbsolutePath(path)
}

func getFullPath(elem ...string) string {
	return filepath.Join(elem...)
}

var structToURLValues = netutil.StructToUrlValues

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
	"unicode/utf8"

	"path"

	"github.com/curol/network/http/internal/ascii"
	"github.com/curol/network/url"
	"github.com/duke-git/lancet/v2/netutil"
	"golang.org/x/net/idna"
)

var structToURLValues = netutil.StructToUrlValues

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

// func parsePostForm(r *Request) (vs url.Values, err error) {
// 	if r.Body == nil {
// 		err = errors.New("missing form body")
// 		return
// 	}
// 	ct := r.Header.Get("Content-Type")
// 	// RFC 7231, section 3.1.1.5 - empty type
// 	//   MAY be treated as application/octet-stream
// 	if ct == "" {
// 		ct = "application/octet-stream"
// 	}
// 	ct, _, err = mime.ParseMediaType(ct)
// 	switch {
// 	case ct == "application/x-www-form-urlencoded":
// 		var reader io.Reader = r.Body
// 		maxFormSize := int64(1<<63 - 1)
// 		if _, ok := r.Body.(*maxBytesReader); !ok {
// 			maxFormSize = int64(10 << 20) // 10 MB is a lot of text.
// 			reader = io.LimitReader(r.Body, maxFormSize+1)
// 		}
// 		b, e := io.ReadAll(reader)
// 		if e != nil {
// 			if err == nil {
// 				err = e
// 			}
// 			break
// 		}
// 		if int64(len(b)) > maxFormSize {
// 			err = errors.New("http: POST too large")
// 			return
// 		}
// 		vs, e = url.ParseQuery(string(b))
// 		if err == nil {
// 			err = e
// 		}
// 	case ct == "multipart/form-data":
// 		// handled by ParseMultipartForm (which is calling us, or should be)
// 		// TODO(bradfitz): there are too many possible
// 		// orders to call too many functions here.
// 		// Clean this up and write more tests.
// 		// request_test.go contains the start of this,
// 		// in TestParseMultipartFormOrder and others.
// 	}
// 	return
// }

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

// StripPrefix returns a handler that serves HTTP requests by removing the
// given prefix from the request URL's Path (and RawPath if set) and invoking
// the handler h. StripPrefix handles a request for a path that doesn't begin
// with prefix by replying with an HTTP 404 not found error. The prefix must
// match exactly: if the prefix in the request contains escaped characters
// the reply is also an HTTP 404 not found error.
func StripPrefix(prefix string, h Handler) Handler {
	if prefix == "" {
		return h
	}
	return HandlerFunc(func(w ResponseWriter, r *Request) {
		p := strings.TrimPrefix(r.URL.Path, prefix)
		rp := strings.TrimPrefix(r.URL.RawPath, prefix)
		if len(p) < len(r.URL.Path) && (r.URL.RawPath == "" || len(rp) < len(r.URL.RawPath)) {
			r2 := new(Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = p
			r2.URL.RawPath = rp
			h.ServeHTTP(w, r2)
		} else {
			NotFound(w, r)
		}
	})
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

// stripHostPort returns h without any trailing ":<port>".
func stripHostPort(h string) string {
	// If no port on host, return unchanged
	if !strings.Contains(h, ":") {
		return h
	}
	host, _, err := net.SplitHostPort(h)
	if err != nil {
		return h // on error, return unchanged
	}
	return host
}

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

func parseURL(rawurl string) (*url.URL, error) {
	return url.Parse(rawurl)
}

var isTokenTable = [127]bool{
	'!':  true,
	'#':  true,
	'$':  true,
	'%':  true,
	'&':  true,
	'\'': true,
	'*':  true,
	'+':  true,
	'-':  true,
	'.':  true,
	'0':  true,
	'1':  true,
	'2':  true,
	'3':  true,
	'4':  true,
	'5':  true,
	'6':  true,
	'7':  true,
	'8':  true,
	'9':  true,
	'A':  true,
	'B':  true,
	'C':  true,
	'D':  true,
	'E':  true,
	'F':  true,
	'G':  true,
	'H':  true,
	'I':  true,
	'J':  true,
	'K':  true,
	'L':  true,
	'M':  true,
	'N':  true,
	'O':  true,
	'P':  true,
	'Q':  true,
	'R':  true,
	'S':  true,
	'T':  true,
	'U':  true,
	'W':  true,
	'V':  true,
	'X':  true,
	'Y':  true,
	'Z':  true,
	'^':  true,
	'_':  true,
	'`':  true,
	'a':  true,
	'b':  true,
	'c':  true,
	'd':  true,
	'e':  true,
	'f':  true,
	'g':  true,
	'h':  true,
	'i':  true,
	'j':  true,
	'k':  true,
	'l':  true,
	'm':  true,
	'n':  true,
	'o':  true,
	'p':  true,
	'q':  true,
	'r':  true,
	's':  true,
	't':  true,
	'u':  true,
	'v':  true,
	'w':  true,
	'x':  true,
	'y':  true,
	'z':  true,
	'|':  true,
	'~':  true,
}

func IsTokenRune(r rune) bool {
	i := int(r)
	return i < len(isTokenTable) && isTokenTable[i]
}

func isNotToken(r rune) bool {
	return !IsTokenRune(r)
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

// HeaderValuesContainsToken reports whether any string in values
// contains the provided token, ASCII case-insensitively.
func HeaderValuesContainsToken(values []string, token string) bool {
	for _, v := range values {
		if headerValueContainsToken(v, token) {
			return true
		}
	}
	return false
}

// isOWS reports whether b is an optional whitespace byte, as defined
// by RFC 7230 section 3.2.3.
func isOWS(b byte) bool { return b == ' ' || b == '\t' }

// trimOWS returns x with all optional whitespace removes from the
// beginning and end.
func trimOWS(x string) string {
	// TODO: consider using strings.Trim(x, " \t") instead,
	// if and when it's fast enough. See issue 10292.
	// But this ASCII-only code will probably always beat UTF-8
	// aware code.
	for len(x) > 0 && isOWS(x[0]) {
		x = x[1:]
	}
	for len(x) > 0 && isOWS(x[len(x)-1]) {
		x = x[:len(x)-1]
	}
	return x
}

// headerValueContainsToken reports whether v (assumed to be a
// 0#element, in the ABNF extension described in RFC 7230 section 7)
// contains token amongst its comma-separated tokens, ASCII
// case-insensitively.
func headerValueContainsToken(v string, token string) bool {
	for comma := strings.IndexByte(v, ','); comma != -1; comma = strings.IndexByte(v, ',') {
		if tokenEqual(trimOWS(v[:comma]), token) {
			return true
		}
		v = v[comma+1:]
	}
	return tokenEqual(trimOWS(v), token)
}

// lowerASCII returns the ASCII lowercase version of b.
func lowerASCII(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

// tokenEqual reports whether t1 and t2 are equal, ASCII case-insensitively.
func tokenEqual(t1, t2 string) bool {
	if len(t1) != len(t2) {
		return false
	}
	for i, b := range t1 {
		if b >= utf8.RuneSelf {
			// No UTF-8 or non-ASCII allowed in tokens.
			return false
		}
		if lowerASCII(byte(b)) != lowerASCII(t2[i]) {
			return false
		}
	}
	return true
}

// isLWS reports whether b is linear white space, according
// to http://www.w3.org/Protocols/rfc2616/rfc2616-sec2.html#sec2.2
//
//	LWS            = [CRLF] 1*( SP | HT )
func isLWS(b byte) bool { return b == ' ' || b == '\t' }

// isCTL reports whether b is a control byte, according
// to http://www.w3.org/Protocols/rfc2616/rfc2616-sec2.html#sec2.2
//
//	CTL            = <any US-ASCII control character
//	                 (octets 0 - 31) and DEL (127)>
func isCTL(b byte) bool {
	const del = 0x7f // a CTL
	return b < ' ' || b == del
}

// ValidHeaderFieldName reports whether v is a valid HTTP/1.x header name.
// HTTP/2 imposes the additional restriction that uppercase ASCII
// letters are not allowed.
//
// RFC 7230 says:
//
//	header-field   = field-name ":" OWS field-value OWS
//	field-name     = token
//	token          = 1*tchar
//	tchar = "!" / "#" / "$" / "%" / "&" / "'" / "*" / "+" / "-" / "." /
//	        "^" / "_" / "`" / "|" / "~" / DIGIT / ALPHA
func ValidHeaderFieldName(v string) bool {
	if len(v) == 0 {
		return false
	}
	for _, r := range v {
		if !IsTokenRune(r) {
			return false
		}
	}
	return true
}

// ValidHostHeader reports whether h is a valid host header.
func ValidHostHeader(h string) bool {
	// The latest spec is actually this:
	//
	// http://tools.ietf.org/html/rfc7230#section-5.4
	//     Host = uri-host [ ":" port ]
	//
	// Where uri-host is:
	//     http://tools.ietf.org/html/rfc3986#section-3.2.2
	//
	// But we're going to be much more lenient for now and just
	// search for any byte that's not a valid byte in any of those
	// expressions.
	for i := 0; i < len(h); i++ {
		if !validHostByte[h[i]] {
			return false
		}
	}
	return true
}

// See the validHostHeader comment.
var validHostByte = [256]bool{
	'0': true, '1': true, '2': true, '3': true, '4': true, '5': true, '6': true, '7': true,
	'8': true, '9': true,

	'a': true, 'b': true, 'c': true, 'd': true, 'e': true, 'f': true, 'g': true, 'h': true,
	'i': true, 'j': true, 'k': true, 'l': true, 'm': true, 'n': true, 'o': true, 'p': true,
	'q': true, 'r': true, 's': true, 't': true, 'u': true, 'v': true, 'w': true, 'x': true,
	'y': true, 'z': true,

	'A': true, 'B': true, 'C': true, 'D': true, 'E': true, 'F': true, 'G': true, 'H': true,
	'I': true, 'J': true, 'K': true, 'L': true, 'M': true, 'N': true, 'O': true, 'P': true,
	'Q': true, 'R': true, 'S': true, 'T': true, 'U': true, 'V': true, 'W': true, 'X': true,
	'Y': true, 'Z': true,

	'!':  true, // sub-delims
	'$':  true, // sub-delims
	'%':  true, // pct-encoded (and used in IPv6 zones)
	'&':  true, // sub-delims
	'(':  true, // sub-delims
	')':  true, // sub-delims
	'*':  true, // sub-delims
	'+':  true, // sub-delims
	',':  true, // sub-delims
	'-':  true, // unreserved
	'.':  true, // unreserved
	':':  true, // IPv6address + Host expression's optional port
	';':  true, // sub-delims
	'=':  true, // sub-delims
	'[':  true,
	'\'': true, // sub-delims
	']':  true,
	'_':  true, // unreserved
	'~':  true, // unreserved
}

// ValidHeaderFieldValue reports whether v is a valid "field-value" according to
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.2 :
//
//	message-header = field-name ":" [ field-value ]
//	field-value    = *( field-content | LWS )
//	field-content  = <the OCTETs making up the field-value
//	                 and consisting of either *TEXT or combinations
//	                 of token, separators, and quoted-string>
//
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec2.html#sec2.2 :
//
//	TEXT           = <any OCTET except CTLs,
//	                  but including LWS>
//	LWS            = [CRLF] 1*( SP | HT )
//	CTL            = <any US-ASCII control character
//	                 (octets 0 - 31) and DEL (127)>
//
// RFC 7230 says:
//
//	field-value    = *( field-content / obs-fold )
//	obj-fold       =  N/A to http2, and deprecated
//	field-content  = field-vchar [ 1*( SP / HTAB ) field-vchar ]
//	field-vchar    = VCHAR / obs-text
//	obs-text       = %x80-FF
//	VCHAR          = "any visible [USASCII] character"
//
// http2 further says: "Similarly, HTTP/2 allows header field values
// that are not valid. While most of the values that can be encoded
// will not alter header field parsing, carriage return (CR, ASCII
// 0xd), line feed (LF, ASCII 0xa), and the zero character (NUL, ASCII
// 0x0) might be exploited by an attacker if they are translated
// verbatim. Any request or response that contains a character not
// permitted in a header field value MUST be treated as malformed
// (Section 8.1.2.6). Valid characters are defined by the
// field-content ABNF rule in Section 3.2 of [RFC7230]."
//
// This function does not (yet?) properly handle the rejection of
// strings that begin or end with SP or HTAB.
func ValidHeaderFieldValue(v string) bool {
	for i := 0; i < len(v); i++ {
		b := v[i]
		if isCTL(b) && !isLWS(b) {
			return false
		}
	}
	return true
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

// PunycodeHostPort returns the IDNA Punycode version
// of the provided "host" or "host:port" string.
func PunycodeHostPort(v string) (string, error) {
	if isASCII(v) {
		return v, nil
	}

	host, port, err := net.SplitHostPort(v)
	if err != nil {
		// The input 'v' argument was just a "host" argument,
		// without a port. This error should not be returned
		// to the caller.
		host = v
		port = ""
	}
	host, err = idna.ToASCII(host)
	if err != nil {
		// Non-UTF-8? Not representable in Punycode, in any
		// case.
		return "", err
	}
	if port == "" {
		return host, nil
	}
	return net.JoinHostPort(host, port), nil
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

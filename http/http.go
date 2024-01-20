package http

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/curol/network/http/internal/ascii"
	"github.com/curol/network/http/internal/token"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
	protocol         = "HTTP/1.1"
	defaultUserAgent = "curol-http-client/1.1"
)

// NoBody is an io.ReadCloser with no bytes. Read always returns EOF
// and Close always returns nil. It can be used in an outgoing client
// request to explicitly signal that a request has zero bytes.
// An alternative, however, is to simply set Request.Body to nil.
var NoBody = noBody{}

type noBody struct{}

func (noBody) Read([]byte) (int, error)         { return 0, io.EOF }
func (noBody) Close() error                     { return nil }
func (noBody) WriteTo(io.Writer) (int64, error) { return 0, nil }

var (
	// verify that an io.Copy from NoBody won't require a buffer:
	_ io.WriterTo   = NoBody
	_ io.ReadCloser = NoBody
)

// getHost returns the host as declared in the headers.
//
// RFC 7230, section 5.3: Must treat
//
//	GET /index.html HTTP/1.1
//	Host: www.google.com
//
// and
//
//	GET http://www.google.com/index.html HTTP/1.1
//	Host: doesntmatter
//
// the same.
// In the second case, any Host line is ignored.
func getHostForWriter(r *Request) (string, error) {
	// errMissingHost is returned by Write when there is no Host or URL present in
	// the Request.
	var errMissingHost = errors.New("http: Request.Write on Request with no Host or URL set")

	// Find the target host. Prefer the Host: header, but if that
	// is not given, use the host from the request URL.
	host := r.Host
	if host == "" {
		if r.URL == nil {
			return "", errMissingHost
		}
		host = r.URL.Host
	}
	host = removeZone(host)

	// TODO: Validate and clean host

	return host, nil
}

func getUserAgent(h Header) string {
	if h.Get("User-Agent") == "" {
		return defaultUserAgent
	}
	return h.Get("User-Agent")
}

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

func getContentType(header Header) string {
	ct := header.Get("Content-Type")
	if ct == "" {
		return ""
	}
	return ct
}

func getCookies(h Header) []*Cookie {
	// Original code:
	// for k, v := range h {
	// 	if k == "Cookie" {
	// 		cookie := parseCookie(v)
	// 		cookies = append(cookies, cookie)
	// 	}
	// }
	// return cookies
	return readCookies(h, "")
}

// removeZone removes IPv6 zone identifier from host.
// E.g., "[fe80::1%en0]:8080" to "[fe80::1]:8080"
func removeZone(host string) string {
	if !strings.HasPrefix(host, "[") {
		return host
	}
	i := strings.LastIndex(host, "]")
	if i < 0 {
		return host
	}
	j := strings.LastIndex(host[:i], "%")
	if j < 0 {
		return host
	}
	return host[:j] + host[i:]
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

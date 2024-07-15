package cookie

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

// A Cookie represents an HTTP cookie as sent in the Set-Cookie header of an
// HTTP response or the Cookie header of an HTTP request.
//
// See https://tools.ietf.org/html/rfc6265 for details.
type Cookie struct {
	Name   string
	Value  string
	Quoted bool // indicates whether the Value was originally quoted

	Path       string    // optional
	Domain     string    // optional
	Expires    time.Time // optional
	RawExpires string    // for reading cookies only

	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite SameSite
	Raw      string
	Unparsed []string // Raw text of unparsed attribute-value pairs
}

type CookieOptions struct {
	Path     string
	Domain   string
	Expires  time.Time
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite SameSite
	Raw      string
	Unparsed []string
}

func NewCookie(name string, value string, opts *CookieOptions) *Cookie {
	cookie := &Cookie{Name: name, Value: value, SameSite: SameSiteDefaultMode}
	if opts != nil {
		if opts.Path != "" {
			cookie.Path = opts.Path
		}
		if opts.Domain != "" {
			cookie.Domain = opts.Domain
		}
		if opts.Expires.IsZero() {
			cookie.Expires = opts.Expires
		}
		if opts.MaxAge != 0 {
			cookie.MaxAge = opts.MaxAge
		}
		if opts.Secure {
			cookie.Secure = opts.Secure
		}
		if opts.HttpOnly {
			cookie.HttpOnly = opts.HttpOnly
		}
		if opts.SameSite != SameSiteDefaultMode {
			cookie.SameSite = opts.SameSite
		}
		if opts.Raw != "" {
			cookie.Raw = opts.Raw
		}
		if len(opts.Unparsed) > 0 {
			cookie.Unparsed = opts.Unparsed
		}
	}
	if !cookie.isStringsValid() {
		return nil
	}
	return cookie
}

func (c *Cookie) String() string {
	return serializeCookie(c)
}

func (c *Cookie) StringWithoutAtrributes() string {
	return c.Name + "=" + c.Value
}

// SerializeCookie serializes a cookie to a string.
func serializeCookie(c *Cookie) string {
	name := c.Name
	val := c.Value
	// TODO: Add quotes to value if quoted?
	// if c.Quoted {
	// 	val = "\"" + val + "\""
	// }
	// Begin with name and value
	cookie := name + "=" + val
	// Attributes
	if c.Path != "" {
		cookie += "; Path=" + c.Path
	}
	if c.Domain != "" {
		cookie += "; Domain=" + c.Domain
	}
	if !c.Expires.IsZero() {
		cookie += "; Expires=" + c.Expires.Format(time.RFC1123)
	}
	if c.MaxAge > 0 {
		cookie += "; Max-Age=" + string(rune(c.MaxAge))
	}
	if c.Secure {
		cookie += "; Secure"
	}
	if c.HttpOnly {
		cookie += "; HttpOnly"
	}
	if c.SameSite != SameSiteDefaultMode {
		ssv := c.SameSite.String()
		if ssv != "" {
			cookie += "; SameSite=" + ssv
		}
	}
	return cookie
}

func (c *Cookie) isStringsValid() bool {
	if c.Name == "" || c.Value == "" {
		return false
	}
	if !isValidCookieName(c.Name) {
		return false
	}
	if !isValidCookieValue(c.Value) {
		return false
	}
	if c.Path != "" && !isValidCookieValue(c.Path) {
		return false
	}
	if c.Domain != "" && !isValidCookieValue(c.Domain) {
		return false
	}
	if c.RawExpires != "" && !isValidCookieValue(c.RawExpires) {
		return false
	}
	if c.Raw != "" && !isValidCookieValue(c.Raw) {
		return false
	}
	if len(c.Unparsed) > 0 {
		for _, unparsed := range c.Unparsed {
			if !isValidCookieValue(unparsed) {
				return false
			}
		}
	}
	return true
}

func (c *Cookie) IsExpired() bool {
	if c.MaxAge < 0 || c.Expires.Before(time.Now()) {
		return true
	}
	return false
}

// IsValid checks if cookie is ok
// func (c *Cookie) IsValid( ) bool {
// 	if c.Name == "" || c.Value == "" {
// 		return false
// 	}
// 	if !isValidCookieName(c.Name) {
// 		return false
// 	}
// 	if !isValidCookieValue(c.Value) {
// 		return false
// 	}
// 	if c.Path != "" && !isValidCookieValue(c.Path) {
// 		return false
// 	}
// 	if c.Domain != "" && !isValidCookieValue(c.Domain) {
// 		return false
// 	}
// 	if c.RawExpires != "" && !isValidCookieValue(c.RawExpires) {
// 		return false
// 	}
// 	if c.MaxAge != 0 {
// 		return false
// 	}
// 	if c.Raw != "" && !isValidCookieValue(c.Raw) {
// 		return false
// 	}
// 	if len(c.Unparsed) > 0 {
// 		for _, unparsed := range c.Unparsed {
// 			if !isValidCookieValue(unparsed) {
// 				return false
// 			}
// 		}
// 	}
// 	if c.SameSite != SameSiteDefaultMode {
// 		return false
// 	}
// 	if c.Expires.IsZero() {
// 		return false
// 	}
// 	if c.MaxAge < 0 {
// 		return false
// 	}

// }

// SameSite allows a server to define a cookie attribute making it impossible for
// the browser to send this cookie along with cross-site requests. The main
// goal is to mitigate the risk of cross-origin information leakage, and provide
// some protection against cross-site request forgery attacks.
//
// See https://tools.ietf.org/html/draft-ietf-httpbis-cookie-same-site-00 for details.
type SameSite int

const (
	SameSiteDefaultMode SameSite = iota + 1
	SameSiteLaxMode
	SameSiteStrictMode
	SameSiteNoneMode
)

var (
	errBlankCookie           = errors.New("http: blank cookie")
	errEqualNotFoundInCookie = errors.New("http: '=' not found in cookie")
	errInvalidCookieName     = errors.New("http: invalid cookie name")
	errInvalidCookieValue    = errors.New("http: invalid cookie value")
)

func (s SameSite) String() string {
	switch s {
	case SameSiteDefaultMode:
		return "Default"
	case SameSiteLaxMode:
		return "Lax"
	case SameSiteStrictMode:
		return "Strict"
	case SameSiteNoneMode:
		return "None"
	default:
		return ""
	}
}

// func ParseCookies(cookieHeader string) ([]*Cookie, error) {
// 	parts := strings.Split(cookieHeader, ";")
// 	cookies := make([]*Cookie, len(parts))
// }

// ParseCookie parses a Cookie header value and returns all the cookies
// which were set in it. Since the same cookie name can appear multiple times
// the returned Values can contain more than one value for a given key.
// Each cookie in the header is separated by a semicolon and a space (; ).
func ParseCookie(line string) ([]*Cookie, error) {
	parts := strings.Split(TrimString(line), ";")
	if len(parts) == 1 && parts[0] == "" {
		return nil, errBlankCookie
	}
	cookies := make([]*Cookie, 0, len(parts))
	for _, s := range parts {
		s = TrimString(s)
		name, value, found := strings.Cut(s, "=")
		if !found {
			return nil, errEqualNotFoundInCookie
		}
		// TODO: Validate
		name = sanitizeCookieName(name)
		value, quoted := sanitizeCookieValue(value)
		cookies = append(cookies, &Cookie{Name: name, Value: value, Quoted: quoted})
	}
	return cookies, nil
}

// ParseCookie parses a single cookie header string and returns the first
// mimheader is with multiple values is seperated by comma
// cookies use semicolons
//
// Example of raw request:
// ````
// GET / HTTP/1.1
// Host: localhost:8080
// User-Agent: curl/7.64.1
// Accept: */*
// Cookie: username=JohnDoe; session_token=abc123
// ````
func ParseCookieToMap(cookieHeader string) map[string]string {
	reqCookies := make(map[string]string)
	cookies := strings.Split(cookieHeader, ";")
	for _, cookie := range cookies {
		parts := strings.SplitN(strings.TrimSpace(cookie), "=", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			reqCookies[name] = val
		}
	}
	return reqCookies
}

// ParseSetCookie parses a single set-cookie header string and returns the
//
// Example of raw response:
// ````
// HTTP/1.1 200 OK
// Content-Type: text/html; charset=UTF-8
// Set-Cookie: session_token=abc123; Path=/; HttpOnly
// Set-Cookie: theme=light; Expires=Wed, 09 Jun 2021 10:18:14 GMT
// ````
func ParseSetCookie(setCookieHeader string) ([]*Cookie, error) {
	parts := strings.Split(TrimString(setCookieHeader), ";")
	if len(parts) == 1 && parts[0] == "" {
		return nil, errBlankCookie
	}
	cookies := make([]*Cookie, 0, len(parts))
	for _, s := range parts {
		s = TrimString(s)
		name, value, found := strings.Cut(s, "=")
		// Validate
		if !found {
			return nil, errEqualNotFoundInCookie
		}
		// Validate and sanitize cookie name
		if !isValidCookieName(name) {
			return nil, errInvalidCookieName
		}
		name = sanitizeCookieName(name)
		// Validate and sanitize cookie value
		if !isValidCookieValue(value) {
			return nil, errInvalidCookieValue
		}
		value, quoted := sanitizeCookieValue(value)
		// Set
		cookies = append(cookies, &Cookie{Name: name, Value: value, Quoted: quoted})
	}
	return cookies, nil
}

func sanitizeCookieName(n string) string {
	return TrimString(n)
}

func sanitizeCookieValue(v string) (string, bool) {
	v = strings.Trim(v, " ")
	// check if quotes in string
	isQuotes := strings.Contains(v, "\"")
	// escape parts of string
	v = url.QueryEscape(v)
	return v, isQuotes
}

// isValidCookieValue is a basic validation to check if the provided string is a valid cookie value.
func isValidCookieValue(value string) bool {
	// If the value is enclosed in double quotes, strip them for validation.
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		value = strings.Trim(value, "\"")
	} else {
		// If not enclosed in double quotes, check for forbidden characters.
		if strings.ContainsAny(value, " ,;\\") {
			return false
		}
	}

	// Further checks can be added here, such as length restrictions.
	return true
}

// isValidCookieName checks if the provided string is a valid cookie name.
func isValidCookieName(name string) bool {
	// Cookie names cannot be empty
	if name == "" {
		return false
	}
	// Check for forbidden characters in cookie names
	forbiddenChars := " ,;=\r\n\t"
	if strings.ContainsAny(name, forbiddenChars) {
		return false
	}
	// Check for control characters (0-31 and 127 in ASCII)
	for _, char := range name {
		if char <= 31 || char == 127 {
			return false
		}
	}
	return true
}

func getCookieAttributes(parts []string) map[string]string {
	attributes := make(map[string]string)
	for _, part := range parts {
		if strings.Contains(part, "=") {
			attribute := strings.SplitN(part, "=", 2)
			attributes[attribute[0]] = attribute[1]
		}
	}
	return attributes
}

func readCookiesFromHeader(header string) []*Cookie {
	cookies := make([]*Cookie, 0)
	parts := strings.Split(header, ";")
	for _, part := range parts {
		cookie := &Cookie{}
		cookieParts := strings.Split(part, "=")
		if len(cookieParts) == 2 {
			cookie.Name = cookieParts[0]
			cookie.Value = cookieParts[1]
			cookies = append(cookies, cookie)
		}
	}
	return cookies
}

func TrimString(s string) string {
	return strings.TrimSpace(s)
}

package message

import (
	"fmt"
	"net/url"
	"time"
)

type Cookie struct {
	Name  string
	Value string

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
	// SameSite SameSite
	Raw      string
	Unparsed []string // Raw text of unparsed attribute-value pairs
}

// String returns the serialization of the cookie for use in a Cookie header (if only Name and Value are set) or a Set-Cookie response header (if other fields are set).
//
//	If c is nil or c.Name is invalid, the empty string is returned.
func (c *Cookie) String() string {
	return c.Name + "=" + c.Value
}

func (c *Cookie) Valid() error {
	if c.Name == "" {
		return fmt.Errorf("Cookie name is empty")
	}
	return nil
}

// A CookieJar manages storage and use of cookies in HTTP requests.
// Implementations of CookieJar must be safe for concurrent use by multiple goroutines.
// The net/http/cookiejar package provides a CookieJar implementation.
type CookieJar interface {
	// SetCookies handles the receipt of the cookies in a reply for the
	// given URL.  It may or may not choose to save the cookies, depending
	// on the jar's policy and implementation.
	SetCookies(u *url.URL, cookies []*Cookie)

	// Cookies returns the cookies to send in a request for the given URL.
	// It is up to the implementation to honor the standard cookie use
	// restrictions such as in RFC 6265.
	Cookies(u *url.URL) []*Cookie
}

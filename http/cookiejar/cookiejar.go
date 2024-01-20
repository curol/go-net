package cookiejar

import (
	"net/http"
	"net/url"
	"strings"
)

// A CookieJar manages storage and use of cookies in HTTP requests.
// Implementations of CookieJar must be safe for concurrent use by multiple goroutines.
// The net/http/cookiejar package provides a CookieJar implementation.
type CookieJarInterface interface {
	// SetCookies handles the receipt of the cookies in a reply for the
	// given URL.  It may or may not choose to save the cookies, depending
	// on the jar's policy and implementation.
	SetCookies(u *url.URL, cookies []*http.Cookie)

	// Cookies returns the cookies to send in a request for the given URL.
	// It is up to the implementation to honor the standard cookie use
	// restrictions such as in RFC 6265.
	Cookies(u *url.URL) []*http.Cookie
}

type CookieJar struct {
	cookies []*http.Cookie
}

func NewCookieJar() *CookieJar {
	return &CookieJar{
		cookies: make([]*http.Cookie, 0),
	}
}

func (c *CookieJar) Len() int {
	return len(c.cookies)
}

// Set adds a cookie to the CookieJar.
// If the cookie exists, it will be updated.
// If the cookie does not exist, it will be added.
//
// Note:
// - If the cookie's domain is blank, the domain of the request is used.
// - If the cookie expires or MaxAge<0, it will be deleted from the CookieJar.
func (c *CookieJar) Set(cookie *http.Cookie) {
	// if cookie exists, update cookie
	for i, ck := range c.cookies {
		if ck.Name == cookie.Name {
			c.cookies[i] = cookie
			return
		}
	}
	// if cookie does not exist, add new cookie
	c.cookies = append(c.cookies, cookie)
}

// SetCookies handles the receipt of the cookies in a reply for the
// given URL.  It may or may not choose to save the cookies, depending
// on the jar's policy and implementation.
func (jar *CookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.cookies = cookies
}

// Cookies returns the cookies to send in a request for the given URL.
// It is up to the implementation to honor the standard cookie use
// restrictions such as in RFC 6265.
func (jar *CookieJar) Cookies(u *url.URL) []*http.Cookie {
	cookies := make([]*http.Cookie, 0)
	for _, c := range jar.cookies {
		if c.Domain == u.Hostname() && strings.HasPrefix(u.Path, c.Path) {
			cookie := &http.Cookie{
				Name:     c.Name,
				Value:    c.Value,
				Domain:   c.Domain,
				Path:     c.Path,
				Secure:   c.Secure,
				HttpOnly: c.HttpOnly,
			}
			if !c.Expires.IsZero() {
				cookie.Expires = c.Expires
			} else if c.MaxAge != 0 {
				cookie.MaxAge = c.MaxAge
			}
			cookies = append(cookies, cookie)
		}
	}
	return cookies
}

func (c *CookieJar) Get(name string) *http.Cookie {
	for _, cookie := range c.cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func (c *CookieJar) Delete(name string) {
	for i, cookie := range c.cookies {
		if cookie.Name == name {
			c.cookies = append(c.cookies[:i], c.cookies[i+1:]...)
		}
	}
}

func (c *CookieJar) Clear() {
	c.cookies = make([]*http.Cookie, 0)
}

func (c *CookieJar) String() string {
	var sb strings.Builder
	for _, cookie := range c.cookies {
		sb.WriteString(cookie.String())
	}
	return sb.String()
}

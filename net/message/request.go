package message

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/curol/network/net/cookie"
	"github.com/curol/network/net/textproto"
)

// Req is a wrapper for textproto.Reader which implements convenience methods for reading requests.
type Request struct {
	*textproto.TextMessage
	Form    map[string][]string
	Cookies []*cookie.Cookie
}

func NewRequest(r *bufio.Reader) (*Request, error) {
	tm, err := textproto.ReadTextMessage(r)
	if err != nil {
		return nil, err
	}
	req := &Request{}
	req.TextMessage = tm
	req.Form = make(map[string][]string)
	req.Cookies = make([]*cookie.Cookie, 0)

	// Set the cookies
	v, ok := req.Headers["Cookie"]
	if ok {
		req.parseCookies(v[0])
	}

	return req, nil
}

func (req *Request) IsForm() bool {
	contentType, ok := req.Headers["Content-Type"]
	if !ok {
		return false
	}
	return IsForm(contentType[0])
}

func (req *Request) ParseForm() error {
	// Parse the form
	if req.IsForm() {
		buf, err := req.readBody() // read body and return bytes
		if err != nil {
			return err
		}
		form, err := ParseForm(string(buf)) // parse form
		if err != nil {
			return err
		}
		req.Form = form
		// }
	}
	return nil
}

func (req *Request) readBody() ([]byte, error) {
	contLen := req.ContentLen
	r := io.LimitReader(req.Body, contLen)
	buf := make([]byte, contLen)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		if err == io.EOF {
			return buf, nil
		}
		return nil, err
	}
	return buf, nil
}

func (req *Request) Cookie(name string) (string, error) {
	for _, cookie := range req.Cookies {
		if cookie.Name == name {
			return cookie.Value, nil
		}
	}
	return "", fmt.Errorf("Cookie not found")
}

func (req *Request) serializeCookies() string {
	cookies := req.Cookies
	cookieHeader := ""
	for _, cookie := range cookies {
		cookieHeader += cookie.Name + "=" + cookie.Value + "; "
	}
	return cookieHeader
}

func (req *Request) parseCookies(cookieHeader string) {
	cookies, err := cookie.ParseCookie(cookieHeader)
	if err != nil {
		return
	}
	req.Cookies = cookies
}

func ParseForm(body string) (map[string][]string, error) {
	form, err := url.ParseQuery(body)
	if err != nil {
		return nil, err
	}
	return form, nil
}

func IsForm(contentType string) bool {
	// Check if the Content-Type header indicates form data
	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") || strings.HasPrefix(contentType, "multipart/form-data") {
		return true
	}
	return false
}

func IsCookie(cookieHeader string) bool {
	// Check if the Content-Type header indicates form data
	if strings.HasPrefix(cookieHeader, "Cookie") {
		return true
	}
	return false
}

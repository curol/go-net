package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	net "net/http"
	"net/http/httputil"
	"reflect"
	"strings"
	"testing"

	"github.com/curol/network/url"
)

func TestNewRequest(t *testing.T) {
	u, err := url.Parse("http://www.example.com")
	h := newHeaderFromMap(map[string][]string{
		"Host":           {"www.example.com"},
		"Content-Length": {"12"},
		"Content-Type":   {"text/plain"},
	})
	// Arrange
	want := &Request{
		// request-line
		Method:     "GET",
		Protocol:   protocol,
		RequestURI: u.RequestURI(),
		URL:        u,
		Host:       u.Host,
		// headers
		Header:        h,
		ContentType:   getContentType(h),
		ContentLength: getContentLength(h),
		Cookies:       getCookies(h),
		// body
		Body:          nil,
		Form:          nil,
		MultipartForm: nil,
		// conn
		RemoteAddress: "",
	}

	// Act
	got, err := NewRequest("GET", "http://www.example.com", h, nil)
	if err != nil {
		t.Errorf("NewRequest() returned error: %v", err)
	}

	// Assert
	assertRequest(got, want, t)

}

func TestReadRequest(t *testing.T) {
	// Arrange
	lines := []string{
		// request-line
		"GET http://www.example.com HTTP/1.1",
		// headers
		"Host: www.example.com",
		"Content-Length: 12",
		"Content-Type: text/plain",
		"",
		// body
		"",
	}
	raw := []byte(strings.Join(lines, "\r\n"))
	body := io.NopCloser(strings.NewReader(lines[len(lines)-1]))
	h := newHeaderFromMap(map[string][]string{
		"Host":           {"www.example.com"},
		"Content-Length": {"12"},
		"Content-Type":   {"text/plain"},
	})
	want := &Request{
		// Request-Line
		Method:     "GET",
		Protocol:   protocol,
		RequestURI: "/",
		URL:        &url.URL{Path: "/"},
		Host:       "www.example.com",
		// Headers
		Header:        h,
		ContentType:   getContentType(h),
		ContentLength: getContentLength(h),
		Cookies:       getCookies(h),
		// Body
		Body:          body,
		Form:          nil,
		MultipartForm: nil,
		// Connection
		RemoteAddress: "",
	}

	// Act
	got, err := ReadRequest(bufio.NewReader(strings.NewReader(string(raw))))
	if err != nil {
		t.Errorf("ReadRequest() returned error: %v", err)
		return
	}

	// Assert
	req2 := netReadRequest(raw, t)
	if req2 == nil {
		t.Errorf("net.ReadRequest() returned nil")
		return
	}
	netWriteRequest(req2, t)

	assertRequest(got, want, t)
}

func netReadRequest(raw []byte, t *testing.T) *net.Request {
	// Read
	hreq, err := net.ReadRequest(bufio.NewReader(strings.NewReader(string(raw)))) // net/http.Request
	if err != nil {
		fmt.Println(hreq)
		t.Errorf("net.ReadRequest() returned error: %v", err)
	}
	return hreq
}

func netWriteRequest(req *net.Request, t *testing.T) *bytes.Buffer {
	// Write
	buf := bytes.NewBuffer(nil)
	req.Write(buf)
	fmt.Println("net/http.Request")
	fmt.Println(buf.String())
	return buf
}

func debugNetRequest(req *net.Request, t *testing.T) {
	// Debug
	buf, err := httputil.DumpRequest(req, true)
	if err != nil {
		t.Errorf("httputil.DumpRequest() returned error: %v", err)
	}
	fmt.Println("net/http.Request")
	fmt.Println(string(buf))
}

func TestReadRequestWithBody(t *testing.T) {
	// Arrange
	lines := []string{
		// request-line
		"POST http://www.example.com HTTP/1.1",
		// headers
		"Host: www.example.com",
		"Content-Length: 12",
		"Content-Type: text/plain",
		"",
		// body
		"Hello World!",
	}
	raw := []byte(strings.Join(lines, "\r\n"))
	body := io.NopCloser(strings.NewReader(lines[len(lines)-1]))
	h := newHeaderFromMap(map[string][]string{
		"Host":           {"www.example.com"},
		"Content-Length": {"12"},
		"Content-Type":   {"text/plain"},
	})
	path := "http://www.example.com"
	u, _ := url.Parse(path)
	want := &Request{
		// Request-Line
		Method:     "POST",
		Protocol:   "HTTP/1.1",
		RequestURI: path,
		URL:        u,
		Host:       u.Host,
		// Headers
		Header:        h,
		ContentType:   getContentType(h),
		ContentLength: getContentLength(h),
		Cookies:       getCookies(h),
		// Body
		Body:          body,
		Form:          nil,
		MultipartForm: nil,
		// Connection
		RemoteAddress: "",
	}

	// Act
	got, err := ReadRequest(bufio.NewReader(strings.NewReader(string(raw))))
	if err != nil {
		t.Errorf("ReadRequest() returned error: %v", err)
		return
	}

	// Assert
	assertRequest(got, want, t)
}

func TestRequestBuffer(t *testing.T) {
	// Act
	// req := NewRequest("GET", "http://www.example.com", nil, nil) // client request
	// buf, _ := req.ToBuffer()
	// reader := bufio.NewReader(buf)
}

func assertRequest(got *Request, want *Request, t *testing.T) {
	// Assert
	if !reflect.DeepEqual(got.Method, want.Method) {
		t.Errorf("ReadRequest() = got.method %v, want.method %v", got.Method, want.Method)
	}
	if !reflect.DeepEqual(got.Protocol, want.Protocol) {
		t.Errorf("ReadRequest() = got.protocol %v, want.protocol %v", got.Protocol, want.Protocol)
	}
	if !reflect.DeepEqual(got.Header, want.Header) {
		t.Errorf("ReadRequest() = \ngot.header: %v\n want.header: %v\n", got.Header, want.Header)
	}
	if !reflect.DeepEqual(got.URL, want.URL) {
		t.Errorf("ReadRequest() = got.url %v, want.url %v", got.URL, want.URL)
	}
	if !reflect.DeepEqual(got.RequestURI, want.RequestURI) {
		t.Errorf("ReadRequest() = got.requestURI %v, want.requestURI %v", got.RequestURI, want.RequestURI)
	}
	if !reflect.DeepEqual(got.ContentLength, want.ContentLength) {
		t.Errorf("ReadRequest() = got.contentLength %v, want.contentLength %v", got.ContentLength, want.ContentLength)
	}
	if !reflect.DeepEqual(got.ContentType, want.ContentType) {
		t.Errorf("ReadRequest() = got.contentType %v, want.contentType %v", got.ContentType, want.ContentType)
	}
	if !reflect.DeepEqual(got.Cookies, want.Cookies) {
		t.Errorf("ReadRequest() = got.cookies %v, want.cookies %v", got.Cookies, want.Cookies)
	}
	if !reflect.DeepEqual(got.Host, want.Host) {
		t.Errorf("ReadRequest() = got.host: %v, want.host %v", got.Host, want.Host)
	}
	if !reflect.DeepEqual(got.RemoteAddress, want.RemoteAddress) {
		t.Errorf("ReadRequest() = got.remoteAddress %v\n, want.remoteAddress\n %v", got.RemoteAddress, want.RemoteAddress)
	}
	if !reflect.DeepEqual(got.Form, want.Form) {
		t.Errorf("ReadRequest() = got.form: %v\n, want.form %v\n", got.Form, want.Form)
	}
	if !reflect.DeepEqual(got.MultipartForm, want.MultipartForm) {
		t.Errorf("ReadRequest() = got.multipartForm\n %v, want.multipartForm\n %v", got.MultipartForm, want.MultipartForm)
	}

	// Assert Body
	if got.Body == nil && want.Body == nil {
		return
	}
	if got.Body != nil && want.Body == nil {
		t.Errorf("ReadRequest() = got.body != nil & want.body == nil")
		return
	}
	if got.Body == nil && want.Body != nil {
		t.Errorf("ReadRequest() = got.body == nil & want.body != nil")
		return
	}
	if got.Body != nil && want.Body != nil {
		gbody, _ := io.ReadAll(got.Body)
		wbody, _ := io.ReadAll(want.Body)
		if !reflect.DeepEqual(gbody, wbody) {
			t.Errorf("ReadRequest() = got.body\n %v, want.body\n %v", got.Body, want.Body)
			return
		}
	}
}

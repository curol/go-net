package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/curol/network/url"
)

type reqWriteTest struct {
	Req  Request
	Body any // optional []byte or func() io.ReadCloser to populate Req.Body

	// Any of these three may be empty to skip that test.
	WantWrite string // Request.Write
	WantProxy string // Request.WriteProxy

	WantError error // wanted error from Request.Write
}

var reqWriteTests = []reqWriteTest{
	// HTTP/1.1 => chunked coding; no body; no trailer
	0: {
		Req: Request{
			Method: "GET",
			URL: &url.URL{
				Scheme: "http",
				Host:   "www.techcrunch.com",
				Path:   "/",
			},
			Protocol: "HTTP/1.1",
			Header: Header{
				"Accept":           {"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
				"Accept-Charset":   {"ISO-8859-1,utf-8;q=0.7,*;q=0.7"},
				"Accept-Encoding":  {"gzip,deflate"},
				"Accept-Language":  {"en-us,en;q=0.5"},
				"Keep-Alive":       {"300"},
				"Proxy-Connection": {"keep-alive"},
				"User-Agent":       {"Fake"},
			},
			Body:  nil,
			Close: false,
			Host:  "www.techcrunch.com",
			Form:  map[string][]string{},
		},

		WantWrite: "GET / HTTP/1.1\r\n" +
			"Host: www.techcrunch.com\r\n" +
			"User-Agent: Fake\r\n" +
			"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\n" +
			"Accept-Charset: ISO-8859-1,utf-8;q=0.7,*;q=0.7\r\n" +
			"Accept-Encoding: gzip,deflate\r\n" +
			"Accept-Language: en-us,en;q=0.5\r\n" +
			"Keep-Alive: 300\r\n" +
			"Proxy-Connection: keep-alive\r\n\r\n",

		WantProxy: "GET http://www.techcrunch.com/ HTTP/1.1\r\n" +
			"Host: www.techcrunch.com\r\n" +
			"User-Agent: Fake\r\n" +
			"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\n" +
			"Accept-Charset: ISO-8859-1,utf-8;q=0.7,*;q=0.7\r\n" +
			"Accept-Encoding: gzip,deflate\r\n" +
			"Accept-Language: en-us,en;q=0.5\r\n" +
			"Keep-Alive: 300\r\n" +
			"Proxy-Connection: keep-alive\r\n\r\n",
	},
	// HTTP/1.1 => chunked coding; body; empty trailer
	1: {
		Req: Request{
			Method: "GET",
			URL: &url.URL{
				Scheme: "http",
				Host:   "www.google.com",
				Path:   "/search",
			},
			Protocol:         "HTTP/1.1",
			Header:           Header{},
			TransferEncoding: []string{"chunked"}, // TODO: fix this
		},

		Body: []byte("abcdef"),

		WantWrite: "GET /search HTTP/1.1\r\n" +
			"Host: www.google.com\r\n" +
			"User-Agent: curol-http-client/1.1\r\n" +
			"Transfer-Encoding: chunked\r\n\r\n" +
			chunk("abcdef") + chunk(""),

		WantProxy: "GET http://www.google.com/search HTTP/1.1\r\n" +
			"Host: www.google.com\r\n" +
			"User-Agent: curol-http-client/1.1\r\n" +
			"Transfer-Encoding: chunked\r\n\r\n" +
			chunk("abcdef") + chunk(""),
	},
	// HTTP/1.1 POST => chunked coding; body; empty trailer
	2: {
		Req: Request{
			Method: "POST",
			URL: &url.URL{
				Scheme: "http",
				Host:   "www.google.com",
				Path:   "/search",
			},
			Protocol:         "HTTP/1.1",
			Header:           Header{},
			Close:            true,
			TransferEncoding: []string{"chunked"}, // TODO: fix this
		},

		Body: []byte("abcdef"),

		WantWrite: "POST /search HTTP/1.1\r\n" +
			"Host: www.google.com\r\n" +
			"User-Agent: curol-http-client/1.1\r\n" +
			"Connection: close\r\n" +
			"Transfer-Encoding: chunked\r\n\r\n" +
			chunk("abcdef") + chunk(""),

		WantProxy: "POST http://www.google.com/search HTTP/1.1\r\n" +
			"Host: www.google.com\r\n" +
			"User-Agent: curol-http-client/1.1\r\n" +
			"Connection: close\r\n" +
			"Transfer-Encoding: chunked\r\n\r\n" +
			chunk("abcdef") + chunk(""),
	},

	// HTTP/1.1 POST with Content-Length, no chunking
	3: {
		Req: Request{
			Method: "POST",
			URL: &url.URL{
				Scheme: "http",
				Host:   "www.google.com",
				Path:   "/search",
			},
			Protocol:      "HTTP/1.1",
			Header:        Header{},
			Close:         true,
			ContentLength: 6,
		},

		Body: []byte("abcdef"),

		WantWrite: "POST /search HTTP/1.1\r\n" +
			"Host: www.google.com\r\n" +
			"User-Agent: curol-http-client/1.1\r\n" +
			"Connection: close\r\n" +
			"Content-Length: 6\r\n" +
			"\r\n" +
			"abcdef",

		WantProxy: "POST http://www.google.com/search HTTP/1.1\r\n" +
			"Host: www.google.com\r\n" +
			"User-Agent: curol-http-client/1.1\r\n" +
			"Connection: close\r\n" +
			"Content-Length: 6\r\n" +
			"\r\n" +
			"abcdef",
	},

	// HTTP/1.1 POST with Content-Length in headers
	4: {
		Req: Request{
			Method: "POST",
			URL:    mustParseURL("http://example.com/"),
			Host:   "example.com",
			Header: Header{
				"Content-Length": []string{"10"}, // ignored
			},
			ContentLength: 6,
		},

		Body: []byte("abcdef"),

		WantWrite: "POST / HTTP/1.1\r\n" +
			"Host: example.com\r\n" +
			"User-Agent: curol-http-client/1.1\r\n" +
			"Content-Length: 6\r\n" +
			"\r\n" +
			"abcdef",

		WantProxy: "POST http://example.com/ HTTP/1.1\r\n" +
			"Host: example.com\r\n" +
			"User-Agent: curol-http-client/1.1\r\n" +
			"Content-Length: 6\r\n" +
			"\r\n" +
			"abcdef",
	},

	// default to HTTP/1.1
	5: {
		Req: Request{
			Method: "GET",
			URL:    mustParseURL("/search"),
			Host:   "www.google.com",
		},

		WantWrite: "GET /search HTTP/1.1\r\n" +
			"Host: www.google.com\r\n" +
			"User-Agent: curol-http-client/1.1\r\n" +
			"\r\n",
	},

	// Request with a 0 ContentLength and a 0 byte body.
	6: {
		Req: Request{
			Method:        "POST",
			URL:           mustParseURL("/"),
			Host:          "example.com",
			Protocol:      "HTTP/1.1",
			ContentLength: 0, // as if unset by user
		},

		Body: func() io.ReadCloser { return io.NopCloser(io.LimitReader(strings.NewReader("xx"), 0)) },

		WantWrite: "POST / HTTP/1.1\r\n" +
			"Host: example.com\r\n" +
			"User-Agent: curol-http-client/1.1\r\n" +
			"Transfer-Encoding: chunked\r\n" +
			"\r\n0\r\n\r\n",

		WantProxy: "POST / HTTP/1.1\r\n" +
			"Host: example.com\r\n" +
			"User-Agent: curol-http-client/1.1\r\n" +
			"Transfer-Encoding: chunked\r\n" +
			"\r\n0\r\n\r\n",
	},

	// Request with a 0 ContentLength and a nil body.
	7: {
		Req: Request{
			Method:        "POST",
			URL:           mustParseURL("/"),
			Host:          "example.com",
			Protocol:      "HTTP/1.1",
			ContentLength: 0, // as if unset by user
		},

		Body: func() io.ReadCloser { return nil },

		WantWrite: "POST / HTTP/1.1\r\n" +
			"Host: example.com\r\n" +
			"User-Agent: curol-http-client/1.1\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n",

		WantProxy: "POST / HTTP/1.1\r\n" +
			"Host: example.com\r\n" +
			"User-Agent: curol-http-client/1.1\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n",
	},

	// Request with a 0 ContentLength and a 1 byte body.
	8: {
		Req: Request{
			Method:        "POST",
			URL:           mustParseURL("/"),
			Host:          "example.com",
			Protocol:      "HTTP/1.1",
			ContentLength: 0, // as if unset by user
		},

		Body: func() io.ReadCloser { return io.NopCloser(io.LimitReader(strings.NewReader("xx"), 1)) },

		WantWrite: "POST / HTTP/1.1\r\n" +
			"Host: example.com\r\n" +
			"User-Agent: curol-http-client/1.1\r\n" +
			"Transfer-Encoding: chunked\r\n\r\n" +
			chunk("x") + chunk(""),

		WantProxy: "POST / HTTP/1.1\r\n" +
			"Host: example.com\r\n" +
			"User-Agent: curol-http-client/1.1\r\n" +
			"Transfer-Encoding: chunked\r\n\r\n" +
			chunk("x") + chunk(""),
	},

	// Request with a ContentLength of 10 but a 5 byte body.
	9: {
		Req: Request{
			Method:        "POST",
			URL:           mustParseURL("/"),
			Host:          "example.com",
			Protocol:      "HTTP/1.1",
			ContentLength: 10, // but we're going to send only 5 bytes
		},
		Body:      []byte("12345"),
		WantError: errors.New("http: ContentLength=10 with Body length 5"),
	},

	// Request with a ContentLength of 4 but an 8 byte body.
	10: {
		Req: Request{
			Method:        "POST",
			URL:           mustParseURL("/"),
			Host:          "example.com",
			Protocol:      "HTTP/1.1",
			ContentLength: 4, // but we're going to try to send 8 bytes
		},
		Body:      []byte("12345678"),
		WantError: errors.New("http: ContentLength=4 with Body length 8"),
	},

	// Request with a 5 ContentLength and nil body.
	11: {
		Req: Request{
			Method:        "POST",
			URL:           mustParseURL("/"),
			Host:          "example.com",
			Protocol:      "HTTP/1.1",
			ContentLength: 5, // but we'll omit the body
		},
		WantError: errors.New("http: Request.ContentLength=5 with nil Body"),
	},

	// Request with a 0 ContentLength and a body with 1 byte content and an error.
	12: {
		Req: Request{
			Method:        "POST",
			URL:           mustParseURL("/"),
			Host:          "example.com",
			Protocol:      "HTTP/1.1",
			ContentLength: 0, // as if unset by user
		},

		Body: func() io.ReadCloser {
			err := errors.New("Custom reader error")
			errReader := iotest.ErrReader(err)
			return io.NopCloser(io.MultiReader(strings.NewReader("x"), errReader))
		},

		WantError: errors.New("Custom reader error"),
	},

	// Request with a 0 ContentLength and a body without content and an error.
	13: {
		Req: Request{
			Method:        "POST",
			URL:           mustParseURL("/"),
			Host:          "example.com",
			Protocol:      "HTTP/1.1",
			ContentLength: 0, // as if unset by user
		},

		Body: func() io.ReadCloser {
			err := errors.New("Custom reader error")
			errReader := iotest.ErrReader(err)
			return io.NopCloser(errReader)
		},

		WantError: errors.New("Custom reader error"),
	},
}

func TestRequestWrite(t *testing.T) {
	for i := range reqWriteTests {
		tt := &reqWriteTests[i]

		setBody := func() {
			if tt.Body == nil {
				return
			}
			switch b := tt.Body.(type) {
			case []byte:
				tt.Req.Body = io.NopCloser(bytes.NewReader(b))
			case func() io.ReadCloser:
				tt.Req.Body = b()
			}
		}
		setBody()
		if tt.Req.Header == nil {
			tt.Req.Header = make(Header)
		}

		var braw strings.Builder
		err := tt.Req.Write(&braw)
		if g, e := fmt.Sprintf("%v", err), fmt.Sprintf("%v", tt.WantError); g != e {
			t.Errorf("writing #%d, err = %q, want %q", i, g, e)
			continue
		}
		if err != nil {
			continue
		}

		if tt.WantWrite != "" {
			sraw := braw.String()
			if sraw != tt.WantWrite {
				t.Errorf("Test %d, expecting:\n%s\nGot:\n%s\n", i, tt.WantWrite, sraw)
				continue
			}
		}

		// if tt.WantProxy != "" {
		// 	setBody()
		// 	var praw strings.Builder
		// 	err = tt.Req.WriteProxy(&praw)
		// 	if err != nil {
		// 		t.Errorf("WriteProxy #%d: %s", i, err)
		// 		continue
		// 	}
		// 	sraw := praw.String()
		// 	if sraw != tt.WantProxy {
		// 		t.Errorf("Test Proxy %d, expecting:\n%s\nGot:\n%s\n", i, tt.WantProxy, sraw)
		// 		continue
		// 	}
		// }
	}
}

func chunk(s string) string {
	return fmt.Sprintf("%x\r\n%s\r\n", len(s), s)
}

func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(fmt.Sprintf("Error parsing URL %q: %v", s, err))
	}
	return u
}

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

// delegateReader is a reader that delegates to another reader,
// once it arrives on a channel.
type delegateReader struct {
	c chan io.Reader
	r io.Reader // nil until received from c
}

func (r *delegateReader) Read(p []byte) (int, error) {
	if r.r == nil {
		r.r = <-r.c
	}
	return r.r.Read(p)
}

// dumpConn is a net.Conn that writes to Writer and reads from Reader.
type dumpConn struct {
	io.Writer
	io.Reader
}

func (c *dumpConn) Close() error                       { return nil }
func (c *dumpConn) LocalAddr() net.Addr                { return nil }
func (c *dumpConn) RemoteAddr() net.Addr               { return nil }
func (c *dumpConn) SetDeadline(t time.Time) error      { return nil }
func (c *dumpConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *dumpConn) SetWriteDeadline(t time.Time) error { return nil }

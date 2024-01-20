// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	net "net/http"
	"reflect"
	"strings"
	"testing"

	token "github.com/curol/network/http/internal/token"
	url "github.com/curol/network/url"
)

type reqTest struct {
	Raw   string
	Want  *Request
	Error string
	Body  any
}

var noError = ""
var noBodyStr = ""

var reqTests = []reqTest{
	0: {
		Raw: "GET / HTTP/1.1\r\nHost: foo.com\r\n\r\n",
		Want: &Request{
			RequestURI: "/",
			Method:     "GET",
			URL: &url.URL{
				Path: "/",
			},
			Protocol: "HTTP/1.1",
			Header:   NewHeader(),
			Host:     "foo.com",
			Cookies:  []*Cookie{},
			// Body:     io.NopCloser(strings.NewReader(noBodyStr)),
			Body: nil,
		},
		Body: noBodyStr,
	},
	1: {
		Raw: "POST / HTTP/1.1\r\nHost: foo.com\r\nContent-Length: 3\r\n\r\nabc",
		Want: &Request{
			RequestURI: "/",
			Method:     "POST",
			URL: &url.URL{
				Path: "/",
			},
			Protocol:      "HTTP/1.1",
			Header:        Header(map[string][]string{"Content-Length": {"3"}}),
			Host:          "foo.com",
			ContentLength: 3,
			Cookies:       []*Cookie{},
			// Body:          io.NopCloser(strings.NewReader("abc")),
			// Body: io.NopCloser(strings.NewReader("abc")),
			Body: nil,
		},
		Body: "abc",
	},
	2: {
		Raw: "POST / HTTP/1.1\r\nHost: foo.com\r\nContent-Length: 4\r\nTransfer-Encoding: chunked\r\n\r\nabcd",
		Want: &Request{
			RequestURI: "/",
			Method:     "POST",
			URL: &url.URL{
				Path: "/",
			},
			Protocol:      "HTTP/1.1",
			Header:        Header(map[string][]string{"Content-Length": {"4"}, "Transfer-Encoding": {"chunked"}}),
			Host:          "foo.com",
			ContentLength: 4,
			TransferEncoding: []string{
				"chunked",
			},
			Cookies: []*Cookie{},
			Body:    nil,
		},
		Body: "abcd",
	},
}

func TestReadRequests(t *testing.T) {
	for i := range reqTests {
		tt := &reqTests[i]

		// Act
		req, err := ReadRequest(bufio.NewReader(strings.NewReader(tt.Raw))) // read
		if err != nil {
			if err.Error() != tt.Error {
				t.Errorf("#%d: error %q, want error %q", i, err.Error(), tt.Error)
			}
			continue
		}
		req2, err := net.ReadRequest(bufio.NewReader(strings.NewReader(tt.Raw))) // read
		if err != nil {
			if err.Error() != tt.Error {
				t.Errorf("#%d: error %q, want error %q", i, err.Error(), tt.Error)
			}
			continue
		}
		rbody := req.Body // set body
		req.Body = nil
		testName := fmt.Sprintf("\nTest %d (%q)", i, tt.Raw)
		diff(t, testName, req, tt.Want)
		var bout strings.Builder // copy body
		if rbody != nil {
			_, err := io.Copy(&bout, rbody)
			if err != nil {
				t.Fatalf("%s: copying body: %v", testName, err)
			}
			rbody.Close()
		}
		body := bout.String()

		// Assert body
		if body != tt.Body {
			t.Errorf("%s: Body = %q\n want %q\n", testName, body, tt.Body)
		}

		compareReqToHttpRequest(req, req2, *tt, t)
	}
}

// Compare github.com/curol/network.Request to http.Request
func compareReqToHttpRequest(req *Request, req2 *net.Request, tt reqTest, t *testing.T) {
	buf := bytes.NewBuffer(nil) // network.Request
	req.Write(buf)              // write request
	b := []byte(tt.Body.(string))
	buf.Write(b)          // write body since its nil for test
	myreq := buf.String() // get request as string

	buf = bytes.NewBuffer(nil) // http.Request
	req2.Write(buf)
	libreq := buf.String()
	if myreq != libreq {
		t.Errorf("github.com/curol/network/http.Request != http.Request\ngithub.com/curol/network/http.Request:\n%s\n\nhttp.Request:\n%s\n", myreq, libreq)
	}
}

// reqBytes treats req as a request (with \n delimiters) and returns it with \r\n delimiters,
// ending in \r\n\r\n
func reqBytes(req string) []byte {
	return []byte(strings.ReplaceAll(strings.TrimSpace(req), "\n", "\r\n") + "\r\n\r\n")
}

func diff(t *testing.T, prefix string, have, want any) {
	t.Helper()
	hv := reflect.ValueOf(have).Elem()
	wv := reflect.ValueOf(want).Elem()
	if hv.Type() != wv.Type() {
		t.Errorf("%s: type mismatch %v want %v", prefix, hv.Type(), wv.Type())
	}
	for i := 0; i < hv.NumField(); i++ {
		name := hv.Type().Field(i).Name
		if !token.IsExported(name) {
			continue
		}
		hf := hv.Field(i).Interface()
		wf := wv.Field(i).Interface()
		if !reflect.DeepEqual(hf, wf) {
			t.Errorf("%s: %s = %v want %v", prefix, name, hf, wf)
		}
	}
}

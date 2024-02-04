package tests

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	libhttp "net/http"

	http "github.com/curol/network/http"

	url "github.com/curol/network/url"

	"github.com/curol/network/http/tests/mock"
)

var defaultUserAgent = "Go-http-client/1.1"

// Compare `github.com/curol/network.Request` to standard library http request `net/http.Request`
func TestCompareRequests(t *testing.T) {
	// Arrange
	want := mock.MockPostJSONRequest(nil)

	// Act
	data := strings.NewReader(`{"key": "value"}`)
	got := mock.MockStandardLibraryHttpPostJsonRequest(data)

	// Assert standard library request is equal to `github.com/curol/network/http.Request`
	compareRequests(want, got, t)
}

// Test server request
func TestReadRequest(t *testing.T) {
	// Arrange
	raw := []byte("GET / HTTP/1.1\r\nHost: golang.org\r\n\r\n") // raw req in wire format
	reader := bufio.NewReader(bytes.NewReader(raw))             // create reader from wire
	want, err := libhttp.ReadRequest(reader)                    // read client request using standard library `net/http`
	if err != nil {
		t.Fatal(err)
	}

	// Act
	buf := bytes.NewBuffer(raw)
	reader = bufio.NewReader(buf)
	got, err := http.ReadRequest(reader) // read client request using package `github.com/curol/network/http`
	if err != nil {
		t.Fatal(err)
	}

	// Assert
	compareRequests(got, want, t) // compare `github.com/curol/network/http.Request` to standard library `net/http.Request`
}

// Test client request
func TestClientRequest(t *testing.T) {
	// Arrange
	lines := []string{
		// Request line
		"POST / HTTP/1.1",
		// Headers (note: sort headers for comparing strings)
		"Accept: application/json",
		"Accept-Encoding: gzip",
		"Content-Length: 16",
		"Content-Type: application/json",
		"Host: example.com",
		"User-Agent: Go-http-client/1.1",
		"",
		// Body
		"{\"key\": \"value\"}",
	}
	raw := strings.Join(lines, "\r\n")
	want := raw
	// Act
	req, err := http.NewRequest( // create client request using package `github.com/curol/network/http`
		"POST",               // method
		"http://example.com", // address
		map[string][]string{ // headers
			"Host":            {"example.com"},
			"Content-Type":    {"application/json"},
			"Content-Length":  {"16"},
			"User-Agent":      {"Mozilla/5.0"},
			"Accept":          {"application/json"},
			"Accept-Encoding": {"gzip"},
		},
		strings.NewReader(`{"key": "value"}`), // body
	)
	if err != nil {
		t.Fatal(err)
	}
	got := req.Dump() // dump client request to wire format
	// Assert
	if got != want { // assert client request is equal to expected wire format
		t.Errorf("got:\n%v, want:\n%v", got, want)
	}
}

func TestDumpRequest(t *testing.T) {
	// Arrange
	raw := []byte("GET / HTTP/1.1\r\nAccept-Encoding: gzip\r\nHost: example.com\r\nUser-Agent: Go-http-client/1.1\r\n\r\n") // raw req in wire format
	want := string(raw)
	// Act
	req, err := http.NewRequest("GET", "http://example.com", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := req.Dump()
	// Assert
	if got != want { // assert dump is equal to expected wire format
		t.Errorf("got:\n%v, want:\n%v", got, want)
	}

}

var newRequestHostTests = []struct {
	in, out string
}{
	{"http://www.example.com/", "www.example.com"},
	{"http://www.example.com:8080/", "www.example.com:8080"},

	{"http://192.168.0.1/", "192.168.0.1"},
	{"http://192.168.0.1:8080/", "192.168.0.1:8080"},
	{"http://192.168.0.1:/", "192.168.0.1"},

	{"http://[fe80::1]/", "[fe80::1]"},
	{"http://[fe80::1]:8080/", "[fe80::1]:8080"},
	{"http://[fe80::1%25en0]/", "[fe80::1%en0]"},
	{"http://[fe80::1%25en0]:8080/", "[fe80::1%en0]:8080"},
	{"http://[fe80::1%25en0]:/", "[fe80::1%en0]"},
}

func TestNewRequestHost(t *testing.T) {
	for i, tt := range newRequestHostTests {
		req, err := http.NewRequest("GET", tt.in, nil, nil)
		if err != nil {
			t.Errorf("#%v: %v", i, err)
			continue
		}
		if req.Host != tt.out {
			t.Errorf("got %q; want %q", req.Host, tt.out)
		}
	}
}

var parseHTTPVersionTests = []struct {
	vers         string
	major, minor int
	ok           bool
}{
	{"HTTP/0.0", 0, 0, true},
	{"HTTP/0.9", 0, 9, true},
	{"HTTP/1.0", 1, 0, true},
	{"HTTP/1.1", 1, 1, true},

	{"HTTP", 0, 0, false},
	{"HTTP/one.one", 0, 0, false},
	{"HTTP/1.1/", 0, 0, false},
	{"HTTP/-1,0", 0, 0, false},
	{"HTTP/0,-1", 0, 0, false},
	{"HTTP/", 0, 0, false},
	{"HTTP/1,1", 0, 0, false},
	{"HTTP/+1.1", 0, 0, false},
	{"HTTP/1.+1", 0, 0, false},
	{"HTTP/0000000001.1", 0, 0, false},
	{"HTTP/1.0000000001", 0, 0, false},
	{"HTTP/3.14", 0, 0, false},
	{"HTTP/12.3", 0, 0, false},
}

func TestParseHTTPVersion(t *testing.T) {
	for _, tt := range parseHTTPVersionTests {
		major, minor, ok := http.ParseHTTPVersion(tt.vers)
		if ok != tt.ok || major != tt.major || minor != tt.minor {
			type version struct {
				major, minor int
				ok           bool
			}
			t.Errorf("failed to parse %q, expected: %#v, got %#v", tt.vers, version{tt.major, tt.minor, tt.ok}, version{major, minor, ok})
		}
	}
}

type logWrites struct {
	t   *testing.T
	dst *[]string
}

func (l logWrites) WriteByte(c byte) error {
	l.t.Fatalf("unexpected WriteByte call")
	return nil
}

func (l logWrites) Write(p []byte) (n int, err error) {
	*l.dst = append(*l.dst, string(p))
	return len(p), nil
}

func TestRequestWriteBufferedWriter(t *testing.T) {
	got := []string{}
	req, _ := http.NewRequest("GET", "http://foo.com/", nil, nil)
	req.Write(logWrites{t, &got})
	want := []string{
		"GET / HTTP/1.1\r\n",
		"Host: foo.com\r\n",
		"User-Agent: " + defaultUserAgent + "\r\n",
		"\r\n",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Writes =\n %q\n  Want =\n %q", got, want)
	}
}

func TestRequestBadHostHeader(t *testing.T) {
	got := []string{}
	req, err := http.NewRequest("GET", "http://foo/after", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "foo.com\nnewline"
	req.URL.Host = "foo.com\nnewline"
	req.Write(logWrites{t, &got})
	want := []string{
		"GET /after HTTP/1.1\r\n",
		"Host: \r\n",
		"User-Agent: " + defaultUserAgent + "\r\n",
		"\r\n",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Writes = %q\n  Want = %q", got, want)
	}
}

func TestQuery(t *testing.T) {
	req := &http.Request{Method: "GET"}
	req.URL, _ = url.Parse("http://www.google.com/search?q=foo&q=bar")
	if q := req.FormValue("q"); q != "foo" {
		t.Errorf(`req.FormValue("q") = %q, want "foo"`, q)
	}
}

func TestNewRequestContentLength(t *testing.T) {
	readByte := func(r io.Reader) io.Reader {
		var b [1]byte
		r.Read(b[:])
		return r
	}
	tests := []struct {
		r    io.Reader
		want int64
	}{
		{bytes.NewReader([]byte("123")), 3},
		{bytes.NewBuffer([]byte("1234")), 4},
		{strings.NewReader("12345"), 5},
		{strings.NewReader(""), 0},
		{http.NoBody, 0},

		// Not detected. During Go 1.8 we tried to make these set to -1, but
		// due to Issue 18117, we keep these returning 0, even though they're
		// unknown.
		{struct{ io.Reader }{strings.NewReader("xyz")}, 0},
		{io.NewSectionReader(strings.NewReader("x"), 0, 6), 0},
		{readByte(io.NewSectionReader(strings.NewReader("xy"), 0, 6)), 0},
	}
	for i, tt := range tests {
		req, err := http.NewRequest("POST", "http://localhost/", nil, tt.r)
		if err != nil {
			t.Fatal(err)
		}
		if req.ContentLength != tt.want {
			t.Errorf("test[%d]: ContentLength(%T) = %d; want %d", i, tt.r, req.ContentLength, tt.want)
		}
	}
}

// verify that NewRequest sets Request.GetBody and that it works
func TestNewRequestGetBody(t *testing.T) {
	tests := []struct {
		r io.Reader
	}{
		{r: strings.NewReader("hello")},
		{r: bytes.NewReader([]byte("hello"))},
		{r: bytes.NewBuffer([]byte("hello"))},
	}
	for i, tt := range tests {
		req, err := http.NewRequest("POST", "http://foo.tld/", nil, tt.r)
		if err != nil {
			t.Errorf("test[%d]: %v", i, err)
			continue
		}
		if req.Body == nil {
			t.Errorf("test[%d]: Body = nil", i)
			continue
		}
		if req.GetBody == nil {
			t.Errorf("test[%d]: GetBody = nil", i)
			continue
		}
		slurp1, err := io.ReadAll(req.Body)
		if err != nil {
			t.Errorf("test[%d]: ReadAll(Body) = %v", i, err)
		}
		newBody, err := req.GetBody()
		if err != nil {
			t.Errorf("test[%d]: GetBody = %v", i, err)
		}
		slurp2, err := io.ReadAll(newBody)
		if err != nil {
			t.Errorf("test[%d]: ReadAll(GetBody()) = %v", i, err)
		}
		if string(slurp1) != string(slurp2) {
			t.Errorf("test[%d]: Body %q != GetBody %q", i, slurp1, slurp2)
		}
	}
}

// Issue 53181: verify Request.Cookie return the correct Cookie.
// Return ErrNoCookie instead of the first cookie when name is "".
func TestRequestCookie(t *testing.T) {
	for _, tt := range []struct {
		name        string
		value       string
		expectedErr error
	}{
		{
			name:        "foo",
			value:       "bar",
			expectedErr: nil,
		},
		{
			name: "",
			// value is ignored when name is "".
			expectedErr: http.ErrNoCookie,
		},
	} {
		req, err := http.NewRequest("GET", "http://example.com/", nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(&http.Cookie{Name: tt.name, Value: tt.value})
		c, err := req.Cookie(tt.name)
		if err != tt.expectedErr {
			t.Errorf("got %v, want %v", err, tt.expectedErr)
		}

		// skip if error occurred.
		if err != nil {
			continue
		}
		if c.Value != tt.value {
			t.Errorf("got %v, want %v", c.Value, tt.value)
		}
		if c.Name != tt.name {
			t.Errorf("got %s, want %v", tt.name, c.Name)
		}
	}
}

func TestParseFormQuery(t *testing.T) {
	req, _ := http.NewRequest(
		"POST",
		"http://www.google.com/search?q=foo&q=bar&both=x&prio=1&orphan=nope&empty=not",
		nil,
		strings.NewReader("z=post&both=y&prio=2&=nokey&orphan&empty=&"),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	if q := req.FormValue("q"); q != "foo" {
		t.Errorf(`req.FormValue("q") = %q, want "foo"`, q)
	}
	if z := req.FormValue("z"); z != "post" {
		t.Errorf(`req.FormValue("z") = %q, want "post"`, z)
	}
	if bq, found := req.PostForm["q"]; found {
		t.Errorf(`req.PostForm["q"] = %q, want no entry in map`, bq)
	}
	if bz := req.PostFormValue("z"); bz != "post" {
		t.Errorf(`req.PostFormValue("z") = %q, want "post"`, bz)
	}
	if qs := req.Form["q"]; !reflect.DeepEqual(qs, []string{"foo", "bar"}) {
		t.Errorf(`req.Form["q"] = %q, want ["foo", "bar"]`, qs)
	}
	if both := req.Form["both"]; !reflect.DeepEqual(both, []string{"y", "x"}) {
		t.Errorf(`req.Form["both"] = %q, want ["y", "x"]`, both)
	}
	if prio := req.FormValue("prio"); prio != "2" {
		t.Errorf(`req.FormValue("prio") = %q, want "2" (from body)`, prio)
	}
	if orphan := req.Form["orphan"]; !reflect.DeepEqual(orphan, []string{"", "nope"}) {
		t.Errorf(`req.FormValue("orphan") = %q, want "" (from body)`, orphan)
	}
	if empty := req.Form["empty"]; !reflect.DeepEqual(empty, []string{"", "not"}) {
		t.Errorf(`req.FormValue("empty") = %q, want "" (from body)`, empty)
	}
	if nokey := req.Form[""]; !reflect.DeepEqual(nokey, []string{"nokey"}) {
		t.Errorf(`req.FormValue("nokey") = %q, want "nokey" (from body)`, nokey)
	}
}

// Tests that we only parse the form automatically for certain methods.
func TestParseFormQueryMethods(t *testing.T) {
	for _, method := range []string{"POST", "PATCH", "PUT", "FOO"} {
		req, _ := http.NewRequest(method, "http://www.google.com/search",
			nil,
			strings.NewReader("foo=bar"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		want := "bar"
		if method == "FOO" {
			want = ""
		}
		if got := req.FormValue("foo"); got != want {
			t.Errorf(`for method %s, FormValue("foo") = %q; want %q`, method, got, want)
		}
	}
}

func TestParseFormUnknownContentType(t *testing.T) {
	for _, test := range []struct {
		name        string
		wantErr     string
		contentType http.Header
	}{
		{"text", "", http.Header{"Content-Type": {"text/plain"}}},
		// Empty content type is legal - may be treated as
		// application/octet-stream (RFC 7231, section 3.1.1.5)
		{"empty", "", http.Header{}},
		{"boundary", "mime: invalid media parameter", http.Header{"Content-Type": {"text/plain; boundary="}}},
		{"unknown", "", http.Header{"Content-Type": {"application/unknown"}}},
	} {
		t.Run(test.name,
			func(t *testing.T) {
				req := &http.Request{
					Method: "POST",
					Header: test.contentType,
					Body:   io.NopCloser(strings.NewReader("body")),
				}
				err := req.ParseForm()
				switch {
				case err == nil && test.wantErr != "":
					t.Errorf("unexpected success; want error %q", test.wantErr)
				case err != nil && test.wantErr == "":
					t.Errorf("want success, got error: %v", err)
				case test.wantErr != "" && test.wantErr != fmt.Sprint(err):
					t.Errorf("got error %q; want %q", err, test.wantErr)
				}
			},
		)
	}
}

func TestMultipartReader(t *testing.T) {
	tests := []struct {
		shouldError bool
		contentType string
	}{
		{false, `multipart/form-data; boundary="foo123"`},
		{false, `multipart/mixed; boundary="foo123"`},
		{true, `text/plain`},
	}

	for i, test := range tests {
		req := &http.Request{
			Method: "POST",
			Header: http.Header{"Content-Type": {test.contentType}},
			Body:   io.NopCloser(new(bytes.Buffer)),
		}
		multipart, err := req.MultipartReader()
		if test.shouldError {
			if err == nil || multipart != nil {
				t.Errorf("test %d: unexpectedly got nil-error (%v) or non-nil-multipart (%v)", i, err, multipart)
			}
			continue
		}
		if err != nil || multipart == nil {
			t.Errorf("test %d: unexpectedly got error (%v) or nil-multipart (%v)", i, err, multipart)
		}
	}
}

// Issue 9305: ParseMultipartForm should populate PostForm too
func TestParseMultipartFormPopulatesPostForm(t *testing.T) {
	postData :=
		`--xxx
Content-Disposition: form-data; name="field1"

value1
--xxx
Content-Disposition: form-data; name="field2"

value2
--xxx
Content-Disposition: form-data; name="file"; filename="file"
Content-Type: application/octet-stream
Content-Transfer-Encoding: binary

binary data
--xxx--
`
	req := &http.Request{
		Method: "POST",
		Header: http.Header{"Content-Type": {`multipart/form-data; boundary=xxx`}},
		Body:   io.NopCloser(strings.NewReader(postData)),
	}

	initialFormItems := map[string]string{
		"language": "Go",
		"name":     "gopher",
		"skill":    "go-ing",
		"field2":   "initial-value2",
	}

	req.Form = make(url.Values)
	for k, v := range initialFormItems {
		req.Form.Add(k, v)
	}

	err := req.ParseMultipartForm(10000)
	if err != nil {
		t.Fatalf("unexpected multipart error %v", err)
	}

	wantForm := url.Values{
		"language": []string{"Go"},
		"name":     []string{"gopher"},
		"skill":    []string{"go-ing"},
		"field1":   []string{"value1"},
		"field2":   []string{"initial-value2", "value2"},
	}
	if !reflect.DeepEqual(req.Form, wantForm) {
		t.Fatalf("req.Form = %v, want %v", req.Form, wantForm)
	}

	wantPostForm := url.Values{
		"field1": []string{"value1"},
		"field2": []string{"value2"},
	}
	if !reflect.DeepEqual(req.PostForm, wantPostForm) {
		t.Fatalf("req.PostForm = %v, want %v", req.PostForm, wantPostForm)
	}
}

func TestParseMultipartForm(t *testing.T) {
	req := &http.Request{
		Method: "POST",
		Header: http.Header{"Content-Type": {`multipart/form-data; boundary="foo123"`}},
		Body:   io.NopCloser(new(bytes.Buffer)),
	}
	err := req.ParseMultipartForm(25)
	if err == nil {
		t.Error("expected multipart EOF, got nil")
	}

	req.Header = http.Header{"Content-Type": {"text/plain"}}
	err = req.ParseMultipartForm(25)
	if err != http.ErrNotMultipart {
		t.Error("expected ErrNotMultipart for text/plain")
	}
}

// Issue 45789: multipart form should not include directory path in filename
func TestParseMultipartFormFilename(t *testing.T) {
	postData :=
		`--xxx
Content-Disposition: form-data; name="file"; filename="../usr/foobar.txt/"
Content-Type: text/plain

--xxx--
`
	req := &http.Request{
		Method: "POST",
		Header: http.Header{"Content-Type": {`multipart/form-data; boundary=xxx`}},
		Body:   io.NopCloser(strings.NewReader(postData)),
	}
	_, hdr, err := req.FormFile("file")
	if err != nil {
		t.Fatal(err)
	}
	if hdr.Filename != "foobar.txt" {
		t.Errorf("expected only the last element of the path, got %q", hdr.Filename)
	}
}

// func TestHandler(t *testing.T) {
//     req, err := http.NewRequest("GET", "", nil)
//     if err != nil {
//         t.Fatal(err)
//     }

//     recorder := httptest.NewRecorder()

//     hf := http.HandlerFunc(HelloWorldHandler)

//     hf.ServeHTTP(recorder, req)

//     if status := recorder.Code; status != http.StatusOK {
//         t.Errorf("handler returned wrong status code: got %v want %v",
//             status, http.StatusOK)
//     }

//     expected := `Hello, world!`
//     actual := recorder.Body.String()
//     if actual != expected {
//         t.Errorf("handler returned unexpected body: got %v want %v",
//             actual, expected)
//     }
// }

// compareReqToHttpRequest(want, got, t)

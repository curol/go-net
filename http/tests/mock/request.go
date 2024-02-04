package mock

import (
	"io"
	"strings"

	libhttp "net/http"

	http "github.com/curol/network/http"
	url "github.com/curol/network/url"
)

// MockPostJSONRequest mock request, which should be used in tests.
// Arguement `opts` is optional for overriding the default values.
func MockPostJSONRequest(opts *http.Request) *http.Request {
	// Default mock request
	req := &http.Request{
		// Request Line
		Method: "POST",
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
		},
		// Headers
		Header: http.Header{
			"Host":           []string{"example.com"},
			"Content-Type":   []string{"application/json"},
			"Content-Length": []string{"16"},
		},
		// Payload
		Body: io.NopCloser(strings.NewReader(`{"key": "value"}`)),
		// Other
		Proto:            "HTTP/1.1",
		ProtoMajor:       1,
		ProtoMinor:       1,
		ContentType:      "application/json",
		ContentLength:    int64(16),
		RequestURI:       "",
		RemoteAddress:    "",
		Form:             nil,
		PostForm:         nil,
		MultipartForm:    nil,
		Trailer:          nil,
		TransferEncoding: nil,
		TLS:              nil,
	}
	// Merge opts if not nil
	if opts != nil {
		req.Merge(opts)
	}
	return req
}

func MockStandardLibraryHttpPostJsonRequest(body io.Reader) *libhttp.Request {
	// data := strings.NewReader(`{"key": "value"}`)
	got, err := libhttp.NewRequest("POST", "http://example.com", body)
	if err != nil {
		return nil
	}
	got.Header.Set("Content-Type", "application/json")
	got.Header.Set("Content-Length", "16")
	got.Header.Set("Host", "example.com")
	return got
}

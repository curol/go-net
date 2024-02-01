package http

import (
	gohttp "net/http"
)

// Header is the metadata of the request or response.
//
// Note, it is a hashmap structure of key-value pairs.
type Header = gohttp.Header // map[string][]string

// NewHeader creates a new Header.
func NewHeader() Header {
	return Header(gohttp.Header{})
}

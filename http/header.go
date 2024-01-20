package http

import (
	net "net/http"
)

// Header is the metadata of the request or response.
//
// Note, it is a hashmap structure of key-value pairs.
type Header = net.Header // map[string][]string

// NewHeader creates a new Header.
func NewHeader() Header {
	return Header(net.Header{})
}

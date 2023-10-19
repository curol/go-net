package message

import "message/hashmap"

// Header is the metadata of the request or response.
//
// Note, it is a hashmap structure of key-value pairs.
type Header = hashmap.HashMap

// NewHeader creates a new Header.
func NewHeader() Header {
	return Header(hashmap.New())
}

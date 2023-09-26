package message // Header is the metadata of the request or response. It is a map of key-value pairs.

import "message/hashmap"

type Header = hashmap.HashMap

func NewHeader() Header {
	return Header(hashmap.New())
}

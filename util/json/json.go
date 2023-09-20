package json

import (
	"encoding/json"
	"io"
)

// *********************************************************************************************************************
// JSON
// *********************************************************************************************************************
// JSONUnmarshal parses the JSON-encoded data and stores the result in the value pointed to by v.
func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// JsonMarshal returns the JSON encoding of v.
func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// JsonMarshalIndent is like Marshal but applies Indent to format the output.
func MarshalIndent(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// JsonNewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *json.Decoder {
	return json.NewDecoder(r)
}

// DecodeJSON decodes r into v.
func Decode(r io.Reader, v any) error {
	decoder := NewDecoder(r)
	return decoder.Decode(v)
}

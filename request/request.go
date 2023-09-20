package request

import (
	"bufio"
	"bytes"
	"message"
	"net"
	"reader"
	"writer"
)

// Request properties
type Request struct {
	*message.Message
	r *reader.RequestReader
	w *writer.ResponseWriter
}

func NewRequest(con net.Conn) *Request {
	return newRequest(con)
}

func newRequest(con net.Conn) *Request {
	req := &Request{}
	// Arrange connection
	req.r = reader.NewRequestReader(con)
	req.w = writer.NewResponseWriter(con)
	// Read data from connection
	reqMessage, err := req.r.Read()
	if err != nil {
		panic(err)
	}
	// Set Message
	req.Message = reqMessage
	return req
}

func NewRequestFromBytes(data []byte) *Request {
	req := &Request{}
	// Get reader
	reader := bufio.NewReader(bytes.NewReader(data))
	// TODO: Arrange connection?
	req.r = nil
	req.w = nil
	// Read
	req.Message, _ = message.NewMessage(reader)
	return req
}

// ********************************************************************************//
// Getters
// ********************************************************************************//
func (req *Request) Reader() *reader.RequestReader {
	return req.r
}

func (req *Request) Writer() *writer.ResponseWriter {
	return req.w
}

// // Decode request body into v.
// // Uses json.Unmarshal function to decode a JSON string to a v object.
// func (req *Request) JSON(v any) error {
// 	// Data
// 	data := req.Body()
// 	// Decode
// 	return json.Unmarshal(data, v)
// }

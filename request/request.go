package request

import (
	"bufio"
	"bytes"
	"encoding/json"
	"message"
	"net"
	"reader"
	"writer"
)

// Request properties
type Request struct {
	// Request line
	// method  string // GET, POST, PUT, DELETE, etc.
	// path    string // /, /index.html, /about.html, etc.
	// version string // HTTP/1.1, HTTP/2.0, etc.
	// headers map[string]string
	// body    []byte
	// size    int
	// Message from client
	// message *message.Message
	*message.RequestMessage
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
	// Set RequestMessage
	req.RequestMessage = reqMessage
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
	req.RequestMessage, _ = message.NewRequestMessage(reader)
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

func (req *Request) Message() message.RequestMessage {
	return *req.RequestMessage.Copy()
}

// Decode request body into v.
// Uses json.Unmarshal function to decode a JSON string to a v object.
func (req *Request) JSON(v any) error {
	// Data
	data := req.Body()
	// Decode
	return json.Unmarshal(data, v)
}

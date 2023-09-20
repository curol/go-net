package message

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"util/ioutil"
	"util/slice"
	"util/url"
)

// Body is a wrapper for the body of a message.
type Body struct {
	// Body is the reader stream of bytes.
	// body io.ReadCloser
	r io.Reader
	// Len is the length of the body.
	// Can be -1 if length is unknown.
	// Message Header should have `Content-Length`.
	len int
	// Content-Type
	typ string
	// Data is the body as a byte slice.
	data []byte
}

// NewBody returns a new Body from the reader.
func NewBody(reader io.Reader, header *Header) *Body {
	body := new(Body)

	body.r = ioutil.NoCloser(reader)

	body.parseHeaders(header)

	// TODO: Read body into buffer?
	// buffer := make([]byte, length)
	// _, err = io.ReadFull(reader, buffer)
	// if err != nil {
	// 	return nil, err
	// }

	return body
}

func (b *Body) parseHeaders(header *Header) {
	// Get content length
	len, err := header.ContentLength()
	if err != nil {
		len = 0
		log.Println(err)
	}
	b.len = len

	// Get content type
	conTyp, err := header.ContentType()
	if err != nil {
		conTyp = ""
		log.Println(err)
	}
	b.typ = conTyp
}

//**********************************************************************************************************************
// Read
//**********************************************************************************************************************

// Read reads up to len(p) bytes into p and returns the number of bytes read and an error.
func (b *Body) Read(p []byte) (n int, err error) {
	return b.r.Read(p)
}

//**********************************************************************************************************************
// Write
//**********************************************************************************************************************

// ToBytes returns the body as a byte slice.
func (bo *Body) ToBytes() ([]byte, error) {
	// Create buffer
	var b bytes.Buffer
	// Write to buffer
	_, err := bo.WriteTo(&b)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Write writes b to writer.
func (bo *Body) Write(b []byte) (int, error) {
	// NewBuffer creates and initializes a new Buffer using buf as its initial contents.
	buf := bytes.NewBuffer(b)
	// Write data to buffer
	n, err := bo.WriteTo(buf)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

// WriteTo writes data to w until the buffer is drained or an error occurs.
func (b *Body) WriteTo(w io.Writer) (int64, error) {
	if b.len >= 0 {
		return 0, fmt.Errorf("can't have len >= 0")
	}

	// Write to buffer
	// n, err := w.Write(b.buffer)

	// CopyN copies n bytes (or until an error) from src to dst
	// Read from b.body and write to w of length b.len
	// return io.CopyN(w, b.body, int64(b.len))
	return ioutil.CopyReaderToWriter(w, b.r, int64(b.len))
}

//**********************************************************************************************************************
// Decode/Encode
//**********************************************************************************************************************

// Form decodes body into form values
func (b *Body) form(data string) (map[string][]string, error) {
	// TODO: Read body into buffer?
	// bodyBytes, err := io.ReadAll(b.body)

	// Return parse body into form values
	return url.DecodeForm(data)
}

// Equals returns true if b equals other.
func (b *Body) Equals(other *Body) bool {
	// Compare b.len to other.len
	if b.len != other.len {
		return false
	}

	// Compare b.data to other.data
	// TODO: Read b.body into b.data?
	return slice.BytesEqual(b.data, other.data)
}

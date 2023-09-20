package message

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"util/slice"
	"util/stream"
	"util/url"
)

// Body is a wrapper for the body of a message.
type Body struct {
	// The source of stream of bytes.
	r *bufio.Reader
	// Len is the size of the body.
	// Can be -1 if length is unknown.
	// Message Header should have `Content-Length`.
	size int
	// Content-Type of the body. Example: `text/html; charset=utf-8`
	typ string
	// Data is the body as a byte slice.
	data []byte
	// TODO: Set maxReadSize for max reading of stream
	// maxReadSize int
}

// NewBody returns a new Body from the reader.
func NewBody(reader io.Reader, header *Header) *Body {
	body := new(Body)

	body.r = bufio.NewReader(reader)

	// Parse header for content length and type
	body.parseHeadersForLengthAndType(header)

	// TODO: Read body into buffer?
	// buffer := make([]byte, length)
	// _, err = io.ReadFull(reader, buffer)
	// if err != nil {
	// 	return nil, err
	// }

	return body
}

func (b *Body) parseHeadersForLengthAndType(header *Header) {
	// Get content length
	size, err := header.ContentLength()
	if err != nil {
		size = 0
	}
	b.size = size

	// Get content type
	conTyp, err := header.ContentType()
	if err != nil {
		conTyp = ""
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
	// TODO: MAX READ SIZE

	// Create buffer
	var b bytes.Buffer // same as bytes.NewBuffer(nil)

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

	// Return bytes read and err
	return int(n), nil
}

// WriteTo writes data to w until the buffer is drained or an error occurs.
//
// Which, reads from b.r and writes to w of size b.len.
func (b *Body) WriteTo(w io.Writer) (int64, error) {

	// Check if size set
	if b.size <= 0 {
		return 0, fmt.Errorf("can't have len <= 0")
	}

	// Write to buffer
	// n, err := w.Write(b.buffer)

	// CopyN copies n bytes (or until an error) from src to dst
	// Read from b.body and write to w of length b.len
	// return io.CopyN(w, b.body, int64(b.len))
	return stream.CopyReaderToWriter(w, b.r, int64(b.size))
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
	if b.size != other.size {
		return false
	}

	// Compare b.data to other.data
	// TODO: Read b.body into b.data?
	return slice.BytesEqual(b.data, other.data)
}

func (b *Body) Size() int {
	return b.size
}

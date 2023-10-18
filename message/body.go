package message

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
)

// Body represents the body of an HTTP request or response.
type Body struct {
	rw   io.ReadWriter
	f    *os.File
	size int64
}

func NewBody() *Body {
	return &Body{
		rw:   nil,
		f:    nil,
		size: 0,
	}
}

// JSON writes JSON to the response body
func (b *Body) JSON(v any) (int, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}
	return b.write(jsonBytes)
}

// Text writes text to the response body
func (b *Body) Text(s string) (int, error) {
	return b.write([]byte(s))
}

// File sets a file to the response body
func (b *Body) File(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	fileInfo, err := f.Stat()
	if err != nil {
		return err
	}
	b.rw = nil
	b.f = f
	b.size = fileInfo.Size()
	return nil
}

// Read reads response body
func (b *Body) Read(p []byte) (int, error) {
	if b.rw != nil {
		return b.rw.Read(p)
	}
	if b.f != nil {
		return b.f.Read(p)
	}
	return 0, io.EOF
}

func (b *Body) Close() error {
	if b.f != nil {
		return b.f.Close()
	}
	b.reset()
	return nil
}

func (b *Body) reset() {
	b.rw = nil
	b.f = nil
	b.size = 0
}

func (b *Body) Size() int64 {
	return b.size
}

func (b *Body) IsContents() bool {
	return b.rw != nil || b.f != nil
}

// newBuffer creates a new buffer and writes v to it
func (b *Body) newBuffer(v []byte) {
	b.rw = bytes.NewBuffer(v) // v is written to the buffer as the initial contents
}

func (b *Body) write(v []byte) (int, error) {
	if b.rw == nil {
		b.newBuffer(v) // set initial contents
		b.size = int64(len(v))
		return len(v), nil
	}
	n, err := b.rw.Write(v) // write v to rw
	if err != nil {
		return 0, err
	}
	b.size += int64(n)
	return n, nil
}

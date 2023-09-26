package message

import (
	"io"
)

type Body struct {
	body   io.Reader
	closer io.Closer
}

func NewBody(rc io.ReadCloser) *Body {
	return &Body{
		body:   rc,
		closer: rc,
	}
}

func (b *Body) Read(p []byte) (int, error) {
	return b.body.Read(p)
}

func (b *Body) Close() error {
	return b.closer.Close()
}

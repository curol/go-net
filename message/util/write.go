package util

import (
	"bufio"
	"io"
)

// NewWriter returns a new Writer whose buffer has the default size.
func newWriter(w io.Writer) *bufio.Writer {
	return bufio.NewWriter(w)
}

// NewReadWriter returns a new ReadWriter with the given buffer size.
func newReadWriter(r io.Reader, w io.Writer) *bufio.ReadWriter {
	return bufio.NewReadWriter(newReader(r), newWriter(w))
}

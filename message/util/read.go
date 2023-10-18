package util

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

//**********************************************************************************************************************
// Reader/Writer
//**********************************************************************************************************************

// CopyReaderToWriter copies from src to dst until either EOF is reached on src or an error occurs. It returns the number of bytes copied and the first error encountered while copying, if any.
// A successful CopyReader returns err == nil, not err == EOF. Because Copy is defined to read from src until EOF, it does not treat an EOF from Read as an error to be reported.
// If src implements the WriterTo interface, the copy is implemented by calling src.WriteTo(dst). Otherwise, if dst implements the ReaderFrom interface, the copy is implemented by calling dst.ReadFrom(src).
func CopyReaderToWriter(dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, src)
}

// CopyReaderToWriterN copies n bytes (or until an error) from src to dst. It returns the number of bytes copied and the earliest error encountered while copying. On return, written == n if and only if err == nil.
//
// If dst implements the ReaderFrom interface, the copy is implemented using it.
func CopyReaderToWriterN(dst io.Writer, src io.Reader, size int64) (int64, error) {
	// CopyN copies n bytes (or until an error) from src to dst
	// Read from b.body and write to w of length b.len
	return io.CopyN(dst, src, size)
}

// ReadAll reads from r until an error or EOF and returns the data it read. A successful call returns err == nil, not err == EOF. Because ReadAll is defined to read from src until EOF, it does not treat an EOF from Read as an error to be reported.
func readAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

// ReadFull reads exactly len(buf) bytes from r into buf. It returns the number of bytes copied and an error if fewer bytes were read. The error is EOF only if no bytes were read. If an EOF happens after reading some but not all the bytes, ReadFull returns ErrUnexpectedEOF. On return, n == len(buf) if and only if err == nil. If r returns an error having read at least len(buf) bytes, the error is dropped.
func readN(r io.Reader, n int64) ([]byte, error) {
	buf := make([]byte, n)
	_, err := io.ReadFull(r, buf)
	return buf, err
}

// NewReaderFromBytes returns a new Reader reading from b.
func newReaderFromBytes(b []byte) *bytes.Reader {
	return bytes.NewReader(b)
}

// NewReaderFromStrings returns a new Reader reading from s.
func newReaderFromStrings(s string) *strings.Reader {
	return strings.NewReader(s)
}

// NewReader returns a new Reader from a reader. The underlying buffer.
func newReader(r io.Reader) *bufio.Reader {
	return bufio.NewReader(r)
}

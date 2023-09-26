package message

import "io"

// CopyN copies n bytes (or until an error) from src to dst.
func copyN(dst io.Writer, src io.Reader, size int64) (int64, error) {
	// CopyN copies n bytes (or until an error) from src to dst
	// Read from b.body and write to w of length b.len
	return io.CopyN(dst, src, size)
}

// TODO: Add a RequestReader interface
// type RequestReader interface {
// 	Method() string
// 	Path() string
// 	Protocol() string
// 	Body() []byte
// 	ContentLength() int
// 	ContentType() string
// }

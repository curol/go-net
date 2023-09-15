package writer

import (
	"bufio"
	"encoding/json"
	"net"
)

// ResponseWriter manages the response to client.
type ResponseWriter struct {
	w *bufio.Writer
}

// NewResponseWriter returns a new ResponseWriter.
func NewResponseWriter(conn net.Conn) *ResponseWriter {
	return &ResponseWriter{
		w: bufio.NewWriter(conn),
	}
}

// Writer returns the writer for the ResponseWriter.
func (r *ResponseWriter) Writer() *bufio.Writer {
	return r.w
}

// Write writes to writer and return number of bytes written
func (r *ResponseWriter) Write(p []byte) int {
	return r.writeAndFlush(p)
}

// Flush writes any buffered data to the underlying io.Writer.
func (r *ResponseWriter) Flush() error {
	return r.w.Flush()
}

func (r *ResponseWriter) writeAndFlush(p []byte) int {
	// Write writes the contents of p into the buffer.
	n, err := r.w.Write(p)
	if err != nil {
		panic(err)
	}

	// Flush writes any buffered data to the underlying io.Writer.
	err = r.Flush()
	if err != nil {
		panic(err)
	}

	// Return number of bytes written
	return n
}

// Text writes plain text to writer.
func (r *ResponseWriter) Text(v string) {
	r.w.WriteString(v)
	r.w.Flush()
}

// JSON writes json to writer and return bytes written.
func (r *ResponseWriter) JSON(v any) (int, error) {
	result, err := json.Marshal(v)
	if err != nil {
		return -1, err
	}
	return r.Write(result), nil
}

// func (c *Connection) ResponseWriter() *ResponseWriter {
// 	return c.Writer()
// }

// // Return connection's writer
// func (c *Connection) Writer() *bufio.Writer {
// 	return c.writer.Writer()
// }

// // Write to connection
// func (c *Connection) Write(p []byte) int {
// 	return c.writer.Write(p)
// }

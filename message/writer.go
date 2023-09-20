package message

import (
	"bytes"
	"io"
)

type ResponseWriter interface {
	Write(b []byte) (int, error)
	WriteTo(w io.Writer) (int64, error)
	Header() map[string]string
}

// Write writes b to w.
func writeMessage(w ResponseWriter, b []byte) (int, error) {
	// NewBuffer creates and initializes a new Buffer
	// using b as its initial contents.
	buf := bytes.NewBuffer(b)
	// Write data to buffer
	n, err := w.WriteTo(buf)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

// ToBytes returns the headers as a byte slice.
func toBytes(w ResponseWriter) ([]byte, error) {
	var buf bytes.Buffer
	_, err := w.WriteTo(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// String returns the text of the header formatted in the same way as in the request.
func toString(w ResponseWriter) string {
	var buf bytes.Buffer
	w.WriteTo(&buf)
	return buf.String()
}

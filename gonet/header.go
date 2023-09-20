package gonet

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Header contains the headers of a message.
// Headers can indeed have values that contain commas. According to the HTTP/1.1 specification, multiple message-header fields with the same field-name may be present in a message if and only if the entire field-value for that header field is defined as a comma-separated list.
// For example, Accept: text/plain, text/html
// For Brevity, we will not be supporting headers with multiple values.
type Header map[string]string

// NewHeader returns a new Header.
func NewHeader() Header {
	return make(Header)
}

// Set sets the header's value.
func (h Header) Set(key, value string) {
	k := strings.TrimSpace(key)
	v := strings.TrimSpace(value)
	h[k] = v
}

// Get gets the value associated with the given key.
func (h Header) Get(key string) (string, bool) {
	k := strings.TrimSpace(key)
	if values, ok := h[k]; !ok || len(values) == 0 {
		return "", false
	}
	return h[key], true
}

// Del deletes the values associated with key.
func (h Header) Del(key string) {
	k := strings.TrimSpace(key)
	delete(h, k)
}

// Clone creates a new Header with the same keys and values as the original.
// It does not create deep copies of the values, so changes to the original
// Header may affect the copied Header if the values are pointers or slices.
func (h Header) Clone() Header {
	h2 := make(Header)
	for k, v := range h {
		h2[k] = v
	}
	return h2
}

// Len returns the number of headers.
func (h Header) Len() int {
	return len(h)
}

// Keys returns the keys of the header.
func (h Header) Keys() []string {
	var keys []string
	for k := range h {
		keys = append(keys, k)
	}
	return keys
}

// Values returns the values of the header.
func (h Header) Values() []string {
	var values []string
	for _, v := range h {
		values = append(values, v)
	}
	return values
}

// Merge merges two Headers.
func (h Header) Merge(other Header) {
	for k, v := range other {
		h[k] = v
	}
}

// String returns the text of the header formatted in the same way as in the request.
func (h Header) ToString() string {
	var buf bytes.Buffer
	h.WriteTo(&buf)
	return buf.String()
}

// ToBytes returns the headers as a byte slice.
func (h Header) ToBytes() []byte {
	var buf bytes.Buffer
	h.WriteTo(&buf)
	return buf.Bytes()
}

// ToSlice returns the headers as a slice of strings.
func (h Header) ToSlice() []string {
	// Extract the keys and sort them
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	// Sort keys
	sort.Strings(keys)
	// Create a new slice with sorted headers
	sortedHeaders := make([]string, 0, len(h))
	for _, k := range keys {
		sortedHeaders = append(sortedHeaders, fmt.Sprintf("%s: %s", k, h[k]))
	}
	return sortedHeaders
}

// contains checks if a slice contains a string.
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// ******************************************************
// Read/Write
// ******************************************************
// ReadFrom reads a sequence of headers from r until io.EOF and adds them to the Header.
// It returns the number of bytes read. If an error occurs before io.EOF, it returns the number of bytes read so far and the error.
// A successful ReadFrom returns err == nil, not err == io.EOF. Because ReadFrom is defined to read from src until EOF, it does not treat an EOF from Read as an error to be reported.
// If the header line is invalid, it returns an error
func (h Header) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	for {
		line, err := readLine(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return read, err
		}
		read += int64(len(line))
		parts := bytes.SplitN(line, []byte{':'}, 2)
		if len(parts) < 2 {
			return read, errors.New("invalid header line")
		}
		key := string(bytes.TrimSpace(parts[0]))
		value := string(bytes.TrimSpace(parts[1]))
		h.Set(key, value)
	}
	return read, nil
}

// readLine reads a line from r until it finds a \n or io.EOF.
func readLine(r io.Reader) ([]byte, error) {
	var line []byte
	buf := make([]byte, 1)

	for {
		_, err := r.Read(buf)
		if err != nil {
			return line, err
		}
		if buf[0] == '\n' {
			break
		}
		line = append(line, buf[0])
	}

	return line, nil
}

// Write writes a sequence of headers to w in the HTTP/1.1 header format.
func (h Header) Write(b []byte) (int, error) {
	n, err := h.WriteTo(bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

// WriteTo writes a sequence of headers to w in the HTTP/1.1 header format.
func (h Header) WriteTo(w io.Writer) (int64, error) {
	var written int64
	for k, vs := range h {
		for _, v := range vs {
			n, err := fmt.Fprintf(w, "%s: %s\r\n", k, v)
			written += int64(n)
			if err != nil {
				return written, err
			}
		}
	}
	return written, nil
}

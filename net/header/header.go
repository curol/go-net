package header

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/textproto"
	"slices"
)

// A MIMEHeader represents a MIME-style header mapping.
type MIMEHeader map[string][]string

type Header map[string][]string

func NewHeader(m map[string][]string) Header {
	h := make(Header)
	if m != nil {
		h.FromMap(m)
	}
	return h
}

func (h Header) FromMap(m map[string][]string) {
	for k, v := range m {
		for _, vv := range v {
			h.Add(k, vv)
		}
	}
}

func (h Header) Add(k, v string) {
	k = textproto.CanonicalMIMEHeaderKey(k)
	h[k] = append(h[k], v)
}

func (h Header) Set(k, v string) {
	k = textproto.CanonicalMIMEHeaderKey(k)
	h[k] = []string{v}
}

func (h Header) Get(k string) string {
	k = textproto.CanonicalMIMEHeaderKey(k)
	if v, ok := h[k]; ok {
		return v[0]
	}
	return ""
}

func (h Header) Del(k string) {
	k = textproto.CanonicalMIMEHeaderKey(k)
	delete(h, k)
}

func (h Header) Has(k string) bool {
	k = textproto.CanonicalMIMEHeaderKey(k)
	_, ok := h[k]
	return ok
}

func (h Header) Keys() []string {
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func (h Header) Clone() Header {
	nh := make(Header)
	for k, v := range h {
		nh[k] = v
	}
	return nh
}

func (h Header) Write(w io.Writer) (int64, error) {
	var n int64
	var err error
	switch v := w.(type) {
	case *bufio.Writer:
		n, err = h.write(v)
	default:
		n, err = h.write(bufio.NewWriter(w))
	}
	return n, err
}

func (h Header) write(w *bufio.Writer) (int64, error) {
	var n int64
	for k, v := range h {
		for _, vv := range v {
			// nn, err := w.WriteString(k + ": " + vv + "\r\n")
			nn, err := fmt.Fprintf(w, "%s: %s\r\n", k, vv)
			n += int64(nn)
			if err != nil {
				return n, err
			}
		}
	}
	return n, w.Flush()
}

func (h Header) Read(r io.Reader) (int, error) {
	var buf *bufio.Reader

	switch v := r.(type) {
	case *bufio.Reader:
		buf = v
	default:
		buf = bufio.NewReader(v)
	}

	tp := textproto.NewReader(buf)
	m, err := tp.ReadMIMEHeader()
	if err != nil {
		return -1, err
	}
	h.FromMap(m)
	return buf.Buffered(), nil
}

func SerializeHeaders(headers map[string][]string) []byte {
	buffer := bytes.NewBuffer(nil)

	for name, values := range headers {
		for _, value := range values {
			buffer.WriteString(name)
			buffer.WriteString(": ")
			buffer.WriteString(value)
			buffer.WriteString("\r\n")
		}
	}

	return buffer.Bytes()
}
func GetHeaderSize(headers map[string][]string) int {
	size := 0
	for name, values := range headers {
		size += len(name) // Add length of header name
		for _, value := range values {
			size += len(value) // Add length of each header value
		}
		size += 4 // Add length of ": \r\n"
	}
	return size
}

func ParseHeaders(r *bufio.Reader) (map[string][]string, int, error) {
	tp := textproto.NewReader(r)
	h, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, -1, err
	}
	n := GetHeaderSize(h)
	return h, n, nil
}

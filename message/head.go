package message

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type head struct {
	// RequestLine is the first line of the message.
	rl *RequestLine
	// Header is the metadata of a message in key-value pairs.
	header *Header
	// Len is the length of the head.
	size int
	//
	reader *bufio.Reader
}

func NewHead(reader *bufio.Reader) (*head, error) {
	h := new(head)
	h.reader = reader
	err := h.parse(reader)
	return h, err
}

func (h *head) parse(reader *bufio.Reader) error {
	rl, err := NewRequestLine(reader)
	if err != nil {
		return err
	}

	// Read and parse headers
	header, err := NewHeader(reader)
	if err != nil {
		return err
	}

	h.rl = rl
	h.header = header
	h.size = rl.size + header.Size()

	return nil
}

func (h *head) ContentLength() (int, error) {
	return h.header.ContentLength()
}

func (h *head) Size() int {
	return h.size
}

func (h *head) RequestLine() *RequestLine {
	return h.rl
}

func (h *head) Header() *Header {
	return h.header
}

func (h *head) Strings() []string {
	lines := []string{
		fmt.Sprintf("\t- Head: %p", h),
		fmt.Sprintf("\t\t- Size: %d", h.size),
		fmt.Sprintf("\t\t- Request Line: %p", h.rl),
		fmt.Sprintf("\t\t\t- Method: %s", h.rl.method),
		fmt.Sprintf("\t\t\t- Path: %s", h.rl.path),
		fmt.Sprintf("\t\t\t- Protocol: %s", h.rl.protocol),
		fmt.Sprintf("\t\t- Header: %p", h.header),
	}
	heds := h.Header().ToStrings()
	for _, h := range heds {
		lines = append(lines, fmt.Sprintf("\t\t\t- %s", h))
	}
	return lines
}

func (h *head) Method() string {
	return h.rl.method
}

func (h *head) Path() string {
	return h.rl.path
}

func (h *head) Protocol() string {
	return h.rl.protocol
}

// Equals checks if two heads are equal.
func (h *head) Equals(other *head) bool {
	if h.Size() != other.Size() {
		return false
	}
	// Request line
	if !h.RequestLine().Equals(other.RequestLine()) {
		return false
	}
	// TODO: Check if header is equal
	// Headers
	// if !h.Header().Equals(other.Header()) {
	// 	return false
	// }
	return true
}

func (h *head) ToBytes() ([]byte, error) {
	var b bytes.Buffer
	_, err := h.WriteTo(&b)
	if err != nil {
		return nil, err
	}
	// writer := bufio.NewWriter(&b)
	// h.rl.Write(writer)
	// h.header.WriteTo(writer)
	// // Ensure all data has been written to the underlying buffer
	// writer.Flush()
	return b.Bytes(), nil
}

func (h *head) Write(b []byte) (int, error) {
	n, err := h.WriteTo(bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func (h *head) WriteTo(w io.Writer) (int64, error) {
	// Request line
	n, err := h.rl.WriteTo(w)
	if err != nil {
		return int64(n), err
	}
	// Header
	n2, err := h.header.WriteTo(w)
	if err != nil {
		return int64(n) + n2, err
	}
	return int64(n) + n2, nil
}

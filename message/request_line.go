package message

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// Request line is the first line of the message.
// The format is: <method> <path> <protocol>
type RequestLine struct {
	method   string
	path     string
	protocol string
	len      int
}

func NewRequestLine(reader *bufio.Reader) (*RequestLine, error) {
	rl := new(RequestLine)
	err := rl.parse(reader)
	if err != nil {
		return nil, err
	}
	return rl, nil
}

func (r *RequestLine) Method() string {
	return r.method
}

func (r *RequestLine) Path() string {
	return r.path
}

func (r *RequestLine) Protocol() string {
	return r.protocol
}

func (r *RequestLine) Len() int {
	return r.len
}

func (r *RequestLine) ToString() string {
	return fmt.Sprintf("%s %s %s", r.method, r.path, r.protocol)
}

func (r *RequestLine) ToBytes() []byte {
	return []byte(r.ToString())
}

func (r *RequestLine) Equals(rl *RequestLine) bool {
	return r.method == rl.method && r.path == rl.path && r.protocol == rl.protocol
}

func (r *RequestLine) parse(reader *bufio.Reader) error {
	rl, err := parseRequestLine(reader)
	if err != nil {
		return err
	}
	r.method = rl.method
	r.path = rl.path
	r.protocol = rl.protocol
	r.len = rl.len
	return nil
}

// ParseRequestLine reads the first line of the reader and returns the method, path, and protocol.
type requestLine struct {
	method   string
	path     string
	protocol string
	len      int
}

func parseRequestLine(reader *bufio.Reader) (*requestLine, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	requestLineComponents := strings.Split(strings.TrimSpace(line), " ")
	if len(requestLineComponents) != 3 {
		return nil, fmt.Errorf("Malformed request line. Expected format: <method> <path> <protocol>")
	}
	return &requestLine{method: requestLineComponents[0], path: requestLineComponents[1], protocol: requestLineComponents[2], len: len(line)}, nil
}

func (r *RequestLine) Write(b []byte) (int, error) {
	n, err := r.WriteTo(bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func (r *RequestLine) WriteTo(w io.Writer) (int64, error) {
	v := fmt.Sprintf("%s\r\n", r.ToString())
	n, err := fmt.Fprintf(w, v)
	return int64(n), err
}

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
	size     int
}

func NewRequestLine(reader *bufio.Reader) (*RequestLine, error) {
	rl := new(RequestLine)
	err := rl.parse(reader)
	return rl, err
}

func (r *RequestLine) parse(reader *bufio.Reader) error {
	rl, err := parseRequestLine(reader)
	if err != nil {
		return err
	}
	r.method = rl.method
	r.path = rl.path
	r.protocol = rl.protocol
	r.size = rl.len
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
	return &requestLine{
			method:   requestLineComponents[0],
			path:     requestLineComponents[1],
			protocol: requestLineComponents[2],
			len:      len(line),
		},
		nil
}

//**********************************************************************************************************************
// Getters
//**********************************************************************************************************************

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
	return r.size
}

func (r *RequestLine) Equals(rl *RequestLine) bool {
	return r.method == rl.method && r.path == rl.path && r.protocol == rl.protocol
}

//**********************************************************************************************************************
// Writers
//**********************************************************************************************************************

// ToString returns the request line as a string.
func (r *RequestLine) ToString() string {
	return string(r.ToBytes())
}

// ToBytes returns the request line as a byte slice.
func (r *RequestLine) ToBytes() []byte {
	b := bytes.NewBuffer(nil)
	_, err := r.WriteTo(b)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

// WriteTo writes the the request line to w.
func (r *RequestLine) WriteTo(w io.Writer) (int64, error) {
	s := fmt.Sprintf("%s %s %s\r\n", r.method, r.path, r.protocol)
	// v := fmt.Sprintf("%s\r\n", r.ToString())
	// n, err := fmt.Fprintf(w, v)
	writer := bufio.NewWriter(w)
	n, err := writer.Write([]byte(s))
	return int64(n), err
}

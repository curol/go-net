package message

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

// Message represents the raw request message.
type Message struct {
	*head
	// Body is the payload or content of the message.
	body *Body
	// Reader
	r *bufio.Reader
	// Length
	size int
	// Max bytes to read
	maxReadSize int
}

// Returns Message from reader
func NewMessage(reader any) (*Message, error) {
	message := &Message{}

	// TODO: Make maxReadSize configurable
	// message.maxReadSize = 1024 * 1024 * 10 // 10 MB

	// Arrange
	switch v := reader.(type) {
	case net.Conn:
		message.r = bufio.NewReader(v)
	case *bufio.Reader:
		message.r = bufio.NewReader(v)
	case []byte:
		message.r = bufio.NewReader(bytes.NewReader(v))
	case io.Reader:
		message.r = bufio.NewReader(v)
	default:
		return nil, fmt.Errorf("Invalid reader type")
	}

	// Head
	h, err := newHead(message.r)
	if err != nil {
		return nil, err
	}
	// message.head = *h
	message.head = h

	// Body
	message.body = NewBody(message.r, message.head.header)

	// Length
	message.size = message.head.Size() + message.body.Size()

	return message, nil
}

// Get Method
func (rm *Message) Method() string {
	return rm.head.rl.method
}

// Get Path
func (rm *Message) Path() string {
	return rm.head.rl.path
}

// Get Protocol
func (rm *Message) Protocol() string {
	return rm.head.rl.protocol
}

// Get Header
func (rm *Message) Header() *Header {
	return rm.head.header
}

func (rm *Message) Body() *Body {
	return rm.body
}

func (m *Message) Reader() *bufio.Reader {
	return m.r
}

// Len returns the length of the message, head, and body.
func (m *Message) Size() (int, int, int) {
	cl, err := m.head.ContentLength()
	if err != nil {
		cl = 0
	}
	return m.size, m.head.Size(), cl
}

func (m *Message) String() string {
	mes := m

	// Message
	mesLen, _, bodyLen := m.Size()
	lines := []string{
		fmt.Sprintf("Message: %p", mes),
		fmt.Sprintf("\t- Size: %d", mesLen),
	}

	// Head
	lines = append(lines, m.Strings()...)

	// Body
	lines = append(lines, fmt.Sprintf("\t- Body: %p", mes.body))
	lines = append(lines, fmt.Sprintf("\t\t- Size: %d", bodyLen))

	// Reader
	lines = append(lines, fmt.Sprintf("\t- Reader: %p", mes.r))

	return strings.Join(lines, "\n")
}

// Read
func (m *Message) Read(p []byte) (n int, err error) {
	return m.r.Read(p)
}

func (m *Message) Equals(other *Message) bool {
	// TODO: Reflect?
	// 	reflect.DeepEqual(m, other)

	// Check if both messages are nil
	if m == nil && other == nil {
		return true
	}
	if m == nil || other == nil {
		return false
	}

	// Compare head fields
	if !m.head.Equals(other.head) {
		return false
	}

	// Compare body fields
	if !m.body.Equals(other.body) {
		return false
	}

	// Compare len fields
	if m.size != other.size {
		return false
	}

	return true
}

// Write to file
func (m *Message) ToFile(fn string) {
	// File stream
	f, err := os.Create(fn)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	// Write head to file
	n, err := m.head.WriteTo(f)
	if err != nil && err != io.EOF {
		panic(err)
	}
	fmt.Printf("Sent file %d bytes.\n", n)

	// Write body to file
	n2, err := m.body.WriteTo(f)
	if err != nil && err != io.EOF {
		panic(err)
	}
	fmt.Printf("Sent file %d bytes.\n", n2)

	// Write bytes to file
	fmt.Printf("%d bytes written to file %s\n", n+n2, fn)
}

func (m *Message) ToBytes() []byte {

	buf := bytes.NewBuffer(nil)

	head, err := m.head.ToBytes()
	if err != nil && err != io.EOF {
		panic(err)
	}

	body, err := m.body.ToBytes()
	if err != nil && err != io.EOF {
		panic(err)
	}

	buf.Write(head)
	buf.Write([]byte("\r\n"))
	buf.Write(body)

	return buf.Bytes()
}

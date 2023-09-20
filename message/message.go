package message

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
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
	len int
}

// Returns Message from reader
func NewMessage(reader interface{}) (*Message, error) {
	message := &Message{}

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
	cl, _ := message.head.ContentLength()
	message.body = NewBody(message.r, message.head.header)

	// Length
	message.len = message.head.Len() + cl

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
func (m *Message) Len() (int, int, int) {
	cl, err := m.head.ContentLength()
	if err != nil {
		log.Println(err)
		cl = 0
	}
	return m.len, m.head.Len(), cl
}

func (m *Message) String() string {
	mes := m

	// Message
	mesLen, _, bodyLen := m.Len()
	lines := []string{
		fmt.Sprintf("Message: %p", mes),
		fmt.Sprintf("\t- Length: %d", mesLen),
	}

	// Head
	lines = append(lines, m.Strings()...)

	// Body
	lines = append(lines, fmt.Sprintf("\t- Body: %p", mes.body))
	lines = append(lines, fmt.Sprintf("\t\t- Length: %d", bodyLen))

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
	if m.len != other.len {
		return false
	}

	return true
}

// Write to file
func (m *Message) ToFile(fn string) {
	f, err := os.Create(fn)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	// fmt.Fprintln(f, m.String())
	b := m.ToBytes()
	n, err := f.Write(b)
	if err != nil {
		log.Println("Error writing to file:", err)
	}
	fmt.Printf("%d bytes written to file %s\n", n, fn)
}

func (m *Message) ToBytes() []byte {
	b, err := m.head.ToBytes()
	if err != nil {
		log.Fatal(err)
	}
	return b
}

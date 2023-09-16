package gonet

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"sort"
)

// Message represents the request message sent from the client.
type Message struct {
	// Request line
	method   string
	path     string
	protocol string
	// Headers
	headers map[string]string
	// Payload
	body []byte
}

// Returns Message from reader
func NewMessage(reader *bufio.Reader) (*Message, error) {
	return parse(reader)
}

// Returns Message from connection
func NewMessageFromConnection(conn net.Conn) (*Message, error) {
	reader := bufio.NewReader(conn)
	return parse(reader)
}

// Returns Message from bytes
func NewMessageFromBytes(data []byte) (*Message, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	return parse(reader)
}

// Convert Message to bytes
func (rm *Message) ToBytes() []byte {
	// Format the request line
	requestLine := fmt.Sprintf("%s %s %s\r\n", rm.method, rm.path, rm.protocol)

	// Format the headers
	headers := ""
	for _, v := range rm.HeadersToSlice() {
		headers += fmt.Sprintf("%s\r\n", v)
		// headers += fmt.Sprintf("%s: %s\r\n", name, value)
	}

	// Combine the request line, headers, and body
	request := requestLine + headers + "\r\n" + string(rm.body)

	// Convert the request to bytes
	return []byte(request)
}

// Convert Message to map
func (rm *Message) ToMap() map[string]string {
	return map[string]string{
		"method":   rm.method,
		"path":     rm.path,
		"protocol": rm.protocol,
		"headers":  fmt.Sprintf("%v", rm.Headers()),
		"body":     string(rm.body),
	}
}

// Compares two Messages
func (rm *Message) Equals(other *Message) bool {
	return bytes.Equal(rm.ToBytes(), other.ToBytes())
}

// Get Method
func (rm *Message) Method() string {
	return rm.method
}

// Get Path
func (rm *Message) Path() string {
	return rm.path
}

// Get Protocol
func (rm *Message) Protocol() string {
	return rm.protocol
}

// Get Headers
func (rm *Message) Headers() map[string]string {
	return rm.headers
}

// Headers sorted by name
func (rm *Message) HeadersToSlice() []string {
	// Extract the keys and sort them
	keys := make([]string, 0, len(rm.headers))
	for k := range rm.headers {
		keys = append(keys, k)
	}

	// Sort keys
	sort.Strings(keys)

	// Create a new slice with sorted headers
	sortedHeaders := make([]string, 0, len(rm.headers))
	for _, k := range keys {
		sortedHeaders = append(sortedHeaders, fmt.Sprintf("%s: %s", k, rm.headers[k]))
	}

	return sortedHeaders
}

// Print Headers
func (rm *Message) PrintHeaders() {
	for _, value := range rm.HeadersToSlice() {
		fmt.Println(value)
	}
}

// Get Body
func (rm *Message) Body() []byte {
	return rm.body
}

// Print
func (rm *Message) Print() {
	fmt.Println("Method:", rm.method)
	fmt.Println("Path:", rm.path)
	fmt.Println("Protocol:", rm.protocol)
	fmt.Println("Headers:", rm.headers)
	fmt.Println("Bytes:", rm.ToBytes())
	fmt.Println("Map:", rm.ToMap())
	fmt.Println("Size:", len(rm.ToBytes()))
	fmt.Println("Body:", string(rm.body))
	rm.PrintHeaders()
}

// Copy Message and return new Message
func (rm *Message) Copy() *Message {
	return &Message{
		method:   rm.method,
		path:     rm.path,
		protocol: rm.protocol,
		headers:  rm.headers,
		body:     rm.body,
	}
}

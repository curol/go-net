package message

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"sort"
)

// RequestMessage represents the request message sent from the client.
type RequestMessage struct {
	// Request line
	method   string
	path     string
	protocol string
	// Headers
	headers map[string]string
	// Payload
	body []byte
}

// Returns RequestMessage from reader
func NewRequestMessage(reader *bufio.Reader) (*RequestMessage, error) {
	return parse(reader)
}

// Returns RequestMessage from connection
func NewRequestMessageFromConnection(conn net.Conn) (*RequestMessage, error) {
	reader := bufio.NewReader(conn)
	return parse(reader)
}

// Returns RequestMessage from bytes
func NewRequestMessageFromBytes(data []byte) (*RequestMessage, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	return parse(reader)
}

// Convert RequestMessage to bytes
func (rm *RequestMessage) ToBytes() []byte {
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

// Convert RequestMessage to map
func (rm *RequestMessage) ToMap() map[string]string {
	return map[string]string{
		"method":   rm.method,
		"path":     rm.path,
		"protocol": rm.protocol,
		"headers":  fmt.Sprintf("%v", rm.Headers()),
		"body":     string(rm.body),
	}
}

// Compares two RequestMessages
func (rm *RequestMessage) Equals(other *RequestMessage) bool {
	return bytes.Equal(rm.ToBytes(), other.ToBytes())
}

// Get Method
func (rm *RequestMessage) Method() string {
	return rm.method
}

// Get Path
func (rm *RequestMessage) Path() string {
	return rm.path
}

// Get Protocol
func (rm *RequestMessage) Protocol() string {
	return rm.protocol
}

// Get Headers
func (rm *RequestMessage) Headers() map[string]string {
	return rm.headers
}

// Headers sorted by name
func (rm *RequestMessage) HeadersToSlice() []string {
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
func (rm *RequestMessage) PrintHeaders() {
	for _, value := range rm.HeadersToSlice() {
		fmt.Println(value)
	}
}

// Get Body
func (rm *RequestMessage) Body() []byte {
	return rm.body
}

// Print
func (rm *RequestMessage) Print() {
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

// Copy RequestMessage and return new RequestMessage
func (rm *RequestMessage) Copy() *RequestMessage {
	return &RequestMessage{
		method:   rm.method,
		path:     rm.path,
		protocol: rm.protocol,
		headers:  rm.headers,
		body:     rm.body,
	}
}

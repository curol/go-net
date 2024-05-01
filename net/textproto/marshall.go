/*
The general term for parsing and serializing is "Data Marshalling" or just "Marshalling".

- **Parsing** is the process of converting data in a specific format (like JSON, XML, etc.) into a format that your program can use, such as a data structure or object. This is also known as "unmarshalling" or "deserialization".

- **Serializing** is the process of converting a data structure or object in your program into a format that can be stored or transmitted, such as a JSON or XML string. This is also known as "marshalling".

Together, these processes are used to convert data between formats, often for the purposes of storage, transmission over a network, or communication between different parts of a program.
*/

package textproto

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"strconv"
	"strings"
)

type parsedTextMessage struct {
	// Status line
	status string
	// Headers
	headers MIMEHeader
	// Body
	body *bufio.Reader
	// Size
	headLen int   // Size of headers
	cl      int64 // Size of body
	ct      string
	// Error
	err error
}

// Serialize converts the parsed data into a byte slice
func (pd *parsedTextMessage) WriteTo(w *bufio.Writer) (n int64, err error) {
	var nn int
	nn, err = fmt.Fprintf(w, "%s\r\n", pd.status)
	n += int64(nn)
	if err != nil {
		return n, err
	}
	nn, err = fmt.Fprintf(w, "%s\r\n", serializeHeaders(pd.headers))
	n += int64(nn)
	if err != nil {
		return n, err
	}
	bcount, err := io.CopyN(w, pd.body, pd.cl)
	n += int64(bcount)
	if err != nil {
		return n, err
	}
	return n, w.Flush()
}

func newParsedTextMessage(r *bufio.Reader) (*parsedTextMessage, error) {
	if r == nil {
		return nil, errors.New("r is nil")
	}
	// Arrange
	pd := &parsedTextMessage{}
	var e error
	delm := byte('\n')

	// 1. Read and parse status line
	status, e := r.ReadString(delm) // 1
	if e != nil {
		pd.err = e
		return pd, e
	}
	status = strings.TrimSpace(status)
	pd.status = status
	pd.headLen += len(status) + 2

	// 2. Read and parse headers
	h, n, e := parseHeaders(r)
	if e != nil {
		pd.err = e
		return pd, e
	}
	pd.headers = h
	pd.headLen += n + 2

	// 3. Set content length and type
	pd.cl = parseContentLength(pd.headers)
	pd.ct, _ = parseContentType(pd.headers)

	// 4. Set body
	pd.body = r

	return pd, nil
}

// parseReaderHeaders parses headers from a bufio.Reader into a map of header names to slices of header values.
//
// Example of parsing raw headers:
// ```
//
//	bytesReader := bytes.NewReader(rawHeaders)
//	reader := bufio.NewReader(bytesReader)
//	parseHeaders(reader)
//
// ```
//
// Example parsing headers:
// ```
// 2. Read and parse headers
// header := NewHeader()
// for { // read each new line until a blank line ("\r\n") is reached.
//
//		line, err := r.ReadString('\n') // read line
//		if err != nil && err != io.EOF {
//			return nil, err
//		}
//		if line == "\r\n" || err == io.EOF { // headers are terminated by a blank line "\r\n"
//			break
//		}
//		parts := strings.SplitN(line, ":", 2) // parse line by splitting line into key and value
//		if len(parts) < 2 {
//			return nil, fmt.Errorf("invalid header line")
//		}
//		// remove leading and trailing whitespace from key and value
//		k := strings.TrimSpace(parts[0])
//		v := strings.TrimSpace(parts[1])
//		header.Set(k, v)
//	}
//
// ```
func parseHeaders(r *bufio.Reader) (map[string][]string, int, error) {
	tp := textproto.NewReader(r)
	h, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, -1, err
	}
	n := getHeaderSize(h)
	return h, n, nil
}

// parseStatusLine parses a raw status line into a method, path, and protocol.
func parseStatusLine(rawStatusLine []byte) (string, string, string, error) {
	s := strings.TrimSpace(string(rawStatusLine))
	ss := strings.Split(s, " ")
	if len(ss) != 3 {
		return "", "", "", errors.New("invalid status line format: " + s)
	}
	method := strings.TrimSpace(ss[0])
	path := strings.TrimSpace(ss[1])
	proto := strings.TrimSpace(ss[2])
	return method, path, proto, nil
}

// parseContentLength parses the Content-Length header from a map of headers.
// returns -1 if the header not found
// returns 0 if the header value is not a valid integer
// returns the content length if the header is found and the value is a valid integer
func parseContentLength(headers map[string][]string) int64 {
	// get content length
	cl, ok := headers["Content-Length"]
	if !ok {
		return -1
	}
	n, e := strconv.ParseInt(cl[0], 10, 64)
	if e != nil {
		return 0
	}
	return n
}

func parseContentType(headers map[string][]string) (string, error) {
	// get content type
	ct, ok := headers["Content-Type"]
	if !ok {
		return "", errors.New("missing content type")
	}
	return ct[0], nil
}

// serializeHeaders serializes a map of headers into a byte slice.
func serializeHeaders(headers map[string][]string) []byte {
	var buffer bytes.Buffer

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

func getHeaderSize(headers map[string][]string) int {
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

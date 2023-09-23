package message

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

//**********************************************************************************************************************
// MessageReader
//
// Read reads the source of data into buffer. Read(buffer []byte) (n int, err error)
// Write writes buffer to the source. Write(buffer []byte) (n int, err error)
//**********************************************************************************************************************

// MessageReader structures the raw request message.
//
// It implements parsing and buffering and parsing from the source reader.
// Payload of message is written
type MessageReader struct {
	// Readers
	r io.Reader // reader provided by the client
	// Parsed data
	method        string            // Parsed method
	path          string            // Parsed path
	protocol      string            // Parsed protocol
	headersMap    map[string]string // Contains parsed headers
	contentType   string            // Parsed header Content-Type
	contentLength int               // Parsed header Content-Length
	len           int               // Size of message (request line + headers + body)
	// Buffers
	reqLineBuf []byte // Buffer for request line
	headersBuf []byte // Buffer for headers
	bodyBuf    []byte
}

func NewMessageReader(r io.Reader) *MessageReader {
	// wb := bufio.NewWriter(body)
	src := io.NopCloser(r)
	m, err := ParseReaderToMessage(src)
	if err != nil {
		panic(err)
	}
	return m
}

func (p *MessageReader) String() string {
	// lines := []string{
	// 	fmt.Sprintf("MessageReader"),
	// 	fmt.Sprintf("\t- Method: %s", p.method),
	// 	fmt.Sprintf("\t- Path: %s", p.path),
	// 	fmt.Sprintf("\t- Protocol: %s", p.protocol),
	// 	fmt.Sprintf("\t- RequestLine: %d", p.reqLineBuf),
	// 	fmt.Sprintf("\t- Headers: %d", p.headers),
	// 	fmt.Sprintf("\t- HeadersMap: %s", p.headersMap),
	// 	fmt.Sprintf("\t- Body: %p", p.body),
	// 	fmt.Sprintf("\t- ContentLength: %d", p.contentLength),
	// 	fmt.Sprintf("\t- ContentType: %s", p.contentType),
	// }
	b := p.ToBytes()
	return string(b)
}

func (p *MessageReader) ToBytes() []byte {
	b := bytes.NewBuffer(nil)
	_, err := p.WriteTo(b)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func (p *MessageReader) ToFile(path string) (int64, error) {
	// File stream
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	return p.WriteTo(f)
}

func (p *MessageReader) WriteTo(w io.Writer) (int64, error) {
	// Write request line to w
	n, err := w.Write(p.reqLineBuf)
	if err != nil {
		return int64(n), err
	}
	// Write headers to w
	n2, err := w.Write(p.headersBuf)
	if err != nil {
		return int64(n + n2), err
	}
	// Write body to w
	n3, err := w.Write(p.bodyBuf)
	return int64(n + n2 + n3), err
}

func (p *MessageReader) copyReaderToWriter(des io.Writer, src io.Reader, n int64) (int64, error) {
	return io.CopyN(des, src, n)
}

func (p *MessageReader) Equals(other *MessageReader) bool {
	// Check size
	if p.Len() != other.Len() {
		return false
	}
	// Check if other map contains same key-value pairs.
	//
	// Note order doesnt matter, so we can't just compare maps.
	for k, v := range p.headersMap {
		if v != other.headersMap[k] {
			return false
		}
	}
	return true
}

//######################################################################################################################
// Getters
//######################################################################################################################

func (p *MessageReader) Len() int { return p.len }

func (p *MessageReader) RequestLine() []byte { return p.reqLineBuf }

func (p *MessageReader) Headers() []byte { return p.headersBuf }

func (p *MessageReader) HeadersMap() map[string]string { return p.headersMap }

func (p *MessageReader) Body() []byte { return p.bodyBuf }

func (p *MessageReader) Method() string { return p.method }

func (p *MessageReader) Path() string { return p.path }

func (p *MessageReader) Protocol() string { return p.protocol }

func (p *MessageReader) ContentType() string { return p.contentType }

func (p *MessageReader) ContentLength() int { return p.contentLength }

//######################################################################################################################
// Helpers
//######################################################################################################################

func ParseReaderToMessage(r io.Reader) (*MessageReader, error) {
	reader := bufio.NewReader(r) // wrap src reader in bufio.Reader

	pm := new(MessageReader)
	pm.r = r // set src reader

	// 1.) Request line
	rl, err := reader.ReadString('\n') // parse first line from reader as the request line
	if err != nil && err != io.EOF {
		return nil, err
	}
	parts := strings.SplitN(rl, " ", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line")
	}
	pm.reqLineBuf = []byte(rl)
	pm.method, pm.path, pm.protocol = parts[0], parts[1], parts[2]

	// 2.) Headers
	headersBytes := bytes.NewBuffer(nil)
	m := make(map[string]string)
	for {
		line, err := reader.ReadString('\n') // read line
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		headersBytes.Write([]byte(line)) // write line to buffer
		if line == "\r\n" {              // headers are terminated by a blank line "\r\n"
			break
		}
		parts := strings.SplitN(line, ":", 2) // split line into key and value
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid header line")
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		m[key] = value
	}
	pm.headersMap = m
	pm.headersBuf = headersBytes.Bytes()
	cl, ok := pm.headersMap["Content-Length"]
	if !ok {
		cl = "0"
	}
	length, err := strconv.Atoi(cl) // convert to int
	if err != nil {
		length = 0
	}
	pm.contentLength = length // set Content-Length
	ct, ok := pm.headersMap["Content-Type"]
	if !ok {
		ct = ""
	}
	pm.contentType = ct // set Content-Type

	pm.len = len(pm.reqLineBuf) + len(pm.headersMap) + pm.contentLength // set size

	// 3.) Body
	// One more read call to get body contents
	//
	// TODO: Check if size is too big for MaxReadSize and MaxWriteSize
	// Write body to w
	// if p.contentLength > MaxReadSize {
	// 	return int64(n + n2), fmt.Errorf("content length too big")
	// }
	buf := bytes.NewBuffer(make([]byte, 0, pm.contentLength))
	_, err = copyN(buf, reader, int64(pm.contentLength)) // copy reader to writer
	if err != nil {
		panic(err)
	}

	return pm, nil
}

func ParseBytesToMessage(data []byte) (*MessageReader, error) {
	r := bufio.NewReader(bytes.NewBuffer(data))
	return ParseReaderToMessage(r)
}

// CopyN copies n bytes (or until an error) from src to dst.
func copyN(dst io.Writer, src io.Reader, size int64) (int64, error) {
	// CopyN copies n bytes (or until an error) from src to dst
	// Read from b.body and write to w of length b.len
	return io.CopyN(dst, src, size)
}

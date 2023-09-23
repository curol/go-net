package message

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"util/hashmap"
)

// Header is the header of a message, which contains the headers.
// It wraps the headers and raw data of a message.
//
// Headers can have values that contain commas.
// According to the HTTP/1.1 specification, multiple message-header fields with the same field-name may be present in a message if and only if the entire field-value for that header field is defined as a comma-separated list.
// I.g., Accept: text/plain, text/html
//
// For Brevity, we will not be supporting headers with multiple values.
type Header struct {
	hashmap.HashMap
	buf *bytes.Buffer
}

// NewHeader returns a new Header.
func NewHeader(reader *bufio.Reader) (*Header, error) {
	header := new(Header)
	header.HashMap = hashmap.New()

	// Parse reader for headers
	err := header.parse(reader)

	return header, err
}

// parse parses the reader for headers.
func (h *Header) parse(reader *bufio.Reader) error {
	// TODO: Store each line read?
	// lines := make([]byte, 0)
	h.buf = bytes.NewBuffer(nil)

	for {
		line, err := reader.ReadString('\n')

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		// Write line to buffer
		h.buf.Write([]byte(line))

		// Headers are terminated by a blank line "\r\n"
		if line == "\r\n" {
			// Break from loop because headers are done
			break
		}

		// Split the header into name and value
		// kv := strings.SplitN(line, ": ", 2)
		// if len(kv) != 2 {
		// 	return fmt.Errorf("malformed header")
		// }
		k, v, err := h.parseLine(line)
		if err != nil {
			return err
		}

		// Add header
		h.Set(k, v)
	}

	return nil
}

// parseLine parses a header line into a key and value.
func (h *Header) parseLine(line string) (string, string, error) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return "", "", errors.New("invalid header line")
	}
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	return key, value, nil
}

// **********************************************************************************************************************
// Getters
// **********************************************************************************************************************

// Size returns the size of the underyling data buffer.
func (h *Header) Size() int {
	if h.buf == nil {
		return 0
	}
	return h.buf.Len()
}

// ContentLength returns the Content-Length header. If the header is not present, it returns an error.
func (h *Header) ContentLength() (int, error) {
	contentLength, ok := h.Get("Content-Length")

	// Check if Content-Length header exists
	if !ok {
		return -1, fmt.Errorf("missing header Content-Length")
	}

	// Convert to int
	length, err := strconv.Atoi(contentLength)
	if err != nil {
		return -1, fmt.Errorf("invalid Content-Length value")
	}
	return length, nil
}

// ContentType returns the Content-Type header. If the header is not present, it returns an error.
//
// Note that the Content-Type header specifies the MIME type of the request body.
func (h *Header) ContentType() (string, error) {
	ct, ok := h.Get("Content-Type")
	if !ok {
		return "", fmt.Errorf("missing header Content-Type")
	}
	return ct, nil
}

// Header returns the header as a map of strings to strings.
func (h *Header) Header() map[string]string {
	return h.HashMap
}

// ToBytes returns the header as a byte slice.
func (h *Header) ToBytes() []byte {
	// return h.buf.Bytes()
	strs := h.HashMap.ToStrings()
	joinedStrs := strings.Join(strs, "\r\n")
	s := joinedStrs + "\r\n"
	fmt.Println(s)
	return []byte(s)
}

// ToString returns the header as a string.
func (h *Header) ToString() string {
	return string(h.ToBytes())
}

func (h *Header) Clone() *Header {
	return &Header{
		HashMap: h.HashMap.Clone(),
		buf:     bytes.NewBuffer(h.ToBytes()),
	}
}

// ******************************************************
// Mutators
// ******************************************************

// ReadFrom reads a sequence of headers from r until io.EOF and adds them to the Header.
// It returns the number of bytes read. If an error occurs before io.EOF, it returns the number of bytes read so far and the error.
// A successful ReadFrom returns err == nil, not err == io.EOF. Because ReadFrom is defined to read from src until EOF, it does not treat an EOF from Read as an error to be reported.
// If the header line is invalid, it returns an error
func (h *Header) ReadFrom(r io.Reader) (int64, error) {
	buf := bytes.NewBuffer(h.ToBytes())
	reader := bufio.NewReader(r)
	n := int64(0)

	for {
		// Read
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		n += int64(len(line))
		if err != nil {
			return int64(buf.Len()), err
		}
		// Parse line
		k, v, err := h.parseLine(string(line))
		if err != nil {
			return int64(buf.Len()), err
		}
		// Write to buffer
		_, err = buf.Write(line)
		if err != nil {
			return int64(buf.Len()), err
		}
		// Add header
		h.Set(k, v)
	}
	h.buf = buf
	return n, nil
}

// WriteTo writes a sequence of headers to w in the HTTP/1.1 header format.
func (h *Header) WriteMapTo(w io.Writer) (int64, error) {
	var written int64
	writer := bufio.NewWriter(w)
	headers := h.HashMap.ToStrings()
	// Add Empty line to headers to mark end of header
	headers = append(headers, "")

	for _, line := range headers {
		n, err := fmt.Fprintf(writer, "%s\r\n", line)
		written += int64(n)
		if err != nil {
			return written, err
		}
	}

	return written, nil
}

func (h *Header) WriteTo(w io.Writer) (int64, error) {
	b := h.ToBytes()
	writer := bufio.NewWriter(w)
	n, err := writer.Write(b)
	return int64(n), err
}

// Equals returns true if the headers are equal. An equality check is done by comparing the size of the headers and the values of the headers.
//
// Note that the order of the headers does not matter.
func (h *Header) Equals(other *Header) bool {
	if h.Size() != other.Size() {
		return false
	}

	// for k, v := range h.HashMap {
	// 	if v != otherHeader.HashMap[k] {
	// 		return false
	// 	}
	// }
	return h.HashMap.Equals(other.HashMap)
}

func (h *Header) EqualsBytesTo(other []byte) bool {
	reader := bufio.NewReader(bytes.NewReader(other))
	otherHeader, err := NewHeader(reader)
	if err != nil {
		return false
	}
	if len(h.HashMap) != len(otherHeader.HashMap) {
		return false
	}
	// for k, v := range h.HashMap {
	// 	if v != otherHeader.HashMap[k] {
	// 		return false
	// 	}
	// }
	return h.HashMap.Equals(otherHeader.HashMap)
}

// TODO: Implement Validate
// Validate headers
// Validate checks if the Header is valid.
// func (h Header) Validate() error {
// 	// Define the list of acceptable keys if necessary.
// 	// acceptableKeys := []string{"Key1", "Key2", "Key3"}

// 	for key, val := range h {
// 		// Check if the key is empty.
// 		if key == "" {
// 			return errors.New("key cannot be empty")
// 		}

// 		if len(val) == 0 || val == "" {
// 			return errors.New("value cannot be empty")
// 		}

// 		// Check if the key is in the list of acceptable keys.
// 		// if !contains(acceptableKeys, key) {
// 		//     return fmt.Errorf("invalid key: %s", key)
// 		// }

// 		// Check if the values are empty.
// 		// for _, value := range values {
// 		// 	if value == "" {
// 		// 		return fmt.Errorf("value for key %s cannot be empty", key)
// 		// 	}
// 		// }
// 	}

// 	return nil
// }

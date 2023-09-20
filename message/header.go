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
	"util/stream"
)

// Header is the header of a message which contains the headers.
// It wraps the headers of a message as a map of strings to strings.
//
// Headers can have values that contain commas.
// According to the HTTP/1.1 specification, multiple message-header fields with the same field-name may be present in a message if and only if the entire field-value for that header field is defined as a comma-separated list.
// I.g., Accept: text/plain, text/html
//
// For Brevity, we will not be supporting headers with multiple values.
type Header struct {
	hashmap.HashMap
}

// NewHeader returns a new Header.
func NewHeader(reader *bufio.Reader) (*Header, int, error) {
	header := new(Header)
	header.HashMap = hashmap.New()

	// Parse
	n, err := header.parse(reader)
	if err != nil {
		return nil, n, err
	}

	// Return values
	return header, n, nil
}

func (h *Header) parse(reader *bufio.Reader) (int, error) {
	// TODO: Store each line read?
	// lines := make([]byte, 0)

	// Bytes read
	n := 0

	for {
		line, err := reader.ReadString('\n')

		n += len(line)

		if err != nil {
			return n, err
		}

		// lines = append(lines, []byte(line)...)

		// Headers are terminated by a blank line "\r\n"
		if line == "\r\n" {
			// Break from loop because headers are done
			break
		}

		// Split the header into name and value
		kv := strings.SplitN(line, ": ", 2)
		if len(kv) != 2 {
			return n, fmt.Errorf("malformed header")
		}

		// Add headers
		h.Set(kv[0], kv[1])
	}

	return n, nil
}

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

// Get the Content-Type header.
//
// Note that the Content-Type header specifies the MIME type of the request body.
func (h *Header) ContentType() (string, error) {
	ct, ok := h.Get("Content-Type")
	if !ok {
		return "", fmt.Errorf("missing header Content-Type")
	}
	return ct, nil
}

func (h *Header) Header() map[string]string {
	return h.HashMap
}

// ******************************************************
// Read
// ******************************************************

// ReadFrom reads a sequence of headers from r until io.EOF and adds them to the Header.
// It returns the number of bytes read. If an error occurs before io.EOF, it returns the number of bytes read so far and the error.
// A successful ReadFrom returns err == nil, not err == io.EOF. Because ReadFrom is defined to read from src until EOF, it does not treat an EOF from Read as an error to be reported.
// If the header line is invalid, it returns an error
func (h *Header) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	for {
		line, err := stream.ReadLine(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return read, err
		}
		read += int64(len(line))
		parts := bytes.SplitN(line, []byte{':'}, 2)
		if len(parts) < 2 {
			return read, errors.New("invalid header line")
		}
		key := string(bytes.TrimSpace(parts[0]))
		value := string(bytes.TrimSpace(parts[1]))
		h.Set(key, value)
	}
	return read, nil
}

// ******************************************************
// Write
// ******************************************************

// Write writes a sequence of headers to w in the HTTP/1.1 header format.
func (h *Header) Write(b []byte) (int, error) {
	n, err := h.WriteTo(bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

// WriteTo writes a sequence of headers to w in the HTTP/1.1 header format.
func (h *Header) WriteTo(w io.Writer) (int64, error) {
	var written int64
	headers := h.ToStrings()
	// Add Empty line to headers to mark end of header
	headers = append(headers, "")

	for _, line := range headers {
		n, err := fmt.Fprintf(w, "%s\r\n", line)
		written += int64(n)
		if err != nil {
			return written, err
		}
	}

	return written, nil
}

// ToBytes returns the headers as a byte slice.
func (h *Header) ToBytes() ([]byte, error) {
	return toBytes(h)
}

// String returns the text of the Map formatted in the same way as in the request.
func (h *Header) ToString() string {
	return toString(h)
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

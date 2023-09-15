package message

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Returns RequestMessage from parsing reader
func parse(reader *bufio.Reader) (*RequestMessage, error) {
	// Read and parse request line
	method, path, protocol, err := parseRequestLine(reader)
	if err != nil {
		return nil, err
	}
	// Read and parse headers
	headers, err := parseHeaders(reader)
	if err != nil {
		return nil, err
	}
	// Parse body
	body, err := parseBody(reader, headers)
	if err != nil {
		return &RequestMessage{
			method:   method,
			path:     path,
			protocol: protocol,
			headers:  headers,
			body:     nil, // No body
		}, nil
	}
	// Return RequestMessage
	return &RequestMessage{
		method:   method,
		path:     path,
		protocol: protocol,
		headers:  headers,
		body:     body,
	}, nil
}

// Request line reads the first line of the connection and returns the method, path, and protocol.
// Format of first line (Request line): <method> <path> <protocol>
func parseRequestLine(reader *bufio.Reader) (string, string, string, error) {
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return "", "", "", err
	}
	requestLineComponents := strings.Split(strings.TrimSpace(requestLine), " ")
	if len(requestLineComponents) != 3 {
		return "", "", "", fmt.Errorf("Malformed request line. Expected format: <method> <path> <protocol>")
	}
	return requestLineComponents[0], requestLineComponents[1], requestLineComponents[2], nil
}

// Headers
// Headers are terminated by a blank line "\r\n"
// Read Headers for metadata about the request.
func parseHeaders(reader *bufio.Reader) (map[string]string, error) {
	// Headers
	// Headers are terminated by a blank line "\r\n"
	// Read Headers for metadata about the request.
	lines := make([]byte, 0)
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		lines = append(lines, []byte(line)...)
		// Headers are terminated by a blank line "\r\n"
		if line == "\r\n" {
			// Break from loop because headers are done
			break
		}
		// Split the header into name and value
		headerComponents := strings.SplitN(line, ": ", 2)
		if len(headerComponents) != 2 {
			return nil, fmt.Errorf("malformed header")
		}
		// Set headers
		key := strings.TrimSpace(headerComponents[0])
		val := strings.TrimSpace(headerComponents[1])
		headers[key] = val
	}
	return headers, nil
}

// Body is the payload of the message.
//
// Why carefully read the body?
// Because the size of the data might be too big for server to read all at once.
// For streams, we don't want to read the entire body at once into memory.
// If its too big, we might run out of memory.
// Instead, implement strategy for different sizes of the data and use chunks or buffers to read from the connection.
func parseBody(reader *bufio.Reader, headers map[string]string) ([]byte, error) {
	contentLength, ok := headers["Content-Length"]
	if !ok {
		return nil, fmt.Errorf("missing header Content-Length")
	}

	length, err := strconv.Atoi(contentLength)
	if err != nil {
		return nil, fmt.Errorf("invalid Content-Length value")
	}

	body := make([]byte, length)
	_, err = io.ReadFull(reader, body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

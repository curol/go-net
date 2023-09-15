package reader

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type HTTPRequest struct {
	Method      string
	URI         string
	HTTPVersion string
	Headers     map[string]string
}

func parseHTTPRequest(conn net.Conn) (*HTTPRequest, error) {
	reader := bufio.NewReader(conn)

	// Read the request line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	// Split the request line into components
	requestLineComponents := strings.Split(strings.TrimSpace(requestLine), " ")
	if len(requestLineComponents) != 3 {
		return nil, fmt.Errorf("malformed request line")
	}

	// Read the headers
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		// Headers are terminated by a blank line
		if line == "\r\n" {
			break
		}

		// Split the header into name and value
		headerComponents := strings.SplitN(line, ": ", 2)
		if len(headerComponents) != 2 {
			return nil, fmt.Errorf("malformed header")
		}

		headers[strings.TrimSpace(headerComponents[0])] = strings.TrimSpace(headerComponents[1])
	}

	return &HTTPRequest{
		Method:      requestLineComponents[0],
		URI:         requestLineComponents[1],
		HTTPVersion: requestLineComponents[2],
		Headers:     headers,
	}, nil
}

func ReadHTTPRequestFromFile(filename string) (*http.Request, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	req, err := http.ReadRequest(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read request: %v", err)
	}

	return req, nil
}

type HTTPMessage struct {
	Headers map[string]string
	Body    string
}

func parseHTTPMessageFromConnection(conn net.Conn) (*HTTPMessage, error) {
	reader := bufio.NewReader(conn)

	// Read headers
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		// HTTP headers are terminated by a blank line
		if line == "\r\n" {
			break
		}

		// Parse headers
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	// Read body
	body, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	return &HTTPMessage{
		Headers: headers,
		Body:    body,
	}, nil
}

func printAllUntilConnectionClosed(reader io.Reader) {
	// Read until connection is closed
	fmt.Println("Reading until connection is closed or err..")
	body, err := io.ReadAll(reader)
	if err != nil {
		fmt.Println("Error io.ReadAll:", err)
		return
	}
	// Print body
	fmt.Println("Body:", body)
}

func printAllHeaders(reader *bufio.Reader) {
	// Print all header lines
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		// HTTP headers are terminated by a blank line
		if line == "\r\n" {
			break
		}

		// Parse headers
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	// Print headers
	for k, v := range headers {
		fmt.Printf("%s: %s\n", k, v)
	}
}

func printRequestLine(reader *bufio.Reader) {
	// Request line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}
	rl := strings.TrimSpace(requestLine)
	fmt.Println("Request Line: ", rl, []byte(rl))
}

func ParseRequestLineFromMessage(request string) (string, string, string) {
	// Parse request line
	requestLine := strings.Split(request, "\r\n")[0]
	method := strings.Split(requestLine, " ")[0]
	path := strings.Split(requestLine, " ")[1]
	protocol := strings.Split(requestLine, " ")[2]

	return method, path, protocol
}

func PrintHttpFromConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Request line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}
	// Print the request line
	rl := strings.TrimSpace(requestLine)
	fmt.Println(rl, []byte(rl))

	// Print all header lines
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		// HTTP headers are terminated by a blank line
		if line == "\r\n" {
			break
		}

		fmt.Println(strings.TrimSpace(line))
	}
}

func PrintExtractedHTTPMessage(conn net.Conn) {
	reader := bufio.NewReader(conn)

	// Request line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print the request line
	rl := strings.TrimSpace(requestLine)
	fmt.Println(rl, []byte(rl))

	// Print all header lines
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		// HTTP headers are terminated by a blank line
		if line == "\r\n" {
			break
		}

		// Parse headers
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	// Print headers
	for k, v := range headers {
		fmt.Printf("%s: %s\n", k, v)
	}

	// Read body
	body, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print body
	fmt.Println(body)
}

func ReadHTTPBody(conn net.Conn) {
	reader := bufio.NewReader(conn)

	// Request line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}
	rl := strings.TrimSpace(requestLine)
	fmt.Println("Request Line: ", rl, []byte(rl))

	// Read headers
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		// HTTP headers are terminated by a blank line
		if line == "\r\n" {
			break
		}
		// Parse headers
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	fmt.Println("Headers:", headers)

	// Read body
	if headers["Transfer-Encoding"] == "chunked" {
		// Read until final chunk
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				return
			}

			// Parse chunk length
			length, err := strconv.ParseInt(strings.TrimSpace(line), 16, 64)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Final chunk
			if length == 0 {
				break
			}

			// Read chunk
			chunk := make([]byte, length)
			_, err = io.ReadFull(reader, chunk)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Print chunk
			fmt.Println(string(chunk))
		}
	} else if contentLength, ok := headers["Content-Length"]; ok {
		// Read specified number of bytes
		length, err := strconv.Atoi(contentLength)
		if err != nil {
			fmt.Println(err)
			return
		}

		body := make([]byte, length)
		_, err = io.ReadFull(reader, body)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Print body
		fmt.Println(string(body))
	} else {
		// Read until connection is closed
		fmt.Println("Reading until connection is closed or err..")
		body, err := io.ReadAll(reader)
		if err != nil {
			fmt.Println("Error io.ReadAll:", err)
			return
		}
		// Print body
		fmt.Println("Body:", body)
	}
}

/*
The general term for parsing and serializing is "Data Marshalling" or just "Marshalling".

- **Parsing** is the process of converting data in a specific format (like JSON, XML, etc.) into a format that your program can use, such as a data structure or object. This is also known as "unmarshalling" or "deserialization".

- **Serializing** is the process of converting a data structure or object in your program into a format that can be stored or transmitted, such as a JSON or XML string. This is also known as "marshalling".

Together, these processes are used to convert data between formats, often for the purposes of storage, transmission over a network, or communication between different parts of a program.
*/

package textproto

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/curol/network/net/header"
)

type parsedTextMessage struct {
	// Status line
	status string
	// Headers
	headers map[string][]string
	// Body
	body *bufio.Reader
	// Size
	cl  int64 // Size of body
	ct  string
	n   int // # of bytes read
	err error
}

func (pd *parsedTextMessage) writeTo(w *bufio.Writer) (n int64, err error) {
	var nn int
	nn, err = fmt.Fprintf(w, "%s\r\n", pd.status)
	n += int64(nn)
	if err != nil {
		return n, err
	}
	nn, err = fmt.Fprintf(w, "%s\r\n", header.SerializeHeaders(pd.headers))
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
	pd.n += len(status) + 2

	// 2. Read and parse headers
	h, nn, e := header.ParseHeaders(r)
	if e != nil {
		pd.err = e
		return pd, e
	}
	pd.headers = h
	pd.n += nn + 2

	// 3. Set content length and type
	pd.cl = parseContentLength(pd.headers)
	pd.ct, _ = parseContentType(pd.headers)

	// 4. Set body
	lr := io.LimitReader(r, pd.cl)
	pd.body = bufio.NewReader(lr)

	return pd, nil
}

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

func parseContentLength(headers map[string][]string) int64 {
	if headers == nil {
		return 0
	}
	// Check if 'Content-Length' is exists
	cl, ok := headers["Content-Length"]
	if !ok {
		return 0
	}
	// Check if there is only one content length
	if len(cl) != 1 {
		return 0
	}
	l := TrimString(cl[0])
	n, e := strconv.ParseInt(l, 10, 64)
	if e != nil {
		return 0
	}
	return n
}

// func getContentLength(h header.Header) int64 {
// 	if h == nil {
// 		return 0
// 	}
// 	if cl, ok := h["Content-Length"]; ok {
// 		if len(cl) == 1 {
// 			l := TrimString(cl[0])
// 			if value, err := strconv.ParseInt(l, 10, 64); err == nil {
// 				return value
// 			}
// 		}
// 	}
// 	return 0
// }

func parseContentType(headers map[string][]string) (string, error) {
	// get content type
	ct, ok := headers["Content-Type"]
	if !ok {
		return "", errors.New("missing content type")
	}
	return ct[0], nil
}

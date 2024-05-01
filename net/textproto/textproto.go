// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package textproto implements generic support for text-based request/response
// protocols in the style of HTTP, NNTP, and SMTP.
//
// The package provides:
//
// [Error], which represents a numeric error response from
// a server.
//
// [Pipeline], to manage pipelined requests and responses
// in a client.
//
// [Reader], to read numeric response code lines,
// key: value headers, lines wrapped with leading spaces
// on continuation lines, and whole text blocks ending
// with a dot on a line by itself.
//
// [Writer], to write dot-encoded text blocks.
//
// [Conn], a convenient packaging of [Reader], [Writer], and [Pipeline] for use
// with a single network connection.
package textproto

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// A Conn represents a textual network protocol connection.
// It consists of a [Reader] and [Writer] to manage I/O
// and a [Pipeline] to sequence concurrent requests on the connection.
// These embedded types carry methods with them;
// see the documentation of those types for details.
type Conn struct {
	R    *bufio.Reader
	W    *bufio.Writer
	conn io.ReadWriteCloser
}

// A MIMEHeader represents a MIME-style header mapping.
type MIMEHeader map[string][]string

// A TextMessage represents a text-based message.
type TextMessage struct {
	// headers
	Status  string     // first line
	Headers MIMEHeader // 2nd line through blank line
	// body
	Body       *bufio.Reader // rest of lines
	ContentLen int64         // Content-Length header for size of content
	ContentTyp string        // Content-Tupe header for type of content
	// meta
	isParsed   bool
	isReadBody bool // true if body read
	bytesRead  int64
}

// Buffer returns the TextMessage as a *bytes.buffer
func (tp *TextMessage) Buffer() *bytes.Buffer {
	b := bytes.NewBuffer(nil)
	tp.WriteTo(b)
	return b
}

// Bytes returns the TextMessage as a slice of bytes
func (tp *TextMessage) Bytes() []byte {
	buf := tp.Buffer()
	return buf.Bytes()
}

// StdOut writes the TextMessage to stdout
func (tp *TextMessage) StdOut() {
	w := bufio.NewWriter(os.Stdout)
	tp.WriteTo(w)
}

// Strings returns the TextMessage as a slice of strings without the \r\n
func (tp *TextMessage) Strings() []string {
	return strings.Split(string(tp.Bytes()), "\r\n")
}

// Head return the head of the TextMessage
func (tm *TextMessage) Head() {
	buf := bytes.NewBuffer(nil)
	w := bufio.NewWriter(buf)
	tm.serialize(w, false)
}

// File saves content to disk
func (tm *TextMessage) File(path string) (int64, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	var n int64
	for {
		nn, err := tm.Body.WriteTo(f)
		n += nn
		if err != nil {
			return n, err
		}
		if nn == 0 {
			break
		}
	}
	return n, nil
}

// Read reads TextMessage.body to p
func (tm *TextMessage) Read(p []byte) (int, error) {
	return tm.Body.Read(p)
}

// Size returns the size of the TextMessage (headers + body)
func (tp *TextMessage) Size() int64 {
	return int64(tp.bytesRead) + tp.ContentLen
}

func (tm *TextMessage) ContentLength() int64 {
	return tm.ContentLen
}

func (tm *TextMessage) ContentType() string {
	return tm.ContentTyp
}

// WriteTo serializes TextMessage to w
func (tm *TextMessage) WriteTo(w io.Writer) (int64, error) {
	var bw *bufio.Writer

	// 1. Type switch w
	switch v := w.(type) {
	case *bufio.Writer:
		bw = v
	case *bytes.Buffer:
		bw = bufio.NewWriter(v)
	default:
		return 0, fmt.Errorf("textproto: writer type '%T' not supported", v)
	}

	n, err := tm.serialize(bw, true)
	if err != nil {
		return n, err
	}
	return n, nil
}

// serialize writes the TextMessage to w and returns the number of bytes written
// It writes the status line, headers, and body to w.
//
// Example:
// tp := &TextMessage{Status: "200 OK", Headers: MIMEHeader{"Content-Type": []string{"text/plain"}}, Body: bufio.NewReader(strings.NewReader("Hello, World!"))}
// bw := bufio.NewWriter(os.Stdout)
// n, err := tp.serialize(bw)
func (tp *TextMessage) serialize(bw *bufio.Writer, isSerializeBody bool) (int64, error) {
	// var bw *bufio.Writer
	var n int64
	var err error
	headers := tp.Headers
	dlm := "\r\n"
	// 1. Clean up
	defer func() {
		err = bw.Flush()
		if err != nil {
			fmt.Println("textproto: Error flushing serialization to w - ", err)
		}
	}()
	// 2. Status line
	sn, _ := fmt.Fprintf(bw, "%s%s", tp.Status, dlm) // Write status line
	n = int64(sn)
	// 3. Headers
	// h := serializeHeaders(tp.Headers)
	// n2, _ = fmt.Fprint(bw, string(h)) // Write headers
	for name, values := range headers {
		for _, value := range values {
			hn, _ := fmt.Fprintf(bw, "%s: %s%s", name, value, dlm)
			n += int64(hn)
		}
	}
	// 4. End of headers
	ehn, _ := fmt.Fprintf(bw, dlm) // Write blank line between headers and body
	n += int64(ehn)
	// 5. Body
	if !isSerializeBody {
		return n, err
	}
	// 5.1. Validate
	cl := tp.ContentLen
	if cl < 0 {
		return 0, fmt.Errorf("textproto: Content-Length less than 0 %d", cl)
	}
	if cl == 0 {
		return 0, nil
	}
	if tp.isReadBody {
		return 0, fmt.Errorf("textproto: Body already read")
	}
	// 5.2. Copy body to dst
	src := tp.Body
	dst := bw
	bn, err := io.CopyN(dst, src, cl)
	tp.isReadBody = true
	n += int64(bn)
	if err != nil {
		return n, err
	}
	// 5.3. Check if body is fully read
	if bn != cl {
		se := fmt.Sprintf("textproto: content length '%d' doesn't match bytes written to w'%d'", cl, bn)
		return n, fmt.Errorf(se)
	}
	// 6. Return bytes read and error
	return n, err
}

// ReadTextMessage reads from r and parses the text message
func ReadTextMessage(r *bufio.Reader) (*TextMessage, error) {
	tp := &TextMessage{}
	// Parse
	parsedMess, err := newParsedTextMessage(r)
	if err != nil {
		return nil, err
	}
	// Set values
	tp.Status = parsedMess.status
	tp.Headers = parsedMess.headers
	tp.ContentLen = parsedMess.cl
	tp.bytesRead = int64(parsedMess.headLen)
	tp.Body = parsedMess.body
	tp.isParsed = true
	tp.isReadBody = false
	return tp, nil
}

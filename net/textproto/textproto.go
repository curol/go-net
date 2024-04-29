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
	isReadTextMessage bool
	isReadBody        bool // true if body read
	r                 *bufio.Reader
	bytesRead         int64
}

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

// Strings returns the TextMessage as a slice of strings without the \r\n
func (tp *TextMessage) Strings() []string {
	return strings.Split(string(tp.Bytes()), "\r\n")
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

// Read body to p
func (tm *TextMessage) Read(p []byte) (int, error) {
	return tm.Body.Read(p)
}

// Size returns the size of the TextMessage
func (tp *TextMessage) Size() int64 {
	return int64(tp.bytesRead) + tp.ContentLen
}

// WriteTo serializes the TextMessage and writes to w
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

	return tm.serialize(bw)
}

// var s strings.Builder
// s.WriteString(tp.Status + "\r\n") // Status line
// h := string(serializeHeaders(tp.Headers))
// s.WriteString(h) // Headers
// s.WriteString("\r\n") // Blank line between headers and body
//
// b := bytes.NewBuffer(nil)
// b.WriteString(tp.Status + "\r\n") // Status line
// b.Write(serializeHeaders(tp.Headers)) // Headers
// b.Write([]byte("\r\n")) // Blank line between headers and body
func (tp *TextMessage) serialize(bw *bufio.Writer) (int64, error) {
	// var bw *bufio.Writer
	var n int64
	var err error
	headers := tp.Headers
	dlm := "\r\n"

	// clean up
	defer func() {
		err = bw.Flush()
		if err != nil {
			fmt.Println("textproto: Error flushing serialization to w - ", err)
		}
	}()

	// status line
	sn, _ := fmt.Fprintf(bw, "%s%s", tp.Status, dlm) // Write status line
	n = int64(sn)

	// headers
	// h := serializeHeaders(tp.Headers)
	// n2, _ = fmt.Fprint(bw, string(h)) // Write headers
	for name, values := range headers {
		for _, value := range values {
			hn, _ := fmt.Fprintf(bw, "%s: %s%s", name, value, dlm)
			n += int64(hn)
		}
	}

	// end of headers
	ehn, _ := fmt.Fprintf(bw, dlm) // Write blank line between headers and body
	n += int64(ehn)

	// body
	// 1. Validate
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

	// 2. Copy body to dst
	src := tp.Body
	dst := bw
	bn, err := io.CopyN(dst, src, cl)
	tp.isReadBody = true
	n += int64(bn)
	if err != nil {
		return n, err
	}

	// 4. Check if body is fully read
	if bn != cl {
		se := fmt.Sprintf("textproto: content length '%d' doesn't match bytes written to w'%d'", cl, bn)
		return n, fmt.Errorf(se)
	}

	// 5. Return bytes read
	return n, nil
}

// func (tm *TextMessage) writeBodyTo(w *bufio.Writer) (int64, error) {
// 	src := tm.Body
// 	dst := w
// 	cl := tm.ContentLen

// 	// 1. Validate
// 	if cl < 0 {
// 		return 0, fmt.Errorf("textproto: Content-Length less than 0 %d", cl)
// 	}
// 	if cl == 0 {
// 		return 0, nil
// 	}
// 	if tm.isReadBody {
// 		return 0, fmt.Errorf("textproto: Body already read")
// 	}

// 	// 2. Copy body to dst
// 	defer dst.Flush()
// 	n, err := io.CopyN(dst, src, cl)
// 	tm.isReadBody = true
// 	if err != nil {
// 		return n, err
// 	}

// 	// 4. Check if body is fully read
// 	if n != cl {
// 		se := fmt.Sprintf("textproto: content length '%d' doesn't match bytes written to w'%d'", cl, bn)
// 		return n, fmt.Errorf(se)
// 	}

// 	// 5. Return bytes read
// 	return n, nil
// }

// serialize to buffer
// func (tm *TextMessage) toBuffer() *bytes.Buffer {
// 	var buffer bytes.Buffer
// 	headers := tm.Headers
// 	sl := tm.Status
// 	dlm := "\r\n"

// 	// sl
// 	buffer.WriteString(sl + dlm)
// 	// headers
// 	for name, values := range headers {
// 		for _, value := range values {
// 			buffer.WriteString(name)
// 			buffer.WriteString(": ")
// 			buffer.WriteString(value)
// 			buffer.WriteString(dlm)
// 		}
// 	}
// 	buffer.WriteString(dlm) // end of headers

// 	// body
// 	// TODO: hande error
// 	buffer.ReadFrom(tm.Body)

// 	// return bufio.NewReader(tm.Body)
// }

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
	tp.isReadTextMessage = true
	tp.isReadBody = false
	tp.r = r
	return tp, nil
}

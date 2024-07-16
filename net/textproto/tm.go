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
	"time"

	"github.com/curol/network/net/header"
)

// A TextMessage represents a text-based message.
type TextMessage struct {
	// Head
	Status  string        // first line
	Headers header.Header // 2nd line through blank line
	// Body
	Body       *bufio.Reader // rest of lines
	ContentLen int64         // Content-Length header for size of content
	ContentTyp string        // Content-Tupe header for type of content
	// Meta
	readTimeout  time.Duration
	writeTimeout time.Duration
	readMaxN     int64
	writeMaxN    int64
	isParsed     bool
	parsedN      int64 // # of bytes parsed
	isSerialized bool
	serializedN  int64 // # of bytes serialized
	isBodyRead   bool  // true if body read
}

func NewDefaultTextMessage() *TextMessage {
	tm := &TextMessage{}
	tm.readTimeout = 30 * time.Second
	tm.writeTimeout = 30 * time.Second
	tm.readMaxN = 10 << 20
	tm.writeMaxN = 10 << 20
	return tm
}

// Buffer returns the TextMessage as a *bytes.buffer
func (tm *TextMessage) Buffer() (*bytes.Buffer, error) {
	b := bytes.NewBuffer(nil)
	_, err := tm.WriteTo(b)
	return b, err
}

// Bytes returns the TextMessage as a slice of bytes
func (tm *TextMessage) Bytes() ([]byte, error) {
	if tm.isSerialized {
		return nil, fmt.Errorf("textproto: TextMessage already serialized")
	}
	buf, err := tm.Buffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// StdOut writes the TextMessage to stdout
func (tm *TextMessage) StdOut() (int64, error) {
	if tm.isSerialized {
		return 0, fmt.Errorf("textproto: TextMessage already serialized")
	}
	f := os.Stdout
	w := bufio.NewWriter(f)
	return tm.WriteTo(w)
}

// Lines returns the TextMessage as a slice of strings
func (tm *TextMessage) Lines() ([]string, error) {
	if tm.isSerialized {
		return nil, fmt.Errorf("textproto: TextMessage already serialized")
	}
	b, err := tm.Bytes()
	if err != nil {
		return nil, err
	}
	return strings.Split(string(b), "\r\n"), nil
}

// Head return the head of the TextMessage
func (tm *TextMessage) Head() {
	buf := bytes.NewBuffer(nil)
	w := bufio.NewWriter(buf)
	tm.serialize(w, false)
}

// File saves the content (Body) to disk
func (tm *TextMessage) Content(path string) (int64, error) {
	if tm.isSerialized {
		return 0, fmt.Errorf("textproto: TextMessage already serialized")
	}
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

// Save saves the TextMessage to disk
func (tm *TextMessage) Save(path string) (int64, error) {
	if tm.isSerialized {
		return 0, fmt.Errorf("textproto: TextMessage already serialized")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return tm.WriteTo(f)
}

// Size returns the size of the TextMessage without the body (headers only)
func (tm *TextMessage) Size() int64 {
	if tm.isParsed {
		return tm.parsedN
	} else if tm.Status != "" && tm.Headers != nil {
		sl := len(tm.Status) + 2
		h := header.GetHeaderSize(tm.Headers)
		return int64(sl + h)
	} else {
		return 0
	}
}

// ContentLength returns the size of the TextMessage.Body
func (tm *TextMessage) ContentLength() int64 {
	return tm.ContentLen
}

// ContentType returns the type of the TextMessage.Body
func (tm *TextMessage) ContentType() string {
	return tm.ContentTyp
}

// Read reads TextMessage.Body to p
func (tm *TextMessage) Read(p []byte) (int, error) {
	return tm.Body.Read(p)
}

// WriteTo serializes tm to bytes and writes to w
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

// serialize serializes the TextMessage to bw and returns the number of bytes written
// It writes the status line, headers, and body to w.
//
// Example:
// tp := &TextMessage{Status: "200 OK", Headers: MIMEHeader{"Content-Type": []string{"text/plain"}}, Body: bufio.NewReader(strings.NewReader("Hello, World!"))}
// bw := bufio.NewWriter(os.Stdout)
// n, err := tm.serialize(bw)
func (tm *TextMessage) serialize(w *bufio.Writer, doSerializeBody bool) (int64, error) {
	return serialize(tm, w, doSerializeBody)
}

func (tm *TextMessage) readBody(dst *bufio.Writer, cl int64) (int64, error) {
	src := tm.Body
	if src == nil {
		return 0, fmt.Errorf("textproto: Body is nil")
	}
	// Set read timeout
	readTimeout := tm.readTimeout
	if readTimeout == 0 {
		readTimeout = 30 * time.Second
	}
	peakN := 1

	// Use go routine so reader doesnt block or hang if not enough bytes available
	// ch := make(chan []byte)
	errCh := make(chan error)
	go func() {
		_, err := src.Peek(peakN)
		if err != nil {
			errCh <- err
			return
		}
		errCh <- nil
		return
	}()

	// Wait for the first byte or error
	select {
	case err := <-errCh:
		// If error, dont read body and return error
		if err != nil {
			fmt.Println("Received error:", err)
			return 0, err
		}
		// If no error, read the body to w
		readMaxN := tm.readMaxN
		if readMaxN == 0 {
			readMaxN = 10 << 20
		}
		if cl > readMaxN {
			return 0, fmt.Errorf("textproto: Content-Length '%d' exceeds readMaxN '%d'", cl, readMaxN)
		}
		n, err := copyBody(dst, src, cl)
		tm.isBodyRead = true
		return n, err
	// Timeout
	case <-time.After(readTimeout):
		return 0, fmt.Errorf("textproto: Timeout")
	}
}

func copyBody(w *bufio.Writer, r io.Reader, cl int64) (int64, error) {
	n, err := io.CopyN(w, r, cl)
	if err != nil {
		return n, err
	}
	return n, nil
}

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
	"fmt"
	"strings"

	"github.com/curol/network/net/header"
)

// ReadTextMessage is a helper that reads from r and returns the parsed TextMessage.
func ReadTextMessage(r *bufio.Reader) (*TextMessage, error) {
	tm := NewDefaultTextMessage()
	// Parse
	parsedMess, err := newParsedTextMessage(r)
	if err != nil {
		return nil, err
	}
	// Set values
	tm.Status = parsedMess.status
	tm.Headers = header.Header(parsedMess.headers)
	tm.Body = parsedMess.body
	tm.ContentLen = parsedMess.cl
	// Set meta
	tm.parsedN = int64(parsedMess.n)
	tm.isParsed = true
	tm.isSerialized = false
	tm.isBodyRead = false
	return tm, nil
}

func TrimString(s string) string {
	return strings.TrimSpace(s)
}

func PrintfLine(s string) string {
	s = TrimString(s)
	return fmt.Sprintf("%s\r\n", s)
}

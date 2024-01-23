package util

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/mitchellh/mapstructure"
)

func NewBuffer(b []byte) *bytes.Buffer {
	return bytes.NewBuffer(b)
}

func NewWriter(w io.Writer) *bufio.Writer {
	return bufio.NewWriter(w)
}

func NewReader(r io.Reader) *bufio.Reader {
	// return bytes.NewReader(nil)
	return bufio.NewReader(r)
}

func NewStringReader(s string) *strings.Reader {
	return strings.NewReader(s)
}

func NewNopCloser(r io.Reader) io.ReadCloser {
	return io.NopCloser(r)
}

func CopyReaderToWriter(r io.Reader, w io.Writer, len int64) (int64, error) {
	if len == 0 {
		return 0, fmt.Errorf("can't copy reader of size 0")
	}
	if w == nil {
		return 0, fmt.Errorf("can't copy writer of type nil")
	}
	if r == nil {
		return 0, fmt.Errorf("can't copy reader of type nil")
	}

	// Copy reader to w
	return io.CopyN(w, r, len) // copy reader to writer of size cl
}

// Compare readers
func IsReadersEqual(r1, r2 io.Reader) (bool, error) {
	buf1 := make([]byte, 1024)
	buf2 := make([]byte, 1024)

	for {
		n1, err1 := r1.Read(buf1)
		n2, err2 := r2.Read(buf2)

		if err1 != nil && err1 != io.EOF {
			return false, err1
		}
		if err2 != nil && err2 != io.EOF {
			return false, err2
		}

		if !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false, nil
		}

		if err1 == io.EOF || err2 == io.EOF {
			return err1 == err2, nil
		}
	}
}

func MergeStructs(s1, s2, result interface{}) {
	var m1 map[string]interface{}
	var m2 map[string]interface{}

	mapstructure.Decode(s1, &m1)
	mapstructure.Decode(s2, &m2)

	// Merge the maps
	for k, v := range m2 {
		m1[k] = v
	}

	// Convert the map back to a struct
	mapstructure.Decode(m1, &result)

	fmt.Println(result)
}

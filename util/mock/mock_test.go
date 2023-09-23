package mock

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"
)

func TestStreamFromSTDIN(t *testing.T) {
	// Dont use `os.Stdin` for testing because it doesn't work with `go test`.
	// Instead, create a new bytes.Buffer to simulate standard input.
	input := bytes.NewBufferString("hello\nworld\n")

	result := ExampleStreamFromSTDIN(input)
	fmt.Println("\nBuffer:", result)

	if !bytes.Equal(result, []byte("hello\nworld\n")) {
		t.Error("Expected `hello\nworld\n` but got", result)
	}
}

func TestStreamFromSTDINWithBufio(t *testing.T) {
	// Dont use `os.Stdin` for testing because it doesn't work with `go test`.
	// Instead, create a new bytes.Buffer to simulate standard input.
	input := bytes.NewBufferString("hello\nworld\n")

	// Most cases will use `bufio.NewReader` to read from standard input.
	// for better performance.
	reader := bufio.NewReader(input)

	// Result
	result := ExampleStreamFromSTDIN(reader)

	if !bytes.Equal(result, []byte("hello\nworld\n")) {
		t.Error("Expected `hello\nworld\n` but got", result)
	}

	fmt.Println("\nBuffer:", result)
}

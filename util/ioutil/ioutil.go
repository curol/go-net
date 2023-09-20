package ioutil

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
)

// *********************************************************************************************************************
// Readers/Writers
// *********************************************************************************************************************
// NewReader returns a new Reader whose buffer has the default size.
func NewReader(r io.Reader) *bufio.Reader {
	return bufio.NewReader(r)
}

// NewBuffer creates and initializes a new Buffer using buf as its initial contents.
func NewBuffer(buf []byte) *bytes.Buffer {
	return bytes.NewBuffer(buf)
}

// NewWriter returns a new Writer whose buffer has the default size.
func NewWriter(w io.Writer) *bufio.Writer {
	return bufio.NewWriter(w)
}

// NewWriterAndBuffer returns a new Writer whose buffer has the default size.
func NewWriterAndBuffer() (*bufio.Writer, *bytes.Buffer) {
	// Create a bytes.Buffer
	var b bytes.Buffer
	// Create a bufio.Writer that writes to the bytes.Buffer
	return NewWriter(&b), &b
}

// NewFile returns a new File with the given name, file flag and file mode.
func Open(name string, flag int, perm os.FileMode) (*os.File, error) {
	// Flags:
	//   os.O_RDONLY = 0
	// 	 os.O_WRONLY = 1
	// 	 os.O_RDWR = 2
	// 	 os.O_CREATE = 64
	// 	 os.O_TRUNC = 512
	// 	 os.O_APPEND = 1024
	// 	 os.O_WRONLY|os.O_CREATE|os.O_TRUNC = 0644
	return os.OpenFile(name, flag, perm)
}

// NewReadOnlyFile returns a new File with the O_RDONLY flag set.
func OpenReadOnly(name string) (*os.File, error) {
	return os.Open(name)
}

// NewRWFile returns a new File with the O_RDWR | os.O_CREATE | os.O_TRUNC flag set
// and permissions 0644, which gives read and write permissions to the owner of the file,
// and read-only permissions to everyone else.
func OpenRWFile(fn string) (*os.File, error) {
	// Write only, create if doesnt exist, and truncate if it does
	flag := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	// `0644` gives read and write permissions to the owner of the file, and read-only permissions to everyone else.
	perm := fs.FileMode(0644)
	// Open stream to file
	return os.OpenFile(fn, flag, perm)
}

// NopCloser returns a ReadCloser with a no-op Close method wrapping the provided Reader r.
func NoCloser(r io.Reader) io.ReadCloser {
	return io.NopCloser(r)
}

// *********************************************************************************************************************
// Read
// *********************************************************************************************************************

// ReadAll reads from r until an error or EOF and returns the data it read.
func ReadAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

// ReadFull reads exactly len(buf) bytes from r into buf.
func ReadFull(r io.Reader, buf []byte) (n int, err error) {
	return io.ReadFull(r, buf)
}

// ReadLine reads a line from r until it finds a \n or io.EOF.
func ReadLine(r io.Reader) ([]byte, error) {
	var line []byte
	// Since buffer is size 1, we can only read 1 byte at a time.
	buf := make([]byte, 1)

	for {
		// Read 1 byte from r into buf
		_, err := r.Read(buf)
		if err != nil {
			return line, err
		}
		// If buf[0] is a newline, break
		if buf[0] == '\n' {
			break
		}
		// Append buf[0] to line
		line = append(line, buf[0])
	}

	return line, nil
}

// ReadAndClose reads from readCloser into buf.
func ReadAndClose(readCloser io.ReadCloser, buf []byte) error {
	// readAndClose reads from readCloser into buf.
	//
	// Example of reading from a ReadCloser:
	//
	// ```
	// // Create a buffer of the desired size n
	// n := 10
	// buf := make([]byte, n)
	// // Read n bytes from readCloser into buf
	// readCloser.Read(buf)
	// readCloser.Close()
	// ```
	// **Note:** The ReadCloser must be closed after reading from it.

	// Read n bytes from readCloser into buf
	_, err := readCloser.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("Error reading data: %v", err)
	}
	// When finished reading from readCloser, close it
	err = readCloser.Close()
	if err != nil {
		return fmt.Errorf("Error closing ReadCloser: %v", err)
	}
	return nil
}

// Read reads up to len(buf) bytes from the File and stores them in buf.
func ReadFile(fn string, buf []byte) (int, error) {
	f, err := OpenReadOnly(fn)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Read(buf)
}

// *********************************************************************************************************************
// Write
// *********************************************************************************************************************

// Write writes len(p) bytes from p to the File.
func WriteFile(fn string, data []byte) (int, error) {
	// Write only, create if doesnt exist, and truncate if it does
	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	// `0644` gives read and write permissions to the owner of the file, and read-only permissions to everyone else.
	perm := fs.FileMode(0644)

	// Open stream to file
	fd, err := os.OpenFile(fn, flag, perm)
	if err != nil {
		return 0, err
	}
	// Close when finished writing
	defer fd.Close()
	// Write data to file
	return fd.Write(data)
}

// Copy copies n bytes (or until an error) from src to dst.
func CopyReaderToWriter(dst io.Writer, src io.Reader, n int64) (int64, error) {
	return io.CopyN(dst, src, n)
}

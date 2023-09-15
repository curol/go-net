package reader

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestStream(t *testing.T) {
	b := make([]byte, 10)

	for i := 0; i < len(b); i++ {
		b[i] = byte(i)
		ReadStream("example.txt")
	}

}

func mockFileStream(data string) (io.Reader, error) {
	// Create a temporary file
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		return nil, err
	}

	// Write some data to the file
	_, err = tmpfile.Write([]byte(data))
	if err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		return nil, err
	}

	// Seek to the beginning of the file
	_, err = tmpfile.Seek(0, 0)
	if err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		return nil, err
	}

	// Return the file as an io.Reader
	return tmpfile, nil
}

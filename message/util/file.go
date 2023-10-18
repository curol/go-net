//**********************************************************************************************************************
// File
//**********************************************************************************************************************

package util

import "os"

// CreateFile creates or truncates the named file. If the file already exists, it is truncated. If the file does not exist, it is created with mode 0666 (before umask). If successful, methods on the returned File can be used for I/O; the associated file descriptor has mode O_RDWR. If there is an error, it will be of type *PathError.
func createFile(filename string) *os.File {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	return file
}

func readFile(filename string) []byte {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return data
}

func writeToFile(filename string, data []byte) int {
	f := createFile(filename)
	n, err := f.Write(data)
	if err != nil {
		panic(err)
	}
	return n
}

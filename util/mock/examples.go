package mock

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
)

func ExampleWriterToBuffer(data string) {
	// Create a bytes.Buffer
	var b bytes.Buffer

	// Create a bufio.Writer that writes to the bytes.Buffer
	w := bufio.NewWriter(&b)

	// Write some data to the bufio.Writer
	w.WriteString(data)

	// Ensure all data has been written to the underlying buffer
	w.Flush()

	// Get the []byte from the bytes.Buffer
	result := b.Bytes()

	fmt.Println(string(result)) // Outputs: Hello, World!
}

// Read the buffer in chunks of 5 bytes using a for loop.
// In each iteration of the loop, we use reader.Peek to read the next chunk of data from the buffer without advancing the reader's position. We then print the chunk as a string using fmt.Println.
func ExampleReadInChunks(str string) {
	// Create a new buffer with some data
	// buffer := bytes.NewBufferString("Hello, World!")
	buffer := bytes.NewBufferString(str)

	// Create a new bufio.Reader from the buffer
	reader := bufio.NewReader(buffer)

	// Read the buffer in chunks of 5 bytes
	chunkSize := 5
	for {
		// Read the next chunk from the buffer
		chunk, err := reader.Peek(chunkSize)
		if err != nil && err != io.EOF {
			panic(err)
		}

		// Print the chunk as a string
		fmt.Println(string(chunk))

		// Advance the reader by the size of the chunk
		_, err = reader.Discard(len(chunk))
		if err != nil {
			panic(err)
		}

		// If we've reached the end of the buffer, break out of the loop
		if len(chunk) < chunkSize {
			break
		}
	}
}

// Read the buffer in chunks of 5 bytes using a for loop.
// In each iteration of the loop, we create a new byte slice to hold the chunk using make. We then use reader.Read to read the next chunk of data from the buffer into the byte slice. The n variable contains the number of bytes read.
func ExampleReadInChunksWithoutPeek(data string) {
	// Create a new buffer with some data
	buffer := bytes.NewBufferString(data)

	// Create new reader wrapping reader with bufio.Reader from the buffer
	reader := bufio.NewReader(buffer)

	// Read the buffer in chunks of 5 bytes
	chunkSize := 5
	for {
		// Create a new byte slice to hold the chunk
		chunk := make([]byte, chunkSize)

		// Read the next chunk from the buffer
		// The bytes are taken from at most one Read on the underlying Reader.
		n, err := reader.Read(chunk)
		if err != nil && err != io.EOF {
			panic(err)
		}

		// Print the chunk as a string
		fmt.Println(string(chunk[:n]))

		// If we've reached the end of the buffer, break out of the loop
		if err == io.EOF {
			break
		}
	}
}

// Read the file in chunks of 5 bytes using a for loop. In each iteration of the loop, we create a new byte slice to hold the chunk using make. We then use reader.Read to read the next chunk of data from the file into the byte slice. The n variable contains the number of bytes read.
func ExampleReadFileInChunks(fn string) {
	// Open the file for reading
	file, err := os.Open(fn)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create a new bufio.Reader from the file
	reader := bufio.NewReader(file)

	// Read the file in chunks of 5 bytes
	chunkSize := 5
	for {
		// Create a new byte slice to hold the chunk
		chunk := make([]byte, chunkSize)

		// Read the next chunk from the file
		n, err := reader.Read(chunk)
		if err != nil && err != io.EOF {
			panic(err)
		}

		// Print the chunk as a string
		fmt.Println(string(chunk[:n]))

		// If we've reached the end of the file, break out of the loop
		if err == io.EOF {
			break
		}
	}
}

// We then read the connection in chunks of 5 bytes using a for loop. In each iteration of the loop, we create a new byte slice to hold the chunk using make. We then use reader.Read to read the next chunk of data from the connection into the byte slice. The n variable contains the number of bytes read.
func ExampleReadConnectionInChunks(network string, address string) {
	// Connect to the remote server
	conn, err := net.Dial(network, address)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Create a new bufio.Reader from the connection
	reader := bufio.NewReader(conn)

	// Read the connection in chunks of 5 bytes
	chunkSize := 5
	for {
		// Create a new byte slice to hold the chunk
		buffer := make([]byte, chunkSize)

		// Readthe next chunk from the connection
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			panic(err)
		}

		// Print the chunk as a string
		fmt.Println(string(buffer[:n]))

		// If we've reached the end of the connection, break out of the loop
		if err == io.EOF {
			break
		}
	}
}

// Read the connection in chunks of 5 bytes using a for loop. We create a byte slice called chunk with a length of chunkSize before the loop. This slice is reused on each iteration of the loop.
// In each iteration of the loop, we use reader.Read to read the next chunk of data from the connection into the chunk byte slice. The n variable contains the number of bytes read.
func ExampleReadConnectionInChunksReusingBuffer() {
	// Connect to the remote server
	conn, err := net.Dial("tcp", "example.com:80")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Create a new bufio.Reader from the connection
	reader := bufio.NewReader(conn)

	// Read the connection in chunks of 5 bytes
	chunkSize := 5
	buffer := make([]byte, chunkSize)
	for {
		// Read the next chunk from the connection
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			panic(err)
		}

		// Print the chunk as a string
		fmt.Println(string(buffer[:n]))

		// If we've reached the end of the connection, break out of the loop
		if err == io.EOF {
			break
		}
	}
}

func ReadInChunksOfFive(reader io.Reader) {
	chunkSize := 5
	buffer := make([]byte, chunkSize)

	for {
		// Read the next chunk from the connection
		n, err := reader.Read(buffer)

		if err != nil && err != io.EOF {
			panic(err)
		}

		// Print the chunk as a string
		fmt.Println(string(buffer[:n]))

		// If we've reached the end of the connection, break out of the loop
		if err == io.EOF {
			break
		}
	}
}

// Read from src and write to dst.
func ExampleReadAndWriteToFileStream(src string, dst string) {
	// Open the input file for reading
	inputFile, err := os.Open(src)
	if err != nil {
		panic(err)
	}
	defer inputFile.Close()

	// Create a new bufio.Reader from the input file
	inputReader := bufio.NewReader(inputFile)

	// Open the output file for writing
	outputFile, err := os.Create(dst)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	// Create a new bufio.Writer for the output file
	outputWriter := bufio.NewWriter(outputFile)

	// Read the input file line by line and write to the output file
	for {
		// Read the next line from the input file
		line, err := inputReader.ReadString('\n')
		if err != nil && err != io.EOF {
			panic(err)
		}

		// Write the line to the output file
		_, err = outputWriter.WriteString(line)
		if err != nil {
			panic(err)
		}

		// If we've reached the end of the input file, break out of the loop
		if err == io.EOF {
			break
		}
	}

	// Flush the output buffer to ensure all data is written to the file
	// Note that calling Flush too frequently can reduce performance, since it involves writing data to the output stream more often.
	// It's generally best to let the buffer fill up before calling Flush, unless you need to ensure that all data has been written to the output stream immediately.
	err = outputWriter.Flush()
	if err != nil {
		panic(err)
	}
}

func ExampleStreamFromSTDIN(r io.Reader) []byte {
	// Buffer
	buf := bytes.NewBuffer(nil)

	// Create Reader
	reader := bufio.NewReader(r)

	for {
		// Read
		line, err := reader.ReadString('\n')

		// If err not io.EOF, panic
		if err != nil && err != io.EOF {
			panic(err)
		}

		// Break when finished reading
		if err == io.EOF {
			fmt.Println("EOF received.")
			break
		}

		// Write to buffer
		fmt.Println("Line:", line)
		buf.Write([]byte(line))
	}

	return buf.Bytes()
}

func ExampleScanSTDIN() {
	// Create a new bufio.Scanner to read from standard input
	scanner := bufio.NewScanner(os.Stdin)

	// Loop over standard input
	for scanner.Scan() {
		// Print the input as a string
		fmt.Println("Text:", scanner.Text())
	}

	// Check if there was an error reading from standard input
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

// ExampleCopyReaderToWriter copies the contents of a reader to a writer.
//
// In Go, you can use the io.Copy function to copy data from one io.Reader to another io.Writer.
// This can be useful when you want to copy the contents of one file to another file, for example.
func ExampleCopyReaderToWriter(r string, w string) {
	// Open the input file for reading
	inputFile, err := os.Open(r)
	if err != nil {
		panic(err)
	}
	defer inputFile.Close()

	// Open the output file for writing
	outputFile, err := os.Create(w)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	// Copy the contents of the input file to the output file
	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		panic(err)
	}
}

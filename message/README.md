
# Message

## Head

## Body
		 Body is the payload of the message.
		
		 For small data, the entire body can be read at once.
		 For large data, the body can be streamed in chunks.
		 For example, a file can be streamed in chunks.
		 For example, a video can be streamed in chunks.
		 For example, a large JSON object can be streamed in chunks.
		
		 Why carefully read the body with content-length?
		 Because the size of the data might be too big for server to read all at once.
		 For streams, we don't want to read the entire body at once into memory.
		 If its too big, we might run out of memory.
		 Instead, implement strategy for different sizes of the data and use chunks or buffers to read from the connection.


## Reader

The io.Reader interface in Go is a fundamental interface that represents the read end of a stream of data.
It has a single method, Read, which reads up to len(p) bytes into p.

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

The Read method takes a byte slice as an argument, fills it with data and returns the number of bytes read and an error, if any occurred.
If there is no more data to be read, Read returns io.EOF (end-of-file) error.

Here's how you might use it:

```go
buf := make([]byte, 1024)
n, err := reader.Read(buf)
if err != nil && err != io.EOF {
    panic(err)
}
fmt.Println("Bytes read:", n)
```

In this example, reader is an io.Reader. We create a buffer buf and pass it to reader.Read.
The Read method fills buf with data and returns the number of bytes read.
If an error occurs during the read operation and it's not io.EOF, we panic.
Otherwise, we print the number of bytes read.

### ReadCloser
ReadCloser is the interface that groups the basic Read and Close methods.
	
```
	 type ReadCloser interface {
	   Reader
	 	 Closer
	 }
```


## Writer

In Go, the io.Writer interface is a fundamental interface that represents the write end of a stream of data.
It has a single method, Write, which writes some data and returns the number of bytes written and an error, if any occurred.

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

The Write method takes a byte slice as input, writes as much of it as possible, and returns the number of bytes it wrote and an error value.
If the number of bytes written is less than the length of the input byte slice, it should return an error explaining why it couldn't write the whole slice.

The Flush method is not part of the io.Writer interface, but it is often found in buffered writers like bufio.Writer. The purpose of Flush is to clear the buffer, writing any buffered data to the underlying writer.

```go
type Writer interface {
    Write(p []byte) (n int, err error)
    Flush() error
}
```

When you write data to a buffered writer, the data is not usually written immediately to the underlying writer.
Instead, it's held in a buffer, and only written when the buffer becomes full.
This can improve performance by reducing the number of write operations.
However, sometimes you need to ensure that all written data is sent immediately, without waiting for the buffer to fill up.
That's when you would call Flush.
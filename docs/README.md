# Network

This package is experimental for research & development in networking and I/O.

Its primary job is to develop simplified networking implementations, wrap existing implementations of such network primitives, such as those in package net, into shared public interfaces that abstract the functionality. 



## Connection

A network connection often refers to the communication link between two programs running on the network. This connection allows the programs to exchange data.

There are various protocols that govern how this data exchange happens, such as TCP (Transmission Control Protocol) and UDP (User Datagram Protocol). These protocols define rules for how the data should be packaged, addressed, transmitted, routed, and received at the destination.

**A network connection can be established using various methods, such as sockets**. Once a connection is established, data can be sent back and forth until the connection is closed.

## Protocol

### TCP

TCP (Transmission Control Protocol) is one of the main protocols in the Internet protocol suite. It's a transport layer protocol that provides reliable, ordered, and error-checked delivery of a stream of bytes between applications running on hosts communicating over an IP network.

TCP is used by many popular application layer protocols, such as HTTP, HTTPS, FTP, SMTP, and more. It's connection-oriented, meaning it requires a connection to be established before data can be sent. This is in contrast to connectionless protocols like UDP, which don't require a connection.

## File

In Go, a file is represented by the `*os.File` type, which is a pointer to an `os.File` value. This type provides methods for file operations, such as reading, writing, and seeking.

When you open a file using the `os.Open` function or create a new file using the `os.Create` function, these functions return a `*os.File` value that you can use to manipulate the file. Here's an example:

```go
file, err := os.Open("myfile.txt")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// Now you can use the `file` value to read from or write to the file.
```

In this example, `os.Open` opens the file named "myfile.txt" and returns a `*os.File` value. The `defer file.Close()` line ensures that the file will be closed when the function returns.

The `*os.File` type also implements the `io.Reader`, `io.Writer`, `io.Seeker`, and `io.Closer` interfaces, so you can use it with functions that work with these interfaces. For example, you can use it with the `io.Copy` function to copy data from one file to another, or with the `bufio.Scanner` type to read the file line by line.

### Everything is a file

In Unix and Unix-like operating systems, including Linux, there's a concept that "everything is a file".
This means that all I/O devices, including network connections, are treated as files.
You can read from and write to a network connection using the same system calls (read, write, etc.) that you would use for regular files.

However, a network connection is not a file in the traditional sense.
It doesn't have a size, you can't seek in it, and reading from or writing to it has side effects (like sending or receiving data over the network).
But from a programming perspective, it's often treated like a file because you can use file-like operations to interact with it.

In many programming languages, including Go, network connections are represented by objects or data structures that provide file-like methods. 
For example, in Go, the net.Conn interface provides Read and Write methods for network I/O, similar to the os.File type for file I/O.

#### FD

A fd (file descriptor) is a low-level concept in Unix and Unix-like operating systems, including Linux. It's an integer that the operating system assigns to an opened file, socket, or other resource to track it or uniquely identify it.

When a process opens a file or other resource, the operating system assigns it a file descriptor. The process can then use this file descriptor to read from or write to the file or resource.

File descriptors are used in many system calls and functions in the C library, and they're also used in many other programming languages when dealing with low-level I/O. In Go, for example, the `os` package provides functions like `os.Open` and `os.Create` that return an `*os.File` value, which is a higher-level wrapper around a file descriptor.
They are used by system calls such as open, read, write, close, select, and poll.

In Go, file descriptors are represented by the os.File type. When you open a file using os.Open, for example, you get an os.File object that represents the file and provides methods for reading from and writing to it. Under the hood, the os.File object is backed by a file descriptor.

### Stream

In computing, a file stream is a sequence of data bytes that are read from or written to a file. A file stream is a higher-level abstraction than a file descriptor, which is a low-level identifier used by the operating system to represent an open file.

A file stream provides a buffered interface for reading from or writing to a file. **This means that data is read from or written to the file in chunks, rather than one byte at a time.** This can improve performance, especially when reading or writing large amounts of data.

In Go, file streams are represented by the `os.File` type. When you open a file using `os.Open`, for example, you get an `os.File` object that represents the file and provides methods for reading from and writing to it. The `os.File` type provides a buffered interface for reading and writing, using the bufio.Reader and bufio.Writer types, respectively.

File streams are used extensively in programming for input/output operations, such as reading from or writing to files, network sockets, and other types of input/output resources.

### Sockets

In the context of networking, a socket is an endpoint in a communication flow between two systems. Sockets provide a mechanism for the exchange of data between a client program and a server program in a network. A socket is bound to a specific port number so that the TCP layer can identify the application that data is destined to be sent to.

There are two main types of sockets:

- TCP Sockets: These are reliable, connection-oriented sockets. They ensure that data arrives intact and in order at the destination. TCP sockets are used where reliability is more critical than speed, such as loading a webpage or sending an email.

- UDP Sockets: These are connectionless and do not guarantee the delivery or the order of data. They are used where speed is more critical than reliability, such as in video streaming or online gaming.

In programming, a socket API (Application Programming Interface) allows for the creation and management of sockets. In Go, the net package provides functionalities for creating TCP and UDP sockets.


### I/O 

#### Reader & Writer

In Go, `io.Reader` and `io.Writer` are two fundamental interfaces used for reading from and writing to a variety of data streams, including files, network connections, and other data sources or sinks.

1. **io.Reader**: This interface represents the read end of a data stream. It has a single method, `Read`, which attempts to fill a provided byte slice with data and returns the number of bytes read and an error value. It's used as the source of input in many functions and types across the Go standard library.

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

2. **io.Writer**: This interface represents the write end of a data stream. It has a single method, `Write`, which takes a byte slice and writes as much of it as possible, returning the number of bytes written and an error value. It's used as the destination of output in many functions and types across the Go standard library.

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

For example, the `os.File` type in Go implements both `io.Reader` and `io.Writer`, so you can use it to read from and write to files. Similarly, the `net.Conn` type in the `net` package also implements these interfaces, so you can use it to read from and write to network connections.

#### Non-blocking I/O model

In a traditional blocking I/O model, when you issue a read or write operation on a file descriptor (which could represent a network socket), the operation blocks until it can be completed. This means if data is not available for reading, the read operation will wait until data becomes available. Similarly, if the system is not ready to accept more data for writing, the write operation will wait. This can be inefficient if you have many connections, as your program can spend a lot of time waiting.

**Non-blocking I/O with polling is a method used in network programming for handling multiple connections simultaneously.**

Non-blocking I/O changes this model. When you set a file descriptor to non-blocking mode, read and write operations will return immediately, even if they can't be completed. If a read operation can't read any data, it will return an error. If a write operation can't write all the data, it will write as much as it can and return the number of bytes written.

**Polling is a technique used in conjunction with non-blocking I/O to check the status of multiple file descriptors.**
Instead of issuing a read or write operation and possibly getting an error, you can use a polling system call like `select`, `poll`, or `epoll` to check which file descriptors are ready for reading or writing. Then, you can issue read or write operations only on those file descriptors, avoiding the need to handle errors from non-blocking operations.

In your code, the `poll.DupCloseOnExec` function is likely a part of this non-blocking I/O with polling model. It seems to duplicate the file descriptor `f.Fd()`, setting the new file descriptor to close-on-exec (meaning it will be closed automatically when a new program is executed) and possibly also setting it to non-blocking mode.











## Examples

### Read

Read input.

#### Read file using bufio.Reader

When reading from a file in Go, it's generally best practice to use a bufio.Reader to buffer the input. This can improve performance by reducing the number of system calls needed to read the file.

```go
// Buffer
chunkSize := 5
buffer := make([]byte,chunkSize)

// Storage
file := os.Open("example.txt")

// Read until all bytes read from underlying data
for{
    // Read data into buffer
    n,err := reader.Read(buffer)

    // Break when done
    if err != nil || err == io.EOF{
        break
    }

    // Write contents to destination
    file.Append(buffer[:n])
}
```

#### Read File

```go
    // Open the file for reading
    file, err := os.Open("example.txt")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    // Create a new bufio.Reader from the file
    reader := bufio.NewReader(file)

    // Read the file line by line
    for {
        // Read the next line from the file
        line, err := reader.ReadString('\n')
        if err != nil && err != io.EOF {
            panic(err)
        }

        // Print the line
        fmt.Print(line)

        // If we've reached the end of the file, break out of the loop
        if err == io.EOF {
            break
        }
    }
```

### Write

Write output.

#### Read and Write to file


```go
package main

import (
    "bufio"
    "io"
    "os"
)

func main() {
    // Open the input file for reading
    inputFile, err := os.Open("input.txt")
    if err != nil {
        panic(err)
    }
    defer inputFile.Close()

    // Create a new bufio.Reader from the input file
    inputReader := bufio.NewReader(inputFile)

    // Open the output file for writing
    outputFile, err := os.Create("output.txt")
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
    err = outputWriter.Flush()
    if err != nil {
        panic(err)
    }
}
```


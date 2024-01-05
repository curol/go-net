# Net

Package network is an experimental package for network I/O.

## Connection

A network connection often refers to the communication link between two programs running on the network. This connection allows the programs to exchange data.

There are various protocols that govern how this data exchange happens, such as TCP (Transmission Control Protocol) and UDP (User Datagram Protocol). These protocols define rules for how the data should be packaged, addressed, transmitted, routed, and received at the destination.

A network connection can be established using various methods, such as sockets in many programming languages. Once a connection is established, data can be sent back and forth until the connection is closed.

## Everything is a file

In Unix and Unix-like operating systems, including Linux, there's a concept that "everything is a file". This means that all I/O devices, including network connections, are treated as files. You can read from and write to a network connection using the same system calls (read, write, etc.) that you would use for regular files.

However, a network connection is not a file in the traditional sense. It doesn't have a size, you can't seek in it, and reading from or writing to it has side effects (like sending or receiving data over the network). But from a programming perspective, it's often treated like a file because you can use file-like operations to interact with it.

In many programming languages, including Go, network connections are represented by objects or data structures that provide file-like methods. For example, in Go, the net.Conn interface provides Read and Write methods for network I/O, similar to the os.File type for file I/O.

## TCP

TCP (Transmission Control Protocol) is one of the main protocols in the Internet protocol suite. It's a transport layer protocol that provides reliable, ordered, and error-checked delivery of a stream of bytes between applications running on hosts communicating over an IP network.

TCP is used by many popular application layer protocols, such as HTTP, HTTPS, FTP, SMTP, and more. It's connection-oriented, meaning it requires a connection to be established before data can be sent. This is in contrast to connectionless protocols like UDP, which don't require a connection.

## Text Protocol

A text protocol network connection is a type of network connection where the communication between the client and the server is done using human-readable text strings. This is in contrast to binary protocols, where communication is done using binary data that is not directly readable by humans.

Examples of text protocols include HTTP, SMTP, FTP, and IRC. These protocols use text commands and responses to communicate between the client and the server. For example, in HTTP, a client might send a text command like `GET /index.html HTTP/1.1` to request a web page from a server.

Text protocols have the advantage of being easy to debug and understand, since you can read the commands and responses directly. However, they can be less efficient than binary protocols, since text data is larger and slower to process than binary data.

## Packages

### Internal Packages

The internal keyword in Go is a special directory name that restricts the accessibility of the packages inside it.
Packages inside an internal directory can only be imported and used by the code that is in the same parent directory.

The internal package is mainly used for internal implementation details that are shared across multiple packages within the parent package.

In the case of net/internal, it contains implementation details and helper functions that are used by other packages within the net package, but are not intended to be directly used by programs that import the net package. This is a way to hide implementation details and prevent them from becoming part of the package's public API.

## Read File

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


## Read and Write to file

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

## Read

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


## FD

In Unix-based operating systems, a file descriptor is a non-negative integer that uniquely identifies an open file or other input/output resource, such as a network socket or a pipe.

When a process opens a file or other resource, the operating system assigns it a file descriptor. The process can then use this file descriptor to read from or write to the file or resource.

File descriptors are used extensively in Unix-based systems for input/output operations. They are used by system calls such as open, read, write, close, select, and poll.

In Go, file descriptors are represented by the os.File type. When you open a file using os.Open, for example, you get an os.File object that represents the file and provides methods for reading from and writing to it. Under the hood, the os.File object is backed by a file descriptor.


## Stream

In computing, a file stream is a sequence of data bytes that are read from or written to a file. A file stream is a higher-level abstraction than a file descriptor, which is a low-level identifier used by the operating system to represent an open file.

A file stream provides a buffered interface for reading from or writing to a file. This means that data is read from or written to the file in chunks, rather than one byte at a time. This can improve performance, especially when reading or writing large amounts of data.

In Go, file streams are represented by the os.File type. When you open a file using os.Open, for example, you get an os.File object that represents the file and provides methods for reading from and writing to it. The os.File type provides a buffered interface for reading and writing, using the bufio.Reader and bufio.Writer types, respectively.

File streams are used extensively in programming for input/output operations, such as reading from or writing to files, network sockets, and other types of input/output resources.

## TCP

TCP stands for Transmission Control Protocol. It is one of the main protocols in the Internet protocol suite, which also includes IP (Internet Protocol). Together, they are often referred to as TCP/IP.

TCP is a transport layer protocol that provides reliable, ordered, and error-checked delivery of a stream of bytes between applications running on hosts communicating over an IP network.

Key features of TCP include:

- Reliability: TCP ensures that data sent from one end of the connection actually gets to the other end and in the same order it was sent. If any data is lost during transmission, TCP will retransmit that data.

- Ordered Data Transfer: If data segments arrive in the wrong order, TCP will reassemble them into their original order.

- Error Checking: TCP includes a checksum in its header for error checking of the header and data. This helps ensure that the data is not corrupted during transmission.

- Flow Control: TCP uses a sliding window for flow control to avoid overwhelming the receiver with more data than it can process.

- Congestion Control: TCP has mechanisms to reduce its data transfer rate when network congestion is detected.

In the context of programming, TCP is often used for applications that require high reliability, and where it is acceptable for there to be a slight delay in order to deliver the data in the correct order and without errors. Examples include web servers, email, file transfers, and virtual private networks (VPNs). In Go, the net package provides functionalities for TCP networking.


### Sockets

In the context of networking, a socket is an endpoint in a communication flow between two systems. Sockets provide a mechanism for the exchange of data between a client program and a server program in a network. A socket is bound to a specific port number so that the TCP layer can identify the application that data is destined to be sent to.

There are two main types of sockets:

- TCP Sockets: These are reliable, connection-oriented sockets. They ensure that data arrives intact and in order at the destination. TCP sockets are used where reliability is more critical than speed, such as loading a webpage or sending an email.

- UDP Sockets: These are connectionless and do not guarantee the delivery or the order of data. They are used where speed is more critical than reliability, such as in video streaming or online gaming.

In programming, a socket API (Application Programming Interface) allows for the creation and management of sockets. In Go, the net package provides functionalities for creating TCP and UDP sockets.

## HTTP
HTTP stands for Hypertext Transfer Protocol. It is a protocol used for transmitting hypertext requests and information between servers and browsers. HTTP is the foundation of data communication on the World Wide Web.

Here are some key points about HTTP:

- Stateless: Each request from client to server is processed independently, without any knowledge of the requests that came before it.

- Connectionless: After a request is made, the client disconnects from the server and waits for a response. The server processes the request and re-establishes the connection with the client to send the response back.

- Media Independent: Any type of data can be sent by HTTP as long as both the client and the server know how to handle the data content.

- Methods: HTTP uses methods (also known as verbs) to indicate the desired action to be performed on the identified resource. The most common methods include GET, POST, PUT, DELETE, and HEAD.

In the context of programming, HTTP is often used for API calls, web scraping, and other tasks that involve sending or receiving data over the internet. In Go, the net/http package provides functionalities for HTTP requests and responses.

### HTTP Message Format

An HTTP message consists of a start-line, headers, and a body, separated by CRLF (\r\n) sequences. The body of the message, also known as the payload, comes after the headers and is optional.

Here's a basic format of an HTTP request message with a body:

```
POST /path HTTP/1.1\r\n
Host: www.example.com\r\n
Content-Type: application/x-www-form-urlencoded\r\n
Content-Length: length\r\n
\r\n
field1=value1&field2=value2
```

    - The start-line is POST /path HTTP/1.1 and indicates this is a POST request to the /path resource.
    - The Host header is required in HTTP/1.1 and specifies the domain of the server.
    - The Content-Type header indicates the media type of the body. In this case, it's application/x-www-form-urlencoded, which is commonly used when submitting HTML forms.
    - The Content-Length header indicates the length of the body in bytes.
    - After the headers, there's an empty line (\r\n), which indicates the end of the headers and the start of the body.
    - The body is field1=value1&field2=value2, which is a URL-encoded form data.

For a JSON payload, the Content-Type would be application/json and the body would contain a JSON object:

```
POST /path HTTP/1.1\r\n
Host: www.example.com\r\n
Content-Type: application/json\r\n
Content-Length: length\r\n
\r\n
{"field1":"value1","field2":"value2"}
```

### HTTP Sessions

An HTTP session is a sequence of network request-response transactions. An HTTP client initiates a request. It then waits for the response. The HTTP server processes the request and sends a response back to the client. This exchange is considered one HTTP session.

However, HTTP itself is stateless, meaning each request is processed independently without any knowledge of the requests that came before it. To maintain state between different requests (such as information about whether a user is logged in), many web applications use session cookies, which are sent in the HTTP headers.

A session cookie might contain a session ID that the server can use to retrieve stored information about the client, such as user preferences, shopping cart contents, or authentication status. This allows the server to provide a personalized experience for each client.

In Go, you can use the net/http package's SetCookie function to set session cookies, and the Request type's Cookie method to retrieve them. There are also third-party packages available that provide more advanced session management features.

## Parsing

Parsing, in the context of programming, is the process of analyzing a string of symbols, either in natural language or computer languages, according to the rules of a formal grammar. Essentially, it involves taking input data (like code or text), breaking it down, and turning it into a format that's more usable for the program.

## Serialization

Serialization of a struct is the process of converting a structured data type (such as a struct) into a format that can be transmitted or stored, such as a byte slice or a string. The serialized data can then be transmitted over a network, stored in a file, or sent to another program.

In the context of the excerpt you provided, the `serializeHead` method of the `Response` struct serializes the head of an HTTP response (the status line and headers) to a byte slice. The serialized data can then be sent over a network to a client that requested the response.

Serialization is often used in distributed systems and network programming, where data needs to be transmitted between different systems or processes. The serialized data can be transmitted in a platform-independent format, allowing systems with different architectures or programming languages to communicate with each other.

## Encoding vs Decoding

Encoding and decoding are two related but opposite operations that are commonly used in computer science.

Encoding is the process of converting data from one format to another. For example, encoding a string as a sequence of bytes, or encoding a data structure as a JSON string.

Decoding is the opposite process of encoding. It is the process of converting data from one format back to its original format. For example, decoding a sequence of bytes back into a string, or decoding a JSON string back into a data structure.

In general, encoding and decoding are used to represent data in a way that is more suitable for a particular purpose. For example, encoding data as a JSON string makes it easier to transmit over a network, while decoding the JSON string back into a data structure makes it easier to work with in a program.

## Cookie
GitHub Copilot: HTTP cookies are small pieces of data that are sent from a web server to a client (usually a web browser) and stored on the client's computer. Cookies are used to maintain state between requests in a stateless protocol like HTTP.

When a client sends an HTTP request to a server, it can include any cookies that were previously set by the server for that domain. The server can then use the cookies to identify the client and maintain state between requests.

Cookies can have various attributes, such as a name, a value, an expiration time, a domain, a path, and a secure flag. The name and value are used to store data, while the other attributes control how the cookie is transmitted and stored.

Cookies can be used for various purposes, such as session management, user tracking, and personalization. However, cookies can also be used for tracking and advertising purposes, which has led to concerns about privacy and security.

Web browsers typically allow users to view and delete cookies, and some browsers also allow users to block cookies entirely or only accept cookies from certain domains.

### Response Example

GitHub Copilot: Here's an example of a raw HTTP response with a cookie:

```
HTTP/1.1 200 OK
Content-Type: text/plain
Set-Cookie: mycookie=value; Path=/; Domain=example.com; Expires=Wed, 21 Oct 2021 07:28:00 GMT; Max-Age=3600; Secure
Content-Length: 13

Hello, world!
```

In this example, the response has a status line "HTTP/1.1 200 OK". The `Content-Type` header is set to "text/plain", indicating that the response body is plain text. The `Set-Cookie` header sets a cookie with a name "mycookie", a value "value", and various attributes such as `Path`, `Domain`, `Expires`, `Max-Age`, and `Secure`. The `Content-Length` header is set to 13, indicating that the response body has 13 bytes.

The response body is "Hello, world!", which is 13 bytes long and matches the `Content-Length` header.

### Request Example

GitHub Copilot: Here's an example of a raw HTTP request with a cookie:

```
GET / HTTP/1.1
Host: example.com
Cookie: mycookie=value; othercookie=othervalue
```

In this example, the request is a GET request for the root path ("/") of the "example.com" domain. The `Host` header is set to "example.com", indicating the domain of the request. The `Cookie` header sets two cookies: "mycookie" with a value "value" and "othercookie" with a value "othervalue".

Note that the `Cookie` header can contain multiple cookies separated by semicolons. Each cookie is a name-value pair separated by an equals sign. The name and value are both URL-encoded.

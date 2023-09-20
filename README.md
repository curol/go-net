# go-net

go-net is an golang net playground.

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
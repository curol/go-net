# http

This package is experimental for research & development in HTTP networking. 

## HTTP

HTTP stands for Hypertext Transfer Protocol. It is a protocol used for transmitting hypertext requests and information between servers and browsers. HTTP is the foundation of data communication on the World Wide Web.

Here are some key points about HTTP:

- Stateless: Each request from client to server is processed independently, without any knowledge of the requests that came before it.

- Connectionless: After a request is made, the client disconnects from the server and waits for a response. The server processes the request and re-establishes the connection with the client to send the response back.

- Media Independent: Any type of data can be sent by HTTP as long as both the client and the server know how to handle the data content.

- Methods: HTTP uses methods (also known as verbs) to indicate the desired action to be performed on the identified resource. The most common methods include GET, POST, PUT, DELETE, and HEAD.

In the context of programming, HTTP is often used for API calls, web scraping, and other tasks that involve sending or receiving data over the internet. In Go, the net/http package provides functionalities for HTTP requests and responses.

#### HTTP Message Format

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

#### HTTP Sessions

An HTTP session is a sequence of network request-response transactions. An HTTP client initiates a request. It then waits for the response. The HTTP server processes the request and sends a response back to the client. This exchange is considered one HTTP session.

However, HTTP itself is stateless, meaning each request is processed independently without any knowledge of the requests that came before it. To maintain state between different requests (such as information about whether a user is logged in), many web applications use session cookies, which are sent in the HTTP headers.

A session cookie might contain a session ID that the server can use to retrieve stored information about the client, such as user preferences, shopping cart contents, or authentication status. This allows the server to provide a personalized experience for each client.

In Go, you can use the net/http package's SetCookie function to set session cookies, and the Request type's Cookie method to retrieve them. There are also third-party packages available that provide more advanced session management features.

#### HTTP Cookie

GitHub Copilot: HTTP cookies are small pieces of data that are sent from a web server to a client (usually a web browser) and stored on the client's computer. Cookies are used to maintain state between requests in a stateless protocol like HTTP.

When a client sends an HTTP request to a server, it can include any cookies that were previously set by the server for that domain. The server can then use the cookies to identify the client and maintain state between requests.

Cookies can have various attributes, such as a name, a value, an expiration time, a domain, a path, and a secure flag. The name and value are used to store data, while the other attributes control how the cookie is transmitted and stored.

Cookies can be used for various purposes, such as session management, user tracking, and personalization. However, cookies can also be used for tracking and advertising purposes, which has led to concerns about privacy and security.

Web browsers typically allow users to view and delete cookies, and some browsers also allow users to block cookies entirely or only accept cookies from certain domains.

### Response

Here's an example of a raw HTTP response with a cookie:

```
HTTP/1.1 200 OK
Content-Type: text/plain
Set-Cookie: mycookie=value; Path=/; Domain=example.com; Expires=Wed, 21 Oct 2021 07:28:00 GMT; Max-Age=3600; Secure
Content-Length: 13

Hello, world!
```

In this example, the response has a status line "HTTP/1.1 200 OK". The `Content-Type` header is set to "text/plain", indicating that the response body is plain text. The `Set-Cookie` header sets a cookie with a name "mycookie", a value "value", and various attributes such as `Path`, `Domain`, `Expires`, `Max-Age`, and `Secure`. The `Content-Length` header is set to 13, indicating that the response body has 13 bytes.

The response body is "Hello, world!", which is 13 bytes long and matches the `Content-Length` header.

### Request 

Here's an example of a raw HTTP request with a cookie:

```
GET / HTTP/1.1
Host: example.com
Cookie: mycookie=value; othercookie=othervalue
```

In this example, the request is a GET request for the root path ("/") of the "example.com" domain. The `Host` header is set to "example.com", indicating the domain of the request. The `Cookie` header sets two cookies: "mycookie" with a value "value" and "othercookie" with a value "othervalue".

Note that the `Cookie` header can contain multiple cookies separated by semicolons. Each cookie is a name-value pair separated by an equals sign. The name and value are both URL-encoded.



## Context

In Go, the `context` package is often used with the `net/http` package to handle cancellation of HTTP requests. When an HTTP request is received by a Go server, it is handled in its own goroutine. If the client disconnects before the request is finished, the context associated with the request is cancelled.

Here's an example of how to use a context with an HTTP handler:

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fmt.Println("handler started")
	defer fmt.Println("handler ended")

	select {
	case <-time.After(5 * time.Second):
		fmt.Fprintln(w, "request processed")
	case <-ctx.Done():
		err := ctx.Err()
		fmt.Println("request cancelled:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
```

In this example, `handler` is an HTTP handler that simulates a long-running operation. It completes after 5 seconds, or cancels if the request is cancelled, whichever happens first. The context for the request is retrieved with `r.Context()`. If the client disconnects before the request is finished, `ctx.Done()` will receive a value and the request will be cancelled.

To test this, you can start the server, make a request, and then disconnect before the request is finished. You will see that the handler detects the disconnection and cancels the request.

### Go Context

In Go, the `context` package is used to pass cancellation signals, deadlines, and other request-scoped values across API boundaries and between processes.

Here are some key points about the `context` package:

1. **Cancellation**: You can cancel a context, which will cancel all contexts derived from it. This is useful to cancel a group of goroutines when any of them encounters an error, or when the operation they are performing is no longer needed.

2. **Deadlines and timeouts**: You can set a deadline or a timeout on a context. When the deadline is exceeded or the timeout is reached, the context is automatically cancelled. This is useful to prevent operations from running longer than a certain time.

3. **Request-scoped values**: You can associate values with a context. These values are scoped to the context and its descendants, and can be retrieved anywhere the context is passed. This is useful to pass request-scoped values such as request IDs, user IDs, and so on.

Here's an example of how to use a context:

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func operation(ctx context.Context) {
	select {
	case <-time.After(2 * time.Second):
		fmt.Println("operation completed")
	case <-ctx.Done():
		fmt.Println("operation cancelled")
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go operation(ctx)

	time.Sleep(3 * time.Second)
}
```

In this example, `operation` is a function that simulates a long-running operation. It completes after 2 seconds, or cancels if the context is done, whichever happens first. In `main`, a context with a timeout of 1 second is created. This context is passed to `operation`, which is run in a goroutine. After 3 seconds, the program ends. Because the context timeout is less than the time it takes for `operation` to complete, `operation` is cancelled before it completes.


## Write

For output, write serializes the request and writes to a writer.

### Request.Write

To serialize an HTTP request in Go, you can use the http.Request.Write method.
`http.Request.Write` writes the HTTP request in wire format, which includes the request line, headers, and body.
The `Host` header is automatically added based on the request's URL.

Here's an example of how to use the `http.Request.Write` method:

```go
package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	// Create a new request
	req, err := http.NewRequest("GET", "http://example.com", strings.NewReader("request body"))
	if err != nil {
		fmt.Println(err)
		return
	}

	// Add headers to the request
	req.Header.Add("Content-Type", "application/json")

	// Create a buffer to write the request to
	var buffer bytes.Buffer

	// Write the request to the buffer




	err = req.Write(&buffer)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print the request
	fmt.Println(buffer.String())
}
```

In this example, a new `http.Request` is created with the `http.NewRequest` function. The `http.Request.Write` method is then used to write the request to a `bytes.Buffer`. The contents of the buffer (which now contain the serialized HTTP request) are then printed to the console.



### fmt.Fprintf


`fmt.Fprintf` writes to the provided `io.Writer` and if that writer is buffered (like `bufio.Writer`), then the output of `fmt.Fprintf` will be buffered automatically.

Example:

```go
fmt.Fprintf(w, "%s %s %s\r\n", r.Method, r.URL.RequestURI(), r.Proto)
```

`fmt.Fprintf` formats according to a format specifier and writes to `w`. If `w` is a `bufio.Writer` or any other type of buffered writer, the output will be stored in the buffer until the buffer is full or manually flushed. This can improve performance by reducing the number of system calls.


> Remember, the data will not be written out until the buffer is full or you manually call the `Flush` method on the `bufio.Writer`. So, if you want to ensure that your data is written out immediately after the `Fprintf` call, you should follow it with a call to `w.Flush()`.

### Header.Write

Convert an http.Header to bytes by using the Write method of http.Header. 
This method writes the header in wire format (as it would appear in an HTTP request or response) to a bytes.Buffer, which you can then convert to bytes.

```go
header := http.Header{
    "Content-Type": []string{"application/json"},
    // Add other headers here
}

var buf bytes.Buffer
// Headers
err := header.Write(&buf)
if err != nil {
    log.Fatal(err)
}
headerBytes := buf.Bytes()

// Write trailing CRLF for end of head
buf.WriteString("\r\n")
headerBytes := buf.Bytes()
```

### Transfer encoding chunked

"Transfer-Encoding: chunked" is a type of HTTP/1.1 message body encoding, also known as chunked transfer encoding. It allows a server to maintain an HTTP connection for dynamically generated content or long-lived connections.

In chunked transfer encoding, the data is divided into a series of chunks. Each chunk is preceded by its size in bytes. The transmission ends when a zero-length chunk is received. This allows the server to start sending data as it is generated without knowing the total size of the data in advance.

Here's a simple example of what chunked transfer encoding might look like:

```
HTTP/1.1 200 OK 
Content-Type: text/plain 
Transfer-Encoding: chunked

7\r\n
Hello, \r\n
6\r\n
world!\r\n
0\r\n 
\r\n
```

In this example, the data "Hello, world!" is sent as two chunks. The first chunk has a size of 7 bytes, and the second chunk has a size of 6 bytes. The end of the message is marked by a chunk with a size of 0 bytes.

Chunked transfer encoding is particularly useful when the server doesn't know the size of the response when it starts transmitting, like when it's generated dynamically or comes from a process that doesn't provide its length in advance.
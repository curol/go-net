# http

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
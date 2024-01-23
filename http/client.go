package http

import (
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/curol/network/url"
)

type Client struct {
	network  string
	protocol string
	method   string
	address  string
	url      *url.URL
	header   map[string][]string
	body     io.Reader

	conn net.Conn  // connection to server
	req  *Request  // request
	res  *Response // response
}

func NewClient(method string, address string, header map[string][]string, body io.Reader) *Client {
	method = strings.ToUpper(strings.TrimSpace(method))
	address = strings.TrimSpace(address)
	u, err := url.Parse(address)
	if err != nil {
		panic(err)
	}

	client := &Client{
		network:  "tcp",
		protocol: "HTTP/1.1",
		method:   method,
		address:  address,
		header:   header,
		body:     body,
		url:      u,
	}
	return client

}

// Connect connects to the server.
//
// Example:
// client.connect()      // 1
// defer client.clean()  // 2
// client.writeRequest() // 3
// client.readResponse() // 4
// return client
func (c *Client) connect() {
	conn, err := net.Dial(c.network, c.address) // start connection
	if err != nil {
		panic(err)
	}
	c.conn = conn
}

// WriteRequest writes the request to the server.
func (c *Client) writeRequest() {
	req, err := NewRequest(c.method, c.address, c.header, io.NopCloser(c.body))
	if err != nil {
		panic(err)
	}

	err = req.Write(c.conn) // write request
	if err != nil {
		if err != io.EOF {
			panic(err)
		}
	}
	c.req = req
}

// ReadResponse reads the response from the server.
func (c *Client) readResponse() {
	resp, err := ReadResponse(c.conn) // read response
	if err != nil {
		if err != io.EOF {
			panic(err)
		}
	}
	c.res = resp
}

func (c *Client) readN(n int, conn net.Conn) (buf []byte, err error) {
	b := make([]byte, n)
	n, err = conn.Read(b)
	if err != nil {
		if err != io.EOF {
			panic(err)
		}
	}
	return b, nil
}

// Clean closes the connection to the server and cleans up client.
func (c *Client) clean() error {
	if c.conn == nil {
		return fmt.Errorf("Connection is nil")
	}
	return c.conn.Close()
}

//**********************************************************************************************************************
// Helpers
//**********************************************************************************************************************

// // ParseURL parses a raw url into a URL structure.
// func parseAddress(rawUrl string) (*url.URL, error) {
// 	parsedURL, err := url.Parse(rawUrl) // parse url
// 	if err != nil {
// 		return nil, fmt.Errorf("Error parsing url %s: %s", rawUrl, err)
// 	}
// 	if parsedURL.Path == "" {
// 		parsedURL.Path = "/"
// 	}
// 	return parsedURL, nil
// }

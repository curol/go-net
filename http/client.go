package http

import (
	"fmt"
	"io"
	"net"

	"github.com/curol/network/url"
)

// ClientRequest represents a request from the user.
type ClientConfig struct {
	Method  string
	Address string
	Header  map[string][]string
	Body    io.Reader
}

type config struct {
	Method string
	URL    *url.URL
	Header Header
	Body   io.Reader
}

func newConfig(cr *ClientConfig) *config {
	url, err := parseAddress(cr.Address)
	if err != nil {
		panic(err)
	}
	return &config{
		Method: cr.Method,
		URL:    url,
		Header: newHeaderFromMap(cr.Header),
		Body:   cr.Body,
	}
}

type Client struct {
	network  string
	protocol string
	*config
	conn net.Conn  // connection to server
	reqN int64     // number of bytes written
	resN int64     // number of bytes read
	req  *Request  // request
	res  *Response // response
}

func NewClient(config *ClientConfig) *Client {
	client := &Client{
		network:  "tcp",
		protocol: "HTTP/1.1",
		config:   newConfig(config),
	}
	client.connect()      // 1
	defer client.clean()  // 2
	client.writeRequest() // 3
	client.readResponse() // 4
	return client
}

// Connect connects to the server.
func (c *Client) connect() {
	conn, err := net.Dial(c.network, c.URL.Host) // start connection
	if err != nil {
		panic(err)
	}
	c.conn = conn
}

// WriteRequest writes the request to the server.
func (c *Client) writeRequest() {
	addr := c.URL.String() // TODO: fix this
	// addr := c.URL.Host

	req, err := NewRequest(c.Method, addr, c.Header, io.NopCloser(c.Body))
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
	// c.reqN = n
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
	return c.conn.Close()
}

//**********************************************************************************************************************
// Helpers
//**********************************************************************************************************************

// ParseURL parses a raw url into a URL structure.
func parseAddress(rawUrl string) (*url.URL, error) {
	parsedURL, err := url.Parse(rawUrl) // parse url
	if err != nil {
		return nil, fmt.Errorf("Error parsing url %s: %s", rawUrl, err)
	}
	if parsedURL.Path == "" {
		parsedURL.Path = "/"
	}
	return parsedURL, nil
}

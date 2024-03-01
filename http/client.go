package http

import (
	"bufio"
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
	// Set request line
	method = strings.ToUpper(strings.TrimSpace(method))
	address = strings.TrimSpace(address)
	u, err := url.Parse(address)
	if err != nil {
		panic(err)
	}
	// Create client
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

func Get(address string, header map[string][]string, body io.Reader) []byte {
	c := NewClient("GET", address, header, body)
	defer c.Clean()
	// Create request
	req, err := NewRequest(c.method, c.address, c.header, body)
	if err != nil {
		panic(err)
	}
	// Write request
	w := bufio.NewWriter(c.conn)
	err = req.Write(w)
	if err != nil {
		if err != io.EOF {
			panic(err)
		}
	}
	// Read response
	r := bufio.NewReader(c.conn)
	if r == nil {
		return nil
	}
	fl, _ := r.ReadString('\n')
	// Finish
	return []byte(fl)
}

func (c *Client) Parse(r *bufio.Reader) {

}

func (c *Client) Do() *Response {
	// 1. Connect
	c.dial()

	// 2. Clean up
	defer c.Clean()

	// 3. Write request
	req, err := NewRequest(c.method, c.address, c.header, io.NopCloser(c.body))
	if err != nil {
		panic(err)
	}
	err = req.Write(c.conn)
	if err != nil {
		if err != io.EOF {
			panic(err)
		}
	}
	c.req = req

	// 4. Read response
	// io.Copy(c.conn, c.res.Body)
	resp, err := ReadResponse(c.conn)
	if err != nil {
		if err != io.EOF {
			panic(err)
		}
	}
	c.res = resp
	return resp
}

// Connect connects to the server.
func (c *Client) dial() {
	conn, err := net.Dial(c.network, c.address) // start connection
	if err != nil {
		panic(err)
	}
	c.conn = conn
}

// WriteRequest writes the request to the server.
func (c *Client) write() error {
	return c.req.Write(c.conn)
}

// ReadResponse reads the response from the server.
func (c *Client) read() (int64, error) {
	return c.res.WriteTo(c.conn)
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
func (c *Client) Clean() error {
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

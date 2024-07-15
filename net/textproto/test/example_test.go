package textproto_test

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/curol/network/net/testutil"
	"github.com/curol/network/net/textproto"
)

func exampleHandlerFunc(conn net.Conn) {
	defer conn.Close()

	// Read the request
	tm, err := textproto.ReadTextMessage(bufio.NewReader(conn))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Server Request:", tm)
	b, err := tm.Bytes()
	if err != nil {
		fmt.Println("error to bytes", err)
		return
	}
	fmt.Println("Server Request size:", len(b))
	fmt.Println("Server Request bytes:", b)
	fmt.Println("Server Request string:")
	fmt.Println("--------Start-----------")
	fmt.Print(string(b))
	fmt.Println("\n---------Finished----------")

	// Write response
	rawRes := []byte("HTTP/1.0 200 OK\r\nUser-Agent: textproto example\r\nType: Response\r\nAccept: */*\r\n\r\n")
	fmt.Fprint(conn, string(rawRes))
}

func printRawRequest(rawReq []byte) {
	fmt.Println("---")
	fmt.Println("- Raw request size:", len(rawReq))
	fmt.Println("- Raw request bytes", rawReq)
	fmt.Println("---")
}

func TestExampleTextMessage(t *testing.T) {
	// 1. Server
	go testutil.MockServer(exampleHandlerFunc)
	time.Sleep(2 * time.Second)

	// 2. Client
	rawReq := []byte("GET / HTTP/1.0\r\nUser-Agent: textproto example\r\nAccept: */*\r\n\r\n")
	printRawRequest(rawReq)

	testutil.MockClientReq(rawReq)
}

func TestExampleTextMessageWithBody1(t *testing.T) {
	// 1. Server
	go testutil.MockServer(exampleHandlerFunc)
	time.Sleep(2 * time.Second)

	// 2. Client request
	lines := []string{
		"GET / HTTP/1.0",
		"User-Agent: textproto example",
		"Accept: */*",
		"Content-Length: 5",
		"",      // end of headers
		"Hello", // body
	}
	rawReq := []byte(strings.Join(lines, "\r\n"))

	printRawRequest(rawReq)

	testutil.MockClientReq(rawReq)
}

func TestExampleTextMessageWithBody2(t *testing.T) {
	// 1. Server
	go testutil.MockServer(exampleHandlerFunc)
	time.Sleep(2 * time.Second)

	// 2. Client request
	lines := []string{
		"GET / HTTP/1.0",
		"User-Agent: textproto example",
		"Accept: */*",
		"Content-Length: 12",
		"",             // end of headers
		"Hello World!", // body
	}
	rawReq := []byte(strings.Join(lines, "\r\n"))

	printRawRequest(rawReq)

	testutil.MockClientReq(rawReq)
}

func TestExampleTextMessageWithBodyAndContentLengthShorter(t *testing.T) {
	// 1. Server
	go testutil.MockServer(exampleHandlerFunc)
	time.Sleep(2 * time.Second)
	// 2. Client
	lines := []string{
		"GET / HTTP/1.0",
		"User-Agent: textproto example",
		"Accept: */*",
		"Content-Length: 5", // content length header shorter than actual contents
		"",                  // end of headers
		"Hello World",
	}
	rawReq := []byte(strings.Join(lines, "\r\n"))

	printRawRequest(rawReq)

	testutil.MockClientReq(rawReq)
}

func TestExampleTextMessageWithContentLengthAndNoBody(t *testing.T) {
	// 1. Server
	go testutil.MockServer(exampleHandlerFunc)
	time.Sleep(2 * time.Second)
	// 2. Client
	lines := []string{
		"GET / HTTP/1.0",
		"User-Agent: textproto example",
		"Accept: */*",
		"Content-Length: 5", // content length header shorter than actual contents
		"",                  // end of headers
		"",                  // no body
	}
	rawReq := []byte(strings.Join(lines, "\r\n"))

	printRawRequest(rawReq)

	testutil.MockClientReq(rawReq)
}

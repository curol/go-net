package testutil

import (
	"bufio"
	"fmt"
	"net"

	"github.com/curol/network"
	"github.com/curol/network/net/textproto"
)

func MockServer(handlerFunc func(conn net.Conn)) {
	// Listen for connections
	ln, err := network.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	// When server finished
	defer ln.Close() // close listener when finished
	defer fmt.Println("Server Finished.")

	fmt.Println("Server started.")

	// Without continious loop, accept single connection
	conn, err := ln.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Connection accepted from:", conn.RemoteAddr().String())

	// Handle the connection
	handlerFunc(conn)
}

func MockClientReq(rawReq []byte) {
	// Connect to the server
	conn, err := network.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Write request
	fmt.Fprint(conn, string(rawReq))

	// Read the response
	tm, err := textproto.ReadTextMessage(bufio.NewReader(conn))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Response:", tm)
}

func MockStdoutHandlerFunc(conn net.Conn) {
	// 1. Close connection when done
	defer conn.Close()
	// 2. Read the request text message
	tm, err := textproto.ReadTextMessage(bufio.NewReader(conn))
	if err != nil {
		fmt.Println(err)
		return
	}
	// 3. Handle the request
	fmt.Println("\n------")
	fmt.Println("\n\n\nHandler Request:\n", tm)
	fmt.Println("- Content Length:", tm.ContentLength())
	fmt.Println("- Content Type:", tm.ContentType())
	fmt.Println("- Size:", tm.Size())
	tm.StdOut()
	fmt.Println("\n------")
	// 4. Write the response
	rawRes := []byte("HTTP/1.0 200 OK\r\nUser-Agent: textproto example\r\nType: Response\r\n\r\n")
	fmt.Fprint(conn, string(rawRes))
}

func MockFilehandlerFunc(conn net.Conn) {
	//
	defer conn.Close()

	// Read the request
	tm, err := textproto.ReadTextMessage(bufio.NewReader(conn))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Request:", tm)

	// Cant use this cause body would already be read
	// fmt.Println("Request string:", string(tm.Bytes()))

	tm.Content("example-test-go.txt")

	// Write response
	rawRes := []byte("HTTP/1.0 200 OK\r\nUser-Agent: textproto example\r\nType: Response\r\nAccept: */*\r\n\r\n")
	fmt.Fprint(conn, string(rawRes))
}

// 	// Write req
// 	fmt.Fprintf(conn, string(rawReq))

// 	// Read the request
// 	req := textproto.NewRequest(bufio.NewReader(conn))
// 	err = req.Read()
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	fmt.Println("Request:", req)

// 	// Write the response
// 	res := textproto.NewResponse(bufio.NewWriter(conn))
// 	// res.StatusLine = "HTTP/1.0 200 OK"
// 	// res.Headers = map[string][]string{
// 	// 	"User-Agent": {"textproto example"},
// 	// 	"Accept":     {"*/*"},
// 	// }
// 	rawData := []byte("HTTP/1.0 200 OK\r\nUser-Agent: textproto example\r\nAccept: */*\r\n\r\n")
// 	_, err = res.Write()
// 	fmt.Println("Response:", res)
// }

// exampleTextMessage()

/*
Package network provides a portable interface for network I/O, including
TCP/IP, UDP, domain name resolution, and Unix domain sockets.

Although the package provides access to low-level networking
primitives, most clients will need only the basic interface provided
by the Dial, Listen, and Accept functions and the associated
Conn and Listener interfaces. The crypto/tls package uses
the same interfaces and similar Dial and Listen functions.

The Dial function connects to a server:

	conn, err := network.Dial("tcp", "golang.org:80")
	if err != nil {
		// handle error
	}
	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	status, err := bufio.NewReader(conn).ReadString('\n')
	// ...

The Listen function creates servers:

	ln, err := network.Listen("tcp", ":8080")
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}

# Name Resolution

The method for resolving domain names, whether indirectly with functions like Dial
or directly with functions like LookupHost and LookupAddr, varies by operating system.

On Unix systems, the resolver has two options for resolving names.
It can use a pure Go resolver that sends DNS requests directly to the servers
listed in /etc/resolv.conf, or it can use a cgo-based resolver that calls C
library routines such as getaddrinfo and getnameinfo.

By default the pure Go resolver is used, because a blocked DNS request consumes
only a goroutine, while a blocked C call consumes an operating system thread.
When cgo is available, the cgo-based resolver is used instead under a variety of
conditions: on systems that do not let programs make direct DNS requests (OS X),
when the LOCALDOMAIN environment variable is present (even if empty),
when the RES_OPTIONS or HOSTALIASES environment variable is non-empty,
when the ASR_CONFIG environment variable is non-empty (OpenBSD only),
when /etc/resolv.conf or /etc/nsswitch.conf specify the use of features that the
Go resolver does not implement, and when the name being looked up ends in .local
or is an mDNS name.

The resolver decision can be overridden by setting the netdns value of the
GODEBUG environment variable (see package runtime) to go or cgo, as in:

	export GODEBUG=netdns=go    # force pure Go resolver
	export GODEBUG=netdns=cgo   # force native resolver (cgo, win32)

The decision can also be forced while building the Go source tree
by setting the netgo or netcgo build tag.

A numeric netdns setting, as in GODEBUG=netdns=1, causes the resolver
to print debugging information about its decisions.
To force a particular resolver while also printing debugging information,
join the two settings by a plus sign, as in GODEBUG=netdns=go+1.

On macOS, if Go code that uses the net package is built with
-buildmode=c-archive, linking the resulting archive into a C program
requires passing -lresolv when linking the C code.

On Plan 9, the resolver always accesses /net/cs and /net/dns.

On Windows, in Go 1.18.x and earlier, the resolver always used C
library functions, such as GetAddrInfo and DnsQuery.
*/

// TODO: Add more functions here.

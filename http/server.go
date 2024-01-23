package http

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/curol/network/url"
)

// A ResponseWriter interface is used by an HTTP handler to
// construct an HTTP response.
//
// A ResponseWriter may not be used after [Handler.ServeHTTP] has returned.
type ResponseWriter interface {
	// Header returns the header map that will be sent by
	// [ResponseWriter.WriteHeader]. The [Header] map also is the mechanism with which
	// [Handler] implementations can set HTTP trailers.
	//
	// Changing the header map after a call to [ResponseWriter.WriteHeader] (or
	// [ResponseWriter.Write]) has no effect unless the HTTP status code was of the
	// 1xx class or the modified headers are trailers.
	//
	// There are two ways to set Trailers. The preferred way is to
	// predeclare in the headers which trailers you will later
	// send by setting the "Trailer" header to the names of the
	// trailer keys which will come later. In this case, those
	// keys of the Header map are treated as if they were
	// trailers. See the example. The second way, for trailer
	// keys not known to the [Handler] until after the first [ResponseWriter.Write],
	// is to prefix the [Header] map keys with the [TrailerPrefix]
	// constant value.
	//
	// To suppress automatic response headers (such as "Date"), set
	// their value to nil.
	Header() Header

	// Write writes the data to the connection as part of an HTTP reply.
	//
	// If [ResponseWriter.WriteHeader] has not yet been called, Write calls
	// WriteHeader(http.StatusOK) before writing the data. If the Header
	// does not contain a Content-Type line, Write adds a Content-Type set
	// to the result of passing the initial 512 bytes of written data to
	// [DetectContentType]. Additionally, if the total size of all written
	// data is under a few KB and there are no Flush calls, the
	// Content-Length header is added automatically.
	//
	// Depending on the HTTP protocol version and the client, calling
	// Write or WriteHeader may prevent future reads on the
	// Request.Body. For HTTP/1.x requests, handlers should read any
	// needed request body data before writing the response. Once the
	// headers have been flushed (due to either an explicit Flusher.Flush
	// call or writing enough data to trigger a flush), the request body
	// may be unavailable. For HTTP/2 requests, the Go HTTP server permits
	// handlers to continue to read the request body while concurrently
	// writing the response. However, such behavior may not be supported
	// by all HTTP/2 clients. Handlers should read before writing if
	// possible to maximize compatibility.
	Write([]byte) (int, error)

	// WriteHeader sends an HTTP response header with the provided
	// status code.
	//
	// If WriteHeader is not called explicitly, the first call to Write
	// will trigger an implicit WriteHeader(http.StatusOK).
	// Thus explicit calls to WriteHeader are mainly used to
	// send error codes or 1xx informational responses.
	//
	// The provided code must be a valid HTTP 1xx-5xx status code.
	// Any number of 1xx headers may be written, followed by at most
	// one 2xx-5xx header. 1xx headers are sent immediately, but 2xx-5xx
	// headers may be buffered. Use the Flusher interface to send
	// buffered data. The header map is cleared when 2xx-5xx headers are
	// sent, but not with 1xx headers.
	//
	// The server will automatically send a 100 (Continue) header
	// on the first read from the request body if the request has
	// an "Expect: 100-continue" header.
	WriteHeader(statusCode int)
}

// A Handler responds to an HTTP request.
//
// [Handler.ServeHTTP] should write reply headers and data to the [ResponseWriter]
// and then return. Returning signals that the request is finished; it
// is not valid to use the [ResponseWriter] or read from the
// [Request.Body] after or concurrently with the completion of the
// ServeHTTP call.
//
// Depending on the HTTP client software, HTTP protocol version, and
// any intermediaries between the client and the Go server, it may not
// be possible to read from the [Request.Body] after writing to the
// [ResponseWriter]. Cautious handlers should read the [Request.Body]
// first, and then reply.
//
// Except for reading the body, handlers should not modify the
// provided Request.
//
// If ServeHTTP panics, the server (the caller of ServeHTTP) assumes
// that the effect of the panic was isolated to the active request.
// It recovers the panic, logs a stack trace to the server error log,
// and either closes the network connection or sends an HTTP/2
// RST_STREAM, depending on the HTTP protocol. To abort a handler so
// the client sees an interrupted response but the server doesn't log
// an error, panic with the value [ErrAbortHandler].
type Handler interface {
	// ServeConn should write reply headers and data to the [ResponseWriter]
	// and then return. Returning signals that the request is finished; it
	// is not valid to use the [ResponseWriter] or read from the
	// [Request.Body] after or concurrently with the completion of the
	// ServeHTTP call.
	ServeConn(ResponseWriter, *Request)
}

// ListenAndServe listens on the TCP network address addr and then calls
// [Serve] with handler to handle requests on incoming connections.
// Accepted connections are configured to enable TCP keep-alives.
//
// The handler is typically nil, in which case [DefaultServeMux] is used.
//
// ListenAndServe always returns a non-nil error.
func ListenAndServe(addr string, handler Handler) error {
	server := &Server{Address: addr, Handler: handler}
	return server.listenAndServe()
}

// Serve starts the server and listens for connections.
// Listen listens on the TCP network and accepts incoming connections concurrently.
// The handler handles a request and response for the client.
func (s *Server) listenAndServe() error {
	network := s.Network
	address := s.Address

	// Listen for connections
	listener, err := net.Listen(network, address)
	if err != nil {
		return err
	} else {
		s.listener = listener
		fmt.Println("Server listening on " + address)
	}

	// Clean
	defer func() error {
		err := s.listener.Close() // close listener
		if err != nil {
			return err
		}
		fmt.Println("Server closed")
		return nil
	}()

	// Accept connections and serve
	for {
		conn, err := listener.Accept() // wait for next connection and accept
		if err != nil {
			s.Logger.Fatal(err)
			continue // skip to next connection
		}
		go s.serve(conn) // serve and handle connection
	}
}

func (s *Server) serve(conn net.Conn) {
	// TODO: What to do after serving connection?
	// TODO: Close response body?
	// TODO: After handler finishes, serialize response?
	// TODO: After handler finishes, flush ResponseWriter?
	// TODO: Serialize response?
	// Write response
	// _, err := res.Write() // write response output to its writer `w`
	// if err != nil {
	// 	panic(err)
	// }

	// Clean
	defer func() {
		err := conn.Close() // close connection
		if err != nil {
			panic(err)
		}
	}()

	// Init
	err := conn.SetDeadline(s.Deadline) // set deadlines
	if err != nil {
		panic(err)
	}

	// Request
	req, err := ReadRequest(bufio.NewReader(conn)) // read request
	if err != nil {
		// TODO: Handle error
		s.Logger.Fatal(err)
	}

	// Log
	s.Logger.Status(conn.RemoteAddr().String(), req.Method, req.RequestURI)

	// Response
	res := NewResponse(conn)

	// Handler
	s.Handler.ServeConn(res, req)
}

// func Run(address string, config *ServerConfig) {
// 	// Server
// 	server := NewServer(address, config)

// 	// Listen for connections and serve connection using config.handler
// 	server.Serve()
// }

type Server struct {
	// Connection
	Network  string
	Address  string
	Deadline time.Time
	// Misc
	Logger  Log
	Handler Handler
	// Connection
	listener net.Listener
}

func NewServer(address string, config *ServerConfig) *Server {
	server := &Server{ // default server
		Address:  address,
		Logger:   NewLogger(),
		Handler:  NewMux(),
		Network:  "tcp",
		Deadline: time.Now().Add(5 * time.Minute),
		listener: nil,
	}
	return server
}

type ServerConfig struct {
	*Server
}

var invalidRequestURIErr = fmt.Errorf("Invalid request URI")

//**********************************************************************************************************************
// Connection
//**********************************************************************************************************************

type parsedConnection struct {
	remoteAddress string
	localAddress  string
	url           *url.URL
	host          string
	hostname      string
	path          string
}

func parseConnection(conn net.Conn) (*parsedConnection, error) {
	pc := new(parsedConnection)

	pc.remoteAddress = conn.RemoteAddr().String()
	pc.localAddress = conn.LocalAddr().String()
	u, err := url.Parse(pc.remoteAddress)
	if err != nil {
		return nil, err
	}
	pc.url = u
	pc.hostname = u.Hostname()
	pc.host = u.Host
	return pc, nil
}

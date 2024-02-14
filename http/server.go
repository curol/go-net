package http

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// Server is a simple HTTP server and architure without all the extra services.
type Server struct {
	// Connection
	Network        string
	Address        string
	Deadline       time.Time
	Logger         Log
	Handler        Handler // handler to invoke, http.DefaultServeMux if nil
	Listener       net.Listener
	MaxHeaderBytes int
	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body. A zero or negative value means
	// there will be no timeout.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration

	// ReadHeaderTimeout is the amount of time allowed to read
	// request headers. The connection's read deadline is reset
	// after reading the headers and the Handler can decide what
	// is considered too slow for the body. If ReadHeaderTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, there is no timeout.
	ReadHeaderTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	// A zero or negative value means there will be no timeout.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled. If IdleTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, there is no timeout.
	IdleTimeout time.Duration

	isShutdown bool
}

func NewServer(network string, address string) *Server {
	if network == "" {
		network = "tcp"
	}
	server := &Server{ // default server
		Network:           network,
		Address:           address,
		Logger:            NewLogger(),
		Handler:           NewMux(),
		Deadline:          time.Now().Add(5 * time.Minute), // TODO: Set default deadlines
		Listener:          nil,
		MaxHeaderBytes:    1 << 20,         // 1 MB
		ReadTimeout:       5 * time.Minute, // Fixed: Use time.Duration value
		ReadHeaderTimeout: 5 * time.Minute, // Fixed: Use time.Duration value
		WriteTimeout:      5 * time.Minute, // Fixed: Use time.Duration value
		IdleTimeout:       5 * time.Minute, // Fixed: Use time.Duration value
		isShutdown:        false,
	}
	return server
}

// Run starts the server and listens for connections.
func (s *Server) Run() error {
	return s.listenAndServe()
}

// listenAndServe listens for connections and serves them.
// Each serve is a service goroutine that reads requests and then calls [Handler].
func (s *Server) listenAndServe() error {
	network := s.Network
	address := s.Address

	// 1. Listen for connections
	listener, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	s.Listener = listener
	s.Logger.Info("Server listening on " + address)

	// 2. Defer server shutdown
	defer s.Shutdown()

	// 3. Listen for new connections and serve
	for {
		// 3.1. Acceept next connection
		conn, err := listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				// The error "use of closed network connection" typically occurs when you're trying to perform a network operation (like Accept, Read, Write, etc.) on a network connection that has already been closed.
				// This can often happen in a server that's being shut down while it's in the middle of accepting new connections
				break
			} else {
				e := fmt.Errorf("Error on listener.Accept(): " + err.Error())
				s.Logger.Fatal(e)
				continue
			}
		}
		// 3.2. Serve connection
		go s.serve(conn)
	}

	// 4. Finish
	return nil
}

// serve serves a new connection and calls the [Handler].
// Moreover, serve handles each new connection, reads requests, and then calls [Handler] to reply to them.
func (s *Server) serve(conn net.Conn) {
	// 1. Defer closing connection
	defer func() {
		err := conn.Close() // close connection
		if err != nil {
			// TODO: Handle error
			panic("Error closing connection in 'Server.serve(conn)': " + err.Error())
		}
	}()

	// 2. Set connection properties
	err := conn.SetReadDeadline(time.Now().Add(s.ReadTimeout))
	if err != nil {
		panic(err)
	}
	err = conn.SetWriteDeadline(time.Now().Add(s.WriteTimeout))
	if err != nil {
		panic(err)
	}

	// 3. Read Request
	req, err := ReadRequest(bufio.NewReader(conn)) // read request
	if err != nil {
		// TODO: Handle error
		s.Logger.Fatal(err)
	}

	// 4. Log status
	s.Logger.Status(conn.RemoteAddr().String(), req.Method, req.RequestURI)

	// 5. Create response writer
	rw := newResponseWriter(conn, req)

	// 6. Serve handler
	s.Handler.ServeHTTP(rw, req)

	// 7. Write response
	_, err = rw.WriteTo(conn)
	if err != nil {
		s.Logger.Warn("Error writing response to connection: " + err.Error())
	}

	// 8. Finish
	// b := res.Bytes() // get response bytes
	// _, err = res.WriteTo(b) // write response output to its writer `w`
	// if err != nil {
	// 	panic(err)
	// }

	// TODO: Finish implementation
}

// Shutdown gracefully shutsdown the server resources and cleans up.
func (s *Server) Shutdown() error {
	// Cleanup server resources
	if s.isShutdown {
		return nil
	}
	time.Sleep(1 * time.Second) // wait for server to shutdown
	err := s.clean()
	if err != nil {
		return err
	}
	// TODO: Add more cleanup
	s.Logger.Info("Succesfully cleaned up server.")
	// TODO: Implement graceful shutdown
	s.Logger.Info("Successfuly shutdown server. Goodbye:)")
	s.isShutdown = true
	time.Sleep(2 * time.Second) // wait for server to shutdown
	return nil
}

func (s *Server) IsShutdown() bool {
	return s.isShutdown
}

// clean cleans up server resources.
func (s *Server) clean() error {
	if s.Listener == nil {
		return nil
	}
	err := s.Listener.Close() // close listener
	if err != nil {
		return err
	}
	return nil
}

// Error replies to the request with the specified error message and HTTP code.
// It does not otherwise end the request; the caller should ensure no further
// writes are done to w.
// The error message should be plain text.
func Error(w ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, error)
}

// NotFound replies to the request with an HTTP 404 not found error.
func NotFound(w ResponseWriter, r *Request) { Error(w, "404 page not found", StatusNotFound) }

// NotFoundHandler returns a simple request handler
// that replies to each request with a “404 page not found” reply.
func NotFoundHandler() Handler { return HandlerFunc(NotFound) }

// The Hijacker interface is implemented by ResponseWriters that allow
// an HTTP handler to take over the connection.
//
// The default [ResponseWriter] for HTTP/1.x connections supports
// Hijacker, but HTTP/2 connections intentionally do not.
// ResponseWriter wrappers may also not support Hijacker. Handlers
// should always test for this ability at runtime.
type Hijacker interface {
	// Hijack lets the caller take over the connection.
	// After a call to Hijack the HTTP server library
	// will not do anything else with the connection.
	//
	// It becomes the caller's responsibility to manage
	// and close the connection.
	//
	// The returned net.Conn may have read or write deadlines
	// already set, depending on the configuration of the
	// Server. It is the caller's responsibility to set
	// or clear those deadlines as needed.
	//
	// The returned bufio.Reader may contain unprocessed buffered
	// data from the client.
	//
	// After a call to Hijack, the original Request.Body must not
	// be used. The original Request's Context remains valid and
	// is not canceled until the Request's ServeHTTP method
	// returns.
	Hijack() (net.Conn, *bufio.ReadWriter, error)
}

func (s *Server) SetHandler(h Handler) {
	s.Handler = h
}

func (s *Server) HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
	if s.Handler == nil {
		s.Handler = NewMux()
	}

	switch s.Handler.(type) {
	case *Mux:
		s.Handler.(*Mux).HandleFunc(pattern, handler)
	default:
		//
		s.Logger.Warn("Server handler is not defined...")
		panic("server.HandleFunc type not found...")
	}

}

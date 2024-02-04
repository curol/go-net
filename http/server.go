package http

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/curol/network/url"
)

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

// // ListenAndServe listens on the TCP network address addr and then calls
// // [Serve] with handler to handle requests on incoming connections.
// // Accepted connections are configured to enable TCP keep-alives.
// //
// // The handler is typically nil, in which case [DefaultServeMux] is used.
// //
// // ListenAndServe always returns a non-nil error.
// func ListenAndServe(addr string, handler Handler) error {
// 	server := &Server{Address: addr, Handler: handler}
// 	return server.listenAndServe()
// }

type Server struct {
	// Connection
	Network  string
	Address  string
	Deadline time.Time
	// ErrorLog specifies an optional logger for errors accepting
	// connections, unexpected behavior from handlers, and
	// underlying FileSystem errors.
	// If nil, logging is done via the log package's standard logger.
	ErrorLog *log.Logger
	// Misc
	Logger Log

	Handler Handler // handler to invoke, http.DefaultServeMux if nil
	// Connection
	listener net.Listener

	// TLSConfig optionally provides a TLS configuration for use
	// by ServeTLS and ListenAndServeTLS. Note that this value is
	// cloned by ServeTLS and ListenAndServeTLS, so it's not
	// possible to modify the configuration with methods like
	// tls.Config.SetSessionTicketKeys. To use
	// SetSessionTicketKeys, use Server.Serve with a TLS Listener
	// instead.
	TLSConfig *tls.Config

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

	// MaxHeaderBytes controls the maximum number of bytes the
	// server will read parsing the request header's keys and
	// values, including the request line. It does not limit the
	// size of the request body.
	// If zero, DefaultMaxHeaderBytes is used.
	MaxHeaderBytes int
}

type ServerConfig struct {
	*Server
}

func NewServer(network string, address string) *Server {
	if network == "" {
		network = "tcp"
	}
	server := &Server{ // default server
		Network:  network,
		Address:  address,
		Logger:   NewLogger(),
		Handler:  NewMux(),
		Deadline: time.Now().Add(5 * time.Minute), // TODO: Set default deadlines
		listener: nil,
	}
	return server
}

func (s *Server) ListenAndServe() error {
	return s.listenAndServe()
}

// listenAndServe starts the server, listens for connections, and serves them concurrently.
// A handler is called for each connection.
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
	rw := newResponseWriter(conn)

	// Serve handler
	s.Handler.ServeHTTP(rw, req)
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

// TimeFormat is the time format to use when generating times in HTTP
// headers. It is like [time.RFC1123] but hard-codes GMT as the time
// zone. The time being formatted must be in UTC for Format to
// generate the correct format.
//
// For parsing this time format, see [ParseTime].
const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

// appendTime is a non-allocating version of []byte(t.UTC().Format(TimeFormat))
func appendTime(b []byte, t time.Time) []byte {
	const days = "SunMonTueWedThuFriSat"
	const months = "JanFebMarAprMayJunJulAugSepOctNovDec"

	t = t.UTC()
	yy, mm, dd := t.Date()
	hh, mn, ss := t.Clock()
	day := days[3*t.Weekday():]
	mon := months[3*(mm-1):]

	return append(b,
		day[0], day[1], day[2], ',', ' ',
		byte('0'+dd/10), byte('0'+dd%10), ' ',
		mon[0], mon[1], mon[2], ' ',
		byte('0'+yy/1000), byte('0'+(yy/100)%10), byte('0'+(yy/10)%10), byte('0'+yy%10), ' ',
		byte('0'+hh/10), byte('0'+hh%10), ':',
		byte('0'+mn/10), byte('0'+mn%10), ':',
		byte('0'+ss/10), byte('0'+ss%10), ' ',
		'G', 'M', 'T')
}

// Helper handlers

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

package http

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/curol/network/url"
)

func Run(address string) {
	// Config
	config := NewConfig(address)

	// Server
	server := NewServer(config)

	// Listen for connections and serve connection using config.handler
	server.listenAndServe()
}

type Server struct {
	// Config
	config *Config
	// Connection
	listener net.Listener
	// Misc
	log     Log
	handler Handler
}

func NewServer(config *Config) *Server {
	server := &Server{
		listener: nil,
		config:   config,
		// Misc
		log:     config.Log,
		handler: config.Handler,
	}
	return server
}

func (s *Server) listenAndServe() {
	// ListenAndServe listens on the TCP network and accepts incoming connections concurrently.
	// The handler handles a request and response for the client.

	network := s.config.Network
	address := s.config.Address

	// Listen for connections
	listener, err := net.Listen(network, address)
	if err != nil {
		panic(err)
	} else {
		s.listener = listener
		fmt.Println("Server listening on " + address)
	}

	// Clean up when finished
	defer s.clean()

	// Run listener forever
	for {
		conn, err := listener.Accept() // wait for next connection and accept
		if err != nil {
			log.Fatal(err) // log status
		}
		go s.serve(conn) // serve and handle connection
	}
}

func (s *Server) serve(conn net.Conn) {
	// Clean
	defer conn.Close() // close connection when this is finished

	// Init
	s.setConnectionProps(conn) // set connection properties

	// Request
	req, err := ReadRequest(bufio.NewReader(conn)) // read request
	// TODO: Handle error
	if err != nil {
		panic(err)
	}

	// Log
	s.log.Status(req.RequestURI, req.Method, conn.RemoteAddr().String())

	// Response
	res := new(Response)

	// Handler
	s.handler.ServeConn(res, req)

	// TODO: What to do after serving connection?
	// TODO: Close response body?
	// TODO: After handler finishes, serialize response?
	// TODO: After handler finishes, flush ResponseWriter?
	// Write response
	// _, err := res.Write() // write response output to its writer `w`
	// if err != nil {
	// 	panic(err)
	// }
}

func (s *Server) setConnectionProps(conn net.Conn) {
	// Set read and write deadlines
	err := conn.SetDeadline(s.config.Deadline)
	if err != nil {
		panic(err)
	}
}

// Clean cleans up the server when listener is stopped
func (s *Server) clean() {
	err := s.listener.Close() // close listener
	if err != nil {
		panic(err)
	}
	fmt.Println("Server closed")
}

type ResponseWriter interface {
	Write(b []byte) (int, error)
	WriteHeader(string, string)
	Header() Header
}

// Mux implements interface Handler for handling requests.
type Mux struct{}

// NewMux returns a new Mux.
func NewMux() *Mux {
	return &Mux{}
}

// ServeConn handles a request and response for the client.
func (m *Mux) ServeConn(w *Response, r *Request) {
	w.NotFound()
}

type Log interface {
	Status(path, method, remoteAddress string) // Log status
	Fatal(error)                               // Log error and exit
	// TODO: Add more logging methods
}

type logger struct{}

// NewLogger returns a new logger
func NewLogger() *logger {
	return &logger{}
}

// Log logs connection status
func (l *logger) Status(path, method, remoteAddress string) {
	// Time
	now := time.Now()
	timeFormat := now.Format("2006-01-02 15:04:05")
	// Connection
	// Request
	s := "%s Status: %s (path: %s) (method: %s)\n"
	fmt.Printf(s, timeFormat, remoteAddress, path, method)
}

// Fatal logs error and exits
func (l *logger) Fatal(err error) {
	log.Fatal(err)
}

// Handler is an interface with the method ServeConn(ResponseWriter, *Request) that handles and responds to an HTTP request.
type Handler interface {
	// ServeHTTP should write reply headers and data to the [ResponseWriter]
	// and then return. Returning signals that the request is finished; it
	// is not valid to use the [ResponseWriter] or read from the
	// [Request.Body] after or concurrently with the completion of the
	// ServeHTTP call.
	ServeConn(*Response, *Request)
}

type Config struct {
	// Connection
	Network  string
	Address  string
	Deadline time.Time
	// Misc
	Log     Log
	Handler Handler
}

func NewConfig(address string) *Config {
	c := &Config{
		Address: address,
	}
	c.setDefaults()
	return c
}

func (c *Config) setDefaults() {
	config := c
	// Config defaults
	if config.Log == nil {
		config.Log = NewLogger()
	}
	if config.Handler == nil {
		config.Handler = NewMux() // handler interface for ServeConn
	}
	if config.Network == "" {
		config.Network = "tcp"
	}
	if config.Address == "" {
		config.Address = "localhost:8080"
	}
	if config.Deadline.IsZero() {
		config.Deadline = time.Now().Add(5 * time.Minute)
	}
}

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

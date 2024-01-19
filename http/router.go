package http

import (
	"bytes"
	"fmt"
)

// HandlerInterface is an interface with the method ServeConn(ResponseWriter, *Request) that handles and responds to an HTTP request.
type HandlerInterface interface {
	// ServeHTTP should write reply headers and data to the [ResponseWriter]
	// and then return. Returning signals that the request is finished; it
	// is not valid to use the [ResponseWriter] or read from the
	// [Request.Body] after or concurrently with the completion of the
	// ServeHTTP call.
	ServeConn(*Response, *Request)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(*Response, *Request)

// ServeConn calls f(w, r).
func (f HandlerFunc) ServeConn(w *Response, r *Request) {
	f(w, r)
}

// Handlers is a map of handlers.
type Handlers map[string]HandlerFunc

// Router is a barbones router that maps requests to handlers.
type Router struct {
	handlers Handlers
}

// Returns a new Router
func NewRouter() *Router {
	router := &Router{
		handlers: make(Handlers, 0),
	}

	addHandler(router, "NotFound", "/", notFoundHandler)
	return router
}

// HandleFunc registers the handler function for the given pattern.
// func HandleFunc(pattern string, handler func(ResponseWriter, *message.Request)) {}

// Add handler to router
func addHandler(r *Router, method string, path string, handler HandlerFunc) {
	r.handlers[method+" "+path] = handler
}

// Get handler from router
func getHandler(r *Router, method string, path string) HandlerFunc {
	return r.handlers[method+" "+path]
}

// Route request to handler
func (r *Router) Route(req *Request, w *Response) {
	fmt.Println("Router: Routing request", req)

	// Get handler
	// TODO: Which path to use req.URL or req.RequestURI?
	handler := getHandler(r, req.Method, req.URL.Path)
	if handler == nil {
		// Not found
		fmt.Println("Route not found.")
		handler = getHandler(r, "NotFound", "/")
	}
	// Call handler
	handler(w, req)
}

// **********************************************************************************************************************
// Default Handlers
// **********************************************************************************************************************

func notFoundHandler(w *Response, r *Request) {
	s := bytes.NewBuffer([]byte("404 Not Found"))
	w.Write(s)
}

func (r *Router) NotFound(path string, handler HandlerFunc) {
	addHandler(r, "NotFound", path, handler)
}

func (r *Router) PING(path string, handler HandlerFunc) {
	addHandler(r, "PING", path, handler)
}

// **********************************************************************************************************************
// CRUD
// **********************************************************************************************************************
func (r *Router) GET(path string, handler HandlerFunc) {
	addHandler(r, "GET", path, handler)
}

func (r *Router) POST(path string, handler HandlerFunc) {
	addHandler(r, "POST", path, handler)
}

func (r *Router) PUT(path string, handler HandlerFunc) {
	addHandler(r, "PUT", path, handler)
}

func (r *Router) DELETE(path string, handler HandlerFunc) {
	addHandler(r, "DELETE", path, handler)
}

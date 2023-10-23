package router

import (
	"fmt"
	"message"
)

// Handler is an interface with the method ServeConn(ResponseWriter, *Request) that handles and responds to an HTTP request.
type Handler interface {
	// ServeHTTP should write reply headers and data to the [ResponseWriter]
	// and then return. Returning signals that the request is finished; it
	// is not valid to use the [ResponseWriter] or read from the
	// [Request.Body] after or concurrently with the completion of the
	// ServeHTTP call.
	ServeConn(*message.Response, *message.Request)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(*message.Response, *message.Request)

// ServeConn calls f(w, r).
func (f HandlerFunc) ServeConn(w *message.Response, r *message.Request) {
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
func (r *Router) Route(req *message.Request, w *message.Response) {
	fmt.Println("Router: Routing request", req)

	// Get handler
	handler := getHandler(r, req.Method(), req.Path())
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

func notFoundHandler(w *message.Response, req *message.Request) {
	w.Write([]byte("404 Not Found"))
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

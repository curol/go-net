package router

import (
	"fmt"
	"message"
)

// Request is a structure that represents an HTTP request received by a server or to be sent by a client.
type Request = message.Request

// ResponseWriter is an interface that is used by an HTTP handler to construct an HTTP response.
type ResponseWriter = message.ResponseWriter

// Handler is a function that handles a client request.
type Handler = func(message.Request, message.ResponseWriter)

// Handlers is a map of handlers.
type Handlers map[string]Handler

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

// Add handler to router
func addHandler(r *Router, method string, path string, handler Handler) {
	r.handlers[method+" "+path] = handler
}

// Get handler from router
func getHandler(r *Router, method string, path string) Handler {
	return r.handlers[method+" "+path]
}

// Route request to handler
func (r *Router) Route(req Request, w ResponseWriter) {
	fmt.Println("Router: Routing request", req)

	// Get handler
	handler := getHandler(r, req.Method(), req.Path())
	if handler == nil {
		// Not found
		fmt.Println("Route not found.")
		handler = getHandler(r, "NotFound", "/")
	}
	// Call handler
	handler(req, w)
}

// **********************************************************************************************************************
// Default Handlers
// **********************************************************************************************************************

func notFoundHandler(req Request, w ResponseWriter) {
	w.Write([]byte("404 Not Found"))
}

func (r *Router) NotFound(path string, handler Handler) {
	addHandler(r, "NotFound", path, handler)
}

func (r *Router) PING(path string, handler Handler) {
	addHandler(r, "PING", path, handler)
}

// **********************************************************************************************************************
// CRUD
// **********************************************************************************************************************
func (r *Router) GET(path string, handler Handler) {
	addHandler(r, "GET", path, handler)
}

func (r *Router) POST(path string, handler Handler) {
	addHandler(r, "POST", path, handler)
}

func (r *Router) PUT(path string, handler Handler) {
	addHandler(r, "PUT", path, handler)
}

func (r *Router) DELETE(path string, handler Handler) {
	addHandler(r, "DELETE", path, handler)
}

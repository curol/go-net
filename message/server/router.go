package server

import (
	"fmt"
	"message"
)

// Request is a structure that represents an HTTP request received by a server or to be sent by a client.
type Request = message.Request

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
func (r *Router) Route(req *Request, w ResponseWriter) {
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

func notFoundHandler(w ResponseWriter, req *Request) {
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

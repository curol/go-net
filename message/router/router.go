package router

import (
	"fmt"
	"message"
)

type Handler = message.Handler

// Handlers is a map of handlers.
type Handlers map[string]message.Handler

// Router is a barbones router that maps requests to handlers.
type Router struct {
	handlers Handlers
}

type Request = message.Request

type ResponseWriter = message.ResponseWriter

// Return new Router
func NewRouter() *Router {
	router := &Router{
		handlers: make(Handlers),
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

func notFoundHandler(req Request, w ResponseWriter) {
	w.Write([]byte("404 Not Found"))
}

func (r *Router) NotFound(path string, handler Handler) {
	addHandler(r, "NotFound", path, handler)
}

func (r *Router) PING(path string, handler Handler) {
	addHandler(r, "PING", path, handler)
}

// CRUD methods
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

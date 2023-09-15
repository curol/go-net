package router

import (
	"fmt"
	"request"
	"writer"
)

// Router maps requests to handlers
type Router struct {
	handlers request.Handlers
}

// Return new Router
func NewRouter() *Router {
	router := &Router{
		handlers: make(request.Handlers),
	}
	addHandler(router, "NotFound", "/", notFoundHandler)
	return router
}

// Add handler to router
func addHandler(r *Router, method string, path string, handler request.Handler) {
	r.handlers[method+" "+path] = handler
}

// Get handler from router
func getHandler(r *Router, method string, path string) request.Handler {
	return r.handlers[method+" "+path]
}

// Route request to handler
func (r *Router) Route(req *request.Request, w *writer.ResponseWriter) {
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

func notFoundHandler(req *request.Request, w *writer.ResponseWriter) {
	w.Write([]byte("404 Not Found"))
}

func (r *Router) NotFound(path string, handler request.Handler) {
	addHandler(r, "NotFound", path, handler)
}

func (r *Router) PING(path string, handler request.Handler) {
	addHandler(r, "PING", path, handler)
}

// CRUD methods
func (r *Router) GET(path string, handler request.Handler) {
	addHandler(r, "GET", path, handler)
}

func (r *Router) POST(path string, handler request.Handler) {
	addHandler(r, "POST", path, handler)
}

func (r *Router) PUT(path string, handler request.Handler) {
	addHandler(r, "PUT", path, handler)
}

func (r *Router) DELETE(path string, handler request.Handler) {
	addHandler(r, "DELETE", path, handler)
}

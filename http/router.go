package http

import (
	"fmt"
)

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

func notFoundHandler(w ResponseWriter, r *Request) {
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

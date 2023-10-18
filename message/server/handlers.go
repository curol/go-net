// ******************************************************
// Handlers
//
// A Handler handles the client's request and response.
// ******************************************************
package server

import "message"

// Handler is an interface with the method ServeConn(ResponseWriter, *Request) that handles and responds to an HTTP request.
type Handler interface {
	// ServeHTTP should write reply headers and data to the [ResponseWriter]
	// and then return. Returning signals that the request is finished; it
	// is not valid to use the [ResponseWriter] or read from the
	// [Request.Body] after or concurrently with the completion of the
	// ServeHTTP call.
	ServeConn(ResponseWriter, *message.Request)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(ResponseWriter, *message.Request)

// ServeConn calls f(w, r).
func (f HandlerFunc) ServeConn(w ResponseWriter, r *message.Request) {
	f(w, r)
}

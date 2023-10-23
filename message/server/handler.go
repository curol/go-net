// ******************************************************
// Handlers
//
// A Handler handles the client's request and response.
// ******************************************************
package server

// Handler is an interface with the method ServeConn(ResponseWriter, *Request) that handles and responds to an HTTP request.
type Handler interface {
	// ServeHTTP should write reply headers and data to the [ResponseWriter]
	// and then return. Returning signals that the request is finished; it
	// is not valid to use the [ResponseWriter] or read from the
	// [Request.Body] after or concurrently with the completion of the
	// ServeHTTP call.
	ServeConn(*Response, *Request)
}

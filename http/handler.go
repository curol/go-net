package http

// A Handler responds to an HTTP request.
//
// [Handler.ServeHTTP] should write reply headers and data to the [ResponseWriter]
// and then return. Returning signals that the request is finished; it
// is not valid to use the [ResponseWriter] or read from the
// [Request.Body] after or concurrently with the completion of the
// ServeHTTP call.
//
// Depending on the HTTP client software, HTTP protocol version, and
// any intermediaries between the client and the Go server, it may not
// be possible to read from the [Request.Body] after writing to the
// [ResponseWriter]. Cautious handlers should read the [Request.Body]
// first, and then reply.
//
// Except for reading the body, handlers should not modify the
// provided Request.
//
// If ServeHTTP panics, the server (the caller of ServeHTTP) assumes
// that the effect of the panic was isolated to the active request.
// It recovers the panic, logs a stack trace to the server error log,
// and either closes the network connection or sends an HTTP/2
// RST_STREAM, depending on the HTTP protocol. To abort a handler so
// the client sees an interrupted response but the server doesn't log
// an error, panic with the value [ErrAbortHandler].
type Handler interface {
	// ServeHTTP should write reply headers and data to the [ResponseWriter]
	// and then return. Returning signals that the request is finished; it
	// is not valid to use the [ResponseWriter] or read from the
	// [Request.Body] after or concurrently with the completion of the
	// ServeHTTP call.
	ServeHTTP(ResponseWriter, *Request)
}

// // HandlerInterface is an interface with the method ServeConn(ResponseWriter, *Request) that handles and responds to an HTTP request.
// type HandlerInterface interface {
// 	// ServeHTTP should write reply headers and data to the [ResponseWriter]
// 	// and then return. Returning signals that the request is finished; it
// 	// is not valid to use the [ResponseWriter] or read from the
// 	// [Request.Body] after or concurrently with the completion of the
// 	// ServeHTTP call.
// 	ServeHTTP(*Response, *Request)
// }

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(ResponseWriter, *Request)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
	f(w, r)
}

// Handlers is a map of handlers.
type Handlers map[string]HandlerFunc

package server

// Mux implements interface Handler for handling requests.
type Mux struct{}

// NewMux returns a new Mux.
func NewMux() *Mux {
	return &Mux{}
}

// ServeConn handles a request and response for the client.
func (m *Mux) ServeConn(w *Response, r *Request) {
	w.NotFound()
}

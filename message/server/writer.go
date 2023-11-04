package server

import "gonet"

// A ResponseWriter interface is used by an HTTP handler to construct an HTTP response.
//
// Note, a ResponseWriter may not be used after [Handler.ServeHTTP] has returned.
type ResponseWriter interface {
	Write(b []byte) (int, error)
	WriteHeader(string, string)
	Header() gonet.Header
}

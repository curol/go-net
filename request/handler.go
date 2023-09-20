package request

import "writer"

// ******************************************************
// Handlers
// Handlers for handling client's requests
// ******************************************************
type Handler func(r *Request, w *writer.ResponseWriter)
type Handlers map[string]Handler

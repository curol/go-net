package textproto

// // Req is a wrapper for textproto.Reader which implements convenience methods for reading requests.
// type Request struct {
// 	*TextMessage
// 	r *bufio.Reader
// }

// func NewRequest(r *bufio.Reader) *Request {
// 	req := &Request{}
// 	req.r = r
// 	req.Status = "GET / HTTP/1.0"           // Default
// 	req.Headers = make(map[string][]string) // Default
// 	return req
// }

// // Read reads from the underlying bufio.Reader and parses the request into the Request.
// func (req *Request) Read() error {
// 	err := req.ReadFrom(req.r)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

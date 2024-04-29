package textproto

// // Res is a wrapper for textproto.Writer which implements convenience methods for reading responses.
// type Response struct {
// 	*TextMessage
// 	W *bufio.Writer
// }

// func NewResponse(tm *TextMessage, w *bufio.Writer) *Response {
// 	res := &Response{}
// 	res.W = w
// 	res.Status = tm.Status
// 	if tm.Headers == nil {
// 		tm.Headers = make(map[string][]string) // Default
// 	}
// 	return res
// }

// // Flush writes any buffered data to the underlying io.Writer.
// func (res *Response) Flush() error {
// 	return res.W.Flush()
// }

// // Close closes the Response, writing any buffered data to the underlying io.Writer.
// func (res *Response) Close() error {
// 	r := bufio.NewReader(res.buf)
// 	err := res.ReadFrom(r) // Parse the response
// 	if err != nil {
// 		return err
// 	}
// 	_, err = res.WriteTo(res.W) // Write the response to W
// 	if err != nil {
// 		return err
// 	}
// 	err = res.Flush() // Flush the response
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

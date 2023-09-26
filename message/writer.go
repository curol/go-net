package message

import (
	"bufio"
	"io"
	"net"
	"strings"
)

//********************************************************************************************************************//
// Response
//********************************************************************************************************************//

// Body represents the response body.
type Response struct {
	// Head
	Status     string // status line
	StatusCode int    // status code
	Protocol   string
	Header     Header
	// Body
	Body          io.ReadCloser
	ContentLength int
	// Buffers
	w *bufio.Writer
}

func NewResponse(conn net.Conn) *Response {
	return &Response{
		Header: NewHeader(),
		w:      bufio.NewWriter(conn),
		Body:   NewBody(conn),
	}
}

func (r *Response) WriteTo(w io.Writer) (int64, error) {
	// Head
	head := r.Head()
	n, err := w.Write([]byte(head))
	if err != nil {
		return int64(n), err
	}

	// Body
	if r.Body == nil {
		r.Body = io.NopCloser(strings.NewReader(""))
	}
	cl := r.ContentLength // get content length
	n2, err := copyN(w, r.Body, int64(cl))

	return int64(n) + n2, err
}

// Head returns the response head as a string.
func (r *Response) Head() string {
	// Head
	statusLine := r.Status + "\r\n"      // status line
	headers := r.Header.Join("\r\n")     // headers
	return statusLine + headers + "\r\n" // head
}

// Close closes the response body.
func (r *Response) Close() error {
	return r.Body.Close()
}

// Write writes r to w in the HTTP/1.x server response format,
// including the status line, headers, body, and optional trailer.
//
// This method consults the following fields of the response r:
//
//	StatusCode
//	ProtoMajor
//	ProtoMinor
//	Request.Method
//	TransferEncoding
//	Trailer
//	Body
//	ContentLength
//	Header, values for non-canonical keys will have unpredictable behavior
//
// The Response Body is closed after it is sent.
func (r *Response) Write(w io.Writer) (int64, error) {
	// r.WriteTo(w)

}

func (r *Response) WriteHeader(k string, v string) {
	r.Header.Set(k, v)
}

// ReadResponse reads and returns an HTTP response from r.
// The req parameter optionally specifies the Request that corresponds
// to this Response. If nil, a GET request is assumed.
// Clients must call resp.Body.Close when finished reading resp.Body.
// After that call, clients can inspect resp.Trailer to find key/value
// pairs included in the response trailer.
func ReadResponse(r io.Reader) (*Response, error) {
	return nil, nil
}

//********************************************************************************************************************//
// ResponseWriter
//********************************************************************************************************************//

// // Write writes b to w.
// func writeMessage(w ResponseWriter, b []byte) (int, error) {
// 	// NewBuffer creates and initializes a new Buffer
// 	// using b as its initial contents.
// 	buf := bytes.NewBuffer(b)
// 	// Write data to buffer
// 	n, err := w.Write(buf)
// 	if err != nil {
// 		return 0, err
// 	}
// 	// Append empty line to mark end of last line
// 	return int(n), nil
// }

// // ToBytes returns the headers as a byte slice.
// func toBytes(w ResponseWriter) ([]byte, error) {
// 	var buf bytes.Buffer
// 	_, err := w.Write(&buf)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return buf.Bytes(), nil
// }

// // String returns the text of the header formatted in the same way as in the request.
// func toString(w ResponseWriter) string {
// 	var buf bytes.Buffer
// 	w.WriteTo(&buf)
// 	return buf.String()
// }

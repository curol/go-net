package textproto

import (
	"bufio"
	"fmt"
)

// serialize serializes the TextMessage to bw and returns the number of bytes written
// It writes the status line, headers, and body to w.
//
// Example:
// tp := &TextMessage{Status: "200 OK", Headers: MIMEHeader{"Content-Type": []string{"text/plain"}}, Body: bufio.NewReader(strings.NewReader("Hello, World!"))}
// bw := bufio.NewWriter(os.Stdout)
// n, err := tm.serialize(bw)
func serialize(tm *TextMessage, w *bufio.Writer, doSerializeBody bool) (int64, error) {
	var n int64
	var err error
	headers := tm.Headers
	dlm := "\r\n"
	// 1. Clean up
	defer func() {
		err = w.Flush()
		if err != nil {
			fmt.Println("textproto: Error with serialization in tm.serialize()", err)
		}
	}()
	// 2. Status line
	sn, _ := fmt.Fprintf(w, "%s%s", tm.Status, dlm) // Write status line
	n = int64(sn)
	// 3. Headers
	// h := serializeHeaders(tm.Headers)
	// n2, _ = fmt.Fprint(w, string(h)) // Write headers
	for name, values := range headers {
		for _, value := range values {
			hn, _ := fmt.Fprintf(w, "%s: %s%s", name, value, dlm)
			n += int64(hn)
		}
	}
	// 4. End of headers
	ehn, _ := fmt.Fprintf(w, dlm) // Write blank line between headers and body
	n += int64(ehn)
	// 5. Body
	if !doSerializeBody || tm.Body == nil {
		return n, err
	}
	// 5.1. Validate
	cl := tm.ContentLen
	// TODO: if cl < 0, then use max read size
	if cl < 0 {
		return 0, fmt.Errorf("textproto: Content-Length less than 0 %d", cl)
	}
	if cl == 0 {
		return 0, nil
	}
	if tm.isBodyRead {
		return 0, fmt.Errorf("textproto: Body already read")
	}
	// 5.2. Copy body to dst
	bn, err := tm.readBody(w, cl)
	n += bn
	if err != nil {
		return n, err
	}
	// 5.3. Check if body is fully read
	if bn != cl {
		se := fmt.Sprintf("textproto: content length '%d' doesn't match bytes written to w'%d'", cl, bn)
		return n, fmt.Errorf(se)
	}
	// 6. Return bytes read and error
	tm.isSerialized = true
	return n, err
}

package tests

import (
	"bytes"
	"fmt"
	"go/token"
	"reflect"
	"strings"
	"testing"

	httplib "net/http" // http standard library
	"net/http/httputil"

	http "github.com/curol/network/http"
)

func mockHttpReq() *httplib.Request {
	req, err := httplib.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		panic(err)
	}
	return req
}

func dumpHttpReq(req *httplib.Request) string {
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		panic(err)
	}
	return string(dump)
}

// Print struct fields and values
func prettyPrint(v any) {
	switch v := v.(type) {
	case *http.Request:
		//
		fmt.Println(v.Dump())
	case *httplib.Request:
		reqDump, err := httputil.DumpRequestOut(v, true)
		if err != nil {
			panic(err)
		} else {
			fmt.Printf(string(reqDump))
		}
	default:
		//
	}
}

// Compare `github.com/curol/network.Request` to standard library http request `net/http.Request`
func compareRequests(r1 *http.Request, r2 *httplib.Request, t *testing.T) {
	// Validate
	myreq := r1
	otherreq := r2
	if myreq == nil || otherreq == nil {
		t.Errorf("One of the requests is nil")
	}

	// Write
	myreqbuf := bytes.NewBuffer(nil)
	myreq.Write(myreqbuf)
	otherreqbuf := bytes.NewBuffer(nil)
	otherreq.Write(otherreqbuf)

	if myreqbuf.String() != otherreqbuf.String() {
		t.Errorf("Requests are not equal")
	}
}

// reqBytes treats req as a request (with \n delimiters) and returns it with \r\n delimiters,
// ending in \r\n\r\n
func reqBytes(req string) []byte {
	return []byte(strings.ReplaceAll(strings.TrimSpace(req), "\n", "\r\n") + "\r\n\r\n")
}

// Compare any two strcuts
func diff(t *testing.T, prefix string, have, want any) {
	t.Helper()
	hv := reflect.ValueOf(have).Elem()
	wv := reflect.ValueOf(want).Elem()
	if hv.Type() != wv.Type() {
		t.Errorf("%s: type mismatch %v want %v", prefix, hv.Type(), wv.Type())
	}
	for i := 0; i < hv.NumField(); i++ {
		name := hv.Type().Field(i).Name
		if !token.IsExported(name) {
			continue
		}
		hf := hv.Field(i).Interface()
		wf := wv.Field(i).Interface()
		if !reflect.DeepEqual(hf, wf) {
			t.Errorf("%s: %s = %v want %v", prefix, name, hf, wf)
		}
	}
}

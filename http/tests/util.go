package tests

import (
	"bytes"
	"go/token"
	net "net/http"
	"reflect"
	"strings"
	"testing"

	http "github.com/curol/network/http"
)

// Compare github.com/curol/network.Request to http.Request
func compareReqToHttpRequest(r1 *http.Request, r2 *net.Request, t *testing.T) {
	buf := bytes.NewBuffer(nil) // network.Request
	r1.Write(buf)               // write request
	myreq := buf.String()       // get request as string

	buf = bytes.NewBuffer(nil) // http.Request
	r2.Write(buf)
	libreq := buf.String()
	if myreq != libreq {
		t.Errorf("github.com/curol/network/http.Request != http.Request\ngithub.com/curol/network/http.Request:\n%s\n\nhttp.Request:\n%s\n", myreq, libreq)
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

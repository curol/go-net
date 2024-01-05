package httptest

import (
	"crypto/tls"
	"net/textproto"
	"strconv"
)

// A Config structure is used to configure a TLS client or server. After one has been passed to a TLS function it must not be modified. A Config may be reused; the tls package will also not modify it.
func newTlsConfig(cert *tls.Certificate) *tls.Config {
	// Now you can use `cert` in a tls.Config struct for example:
	return &tls.Config{
		Certificates: []tls.Certificate{*cert},
	}
}

// parseContentLength trims whitespace from s and returns -1 if no value
// is set, or the value if it's >= 0.
//
// This a modified version of same function found in net/http/transfer.go. This
// one just ignores an invalid header.
func parseContentLength(cl string) int64 {
	cl = textproto.TrimString(cl)
	if cl == "" {
		return -1
	}
	n, err := strconv.ParseUint(cl, 10, 63)
	if err != nil {
		return -1
	}
	return int64(n)
}

package slice

import (
	"bytes"
	"strings"
)

// *********************************************************************************************************************
// Slices
// *********************************************************************************************************************

// BytesEqual checks if two byte slices are equal.
func BytesEqual(a, b []byte) bool {
	return bytes.Equal(a, b)
}

// Contains checks if a slice contains a string.
func Contains(sl []string, str string) bool {
	for _, v := range sl {
		if v == str {
			return true
		}
	}
	return false
}

// JoinCRLF joins a slice of strings with seperator "\r\n".
func JoinCRLF(sl []string) string {
	return strings.Join(sl, "\r\n")
}

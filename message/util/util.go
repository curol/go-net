package util

import (
	"bytes"
	"sort"
)

//**********************************************************************************************************************
// Slice
//**********************************************************************************************************************

// CopyBytes copies elements from a source slice into a destination slice. (As a special case, it also will copy bytes from a string to a slice of bytes.) The source and destination may overlap. Copy returns the number of elements copied, which will be the minimum of len(src) and len(dst).
func copyBytes(dst, src []byte) int {
	return copy(dst, src)
}

// SortNumbers sorts a slice of integers in increasing order.
func sortNumbers(numbers []int) {
	sort.Ints(numbers)
}

// SortStrings sorts a slice of strings in increasing order.
func sortStrings(strings []string) {
	sort.Strings(strings)
}

// Merge  takes two sorted arrays as input and returns a single sorted array as output.
func mergeInt(a, b []int) []int {
	result := make([]int, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] < b[j] {
			result = append(result, a[i])
			i++
		} else {
			result = append(result, b[j])
			j++
		}
	}
	result = append(result, a[i:]...)
	result = append(result, b[j:]...)
	return result
}

// MergeSliceOfStrings takes a slice of slices of strings and returns a single slice of strings.
func mergeSlicesOfStrings(slices [][]string) []string {
	result := make([]string, 0)
	for _, slice := range slices {
		result = append(result, slice...)
	}
	return result
}

//**********************************************************************************************************************
// Map
//**********************************************************************************************************************

// SortMapKeys sorts the keys of a map in increasing order.
func sortMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

//**********************************************************************************************************************
// Buffer
//**********************************************************************************************************************

// NewBuffer creates and initializes a new Buffer using buf as its initial contents. The new Buffer takes ownership of buf, and the caller should not use buf after this call. NewBuffer is intended to prepare a Buffer to read existing data. It can also be used to set the initial size of the internal buffer for writing. To do that, buf should have the desired capacity but a length of zero.
//
// In most cases, new(Buffer) (or just declaring a Buffer variable) is sufficient to initialize a Buffer.
func newBuffer(data []byte) *bytes.Buffer {
	return bytes.NewBuffer(data)
}

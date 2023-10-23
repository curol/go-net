package hashmap

import (
	"fmt"
	"slices"
	"strings"
)

type HashMap map[string]string

func New() HashMap {
	return make(HashMap)
}

func NewFromStrings(strs []string) HashMap {
	m := make(HashMap)
	for _, s := range strs {
		kv := strings.SplitN(s, ":", 2)
		if len(kv) == 2 {
			m.Set(kv[0], kv[1])
		}
	}
	return m
}

func NewFromString(s string) HashMap {
	return NewFromStrings(strings.Split(s, "\r\n"))
}

func NewFromBytes(b []byte) HashMap {
	return NewFromString(string(b))
}

func NewFromMap(m map[string]string) HashMap {
	var lines []string
	for k, v := range m {
		line := fmt.Sprintf("%s: %s", k, v)
		lines = append(lines, line)
	}
	return NewFromStrings(lines)
}

// Set sets the HashMap's value.
func (m HashMap) Set(key, value string) {
	k := strings.TrimSpace(key)
	v := strings.TrimSpace(value)
	m[k] = v
}

// Get gets the value associated with the given key.
func (m HashMap) Get(key string) (string, bool) {
	k := strings.TrimSpace(key)
	if values, ok := m[k]; !ok || len(values) == 0 {
		return "", false
	}
	return m[key], true
}

// Del deletes the values associated with key.
func (m HashMap) Del(key string) {
	k := strings.TrimSpace(key)
	delete(m, k)
}

// Clone creates a new HashMap with the same keys and values as the original.
// It does not create deep copies of the values, so changes to the original
// HashMap may affect the copied HashMap if the values are pointers or slices.
func (m HashMap) Clone() HashMap {
	m2 := New()
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

// // Len returns the number of keys.
// func (m HashMap) Len() int {
// 	return len(m)
// }

// Keys returns the keys of the HashMap.
func (m HashMap) Keys() []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values returns the values of the HashMap.
func (m HashMap) Values() []string {
	var values []string
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// Merge merges other into m.
func (m HashMap) Merge(other HashMap) {
	for k, v := range other {
		m[k] = v
	}
}

// ToStrings returns the HashMaps as a slice of sorted strings.
func (m HashMap) ToStrings() []string {
	// Get keys
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	// Sort keys
	slices.Sort(keys)
	// Create slice of strings from map
	sl := make([]string, 0, len(m))
	for _, k := range keys {
		s := fmt.Sprintf("%s: %s", k, m[k])
		sl = append(sl, s)
	}
	return sl
}

// ToString returns the HashMaps as a string.
func (m HashMap) ToString() string {
	strs := m.ToStrings()
	return strings.Join(strs, "\r\n")
}

// ToBytes returns the HashMaps as a byte slice.
func (m HashMap) ToBytes() []byte {
	return []byte(m.ToString())
}

// Size is len(hashMap.ToString())
func (m HashMap) Size() int {
	return len(m.ToBytes())
}

// Clear clears the HashMap.
func (m HashMap) Clear() {
	// for k := range m {
	// 	delete(m, k)
	// }
	m = make(HashMap)
}

// Equals checks if two HashMaps are equal.
func (m HashMap) Equals(other HashMap) bool {
	if len(m) != len(other) {
		return false
	}
	for k, v := range m {
		if v != other[k] {
			return false
		}
	}
	return true
}

// Join returns the HashMaps as a string joined by sep.
func (m HashMap) Join(sep string) string {
	return strings.Join(m.ToStrings(), sep)
}

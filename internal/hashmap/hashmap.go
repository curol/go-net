package hashmap

import (
	"fmt"
	"slices"
	"strings"
)

type HashMap map[string]string

func New(m map[string]string) HashMap {
	if m == nil {
		return make(HashMap)
	}

	h := make(HashMap, len(m))
	h.FromMap(m)
	return h
}

//*********************************************************************************************************************
// General
//*********************************************************************************************************************

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
	m2 := New(nil)
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

// Len returns the number of keys.
func (m HashMap) Len() int {
	return len(m)
}

// Size is len(hashMap.ToString())
// func (m HashMap) Size() int {
// 	return len(m.ToBytes())
// }

// Keys returns the keys of the HashMap.
func (m HashMap) Keys() []string {
	var keys []string
	for k := range m {
		keys = append(keys, k) // append key
	}
	slices.Sort(keys) // sort keys
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

// Clear clears the HashMap.
func (m HashMap) Clear() {
	m = New(nil)
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

//*********************************************************************************************************************
// Serialize
//*********************************************************************************************************************

// Serialize returns the HashMaps as a slice of sorted strings.
//
// Note:
//   - Format is ["key: value", "key: value", ...]
//   - This is the same as HashMap.ToStrings().
func (m HashMap) serialize() []string {
	keys := m.Keys()
	// Create slice of strings from map
	sl := make([]string, 0, m.Len())
	for _, key := range keys {
		val := m[key]
		s := fmt.Sprintf("%s: %s", key, val)
		sl = append(sl, s)
	}
	return sl
}

//*********************************************************************************************************************
// Encode
//*********************************************************************************************************************

// ToStrings returns the HashMaps as a slice of sorted strings.
func (m HashMap) ToStrings() []string {
	return m.serialize()
}

// ToString returns the HashMaps as a string with seperator delm.
//
// Note:
//   - string ends with blank line "\r\n" signaling end.
func (m HashMap) ToString(delm string) string {
	if delm == "" {
		delm = "\r\n"
	}
	if m.Len() == 0 {
		return delm
	}
	strs := m.ToStrings()
	return strings.Join(strs, delm) + delm
}

// ToBytes returns the HashMaps as a byte slice with seperator delm.
//
// Note:
//   - byte slice ends with blank line "\r\n" signaling end.
func (m HashMap) ToBytes(delm string) []byte {
	return []byte(m.ToString(delm))
}

//*********************************************************************************************************************
// Decode
//*********************************************************************************************************************

// FromMap sets the HashMap's values from a map of strings.
func (m HashMap) FromMap(m2 map[string]string) {
	for k, v := range m2 {
		m.Set(k, v)
	}
}

// FromStrings sets the HashMap's values from a slice of strings.
//
// Note, each string must be in the format:
//
//	"key: value"
func (m HashMap) FromStrings(strs []string) {
	for _, s := range strs {
		kv := strings.SplitN(s, ":", 2)
		if len(kv) == 2 {
			k := strings.TrimSpace(kv[0])
			v := strings.TrimSpace(kv[1])
			m.Set(k, v)
		}
	}
}

// FromString sets the HashMap's values from a string.
//
// Note, the string must be in the format:
//
//	"key1: value1\r\nkey2: value2\r\n"
func (m HashMap) FromString(s string, sep string) {
	m.FromStrings(strings.Split(s, sep))
}

// FromBytes sets the HashMap's values from a byte slice.
//
// Note, the byte slice must be in the format:
//
//	"key1: value1\r\nkey2: value2\r\n"
func (m HashMap) FromBytes(b []byte, sep string) {
	m.FromString(string(b), sep)
}

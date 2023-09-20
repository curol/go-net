package url

import (
	"errors"
	"net/url"
	"strings"
)

// *********************************************************************************************************************
// Queries
// *********************************************************************************************************************
// DecodeQuery parses the URL-encoded query string and returns a map listing the values specified for each key.
func DecodeQuery(query string) (map[string][]string, error) {
	// ParseQuery parses the URL-encoded query string and returns a map listing the values specified for each key.
	return url.ParseQuery(query)
}

// EncodeQuery encodes the values into “URL encoded” form ("bar=baz&foo=quux") sorted by key.
func EncodeQuery(values map[string][]string) string {
	// for key, value := range values {
	// 	values[key] = []string{strings.Join(value, ",")}
	// }
	// Encode encodes the values into “URL encoded” form ("bar=baz&foo=quux") sorted by key.
	return url.Values(values).Encode()
}

// DecodeForm parses the form data and returns a map listing the values specified for each key.
func DecodeForm(data string) (map[string][]string, error) {
	return DecodeQuery(data)
}

// ParseQueriesFromBytes parses a byte slice for ampersand-separated key-value pairs.
// Same as url.ParseQuery, but takes a byte slice instead of a string.
func ParseQueriesFromBytes(b []byte) (map[string]string, error) {
	str := string(b)
	pairs := strings.Split(str, "&")
	result := make(map[string]string)

	for _, pair := range pairs {
		keyValue := strings.Split(pair, "=")
		if len(keyValue) != 2 {
			return nil, errors.New("input not properly formatted")
		}
		result[keyValue[0]] = keyValue[1]
	}

	return result, nil
}

package cookie

import (
	"strconv"
	"strings"
	"time"
)

// ParseResponse parses cookies from a response header.
//
// For example:
//
//  1. Raw response:
//     ```
//     HTTP/1.1 200 OK
//     Content-Type: text/plain
//     Set-Cookie: mycookie=value; Path=/; Domain=example.com; Expires=Wed, 21 Oct 2021 07:28:00 GMT; Max-Age=3600; Secure
//     Content-Length: 13
//
//     Hello, World!
//     ```
//
//  2. Parse header 'Set-Cookie' from response header:
//     ```
//     header := map[string]string{
//     "Content-Type": "text/plain",
//     "Set-Cookie":   "mycookie=value; Path=/; Domain=example.com; Expires=Wed, 21 Oct 2021 07:28:00 GMT; Max-Age=3600; Secure",
//     "Content-Length": "13",
//     }
//     cookies := ParseResponse(header)
//     ```
func ParseResponse(header map[string]string) []*Cookie {
	cookies := make([]*Cookie, 0)
	for k, v := range header {
		// TODO: Since message.Header is a map[string]string, we can't have multiple Set-Cookie headers.
		// TODO: We should probably change message.Header to map[string][]string.
		if k == "Set-Cookie" {
			cookie := parseCookie(v)
			cookies = append(cookies, cookie)
		}
	}
	return cookies
}

// ParseRequest parses a cookie from a request header.
//
// For example:
//
//  1. Raw request:
//     ```
//     GET / HTTP/1.1
//     Host: example.com
//     Cookie: mycookie=value; othercookie=othervalue
//     ```
//
//  2. Parse header 'Cookie' from raw request:
//     ```
//     cookie := ParseCookieRequest("Cookie: mycookie=value; othercookie=othervalue")
//     ```
func ParseRequest(raw string) *Cookie {
	cookie := &Cookie{}
	parts := strings.Split(raw, "=")
	if len(parts) == 2 {
		cookie.Name = strings.TrimSpace(parts[0])
		cookie.Value = strings.TrimSpace(parts[1])
	}
	return cookie
}

// ParseResponseString parses cookies from a response
func parseResponseLines(lines []string) []*Cookie {
	// lines := strings.Split(raw, "\r\n")
	cookies := make([]*Cookie, 0)
	for _, line := range lines {
		if strings.HasPrefix(line, "Set-Cookie: ") {
			cookie := parseCookie(line)
			cookies = append(cookies, cookie)
		}
	}
	return cookies
}

func parseCookie(raw string) *Cookie {
	cookie := &Cookie{}
	parts := strings.Split(raw, ";")
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			switch key {
			case "Path":
				cookie.Path = value
			case "Domain":
				cookie.Domain = value
			case "Expires":
				cookie.RawExpires = value
				if t, err := time.Parse(time.RFC1123, value); err == nil {
					cookie.Expires = t
				}
			case "Max-Age":
				if i, err := strconv.Atoi(value); err == nil {
					cookie.MaxAge = i
				}
			case "Secure":
				cookie.Secure = true
			case "HttpOnly":
				cookie.HttpOnly = true
			default:
				cookie.Name = key
				cookie.Value = value
			}
		}
	}
	return cookie
}

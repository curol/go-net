package cookie

import (
	"fmt"
	"time"
)

// Cookies are for state management since the connection is stateless.
//
// Note:
//   - The header 'Cookie' is for requests.
//   - The header 'Set-Cookie' is for responses.
type Cookie struct {
	Name  string
	Value string

	Path       string    // optional
	Domain     string    // optional
	Expires    time.Time // optional
	RawExpires string    // for reading cookies only

	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge   int
	Secure   bool
	HttpOnly bool
	// SameSite SameSite
	Raw      string
	Unparsed []string // Raw text of unparsed attribute-value pairs
}

// String returns the serialization of the cookie for use in a Cookie header (if only Name and Value are set) or a Set-Cookie response header (if other fields are set).
//
//	If c is nil or c.Name is invalid, the empty string is returned.
func (c *Cookie) String() string {
	return c.Name + "=" + c.Value
}

func (c *Cookie) Valid() error {
	if c.Name == "" {
		return fmt.Errorf("Cookie name is empty")
	}
	return nil
}

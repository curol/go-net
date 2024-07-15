package cookie

import (
	"fmt"
	"testing"
)

func TestParseCookies(t *testing.T) {
	v, e := ParseCookie("foo=value1; boo=value2")
	if e != nil {
		fmt.Println(e)
		return
	}
	for i, c := range v {
		fmt.Println(i)
		fmt.Println(c)
	}
}

// func TestNewCookie(t *testing.T) {
// 	c := cookie.NewCookie("foo", "value1", nil)
// 	fmt.Println(c)
// 	c2 := cookie.NewCookie("boo", "v2", &cookie.CookieOptions{Path: "/path", Domain: "example.com"})
// 	fmt.Println(c2)
// }

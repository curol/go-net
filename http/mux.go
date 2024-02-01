package http

import (
	"sort"
	"strings"
	"sync"
)

// DefaultServeMux is the default [ServeMux] used by [Serve].
var DefaultServeMux = &defaultServeMux

var defaultServeMux ServeMux

// Mux implements interface Handler for handling requests.
// ServeMux is an HTTP request multiplexer.
// It matches the URL of each incoming request against a list of registered
// patterns and calls the handler for the pattern that
// most closely matches the URL.
//
// # Patterns
//
// Patterns can match the method, host and path of a request.
// Some examples:
//
//   - "/index.html" matches the path "/index.html" for any host and method.
//   - "GET /static/" matches a GET request whose path begins with "/static/".
//   - "example.com/" matches any request to the host "example.com".
//   - "example.com/{$}" matches requests with host "example.com" and path "/".
//   - "/b/{bucket}/o/{objectname...}" matches paths whose first segment is "b"
//     and whose third segment is "o". The name "bucket" denotes the second
//     segment and "objectname" denotes the remainder of the path.
//
// In general, a pattern looks like
//
//	[METHOD ][HOST]/[PATH]
//
// All three parts are optional; "/" is a valid pattern.
// If METHOD is present, it must be followed by a single space.
//
// Literal (that is, non-wildcard) parts of a pattern match
// the corresponding parts of a request case-sensitively.
//
// A pattern with no method matches every method. A pattern
// with the method GET matches both GET and HEAD requests.
// Otherwise, the method must match exactly.
//
// A pattern with no host matches every host.
// A pattern with a host matches URLs on that host only.
//
// A path can include wildcard segments of the form {NAME} or {NAME...}.
// For example, "/b/{bucket}/o/{objectname...}".
// The wildcard name must be a valid Go identifier.
// Wildcards must be full path segments: they must be preceded by a slash and followed by
// either a slash or the end of the string.
// For example, "/b_{bucket}" is not a valid pattern.
//
// Normally a wildcard matches only a single path segment,
// ending at the next literal slash (not %2F) in the request URL.
// But if the "..." is present, then the wildcard matches the remainder of the URL path, including slashes.
// (Therefore it is invalid for a "..." wildcard to appear anywhere but at the end of a pattern.)
// The match for a wildcard can be obtained by calling [Request.PathValue] with the wildcard's name.
// A trailing slash in a path acts as an anonymous "..." wildcard.
//
// The special wildcard {$} matches only the end of the URL.
// For example, the pattern "/{$}" matches only the path "/",
// whereas the pattern "/" matches every path.
//
// For matching, both pattern paths and incoming request paths are unescaped segment by segment.
// So, for example, the path "/a%2Fb/100%25" is treated as having two segments, "a/b" and "100%".
// The pattern "/a%2fb/" matches it, but the pattern "/a/b/" does not.
//
// # Precedence
//
// If two or more patterns match a request, then the most specific pattern takes precedence.
// A pattern P1 is more specific than P2 if P1 matches a strict subset of P2’s requests;
// that is, if P2 matches all the requests of P1 and more.
// If neither is more specific, then the patterns conflict.
// There is one exception to this rule, for backwards compatibility:
// if two patterns would otherwise conflict and one has a host while the other does not,
// then the pattern with the host takes precedence.
// If a pattern passed [ServeMux.Handle] or [ServeMux.HandleFunc] conflicts with
// another pattern that is already registered, those functions panic.
//
// As an example of the general rule, "/images/thumbnails/" is more specific than "/images/",
// so both can be registered.
// The former matches paths beginning with "/images/thumbnails/"
// and the latter will match any other path in the "/images/" subtree.
//
// As another example, consider the patterns "GET /" and "/index.html":
// both match a GET request for "/index.html", but the former pattern
// matches all other GET and HEAD requests, while the latter matches any
// request for "/index.html" that uses a different method.
// The patterns conflict.
//
// # Trailing-slash redirection
//
// Consider a [ServeMux] with a handler for a subtree, registered using a trailing slash or "..." wildcard.
// If the ServeMux receives a request for the subtree root without a trailing slash,
// it redirects the request by adding the trailing slash.
// This behavior can be overridden with a separate registration for the path without
// the trailing slash or "..." wildcard. For example, registering "/images/" causes ServeMux
// to redirect a request for "/images" to "/images/", unless "/images" has
// been registered separately.
//
// # Request sanitizing
//
// ServeMux also takes care of sanitizing the URL request path and the Host
// header, stripping the port number and redirecting any request containing . or
// .. segments or repeated slashes to an equivalent, cleaner URL.
//
// # Compatibility
//
// The pattern syntax and matching behavior of ServeMux changed significantly
// in Go 1.22. To restore the old behavior, set the GODEBUG environment variable
// to "httpmuxgo121=1". This setting is read once, at program startup; changes
// during execution will be ignored.
//
// The backwards-incompatible changes include:
//   - Wildcards are just ordinary literal path segments in 1.21.
//     For example, the pattern "/{x}" will match only that path in 1.21,
//     but will match any one-segment path in 1.22.
//   - In 1.21, no pattern was rejected, unless it was empty or conflicted with an existing pattern.
//     In 1.22, syntactically invalid patterns will cause [ServeMux.Handle] and [ServeMux.HandleFunc] to panic.
//     For example, in 1.21, the patterns "/{"  and "/a{x}" match themselves,
//     but in 1.22 they are invalid and will cause a panic when registered.
//   - In 1.22, each segment of a pattern is unescaped; this was not done in 1.21.
//     For example, in 1.22 the pattern "/%61" matches the path "/a" ("%61" being the URL escape sequence for "a"),
//     but in 1.21 it would match only the path "/%2561" (where "%25" is the escape for the percent sign).
//   - When matching patterns to paths, in 1.22 each segment of the path is unescaped; in 1.21, the entire path is unescaped.
//     This change mostly affects how paths with %2F escapes adjacent to slashes are treated.
//     See https://go.dev/issue/21955 for details.
//
// # Example
// ```
//
//	func main() {
//	    r := mux.NewRouter()
//	    r.HandleFunc("/", HomeHandler)
//	    r.HandleFunc("/products", ProductsHandler)
//	    r.HandleFunc("/articles", ArticlesHandler)
//	    http.Handle("/", r)
//	}
//
// ```
type Mux struct {
	// mu sync.RWMutex
	// tree routingNode
	// index  routingIndex
	// patterns []*pattern
}

// NewMux returns a new Mux.
func NewMux() *Mux {
	return &Mux{}
}

// ServeConn handles a request and response for the client.
func (m *Mux) ServeHTTP(w ResponseWriter, r *Request) {}

// Default server mux
type ServeMux struct {
	mu    sync.RWMutex
	m     map[string]muxEntry
	es    []muxEntry // slice of entries sorted from longest to shortest.
	hosts bool       // whether any patterns contain hostnames
}

// ServeConn handles a request and response for the client.
func (mux *ServeMux) ServeHTTP(w ResponseWriter, r *Request) {
	if r.RequestURI == "*" {
		// TODO: Handle OPTIONS request
		// if r.ProtoAtLeast(1, 1) {
		// 	w.Header().Set("Connection", "close")
		// }
		w.WriteHeader(StatusBadRequest)
		return
	}
	h, _ := mux.FindHandler(r)
	h.ServeHTTP(w, r)
}

type muxEntry struct {
	h       Handler
	pattern string
}

// NewServeMux allocates and returns a new ServeMux.
func NewServeMux() *ServeMux {
	return &ServeMux{}
}

// HandleFunc registers the handler function for the given pattern.
func (mux *ServeMux) HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
	mux.handleFunc(pattern, handler)
}

func (mux *ServeMux) handleFunc(pattern string, handler func(ResponseWriter, *Request)) {
	if handler == nil {
		panic("http: nil handler")
	}
	mux.Register(pattern, HandlerFunc(handler))
}

func (mux *ServeMux) Register(pattern string, handler Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()
	mux.register(pattern, handler)
}

func (mux *ServeMux) register(pattern string, handler Handler) {
	if pattern == "" {
		panic("http: invalid pattern " + pattern)
	}
	if handler == nil {
		panic("http: nil handler")
	}
	if _, exist := mux.m[string(pattern)]; exist {
		panic("http: multiple registrations for " + pattern)
	}
	if mux.m == nil {
		mux.m = make(map[string]muxEntry)
	}

	e := muxEntry{h: handler, pattern: pattern}
	mux.m[pattern] = e
	if pattern[len(pattern)-1] == '/' {
		mux.es = appendSorted(mux.es, e)
	}

	if pattern[0] != '/' {
		mux.hosts = true
	}
}

func appendSorted(es []muxEntry, e muxEntry) []muxEntry {
	n := len(es)
	i := sort.Search(n, func(i int) bool {
		return len(es[i].pattern) < len(e.pattern)
	})
	if i == n {
		return append(es, e)
	}
	// we now know that i points at where we want to insert
	es = append(es, muxEntry{}) // try to grow the slice in place, any entry works.
	copy(es[i+1:], es[i:])      // Move shorter entries down
	es[i] = e
	return es
}

// FindHandler returns the handler to use for the given request, consulting r.Method, r.Host, and r.URL.Path.
// It always returns a non-nil handler.
// If the path is not in its canonical form, the handler will be an internally-generated handler that
// redirects to the canonical path.
func (mux *ServeMux) FindHandler(r *Request) (h Handler, pattern string) {
	return mux.findHandler(r.Host, r.URL.Path)
}

// handler is the main implementation of Handler.
// The path is known to be in canonical form, except for CONNECT methods.
func (mux *ServeMux) findHandler(host, path string) (h Handler, pattern string) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()
	// Host-specific pattern takes precedence over generic ones
	if mux.m != nil {
		// if e, ok := mux.m[path]; ok {
		// 	return e.h, e.pattern
		// }
		return mux.match(path)
	}
	return NotFoundHandler(), ""
}

// Find a handler on a handler map given a path string.
// Most-specific (longest) pattern wins.
func (mux *ServeMux) match(path string) (h Handler, pattern string) {
	// Check for exact match first.
	v, ok := mux.m[path]
	if ok {
		return v.h, v.pattern
	}

	// Check for longest valid match.  mux.es contains all patterns
	// that end in / sorted from longest to shortest.
	for _, e := range mux.es {
		if strings.HasPrefix(path, e.pattern) {
			return e.h, e.pattern
		}
	}
	return nil, ""
}

// // A mapping is a collection of key-value pairs where the keys are unique.
// // A zero mapping is empty and ready to use.
// // A mapping tries to pick a representation that makes [mapping.find] most efficient.
// type mapping[K comparable, V any] struct {
// 	s []entry[K, V] // for few pairs
// 	m map[K]V       // for many pairs
// }

// type entry[K comparable, V any] struct {
// 	key   K
// 	value V
// }

// Handle registers the handler for the given pattern.
//
// Example:
//
//	http.Handle("/", http.FileServer(http.Dir("/tmp")))
//	http.Handle("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
//	}))
func Handle(pattern string, handler Handler) {
	DefaultServeMux.Register(pattern, handler)
}

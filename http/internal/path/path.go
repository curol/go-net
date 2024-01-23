package path

import (
	"errors"
	"path/filepath"
)

var errInvalidPath = errors.New("invalid path")

// FromFS converts a slash-separated path into an operating-system path.
//
// FromFS returns an error if the path cannot be represented by the operating
// system. For example, paths containing '\' and ':' characters are rejected
// on Windows.
func FromFS(path string) (string, error) {
	s := filepath.FromSlash(path) // clean and convert slashes
	if !filepath.IsAbs(s) {       // reject relative paths
		return "", errInvalidPath
	}
	return s, nil
}

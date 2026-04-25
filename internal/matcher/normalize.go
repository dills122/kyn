package matcher

import (
	"path"
	"strings"
)

// NormalizePath converts a path to slash-separated, clean, relative form.
func NormalizePath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.ReplaceAll(p, "\\", "/")
	p = strings.TrimPrefix(p, "./")
	p = path.Clean(p)
	if p == "." {
		return ""
	}
	return p
}

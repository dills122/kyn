package matcher

import (
	"fmt"
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

// NormalizeRelativePath normalizes a repository path and rejects values that
// are absolute or escape the repository through a parent traversal.
func NormalizeRelativePath(p string) (string, error) {
	normalized := NormalizePath(p)
	if normalized == "" {
		return "", nil
	}

	unsafe := strings.IndexByte(normalized, 0) >= 0 ||
		path.IsAbs(normalized) ||
		normalized == ".." ||
		strings.HasPrefix(normalized, "../") ||
		isWindowsDrivePath(normalized)
	if unsafe {
		return "", fmt.Errorf("path %q must be repository-relative and remain within --cwd", p)
	}

	return normalized, nil
}

func isWindowsDrivePath(p string) bool {
	if len(p) < 2 || p[1] != ':' {
		return false
	}
	first := p[0]
	return first >= 'a' && first <= 'z' || first >= 'A' && first <= 'Z'
}

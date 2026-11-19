package globv3

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

var driveRe = regexp.MustCompile(`^[A-Za-z]:`)

// WhyNotClean returns why p is not a clean path, or "" when it is. A clean path
// is non-empty valid UTF-8, relative, has no empty, "." or ".." segment, no
// backslash, no control character or byte order mark, no segment that begins or
// ends with a blank character, and no first segment that begins with an ASCII
// letter and a colon. See "Matching" in 04-grammar.md.
func WhyNotClean(p string) string {
	switch {
	case p == "":
		return "it is empty"
	case !utf8.ValidString(p):
		return "it is not valid UTF-8"
	case hasHidden(p):
		return "it contains a control character or a byte order mark"
	case strings.HasPrefix(p, "/"):
		return "it is absolute"
	case strings.Contains(p, `\`):
		return "it contains a backslash, a path separator on Windows"
	}
	for i, sg := range strings.Split(p, "/") {
		switch {
		case sg == "":
			return "it has an empty segment"
		case sg == "." || sg == "..":
			return fmt.Sprintf("a segment is %q", sg)
		case edgeBlank(sg):
			return fmt.Sprintf("segment %q has leading or trailing whitespace", sg)
		case i == 0 && driveRe.MatchString(sg):
			return "its first segment looks like a Windows drive"
		}
	}
	return ""
}

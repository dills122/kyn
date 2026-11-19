package globv3

import (
	"fmt"
	"sort"
	"strings"
)

// templateVars returns the distinct variables a related template uses, in
// order of first appearance, or an error for anything that is not a variable.
func templateVars(tpl string) ([]string, error) {
	if tpl == "" {
		return nil, fmt.Errorf("related must not be empty")
	}
	if strings.HasPrefix(tpl, "/") {
		return nil, fmt.Errorf("related %q must be a relative path", tpl)
	}
	if strings.HasSuffix(tpl, "/") {
		return nil, fmt.Errorf("related %q must name a file, not end with /", tpl)
	}
	if strings.Contains(tpl, `\`) {
		return nil, fmt.Errorf("related %q contains a backslash", tpl)
	}
	if hasHidden(tpl) {
		return nil, fmt.Errorf("related %q contains a control character or a byte order mark", tpl)
	}
	for i, seg := range strings.Split(tpl, "/") {
		if edgeBlank(seg) {
			return nil, fmt.Errorf("related %q: segment %q has leading or trailing whitespace", tpl, seg)
		}
		if i == 0 && driveRe.MatchString(seg) {
			return nil, fmt.Errorf("related %q: its first segment looks like a Windows drive", tpl)
		}
	}
	if strings.ContainsAny(strings.NewReplacer("{", "", "}", "").Replace(tpl), "*?[]") {
		return nil, fmt.Errorf("related %q contains glob metacharacters; it names exactly one path", tpl)
	}
	for _, seg := range strings.Split(tpl, "/") {
		if seg == ".." || seg == "." {
			return nil, fmt.Errorf("related %q contains \"%s\"", tpl, seg)
		}
	}
	if strings.Contains(tpl, "}{") {
		return nil, fmt.Errorf("related %q has adjacent variables; put literal text between them", tpl)
	}
	rest := varRe.ReplaceAllString(tpl, "")
	if strings.ContainsAny(rest, "{}") {
		return nil, fmt.Errorf("related %q has an unmatched brace", tpl)
	}
	var vars []string
	seen := map[string]bool{}
	for _, m := range varRe.FindAllStringSubmatch(tpl, -1) {
		if !nameRe.MatchString(m[1]) {
			return nil, fmt.Errorf("related %q: %q is not a variable", tpl, "{"+m[1]+"}")
		}
		if !seen[m[1]] {
			seen[m[1]] = true
			vars = append(vars, m[1])
		}
	}
	return vars, nil
}

// wholeSegments requires {dir} and the {**name} capture to fill a whole path
// segment of related, and requires related to contain something else, so it can
// never resolve to nothing. Without this, "{dir}.md" becomes ".md" at the root.
func wholeSegments(tpl, multi string) error {
	bare := 0
	segs := strings.Split(tpl, "/")
	for _, sg := range segs {
		hasSlashy := strings.Contains(sg, "{dir}") || (multi != "" && strings.Contains(sg, "{"+multi+"}"))
		if !hasSlashy {
			bare++
			continue
		}
		if sg != "{dir}" && sg != "{"+multi+"}" {
			return fmt.Errorf("related %q: {dir} and {**name} must fill a whole path segment, because either can be empty", tpl)
		}
	}
	if bare == 0 {
		return fmt.Errorf("related %q is made only of {dir} and {**name}, so it can resolve to nothing; add a file name", tpl)
	}
	return nil
}

func sortedStrings(s []string) []string {
	c := append([]string(nil), s...)
	sort.Strings(c)
	return c
}

func containsString(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func sameBindings(a, b Bindings) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || w != v {
			return false
		}
	}
	return true
}

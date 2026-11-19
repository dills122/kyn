package globv3

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

var varRe = regexp.MustCompile(`\{([^{}]*)\}`)

// Selector is the match and exclude globs of one rule, checked together.
type Selector struct {
	match   []*Glob
	exclude []*Glob
	caps    []Capture
}

// NewSelector compiles a rule's match and exclude globs. Every match glob must
// bind the same capture names, each of the same kind; an exclude has no
// captures and may not be identical to a match glob.
func NewSelector(match, exclude []string) (*Selector, error) {
	if len(match) == 0 {
		return nil, fmt.Errorf("match must have at least one glob")
	}
	s := &Selector{}
	for i, g := range match {
		cg, err := Compile(g, true)
		if err != nil {
			return nil, err
		}
		if i == 0 {
			s.caps = cg.Captures()
		} else if strings.Join(sortedStrings(cg.names()), ",") != strings.Join(sortedStrings(s.match[0].names()), ",") {
			return nil, fmt.Errorf("every match glob must bind the same capture names (%v versus %v)", s.match[0].names(), cg.names())
		} else if strings.Join(sortedStrings(cg.kinds()), ",") != strings.Join(sortedStrings(s.match[0].kinds()), ",") {
			return nil, fmt.Errorf("every match glob must use the same kind for a capture name: {**name} spans segments, {name} does not (%v versus %v)", s.match[0].kinds(), cg.kinds())
		}
		s.match = append(s.match, cg)
	}
	for _, g := range exclude {
		cg, err := Compile(g, false)
		if err != nil {
			return nil, err
		}
		for _, m := range match {
			if m == g {
				return nil, fmt.Errorf("exclude %q is identical to a match glob: delete that match glob instead", g)
			}
		}
		s.exclude = append(s.exclude, cg)
	}
	return s, nil
}

// Captures returns the variables the match globs bind, in the first glob's order.
func (s *Selector) Captures() []Capture { return append([]Capture(nil), s.caps...) }

// Bindings maps each variable, including the built-in dir, to its value.
type Bindings map[string]string

// Bind applies exclude, then every match glob, and returns the variables path
// binds. ok is false when the path is excluded or unmatched. A path that two
// match globs bind differently is an error (rule 9).
func (s *Selector) Bind(p string) (b Bindings, ok bool, err error) {
	for _, e := range s.exclude {
		if _, hit := e.Bind(p); hit {
			return nil, false, nil
		}
	}
	var first Bindings
	for _, g := range s.match {
		vals, hit := g.Bind(p)
		if !hit {
			continue
		}
		v := Bindings{"dir": path.Dir(p)}
		if v["dir"] == "." {
			v["dir"] = ""
		}
		for i, c := range g.captures {
			v[c.Name] = vals[i]
		}
		if first == nil {
			first = v
		} else if !sameBindings(first, v) {
			return nil, false, fmt.Errorf("%q matches two match globs with different bindings; make the globs not overlap", p)
		}
	}
	return first, first != nil, nil
}

// Template is a validated related path template.
type Template struct {
	src  string
	vars []string
}

// ParseTemplate validates a related template against the rule's selector: the
// template is a clean-looking relative path with no glob characters, uses only
// {dir} and captured names, never two variables side by side, at most one
// variable that can contain a slash, whole segments for {dir} and {**name}, and
// every capture is used. A template with no variables must not match its own
// rule's selector.
func ParseTemplate(tpl string, s *Selector) (*Template, error) {
	vars, err := templateVars(tpl)
	if err != nil {
		return nil, err
	}
	names := map[string]bool{}
	multi := ""
	for _, c := range s.caps {
		names[c.Name] = true
		if c.Multi {
			multi = c.Name
		}
	}
	slashy := 0
	for _, v := range vars {
		if v != "dir" && !names[v] {
			return nil, fmt.Errorf("related uses {%s} but match does not capture it", v)
		}
		if v == "dir" || v == multi {
			slashy++
		}
	}
	if slashy > 1 {
		return nil, fmt.Errorf("related uses more than one variable that can contain /; use at most one of {dir} and the {**name} capture")
	}
	if err := wholeSegments(tpl, multi); err != nil {
		return nil, err
	}
	for _, c := range s.caps {
		if !containsString(vars, c.Name) {
			return nil, fmt.Errorf("capture {%s} is never used in related", c.Name)
		}
	}
	t := &Template{src: tpl, vars: vars}
	if len(vars) == 0 {
		rel, err := t.Resolve(nil)
		if err != nil {
			return nil, err
		}
		if _, self, _ := s.Bind(rel); self {
			return nil, fmt.Errorf("related path %q also matches this rule's own match; add an exclude or narrow the match", rel)
		}
	}
	return t, nil
}

// Vars returns the distinct variables the template uses, in order of first
// appearance. They are exactly the instance key (03-instances-and-reports.md).
func (t *Template) Vars() []string { return append([]string(nil), t.vars...) }

// Resolve substitutes every variable in one pass (a value is never rescanned),
// drops empty segments, and requires the result to be a clean path (rule 11).
func (t *Template) Resolve(b Bindings) (string, error) {
	s := varRe.ReplaceAllStringFunc(t.src, func(m string) string { return b[m[1:len(m)-1]] })
	var segs []string
	for _, sg := range strings.Split(s, "/") {
		if sg != "" {
			segs = append(segs, sg)
		}
	}
	if len(segs) == 0 {
		return "", fmt.Errorf("related %q resolves to an invalid path: it is empty", t.src)
	}
	out := strings.Join(segs, "/")
	if why := WhyNotClean(out); why != "" {
		return "", fmt.Errorf("related %q resolves to an invalid path %q: %s", t.src, out, why)
	}
	return out, nil
}

// CheckSelfMatch is rule 10: a resolved related path may not match its own
// rule's match globs (after exclude).
func (s *Selector) CheckSelfMatch(related string) error {
	if _, self, err := s.Bind(related); err != nil {
		return err
	} else if self {
		return fmt.Errorf("related path %q also matches this rule's own match; add an exclude or narrow the match", related)
	}
	return nil
}

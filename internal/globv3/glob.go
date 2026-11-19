// Package globv3 is a prototype of the v3 matcher specified in
// docs/design/v3/04-grammar.md: its own glob dialect, captures, longest-first
// binding, related templates and the clean-path predicate. It is not wired into
// the CLI and does not replace internal/matcher.
//
// Binding is defined by enumerating every way a glob can match a path, so the
// implementation here is its own reference. It is total only because Compile
// rejects a capture between two ** segments.
package globv3

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var nameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type tokKind int

const (
	tLit      tokKind = iota // one literal character
	tStar                    // *
	tQ                       // ?
	tAlt                     // {a,b}: literal alternatives
	tCap                     // {name}: one or more characters inside a segment
	tSegs                    // ** in the middle: zero or more whole directories
	tSegsTail                // a trailing **: one or more segments
	tMCap                    // {**name} in the middle: zero or more directories, captured
	tMCapTail                // a trailing {**name}: one or more segments, captured
)

type token struct {
	k    tokKind
	s    string
	alts []string
}

// Capture is a variable a glob binds: {name}, or {**name} when Multi.
type Capture struct {
	Name  string
	Multi bool
}

type Glob struct {
	toks     []token
	captures []Capture // in order of appearance
}

func (g *Glob) multiName() string {
	for _, c := range g.captures {
		if c.Multi {
			return c.Name
		}
	}
	return ""
}

func (g *Glob) names() []string {
	var n []string
	for _, c := range g.captures {
		n = append(n, c.Name)
	}
	return n
}

func (g *Glob) kinds() []string {
	var n []string
	for _, c := range g.captures {
		if c.Multi {
			n = append(n, "**"+c.Name)
		} else {
			n = append(n, c.Name)
		}
	}
	return n
}

// Compile parses a match/exclude glob in the v3 dialect (04-grammar.md,
// "Globs") and returns a config error for every construct outside it. Text is
// read as Unicode code points: `?` is one code point, and "longest" counts them.
func Compile(g string, allowCaptures bool) (*Glob, error) {
	bad := func(format string, a ...any) (*Glob, error) {
		return nil, fmt.Errorf("glob %q: "+format, append([]any{g}, a...)...)
	}
	if g == "" {
		return bad("is empty")
	}
	if !utf8.ValidString(g) {
		return bad("is not valid UTF-8")
	}
	if hasHidden(g) {
		return bad("contains a control character or a byte order mark")
	}
	if strings.HasPrefix(g, "/") {
		return bad("must be relative to the repository root")
	}
	if driveRe.MatchString(strings.SplitN(g, "/", 2)[0]) {
		return bad("its first segment begins with a letter and a colon; no clean path matches it")
	}
	if strings.ContainsAny(g, "[]") {
		return bad("character classes ([ ]) are not supported")
	}
	if strings.Contains(g, `\`) {
		return bad("backslash escapes are not supported")
	}
	// A / inside a brace group can never be part of an alternative or a capture
	// name. Report it by name instead of letting the split below call it unclosed.
	depth := 0
	for i, c := range g {
		switch c {
		case '{':
			depth++
		case '}':
			if depth > 0 {
				depth--
			}
		case '/':
			if depth > 0 && strings.Contains(g[i:], "}") {
				return bad("a / inside {} is not supported: alternatives and capture names cannot contain a path separator")
			}
		}
	}
	out := &Glob{}
	seen := map[string]bool{}
	capture := func(name string, multi bool) error {
		if !allowCaptures {
			return fmt.Errorf("glob %q: captures are not allowed in exclude", g)
		}
		if !nameRe.MatchString(name) {
			return fmt.Errorf("glob %q: %q is not a valid capture name (ASCII letters, digits and underscore, not starting with a digit)", g, name)
		}
		if name == "dir" {
			return fmt.Errorf("glob %q: dir is reserved and cannot be captured", g)
		}
		if seen[name] {
			return fmt.Errorf("glob %q: capture %q appears twice", g, name)
		}
		if multi && out.multiName() != "" {
			return fmt.Errorf("glob %q: at most one {**name} capture per glob", g)
		}
		seen[name] = true
		out.captures = append(out.captures, Capture{name, multi})
		return nil
	}

	segs := strings.Split(g, "/")
	for si, seg := range segs {
		last := si == len(segs)-1
		switch {
		case seg == "":
			return bad("has an empty path segment")
		case edgeBlank(seg):
			return bad("segment %q has leading or trailing whitespace", seg)
		case seg == "." || seg == "..":
			return bad("must not contain a %q segment", seg)
		case seg == "**":
			if last {
				out.toks = append(out.toks, token{k: tSegsTail})
			} else {
				out.toks = append(out.toks, token{k: tSegs})
			}
			continue
		case strings.Contains(seg, "{**"):
			if !strings.HasPrefix(seg, "{**") || strings.IndexByte(seg, '}') != len(seg)-1 {
				return bad("{**name} must fill whole path segments")
			}
			name := seg[3 : len(seg)-1]
			if strings.Contains(name, ",") {
				return bad("alternatives must be literal text: * is not allowed inside {}")
			}
			if err := capture(name, true); err != nil {
				return nil, err
			}
			if last {
				out.toks = append(out.toks, token{k: tMCapTail, s: name})
			} else {
				out.toks = append(out.toks, token{k: tMCap, s: name})
			}
			continue
		case strings.Contains(seg, "**"):
			return bad("** must be a whole path segment")
		}

		var st []token
		for i := 0; i < len(seg); {
			c, size := utf8.DecodeRuneInString(seg[i:])
			switch c {
			case '*':
				st = append(st, token{k: tStar})
				i++
			case '?':
				st = append(st, token{k: tQ})
				i++
			case '}':
				return bad("unmatched }")
			case '{':
				end := strings.IndexByte(seg[i:], '}')
				if end < 0 {
					return bad("unclosed {")
				}
				end += i
				body := seg[i+1 : end]
				if strings.IndexFunc(body, isBlank) >= 0 {
					return bad("whitespace inside {}: %q", "{"+body+"}")
				}
				if strings.Contains(body, "{") {
					return bad("nested braces are not supported")
				}
				if strings.Contains(body, ",") { // alternation
					parts := strings.Split(body, ",")
					for _, p := range parts {
						if p == "" {
							return bad("alternation has an empty alternative")
						}
						if strings.ContainsAny(p, "*?") {
							return bad("alternatives must be literal text, but %q has a wildcard", p)
						}
					}
					st = append(st, token{k: tAlt, alts: parts})
				} else {
					if err := capture(body, false); err != nil {
						return nil, err
					}
					st = append(st, token{k: tCap, s: body})
				}
				i = end + 1
			default:
				st = append(st, token{k: tLit, s: string(c)})
				i += size
			}
		}
		for i, t := range st {
			if t.k != tCap {
				continue
			}
			for _, j := range []int{i - 1, i + 1} {
				if j < 0 || j >= len(st) {
					continue
				}
				switch st[j].k {
				case tCap:
					return bad("adjacent captures need literal text between them")
				case tStar, tQ, tAlt:
					return bad("a capture cannot touch *, ? or an alternation; put literal text between them")
				}
			}
		}
		if dots := dotSegment(st); dots != "" {
			return bad("a segment can expand to %q through its alternatives; globs name repository paths", dots)
		}
		out.toks = append(out.toks, st...)
		if !last {
			out.toks = append(out.toks, token{k: tLit, s: "/"})
		}
	}
	// A {name} capture with ** or {**x} on both sides would bind by
	// directory-name length, so renaming an unrelated directory could move
	// a file to another instance.
	for i, t := range out.toks {
		if t.k != tCap && t.k != tMCap {
			continue
		}
		before, after := false, false
		for _, u := range out.toks[:i] {
			before = before || u.k == tSegs || u.k == tMCap
		}
		for _, u := range out.toks[i+1:] {
			after = after || u.k == tSegs || u.k == tSegsTail || u.k == tMCap || u.k == tMCapTail
		}
		if before && after {
			return bad("capture %q sits between two ** segments, so it would bind by directory-name length; put ** or {**name} on one side only", t.s)
		}
	}
	return out, nil
}

// dotSegment reports "." or ".." if a segment made only of literals and
// alternatives can expand to it, so `{..,a}` is rejected like a literal `..`.
func dotSegment(st []token) string {
	cur := map[string]bool{"": true}
	for _, t := range st {
		next := map[string]bool{}
		for s := range cur {
			switch t.k {
			case tLit:
				next[s+t.s] = true
			case tAlt:
				for _, a := range t.alts {
					next[s+a] = true
				}
			default:
				return ""
			}
		}
		for s := range next {
			if len(s) > 2 {
				delete(next, s) // longer than ".." can never be one
			}
		}
		cur = next
	}
	for _, d := range []string{".", ".."} {
		if cur[d] {
			return d
		}
	}
	return ""
}

// A blank character is Unicode White_Space or a control character (Cc). Format
// characters (Cf, such as a zero-width joiner) are not blank, so names that use
// them are legal. Controls and the BOM are rejected anywhere in a glob, a related
// template or an input path; a segment may not begin or end with a blank
// character; and none may appear inside {}.
func isBlank(r rune) bool {
	return unicode.IsSpace(r) || unicode.IsControl(r)
}

func hasHidden(s string) bool {
	for _, r := range s {
		if r == 0xFEFF || unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func edgeBlank(seg string) bool {
	f, _ := utf8.DecodeRuneInString(seg)
	l, _ := utf8.DecodeLastRuneInString(seg)
	return seg != "" && (isBlank(f) || isBlank(l))
}

// parses enumerates every way the glob can match p, as the captured values.
type capv struct {
	s     string
	start int
}

func (g *Glob) parses(p string) [][]capv {
	var all [][]capv
	var caps []capv
	var walk func(ti, pos int)
	walk = func(ti, pos int) {
		if ti == len(g.toks) {
			if pos == len(p) {
				all = append(all, append([]capv(nil), caps...))
			}
			return
		}
		t := g.toks[ti]
		switch t.k {
		case tLit:
			if strings.HasPrefix(p[pos:], t.s) {
				walk(ti+1, pos+len(t.s))
			}
		case tStar:
			for end := pos; ; {
				walk(ti+1, end)
				if end >= len(p) || p[end] == '/' {
					break
				}
				_, sz := utf8.DecodeRuneInString(p[end:])
				end += sz
			}
		case tQ:
			if pos < len(p) && p[pos] != '/' {
				_, sz := utf8.DecodeRuneInString(p[pos:])
				walk(ti+1, pos+sz)
			}
		case tAlt:
			for _, a := range t.alts {
				if strings.HasPrefix(p[pos:], a) {
					walk(ti+1, pos+len(a))
				}
			}
		case tCap:
			for end := pos; end < len(p) && p[end] != '/'; {
				_, sz := utf8.DecodeRuneInString(p[end:])
				end += sz
				caps = append(caps, capv{p[pos:end], pos})
				walk(ti+1, end)
				caps = caps[:len(caps)-1]
			}
		case tSegs, tMCap:
			push := func(v string) {
				if t.k == tMCap {
					caps = append(caps, capv{v, pos})
				}
			}
			pop := func() {
				if t.k == tMCap {
					caps = caps[:len(caps)-1]
				}
			}
			push("")
			walk(ti+1, pos)
			pop()
			cur := pos
			for {
				idx := strings.IndexByte(p[cur:], '/')
				if idx <= 0 {
					break
				}
				cur += idx + 1
				push(p[pos : cur-1])
				walk(ti+1, cur)
				pop()
			}
		case tSegsTail:
			if pos < len(p) {
				walk(ti+1, len(p))
			}
		case tMCapTail:
			if pos < len(p) {
				caps = append(caps, capv{p[pos:], pos})
				walk(ti+1, len(p))
				caps = caps[:len(caps)-1]
			}
		}
	}
	walk(0, 0)
	return all
}

// cmpLen compares two parses by capture length in code points, first capture
// first: 1 if a is longer, -1 if b is, 0 if every capture is the same length.
func cmpLen(a, b []capv) int {
	for i := range a {
		la, lb := utf8.RuneCountInString(a[i].s), utf8.RuneCountInString(b[i].s)
		if la != lb {
			if la > lb {
				return 1
			}
			return -1
		}
	}
	return 0
}

// Bind is the reference matcher: it enumerates every way the glob can match p
// and returns the parse whose captures are longest, first capture first. This is
// the normative definition of binding in 04-grammar.md; an implementation may
// use any algorithm that returns the same answer. It is total only because
// Compile rejects a capture between two ** segments: `-ties` checks that no
// accepted glob has two longest parses that bind different values.
func (g *Glob) Bind(p string) ([]string, bool) {
	var best []capv
	found := false
	for _, c := range g.parses(p) {
		if !found {
			best, found = c, true
		} else if cmpLen(c, best) > 0 {
			best = c
		}
	}
	var out []string
	for _, c := range best {
		out = append(out, c.s)
	}
	return out, found
}

// Captures returns the variables the glob binds, in order of appearance.
func (g *Glob) Captures() []Capture { return append([]Capture(nil), g.captures...) }

// src reconstructs a printable form of the glob for test failures.
func (g *Glob) src() string {
	var sb strings.Builder
	for _, t := range g.toks {
		switch t.k {
		case tLit:
			sb.WriteString(t.s)
		case tStar:
			sb.WriteString("*")
		case tQ:
			sb.WriteString("?")
		case tAlt:
			sb.WriteString("{" + strings.Join(t.alts, ",") + "}")
		case tCap:
			sb.WriteString("{" + t.s + "}")
		case tSegs:
			sb.WriteString("**/")
		case tSegsTail:
			sb.WriteString("**")
		case tMCap:
			sb.WriteString("{**" + t.s + "}/")
		case tMCapTail:
			sb.WriteString("{**" + t.s + "}")
		}
	}
	return sb.String()
}

//go:build ignore

// Simulator for the v3 configuration design in docs/design/v3/.
//
// This is NOT Kyn code and shares nothing with internal/. It is a small model
// of the designed semantics: captures in `match`, instance keys derived from
// the variables `related` uses, the three related-path states, `skipIf`,
// `expect`, flags, and the validation rules in 04-grammar.md. It exists so the
// design can be tried against realistic scenarios before anything is built.
//
// Every scenario in scenarios.yaml states its expected outcome, and the
// simulator exits non-zero on any mismatch. The simulator and the scenarios
// were written by the same person, so a green run shows the model is
// internally consistent, not that the design is right.
//
//	go run docs/design/v3/experiments/sim/main.go docs/design/v3/experiments/sim/scenarios.yaml
//	go run docs/design/v3/experiments/sim/main.go -v docs/design/v3/experiments/sim/scenarios.yaml
//
// The build tag keeps it out of `go build ./...`, `go vet ./...` and
// `go test ./...`.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// ---- configuration ----------------------------------------------------------

type strs []string

func (s *strs) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		*s = []string{n.Value}
		return nil
	}
	var l []string
	if err := n.Decode(&l); err != nil {
		return err
	}
	*s = l
	return nil
}

type Rule struct {
	Match      strs    `yaml:"match"`
	Exclude    strs    `yaml:"exclude"`
	Related    *string `yaml:"related"`
	Expect     strs    `yaml:"expect"`
	SkipIf     string  `yaml:"skipIf"`
	ChangeType strs    `yaml:"changeType"`
	Severity   string  `yaml:"severity"`
	Message    string  `yaml:"message"`
	Why        string  `yaml:"why"`
	Flags      strs    `yaml:"flags"`
}

type Config struct {
	Version int             `yaml:"version"`
	Rules   map[string]Rule `yaml:"rules"`
}

type Scenario struct {
	Name    string   `yaml:"name"`
	Config  string   `yaml:"config"`
	Disk    []string `yaml:"disk"`    // files present after the change
	Changes []string `yaml:"changes"` // git name-status lines, or bare paths when mode is plain
	Mode    string   `yaml:"mode"`    // "" (git) or "plain"
	FailOn  string   `yaml:"failOn"`  // "" (error) or "warn"
	Want    Want     `yaml:"want"`
}

type Want struct {
	Error    string   `yaml:"error"`    // substring of the expected config error
	Results  []string `yaml:"results"`  // "STATUS rule instance", in output order
	Headline string   `yaml:"headline"` // FAIL | NON-BLOCKING | PASS
	Flags    []string `yaml:"flags"`
}

// ---- glob compilation -------------------------------------------------------

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

type capInfo struct {
	name  string
	multi bool
}

type glob struct {
	toks     []token
	captures []capInfo // in order of appearance
}

func (g *glob) multiName() string {
	for _, c := range g.captures {
		if c.multi {
			return c.name
		}
	}
	return ""
}

func (g *glob) names() []string {
	var n []string
	for _, c := range g.captures {
		n = append(n, c.name)
	}
	return n
}

func (g *glob) kinds() []string {
	var n []string
	for _, c := range g.captures {
		if c.multi {
			n = append(n, "**"+c.name)
		} else {
			n = append(n, c.name)
		}
	}
	return n
}

// compileGlob parses a match/exclude glob in the v3 dialect (04-grammar.md,
// "Globs") and returns a config error for every construct outside it. Text is
// read as Unicode code points: `?` is one code point, and "longest" counts them.
func compileGlob(g string, allowCaptures bool) (*glob, error) {
	bad := func(format string, a ...any) (*glob, error) {
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
	out := &glob{}
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
		out.captures = append(out.captures, capInfo{name, multi})
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

// A blank character is Unicode White_Space, a control character (Cc) or a
// format character (Cf: zero-width space and joiner, BOM, and the like).
// Controls and the BOM are rejected anywhere in a glob, a related template or an
// input path; a segment may not begin or end with a blank character; and none may
// appear inside {}. A format character in the middle of a segment is legal
// (emoji sequences need the joiner).
func isBlank(r rune) bool {
	return unicode.IsSpace(r) || unicode.IsControl(r) || unicode.Is(unicode.Cf, r)
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

func (g *glob) parses(p string) [][]capv {
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

// bind is the reference matcher: it enumerates every way the glob can match p
// and returns the parse whose captures are longest, first capture first. This is
// the normative definition of binding in 04-grammar.md; an implementation may
// use any algorithm that returns the same answer. It is total only because
// compileGlob rejects a capture between two ** segments: `-ties` checks that no
// accepted glob has two longest parses that bind different values.
func (g *glob) bind(p string) ([]string, bool) {
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

// ---- related templates ------------------------------------------------------

var varRe = regexp.MustCompile(`\{([^{}]*)\}`)

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

// resolveRelated substitutes every variable in one pass (a captured value is
// never rescanned), drops empty segments, and rejects a result that is empty or
// has a "." or ".." segment. Rule 11 in 04-grammar.md.
func resolveRelated(tpl string, vals map[string]string) (string, error) {
	s := varRe.ReplaceAllStringFunc(tpl, func(m string) string { return vals[m[1:len(m)-1]] })
	var segs []string
	for _, sg := range strings.Split(s, "/") {
		if sg != "" {
			segs = append(segs, sg)
		}
	}
	if len(segs) == 0 {
		return "", fmt.Errorf("related %q resolves to an invalid path: it is empty", tpl)
	}
	out := strings.Join(segs, "/")
	if why := badPath(out); why != "" {
		return "", fmt.Errorf("related %q resolves to an invalid path %q: %s", tpl, out, why)
	}
	return out, nil
}

// badPath is the one predicate for a clean repository path, used for input
// paths and for resolved related paths (rule 11). It returns why a path is not
// clean, or "".
func badPath(p string) string {
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

var driveRe = regexp.MustCompile(`^[A-Za-z]:`)

// badInputPath adds git's quoted form to badPath for paths read from input.
func badInputPath(p string) string {
	if strings.HasPrefix(p, `"`) {
		return "it is in git's quoted form; feed git -c core.quotePath=false"
	}
	return badPath(p)
}

// ---- evaluation -------------------------------------------------------------

type Res struct{ Status, Rule, Instance, Detail string }

type Outcome struct {
	Err      string
	Results  []Res
	Headline string
	Flags    []string
	Counts   map[string]int
	Trace    []string
}

func cfgErr(format string, a ...any) Outcome {
	return Outcome{Err: fmt.Sprintf(format, a...)}
}

var validExpect = map[string]bool{"touched": true, "untouched": true, "exists": true, "missing": true, "none": true}
var validSeverity = map[string]bool{"": true, "info": true, "warn": true, "error": true}

type compiled struct {
	id       string
	r        Rule
	match    []*glob
	exclude  []*glob
	vars     []string
	sev      string
	expect   []string
	hasRelat bool
	related  string
}

// bind applies exclude, then every match glob, and returns the variables a path
// binds. A path two globs bind differently is rule 9.
func (c *compiled) bind(p string) (map[string]string, bool, error) {
	for _, e := range c.exclude {
		if _, ok := e.bind(p); ok {
			return nil, false, nil
		}
	}
	var first map[string]string
	for _, g := range c.match {
		m, ok := g.bind(p)
		if !ok {
			continue
		}
		v := map[string]string{"dir": path.Dir(p)}
		if v["dir"] == "." {
			v["dir"] = ""
		}
		for i, ci := range g.captures {
			v[ci.name] = m[i]
		}
		if first == nil {
			first = v
		} else if !equal(first, v) {
			return nil, false, fmt.Errorf("rule %q: %q matches two match globs with different bindings; make the globs not overlap", c.id, p)
		}
	}
	return first, first != nil, nil
}

type inst struct {
	name    string
	vars    map[string]string
	sources []string
	status  map[string]bool
}

func evaluate(sc Scenario) Outcome {
	var cfg Config
	dec := yaml.NewDecoder(strings.NewReader(sc.Config))
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return cfgErr("%v", err)
	}
	if cfg.Version != 3 {
		return cfgErr("version must be 3")
	}

	disk := map[string]bool{}
	for _, f := range sc.Disk {
		disk[f] = true
	}
	touched := map[string]string{} // path -> added|modified|renamed
	vanished := map[string]bool{}
	subject := map[string]string{} // path -> the line that named it
	note := func(p, l string) error {
		if why := badInputPath(p); why != "" {
			return fmt.Errorf("invalid input path %q: %s", p, why)
		}
		if prev, ok := subject[p]; ok && prev != l {
			return fmt.Errorf("path %q appears on more than one line (%q and %q); lines must not conflict", p, prev, l)
		}
		subject[p] = l
		return nil
	}
	fromPlain := map[string]bool{} // sources that came from a plain-path line
	statusRe := regexp.MustCompile(`^([AMDT]|R[0-9]{1,3})$`)
	// A TAB-less line that starts with a status token and a space is almost
	// certainly a name-status line whose TAB became spaces, or git status output.
	statusLike := regexp.MustCompile(`^([AMDTUCX]|R[0-9]+|\?\?|!!) `)
	for _, l := range sc.Changes {
		l = strings.TrimSuffix(l, "\r") // CRLF input
		if strings.TrimSpace(l) == "" {
			continue // empty and whitespace-only lines are ignored
		}
		// Each line is read on its own: a TAB makes it a name-status line, no TAB a
		// plain path. mode "plain" and "mixed" treat a TAB-less line as a path.
		if !strings.Contains(l, "\t") && (sc.Mode == "plain" || sc.Mode == "mixed") {
			if statusLike.MatchString(l) {
				return cfgErr("line %q looks like a status line with spaces instead of a TAB; a plain path may not start with a status token and a space", l)
			}
			if err := note(l, l); err != nil {
				return cfgErr("%v", err)
			}
			touched[l] = "modified"
			fromPlain[l] = true
			continue
		}
		// Scenarios may write a space instead of a TAB when no path contains one.
		f := strings.Split(l, "\t")
		if len(f) == 1 {
			f = strings.Fields(l)
		}
		if len(f) == 0 || f[0] == "" {
			return cfgErr("empty status on line %q", l)
		}
		if sc := f[0]; len(sc) > 1 && sc[0] == 'R' && statusRe.MatchString(sc) {
			n := 0
			fmt.Sscan(sc[1:], &n)
			if n > 100 {
				return cfgErr("unrecognised status %q on line %q: a rename score is 0 to 100", sc, l)
			}
		}
		if !statusRe.MatchString(f[0]) {
			if f[0] == "U" {
				return cfgErr("unmerged path on line %q: resolve the merge first", l)
			}
			return cfgErr("unrecognised status %q on line %q: want A, M, D, T or R followed by a score of 0 to 100", f[0], l)
		}
		want := 2
		if f[0][0] == 'R' {
			want = 3
		}
		if len(f) != want {
			return cfgErr("line %q has %d fields, want %d for status %s", l, len(f), want, f[0])
		}
		for _, q := range f[1:] {
			if err := note(q, l); err != nil {
				return cfgErr("%v", err)
			}
		}
		switch f[0][0] {
		case 'A':
			touched[f[1]] = "added"
		case 'M', 'T':
			touched[f[1]] = "modified"
		case 'D':
			vanished[f[1]] = true
		case 'R':
			vanished[f[1]] = true
			touched[f[2]] = "renamed"
		}
	}

	ids := make([]string, 0, len(cfg.Rules))
	for id := range cfg.Rules {
		ids = append(ids, id)
	}
	sort.Strings(ids) // rule IDs are always processed in sorted order

	// ---- validation, per rule, before any evaluation ----
	var rules []*compiled
	for _, id := range ids {
		r := cfg.Rules[id]
		c := &compiled{id: id, r: r}
		fail := func(format string, a ...any) Outcome {
			return cfgErr("rule %q: "+format, append([]any{id}, a...)...)
		}
		if len(r.Match) == 0 {
			return fail("match is required")
		}
		if len(r.Expect) == 0 {
			return fail("expect is required")
		}
		for _, e := range r.Expect {
			if !validExpect[e] {
				return fail("expect %q is not one of touched, untouched, exists, missing, none", e)
			}
		}
		set := map[string]bool{}
		for _, e := range r.Expect {
			set[e] = true
		}
		if set["none"] && len(set) > 1 {
			return fail("expect: none cannot be combined with other values")
		}
		if set["touched"] && set["untouched"] {
			return fail("expect lists touched and untouched, which contradict")
		}
		if set["exists"] && set["missing"] {
			return fail("expect lists exists and missing, which contradict")
		}
		for e := range set {
			c.expect = append(c.expect, e)
		}
		sort.Strings(c.expect)
		none := set["none"]

		if r.SkipIf != "" && r.SkipIf != "missing" && r.SkipIf != "exists" {
			return fail("skipIf %q is not one of missing, exists", r.SkipIf)
		}
		for _, ct := range r.ChangeType {
			if ct != "added" && ct != "modified" && ct != "renamed" {
				return fail("changeType %q is not one of added, modified, renamed", ct)
			}
		}
		if !validSeverity[r.Severity] {
			return fail("severity %q is not one of info, warn, error", r.Severity)
		}
		c.sev = r.Severity
		if c.sev == "" {
			c.sev = "error"
			if none {
				c.sev = "info"
			}
		}
		if none {
			if c.sev != "info" {
				return fail("expect: none requires severity: info")
			}
			if r.Message == "" && len(r.Flags) == 0 {
				return fail("expect: none requires at least one of message or flags")
			}
		}
		c.hasRelat = r.Related != nil
		if c.hasRelat {
			c.related = *r.Related
		}
		if !c.hasRelat && !none {
			return fail("related is required unless expect is none")
		}
		if !c.hasRelat && r.SkipIf != "" {
			return fail("skipIf needs a related path")
		}
		if c.hasRelat && !none {
			if msg := passability(r.SkipIf, set); msg != "" {
				return fail("skipIf %q with expect %v %s", r.SkipIf, c.expect, msg)
			}
		}

		var names []string
		for gi, g := range r.Match {
			cg, err := compileGlob(g, true)
			if err != nil {
				return fail("%v", err)
			}
			if gi == 0 {
				names = cg.names()
			} else if strings.Join(sorted(cg.names()), ",") != strings.Join(sorted(names), ",") {
				return fail("every match glob must bind the same capture names (%v versus %v)", names, cg.names())
			} else if strings.Join(sorted(cg.kinds()), ",") != strings.Join(sorted(c.match[0].kinds()), ",") {
				return fail("every match glob must use the same kind for a capture name: {**name} spans segments, {name} does not (%v versus %v)", c.match[0].kinds(), cg.kinds())
			}
			c.match = append(c.match, cg)
		}
		for _, g := range r.Exclude {
			cg, err := compileGlob(g, false)
			if err != nil {
				return fail("%v", err)
			}
			if contains(r.Match, g) {
				return fail("exclude %q is identical to a match glob: delete that match glob instead", g)
			}
			c.exclude = append(c.exclude, cg)
		}
		if c.hasRelat {
			vars, err := templateVars(c.related)
			if err != nil {
				return fail("%v", err)
			}
			c.vars = vars
			slashy := 0
			for _, v := range vars {
				if v != "dir" && !contains(names, v) {
					return fail("related uses {%s} but match does not capture it", v)
				}
				if v == "dir" || v == c.match[0].multiName() {
					slashy++
				}
			}
			if slashy > 1 {
				return fail("related uses more than one variable that can contain /; use at most one of {dir} and the {**name} capture")
			}
			if err := wholeSegments(c.related, c.match[0].multiName()); err != nil {
				return fail("%v", err)
			}
		}
		for _, n := range names {
			if !contains(c.vars, n) {
				return fail("capture {%s} is never used in related", n)
			}
		}
		if c.hasRelat && len(c.vars) == 0 {
			rel, err := resolveRelated(c.related, nil)
			if err != nil {
				return fail("%v", err)
			}
			if _, self, _ := c.bind(rel); self {
				return fail("related path %q also matches this rule's own match; add an exclude or narrow the match", rel)
			}
		}
		rules = append(rules, c)
	}

	// ---- evaluation ----
	out := Outcome{Counts: map[string]int{}}
	flags := map[string]bool{}
	blocked := false
	paths := make([]string, 0, len(touched))
	for p := range touched {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, c := range rules {
		r := c.r
		insts := map[string]*inst{}
		for _, p := range paths {
			if fromPlain[p] && !disk[p] {
				continue // a listed source that is not on disk cannot be matched, as with a deleted one in git mode
			}
			v, ok, err := c.bind(p)
			if err != nil {
				return cfgErr("%v", err)
			}
			if !ok {
				continue
			}
			var key, parts []string
			for _, n := range c.vars {
				key = append(key, v[n])
				if v[n] != "" {
					parts = append(parts, v[n])
				}
			}
			name := strings.Join(parts, "/")
			switch {
			case len(c.vars) == 0:
				name = "(all)"
			case name == "":
				name = "."
			}
			k := strings.Join(key, "\x00")
			in := insts[k]
			if in == nil {
				in = &inst{name: name, vars: v, status: map[string]bool{}}
				insts[k] = in
			}
			in.sources = append(in.sources, p)
			in.status[touched[p]] = true
		}
		keys := make([]string, 0, len(insts))
		for k := range insts {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return insts[keys[i]].name < insts[keys[j]].name })

		for _, k := range keys {
			in := insts[k]
			add := func(status, detail string) {
				out.Results = append(out.Results, Res{status, c.id, in.name, detail})
				out.Counts[status]++
			}
			// Rules 9 to 11 run for every instance, before changeType and skipIf can
			// skip it, so a skipped instance cannot hide a broken rule.
			rel := ""
			if c.hasRelat {
				var err error
				if rel, err = resolveRelated(c.related, in.vars); err != nil {
					return cfgErr("rule %q: %v", c.id, err)
				}
				if _, self, err := c.bind(rel); err != nil {
					return cfgErr("%v", err)
				} else if self {
					return cfgErr("rule %q: related path %q also matches this rule's own match; add an exclude or narrow the match", c.id, rel)
				}
			}
			if len(r.ChangeType) > 0 {
				ok := false
				for _, ct := range r.ChangeType {
					ok = ok || in.status[ct]
				}
				if !ok {
					add("skipped", "no source has a listed changeType")
					continue
				}
			}
			if !c.hasRelat { // informational rule with no related path
				for _, f := range r.Flags {
					flags[f] = true
				}
				add("info", "")
				continue
			}

			state := "missing"
			if disk[rel] {
				state = "exists"
			} else if vanished[rel] {
				state = "removed"
			}
			if (r.SkipIf == "missing" && state == "missing") || (r.SkipIf == "exists" && state == "exists") {
				add("skipped", "related is "+state)
				continue
			}

			status, detail := "pass", ""
			_, isTouched := touched[rel]
			for _, e := range c.expect {
				bad := ""
				switch e {
				case "touched":
					switch {
					case isTouched:
					case state == "removed":
						bad = "removed by this change"
					case state == "missing":
						bad = "does not exist"
					default:
						bad = "not touched"
					}
				case "untouched":
					switch {
					case isTouched:
						bad = "was touched"
					case state == "removed":
						bad = "removed by this change"
					}
				case "exists":
					if state != "exists" {
						bad = "does not exist"
						if state == "removed" {
							bad = "removed by this change"
						}
					}
				case "missing":
					if state == "exists" {
						bad = "exists but must not"
					}
				case "none":
					status = "info"
				}
				if bad != "" {
					status = "fail"
					if detail == "" {
						detail = bad
					}
				}
			}
			for _, f := range r.Flags {
				flags[f] = true
			}
			add(status, detail+" ["+state+"]")
			if status == "fail" {
				if c.sev == "error" || (sc.FailOn == "warn" && c.sev == "warn") {
					blocked = true
				}
			}
		}
	}

	out.Headline = "PASS"
	if blocked {
		out.Headline = "FAIL"
	} else if out.Counts["fail"] > 0 {
		out.Headline = "NON-BLOCKING"
	}
	for f := range flags {
		out.Flags = append(out.Flags, f)
	}
	sort.Strings(out.Flags)
	return out
}

// passability is rule 7. It enumerates the five reachable (state, touched)
// pairs, drops the ones skipIf skips, and reports a rule that can never pass,
// or that can pass only on the pair where the diff and the working tree disagree.
func passability(skipIf string, set map[string]bool) string {
	type pair struct {
		state   string
		touched bool
	}
	pairs := []pair{{"exists", true}, {"exists", false}, {"removed", false}, {"missing", false}, {"missing", true}}
	var passing []pair
	applicable := 0
	for _, p := range pairs {
		if (skipIf == "missing" && p.state == "missing") || (skipIf == "exists" && p.state == "exists") {
			continue
		}
		applicable++
		ok := true
		if set["touched"] && !p.touched {
			ok = false
		}
		if set["untouched"] && (p.touched || p.state == "removed") {
			ok = false
		}
		if set["exists"] && p.state != "exists" {
			ok = false
		}
		if set["missing"] && p.state == "exists" {
			ok = false
		}
		if ok {
			passing = append(passing, p)
		}
	}
	if len(passing) == 0 {
		return "can never pass: no state the rule applies to satisfies every expectation"
	}
	if len(passing) == applicable {
		return "can never fail: every state the rule applies to satisfies the expectation; use expect: none with flags to report it"
	}
	if len(passing) == 1 && passing[0] == (pair{"missing", true}) {
		return "can pass only when the diff and the working tree disagree (the related file is in the diff but not on disk)"
	}
	return ""
}

func sorted(s []string) []string { c := append([]string(nil), s...); sort.Strings(c); return c }
func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
func equal(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// ---- driver -----------------------------------------------------------------

// ties generates random globs, keeps those compileGlob accepts, and looks for a
// path with two parses whose captures are equally long but bind different
// values. Binding is total only if there are none.
func ties(n int) {
	rnd := rand.New(rand.NewSource(1))
	segTok := []string{"a", "b", "é", "-", ".", "*", "?", "{x}", "{y}", "{z}", "{ts,html}", "{a,b}", "a{x}b", "{x}.{y}", "{x}-b", "a-{y}"}
	special := []string{"**", "**", "{**p}"}
	alpha := []string{"a", "b", "é", "-", "."}
	acc, matched, bad := 0, 0, 0
	for i := 0; i < n; i++ {
		var segs []string
		for j, k := 0, 1+rnd.Intn(4); j < k; j++ {
			if rnd.Intn(3) == 0 {
				segs = append(segs, special[rnd.Intn(len(special))])
				continue
			}
			t := ""
			for m, c := 0, 1+rnd.Intn(3); m < c; m++ {
				t += segTok[rnd.Intn(len(segTok))]
			}
			segs = append(segs, t)
		}
		g, err := compileGlob(strings.Join(segs, "/"), true)
		if err != nil {
			continue
		}
		acc++
		for q := 0; q < 40; q++ {
			var ps []string
			for j, k := 0, 1+rnd.Intn(5); j < k; j++ {
				w := ""
				for m, c := 0, 1+rnd.Intn(5); m < c; m++ {
					w += alpha[rnd.Intn(len(alpha))]
				}
				ps = append(ps, w)
			}
			p := strings.Join(ps, "/")
			all := g.parses(p)
			if len(all) > 0 {
				matched++
			}
			var top [][]capv
			for _, c := range all {
				switch {
				case top == nil || cmpLen(c, top[0]) > 0:
					top = [][]capv{c}
				case cmpLen(c, top[0]) == 0:
					top = append(top, c)
				}
			}
			for x := 1; x < len(top); x++ {
				if fmt.Sprint(top[x]) != fmt.Sprint(top[0]) {
					bad++
					if bad <= 5 {
						fmt.Printf("TIE  glob %q path %q: %q\n", strings.Join(segs, "/"), p, top)
					}
					break
				}
			}
		}
	}
	fmt.Printf("%d globs tried, %d accepted, %d matching (glob, path) pairs, %d with two longest parses that bind different values\n", n, acc, matched, bad)
	if bad > 0 {
		os.Exit(1)
	}
}

func main() {
	verbose := false
	args := os.Args[1:]
	if len(args) == 2 && args[0] == "-ties" {
		var n int
		fmt.Sscan(args[1], &n)
		ties(n)
		return
	}
	if len(args) > 0 && args[0] == "-v" {
		verbose, args = true, args[1:]
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		panic(err)
	}
	var scs []Scenario
	if err := yaml.Unmarshal(data, &scs); err != nil {
		panic(err)
	}
	bad := 0
	for _, sc := range scs {
		o := evaluate(sc)
		var problems []string
		switch {
		case sc.Want.Error != "":
			if o.Err == "" {
				problems = append(problems, fmt.Sprintf("wanted config error containing %q, got none", sc.Want.Error))
			} else if !strings.Contains(o.Err, sc.Want.Error) {
				problems = append(problems, fmt.Sprintf("wanted error containing %q, got %q", sc.Want.Error, o.Err))
			}
		case o.Err != "":
			problems = append(problems, "unexpected config error: "+o.Err)
		default:
			var got []string
			for _, r := range o.Results {
				got = append(got, strings.ToUpper(r.Status)+" "+r.Rule+" "+r.Instance)
			}
			if strings.Join(got, "\n") != strings.Join(sc.Want.Results, "\n") {
				problems = append(problems, fmt.Sprintf("results\n      got:  %q\n      want: %q", got, sc.Want.Results))
			}
			if o.Headline != sc.Want.Headline {
				problems = append(problems, fmt.Sprintf("headline got %s want %s", o.Headline, sc.Want.Headline))
			}
			if strings.Join(o.Flags, ",") != strings.Join(sc.Want.Flags, ",") {
				problems = append(problems, fmt.Sprintf("flags got %v want %v", o.Flags, sc.Want.Flags))
			}
		}
		if len(problems) > 0 {
			bad++
			fmt.Printf("MISMATCH  %s\n", sc.Name)
			for _, p := range problems {
				fmt.Println("    " + p)
			}
		} else {
			fmt.Printf("ok        %s\n", sc.Name)
		}
		if verbose {
			for _, l := range strings.Split(strings.TrimRight(sc.Config, "\n"), "\n") {
				fmt.Println("      | " + l)
			}
			if o.Err != "" {
				fmt.Println("      CONFIG ERROR:", o.Err)
			}
			for _, r := range o.Results {
				fmt.Printf("      %-7s %s  %s  %s\n", strings.ToUpper(r.Status), r.Rule, r.Instance, r.Detail)
			}
			if o.Err == "" {
				fmt.Printf("      => %s %v flags=%v\n", o.Headline, o.Counts, o.Flags)
			}
		}
	}
	fmt.Printf("\n%d scenarios, %d mismatched\n", len(scs), bad)
	if bad > 0 {
		os.Exit(1)
	}
}

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
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

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
// "Globs") and returns a config error for every construct outside it.
func compileGlob(g string, allowCaptures bool) (*glob, error) {
	bad := func(format string, a ...any) (*glob, error) {
		return nil, fmt.Errorf("glob %q: "+format, append([]any{g}, a...)...)
	}
	if g == "" {
		return bad("is empty")
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
	out := &glob{}
	seen := map[string]bool{}
	capture := func(name string, multi bool) error {
		if !allowCaptures {
			return fmt.Errorf("glob %q: captures are not allowed in exclude", g)
		}
		if !nameRe.MatchString(name) {
			return fmt.Errorf("glob %q: %q is not a valid capture name (letters, digits and underscore, not starting with a digit)", g, name)
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
			c := seg[i]
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
				i++
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
		out.toks = append(out.toks, st...)
		if !last {
			out.toks = append(out.toks, token{k: tLit, s: "/"})
		}
	}
	return out, nil
}

// bind is the reference matcher: it enumerates every way the glob can match p
// and returns the parse whose captures are longest, comparing the first capture
// first. This is the normative definition of binding in 04-grammar.md; an
// implementation may use any algorithm that returns the same answer.
func (g *glob) bind(p string) ([]string, bool) {
	var best []string
	found := false
	var caps []string
	better := func(a, b []string) bool {
		for i := range a {
			if len(a[i]) != len(b[i]) {
				return len(a[i]) > len(b[i])
			}
		}
		return false
	}
	var walk func(ti, pos int)
	walk = func(ti, pos int) {
		if ti == len(g.toks) {
			if pos == len(p) {
				if !found || better(caps, best) {
					best = append([]string(nil), caps...)
					found = true
				}
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
			for end := pos; ; end++ {
				walk(ti+1, end)
				if end >= len(p) || p[end] == '/' {
					break
				}
			}
		case tQ:
			if pos < len(p) && p[pos] != '/' {
				walk(ti+1, pos+1)
			}
		case tAlt:
			for _, a := range t.alts {
				if strings.HasPrefix(p[pos:], a) {
					walk(ti+1, pos+len(a))
				}
			}
		case tCap:
			for end := pos + 1; end <= len(p) && p[end-1] != '/'; end++ {
				caps = append(caps, p[pos:end])
				walk(ti+1, end)
				caps = caps[:len(caps)-1]
			}
		case tSegs, tMCap:
			push := func(v string) {
				if t.k == tMCap {
					caps = append(caps, v)
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
				caps = append(caps, p[pos:])
				walk(ti+1, len(p))
				caps = caps[:len(caps)-1]
			}
		}
	}
	walk(0, 0)
	return best, found
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

// resolveRelated substitutes every variable in one pass (a captured value is
// never rescanned), drops empty segments, and rejects a result that is empty or
// has a "." or ".." segment. Rule 11 in 04-grammar.md.
func resolveRelated(tpl string, vals map[string]string) (string, error) {
	s := varRe.ReplaceAllStringFunc(tpl, func(m string) string { return vals[m[1:len(m)-1]] })
	var segs []string
	for _, sg := range strings.Split(s, "/") {
		if sg == "" {
			continue
		}
		if sg == "." || sg == ".." {
			return "", fmt.Errorf("related %q resolves to an invalid path %q: a captured value became a %q segment", tpl, s, sg)
		}
		segs = append(segs, sg)
	}
	if len(segs) == 0 {
		return "", fmt.Errorf("related %q resolves to an invalid path: it is empty", tpl)
	}
	return strings.Join(segs, "/"), nil
}

// badInputPath reports an input path that is not relative and clean.
func badInputPath(p string) bool {
	if p == "" || strings.HasPrefix(p, "/") {
		return true
	}
	for _, s := range strings.Split(p, "/") {
		if s == "" || s == "." || s == ".." {
			return true
		}
	}
	return false
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
		if badInputPath(p) {
			return fmt.Errorf("invalid input path %q: must be relative, with no empty, . or .. segment", p)
		}
		if prev, ok := subject[p]; ok && prev != l {
			return fmt.Errorf("path %q appears on more than one line (%q and %q); lines must not conflict", p, prev, l)
		}
		subject[p] = l
		return nil
	}
	for _, l := range sc.Changes {
		f := strings.Fields(l)
		if sc.Mode == "plain" {
			if err := note(f[0], l); err != nil {
				return cfgErr("%v", err)
			}
			touched[f[0]] = "modified"
			continue
		}
		switch f[0][0] {
		case 'A', 'M', 'T', 'D':
			if err := note(f[1], l); err != nil {
				return cfgErr("%v", err)
			}
		case 'R':
			if err := note(f[1], l); err != nil {
				return cfgErr("%v", err)
			}
			if err := note(f[2], l); err != nil {
				return cfgErr("%v", err)
			}
		default:
			return cfgErr("unrecognised status line %q", l)
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
		if r.SkipIf == "exists" && set["exists"] {
			return fail("skipIf: exists with expect: exists can never pass: the rule only applies when the related file is not there")
		}
		if r.SkipIf == "exists" && set["touched"] {
			return fail("skipIf: exists with expect: touched can pass only when the diff and the working tree disagree: the rule applies only when the related file is not there")
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
				return fail("exclude %q is identical to a match glob, so it removes every match", g)
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
			if sc.Mode == "plain" && !disk[p] {
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
						bad = "exists"
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

func main() {
	verbose := false
	args := os.Args[1:]
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

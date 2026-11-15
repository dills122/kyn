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
	Match      strs   `yaml:"match"`
	Exclude    strs   `yaml:"exclude"`
	Related    string `yaml:"related"`
	Expect     strs   `yaml:"expect"`
	SkipIf     string `yaml:"skipIf"`
	ChangeType strs   `yaml:"changeType"`
	Severity   string `yaml:"severity"`
	Message    string `yaml:"message"`
	Why        string `yaml:"why"`
	Flags      strs   `yaml:"flags"`
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

type glob struct {
	re       *regexp.Regexp
	captures []string // in order of appearance; "**" captures included
	multi    string   // name of the {**x} capture, or ""
}

// compileGlob turns a match/exclude glob into a regexp. allowCaptures is false
// for exclude globs. It returns a config error for every malformed construct.
func compileGlob(g string, allowCaptures bool) (*glob, error) {
	out := &glob{}
	var b strings.Builder
	b.WriteString("^")
	seen := map[string]bool{}
	lastWasCapture := false
	for i := 0; i < len(g); {
		c := g[i]
		switch {
		case strings.HasPrefix(g[i:], "**/"):
			b.WriteString("(?:.*/)?")
			i += 3
			lastWasCapture = false
		case strings.HasPrefix(g[i:], "**"):
			b.WriteString(".*")
			i += 2
			lastWasCapture = false
		case c == '*':
			b.WriteString("[^/]*")
			i++
			lastWasCapture = false
		case c == '?':
			b.WriteString("[^/]")
			i++
			lastWasCapture = false
		case c == '{':
			end := strings.IndexByte(g[i:], '}')
			if end < 0 {
				return nil, fmt.Errorf("glob %q: unclosed {", g)
			}
			end += i
			body := g[i+1 : end]
			if strings.Contains(body, ",") { // alternation
				parts := strings.Split(body, ",")
				for j := range parts {
					parts[j] = regexp.QuoteMeta(parts[j])
				}
				b.WriteString("(?:" + strings.Join(parts, "|") + ")")
				i = end + 1
				lastWasCapture = false
				continue
			}
			// a capture
			if !allowCaptures {
				return nil, fmt.Errorf("glob %q: captures are not allowed in exclude", g)
			}
			multi := strings.HasPrefix(body, "**")
			name := strings.TrimPrefix(body, "**")
			if !nameRe.MatchString(name) {
				return nil, fmt.Errorf("glob %q: %q is not a valid capture name (letters, digits and underscore, not starting with a digit)", g, name)
			}
			if name == "dir" {
				return nil, fmt.Errorf("glob %q: dir is reserved and cannot be captured", g)
			}
			if seen[name] {
				return nil, fmt.Errorf("glob %q: capture %q appears twice", g, name)
			}
			seen[name] = true
			if multi {
				if out.multi != "" {
					return nil, fmt.Errorf("glob %q: at most one {**name} capture per glob", g)
				}
				if i > 0 && g[i-1] != '/' {
					return nil, fmt.Errorf("glob %q: {**%s} must start a path segment", g, name)
				}
				after := end + 1
				if after < len(g) && g[after] != '/' {
					return nil, fmt.Errorf("glob %q: {**%s} must fill whole path segments", g, name)
				}
				out.multi = name
				out.captures = append(out.captures, name)
				if after < len(g) { // followed by "/": zero or more segments
					b.WriteString("(?:(.+)/)?")
					i = after + 1
				} else { // trailing: one or more segments
					b.WriteString("(.+)")
					i = after
				}
				lastWasCapture = false
				continue
			}
			if lastWasCapture {
				return nil, fmt.Errorf("glob %q: adjacent captures need literal text between them", g)
			}
			out.captures = append(out.captures, name)
			b.WriteString("([^/]+)")
			i = end + 1
			lastWasCapture = true
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
			i++
			lastWasCapture = false
		}
	}
	b.WriteString("$")
	re, err := regexp.Compile(b.String())
	if err != nil {
		return nil, err
	}
	out.re = re
	return out, nil
}

// ---- related templates ------------------------------------------------------

var varRe = regexp.MustCompile(`\{([^{}]*)\}`)

// templateVars returns the distinct variables a related template uses, in
// order of first appearance, or an error for anything that is not a variable.
func templateVars(tpl string) ([]string, error) {
	if strings.HasPrefix(tpl, "/") {
		return nil, fmt.Errorf("related %q must be a relative path", tpl)
	}
	if strings.ContainsAny(strings.NewReplacer("{", "", "}", "").Replace(tpl), "*?[]") {
		return nil, fmt.Errorf("related %q contains glob metacharacters; it names exactly one path", tpl)
	}
	for _, seg := range strings.Split(tpl, "/") {
		if seg == ".." {
			return nil, fmt.Errorf("related %q contains \"..\"", tpl)
		}
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
	for _, l := range sc.Changes {
		f := strings.Fields(l)
		if sc.Mode == "plain" {
			touched[f[0]] = "modified"
			continue
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
		default:
			return cfgErr("unrecognised status line %q", l)
		}
	}

	ids := make([]string, 0, len(cfg.Rules))
	for id := range cfg.Rules {
		ids = append(ids, id)
	}
	sort.Strings(ids) // rule IDs are always processed in sorted order

	// ---- validation, per rule, before any evaluation ----
	type compiled struct {
		id       string
		r        Rule
		match    []*glob
		exclude  []*glob
		vars     []string
		sev      string
		expect   []string
		hasRelat bool
	}
	var rules []compiled
	for _, id := range ids {
		r := cfg.Rules[id]
		c := compiled{id: id, r: r}
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
		c.hasRelat = r.Related != ""
		if !c.hasRelat && !none {
			return fail("related is required unless expect is none")
		}
		if !c.hasRelat && r.SkipIf != "" {
			return fail("skipIf needs a related path")
		}
		if r.SkipIf == "exists" && set["exists"] {
			return fail("skipIf: exists with expect: exists can never pass: the rule only applies when the related file is not there")
		}

		var names []string
		for gi, g := range r.Match {
			cg, err := compileGlob(g, true)
			if err != nil {
				return fail("%v", err)
			}
			if gi == 0 {
				names = cg.captures
			} else if strings.Join(sorted(cg.captures), ",") != strings.Join(sorted(names), ",") {
				return fail("every match glob must bind the same capture names (%v versus %v)", names, cg.captures)
			}
			c.match = append(c.match, cg)
		}
		for _, g := range r.Exclude {
			cg, err := compileGlob(g, false)
			if err != nil {
				return fail("%v", err)
			}
			c.exclude = append(c.exclude, cg)
		}
		if c.hasRelat {
			vars, err := templateVars(r.Related)
			if err != nil {
				return fail("%v", err)
			}
			c.vars = vars
			slashy := 0
			for _, v := range vars {
				if v != "dir" && !contains(names, v) {
					return fail("related uses {%s} but match does not capture it", v)
				}
				if v == "dir" || v == c.match[0].multi {
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
		bind := func(p string) (map[string]string, bool, error) {
			for _, e := range c.exclude {
				if e.re.MatchString(p) {
					return nil, false, nil
				}
			}
			var first map[string]string
			for _, g := range c.match {
				m := g.re.FindStringSubmatch(p)
				if m == nil {
					continue
				}
				v := map[string]string{"dir": path.Dir(p)}
				if v["dir"] == "." {
					v["dir"] = ""
				}
				for i, n := range g.captures {
					v[n] = m[i+1]
				}
				if first == nil {
					first = v
				} else if !equal(first, v) {
					return nil, false, fmt.Errorf("rule %q: %q matches two match globs with different bindings; make the globs not overlap", c.id, p)
				}
			}
			return first, first != nil, nil
		}

		insts := map[string]*inst{}
		for _, p := range paths {
			v, ok, err := bind(p)
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
			rel := r.Related
			for _, n := range c.vars {
				rel = strings.ReplaceAll(rel, "{"+n+"}", in.vars[n])
			}
			rel = strings.TrimPrefix(path.Clean(rel), "/") // an empty {dir} leaves a leading "/"
			if _, self, _ := bind(rel); self {
				return cfgErr("rule %q: related path %q also matches this rule's own match; add an exclude or narrow the match", c.id, rel)
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

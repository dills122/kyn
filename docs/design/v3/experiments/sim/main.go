//go:build ignore

// Simulator for the v3 configuration shape in docs/design/v3/04-grammar.md.
//
// This is NOT Kyn code and shares nothing with internal/. It is a small model
// of the designed semantics -- captures in `match`, instance keys derived from
// the variables `related` uses, the three related-path states, `skipIf`,
// `expect`, flags, and self-match rejection -- used to run realistic scenarios
// before anything is implemented.
//
//	go run docs/design/v3/experiments/sim/main.go docs/design/v3/experiments/sim/scenarios.yaml
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
	Changes []string `yaml:"changes"` // git diff --name-status lines
}

func compile(glob string) (*regexp.Regexp, []string, error) {
	var b strings.Builder
	var names []string
	b.WriteString("^")
	for i := 0; i < len(glob); {
		switch {
		case strings.HasPrefix(glob[i:], "**/"):
			b.WriteString("(?:.*/)?")
			i += 3
		case strings.HasPrefix(glob[i:], "**"):
			b.WriteString(".*")
			i += 2
		case glob[i] == '*':
			b.WriteString("[^/]*")
			i++
		case glob[i] == '{':
			end := strings.IndexByte(glob[i:], '}')
			if end < 0 {
				return nil, nil, fmt.Errorf("unclosed { in %q", glob)
			}
			end += i
			body := glob[i+1 : end]
			switch {
			case strings.Contains(body, ","):
				parts := strings.Split(body, ",")
				for j := range parts {
					parts[j] = regexp.QuoteMeta(parts[j])
				}
				b.WriteString("(?:" + strings.Join(parts, "|") + ")")
			case strings.HasPrefix(body, "**"):
				names = append(names, body[2:])
				b.WriteString("(.+)")
			default:
				names = append(names, body)
				b.WriteString("([^/]+)")
			}
			i = end + 1
		default:
			b.WriteString(regexp.QuoteMeta(string(glob[i])))
			i++
		}
	}
	b.WriteString("$")
	re, err := regexp.Compile(b.String())
	return re, names, err
}

var varRe = regexp.MustCompile(`\{([A-Za-z]+)\}`)

type inst struct {
	name    string
	vars    map[string]string
	sources []string
	status  map[string]bool
}

func run(sc Scenario) {
	fmt.Printf("\n━━ %s\n", sc.Name)
	for _, l := range strings.Split(strings.TrimRight(sc.Config, "\n"), "\n") {
		fmt.Println("    " + l)
	}
	var cfg Config
	dec := yaml.NewDecoder(strings.NewReader(sc.Config))
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		fmt.Println("  CONFIG ERROR:", err)
		return
	}
	disk := map[string]bool{}
	for _, f := range sc.Disk {
		disk[f] = true
	}
	touched := map[string]string{} // path -> added|modified|renamed
	vanished := map[string]bool{}
	for _, l := range sc.Changes {
		f := strings.Fields(l)
		switch f[0][0] {
		case 'A':
			touched[f[1]] = "added"
		case 'M':
			touched[f[1]] = "modified"
		case 'D':
			vanished[f[1]] = true
		case 'R':
			vanished[f[1]] = true
			touched[f[2]] = "renamed"
		}
	}
	fmt.Println("  change: " + strings.Join(sc.Changes, " | "))

	ids := make([]string, 0, len(cfg.Rules))
	for id := range cfg.Rules {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	flags := map[string]bool{}
	counts := map[string]int{}
	blocked := false
	for _, id := range ids {
		r := cfg.Rules[id]
		if len(r.Expect) == 0 {
			fmt.Printf("  CONFIG ERROR: rule %q: expect is required\n", id)
			return
		}
		if r.Related == "" && r.Expect[0] != "none" {
			fmt.Printf("  CONFIG ERROR: rule %q: related is required unless expect is none\n", id)
			return
		}
		if r.Related == "" && r.SkipIf != "" {
			fmt.Printf("  CONFIG ERROR: rule %q: skipIf needs a related path\n", id)
			return
		}
		sev := r.Severity
		if sev == "" {
			sev = "error"
			if r.Expect[0] == "none" {
				sev = "info"
			}
		}
		type cg struct {
			re    *regexp.Regexp
			names []string
		}
		var globs []cg
		for _, g := range r.Match {
			re, names, err := compile(g)
			if err != nil {
				fmt.Println("  CONFIG ERROR:", err)
				return
			}
			globs = append(globs, cg{re, names})
		}
		var excl []*regexp.Regexp
		for _, g := range r.Exclude {
			re, _, _ := compile(g)
			excl = append(excl, re)
		}
		matchSrc := func(p string) (map[string]string, bool) {
			for _, e := range excl {
				if e.MatchString(p) {
					return nil, false
				}
			}
			for _, g := range globs {
				if m := g.re.FindStringSubmatch(p); m != nil {
					v := map[string]string{"dir": path.Dir(p)}
					for i, n := range g.names {
						v[n] = m[i+1]
					}
					return v, true
				}
			}
			return nil, false
		}
		// footprint: variables the related template actually uses
		var fp []string
		for _, m := range varRe.FindAllStringSubmatch(r.Related, -1) {
			fp = append(fp, m[1])
		}
		insts := map[string]*inst{}
		paths := make([]string, 0, len(touched))
		for p := range touched {
			paths = append(paths, p)
		}
		sort.Strings(paths)
		for _, p := range paths {
			v, ok := matchSrc(p)
			if !ok {
				continue
			}
			var keyParts []string
			for _, n := range fp {
				if _, has := v[n]; !has {
					fmt.Printf("  CONFIG ERROR: rule %q: related uses {%s} but match does not capture it\n", id, n)
					return
				}
				part := v[n]
				if n == "dir" && part == "." {
					part = "(root)"
				}
				keyParts = append(keyParts, part)
			}
			name := strings.Join(keyParts, "/")
			if len(keyParts) == 0 {
				name = "(all)"
			}
			in := insts[name]
			if in == nil {
				in = &inst{name: name, vars: v, status: map[string]bool{}}
				insts[name] = in
			}
			in.sources = append(in.sources, p)
			in.status[touched[p]] = true
		}
		names := make([]string, 0, len(insts))
		for n := range insts {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			in := insts[n]
			if len(r.ChangeType) > 0 {
				ok := false
				for _, ct := range r.ChangeType {
					ok = ok || in.status[ct]
				}
				if !ok {
					continue
				}
			}
			if r.Related == "" { // flag-only rule: nothing to resolve
				for _, f := range r.Flags {
					flags[f] = true
				}
				counts["info"]++
				fmt.Printf("  %-7s [%s] %s  %s  [%d sources]\n", "INFO", sev, id, in.name, len(in.sources))
				continue
			}
			rel := r.Related
			for k, v := range in.vars {
				rel = strings.ReplaceAll(rel, "{"+k+"}", v)
			}
			rel = path.Clean(rel)
			if _, self := matchSrc(rel); self {
				fmt.Printf("  CONFIG ERROR: rule %q: related path %q also matches this rule's own match; add an exclude or narrow the match\n", id, rel)
				return
			}
			state := "missing"
			if disk[rel] {
				state = "exists"
			} else if vanished[rel] {
				state = "removed-by-this-change"
			}
			status, detail := "pass", ""
			switch {
			case r.SkipIf == "missing" && state == "missing":
				status = "skipped"
			case r.SkipIf == "exists" && state != "missing":
				status = "skipped"
			default:
				for _, e := range r.Expect {
					bad := ""
					switch e {
					case "touched":
						if _, ok := touched[rel]; !ok {
							bad = "not touched"
							if state == "missing" {
								bad = "does not exist"
							}
						}
					case "untouched":
						if _, ok := touched[rel]; ok {
							bad = "was touched"
						}
					case "exists":
						if !disk[rel] {
							bad = "does not exist"
						}
					case "missing":
						if disk[rel] {
							bad = "exists"
						}
					case "none":
						status = "info"
					}
					if bad != "" {
						status = "fail"
						if detail == "" || bad == "does not exist" {
							detail = bad
						}
					}
				}
				for _, f := range r.Flags {
					flags[f] = true
				}
			}
			counts[status]++
			if status == "fail" && sev == "error" {
				blocked = true
			}
			line := fmt.Sprintf("  %-7s [%s] %s  %s -> %s (%s)", strings.ToUpper(status), sev, id, in.name, rel, state)
			if detail != "" {
				line += ": " + detail
			}
			if len(in.sources) > 1 {
				line += fmt.Sprintf("  [%d sources]", len(in.sources))
			}
			fmt.Println(line)
		}
	}
	head := "PASS"
	if blocked {
		head = "FAIL"
	} else if counts["fail"] > 0 {
		head = "NON-BLOCKING"
	}
	var fl []string
	for f := range flags {
		fl = append(fl, f)
	}
	sort.Strings(fl)
	fmt.Printf("  => %s  (passed %d, failed %d, info %d, skipped %d)", head, counts["pass"], counts["fail"], counts["info"], counts["skipped"])
	if len(fl) > 0 {
		fmt.Printf("  flags=%v", fl)
	}
	if counts["pass"]+counts["fail"]+counts["info"]+counts["skipped"] == 0 {
		fmt.Print("  (no rule matched anything)")
	}
	fmt.Println()
}

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var scs []Scenario
	if err := yaml.Unmarshal(data, &scs); err != nil {
		panic(err)
	}
	for _, sc := range scs {
		run(sc)
	}
}

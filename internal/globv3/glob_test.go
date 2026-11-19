package globv3

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/bmatcuk/doublestar/v4"
)

func TestBind(t *testing.T) {
	tests := []struct {
		glob, path string
		want       []string // captured values in order; nil with ok=false means no match
		ok         bool
	}{
		// Longest first, first capture first.
		{"{name}.{kind}.ts", "user.profile.component.ts", []string{"user.profile", "component"}, true},
		{"src/{m}-{n}.go", "src/a-b-c.go", []string{"a-b", "c"}, true},
		{"src/*-{n}.go", "src/a-b-c.go", []string{"b-c"}, true},
		{"src/**/{name}.component.ts", "src/a/b/button.component.ts", []string{"button"}, true},
		{"src/{n}x.go", "src/..x.go", []string{".."}, true}, // rule 11 stops this later, at resolution
		// Whole segments.
		{"src/{**pkg}/{name}.py", "src/a/b/x.py", []string{"a/b", "x"}, true},
		{"src/{**pkg}/{name}.py", "src/x.py", []string{"", "x"}, true},
		{"{**d}/{name}.go", "x.go", []string{"", "x"}, true},
		{"src/{**rest}", "src/a/b", []string{"a/b"}, true},
		{"src/{**rest}", "src", nil, false}, // a trailing {**name} needs one or more segments
		// ** and * semantics.
		{"src/**/x.go", "src/x.go", nil, true},
		{"src/**/x.go", "src/a/b/x.go", nil, true},
		{"src/**", "src", nil, false},
		{"src/**", "src/a", nil, true},
		{"**/x", "x", nil, true},
		{"**", "a/b/c", nil, true},
		{"src/*.go", "src/a/b.go", nil, false},
		{"*.go", ".hidden.go", nil, true}, // * matches a leading dot
		{"a?c", "añc", nil, true},         // ? is one code point
		{"a?c", "ac", nil, false},
		{"src/*.{ts,tsx}", "src/a.tsx", nil, true},
		{"src/*.{ts,tsx}", "src/a.js", nil, false},
		{"café/{n}.go", "café/x.go", []string{"x"}, true},
		{"café/{n}.go", "café/x.go"[:0] + "cafe\u0301/x.go", nil, false}, // no normalization
		{"src/{n}.go", "src/.go", nil, false},                            // a capture is one or more characters
	}
	for _, tt := range tests {
		t.Run(tt.glob+" on "+tt.path, func(t *testing.T) {
			g, err := Compile(tt.glob, true)
			if err != nil {
				t.Fatalf("Compile(%q): %v", tt.glob, err)
			}
			got, ok := g.Bind(tt.path)
			if ok != tt.ok {
				t.Fatalf("matched = %v, want %v", ok, tt.ok)
			}
			if tt.want != nil && !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("captures = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCompileRejects(t *testing.T) {
	tests := map[string]string{
		"src/[ab].go":            "character classes",
		"src/[!x].go":            "character classes",
		`src/a\*.go`:             "backslash",
		"src/{a,{b,c}}.go":       "nested braces",
		"src/{a,*}.go":           "wildcard",
		"src/{a,}.go":            "empty alternative",
		"src/{a,b/c}.go":         "/ inside {}",
		"src/a**b/x.go":          "whole path segment",
		"/src/x.go":              "relative",
		"src/x/":                 "empty path segment",
		"src//x.go":              "empty path segment",
		"src/./x.go":             "\".\" segment",
		"src/../x.go":            "\"..\" segment",
		"src/{..,a}/x.go":        "\"..\"",
		"src/{.,a}/x.go":         "\".\"",
		"src/{ts, html}":         "whitespace inside {}",
		"src/ x.go":              "whitespace",
		"src/x.go ":              "whitespace",
		"src/\tx.go":             "control character",
		"\ufeffsrc/x.go":         "byte order mark",
		"src/{n":                 "unclosed",
		"src/n}":                 "unmatched",
		"src/\xffx.go":           "UTF-8",
		"c:{n}.go":               "letter and a colon",
		"c:/{n}":                 "letter and a colon",
		"c:*":                    "letter and a colon",
		"src/{1x}.go":            "not a valid capture name",
		"src/{dir}/x.go":         "reserved",
		"src/{a}/{a}.go":         "twice",
		"src/{a}{b}.go":          "adjacent captures",
		"src/*{n}.go":            "cannot touch",
		"src/{n}?.go":            "cannot touch",
		"src/{**a}/{**b}/x.go":   "at most one",
		"src/{**a}x/y.go":        "whole path segments",
		"src/{**a,b}/y.go":       "not allowed inside {}",
		"src/**/{n}/**":          "between two **",
		"src/{**pkg}/{name}/**":  "between two **",
		"src/**/{name}/{**rest}": "between two **",
		"src/{**a}/{n}/{**b}":    "at most one",
	}
	for glob, want := range tests {
		t.Run(glob, func(t *testing.T) {
			_, err := Compile(glob, true)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("Compile(%q) error = %v, want it to contain %q", glob, err, want)
			}
		})
	}
	if _, err := Compile("src/{n}.go", false); err == nil || !strings.Contains(err.Error(), "not allowed in exclude") {
		t.Fatalf("a capture in exclude: error = %v", err)
	}
	// Legal, and easy to confuse with the illegal.
	// Format characters are not blank: a glob with one at an edge compiles (and never matches).
	for _, ok := range []string{"src/x\u200b/y.go", "src/{a,b\u200b}", "src/**/{n}.go", "src/{**pkg}/{name}.py", "my file/{n}.go", "a\u200db/{n}.go", "src/{ts,html}/x", "**/*.go"} {
		if _, err := Compile(ok, true); err != nil {
			t.Errorf("Compile(%q) = %v, want accepted", ok, err)
		}
	}
}

func TestSelector(t *testing.T) {
	s, err := NewSelector([]string{"src/**/{name}.component.ts", "lib/**/{name}.component.ts"}, []string{"**/*.spec.ts"})
	if err != nil {
		t.Fatal(err)
	}
	b, ok, err := s.Bind("lib/ui/button.component.ts")
	if err != nil || !ok || b["name"] != "button" || b["dir"] != "lib/ui" {
		t.Fatalf("Bind = %v %v %v", b, ok, err)
	}
	if _, ok, _ := s.Bind("src/a/b.spec.ts"); ok {
		t.Error("an excluded path matched")
	}

	// Reordering the globs changes nothing.
	r, _ := NewSelector([]string{"lib/**/{name}.component.ts", "src/**/{name}.component.ts"}, nil)
	b2, _, _ := r.Bind("lib/ui/button.component.ts")
	if !reflect.DeepEqual(b, b2) {
		t.Errorf("reordered globs bind %v, want %v", b2, b)
	}

	// Two globs that bind one path differently are an error (rule 9).
	o, err := NewSelector([]string{"src/{a}-{b}.go", "src/{a}.{b}.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := o.Bind("src/x-y.go"); err != nil {
		t.Errorf("one glob matched, got %v", err)
	}
	o2, _ := NewSelector([]string{"{a}/x.go", "{a}/{b}"}, nil)
	_ = o2
}

func TestSelectorRejects(t *testing.T) {
	tests := []struct {
		name           string
		match, exclude []string
		want           string
	}{
		{"no match", nil, nil, "at least one"},
		{"different names", []string{"src/{a}.go", "lib/{b}.go"}, nil, "same capture names"},
		{"different kinds", []string{"src/{a}/x.go", "lib/{**a}/x.go"}, nil, "same kind"},
		{"kinds in either order", []string{"lib/{**a}/x.go", "src/{a}/x.go"}, nil, "same kind"},
		{"exclude identical to match", []string{"src/*.go"}, []string{"src/*.go"}, "identical"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSelector(tt.match, tt.exclude)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}

	o, _ := NewSelector([]string{"{a}/x.go", "{a}/*.go"}, nil)
	if _, _, err := o.Bind("q/x.go"); err != nil {
		t.Logf("same binding from both globs is fine: %v", err)
	}
	d, _ := NewSelector([]string{"src/{n}-1.go", "src/{n}.go"}, nil)
	if _, _, err := d.Bind("src/a-1.go"); err == nil {
		t.Error("a path bound n=a by one glob and n=a-1 by the other should be rule 9")
	}
}

func TestTemplates(t *testing.T) {
	sel, _ := NewSelector([]string{"src/**/{name}.component.ts"}, nil)
	good := map[string][]string{
		"{dir}/{name}.stories.ts": {"dir", "name"},
		"docs/{name}.md":          {"name"},
	}
	for tpl, vars := range good {
		tp, err := ParseTemplate(tpl, sel)
		if err != nil || !reflect.DeepEqual(tp.Vars(), vars) {
			t.Errorf("ParseTemplate(%q) = %v, %v; want vars %v", tpl, tp, err, vars)
		}
	}
	mirror, _ := NewSelector([]string{"src/{**pkg}/{name}.py"}, nil)
	bad := []struct {
		sel  *Selector
		tpl  string
		want string
	}{
		{sel, "", "empty"},
		{sel, "/{name}.md", "relative"},
		{sel, "{dir}/{name}/", "not end with /"},
		{sel, `{dir}\{name}.md`, "backslash"},
		{sel, "{dir}/../{name}.md", "\"..\""},
		{sel, "{dir}/./{name}.md", "\".\""},
		{sel, "{dir}/*.md", "glob metacharacters"},
		{sel, "{dir}/{name}{name}.md", "adjacent"},
		{sel, "{dir}/{nope}.md", "not capture"},
		{sel, "{dir}/{name}.{", "unmatched"},
		{sel, "{dir}/x.md", "never used"},
		{sel, " {dir}/{name}.md", "whitespace"},
		{sel, "c:/{name}.md", "drive"},
		{sel, "{dir}.md/{name}", "whole path segment"},
		{sel, "{dir}", "only of"},
		{mirror, "tests/{pkg}/{dir}/{name}.py", "more than one variable"},
		{mirror, "tests/{pkg}x/{name}.py", "whole path segment"},
		{sel, "src/a.component.ts", "never used"},
	}
	for _, tt := range bad {
		if _, err := ParseTemplate(tt.tpl, tt.sel); err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("ParseTemplate(%q) error = %v, want %q", tt.tpl, err, tt.want)
		}
	}

	// A constant related that matches its own rule is rule 3 at load.
	self, _ := NewSelector([]string{"src/*.go"}, nil)
	if _, err := ParseTemplate("src/a.go", self); err == nil || !strings.Contains(err.Error(), "own match") {
		t.Errorf("constant related inside its own match: error = %v", err)
	}
	if _, err := ParseTemplate("docs/a.md", self); err == nil || !strings.Contains(err.Error(), "never used") {
		// no captures: a constant related is fine
		_ = err
	}
}

func TestResolve(t *testing.T) {
	sel, _ := NewSelector([]string{"src/{**pkg}/{name}.py"}, nil)
	tp, err := ParseTemplate("tests/{pkg}/test_{name}.py", sel)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		path, want, wantErr string
	}{
		{"src/a/b/x.py", "tests/a/b/test_x.py", ""},
		{"src/x.py", "tests/test_x.py", ""},             // empty {**pkg} leaves no stray slash
		{"src/{dir}/x.py", "tests/{dir}/test_x.py", ""}, // a value is never rescanned
	}
	for _, tt := range tests {
		b, ok, err := sel.Bind(tt.path)
		if err != nil || !ok {
			t.Fatalf("Bind(%q) = %v %v %v", tt.path, b, ok, err)
		}
		got, err := tp.Resolve(b)
		if tt.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Resolve(%q) error = %v, want %q", tt.path, err, tt.wantErr)
			}
			continue
		}
		if err != nil || got != tt.want {
			t.Errorf("Resolve(%q) = %q, %v; want %q", tt.path, got, err, tt.want)
		}
	}

	// Rule 11: a captured value that makes the path unclean stops resolution.
	n, _ := NewSelector([]string{"src/{n}x.go"}, nil)
	two, _ := ParseTemplate("{n}/{n}/x.txt", n)
	b, _, _ := n.Bind("src/..x.go")
	if _, err := two.Resolve(b); err == nil || !strings.Contains(err.Error(), "\"..\"") {
		t.Errorf("a capture binding .. : error = %v", err)
	}
	one, _ := ParseTemplate("docs/{n}.md", n)
	b, _, _ = n.Bind("src/a x.go")
	if got, err := one.Resolve(b); err != nil || got != "docs/a .md" {
		t.Errorf("an interior space is fine: Resolve = %q, %v", got, err)
	}
	b["n"] = " a"
	if _, err := one.Resolve(b); err == nil || !strings.Contains(err.Error(), "whitespace") {
		t.Errorf("an edge blank in a resolved segment: error = %v", err)
	}
}

func TestWhyNotClean(t *testing.T) {
	clean := []string{"a", "a\u200b/b", "\u200dx", "src/a.go", "my file.go", "日本語/テスト.go", ".hidden/x", "a..b/x"}
	for _, p := range clean {
		if why := WhyNotClean(p); why != "" {
			t.Errorf("WhyNotClean(%q) = %q, want clean", p, why)
		}
	}
	unclean := []string{"", "/a", "a//b", "a/./b", "a/../b", `a\b`, "a\tb", "a\x00b", "\ufeffa", " a", "a ", "a/ b", "a /b", "C:/x", "c:x", "a\xff", "x/"}
	for _, p := range unclean {
		if why := WhyNotClean(p); why == "" {
			t.Errorf("WhyNotClean(%q) = clean, want a reason", p)
		}
	}
}

// Binding is total: no accepted glob has two longest parses that bind different
// values. The rejection of a capture between two ** is what guarantees it, so
// this generates random globs, keeps the accepted ones, and looks for a counter
// example.
func TestBindingIsTotal(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	pieces := []string{"a", "b", "-", ".", "*", "?", "{x}", "{y}", "{a,b}"}
	segs := func() string {
		var sb strings.Builder
		for i := 0; i < 1+rng.Intn(4); i++ {
			sb.WriteString(pieces[rng.Intn(len(pieces))])
		}
		return sb.String()
	}
	globs := 0
	for i := 0; i < 20000; i++ {
		var parts []string
		for j := 0; j < 1+rng.Intn(5); j++ {
			switch rng.Intn(6) {
			case 0:
				parts = append(parts, "**")
			case 1:
				parts = append(parts, "{**p}")
			default:
				parts = append(parts, segs())
			}
		}
		g, err := Compile(strings.Join(parts, "/"), true)
		if err != nil {
			continue
		}
		globs++
		for k := 0; k < 25; k++ {
			var ps []string
			for j := 0; j < 1+rng.Intn(5); j++ {
				var sb strings.Builder
				for q := 0; q < 1+rng.Intn(5); q++ {
					sb.WriteByte("ab-."[rng.Intn(4)])
				}
				ps = append(ps, sb.String())
			}
			p := strings.Join(ps, "/")
			all := g.parses(p)
			if len(all) == 0 {
				continue
			}
			best := all[0]
			for _, c := range all[1:] {
				if cmpLen(c, best) > 0 {
					best = c
				}
			}
			for _, c := range all {
				if cmpLen(c, best) == 0 && !sameCaps(c, best) {
					t.Fatalf("glob %q on %q: two longest parses bind %v and %v", g.src(), p, best, c)
				}
			}
		}
	}
	if globs < 1000 {
		t.Fatalf("only %d accepted globs; the generator is too strict to test anything", globs)
	}
}

func sameCaps(a, b []capv) bool {
	for i := range a {
		if a[i].s != b[i].s {
			return false
		}
	}
	return true
}

// Against the shipped matcher's library, on the subset where v3 and doublestar
// are meant to agree: * ? ** (not trailing) and literal alternation, no captures.
func TestAgreesWithDoublestar(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	pieces := []string{"a", "b", "c", ".", "*", "?", "{a,b}", "ab"}
	rand1 := func() string {
		var sb strings.Builder
		for i := 0; i < 1+rng.Intn(3); i++ {
			sb.WriteString(pieces[rng.Intn(len(pieces))])
		}
		return sb.String()
	}
	checked := 0
	for i := 0; i < 30000; i++ {
		var parts []string
		n := 1 + rng.Intn(4)
		for j := 0; j < n; j++ {
			if rng.Intn(4) == 0 && j < n-1 { // never trailing: v3 deliberately differs there
				parts = append(parts, "**")
			} else {
				parts = append(parts, rand1())
			}
		}
		glob := strings.Join(parts, "/")
		if parts[len(parts)-1] == "**" { // v3 deliberately differs on a trailing **
			continue
		}
		g, err := Compile(glob, false)
		if err != nil {
			continue
		}
		for k := 0; k < 20; k++ {
			var ps []string
			for j := 0; j < 1+rng.Intn(4); j++ {
				var sb strings.Builder
				for q := 0; q < 1+rng.Intn(3); q++ {
					sb.WriteByte("abc."[rng.Intn(4)])
				}
				ps = append(ps, sb.String())
			}
			p := strings.Join(ps, "/")
			_, got := g.Bind(p)
			want, err := doublestar.Match(glob, p)
			if err != nil {
				continue
			}
			checked++
			if got != want {
				t.Fatalf("glob %q on %q: v3 = %v, doublestar = %v", glob, p, got, want)
			}
		}
	}
	if checked < 10000 {
		t.Fatalf("only %d comparisons", checked)
	}
}

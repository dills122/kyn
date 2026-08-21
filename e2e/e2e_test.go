// Package e2e runs the compiled kyn binary against realistic example
// projects under e2e/projects, black-box, the way a CI pipeline would.
//
// Unlike the unit tests under internal/*, these tests build the real
// cmd/kyn binary and invoke it as a subprocess against on-disk fixture
// repos, asserting exit codes and rendered output. This catches
// regressions in flag wiring, file-system behavior (kinExists is a real
// os.Stat), and documented workflows (presets, migration) that in-process
// unit tests can miss.
//
// Update goldens after an intentional output change:
//
//	go test ./e2e/... -update
package e2e

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

var updateGolden = flag.Bool("update", false, "update e2e golden files")

var kynBinary string

func TestMain(m *testing.M) {
	os.Exit(runMain(m))
}

func runMain(m *testing.M) int {
	tmpDir, err := os.MkdirTemp("", "kyn-e2e-bin")
	if err != nil {
		fmt.Fprintln(os.Stderr, "e2e: create temp dir:", err)
		return 1
	}
	defer os.RemoveAll(tmpDir)

	binName := "kyn"
	if runtime.GOOS == "windows" {
		binName = "kyn.exe"
	}
	kynBinary = filepath.Join(tmpDir, binName)

	repoRoot, err := filepath.Abs("..")
	if err != nil {
		fmt.Fprintln(os.Stderr, "e2e: resolve repo root:", err)
		return 1
	}

	build := exec.Command("go", "build", "-o", kynBinary, "./cmd/kyn")
	build.Dir = repoRoot
	build.Stdout = os.Stderr
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "e2e: build kyn binary:", err)
		return 1
	}

	return m.Run()
}

// scenario is one CLI invocation against a project fixture, along with the
// expected outcome. Scenarios live in each project's scenarios.json so
// fixtures stay self-describing.
type scenario struct {
	Name         string   `json:"name"`
	Args         []string `json:"args"`
	WantExitCode int      `json:"wantExitCode"`
	Golden       string   `json:"golden,omitempty"`
	Contains     []string `json:"contains,omitempty"`
	NotContains  []string `json:"notContains,omitempty"`
}

// TestE2EProjects walks every fixture under e2e/projects and runs its
// declared scenarios against the compiled binary.
func TestE2EProjects(t *testing.T) {
	const projectsDir = "projects"

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		t.Fatalf("read %s: %v", projectsDir, err)
	}

	var projects []string
	for _, entry := range entries {
		if entry.IsDir() {
			projects = append(projects, entry.Name())
		}
	}
	sort.Strings(projects)
	if len(projects) == 0 {
		t.Fatal("no project fixtures found under e2e/projects")
	}

	for _, project := range projects {
		project := project
		scenarioPath := filepath.Join(projectsDir, project, "scenarios.json")
		data, err := os.ReadFile(scenarioPath)
		if err != nil {
			t.Fatalf("project %q: read scenarios.json: %v", project, err)
		}

		var scenarios []scenario
		if err := json.Unmarshal(data, &scenarios); err != nil {
			t.Fatalf("project %q: parse scenarios.json: %v", project, err)
		}
		if len(scenarios) == 0 {
			t.Fatalf("project %q: scenarios.json declares no scenarios", project)
		}

		for _, sc := range scenarios {
			sc := sc
			t.Run(project+"/"+sc.Name, func(t *testing.T) {
				runScenario(t, project, sc)
			})
		}
	}
}

func runScenario(t *testing.T, project string, sc scenario) {
	t.Helper()

	projectDir, err := filepath.Abs(filepath.Join("projects", project))
	if err != nil {
		t.Fatalf("resolve project dir: %v", err)
	}

	stdout, stderr, exitCode := runKyn(t, projectDir, sc.Args)

	if exitCode != sc.WantExitCode {
		t.Fatalf(
			"exit code = %d, want %d\nargs: %v\n--- stdout ---\n%s\n--- stderr ---\n%s",
			exitCode, sc.WantExitCode, sc.Args, stdout, stderr,
		)
	}

	if sc.Golden != "" {
		assertGolden(t, project, sc.Golden, stdout)
	}

	for _, want := range sc.Contains {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing expected substring %q\n--- stdout ---\n%s", want, stdout)
		}
	}
	for _, unwanted := range sc.NotContains {
		if strings.Contains(stdout, unwanted) {
			t.Fatalf("stdout contains unexpected substring %q\n--- stdout ---\n%s", unwanted, stdout)
		}
	}
}

// runKyn runs the compiled kyn binary with args, cwd set to dir, and
// returns captured stdout, stderr, and the process exit code.
func runKyn(t *testing.T, dir string, args []string) (stdout string, stderr string, exitCode int) {
	t.Helper()
	return runKynStdin(t, dir, args, "")
}

// runKynStdin is runKyn with an explicit stdin payload, for scenarios
// that need e.g. `--files-from -` with no changed files.
func runKynStdin(t *testing.T, dir string, args []string, stdin string) (stdout string, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(kynBinary, args...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(stdin)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	switch e := err.(type) {
	case nil:
		exitCode = 0
	case *exec.ExitError:
		exitCode = e.ExitCode()
	default:
		t.Fatalf("run kyn %v: %v", args, err)
	}

	return outBuf.String(), errBuf.String(), exitCode
}

func assertGolden(t *testing.T, project string, name string, got string) {
	t.Helper()

	got = normalizeGoldenText(got)
	path := filepath.Join("projects", project, "testdata", name)

	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create golden dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatalf("write golden %s: %v", path, err)
		}
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (run with -update to create it)", path, err)
	}

	wantText := normalizeGoldenText(string(want))
	if got != wantText {
		t.Fatalf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", path, got, wantText)
	}
}

// normalizeGoldenText keeps golden comparisons stable across the OS
// matrix; kyn output itself uses '\n', but this guards against
// checkout-time line-ending translation on Windows runners.
func normalizeGoldenText(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

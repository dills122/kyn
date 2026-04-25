package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteUnknownCommandReturnsUsage(t *testing.T) {
	code := runWithArgs(t, []string{"kyn", "wat"})
	if code != ExitUsage {
		t.Fatalf("expected exit %d, got %d", ExitUsage, code)
	}
}

func TestExecuteRuleFailureReturnsOne(t *testing.T) {
	dir := t.TempDir()
	cfg := `version: 1
families:
  - id: angular-component
    include:
      - "libs/**/*.component.ts"
    baseName:
      stripSuffixes:
        - ".component"
    kin:
      story: "{dir}/{base}.stories.ts"
rules:
  - id: storybook-sync
    family: angular-component
    severity: error
    when:
      changedAny: [source]
      kinExists: [story]
    require:
      kinChanged: [story]
    message: "sync story"
`
	if err := os.MkdirAll(filepath.Join(dir, "libs/ui/button"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "kyn.config.yaml"), []byte(cfg), 0o600); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "libs/ui/button/button.component.ts"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write component: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "libs/ui/button/button.stories.ts"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write story: %v", err)
	}

	code := runWithArgs(t, []string{
		"kyn", "check",
		"--cwd", dir,
		"--config", "kyn.config.yaml",
		"--files", "libs/ui/button/button.component.ts",
	})
	if code != ExitRuleFailure {
		t.Fatalf("expected exit %d, got %d", ExitRuleFailure, code)
	}
}

func TestExecuteSuccessReturnsZero(t *testing.T) {
	dir := t.TempDir()
	cfg := `version: 1
families:
  - id: angular-component
    include:
      - "libs/**/*.component.ts"
    baseName:
      stripSuffixes:
        - ".component"
    kin:
      story: "{dir}/{base}.stories.ts"
rules:
  - id: storybook-sync
    family: angular-component
    severity: error
    when:
      changedAny: [source]
      kinExists: [story]
    require:
      kinChanged: [story]
    message: "sync story"
`
	if err := os.MkdirAll(filepath.Join(dir, "libs/ui/button"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "kyn.config.yaml"), []byte(cfg), 0o600); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "libs/ui/button/button.component.ts"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write component: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "libs/ui/button/button.stories.ts"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write story: %v", err)
	}

	code := runWithArgs(t, []string{
		"kyn", "check",
		"--cwd", dir,
		"--config", "kyn.config.yaml",
		"--files", "libs/ui/button/button.component.ts,libs/ui/button/button.stories.ts",
	})
	if code != ExitOK {
		t.Fatalf("expected exit %d, got %d", ExitOK, code)
	}
}

func TestExecuteAutoGitModeReturnsZero(t *testing.T) {
	dir := t.TempDir()
	cfg := `version: 1
families:
  - id: angular-component
    include:
      - "libs/**/*.component.ts"
    baseName:
      stripSuffixes:
        - ".component"
    kin:
      story: "{dir}/{base}.stories.ts"
rules:
  - id: storybook-sync
    family: angular-component
    severity: error
    when:
      changedAny: [source]
      kinExists: [story]
    require:
      kinChanged: [story]
    message: "sync story"
`
	if err := os.MkdirAll(filepath.Join(dir, "libs/ui/button"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "kyn.config.yaml"), []byte(cfg), 0o600); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test")

	if err := os.WriteFile(filepath.Join(dir, "libs/ui/button/button.component.ts"), []byte("v1"), 0o600); err != nil {
		t.Fatalf("write component: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "libs/ui/button/button.stories.ts"), []byte("v1"), 0o600); err != nil {
		t.Fatalf("write story: %v", err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "base")
	base := strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))

	if err := os.WriteFile(filepath.Join(dir, "libs/ui/button/button.component.ts"), []byte("v2"), 0o600); err != nil {
		t.Fatalf("write component: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "libs/ui/button/button.stories.ts"), []byte("v2"), 0o600); err != nil {
		t.Fatalf("write story: %v", err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "head")

	t.Setenv("KYN_BASE_REF", base)
	t.Setenv("KYN_HEAD_REF", "HEAD")
	code := runWithArgs(t, []string{
		"kyn", "check",
		"--cwd", dir,
		"--config", "kyn.config.yaml",
	})
	if code != ExitOK {
		t.Fatalf("expected exit %d, got %d", ExitOK, code)
	}
}

func TestExecuteDryRunReturnsZeroOnWouldFail(t *testing.T) {
	dir := t.TempDir()
	cfg := `version: 1
families:
  - id: angular-component
    include:
      - "libs/**/*.component.ts"
    baseName:
      stripSuffixes:
        - ".component"
    kin:
      story: "{dir}/{base}.stories.ts"
rules:
  - id: storybook-sync
    family: angular-component
    severity: error
    when:
      changedAny: [source]
      kinExists: [story]
    require:
      kinChanged: [story]
    message: "sync story"
`
	if err := os.MkdirAll(filepath.Join(dir, "libs/ui/button"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "kyn.config.yaml"), []byte(cfg), 0o600); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "libs/ui/button/button.component.ts"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write component: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "libs/ui/button/button.stories.ts"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write story: %v", err)
	}

	code := runWithArgs(t, []string{
		"kyn", "check",
		"--cwd", dir,
		"--config", "kyn.config.yaml",
		"--files", "libs/ui/button/button.component.ts",
		"--dry-run-resolve",
	})
	if code != ExitOK {
		t.Fatalf("expected exit %d, got %d", ExitOK, code)
	}
}

func TestExecuteExplainReturnsZeroOnRuleFailure(t *testing.T) {
	dir := t.TempDir()
	cfg := `version: 1
families:
  - id: angular-component
    include:
      - "libs/**/*.component.ts"
    baseName:
      stripSuffixes:
        - ".component"
    kin:
      story: "{dir}/{base}.stories.ts"
rules:
  - id: storybook-sync
    family: angular-component
    severity: error
    when:
      changedAny: [source]
      kinExists: [story]
    require:
      kinChanged: [story]
    message: "sync story"
`
	if err := os.MkdirAll(filepath.Join(dir, "libs/ui/button"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "kyn.config.yaml"), []byte(cfg), 0o600); err != nil {
		t.Fatalf("write cfg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "libs/ui/button/button.component.ts"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write component: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "libs/ui/button/button.stories.ts"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write story: %v", err)
	}

	code := runWithArgs(t, []string{
		"kyn", "explain",
		"--cwd", dir,
		"--config", "kyn.config.yaml",
		"--files", "libs/ui/button/button.component.ts",
	})
	if code != ExitOK {
		t.Fatalf("expected exit %d, got %d", ExitOK, code)
	}
}

func TestExecuteExplainStrictModeNoInputFails(t *testing.T) {
	dir := t.TempDir()
	cfg := `version: 1
families: []
rules: []
`
	if err := os.WriteFile(filepath.Join(dir, "kyn.config.yaml"), []byte(cfg), 0o600); err != nil {
		t.Fatalf("write cfg: %v", err)
	}

	code := runWithArgs(t, []string{
		"kyn", "explain",
		"--cwd", dir,
		"--config", "kyn.config.yaml",
		"--strict-input-mode",
	})
	if code != ExitUsage {
		t.Fatalf("expected exit %d, got %d", ExitUsage, code)
	}
}

func runWithArgs(t *testing.T, args []string) int {
	t.Helper()
	prev := os.Args
	t.Cleanup(func() { os.Args = prev })
	os.Args = args
	return Execute()
}

func TestErrorHelpers(t *testing.T) {
	u := usageError("bad %s", "input")
	if u.Error() == "" {
		t.Fatalf("expected usage error message")
	}
	r := runtimeError("runtime %s", "issue")
	if r.Error() == "" {
		t.Fatalf("expected runtime error message")
	}
}

func runGit(t *testing.T, cwd string, args ...string) string {
	t.Helper()
	prefix := []string{"-C", cwd, "-c", "commit.gpgsign=false"}
	cmd := exec.Command("git", append(prefix, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
	return string(out)
}

package cli

import (
	"os"
	"path/filepath"
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

func runWithArgs(t *testing.T, args []string) int {
	t.Helper()
	prev := os.Args
	t.Cleanup(func() { os.Args = prev })
	os.Args = args
	return Execute()
}

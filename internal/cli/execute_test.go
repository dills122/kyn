package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"kyn/internal/config"
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

func TestExecuteInitCreatesConfig(t *testing.T) {
	dir := t.TempDir()

	code := runWithArgs(t, []string{
		"kyn", "init",
		"--cwd", dir,
	})
	if code != ExitOK {
		t.Fatalf("expected exit %d, got %d", ExitOK, code)
	}

	target := filepath.Join(dir, "kyn.config.yaml")
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}
}

func TestExecuteInitExistingWithoutForceFails(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "kyn.config.yaml")
	if err := os.WriteFile(target, []byte("version: 2\n"), 0o600); err != nil {
		t.Fatalf("write existing config: %v", err)
	}

	code := runWithArgs(t, []string{
		"kyn", "init",
		"--cwd", dir,
	})
	if code != ExitUsage {
		t.Fatalf("expected exit %d, got %d", ExitUsage, code)
	}
}

func TestExecuteInitSupportedPresets(t *testing.T) {
	presets := []string{"web-ui", "api", "proto", "iac"}
	for _, preset := range presets {
		t.Run(preset, func(t *testing.T) {
			dir := t.TempDir()
			target := preset + ".yaml"

			code := runWithArgs(t, []string{
				"kyn", "init",
				"--cwd", dir,
				"--config", target,
				"--preset", preset,
			})
			if code != ExitOK {
				t.Fatalf("expected exit %d, got %d", ExitOK, code)
			}

			cfg, _, err := config.Load(dir, target)
			if err != nil {
				t.Fatalf("load generated config: %v", err)
			}
			if cfg.Version != 2 {
				t.Fatalf("expected version 2, got %d", cfg.Version)
			}
			if len(cfg.Families) == 0 || len(cfg.Rules) == 0 {
				t.Fatalf("expected non-empty generated config for preset %s", preset)
			}
		})
	}
}

func TestExecuteInitInvalidPresetFails(t *testing.T) {
	dir := t.TempDir()

	code := runWithArgs(t, []string{
		"kyn", "init",
		"--cwd", dir,
		"--preset", "unknown",
	})
	if code != ExitUsage {
		t.Fatalf("expected exit %d, got %d", ExitUsage, code)
	}
}

func TestExecuteConfigMigrateWritesOutput(t *testing.T) {
	dir := t.TempDir()
	cfg := `version: 1
families:
  - id: angular-component
    include:
      - "libs/**/*.component.ts"
    kin:
      story: "{dir}/{base}.stories.ts"
rules:
  - id: storybook-sync
    family: angular-component
    severity: error
    when:
      changedAny: [source]
    require:
      kinChanged: [story]
      emitFlag: figmaPublishRequired
    message: "sync story"
`
	inPath := filepath.Join(dir, "kyn.config.yaml")
	if err := os.WriteFile(inPath, []byte(cfg), 0o600); err != nil {
		t.Fatalf("write cfg: %v", err)
	}

	code := runWithArgs(t, []string{
		"kyn", "config", "migrate",
		"--cwd", dir,
		"--config", "kyn.config.yaml",
		"--from", "v1",
		"--to", "v2",
	})
	if code != ExitOK {
		t.Fatalf("expected exit %d, got %d", ExitOK, code)
	}

	outPath := filepath.Join(dir, "kyn.config.v2.yaml")
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected migrated config to exist: %v", err)
	}
	outCfg, _, err := config.Load(dir, "kyn.config.v2.yaml")
	if err != nil {
		t.Fatalf("load migrated config: %v", err)
	}
	if outCfg.Version != 2 {
		t.Fatalf("expected migrated version 2, got %d", outCfg.Version)
	}
	if len(outCfg.Rules) != 1 || len(outCfg.Rules[0].Actions.Emit) != 1 {
		t.Fatalf("expected migrated rule emit actions, got %+v", outCfg.Rules)
	}
}

func TestExecuteConfigMigrateInPlaceCreatesBackup(t *testing.T) {
	dir := t.TempDir()
	cfg := `version: 1
families:
  - id: angular-component
    include:
      - "libs/**/*.component.ts"
    kin:
      story: "{dir}/{base}.stories.ts"
rules:
  - id: storybook-sync
    family: angular-component
    severity: error
    when:
      changedAny: [source]
    require:
      kinChanged: [story]
    message: "sync story"
`
	inPath := filepath.Join(dir, "kyn.config.yaml")
	if err := os.WriteFile(inPath, []byte(cfg), 0o600); err != nil {
		t.Fatalf("write cfg: %v", err)
	}

	code := runWithArgs(t, []string{
		"kyn", "config", "migrate",
		"--cwd", dir,
		"--config", "kyn.config.yaml",
		"--from", "v1",
		"--to", "v2",
		"--in-place",
	})
	if code != ExitOK {
		t.Fatalf("expected exit %d, got %d", ExitOK, code)
	}

	if _, err := os.Stat(filepath.Join(dir, "kyn.config.yaml.bak")); err != nil {
		t.Fatalf("expected backup file to exist: %v", err)
	}
	outCfg, _, err := config.Load(dir, "kyn.config.yaml")
	if err != nil {
		t.Fatalf("load migrated in-place config: %v", err)
	}
	if outCfg.Version != 2 {
		t.Fatalf("expected migrated version 2, got %d", outCfg.Version)
	}
}

func TestExecuteCheckSARIFReturnsRuleFailure(t *testing.T) {
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
		"--format", "sarif",
	})
	if code != ExitRuleFailure {
		t.Fatalf("expected exit %d, got %d", ExitRuleFailure, code)
	}
}

func TestExecuteExplainSARIFReturnsUsage(t *testing.T) {
	dir := t.TempDir()
	cfg := `version: 1
families:
  - id: angular-component
    include:
      - "libs/**/*.component.ts"
    kin:
      story: "{dir}/{base}.stories.ts"
rules:
  - id: storybook-sync
    family: angular-component
    severity: error
    when:
      changedAny: [source]
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

	code := runWithArgs(t, []string{
		"kyn", "explain",
		"--cwd", dir,
		"--config", "kyn.config.yaml",
		"--files", "libs/ui/button/button.component.ts",
		"--format", "sarif",
	})
	if code != ExitUsage {
		t.Fatalf("expected exit %d, got %d", ExitUsage, code)
	}
}

func TestExecuteCheckRDJSONReturnsRuleFailure(t *testing.T) {
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
		"--format", "rdjson",
	})
	if code != ExitRuleFailure {
		t.Fatalf("expected exit %d, got %d", ExitRuleFailure, code)
	}
}

func TestExecuteExplainRDJSONReturnsUsage(t *testing.T) {
	dir := t.TempDir()
	cfg := `version: 1
families:
  - id: angular-component
    include:
      - "libs/**/*.component.ts"
    kin:
      story: "{dir}/{base}.stories.ts"
rules:
  - id: storybook-sync
    family: angular-component
    severity: error
    when:
      changedAny: [source]
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

	code := runWithArgs(t, []string{
		"kyn", "explain",
		"--cwd", dir,
		"--config", "kyn.config.yaml",
		"--files", "libs/ui/button/button.component.ts",
		"--format", "rdjson",
	})
	if code != ExitUsage {
		t.Fatalf("expected exit %d, got %d", ExitUsage, code)
	}
}

func TestExecuteCheckCheckstyleReturnsRuleFailure(t *testing.T) {
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
		"--format", "checkstyle",
	})
	if code != ExitRuleFailure {
		t.Fatalf("expected exit %d, got %d", ExitRuleFailure, code)
	}
}

func TestExecuteExplainCheckstyleReturnsUsage(t *testing.T) {
	dir := t.TempDir()
	cfg := `version: 1
families:
  - id: angular-component
    include:
      - "libs/**/*.component.ts"
    kin:
      story: "{dir}/{base}.stories.ts"
rules:
  - id: storybook-sync
    family: angular-component
    severity: error
    when:
      changedAny: [source]
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

	code := runWithArgs(t, []string{
		"kyn", "explain",
		"--cwd", dir,
		"--config", "kyn.config.yaml",
		"--files", "libs/ui/button/button.component.ts",
		"--format", "checkstyle",
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

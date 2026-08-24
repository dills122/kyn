package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestVerboseMetadataUsesStderrAndPreservesJSONStdout(t *testing.T) {
	dir := t.TempDir()
	configYAML := `version: 2
families:
  - id: web-component
    groups:
      source:
        include: ["src/*.ts"]
    kin:
      story: "{dir}/{base}.stories.ts"
rules: []
`
	if err := os.WriteFile(filepath.Join(dir, "kyn.config.yaml"), []byte(configYAML), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	tests := []struct {
		name       string
		newCommand func() *cobra.Command
	}{
		{name: "check", newCommand: newCheckCommand},
		{name: "explain", newCommand: newExplainCommand},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.newCommand()
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs([]string{
				"--cwd", dir,
				"--files", "src/button.ts",
				"--format", "json",
				"--verbose",
			})

			if err := cmd.Execute(); err != nil {
				t.Fatalf("execute: %v", err)
			}
			if !json.Valid(stdout.Bytes()) {
				t.Fatalf("stdout is not valid JSON:\n%s", stdout.String())
			}
			if strings.Contains(stdout.String(), "config=") {
				t.Fatalf("stdout contains verbose metadata:\n%s", stdout.String())
			}
			if !strings.Contains(stderr.String(), "config=") {
				t.Fatalf("stderr does not contain verbose metadata:\n%s", stderr.String())
			}
		})
	}
}

func TestRootHelpShowsGoldenPath(t *testing.T) {
	cmd := newRootCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{
		"kyn init --preset web-ui",
		"kyn check --dry-run-resolve",
		"kyn check",
		"kyn explain",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("help does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestInitSuggestsReviewAndDryRun(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "repo $with % and ' quote")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create cwd: %v", err)
	}
	configPath := "config files/kyn's $config%.yaml"
	cmd := newInitCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--cwd", dir, "--config", configPath, "--preset", "web-ui"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	quote := func(value string) string {
		if runtime.GOOS == "windows" {
			return `'` + strings.ReplaceAll(value, `'`, `''`) + `'`
		}
		return `'` + strings.ReplaceAll(value, `'`, `'"'"'`) + `'`
	}
	quotedCWD := quote(dir)
	quotedConfig := quote(configPath)
	previous := -1
	for _, want := range []string{
		"Next steps (" + commandShellName() + "):",
		"1. Review " + quotedConfig,
		"2. Preview: kyn check --cwd " + quotedCWD + " -c " + quotedConfig + " --dry-run-resolve -f path/to/source-file",
		"3. Enforce: kyn check --cwd " + quotedCWD + " -c " + quotedConfig,
		"4. Diagnose: kyn explain --cwd " + quotedCWD + " -c " + quotedConfig,
	} {
		position := strings.Index(stdout.String(), want)
		if position == -1 {
			t.Fatalf("init output does not contain %q:\n%s", want, stdout.String())
		}
		if position <= previous {
			t.Fatalf("init output does not present %q in order:\n%s", want, stdout.String())
		}
		previous = position
	}

	for _, newCommand := range []func() *cobra.Command{newCheckCommand, newExplainCommand} {
		guided := newCommand()
		guided.SetOut(&bytes.Buffer{})
		guided.SetErr(&bytes.Buffer{})
		guided.SetArgs([]string{
			"--cwd", dir,
			"--config", configPath,
			"--files", "src/button.component.ts",
			"--dry-run-resolve",
		})
		if guided.Name() == "explain" {
			guided.SetArgs([]string{
				"--cwd", dir,
				"--config", configPath,
				"--files", "src/button.component.ts",
			})
		}
		if err := guided.Execute(); err != nil {
			t.Fatalf("execute guided %s command: %v", guided.Name(), err)
		}
	}
}

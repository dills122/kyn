package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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
	dir := t.TempDir()
	cmd := newInitCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--cwd", dir, "--preset", "web-ui"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{
		"Review kyn.config.yaml",
		"kyn check -c kyn.config.yaml --dry-run-resolve -f path/to/source-file",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("init output does not contain %q:\n%s", want, stdout.String())
		}
	}
}

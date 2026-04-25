package rules

import (
	"os"
	"path/filepath"
	"testing"

	"kyn/internal/config"
	"kyn/internal/family"
)

func TestEvaluateStoryAndFlag(t *testing.T) {
	cwd := t.TempDir()
	mustWrite(t, filepath.Join(cwd, "libs/ui/button/button.stories.ts"))
	mustWrite(t, filepath.Join(cwd, "libs/ui/button/figma.button.json"))

	cfg := config.Config{
		Version: 1,
		Rules: []config.Rule{
			{
				ID:       "storybook-sync",
				Family:   "angular-component",
				Severity: "error",
				When: config.RuleClauses{
					ChangedAny: []string{"source"},
					KinExists:  []string{"story"},
				},
				Require: config.RuleClauses{
					KinChanged: []string{"story"},
				},
				Message: "Component changed but story did not.",
			},
			{
				ID:       "figma-publish-required",
				Family:   "angular-component",
				Severity: "warn",
				When: config.RuleClauses{
					ChangedAny: []string{"source"},
					KinExists:  []string{"figma"},
				},
				Require: config.RuleClauses{
					EmitFlag: "figmaPublishRequired",
				},
				Message: "Figma publish may be required.",
			},
		},
	}

	inst := family.Instance{
		FamilyID:    "angular-component",
		Name:        "libs/ui/button/button",
		SourceFiles: []string{"libs/ui/button/button.component.html", "libs/ui/button/button.component.ts"},
		Kin: map[string]string{
			"story": "libs/ui/button/button.stories.ts",
			"figma": "libs/ui/button/figma.button.json",
		},
	}
	changed := map[string]struct{}{
		"libs/ui/button/button.component.ts":   {},
		"libs/ui/button/button.component.html": {},
	}

	summary, err := Evaluate(EvalInput{
		Cwd:       cwd,
		FailOn:    "error",
		Changed:   changed,
		Rules:     cfg.Rules,
		Instances: []family.Instance{inst},
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if summary.OK {
		t.Fatalf("expected summary to fail")
	}
	if summary.Failed != 1 {
		t.Fatalf("expected 1 failed result, got %d", summary.Failed)
	}
	if summary.Errors != 1 {
		t.Fatalf("expected 1 error severity result, got %d", summary.Errors)
	}
	if summary.Warnings != 1 {
		t.Fatalf("expected 1 warning severity result, got %d", summary.Warnings)
	}
	if len(summary.Flags) != 1 || summary.Flags[0] != "figmaPublishRequired" {
		t.Fatalf("unexpected flags: %v", summary.Flags)
	}
}

func TestEvaluateFailOnEmpty(t *testing.T) {
	summary, err := Evaluate(EvalInput{
		Cwd:         t.TempDir(),
		FailOn:      "error",
		FailOnEmpty: true,
		Changed:     map[string]struct{}{},
		Rules:       nil,
		Instances:   nil,
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if summary.OK {
		t.Fatalf("expected summary to fail when fail-on-empty is set")
	}
	if summary.Failed != 1 {
		t.Fatalf("expected failed=1, got %d", summary.Failed)
	}
}

func mustWrite(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
}

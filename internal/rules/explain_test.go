package rules

import (
	"path/filepath"
	"testing"

	"kyn/internal/changes"
	"kyn/internal/config"
	"kyn/internal/family"
)

func TestExplain_StatusesAndFlags(t *testing.T) {
	cwd := t.TempDir()
	mustWrite(t, filepath.Join(cwd, "libs/ui/button/button.stories.ts"))
	mustWrite(t, filepath.Join(cwd, "libs/ui/button/figma.button.json"))

	inst := family.Instance{
		FamilyID:    "angular-component",
		Name:        "libs/ui/button/button",
		SourceFiles: []string{"libs/ui/button/button.component.ts"},
		Kin: map[string]string{
			"story": "libs/ui/button/button.stories.ts",
			"figma": "libs/ui/button/figma.button.json",
		},
	}
	changed := map[string]struct{}{
		"libs/ui/button/button.component.ts": {},
	}

	summary, err := Explain(EvalInput{
		Cwd:     cwd,
		FailOn:  "error",
		Changed: changed,
		StatusByFile: map[string]changes.Status{
			"libs/ui/button/button.component.ts": changes.StatusModified,
		},
		Instances: []family.Instance{inst},
		Rules: []config.Rule{
			{
				ID:       "story-sync",
				Family:   "angular-component",
				Severity: "error",
				If: config.RuleClauses{
					ChangedAny: []string{"source"},
					KinExists:  []string{"story"},
				},
				Assert: config.RuleClauses{
					KinChanged: []string{"story"},
				},
				Message: "Story was not updated.",
			},
			{
				ID:       "figma-flag",
				Family:   "angular-component",
				Severity: "warn",
				If: config.RuleClauses{
					ChangedStatusAny: []string{"modified"},
					KinExists:        []string{"figma"},
				},
				Actions: config.RuleActions{
					Emit: []string{"figmaPublishRequired"},
				},
				Message: "Figma publish may be required.",
			},
			{
				ID:       "renamed-only",
				Family:   "angular-component",
				Severity: "warn",
				If: config.RuleClauses{
					ChangedStatusAny: []string{"renamed"},
				},
				Message: "Only for renamed changes.",
			},
		},
	})
	if err != nil {
		t.Fatalf("Explain returned error: %v", err)
	}

	if summary.Failed != 1 {
		t.Fatalf("failed=%d want 1", summary.Failed)
	}
	if summary.Infos != 0 {
		t.Fatalf("infos=%d want 0", summary.Infos)
	}
	if summary.Skipped != 1 {
		t.Fatalf("skipped=%d want 1", summary.Skipped)
	}
	if len(summary.Flags) != 1 || summary.Flags[0] != "figmaPublishRequired" {
		t.Fatalf("unexpected flags: %v", summary.Flags)
	}
	if len(summary.Results) != 3 {
		t.Fatalf("results=%d want 3", len(summary.Results))
	}
}

func TestExplain_FailOnEmpty(t *testing.T) {
	summary, err := Explain(EvalInput{
		Cwd:         t.TempDir(),
		FailOn:      "error",
		FailOnEmpty: true,
		Changed:     map[string]struct{}{},
	})
	if err != nil {
		t.Fatalf("Explain returned error: %v", err)
	}
	if summary.OK {
		t.Fatal("expected summary to fail when fail-on-empty is set")
	}
	if summary.Failed != 1 {
		t.Fatalf("failed=%d want 1", summary.Failed)
	}
}

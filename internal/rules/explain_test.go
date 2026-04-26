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

func TestExplain_AssertTraceBranches(t *testing.T) {
	cwd := t.TempDir()
	mustWrite(t, filepath.Join(cwd, "libs/ui/button/button.exists.ts"))

	inst := family.Instance{
		FamilyID:    "angular-component",
		Name:        "libs/ui/button/button",
		SourceFiles: []string{"libs/ui/button/button.component.ts"},
		Kin: map[string]string{
			"story":      "libs/ui/button/button.stories.ts",
			"spec":       "libs/ui/button/button.spec.ts",
			"missing":    "libs/ui/button/button.missing.ts",
			"exists":     "libs/ui/button/button.exists.ts",
			"otherstory": "libs/ui/button/other.stories.ts",
		},
	}
	changed := map[string]struct{}{
		"libs/ui/button/button.component.ts": {},
		"libs/ui/button/button.spec.ts":      {},
	}

	summary, err := Explain(EvalInput{
		Cwd:       cwd,
		FailOn:    "warn",
		Changed:   changed,
		Instances: []family.Instance{inst},
		Rules: []config.Rule{
			{
				ID:       "assert-branches",
				Family:   "angular-component",
				Severity: "warn",
				If: config.RuleClauses{
					ChangedAny: []string{"source"},
				},
				Assert: config.RuleClauses{
					KinChanged:   []string{"story"},
					KinUnchanged: []string{"spec"},
					KinExists:    []string{"missing"},
					KinMissing:   []string{"exists"},
				},
				Actions: config.RuleActions{
					Emit: []string{"flagA", ""},
				},
				Message: "exercise assert branches",
			},
		},
	})
	if err != nil {
		t.Fatalf("Explain returned error: %v", err)
	}
	if len(summary.Results) != 1 {
		t.Fatalf("results=%d want 1", len(summary.Results))
	}
	result := summary.Results[0]
	if result.Status != ExplainStatusFail {
		t.Fatalf("status=%s want fail", result.Status)
	}
	if len(result.AssertTrace) < 5 {
		t.Fatalf("expected multiple assert traces, got %v", result.AssertTrace)
	}
	if len(result.EmittedFlags) != 2 {
		t.Fatalf("expected raw emitted flags slice to preserve entries, got %v", result.EmittedFlags)
	}
	if len(summary.Flags) != 1 || summary.Flags[0] != "flagA" {
		t.Fatalf("unexpected deduped summary flags: %v", summary.Flags)
	}
}

func TestExplain_WhenTraceBranchesAndHelpers(t *testing.T) {
	cwd := t.TempDir()
	mustWrite(t, filepath.Join(cwd, "libs/ui/button/button.exists.ts"))

	inst := family.Instance{
		FamilyID:    "angular-component",
		Name:        "libs/ui/button/button",
		SourceFiles: []string{"libs/ui/button/button.component.ts"},
		Kin: map[string]string{
			"exists": "libs/ui/button/button.exists.ts",
			"ghost":  "libs/ui/button/button.ghost.ts",
		},
	}

	summary, err := Explain(EvalInput{
		Cwd:    cwd,
		FailOn: "error",
		Changed: map[string]struct{}{
			"libs/ui/button/button.component.ts": {},
		},
		StatusByFile: map[string]changes.Status{
			"libs/ui/button/button.component.ts": changes.StatusAdded,
		},
		Instances: []family.Instance{inst},
		Rules: []config.Rule{
			{
				ID:       "missing-source",
				Family:   "other-family",
				Severity: "warn",
				Message:  "ignored by family mismatch",
			},
			{
				ID:       "when-pass",
				Family:   "angular-component",
				Severity: "info",
				If: config.RuleClauses{
					ChangedAny:       []string{"source"},
					ChangedStatusAny: []string{"added"},
					KinExists:        []string{"exists"},
					KinMissing:       []string{"ghost"},
				},
				Message: "all when clauses pass",
			},
			{
				ID:       "when-fail-missing",
				Family:   "angular-component",
				Severity: "warn",
				If: config.RuleClauses{
					KinExists: []string{"ghost"},
				},
				Message: "kin missing should skip",
			},
			{
				ID:       "when-fail-exists",
				Family:   "angular-component",
				Severity: "warn",
				If: config.RuleClauses{
					KinMissing: []string{"exists"},
				},
				Message: "kin exists should skip",
			},
		},
	})
	if err != nil {
		t.Fatalf("Explain returned error: %v", err)
	}
	if len(summary.Results) != 3 {
		t.Fatalf("results=%d want 3", len(summary.Results))
	}

	if explainStatusRank(ExplainStatusFail) != 0 || explainStatusRank("unknown") != 4 {
		t.Fatalf("unexpected explainStatusRank behavior")
	}
	if got := toRuleResult(ExplainResult{RuleID: "r1", Status: ExplainStatusPass, Message: "x"}, StatusPass); got.RuleID != "r1" || got.Status != StatusPass {
		t.Fatalf("unexpected toRuleResult conversion: %+v", got)
	}
}

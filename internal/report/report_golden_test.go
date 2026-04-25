package report

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"kyn/internal/rules"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func TestRenderText_Default(t *testing.T) {
	summary := sampleSummary()
	got := RenderText(summary, TextOptions{ShowPasses: false})
	assertGolden(t, "text_default.golden", got)
}

func TestRenderText_ShowPasses(t *testing.T) {
	summary := sampleSummary()
	got := RenderText(summary, TextOptions{ShowPasses: true})
	assertGolden(t, "text_show_passes.golden", got)
}

func TestRenderJSON(t *testing.T) {
	summary := sampleSummary()
	out, err := RenderJSON(summary)
	if err != nil {
		t.Fatalf("RenderJSON returned error: %v", err)
	}
	assertGolden(t, "json.golden", string(out)+"\n")
}

func sampleSummary() rules.Summary {
	return rules.Summary{
		OK:       false,
		Passed:   1,
		Failed:   1,
		Infos:    1,
		Warnings: 1,
		Errors:   1,
		Flags:    []string{"alphaFlag", "zetaFlag"},
		Results: []rules.RuleResult{
			{
				RuleID:       "story-sync",
				FamilyID:     "angular-component",
				FamilyName:   "libs/ui/button/button",
				Severity:     rules.SeverityError,
				Status:       rules.StatusFail,
				Message:      "Story was not updated.",
				ChangedFiles: []string{"libs/ui/button/button.component.ts"},
				ExpectedFiles: []string{
					"libs/ui/button/button.stories.ts",
				},
			},
			{
				RuleID:       "figma-flag",
				FamilyID:     "angular-component",
				FamilyName:   "libs/ui/button/button",
				Severity:     rules.SeverityWarn,
				Status:       rules.StatusInfo,
				Message:      "Figma publish may be required.",
				ChangedFiles: []string{"libs/ui/button/button.component.ts"},
			},
			{
				RuleID:       "docs-sync",
				FamilyID:     "angular-component",
				FamilyName:   "libs/ui/button/button",
				Severity:     rules.SeverityInfo,
				Status:       rules.StatusPass,
				Message:      "Docs updated.",
				ChangedFiles: []string{"libs/ui/button/button.docs.md"},
			},
		},
	}
}

func assertGolden(t *testing.T, name string, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *updateGolden {
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if got != string(want) {
		t.Fatalf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", name, got, string(want))
	}
}

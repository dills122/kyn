package report

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
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

func TestRenderText_SummaryOnly(t *testing.T) {
	summary := sampleSummary()
	got := RenderText(summary, TextOptions{SummaryOnly: true})
	assertGolden(t, "text_summary_only.golden", got)
}

func TestRenderJSON(t *testing.T) {
	summary := sampleSummary()
	out, err := RenderJSON(summary)
	if err != nil {
		t.Fatalf("RenderJSON returned error: %v", err)
	}
	assertGolden(t, "json.golden", string(out)+"\n")
}

func TestRenderJSONSummary(t *testing.T) {
	summary := sampleSummary()
	out, err := RenderJSONSummary(summary)
	if err != nil {
		t.Fatalf("RenderJSONSummary returned error: %v", err)
	}
	assertGolden(t, "json_summary_only.golden", string(out)+"\n")
}

func TestRenderResolveText(t *testing.T) {
	resolve := sampleResolveReport()
	got := RenderResolveText(resolve)
	assertGolden(t, "dry_run_text.golden", got)
}

func TestRenderResolveJSON(t *testing.T) {
	resolve := sampleResolveReport()
	out, err := RenderResolveJSON(resolve)
	if err != nil {
		t.Fatalf("RenderResolveJSON returned error: %v", err)
	}
	assertGolden(t, "dry_run_json.golden", string(out)+"\n")
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

func sampleResolveReport() ResolveReport {
	return ResolveReport{
		Mode:             "git",
		Base:             "origin/main",
		Head:             "HEAD",
		ChangedCount:     2,
		MatchedInstances: 1,
		ChangedFiles: []string{
			"libs/ui/button/button.component.ts",
			"libs/ui/button/button.stories.ts",
		},
		Instances: []ResolveInstance{
			{
				FamilyID:    "angular-component",
				Name:        "libs/ui/button/button",
				SourceFiles: []string{"libs/ui/button/button.component.ts"},
				Kin: map[string]string{
					"spec":  "libs/ui/button/button.spec.ts",
					"story": "libs/ui/button/button.stories.ts",
				},
			},
		},
	}
}

func assertGolden(t *testing.T, name string, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	got = normalizeGoldenText(got)
	if *updateGolden {
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	wantText := normalizeGoldenText(string(want))
	if got != wantText {
		t.Fatalf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", name, got, wantText)
	}
}

func normalizeGoldenText(s string) string {
	// Ensure cross-platform stable snapshots: compare on LF regardless of runner OS.
	return strings.ReplaceAll(s, "\r\n", "\n")
}

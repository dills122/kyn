package report

import (
	"encoding/json"
	"fmt"
	"strings"

	"kyn/internal/rules"
)

func RenderExplainJSON(summary rules.ExplainSummary) ([]byte, error) {
	return json.MarshalIndent(summary, "", "  ")
}

func RenderExplainText(summary rules.ExplainSummary, summaryOnly bool) string {
	var b strings.Builder

	_, _ = fmt.Fprintf(&b, "kyn explain\n\n")
	if summary.OK {
		_, _ = fmt.Fprintf(&b, "PASS\n\n")
	} else {
		_, _ = fmt.Fprintf(&b, "FAIL\n\n")
	}
	_, _ = fmt.Fprintf(&b, "Rules failed: %d\n", summary.Failed)
	_, _ = fmt.Fprintf(&b, "Rules passed: %d\n", summary.Passed)
	_, _ = fmt.Fprintf(&b, "Rules info: %d\n", summary.Infos)
	_, _ = fmt.Fprintf(&b, "Rules skipped: %d\n", summary.Skipped)
	_, _ = fmt.Fprintf(&b, "Warnings: %d\n", summary.Warnings)
	_, _ = fmt.Fprintf(&b, "Errors: %d\n", summary.Errors)

	if summaryOnly {
		if len(summary.Flags) > 0 {
			_, _ = fmt.Fprintf(&b, "\nFlags:\n")
			for _, flag := range summary.Flags {
				_, _ = fmt.Fprintf(&b, "  - %s\n", flag)
			}
		}
		return b.String()
	}

	for _, result := range summary.Results {
		_, _ = fmt.Fprintf(&b, "\n[%s] %s\n", strings.ToUpper(string(result.Severity)), result.RuleID)
		if result.FamilyID != "" {
			_, _ = fmt.Fprintf(&b, "Family: %s\n", result.FamilyID)
		}
		if result.FamilyName != "" {
			_, _ = fmt.Fprintf(&b, "Instance: %s\n", result.FamilyName)
		}
		_, _ = fmt.Fprintf(&b, "Status: %s\n", result.Status)
		_, _ = fmt.Fprintf(&b, "Message: %s\n", result.Message)

		if len(result.IfTrace) > 0 {
			_, _ = fmt.Fprintf(&b, "If:\n")
			for _, trace := range result.IfTrace {
				_, _ = fmt.Fprintf(&b, "  - %s: %s (%s)\n", trace.Clause, trace.Result, trace.Detail)
			}
		}
		if len(result.AssertTrace) > 0 {
			_, _ = fmt.Fprintf(&b, "Assert:\n")
			for _, trace := range result.AssertTrace {
				_, _ = fmt.Fprintf(&b, "  - %s: %s (%s)\n", trace.Clause, trace.Result, trace.Detail)
			}
		}
		if len(result.ChangedFiles) > 0 {
			_, _ = fmt.Fprintf(&b, "Changed files:\n")
			for _, file := range result.ChangedFiles {
				_, _ = fmt.Fprintf(&b, "  - %s\n", file)
			}
		}
		if len(result.ExpectedFiles) > 0 {
			_, _ = fmt.Fprintf(&b, "Expected files:\n")
			for _, file := range result.ExpectedFiles {
				_, _ = fmt.Fprintf(&b, "  - %s\n", file)
			}
		}
		if len(result.EmittedFlags) > 0 {
			_, _ = fmt.Fprintf(&b, "Emitted flags:\n")
			for _, flag := range result.EmittedFlags {
				_, _ = fmt.Fprintf(&b, "  - %s\n", flag)
			}
		}
	}

	if len(summary.Flags) > 0 {
		_, _ = fmt.Fprintf(&b, "\nFlags:\n")
		for _, flag := range summary.Flags {
			_, _ = fmt.Fprintf(&b, "  - %s\n", flag)
		}
	}

	return b.String()
}

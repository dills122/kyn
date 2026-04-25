package report

import (
	"fmt"
	"strings"

	"kyn/internal/rules"
)

func RenderText(summary rules.Summary) string {
	var b strings.Builder

	_, _ = fmt.Fprintf(&b, "kyn check\n\n")
	if summary.OK {
		_, _ = fmt.Fprintf(&b, "PASS\n\n")
	} else {
		_, _ = fmt.Fprintf(&b, "FAIL\n\n")
	}
	_, _ = fmt.Fprintf(&b, "Rules failed: %d\n", summary.Failed)
	_, _ = fmt.Fprintf(&b, "Warnings: %d\n", summary.Warnings)
	_, _ = fmt.Fprintf(&b, "Infos: %d\n", summary.Infos)

	for _, r := range summary.Results {
		_, _ = fmt.Fprintf(&b, "\n[%s] %s\n", strings.ToUpper(string(r.Severity)), r.RuleID)
		if r.FamilyID != "" {
			_, _ = fmt.Fprintf(&b, "Family: %s\n", r.FamilyID)
		}
		if r.FamilyName != "" {
			_, _ = fmt.Fprintf(&b, "Instance: %s\n", r.FamilyName)
		}
		_, _ = fmt.Fprintf(&b, "Status: %s\n", r.Status)
		_, _ = fmt.Fprintf(&b, "Message: %s\n", r.Message)
		if len(r.ChangedFiles) > 0 {
			_, _ = fmt.Fprintf(&b, "Changed files:\n")
			for _, f := range r.ChangedFiles {
				_, _ = fmt.Fprintf(&b, "  - %s\n", f)
			}
		}
		if len(r.ExpectedFiles) > 0 {
			_, _ = fmt.Fprintf(&b, "Expected files:\n")
			for _, f := range r.ExpectedFiles {
				_, _ = fmt.Fprintf(&b, "  - %s\n", f)
			}
		}
	}

	if len(summary.Flags) > 0 {
		_, _ = fmt.Fprintf(&b, "\nFlags:\n")
		for _, f := range summary.Flags {
			_, _ = fmt.Fprintf(&b, "  - %s\n", f)
		}
	}

	return b.String()
}

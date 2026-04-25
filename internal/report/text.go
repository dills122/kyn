package report

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"kyn/internal/rules"
)

type TextOptions struct {
	ShowPasses bool
}

func RenderText(summary rules.Summary, opts TextOptions) string {
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

	results := orderResults(summary.Results, opts.ShowPasses)
	for _, r := range results {
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

func orderResults(in []rules.RuleResult, showPasses bool) []rules.RuleResult {
	results := slices.Clone(in)
	if !showPasses {
		filtered := results[:0]
		for _, r := range results {
			if r.Status == rules.StatusPass {
				continue
			}
			filtered = append(filtered, r)
		}
		results = filtered
	}

	sort.Slice(results, func(i, j int) bool {
		ri := resultRank(results[i])
		rj := resultRank(results[j])
		if ri != rj {
			return ri < rj
		}
		if results[i].FamilyID != results[j].FamilyID {
			return results[i].FamilyID < results[j].FamilyID
		}
		if results[i].FamilyName != results[j].FamilyName {
			return results[i].FamilyName < results[j].FamilyName
		}
		return results[i].RuleID < results[j].RuleID
	})

	return results
}

func resultRank(r rules.RuleResult) int {
	statusRank := map[rules.ResultStatus]int{
		rules.StatusFail: 0,
		rules.StatusInfo: 1,
		rules.StatusPass: 2,
	}
	severityRank := map[rules.Severity]int{
		rules.SeverityError: 0,
		rules.SeverityWarn:  1,
		rules.SeverityInfo:  2,
	}
	return statusRank[r.Status]*10 + severityRank[r.Severity]
}

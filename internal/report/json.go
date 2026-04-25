package report

import (
	"encoding/json"

	"kyn/internal/rules"
)

func RenderJSON(summary rules.Summary) ([]byte, error) {
	return json.MarshalIndent(summary, "", "  ")
}

type summaryOnly struct {
	OK       bool     `json:"ok"`
	Passed   int      `json:"passed"`
	Failed   int      `json:"failed"`
	Infos    int      `json:"infos"`
	Warnings int      `json:"warnings"`
	Errors   int      `json:"errors"`
	Flags    []string `json:"flags,omitempty"`
}

func RenderJSONSummary(summary rules.Summary) ([]byte, error) {
	return json.MarshalIndent(summaryOnly{
		OK:       summary.OK,
		Passed:   summary.Passed,
		Failed:   summary.Failed,
		Infos:    summary.Infos,
		Warnings: summary.Warnings,
		Errors:   summary.Errors,
		Flags:    summary.Flags,
	}, "", "  ")
}

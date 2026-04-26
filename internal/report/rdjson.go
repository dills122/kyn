package report

import (
	"encoding/json"
	"fmt"

	"kyn/internal/rules"
)

type rdjsonReport struct {
	Source      rdjsonSource       `json:"source"`
	Diagnostics []rdjsonDiagnostic `json:"diagnostics"`
}

type rdjsonSource struct {
	Name string `json:"name"`
}

type rdjsonDiagnostic struct {
	Message         string                  `json:"message"`
	Location        rdjsonLocation          `json:"location"`
	Severity        string                  `json:"severity"`
	Source          rdjsonSource            `json:"source"`
	Code            rdjsonCode              `json:"code"`
	OriginalOutput  string                  `json:"originalOutput,omitempty"`
	RelatedLocation []rdjsonRelatedLocation `json:"relatedLocations,omitempty"`
}

type rdjsonCode struct {
	Value string `json:"value"`
}

type rdjsonLocation struct {
	Path  string      `json:"path"`
	Range rdjsonRange `json:"range"`
}

type rdjsonRelatedLocation struct {
	Message  string         `json:"message,omitempty"`
	Location rdjsonLocation `json:"location"`
}

type rdjsonRange struct {
	Start rdjsonPosition `json:"start"`
}

type rdjsonPosition struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func RenderRDJSON(summary rules.Summary) ([]byte, error) {
	diagnostics := make([]rdjsonDiagnostic, 0, len(summary.Results))
	for _, result := range summary.Results {
		if result.Status != rules.StatusFail {
			continue
		}
		locationPath := primaryLocationPath(result)
		diag := rdjsonDiagnostic{
			Message:  renderRDJSONMessage(result),
			Location: newRDJSONLocation(locationPath),
			Severity: rdjsonSeverity(result.Severity),
			Source: rdjsonSource{
				Name: "kyn",
			},
			Code: rdjsonCode{
				Value: result.RuleID,
			},
			OriginalOutput: renderRDJSONOriginalOutput(result),
		}

		for _, file := range result.ChangedFiles {
			if file == "" || file == locationPath {
				continue
			}
			diag.RelatedLocation = append(diag.RelatedLocation, rdjsonRelatedLocation{
				Message:  "changed file",
				Location: newRDJSONLocation(file),
			})
		}
		for _, file := range result.ExpectedFiles {
			if file == "" {
				continue
			}
			diag.RelatedLocation = append(diag.RelatedLocation, rdjsonRelatedLocation{
				Message:  "expected related file",
				Location: newRDJSONLocation(file),
			})
		}

		diagnostics = append(diagnostics, diag)
	}

	report := rdjsonReport{
		Source: rdjsonSource{
			Name: "kyn",
		},
		Diagnostics: diagnostics,
	}
	return json.MarshalIndent(report, "", "  ")
}

func newRDJSONLocation(path string) rdjsonLocation {
	return rdjsonLocation{
		Path: path,
		Range: rdjsonRange{
			Start: rdjsonPosition{
				Line:   1,
				Column: 1,
			},
		},
	}
}

func primaryLocationPath(result rules.RuleResult) string {
	if len(result.ChangedFiles) > 0 && result.ChangedFiles[0] != "" {
		return result.ChangedFiles[0]
	}
	if len(result.ExpectedFiles) > 0 && result.ExpectedFiles[0] != "" {
		return result.ExpectedFiles[0]
	}
	return result.FamilyName
}

func rdjsonSeverity(severity rules.Severity) string {
	switch severity {
	case rules.SeverityError:
		return "ERROR"
	case rules.SeverityWarn:
		return "WARNING"
	default:
		return "INFO"
	}
}

func renderRDJSONMessage(result rules.RuleResult) string {
	if result.FamilyName == "" {
		return result.Message
	}
	return fmt.Sprintf("%s (family instance: %s)", result.Message, result.FamilyName)
}

func renderRDJSONOriginalOutput(result rules.RuleResult) string {
	location := primaryLocationPath(result)
	if location == "" {
		return result.Message
	}
	return fmt.Sprintf("%s:1:1: [%s] %s", location, result.RuleID, result.Message)
}

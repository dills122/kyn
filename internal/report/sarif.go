package report

import (
	"encoding/json"
	"fmt"
	"sort"

	"kyn/internal/rules"
)

type sarifReport struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string      `json:"name"`
	InformationURI string      `json:"informationUri,omitempty"`
	Version        string      `json:"version,omitempty"`
	Rules          []sarifRule `json:"rules,omitempty"`
}

type sarifRule struct {
	ID               string              `json:"id"`
	ShortDescription sarifMessage        `json:"shortDescription"`
	FullDescription  sarifMessage        `json:"fullDescription,omitempty"`
	Properties       sarifRuleProperties `json:"properties,omitempty"`
	DefaultConfig    sarifDefaultConfig  `json:"defaultConfiguration,omitempty"`
}

type sarifRuleProperties struct {
	FamilyID string `json:"familyId,omitempty"`
}

type sarifDefaultConfig struct {
	Level string `json:"level,omitempty"`
}

type sarifResult struct {
	RuleID     string          `json:"ruleId"`
	Level      string          `json:"level"`
	Message    sarifMessage    `json:"message"`
	Locations  []sarifLocation `json:"locations,omitempty"`
	Related    []sarifLocation `json:"relatedLocations,omitempty"`
	Properties map[string]any  `json:"properties,omitempty"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
	Message          sarifMessage          `json:"message,omitempty"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

func RenderSARIF(summary rules.Summary) ([]byte, error) {
	ruleMap := make(map[string]sarifRule, len(summary.Results))
	results := make([]sarifResult, 0, len(summary.Results))

	for _, result := range summary.Results {
		if _, exists := ruleMap[result.RuleID]; !exists {
			ruleMap[result.RuleID] = sarifRule{
				ID:               result.RuleID,
				ShortDescription: sarifMessage{Text: result.Message},
				FullDescription:  sarifMessage{Text: result.Message},
				Properties: sarifRuleProperties{
					FamilyID: result.FamilyID,
				},
				DefaultConfig: sarifDefaultConfig{
					Level: sarifLevel(result.Severity),
				},
			}
		}

		sarifResult := sarifResult{
			RuleID:  result.RuleID,
			Level:   sarifLevel(result.Severity),
			Message: sarifMessage{Text: renderSARIFMessage(result)},
			Properties: map[string]any{
				"familyId":   result.FamilyID,
				"familyName": result.FamilyName,
				"status":     string(result.Status),
				"severity":   string(result.Severity),
			},
		}

		if len(result.ChangedFiles) > 0 {
			sarifResult.Locations = make([]sarifLocation, 0, len(result.ChangedFiles))
			for _, file := range result.ChangedFiles {
				sarifResult.Locations = append(sarifResult.Locations, sarifLocation{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{URI: file},
					},
					Message: sarifMessage{Text: "changed file"},
				})
			}
		}

		if len(result.ExpectedFiles) > 0 {
			sarifResult.Related = make([]sarifLocation, 0, len(result.ExpectedFiles))
			for _, file := range result.ExpectedFiles {
				sarifResult.Related = append(sarifResult.Related, sarifLocation{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{URI: file},
					},
					Message: sarifMessage{Text: "expected related file"},
				})
			}
			sarifResult.Properties["expectedFiles"] = result.ExpectedFiles
		}

		results = append(results, sarifResult)
	}

	ruleIDs := make([]string, 0, len(ruleMap))
	for id := range ruleMap {
		ruleIDs = append(ruleIDs, id)
	}
	sort.Strings(ruleIDs)

	toolRules := make([]sarifRule, 0, len(ruleIDs))
	for _, id := range ruleIDs {
		toolRules = append(toolRules, ruleMap[id])
	}

	report := sarifReport{
		Version: "2.1.0",
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:           "kyn",
						InformationURI: "https://github.com/",
						Rules:          toolRules,
					},
				},
				Results: results,
			},
		},
	}

	return json.MarshalIndent(report, "", "  ")
}

func sarifLevel(severity rules.Severity) string {
	switch severity {
	case rules.SeverityError:
		return "error"
	case rules.SeverityWarn:
		return "warning"
	default:
		return "note"
	}
}

func renderSARIFMessage(result rules.RuleResult) string {
	if result.FamilyName == "" {
		return result.Message
	}
	return fmt.Sprintf("%s (family instance: %s)", result.Message, result.FamilyName)
}

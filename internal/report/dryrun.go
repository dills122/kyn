package report

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"kyn/internal/family"
)

type ResolveReport struct {
	Mode             string            `json:"mode"`
	Base             string            `json:"base,omitempty"`
	Head             string            `json:"head,omitempty"`
	ChangedCount     int               `json:"changedCount"`
	MatchedInstances int               `json:"matchedInstances"`
	ChangedFiles     []string          `json:"changedFiles,omitempty"`
	Instances        []ResolveInstance `json:"instances,omitempty"`
}

type ResolveInstance struct {
	FamilyID    string            `json:"familyId"`
	Name        string            `json:"name"`
	SourceFiles []string          `json:"sourceFiles,omitempty"`
	Kin         map[string]string `json:"kin,omitempty"`
}

func NewResolveReport(
	mode string,
	base string,
	head string,
	changedFiles []string,
	instances []family.Instance,
	summaryOnly bool,
) ResolveReport {
	report := ResolveReport{
		Mode:             mode,
		Base:             base,
		Head:             head,
		ChangedCount:     len(changedFiles),
		MatchedInstances: len(instances),
	}
	if !summaryOnly {
		report.ChangedFiles = changedFiles
		report.Instances = make([]ResolveInstance, 0, len(instances))
		for _, inst := range instances {
			report.Instances = append(report.Instances, ResolveInstance{
				FamilyID:    inst.FamilyID,
				Name:        inst.Name,
				SourceFiles: inst.SourceFiles,
				Kin:         inst.Kin,
			})
		}
	}
	return report
}

func RenderResolveJSON(resolve ResolveReport) ([]byte, error) {
	return json.MarshalIndent(resolve, "", "  ")
}

func RenderResolveText(resolve ResolveReport) string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "kyn check --dry-run-resolve\n\n")
	if resolve.Mode == "git" {
		_, _ = fmt.Fprintf(&b, "Mode: %s (%s...%s)\n", resolve.Mode, resolve.Base, resolve.Head)
	} else {
		_, _ = fmt.Fprintf(&b, "Mode: %s\n", resolve.Mode)
	}
	_, _ = fmt.Fprintf(&b, "Changed files: %d\n", resolve.ChangedCount)
	_, _ = fmt.Fprintf(&b, "Matched instances: %d\n", resolve.MatchedInstances)

	if len(resolve.ChangedFiles) > 0 {
		_, _ = fmt.Fprintf(&b, "\nChanged file list:\n")
		for _, file := range resolve.ChangedFiles {
			_, _ = fmt.Fprintf(&b, "  - %s\n", file)
		}
	}

	for _, inst := range resolve.Instances {
		_, _ = fmt.Fprintf(&b, "\n[%s] %s\n", inst.FamilyID, inst.Name)
		if len(inst.SourceFiles) > 0 {
			_, _ = fmt.Fprintf(&b, "Source files:\n")
			for _, file := range inst.SourceFiles {
				_, _ = fmt.Fprintf(&b, "  - %s\n", file)
			}
		}
		if len(inst.Kin) > 0 {
			_, _ = fmt.Fprintf(&b, "Kin:\n")
			for _, kinName := range sortedStringMapKeys(inst.Kin) {
				_, _ = fmt.Fprintf(&b, "  - %s: %s\n", kinName, inst.Kin[kinName])
			}
		}
	}

	return b.String()
}

func sortedStringMapKeys(input map[string]string) []string {
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	// Keep output deterministic for tests and CI logs.
	sort.Strings(keys)
	return keys
}

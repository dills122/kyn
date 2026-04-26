package report

import (
	"encoding/xml"
	"sort"

	"kyn/internal/rules"
)

type checkstyleReport struct {
	XMLName xml.Name         `xml:"checkstyle"`
	Version string           `xml:"version,attr"`
	Files   []checkstyleFile `xml:"file"`
}

type checkstyleFile struct {
	Name   string            `xml:"name,attr"`
	Errors []checkstyleError `xml:"error"`
}

type checkstyleError struct {
	Line     int    `xml:"line,attr"`
	Column   int    `xml:"column,attr"`
	Severity string `xml:"severity,attr"`
	Message  string `xml:"message,attr"`
	Source   string `xml:"source,attr"`
}

func RenderCheckstyle(summary rules.Summary) ([]byte, error) {
	grouped := make(map[string][]checkstyleError)
	for _, result := range summary.Results {
		if result.Status != rules.StatusFail {
			continue
		}
		path := primaryLocationPath(result)
		if path == "" {
			path = result.RuleID
		}
		grouped[path] = append(grouped[path], checkstyleError{
			Line:     1,
			Column:   1,
			Severity: checkstyleSeverity(result.Severity),
			Message:  renderCheckstyleMessage(result),
			Source:   "kyn." + result.RuleID,
		})
	}

	paths := make([]string, 0, len(grouped))
	for path := range grouped {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	files := make([]checkstyleFile, 0, len(paths))
	for _, path := range paths {
		files = append(files, checkstyleFile{
			Name:   path,
			Errors: grouped[path],
		})
	}

	out, err := xml.MarshalIndent(checkstyleReport{
		Version: "10.0",
		Files:   files,
	}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), out...), nil
}

func checkstyleSeverity(severity rules.Severity) string {
	switch severity {
	case rules.SeverityError:
		return "error"
	case rules.SeverityWarn:
		return "warning"
	default:
		return "info"
	}
}

func renderCheckstyleMessage(result rules.RuleResult) string {
	if len(result.ExpectedFiles) == 0 {
		return renderRDJSONMessage(result)
	}
	return renderRDJSONMessage(result) + " Expected: " + result.ExpectedFiles[0]
}

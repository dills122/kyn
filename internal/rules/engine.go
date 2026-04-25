package rules

import "sort"

func summarize(results []RuleResult, flags []string, failOn string) Summary {
	s := Summary{
		OK:      true,
		Results: results,
		Flags:   flags,
	}

	for _, r := range results {
		switch r.Status {
		case StatusPass:
			s.Passed++
		case StatusFail:
			s.Failed++
		case StatusInfo:
			// informational status does not affect pass/fail counts
		}

		switch r.Severity {
		case SeverityInfo:
			s.Infos++
		case SeverityWarn:
			s.Warnings++
		case SeverityError:
			s.Errors++
		}
	}

	if failOn == "warn" {
		for _, r := range results {
			if r.Status != StatusFail {
				continue
			}
			if r.Severity == SeverityWarn || r.Severity == SeverityError {
				s.OK = false
				return s
			}
		}
	} else {
		for _, r := range results {
			if r.Status == StatusFail && r.Severity == SeverityError {
				s.OK = false
				return s
			}
		}
	}

	return s
}

func mapKeysSorted(in map[string]struct{}) []string {
	out := make([]string, 0, len(in))
	for k := range in {
		if k != "" {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func cloneSorted(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

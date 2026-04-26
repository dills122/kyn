package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"kyn/internal/changes"
	"kyn/internal/config"
	"kyn/internal/family"
)

type ExplainStatus string

const (
	ExplainStatusPass    ExplainStatus = "pass"
	ExplainStatusFail    ExplainStatus = "fail"
	ExplainStatusInfo    ExplainStatus = "info"
	ExplainStatusSkipped ExplainStatus = "skipped"
)

type ClauseResult string

const (
	ClauseResultPass ClauseResult = "pass"
	ClauseResultFail ClauseResult = "fail"
)

type ClauseTrace struct {
	Clause string       `json:"clause"`
	Result ClauseResult `json:"result"`
	Detail string       `json:"detail,omitempty"`
}

type ExplainResult struct {
	RuleID        string        `json:"ruleId"`
	FamilyID      string        `json:"familyId"`
	FamilyName    string        `json:"familyName"`
	Severity      Severity      `json:"severity"`
	Status        ExplainStatus `json:"status"`
	Message       string        `json:"message"`
	ChangedFiles  []string      `json:"changedFiles,omitempty"`
	ExpectedFiles []string      `json:"expectedFiles,omitempty"`
	EmittedFlags  []string      `json:"emittedFlags,omitempty"`
	IfTrace       []ClauseTrace `json:"if,omitempty"`
	AssertTrace   []ClauseTrace `json:"assert,omitempty"`
}

type ExplainSummary struct {
	OK       bool            `json:"ok"`
	Passed   int             `json:"passed"`
	Failed   int             `json:"failed"`
	Infos    int             `json:"infos"`
	Skipped  int             `json:"skipped"`
	Warnings int             `json:"warnings"`
	Errors   int             `json:"errors"`
	Results  []ExplainResult `json:"results"`
	Flags    []string        `json:"flags,omitempty"`
}

func Explain(in EvalInput) (ExplainSummary, error) {
	results := make([]ExplainResult, 0)
	flagSet := map[string]struct{}{}

	for _, rule := range in.Rules {
		for _, inst := range in.Instances {
			if inst.FamilyID != rule.Family {
				continue
			}

			ifClauses := rule.IfClauses()
			assertClauses := rule.AssertClauses()

			result := ExplainResult{
				RuleID:       rule.ID,
				FamilyID:     inst.FamilyID,
				FamilyName:   inst.Name,
				Severity:     Severity(rule.Severity),
				Message:      rule.Message,
				ChangedFiles: cloneSorted(inst.SourceFiles),
			}

			whenOK, ifTrace, err := evalWhenTrace(in.Cwd, ifClauses, inst, in.StatusByFile)
			if err != nil {
				return ExplainSummary{}, err
			}
			result.IfTrace = ifTrace
			if !whenOK {
				result.Status = ExplainStatusSkipped
				results = append(results, result)
				continue
			}

			expectedSet := map[string]struct{}{}
			failed, emitted, assertTrace, err := evalRequireTrace(
				in.Cwd,
				in.Changed,
				assertClauses,
				rule.EmitFlags(),
				inst,
				expectedSet,
			)
			if err != nil {
				return ExplainSummary{}, err
			}

			result.AssertTrace = assertTrace
			result.EmittedFlags = emitted
			for _, flag := range emitted {
				if flag != "" {
					flagSet[flag] = struct{}{}
				}
			}

			if len(expectedSet) > 0 {
				result.ExpectedFiles = mapKeysSorted(expectedSet)
			}

			if failed {
				result.Status = ExplainStatusFail
			} else if hasRequireChecks(assertClauses) {
				result.Status = ExplainStatusPass
			} else {
				result.Status = ExplainStatusInfo
			}

			results = append(results, result)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		ri := explainStatusRank(results[i].Status)
		rj := explainStatusRank(results[j].Status)
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

	applied := make([]RuleResult, 0, len(results))
	skipped := 0
	for _, result := range results {
		switch result.Status {
		case ExplainStatusSkipped:
			skipped++
		case ExplainStatusFail:
			applied = append(applied, toRuleResult(result, StatusFail))
		case ExplainStatusPass:
			applied = append(applied, toRuleResult(result, StatusPass))
		case ExplainStatusInfo:
			applied = append(applied, toRuleResult(result, StatusInfo))
		}
	}
	flags := mapKeysSorted(flagSet)
	base := summarize(applied, flags, in.FailOn)

	summary := ExplainSummary{
		OK:       base.OK,
		Passed:   base.Passed,
		Failed:   base.Failed,
		Infos:    base.Infos,
		Skipped:  skipped,
		Warnings: base.Warnings,
		Errors:   base.Errors,
		Results:  results,
		Flags:    flags,
	}

	if in.FailOnEmpty && len(in.Instances) == 0 {
		summary.OK = false
		summary.Failed++
		summary.Errors++
		summary.Results = append(summary.Results, ExplainResult{
			RuleID:   "fail-on-empty",
			Severity: SeverityError,
			Status:   ExplainStatusFail,
			Message:  "No family instances matched and --fail-on-empty is set.",
		})
	}

	return summary, nil
}

func evalWhenTrace(
	cwd string,
	when config.RuleClauses,
	inst family.Instance,
	statusByFile map[string]changes.Status,
) (bool, []ClauseTrace, error) {
	traces := make([]ClauseTrace, 0, 4)

	if len(when.ChangedAny) > 0 {
		if len(inst.SourceFiles) == 0 {
			traces = append(traces, ClauseTrace{
				Clause: "if.changedAny",
				Result: ClauseResultFail,
				Detail: "no source files changed in this instance",
			})
			return false, traces, nil
		}
		traces = append(traces, ClauseTrace{
			Clause: "if.changedAny",
			Result: ClauseResultPass,
			Detail: fmt.Sprintf("%d source files changed", len(inst.SourceFiles)),
		})
	}

	if len(when.ChangedStatusAny) > 0 {
		if !matchesAnyChangedStatus(inst.SourceFiles, statusByFile, when.ChangedStatusAny) {
			traces = append(traces, ClauseTrace{
				Clause: "if.changedStatusAny",
				Result: ClauseResultFail,
				Detail: fmt.Sprintf("no source file status matched allowed set [%s]", strings.Join(when.ChangedStatusAny, ", ")),
			})
			return false, traces, nil
		}
		traces = append(traces, ClauseTrace{
			Clause: "if.changedStatusAny",
			Result: ClauseResultPass,
			Detail: fmt.Sprintf("matched one of [%s]", strings.Join(when.ChangedStatusAny, ", ")),
		})
	}

	for _, name := range when.KinExists {
		ok, err := explainKinExistence(cwd, inst, name, true)
		if err != nil {
			return false, traces, err
		}
		if !ok {
			traces = append(traces, ClauseTrace{
				Clause: "if.kinExists",
				Result: ClauseResultFail,
				Detail: fmt.Sprintf("%s missing (%s)", name, inst.Kin[name]),
			})
			return false, traces, nil
		}
		traces = append(traces, ClauseTrace{
			Clause: "if.kinExists",
			Result: ClauseResultPass,
			Detail: fmt.Sprintf("%s exists (%s)", name, inst.Kin[name]),
		})
	}

	for _, name := range when.KinMissing {
		ok, err := explainKinExistence(cwd, inst, name, false)
		if err != nil {
			return false, traces, err
		}
		if !ok {
			traces = append(traces, ClauseTrace{
				Clause: "if.kinMissing",
				Result: ClauseResultFail,
				Detail: fmt.Sprintf("%s exists but must be missing (%s)", name, inst.Kin[name]),
			})
			return false, traces, nil
		}
		traces = append(traces, ClauseTrace{
			Clause: "if.kinMissing",
			Result: ClauseResultPass,
			Detail: fmt.Sprintf("%s missing (%s)", name, inst.Kin[name]),
		})
	}

	return true, traces, nil
}

func evalRequireTrace(
	cwd string,
	changed map[string]struct{},
	req config.RuleClauses,
	emitFlags []string,
	inst family.Instance,
	expected map[string]struct{},
) (bool, []string, []ClauseTrace, error) {
	traces := make([]ClauseTrace, 0, 4)
	failed := false

	for _, name := range req.KinChanged {
		p := inst.Kin[name]
		if _, ok := changed[p]; !ok {
			expected[p] = struct{}{}
			failed = true
			traces = append(traces, ClauseTrace{
				Clause: "assert.kinChanged",
				Result: ClauseResultFail,
				Detail: fmt.Sprintf("%s was not changed (%s)", name, p),
			})
			continue
		}
		traces = append(traces, ClauseTrace{
			Clause: "assert.kinChanged",
			Result: ClauseResultPass,
			Detail: fmt.Sprintf("%s changed (%s)", name, p),
		})
	}

	for _, name := range req.KinUnchanged {
		p := inst.Kin[name]
		if _, ok := changed[p]; ok {
			expected[p] = struct{}{}
			failed = true
			traces = append(traces, ClauseTrace{
				Clause: "assert.kinUnchanged",
				Result: ClauseResultFail,
				Detail: fmt.Sprintf("%s changed but must stay unchanged (%s)", name, p),
			})
			continue
		}
		traces = append(traces, ClauseTrace{
			Clause: "assert.kinUnchanged",
			Result: ClauseResultPass,
			Detail: fmt.Sprintf("%s unchanged (%s)", name, p),
		})
	}

	for _, name := range req.KinExists {
		ok, err := explainKinExistence(cwd, inst, name, true)
		if err != nil {
			return false, nil, nil, err
		}
		if !ok {
			expected[inst.Kin[name]] = struct{}{}
			failed = true
			traces = append(traces, ClauseTrace{
				Clause: "assert.kinExists",
				Result: ClauseResultFail,
				Detail: fmt.Sprintf("%s missing (%s)", name, inst.Kin[name]),
			})
			continue
		}
		traces = append(traces, ClauseTrace{
			Clause: "assert.kinExists",
			Result: ClauseResultPass,
			Detail: fmt.Sprintf("%s exists (%s)", name, inst.Kin[name]),
		})
	}

	for _, name := range req.KinMissing {
		ok, err := explainKinExistence(cwd, inst, name, false)
		if err != nil {
			return false, nil, nil, err
		}
		if !ok {
			expected[inst.Kin[name]] = struct{}{}
			failed = true
			traces = append(traces, ClauseTrace{
				Clause: "assert.kinMissing",
				Result: ClauseResultFail,
				Detail: fmt.Sprintf("%s exists but must be missing (%s)", name, inst.Kin[name]),
			})
			continue
		}
		traces = append(traces, ClauseTrace{
			Clause: "assert.kinMissing",
			Result: ClauseResultPass,
			Detail: fmt.Sprintf("%s missing (%s)", name, inst.Kin[name]),
		})
	}

	for _, flag := range emitFlags {
		if flag == "" {
			continue
		}
		traces = append(traces, ClauseTrace{
			Clause: "actions.emit",
			Result: ClauseResultPass,
			Detail: fmt.Sprintf("emit flag %s", flag),
		})
	}

	return failed, emitFlags, traces, nil
}

func explainKinExistence(cwd string, inst family.Instance, kinName string, shouldExist bool) (bool, error) {
	p := inst.Kin[kinName]
	abs := filepath.Join(cwd, filepath.FromSlash(p))
	_, err := os.Stat(abs)
	exists := err == nil
	if !exists && err != nil && !os.IsNotExist(err) {
		return false, err
	}
	if shouldExist {
		return exists, nil
	}
	return !exists, nil
}

func explainStatusRank(status ExplainStatus) int {
	switch status {
	case ExplainStatusFail:
		return 0
	case ExplainStatusInfo:
		return 1
	case ExplainStatusPass:
		return 2
	case ExplainStatusSkipped:
		return 3
	default:
		return 4
	}
}

func toRuleResult(result ExplainResult, status ResultStatus) RuleResult {
	return RuleResult{
		RuleID:        result.RuleID,
		FamilyID:      result.FamilyID,
		FamilyName:    result.FamilyName,
		Severity:      result.Severity,
		Status:        status,
		Message:       result.Message,
		ChangedFiles:  result.ChangedFiles,
		ExpectedFiles: result.ExpectedFiles,
	}
}

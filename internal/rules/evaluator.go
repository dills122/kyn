package rules

import (
	"os"
	"path/filepath"
	"sort"

	"kyn/internal/changes"
	"kyn/internal/config"
	"kyn/internal/family"
)

type EvalInput struct {
	Cwd          string
	FailOn       string
	FailOnEmpty  bool
	Changed      map[string]struct{}
	StatusByFile map[string]changes.Status
	Rules        []config.Rule
	Instances    []family.Instance
}

func Evaluate(in EvalInput) (Summary, error) {
	results := make([]RuleResult, 0)
	flagSet := map[string]struct{}{}

	for _, rule := range in.Rules {
		for _, inst := range in.Instances {
			if inst.FamilyID != rule.Family {
				continue
			}

			ifClauses := rule.IfClauses()
			assertClauses := rule.AssertClauses()

			whenOK, err := evalWhen(in.Cwd, ifClauses, inst, in.StatusByFile)
			if err != nil {
				return Summary{}, err
			}
			if !whenOK {
				continue
			}

			expectedSet := map[string]struct{}{}
			failed, emitted, err := evalRequire(in.Cwd, in.Changed, assertClauses, rule.EmitFlags(), inst, expectedSet)
			if err != nil {
				return Summary{}, err
			}
			for _, e := range emitted {
				if e != "" {
					flagSet[e] = struct{}{}
				}
			}

			result := RuleResult{
				RuleID:       rule.ID,
				FamilyID:     inst.FamilyID,
				FamilyName:   inst.Name,
				Severity:     Severity(rule.Severity),
				Message:      rule.Message,
				ChangedFiles: cloneSorted(inst.SourceFiles),
			}

			if len(expectedSet) > 0 {
				result.ExpectedFiles = mapKeysSorted(expectedSet)
			}

			if failed {
				result.Status = StatusFail
			} else if hasRequireChecks(assertClauses) {
				result.Status = StatusPass
			} else {
				result.Status = StatusInfo
			}

			results = append(results, result)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].FamilyID == results[j].FamilyID {
			if results[i].FamilyName == results[j].FamilyName {
				return results[i].RuleID < results[j].RuleID
			}
			return results[i].FamilyName < results[j].FamilyName
		}
		return results[i].FamilyID < results[j].FamilyID
	})

	flags := mapKeysSorted(flagSet)
	summary := summarize(results, flags, in.FailOn)

	if in.FailOnEmpty && len(in.Instances) == 0 {
		summary.OK = false
		summary.Failed++
		summary.Errors++
		summary.Results = append(summary.Results, RuleResult{
			RuleID:   "fail-on-empty",
			Severity: SeverityError,
			Status:   StatusFail,
			Message:  "No family instances matched and --fail-on-empty is set.",
		})
	}

	return summary, nil
}

func evalWhen(cwd string, when config.RuleClauses, inst family.Instance, statusByFile map[string]changes.Status) (bool, error) {
	if len(when.ChangedAny) > 0 {
		if len(inst.SourceFiles) == 0 {
			return false, nil
		}
	}
	if len(when.ChangedStatusAny) > 0 {
		if !matchesAnyChangedStatus(inst.SourceFiles, statusByFile, when.ChangedStatusAny) {
			return false, nil
		}
	}
	if len(when.KinExists) > 0 {
		ok, err := kinExistence(cwd, inst, when.KinExists, true)
		if err != nil || !ok {
			return ok, err
		}
	}
	if len(when.KinMissing) > 0 {
		ok, err := kinExistence(cwd, inst, when.KinMissing, false)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func matchesAnyChangedStatus(sourceFiles []string, statusByFile map[string]changes.Status, allowed []string) bool {
	if len(sourceFiles) == 0 {
		return false
	}
	allowedSet := make(map[changes.Status]struct{}, len(allowed))
	for _, a := range allowed {
		allowedSet[changes.Status(a)] = struct{}{}
	}
	for _, f := range sourceFiles {
		status, ok := statusByFile[f]
		if !ok {
			continue
		}
		if _, ok := allowedSet[status]; ok {
			return true
		}
	}
	return false
}

func evalRequire(cwd string, changed map[string]struct{}, req config.RuleClauses, emitFlags []string, inst family.Instance, expected map[string]struct{}) (bool, []string, error) {
	failed := false
	if len(req.KinChanged) > 0 {
		for _, name := range req.KinChanged {
			p := inst.Kin[name]
			if _, ok := changed[p]; !ok {
				expected[p] = struct{}{}
				failed = true
			}
		}
	}
	if len(req.KinUnchanged) > 0 {
		for _, name := range req.KinUnchanged {
			p := inst.Kin[name]
			if _, ok := changed[p]; ok {
				expected[p] = struct{}{}
				failed = true
			}
		}
	}
	if len(req.KinExists) > 0 {
		ok, err := kinExistence(cwd, inst, req.KinExists, true)
		if err != nil {
			return false, nil, err
		}
		if !ok {
			for _, name := range req.KinExists {
				expected[inst.Kin[name]] = struct{}{}
			}
			failed = true
		}
	}
	if len(req.KinMissing) > 0 {
		ok, err := kinExistence(cwd, inst, req.KinMissing, false)
		if err != nil {
			return false, nil, err
		}
		if !ok {
			for _, name := range req.KinMissing {
				expected[inst.Kin[name]] = struct{}{}
			}
			failed = true
		}
	}

	return failed, emitFlags, nil
}

func kinExistence(cwd string, inst family.Instance, kinNames []string, shouldExist bool) (bool, error) {
	for _, name := range kinNames {
		p := inst.Kin[name]
		abs := filepath.Join(cwd, filepath.FromSlash(p))
		_, err := os.Stat(abs)
		exists := err == nil
		if !exists && err != nil && !os.IsNotExist(err) {
			return false, err
		}
		if shouldExist && !exists {
			return false, nil
		}
		if !shouldExist && exists {
			return false, nil
		}
	}
	return true, nil
}

func hasRequireChecks(req config.RuleClauses) bool {
	return len(req.KinChanged) > 0 ||
		len(req.KinUnchanged) > 0 ||
		len(req.KinExists) > 0 ||
		len(req.KinMissing) > 0
}

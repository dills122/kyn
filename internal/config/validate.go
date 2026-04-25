package config

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

var kinTemplateVarRe = regexp.MustCompile(`\{([a-zA-Z]+)\}`)

var allowedTemplateVars = map[string]struct{}{
	"dir":  {},
	"file": {},
	"name": {},
	"base": {},
	"ext":  {},
}

var allowedSeverities = []string{"info", "warn", "error"}

func Validate(cfg Config) error {
	if cfg.Version != 1 {
		return fmt.Errorf("unsupported config version %d; expected version 1", cfg.Version)
	}

	familyIDs := make(map[string]struct{}, len(cfg.Families))
	familyByID := make(map[string]Family, len(cfg.Families))
	for i, fam := range cfg.Families {
		if strings.TrimSpace(fam.ID) == "" {
			return fmt.Errorf("families[%d].id is required", i)
		}
		if _, exists := familyIDs[fam.ID]; exists {
			return fmt.Errorf("duplicate family id: %q", fam.ID)
		}
		familyIDs[fam.ID] = struct{}{}
		familyByID[fam.ID] = fam

		if len(fam.Include) == 0 {
			return fmt.Errorf("family %q include must contain at least one glob pattern", fam.ID)
		}
		for j, pattern := range fam.Include {
			if strings.TrimSpace(pattern) == "" {
				return fmt.Errorf("family %q include[%d] must not be empty", fam.ID, j)
			}
		}

		if len(fam.Kin) == 0 {
			return fmt.Errorf("family %q must define at least one kin template", fam.ID)
		}
		for kinName, template := range fam.Kin {
			if strings.TrimSpace(kinName) == "" {
				return fmt.Errorf("family %q has empty kin name", fam.ID)
			}
			if strings.TrimSpace(template) == "" {
				return fmt.Errorf("family %q kin %q has empty template", fam.ID, kinName)
			}
			if err := validateTemplateVars(template); err != nil {
				return fmt.Errorf("family %q kin %q: %w", fam.ID, kinName, err)
			}
		}
	}

	ruleIDs := make(map[string]struct{}, len(cfg.Rules))
	for i, rule := range cfg.Rules {
		if strings.TrimSpace(rule.ID) == "" {
			return fmt.Errorf("rules[%d].id is required", i)
		}
		if _, exists := ruleIDs[rule.ID]; exists {
			return fmt.Errorf("duplicate rule id: %q", rule.ID)
		}
		ruleIDs[rule.ID] = struct{}{}

		if strings.TrimSpace(rule.Family) == "" {
			return fmt.Errorf("rule %q family is required", rule.ID)
		}
		if _, exists := familyIDs[rule.Family]; !exists {
			return fmt.Errorf("rule %q references unknown family id %q", rule.ID, rule.Family)
		}
		if !slices.Contains(allowedSeverities, rule.Severity) {
			return fmt.Errorf("rule %q has invalid severity %q", rule.ID, rule.Severity)
		}
		if strings.TrimSpace(rule.Message) == "" {
			return fmt.Errorf("rule %q message is required", rule.ID)
		}

		if err := validateChangedAny(rule.ID, "when", rule.When.ChangedAny); err != nil {
			return err
		}
		if err := validateChangedAny(rule.ID, "require", rule.Require.ChangedAny); err != nil {
			return err
		}
		fam := familyByID[rule.Family]
		if err := validateKinRefs(rule.ID, "when.kinExists", rule.When.KinExists, fam.Kin); err != nil {
			return err
		}
		if err := validateKinRefs(rule.ID, "when.kinMissing", rule.When.KinMissing, fam.Kin); err != nil {
			return err
		}
		if err := validateKinRefs(rule.ID, "require.kinChanged", rule.Require.KinChanged, fam.Kin); err != nil {
			return err
		}
		if err := validateKinRefs(rule.ID, "require.kinUnchanged", rule.Require.KinUnchanged, fam.Kin); err != nil {
			return err
		}
		if err := validateKinRefs(rule.ID, "require.kinExists", rule.Require.KinExists, fam.Kin); err != nil {
			return err
		}
		if err := validateKinRefs(rule.ID, "require.kinMissing", rule.Require.KinMissing, fam.Kin); err != nil {
			return err
		}
	}

	return nil
}

func validateTemplateVars(template string) error {
	matches := kinTemplateVarRe.FindAllStringSubmatch(template, -1)
	for _, m := range matches {
		v := m[1]
		if _, ok := allowedTemplateVars[v]; !ok {
			return fmt.Errorf("unsupported template variable {%s}", v)
		}
	}
	return nil
}

func validateChangedAny(ruleID string, clause string, groups []string) error {
	for i, g := range groups {
		if g != "source" {
			return fmt.Errorf("rule %q %s.changedAny[%d] invalid group %q; only \"source\" is supported in MVP", ruleID, clause, i, g)
		}
	}
	return nil
}

func validateKinRefs(ruleID string, field string, kinRefs []string, kinMap KinMap) error {
	for i, name := range kinRefs {
		if _, ok := kinMap[name]; !ok {
			return fmt.Errorf("rule %q %s[%d] references unknown kin %q", ruleID, field, i, name)
		}
	}
	return nil
}

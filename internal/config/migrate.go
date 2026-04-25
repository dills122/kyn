package config

import (
	"fmt"
	"slices"
)

func MigrateV1ToV2(in Config) (Config, error) {
	if in.Version != 1 {
		return Config{}, fmt.Errorf("migration requires version 1 input, got version %d", in.Version)
	}

	out := Config{
		Version:  2,
		Families: make([]Family, 0, len(in.Families)),
		Rules:    make([]Rule, 0, len(in.Rules)),
	}

	for _, fam := range in.Families {
		migratedFamily := fam
		if len(migratedFamily.Groups) == 0 {
			migratedFamily.Groups = GroupMap{
				"source": {
					Include: slices.Clone(fam.Include),
					Exclude: slices.Clone(fam.Exclude),
				},
			}
		}
		// v2 source matching should be expressed via groups.
		migratedFamily.Include = nil
		migratedFamily.Exclude = nil
		out.Families = append(out.Families, migratedFamily)
	}

	for _, rule := range in.Rules {
		migratedRule := rule
		ifClauses := rule.IfClauses()
		assertClauses := rule.AssertClauses()
		assertClauses.EmitFlag = ""
		emits := dedupeStrings(rule.EmitFlags())

		migratedRule.When = RuleClauses{}
		migratedRule.Require = RuleClauses{}
		migratedRule.If = ifClauses
		migratedRule.Assert = assertClauses
		if len(emits) > 0 {
			migratedRule.Actions = RuleActions{Emit: emits}
		} else {
			migratedRule.Actions = RuleActions{}
		}

		out.Rules = append(out.Rules, migratedRule)
	}

	if err := Validate(out); err != nil {
		return Config{}, fmt.Errorf("migrated config failed validation: %w", err)
	}

	return out, nil
}

func dedupeStrings(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, value := range in {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

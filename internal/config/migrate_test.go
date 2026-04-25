package config

import "testing"

func TestMigrateV1ToV2(t *testing.T) {
	in := Config{
		Version: 1,
		Families: []Family{
			{
				ID:      "web-component",
				Include: []string{"src/**/*.component.ts"},
				Exclude: []string{"**/*.generated.ts"},
				Kin: KinMap{
					"story": "{dir}/{base}.stories.ts",
				},
			},
		},
		Rules: []Rule{
			{
				ID:       "story-sync",
				Family:   "web-component",
				Severity: "error",
				When: RuleClauses{
					ChangedAny: []string{"source"},
				},
				Require: RuleClauses{
					KinChanged: []string{"story"},
					EmitFlag:   "figmaPublishRequired",
				},
				Message: "Sync story.",
			},
		},
	}

	out, err := MigrateV1ToV2(in)
	if err != nil {
		t.Fatalf("MigrateV1ToV2 returned error: %v", err)
	}
	if out.Version != 2 {
		t.Fatalf("version=%d want 2", out.Version)
	}
	if len(out.Families) != 1 {
		t.Fatalf("families=%d want 1", len(out.Families))
	}
	sourceGroup := out.Families[0].Groups["source"]
	if len(sourceGroup.Include) != 1 || sourceGroup.Include[0] != "src/**/*.component.ts" {
		t.Fatalf("unexpected source include: %v", sourceGroup.Include)
	}
	if len(sourceGroup.Exclude) != 1 || sourceGroup.Exclude[0] != "**/*.generated.ts" {
		t.Fatalf("unexpected source exclude: %v", sourceGroup.Exclude)
	}
	if len(out.Families[0].Include) != 0 || len(out.Families[0].Exclude) != 0 {
		t.Fatalf("expected include/exclude to be cleared in v2 family")
	}

	if len(out.Rules) != 1 {
		t.Fatalf("rules=%d want 1", len(out.Rules))
	}
	rule := out.Rules[0]
	if len(rule.When.ChangedAny) != 0 || len(rule.Require.KinChanged) != 0 || rule.Require.EmitFlag != "" {
		t.Fatalf("expected v1 when/require fields cleared, got when=%+v require=%+v", rule.When, rule.Require)
	}
	if len(rule.If.ChangedAny) != 1 || rule.If.ChangedAny[0] != "source" {
		t.Fatalf("unexpected if clauses: %+v", rule.If)
	}
	if len(rule.Assert.KinChanged) != 1 || rule.Assert.KinChanged[0] != "story" {
		t.Fatalf("unexpected assert clauses: %+v", rule.Assert)
	}
	if len(rule.Actions.Emit) != 1 || rule.Actions.Emit[0] != "figmaPublishRequired" {
		t.Fatalf("unexpected actions emit: %+v", rule.Actions.Emit)
	}
}

func TestMigrateV1ToV2RejectsNonV1(t *testing.T) {
	_, err := MigrateV1ToV2(Config{Version: 2})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

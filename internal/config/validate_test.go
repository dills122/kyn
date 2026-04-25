package config

import "testing"

func TestValidate(t *testing.T) {
	valid := Config{
		Version: 1,
		Families: []Family{
			{
				ID:      "angular-component",
				Include: []string{"libs/**/*.component.ts"},
				Kin: KinMap{
					"story": "{dir}/{base}.stories.ts",
				},
			},
		},
		Rules: []Rule{
			{
				ID:       "storybook-sync",
				Family:   "angular-component",
				Severity: "error",
				When: RuleClauses{
					ChangedAny: []string{"source"},
				},
				Require: RuleClauses{
					KinChanged: []string{"story"},
				},
				Message: "Component changed but story did not.",
			},
		},
	}

	tests := []struct {
		name    string
		modify  func(cfg *Config)
		wantErr bool
	}{
		{name: "valid"},
		{
			name: "invalid version",
			modify: func(cfg *Config) {
				cfg.Version = 3
			},
			wantErr: true,
		},
		{
			name: "duplicate family id",
			modify: func(cfg *Config) {
				cfg.Families = append(cfg.Families, cfg.Families[0])
			},
			wantErr: true,
		},
		{
			name: "missing family reference",
			modify: func(cfg *Config) {
				cfg.Rules[0].Family = "missing"
			},
			wantErr: true,
		},
		{
			name: "invalid changedAny group",
			modify: func(cfg *Config) {
				cfg.Rules[0].When.ChangedAny = []string{"docs"}
			},
			wantErr: true,
		},
		{
			name: "invalid kin reference",
			modify: func(cfg *Config) {
				cfg.Rules[0].Require.KinChanged = []string{"missing-kin"}
			},
			wantErr: true,
		},
		{
			name: "invalid template variable",
			modify: func(cfg *Config) {
				cfg.Families[0].Kin["story"] = "{dir}/{wat}.stories.ts"
			},
			wantErr: true,
		},
		{
			name: "invalid severity",
			modify: func(cfg *Config) {
				cfg.Rules[0].Severity = "fatal"
			},
			wantErr: true,
		},
		{
			name: "empty include pattern",
			modify: func(cfg *Config) {
				cfg.Families[0].Include = []string{"", "libs/**/*.component.ts"}
			},
			wantErr: true,
		},
		{
			name: "missing kin template",
			modify: func(cfg *Config) {
				cfg.Families[0].Kin = KinMap{}
			},
			wantErr: true,
		},
		{
			name: "valid v2 if/assert/actions with groups",
			modify: func(cfg *Config) {
				cfg.Version = 2
				cfg.Families[0].Include = nil
				cfg.Families[0].Groups = GroupMap{
					"source": {
						Include: []string{"libs/**/*.component.ts"},
					},
				}
				cfg.Rules[0].When = RuleClauses{}
				cfg.Rules[0].Require = RuleClauses{}
				cfg.Rules[0].If = RuleClauses{ChangedAny: []string{"source"}}
				cfg.Rules[0].Assert = RuleClauses{KinChanged: []string{"story"}}
				cfg.Rules[0].Actions = RuleActions{Emit: []string{"storybook-sync-required"}}
			},
		},
		{
			name: "invalid changed status",
			modify: func(cfg *Config) {
				cfg.Version = 2
				cfg.Families[0].Groups = GroupMap{
					"source": {Include: []string{"libs/**/*.component.ts"}},
				}
				cfg.Families[0].Include = nil
				cfg.Rules[0].If = RuleClauses{
					ChangedAny:       []string{"source"},
					ChangedStatusAny: []string{"moved"},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := cloneConfig(valid)
			if tt.modify != nil {
				tt.modify(&cfg)
			}

			err := Validate(cfg)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func cloneConfig(in Config) Config {
	out := in
	out.Families = append([]Family(nil), in.Families...)
	for i := range out.Families {
		out.Families[i].Include = append([]string(nil), in.Families[i].Include...)
		out.Families[i].Exclude = append([]string(nil), in.Families[i].Exclude...)
		out.Families[i].BaseName.StripSuffixes = append([]string(nil), in.Families[i].BaseName.StripSuffixes...)
		out.Families[i].Kin = make(KinMap, len(in.Families[i].Kin))
		for k, v := range in.Families[i].Kin {
			out.Families[i].Kin[k] = v
		}
		out.Families[i].Groups = make(GroupMap, len(in.Families[i].Groups))
		for k, v := range in.Families[i].Groups {
			out.Families[i].Groups[k] = GroupDef{
				Include: append([]string(nil), v.Include...),
				Exclude: append([]string(nil), v.Exclude...),
			}
		}
	}
	out.Rules = append([]Rule(nil), in.Rules...)
	for i := range out.Rules {
		out.Rules[i].When = cloneClauses(in.Rules[i].When)
		out.Rules[i].Require = cloneClauses(in.Rules[i].Require)
		out.Rules[i].If = cloneClauses(in.Rules[i].If)
		out.Rules[i].Assert = cloneClauses(in.Rules[i].Assert)
		out.Rules[i].Actions.Emit = append([]string(nil), in.Rules[i].Actions.Emit...)
	}
	return out
}

func cloneClauses(in RuleClauses) RuleClauses {
	return RuleClauses{
		ChangedAny:       append([]string(nil), in.ChangedAny...),
		ChangedStatusAny: append([]string(nil), in.ChangedStatusAny...),
		KinExists:        append([]string(nil), in.KinExists...),
		KinMissing:       append([]string(nil), in.KinMissing...),
		KinChanged:       append([]string(nil), in.KinChanged...),
		KinUnchanged:     append([]string(nil), in.KinUnchanged...),
		EmitFlag:         in.EmitFlag,
	}
}

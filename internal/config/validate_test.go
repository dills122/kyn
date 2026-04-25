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
				cfg.Version = 2
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := valid
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

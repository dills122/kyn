package family

import (
	"slices"
	"testing"

	"kyn/internal/config"
)

func TestResolve(t *testing.T) {
	cfg := config.Config{
		Version: 1,
		Families: []config.Family{
			{
				ID: "angular-component",
				Include: []string{
					"libs/**/*.component.ts",
					"libs/**/*.component.html",
				},
				BaseName: config.BaseName{
					StripSuffixes: []string{".component"},
				},
				Kin: config.KinMap{
					"story": "{dir}/{base}.stories.ts",
					"spec":  "{dir}/{base}.spec.ts",
				},
			},
		},
	}

	changed := []string{
		"libs/ui/button/button.component.ts",
		"libs/ui/button/button.component.html",
		"libs/ui/button/button.component.ts",
		"libs/ui/card/card.component.ts",
		"libs/ui/button/button.spec.ts",
	}

	instances, err := Resolve(cfg, changed)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}

	button := instances[0]
	if button.FamilyID != "angular-component" {
		t.Fatalf("unexpected family id %q", button.FamilyID)
	}
	if button.Name != "libs/ui/button/button" {
		t.Fatalf("unexpected instance name %q", button.Name)
	}
	wantSources := []string{
		"libs/ui/button/button.component.html",
		"libs/ui/button/button.component.ts",
	}
	if !slices.Equal(button.SourceFiles, wantSources) {
		t.Fatalf("button source files = %v, want %v", button.SourceFiles, wantSources)
	}
	if button.Kin["story"] != "libs/ui/button/button.stories.ts" {
		t.Fatalf("unexpected button story kin: %q", button.Kin["story"])
	}
	if button.Kin["spec"] != "libs/ui/button/button.spec.ts" {
		t.Fatalf("unexpected button spec kin: %q", button.Kin["spec"])
	}
}

func TestResolveV2SourceGroup(t *testing.T) {
	cfg := config.Config{
		Version: 2,
		Families: []config.Family{
			{
				ID: "angular-component",
				Groups: config.GroupMap{
					"source": {
						Include: []string{
							"libs/**/*.component.ts",
						},
					},
				},
				BaseName: config.BaseName{
					StripSuffixes: []string{".component"},
				},
				Kin: config.KinMap{
					"story": "{dir}/{base}.stories.ts",
				},
			},
		},
	}

	changed := []string{
		"libs/ui/button/button.component.ts",
		"libs/ui/button/button.stories.ts",
	}

	instances, err := Resolve(cfg, changed)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(instances))
	}
	if instances[0].Name != "libs/ui/button/button" {
		t.Fatalf("unexpected instance name %q", instances[0].Name)
	}
}

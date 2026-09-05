package family

import (
	"slices"
	"strings"
	"testing"

	"github.com/dills122/kyn/internal/config"
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

func TestResolveRejectsUnsafePaths(t *testing.T) {
	base := config.Config{
		Version: 1,
		Families: []config.Family{{
			ID:      "go-source",
			Include: []string{"**/*.go"},
			Kin:     config.KinMap{"test": "{dir}/{base}_test.go"},
		}},
	}

	t.Run("changed path", func(t *testing.T) {
		_, err := Resolve(base, []string{"../outside.go"})
		if err == nil || !strings.Contains(err.Error(), "repository-relative") {
			t.Fatalf("Resolve() error = %v, want repository-relative path error", err)
		}
	})

	t.Run("resolved kin path", func(t *testing.T) {
		cfg := base
		cfg.Families = append([]config.Family(nil), base.Families...)
		cfg.Families[0].Kin = config.KinMap{"test": "../../{base}_test.go"}
		_, err := Resolve(cfg, []string{"src/button.go"})
		if err == nil || !strings.Contains(err.Error(), "resolved unsafe path") {
			t.Fatalf("Resolve() error = %v, want unsafe kin path error", err)
		}
	})

	t.Run("multiple invalid kin paths are reported deterministically", func(t *testing.T) {
		cfg := base
		cfg.Families = append([]config.Family(nil), base.Families...)
		cfg.Families[0].Kin = config.KinMap{
			"zeta":  "../../zeta/{base}_test.go",
			"alpha": "../../alpha/{base}_test.go",
		}
		for i := 0; i < 100; i++ {
			_, err := Resolve(cfg, []string{"src/button.go"})
			if err == nil || !strings.Contains(err.Error(), `kin "alpha"`) {
				t.Fatalf("Resolve() error = %v, want lexicographically first invalid kin", err)
			}
		}
	})
}

// A kin template that embeds {ext} (or {file}/{name}) resolves differently
// depending on which source file's extension it was built from. When two
// source files with different extensions fall into the same family
// instance, Resolve must fail loudly instead of silently keeping whichever
// file the sorted changed-file list happened to visit first.
func TestResolveRejectsAmbiguousKinAcrossExtensions(t *testing.T) {
	cfg := config.Config{
		Version: 1,
		Families: []config.Family{{
			ID: "web-component",
			Include: []string{
				"libs/**/*.component.ts",
				"libs/**/*.component.html",
			},
			BaseName: config.BaseName{StripSuffixes: []string{".component"}},
			Kin: config.KinMap{
				"generated": "{dir}/{base}{ext}.g.go",
			},
		}},
	}

	changed := []string{
		"libs/ui/button/button.component.html",
		"libs/ui/button/button.component.ts",
	}

	_, err := Resolve(cfg, changed)
	if err == nil {
		t.Fatal("Resolve() error = nil, want an ambiguous-kin error")
	}
	for _, want := range []string{
		`family "web-component"`,
		`instance "libs/ui/button/button"`,
		`kin "generated"`,
		"resolves to different paths",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("Resolve() error = %q, want it to contain %q", err.Error(), want)
		}
	}
}

// Instances spanning multiple source-file extensions are fine as long as no
// kin template's resolved path actually depends on the extension.
func TestResolveAllowsMultiExtensionInstanceWhenKinIsExtensionAgnostic(t *testing.T) {
	cfg := config.Config{
		Version: 1,
		Families: []config.Family{{
			ID: "web-component",
			Include: []string{
				"libs/**/*.component.ts",
				"libs/**/*.component.html",
			},
			BaseName: config.BaseName{StripSuffixes: []string{".component"}},
			Kin: config.KinMap{
				"story": "{dir}/{base}.stories.ts",
			},
		}},
	}

	changed := []string{
		"libs/ui/button/button.component.html",
		"libs/ui/button/button.component.ts",
	}

	instances, err := Resolve(cfg, changed)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(instances))
	}
	if got := instances[0].Kin["story"]; got != "libs/ui/button/button.stories.ts" {
		t.Fatalf("kin[story] = %q, want libs/ui/button/button.stories.ts", got)
	}
}

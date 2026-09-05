package e2e

import (
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGoReleaserPreservesContainerTagContracts(t *testing.T) {
	type dockerImage struct {
		ID        string   `yaml:"id"`
		IDs       []string `yaml:"ids"`
		Images    []string `yaml:"images"`
		Tags      []string `yaml:"tags"`
		Platforms []string `yaml:"platforms"`
		SBOM      *bool    `yaml:"sbom"`
		Flags     []string `yaml:"flags"`
	}
	type releaseConfig struct {
		DockersV2 []dockerImage `yaml:"dockers_v2"`
	}

	contents, err := os.ReadFile("../.goreleaser.yml")
	if err != nil {
		t.Fatalf("read GoReleaser config: %v", err)
	}
	var config releaseConfig
	if err := yaml.Unmarshal(contents, &config); err != nil {
		t.Fatalf("parse GoReleaser config: %v", err)
	}

	want := map[string]dockerImage{
		"kyn": {
			IDs:       []string{"kyn"},
			Images:    []string{"ghcr.io/{{ .Env.GITHUB_REPOSITORY_OWNER }}/kyn"},
			Tags:      []string{"{{ .Version }}", "latest"},
			Platforms: []string{"linux/amd64", "linux/arm64"},
		},
		"kyn-amd64-aliases": {
			IDs:       []string{"kyn"},
			Images:    []string{"ghcr.io/{{ .Env.GITHUB_REPOSITORY_OWNER }}/kyn"},
			Tags:      []string{"{{ .Version }}-amd64", "latest-amd64"},
			Platforms: []string{"linux/amd64"},
			Flags:     []string{"--provenance=false"},
		},
		"kyn-arm64-aliases": {
			IDs:       []string{"kyn"},
			Images:    []string{"ghcr.io/{{ .Env.GITHUB_REPOSITORY_OWNER }}/kyn"},
			Tags:      []string{"{{ .Version }}-arm64", "latest-arm64"},
			Platforms: []string{"linux/arm64"},
			Flags:     []string{"--provenance=false"},
		},
	}

	seen := make(map[string]bool, len(config.DockersV2))
	for _, got := range config.DockersV2 {
		expected, ok := want[got.ID]
		if !ok {
			t.Fatalf("unexpected dockers_v2 id %q", got.ID)
		}
		if seen[got.ID] {
			t.Fatalf("duplicate dockers_v2 id %q", got.ID)
		}
		seen[got.ID] = true
		if !reflect.DeepEqual(got.IDs, expected.IDs) {
			t.Errorf("%s build ids = %v, want %v", got.ID, got.IDs, expected.IDs)
		}
		if !reflect.DeepEqual(got.Images, expected.Images) {
			t.Errorf("%s images = %v, want %v", got.ID, got.Images, expected.Images)
		}
		if !reflect.DeepEqual(got.Tags, expected.Tags) {
			t.Errorf("%s tags = %v, want %v", got.ID, got.Tags, expected.Tags)
		}
		if !reflect.DeepEqual(got.Platforms, expected.Platforms) {
			t.Errorf("%s platforms = %v, want %v", got.ID, got.Platforms, expected.Platforms)
		}
		if !reflect.DeepEqual(got.Flags, expected.Flags) {
			t.Errorf("%s flags = %v, want %v", got.ID, got.Flags, expected.Flags)
		}
		if got.SBOM == nil || *got.SBOM {
			t.Errorf("%s must explicitly disable SBOMs", got.ID)
		}
	}
	for id := range want {
		if !seen[id] {
			t.Errorf("missing required dockers_v2 id %q", id)
		}
	}
}

func TestReleaseWorkflowUsesStableGoToolchain(t *testing.T) {
	type workflowStep struct {
		Name string            `yaml:"name"`
		Uses string            `yaml:"uses"`
		With map[string]string `yaml:"with"`
	}
	type workflowJob struct {
		Steps []workflowStep `yaml:"steps"`
	}
	type workflowConfig struct {
		Jobs map[string]workflowJob `yaml:"jobs"`
	}

	contents, err := os.ReadFile("../.github/workflows/release.yml")
	if err != nil {
		t.Fatalf("read release workflow: %v", err)
	}
	var config workflowConfig
	if err := yaml.Unmarshal(contents, &config); err != nil {
		t.Fatalf("parse release workflow: %v", err)
	}

	job, ok := config.Jobs["goreleaser"]
	if !ok {
		t.Fatal("release workflow is missing goreleaser job")
	}
	for _, step := range job.Steps {
		if step.Uses != "actions/setup-go@v5" {
			continue
		}
		if got := step.With["go-version"]; got != "1.27.1" {
			t.Fatalf("release Go version = %q, want 1.27.1", got)
		}
		if source := step.With["go-version-file"]; source != "" {
			t.Fatalf("release Go toolchain must not be derived from minimum-version file %q", source)
		}
		return
	}

	t.Fatal("release workflow is missing actions/setup-go@v5 step")
}

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

	if len(config.DockersV2) != len(want) {
		t.Fatalf("dockers_v2 entries = %d, want %d", len(config.DockersV2), len(want))
	}
	for _, got := range config.DockersV2 {
		expected, ok := want[got.ID]
		if !ok {
			t.Fatalf("unexpected dockers_v2 id %q", got.ID)
		}
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
		if got.ID != "kyn" && (got.SBOM == nil || *got.SBOM) {
			t.Errorf("%s must disable SBOMs for a single-platform image", got.ID)
		}
	}
}

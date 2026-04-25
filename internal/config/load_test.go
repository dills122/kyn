package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultDiscovery(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kyn.config.yaml")
	if err := os.WriteFile(path, []byte(testYAML), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, foundPath, err := Load(dir, "")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if foundPath != path {
		t.Fatalf("expected path %q, got %q", path, foundPath)
	}
	if cfg.Version != 1 {
		t.Fatalf("expected version 1, got %d", cfg.Version)
	}
}

func TestLoadExplicitNotFound(t *testing.T) {
	dir := t.TempDir()
	_, _, err := Load(dir, "missing.yaml")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrConfigNotFound) {
		t.Fatalf("expected ErrConfigNotFound, got %v", err)
	}
}

func TestLoadInvalidYAMLKnownField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kyn.config.yaml")
	bad := `version: 1
families:
  - id: angular-component
    include:
      - "libs/**/*.component.ts"
    kin:
      story: "{dir}/{base}.stories.ts"
    unknown: true
rules: []
`
	if err := os.WriteFile(path, []byte(bad), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, _, err := Load(dir, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

const testYAML = `version: 1
families:
  - id: angular-component
    include:
      - "libs/**/*.component.ts"
    kin:
      story: "{dir}/{base}.stories.ts"
rules:
  - id: storybook-sync
    family: angular-component
    severity: error
    when:
      changedAny: [source]
    require:
      kinChanged: [story]
    message: "sync story"
`

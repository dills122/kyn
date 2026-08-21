package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommandPrintsVersionString(t *testing.T) {
	root := newRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(out.String())
	want := versionString()
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestVersionFlagPrintsVersionString(t *testing.T) {
	root := newRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(out.String())
	want := versionString()
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestVersionStringDefaultsWhenUnset(t *testing.T) {
	prevVersion, prevCommit, prevDate := Version, Commit, Date
	t.Cleanup(func() { Version, Commit, Date = prevVersion, prevCommit, prevDate })
	Version, Commit, Date = "dev", "none", "unknown"

	got := versionString()
	want := "kyn dev (commit none, built unknown)"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

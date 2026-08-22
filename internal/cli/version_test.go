package cli

import (
	"bytes"
	"runtime/debug"
	"strings"
	"testing"
)

func TestEffectiveVersion(t *testing.T) {
	tests := []struct {
		name          string
		linkedVersion string
		buildInfo     *debug.BuildInfo
		buildInfoOK   bool
		want          string
	}{
		{
			name:          "prefers release linker metadata",
			linkedVersion: "0.1.2",
			buildInfo:     &debug.BuildInfo{Main: debug.Module{Version: "v9.9.9"}},
			buildInfoOK:   true,
			want:          "0.1.2",
		},
		{
			name:          "uses module version for go install binary",
			linkedVersion: "dev",
			buildInfo:     &debug.BuildInfo{Main: debug.Module{Version: "v0.1.2"}},
			buildInfoOK:   true,
			want:          "v0.1.2",
		},
		{
			name:          "keeps dev for local build",
			linkedVersion: "dev",
			buildInfo:     &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			buildInfoOK:   true,
			want:          "dev",
		},
		{
			name:          "keeps dev without build info",
			linkedVersion: "dev",
			buildInfoOK:   false,
			want:          "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := effectiveVersion(tt.linkedVersion, tt.buildInfo, tt.buildInfoOK); got != tt.want {
				t.Fatalf("effectiveVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

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

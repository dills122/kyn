package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestDedentExample(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "leading blank line and indented content, as written in command Example fields",
			in:   "\n  # comment\n  kyn check -c kyn.config.yaml\n",
			want: "  # comment\n  kyn check -c kyn.config.yaml",
		},
		{
			name: "no leading blank line",
			in:   "  # comment\n  kyn check\n",
			want: "  # comment\n  kyn check",
		},
		{
			name: "trailing blank lines trimmed",
			in:   "\n  kyn check\n\n\n",
			want: "  kyn check",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dedentExample(tt.in); got != tt.want {
				t.Fatalf("dedentExample(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestCommandExamplesAreUniformlyIndented guards against a regression where
// strings.TrimSpace() on an Example block strips the first content line's
// leading indent along with the blank line before it, leaving the first
// Examples line flush against the margin while every later line stays
// indented (see dedentExample's doc comment). Every non-blank line in every
// command's Example must be indented, matching how it's written in source.
func TestCommandExamplesAreUniformlyIndented(t *testing.T) {
	var walk func(cmd *cobra.Command)
	walk = func(cmd *cobra.Command) {
		if cmd.Example != "" {
			for _, line := range strings.Split(cmd.Example, "\n") {
				if strings.TrimSpace(line) == "" {
					continue
				}
				if !strings.HasPrefix(line, " ") {
					t.Errorf(
						"command %q: Example line is not indented: %q\nfull Example:\n%s",
						cmd.CommandPath(), line, cmd.Example,
					)
				}
			}
		}
		for _, sub := range cmd.Commands() {
			walk(sub)
		}
	}
	walk(newRootCommand())
}

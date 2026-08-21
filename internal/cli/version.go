package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// These are injected at build time via -ldflags, e.g.:
//
//	go build -ldflags "-X kyn/internal/cli.Version=v0.1.1 -X kyn/internal/cli.Commit=abc123 -X kyn/internal/cli.Date=2026-08-20"
//
// GoReleaser sets these automatically for release builds; local/dev builds
// fall back to the defaults below.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// versionString renders the single-line version string shared by
// `kyn version`, `kyn --version`, and any diagnostic output.
func versionString() string {
	return fmt.Sprintf("kyn %s (commit %s, built %s)", Version, Commit, Date)
}

func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the Kyn version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), versionString())
			return nil
		},
	}
	cmd.SilenceUsage = true
	return cmd
}

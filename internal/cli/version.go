package cli

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// These are injected at build time via -ldflags, e.g.:
//
//	go build -ldflags "-X github.com/dills122/kyn/internal/cli.Version=v0.1.1 -X github.com/dills122/kyn/internal/cli.Commit=abc123 -X github.com/dills122/kyn/internal/cli.Date=2026-08-20"
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
	info, ok := debug.ReadBuildInfo()
	return fmt.Sprintf("kyn %s (commit %s, built %s)", effectiveVersion(Version, info, ok), Commit, Date)
}

func effectiveVersion(linkedVersion string, info *debug.BuildInfo, ok bool) string {
	if linkedVersion != "dev" || !ok || info == nil {
		return linkedVersion
	}
	if info.Main.Version == "" || info.Main.Version == "(devel)" {
		return linkedVersion
	}
	return info.Main.Version
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

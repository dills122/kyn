package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

// dedentExample trims the leading blank line and trailing whitespace from a
// command's multi-line `Example:` raw string literal, without disturbing
// each line's own indentation the way strings.TrimSpace does.
//
// Every Example block in this package is written with every line — first
// line included — indented two spaces, matching a shell prompt. TrimSpace
// strips leading whitespace as one run, so it removes both the blank line
// after the opening backtick *and* the first content line's two-space
// indent, while every later line keeps its indent because it follows a
// literal "\n" that TrimSpace doesn't look past. The result is a `--help`
// screen where the first Examples line sits flush left and every other line
// is indented, e.g.:
//
//	Examples:
//	# Fastest happy path
//	  kyn check -c kyn.config.yaml
//
// dedentExample only removes the leading "\n" (not the spaces after it), so
// the first line's indent survives and lines up with the rest.
func dedentExample(s string) string {
	return strings.TrimRight(strings.TrimPrefix(s, "\n"), " \t\n")
}

// registerCommonInputFlags wires the changed-file input, git ref, config
// path, and working-directory flags that `kyn check` and `kyn explain`
// expose identically (same name, shorthand, default, and help text).
//
// Flags whose help text or presence differs between the two commands
// (--format, --fail-on, --fail-on-empty, --summary-only, and check-only
// --show-passes / --dry-run-resolve) stay registered per-command so each
// command's --help keeps its own accurate wording.
func registerCommonInputFlags(cmd *cobra.Command, opts *checkOptions) {
	cmd.Flags().StringVarP(&opts.ConfigPath, "config", "c", "", "Path to Kyn config file")
	cmd.Flags().StringVarP(&opts.FilesCSV, "files", "f", "", "Comma-separated changed files")
	cmd.Flags().StringVar(&opts.FilesFrom, "files-from", "", "Path to changed files list (one per line); use '-' for stdin")
	cmd.Flags().BoolVar(&opts.Stdin, "stdin", false, "Read changed files from stdin (alias for --files-from -)")
	cmd.Flags().StringVar(&opts.Base, "base", "", "Git base ref/SHA for diff detection")
	cmd.Flags().StringVar(&opts.Head, "head", "", "Git head ref/SHA for diff detection")
	cmd.Flags().BoolVar(&opts.StrictInput, "strict-input-mode", false, "Require an explicit single input mode; disable auto git mode")
	cmd.Flags().StringVar(&opts.Cwd, "cwd", ".", "Working directory")
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, "Enable diagnostic output")
}

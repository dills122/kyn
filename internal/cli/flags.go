package cli

import "github.com/spf13/cobra"

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

package cli

import (
	"strings"

	"github.com/dills122/kyn/internal/report"
	"github.com/dills122/kyn/internal/rules"

	"github.com/spf13/cobra"
)

func newExplainCommand() *cobra.Command {
	opts := checkOptions{}
	cmd := &cobra.Command{
		Use:   "explain",
		Short: "Show per-rule diagnostics for changed files and family instances",
		Long: strings.TrimSpace(`
Show per-rule diagnostics for changed files and family instances.

Happy path:
  kyn explain -c kyn.config.yaml

Core flags:
  -c, --config
  -o, --format
  --summary-only

Input behavior matches 'kyn check':
  - explicit mode: --files | --files-from | --stdin | --base + --head
  - auto mode: if none selected and --cwd is a git repo, use origin/main...HEAD

Advanced flags:
  --strict-input-mode
  --fail-on
  --fail-on-empty
  --verbose
  --cwd
`),
		Example: strings.TrimSpace(`
  # Fastest happy path
  kyn explain -c kyn.config.yaml

  # Explain with explicit git refs
  kyn explain -c kyn.config.yaml --base origin/main --head HEAD

  # Explain using stdin list
  git diff --name-only origin/main...HEAD | kyn explain -c kyn.config.yaml --stdin

  # Summary-only diagnostics in JSON
  kyn explain -c kyn.config.yaml --summary-only -o json
`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			run, err := prepareRun(opts, "explain", false)
			if err != nil {
				return err
			}

			summary, err := rules.Explain(run.evalInput())
			if err != nil {
				return runtimeError("explain evaluation failed: %v", err)
			}

			run.writeVerbose(cmd.ErrOrStderr())

			if run.opts.Format == "json" {
				var (
					out []byte
					err error
				)
				if run.opts.SummaryOnly {
					out, err = report.RenderExplainJSONSummary(summary)
				} else {
					out, err = report.RenderExplainJSON(summary)
				}
				if err != nil {
					return runtimeError("json render failed: %v", err)
				}
				_, _ = cmd.OutOrStdout().Write(out)
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
			} else {
				_, _ = cmd.OutOrStdout().Write([]byte(report.RenderExplainText(summary, run.opts.SummaryOnly)))
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
			}

			return nil
		},
	}
	cmd.SilenceUsage = true

	cmd.Flags().StringVarP(&opts.ConfigPath, "config", "c", "", "Path to Kyn config file")
	cmd.Flags().StringVarP(&opts.FilesCSV, "files", "f", "", "Comma-separated changed files")
	cmd.Flags().StringVar(&opts.FilesFrom, "files-from", "", "Path to changed files list (one per line); use '-' for stdin")
	cmd.Flags().BoolVar(&opts.Stdin, "stdin", false, "Read changed files from stdin (alias for --files-from -)")
	cmd.Flags().StringVar(&opts.Base, "base", "", "Git base ref/SHA for diff detection")
	cmd.Flags().StringVar(&opts.Head, "head", "", "Git head ref/SHA for diff detection")
	cmd.Flags().BoolVar(&opts.StrictInput, "strict-input-mode", false, "Require an explicit single input mode; disable auto git mode")
	cmd.Flags().StringVar(&opts.Cwd, "cwd", ".", "Working directory")
	cmd.Flags().StringVarP(&opts.Format, "format", "o", "text", "Output format: text|json")
	cmd.Flags().StringVar(&opts.FailOn, "fail-on", "error", "Severity threshold used for diagnostics: error|warn")
	cmd.Flags().BoolVar(&opts.FailOnEmpty, "fail-on-empty", false, "Mark diagnostics failed if no family instances match")
	cmd.Flags().BoolVar(&opts.SummaryOnly, "summary-only", false, "Print only aggregate diagnostics")
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, "Enable diagnostic output")

	return cmd
}

package cli

import (
	"errors"
	"fmt"
	"strings"

	"kyn/internal/changes"
	"kyn/internal/config"
	"kyn/internal/family"
	"kyn/internal/report"
	"kyn/internal/rules"

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
			cwd, err := resolveCWD(opts.Cwd)
			if err != nil {
				return usageError("invalid --cwd: %v", err)
			}

			effectiveOpts, autoMode, err := applyAutoInputMode(opts, cwd)
			if err != nil {
				return usageError("invalid options: %v", err)
			}
			if err := validateCheckOptions(effectiveOpts, "explain"); err != nil {
				return usageError("invalid options: %v", err)
			}

			cfg, cfgPath, err := config.Load(cwd, effectiveOpts.ConfigPath)
			if err != nil {
				return usageError("invalid config: %v", err)
			}

			filesFrom := effectiveOpts.FilesFrom
			if effectiveOpts.Stdin {
				filesFrom = "-"
			}

			changedResult, err := changes.CollectDetailed(changes.Input{
				Cwd:       cwd,
				FilesCSV:  effectiveOpts.FilesCSV,
				FilesFrom: filesFrom,
				Base:      effectiveOpts.Base,
				Head:      effectiveOpts.Head,
			})
			if err != nil {
				if errors.Is(err, changes.ErrGitFailure) {
					return runtimeError("git change detection failed: %v", err)
				}
				return usageError("invalid change input: %v", err)
			}

			instances, err := family.Resolve(cfg, changedResult.Files)
			if err != nil {
				return runtimeError("family resolution failed: %v", err)
			}

			changedSet := make(map[string]struct{}, len(changedResult.Files))
			for _, f := range changedResult.Files {
				changedSet[f] = struct{}{}
			}

			summary, err := rules.Explain(rules.EvalInput{
				Cwd:          cwd,
				FailOn:       effectiveOpts.FailOn,
				FailOnEmpty:  effectiveOpts.FailOnEmpty,
				Changed:      changedSet,
				StatusByFile: changedResult.StatusByFile,
				Rules:        cfg.Rules,
				Instances:    instances,
			})
			if err != nil {
				return runtimeError("explain evaluation failed: %v", err)
			}

			selectedModes, err := selectedInputModes(effectiveOpts)
			if err != nil {
				return usageError("invalid options: %v", err)
			}
			mode := "unknown"
			if len(selectedModes) > 0 {
				mode = selectedModes[0]
			}

			if effectiveOpts.Verbose {
				_, _ = fmt.Fprintf(
					cmd.OutOrStdout(),
					"config=%s families=%d rules=%d changed=%d instances=%d mode=%s autoMode=%t\n\n",
					cfgPath,
					len(cfg.Families),
					len(cfg.Rules),
					len(changedResult.Files),
					len(instances),
					mode,
					autoMode,
				)
			}

			if effectiveOpts.Format == "json" {
				out, err := report.RenderExplainJSON(summary)
				if err != nil {
					return runtimeError("json render failed: %v", err)
				}
				_, _ = cmd.OutOrStdout().Write(out)
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
			} else {
				_, _ = cmd.OutOrStdout().Write([]byte(report.RenderExplainText(summary, effectiveOpts.SummaryOnly)))
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

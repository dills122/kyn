package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"kyn/internal/changes"
	"kyn/internal/config"
	"kyn/internal/family"
	"kyn/internal/report"
	"kyn/internal/rules"

	"github.com/spf13/cobra"
)

func newCheckCommand() *cobra.Command {
	opts := checkOptions{}
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Evaluate changed files against configured family/rule relationships",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateCheckOptions(opts); err != nil {
				return usageError("invalid options: %v", err)
			}

			cwd, err := resolveCWD(opts.Cwd)
			if err != nil {
				return usageError("invalid --cwd: %v", err)
			}

			cfg, cfgPath, err := config.Load(cwd, opts.ConfigPath)
			if err != nil {
				return usageError("invalid config: %v", err)
			}

			changedFiles, err := changes.Collect(changes.Input{
				Cwd:       cwd,
				FilesCSV:  opts.FilesCSV,
				FilesFrom: opts.FilesFrom,
				Base:      opts.Base,
				Head:      opts.Head,
			})
			if err != nil {
				if errors.Is(err, changes.ErrGitFailure) {
					return runtimeError("git change detection failed: %v", err)
				}
				return usageError("invalid change input: %v", err)
			}

			instances, err := family.Resolve(cfg, changedFiles)
			if err != nil {
				return runtimeError("family resolution failed: %v", err)
			}

			changedSet := make(map[string]struct{}, len(changedFiles))
			for _, f := range changedFiles {
				changedSet[f] = struct{}{}
			}

			summary, err := rules.Evaluate(rules.EvalInput{
				Cwd:         cwd,
				FailOn:      opts.FailOn,
				FailOnEmpty: opts.FailOnEmpty,
				Changed:     changedSet,
				Rules:       cfg.Rules,
				Instances:   instances,
			})
			if err != nil {
				return runtimeError("rule evaluation failed: %v", err)
			}

			if opts.Verbose {
				_, _ = fmt.Fprintf(
					cmd.OutOrStdout(),
					"config=%s families=%d rules=%d changed=%d instances=%d\n\n",
					cfgPath,
					len(cfg.Families),
					len(cfg.Rules),
					len(changedFiles),
					len(instances),
				)
			}

			if opts.Format == "json" {
				out, err := report.RenderJSON(summary)
				if err != nil {
					return runtimeError("json render failed: %v", err)
				}
				_, _ = cmd.OutOrStdout().Write(out)
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
			} else {
				_, _ = cmd.OutOrStdout().Write([]byte(report.RenderText(summary)))
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
			}

			if !summary.OK {
				return ruleFailureError()
			}
			return nil
		},
	}
	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&opts.ConfigPath, "config", "", "Path to Kyn config file")
	cmd.Flags().StringVar(&opts.FilesCSV, "files", "", "Comma-separated changed files")
	cmd.Flags().StringVar(&opts.FilesFrom, "files-from", "", "Path to changed files list (one per line)")
	cmd.Flags().StringVar(&opts.Base, "base", "", "Git base ref/SHA for diff detection")
	cmd.Flags().StringVar(&opts.Head, "head", "", "Git head ref/SHA for diff detection")
	cmd.Flags().StringVar(&opts.Cwd, "cwd", ".", "Working directory")
	cmd.Flags().StringVar(&opts.Format, "format", "text", "Output format: text|json")
	cmd.Flags().StringVar(&opts.FailOn, "fail-on", "error", "Minimum severity that fails command: error|warn")
	cmd.Flags().BoolVar(&opts.FailOnEmpty, "fail-on-empty", false, "Fail if no family instances match")
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, "Enable diagnostic output")

	return cmd
}

func validateCheckOptions(opts checkOptions) error {
	switch opts.Format {
	case "text", "json":
	default:
		return fmt.Errorf("invalid --format %q; expected text|json", opts.Format)
	}

	switch opts.FailOn {
	case "error", "warn":
	default:
		return fmt.Errorf("invalid --fail-on %q; expected error|warn", opts.FailOn)
	}

	modes := 0
	if strings.TrimSpace(opts.FilesCSV) != "" {
		modes++
	}
	if strings.TrimSpace(opts.FilesFrom) != "" {
		modes++
	}
	if opts.Base != "" || opts.Head != "" {
		if opts.Base == "" || opts.Head == "" {
			return errors.New("--base and --head must be provided together")
		}
		modes++
	}

	if modes != 1 {
		return errors.New("exactly one input mode is required: --files | --files-from | --base+--head")
	}

	return nil
}

func resolveCWD(cwd string) (string, error) {
	if strings.TrimSpace(cwd) == "" {
		cwd = "."
	}
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return "", err
	}
	stat, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !stat.IsDir() {
		return "", fmt.Errorf("%s is not a directory", abs)
	}
	return abs, nil
}

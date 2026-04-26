package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
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
		Long: strings.TrimSpace(`
Evaluate changed files against configured family/rule relationships.

Happy path:
  kyn check -c kyn.config.yaml

Core flags:
  -c, --config
  -o, --format
  --fail-on

Input mode flags:
  --files
  --files-from (use '-' to read from stdin)
  --stdin
  --base + --head
  --strict-input-mode

Auto mode (default unless --strict-input-mode):
  - If no input mode is selected and --cwd is a git repo, Kyn uses git diff
    with default refs: base=origin/main, head=HEAD.
  - Override defaults with env vars KYN_BASE_REF and KYN_HEAD_REF.

Advanced flags:
  --summary-only
  --dry-run-resolve
  --show-passes
  --fail-on-empty
  --verbose
  --cwd
`),
		Example: strings.TrimSpace(`
  # Fastest happy path
  kyn check -c kyn.config.yaml

  # CI happy path (git refs)
  kyn check -c kyn.config.yaml --base origin/main --head HEAD

  # Piped changed-file list
  git diff --name-only origin/main...HEAD | kyn check -c kyn.config.yaml --files-from -

  # Explicit files
  kyn check -c kyn.config.yaml -f libs/ui/button/button.component.ts,libs/ui/button/button.component.html
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
			if err := validateCheckOptions(effectiveOpts, "check", true); err != nil {
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

			selectedModes, err := selectedInputModes(effectiveOpts)
			if err != nil {
				return usageError("invalid options: %v", err)
			}
			mode := "unknown"
			if len(selectedModes) > 0 {
				mode = selectedModes[0]
			}
			if effectiveOpts.DryRun {
				resolveReport := report.NewResolveReport(
					mode,
					effectiveOpts.Base,
					effectiveOpts.Head,
					changedResult.Files,
					instances,
					effectiveOpts.SummaryOnly,
				)
				if effectiveOpts.Format == "json" {
					out, err := report.RenderResolveJSON(resolveReport)
					if err != nil {
						return runtimeError("json render failed: %v", err)
					}
					_, _ = cmd.OutOrStdout().Write(out)
					_, _ = cmd.OutOrStdout().Write([]byte("\n"))
				} else {
					_, _ = cmd.OutOrStdout().Write([]byte(report.RenderResolveText(resolveReport)))
					_, _ = cmd.OutOrStdout().Write([]byte("\n"))
				}
				return nil
			}

			changedSet := make(map[string]struct{}, len(changedResult.Files))
			for _, f := range changedResult.Files {
				changedSet[f] = struct{}{}
			}

			summary, err := rules.Evaluate(rules.EvalInput{
				Cwd:          cwd,
				FailOn:       effectiveOpts.FailOn,
				FailOnEmpty:  effectiveOpts.FailOnEmpty,
				Changed:      changedSet,
				StatusByFile: changedResult.StatusByFile,
				Rules:        cfg.Rules,
				Instances:    instances,
			})
			if err != nil {
				return runtimeError("rule evaluation failed: %v", err)
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
				var (
					out []byte
					err error
				)
				if effectiveOpts.SummaryOnly {
					out, err = report.RenderJSONSummary(summary)
				} else {
					out, err = report.RenderJSON(summary)
				}
				if err != nil {
					return runtimeError("json render failed: %v", err)
				}
				_, _ = cmd.OutOrStdout().Write(out)
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
			} else if effectiveOpts.Format == "sarif" {
				out, err := report.RenderSARIF(summary)
				if err != nil {
					return runtimeError("sarif render failed: %v", err)
				}
				_, _ = cmd.OutOrStdout().Write(out)
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
			} else {
				_, _ = cmd.OutOrStdout().Write([]byte(report.RenderText(summary, report.TextOptions{
					ShowPasses:  effectiveOpts.ShowPasses,
					SummaryOnly: effectiveOpts.SummaryOnly,
				})))
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
			}

			if !summary.OK {
				return ruleFailureError()
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
	cmd.Flags().StringVarP(&opts.Format, "format", "o", "text", "Output format: text|json|sarif")
	cmd.Flags().StringVar(&opts.FailOn, "fail-on", "error", "Minimum severity that fails command: error|warn")
	cmd.Flags().BoolVar(&opts.FailOnEmpty, "fail-on-empty", false, "Fail if no family instances match")
	cmd.Flags().BoolVar(&opts.SummaryOnly, "summary-only", false, "Print only aggregate results")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run-resolve", false, "Resolve families/kin only; skip rule evaluation")
	cmd.Flags().BoolVar(&opts.ShowPasses, "show-passes", false, "Include passing rule results in text output")
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, "Enable diagnostic output")

	return cmd
}

func validateCheckOptions(opts checkOptions, command string, allowSARIF bool) error {
	switch opts.Format {
	case "text", "json":
	case "sarif":
		if !allowSARIF {
			return fmt.Errorf("invalid --format %q; %s supports text|json", opts.Format, command)
		}
	default:
		if allowSARIF {
			return fmt.Errorf("invalid --format %q; expected text|json|sarif", opts.Format)
		}
		return fmt.Errorf("invalid --format %q; expected text|json", opts.Format)
	}

	switch opts.FailOn {
	case "error", "warn":
	default:
		return fmt.Errorf("invalid --fail-on %q; expected error|warn", opts.FailOn)
	}

	if opts.DryRun && opts.Format == "sarif" {
		return fmt.Errorf("--dry-run-resolve does not support --format sarif; use text or json")
	}

	selectedModes, err := selectedInputModes(opts)
	if err != nil {
		return err
	}

	if len(selectedModes) != 1 {
		if len(selectedModes) == 0 {
			if !opts.StrictInput {
				return nil
			}
			return fmt.Errorf(
				"invalid input mode: expected exactly one mode, observed none.\n"+
					"Choose one: --files | --files-from | --stdin | --base+--head.\n"+
					"Try: kyn %s --strict-input-mode --base origin/main --head HEAD",
				command,
			)
		}
		return fmt.Errorf(
			"invalid input mode: expected exactly one mode, observed multiple (%s).\n"+
				"Choose one: --files | --files-from | --stdin | --base+--head.\n"+
				"Try: kyn %s --base origin/main --head HEAD",
			strings.Join(selectedModes, " + "),
			command,
		)
	}

	return nil
}

func selectedInputModes(opts checkOptions) ([]string, error) {
	selectedModes := make([]string, 0, 4)
	if strings.TrimSpace(opts.FilesCSV) != "" {
		selectedModes = append(selectedModes, "files")
	}
	if strings.TrimSpace(opts.FilesFrom) != "" {
		selectedModes = append(selectedModes, "files-from")
	}
	if opts.Stdin {
		selectedModes = append(selectedModes, "stdin")
	}
	if opts.Base != "" || opts.Head != "" {
		if opts.Base == "" || opts.Head == "" {
			return nil, fmt.Errorf(
				"invalid git input mode: expected both --base and --head, observed base=%q head=%q.\n"+
					"Try: provide both --base <ref> and --head <ref> together",
				opts.Base,
				opts.Head,
			)
		}
		selectedModes = append(selectedModes, "git")
	}
	return selectedModes, nil
}

func applyAutoInputMode(opts checkOptions, cwd string) (checkOptions, bool, error) {
	selectedModes, err := selectedInputModes(opts)
	if err != nil {
		return opts, false, err
	}
	if len(selectedModes) > 0 || opts.StrictInput {
		return opts, false, nil
	}

	if !isGitRepo(cwd) {
		return opts, false, errors.New(
			"auto input mode unavailable: no explicit mode provided and --cwd is not a git repository.\n" +
				"Choose one: --files | --files-from | --stdin | --base+--head.\n" +
				"Try: kyn check --files-from -",
		)
	}

	opts.Base = firstNonEmpty(strings.TrimSpace(os.Getenv("KYN_BASE_REF")), "origin/main")
	opts.Head = firstNonEmpty(strings.TrimSpace(os.Getenv("KYN_HEAD_REF")), "HEAD")
	return opts, true, nil
}

func isGitRepo(cwd string) bool {
	cmd := exec.Command("git", "-C", cwd, "rev-parse", "--is-inside-work-tree")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
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

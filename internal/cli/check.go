package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dills122/kyn/internal/changes"
	"github.com/dills122/kyn/internal/report"
	"github.com/dills122/kyn/internal/rules"

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
			run, err := prepareRun(opts, "check", true)
			if err != nil {
				return err
			}
			if run.opts.DryRun {
				resolveReport := report.NewResolveReport(
					run.mode,
					run.opts.Base,
					run.opts.Head,
					run.changedResult.Files,
					run.instances,
					run.opts.SummaryOnly,
				)
				if run.opts.Format == "json" {
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

			summary, err := rules.Evaluate(run.evalInput())
			if err != nil {
				return runtimeError("rule evaluation failed: %v", err)
			}

			run.writeVerbose(cmd.ErrOrStderr())

			switch run.opts.Format {
			case "json":
				var (
					out []byte
					err error
				)
				if run.opts.SummaryOnly {
					out, err = report.RenderJSONSummary(summary)
				} else {
					out, err = report.RenderJSON(summary)
				}
				if err != nil {
					return runtimeError("json render failed: %v", err)
				}
				_, _ = cmd.OutOrStdout().Write(out)
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
			case "sarif":
				out, err := report.RenderSARIF(summary)
				if err != nil {
					return runtimeError("sarif render failed: %v", err)
				}
				_, _ = cmd.OutOrStdout().Write(out)
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
			case "rdjson":
				out, err := report.RenderRDJSON(summary)
				if err != nil {
					return runtimeError("rdjson render failed: %v", err)
				}
				_, _ = cmd.OutOrStdout().Write(out)
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
			case "checkstyle":
				out, err := report.RenderCheckstyle(summary)
				if err != nil {
					return runtimeError("checkstyle render failed: %v", err)
				}
				_, _ = cmd.OutOrStdout().Write(out)
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
			default:
				_, _ = cmd.OutOrStdout().Write([]byte(report.RenderText(summary, report.TextOptions{
					ShowPasses:  run.opts.ShowPasses,
					SummaryOnly: run.opts.SummaryOnly,
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

	registerCommonInputFlags(cmd, &opts)
	cmd.Flags().StringVarP(&opts.Format, "format", "o", "text", "Output format: text|json|sarif|rdjson|checkstyle")
	cmd.Flags().StringVar(&opts.FailOn, "fail-on", "error", "Minimum severity that fails command: error|warn")
	cmd.Flags().BoolVar(&opts.FailOnEmpty, "fail-on-empty", false, "Fail if no family instances match")
	cmd.Flags().BoolVar(&opts.SummaryOnly, "summary-only", false, "Print only aggregate results")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run-resolve", false, "Resolve families/kin only; skip rule evaluation")
	cmd.Flags().BoolVar(&opts.ShowPasses, "show-passes", false, "Include passing rule results in text output")

	return cmd
}

// machineOnlyFormats are report formats gated behind allowMachineFormats
// (currently: available to `kyn check`, not `kyn explain`). Keeping this as
// the single source of truth means adding a new machine-readable format only
// requires updating this set, not every switch that cares which formats are
// "machine-only".
var machineOnlyFormats = map[string]struct{}{
	"sarif":      {},
	"rdjson":     {},
	"checkstyle": {},
}

func isMachineOnlyFormat(format string) bool {
	_, ok := machineOnlyFormats[format]
	return ok
}

func validateCheckOptions(opts checkOptions, command string, allowMachineFormats bool) error {
	switch {
	case opts.Format == "text" || opts.Format == "json":
	case isMachineOnlyFormat(opts.Format):
		if !allowMachineFormats {
			return fmt.Errorf("invalid --format %q; %s supports text|json", opts.Format, command)
		}
	default:
		if allowMachineFormats {
			return fmt.Errorf("invalid --format %q; expected text|json|sarif|rdjson|checkstyle", opts.Format)
		}
		return fmt.Errorf("invalid --format %q; expected text|json", opts.Format)
	}
	if opts.SummaryOnly && opts.Format != "text" && opts.Format != "json" {
		return fmt.Errorf("--summary-only supports only text or json; format %s requires per-rule diagnostics", opts.Format)
	}

	switch opts.FailOn {
	case "error", "warn":
	default:
		return fmt.Errorf("invalid --fail-on %q; expected error|warn", opts.FailOn)
	}

	if opts.DryRun && isMachineOnlyFormat(opts.Format) {
		return fmt.Errorf("--dry-run-resolve does not support --format %s; use text or json", opts.Format)
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

	isRepo, err := changes.IsGitRepository(cwd)
	if err != nil {
		return opts, false, err
	}
	if !isRepo {
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

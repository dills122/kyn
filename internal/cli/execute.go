package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	ExitOK          = 0
	ExitRuleFailure = 1
	ExitUsage       = 2
	ExitRuntime     = 3
)

type checkOptions struct {
	ConfigPath  string
	FilesCSV    string
	FilesFrom   string
	Stdin       bool
	Base        string
	Head        string
	StrictInput bool
	Cwd         string
	Format      string
	FailOn      string
	FailOnEmpty bool
	SummaryOnly bool
	DryRun      bool
	ShowPasses  bool
	Verbose     bool
}

type codedError struct {
	code int
	msg  string
}

func (e codedError) Error() string {
	return e.msg
}

func usageError(format string, args ...any) error {
	return codedError{
		code: ExitUsage,
		msg:  fmt.Sprintf(format, args...),
	}
}

func runtimeError(format string, args ...any) error {
	return codedError{
		code: ExitRuntime,
		msg:  fmt.Sprintf(format, args...),
	}
}

func ruleFailureError() error {
	return codedError{code: ExitRuleFailure}
}

// Execute is the entrypoint for the kyn CLI binary.
func Execute() int {
	root := newRootCommand()
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)

	if err := root.Execute(); err != nil {
		var coded codedError
		if errors.As(err, &coded) {
			if coded.msg != "" {
				_, _ = fmt.Fprintln(os.Stderr, coded.Error())
			}
			return coded.code
		}
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return ExitUsage
	}

	return ExitOK
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kyn",
		Short: "Evaluate changed files against related-file rules in CI",
		Long: strings.TrimSpace(`
Kyn enforces change contracts between source files and related stories, tests,
documentation, generated artifacts, and configuration.

Start with 'kyn init', preview path resolution with 'kyn check --dry-run-resolve',
then use 'kyn check' as the policy gate. Use 'kyn explain' to investigate why a
rule applied, passed, failed, or skipped.
`),
		Example: dedentExample(`
  # Generate a version 2 starter config
  kyn init --preset web-ui

  # Preview one changed file while tuning the config
  kyn check --dry-run-resolve -f src/components/Button.component.ts

  # Check the current Git change set
  kyn check

  # Diagnose rule decisions without using explain as a policy gate
  kyn explain
`),
		Version: versionString(),
	}
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetVersionTemplate("{{.Version}}\n")

	cmd.AddCommand(newCheckCommand())
	cmd.AddCommand(newExplainCommand())
	cmd.AddCommand(newInitCommand())
	cmd.AddCommand(newConfigCommand())
	cmd.AddCommand(newVersionCommand())
	return cmd
}

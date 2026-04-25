package cli

import (
	"errors"
	"fmt"
	"os"

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
	Base        string
	Head        string
	Cwd         string
	Format      string
	FailOn      string
	FailOnEmpty bool
	Verbose     bool
}

type codedError struct {
	code int
	err  error
}

func (e codedError) Error() string {
	return e.err.Error()
}

func usageError(format string, args ...any) error {
	return codedError{
		code: ExitUsage,
		err:  fmt.Errorf(format, args...),
	}
}

func runtimeError(format string, args ...any) error {
	return codedError{
		code: ExitRuntime,
		err:  fmt.Errorf(format, args...),
	}
}

// Execute is the entrypoint for the kyn CLI binary.
func Execute() int {
	root := newRootCommand()
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)

	if err := root.Execute(); err != nil {
		var coded codedError
		if errors.As(err, &coded) {
			_, _ = fmt.Fprintln(os.Stderr, coded.Error())
			return coded.code
		}
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return ExitRuntime
	}

	return ExitOK
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kyn",
		Short: "Evaluate changed files against related-file rules in CI",
		Long:  "Kyn is a CLI for enforcing related-file relationship rules in CI.",
	}
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	cmd.AddCommand(newCheckCommand())
	return cmd
}

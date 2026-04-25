package cli

import (
	"fmt"
	"io"
	"os"
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

// Execute is the entrypoint for the `kyn` CLI binary.
func Execute() int {
	args := os.Args[1:]
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		printRootHelp(os.Stdout)
		return ExitOK
	}

	switch args[0] {
	case "check":
		return runCheck(args[1:], os.Stdout, os.Stderr)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		printRootHelp(os.Stderr)
		return ExitUsage
	}
}

func printRootHelp(w io.Writer) {
	_, _ = io.WriteString(w, `kyn is a CLI for enforcing related-file rules in CI.

Usage:
  kyn <command> [flags]

Commands:
  check    Evaluate changed files against configured family/rule relationships
  help     Show help

Run "kyn check --help" for command details.
`)
}

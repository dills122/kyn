package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

func runCheck(args []string, stdout io.Writer, stderr io.Writer) int {
	opts := checkOptions{}
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(stderr)

	fs.StringVar(&opts.ConfigPath, "config", "", "Path to Kyn config file")
	fs.StringVar(&opts.FilesCSV, "files", "", "Comma-separated changed files")
	fs.StringVar(&opts.FilesFrom, "files-from", "", "Path to changed files list (one per line)")
	fs.StringVar(&opts.Base, "base", "", "Git base ref/SHA for diff detection")
	fs.StringVar(&opts.Head, "head", "", "Git head ref/SHA for diff detection")
	fs.StringVar(&opts.Cwd, "cwd", ".", "Working directory")
	fs.StringVar(&opts.Format, "format", "text", "Output format: text|json")
	fs.StringVar(&opts.FailOn, "fail-on", "error", "Minimum severity that fails command: error|warn")
	fs.BoolVar(&opts.FailOnEmpty, "fail-on-empty", false, "Fail if no family instances match")
	fs.BoolVar(&opts.Verbose, "verbose", false, "Enable diagnostic output")

	fs.Usage = func() {
		_, _ = io.WriteString(stderr, `Usage:
  kyn check [flags]

Flags:
  --config <path>        Path to Kyn config file
  --files <csv>          Comma-separated changed files
  --files-from <path>    Path to file containing changed files (one per line)
  --base <ref>           Git base ref/SHA for diff detection
  --head <ref>           Git head ref/SHA for diff detection
  --cwd <path>           Working directory (default ".")
  --format <format>      Output format: text|json (default "text")
  --fail-on <level>      Minimum failing severity: error|warn (default "error")
  --fail-on-empty        Fail if no results are produced
  --verbose              Print diagnostic information
  -h, --help             Show help
`)
	}

	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintf(stderr, "unexpected args: %s\n", strings.Join(fs.Args(), " "))
		fs.Usage()
		return ExitUsage
	}

	if err := validateCheckOptions(opts); err != nil {
		fmt.Fprintf(stderr, "invalid options: %v\n\n", err)
		fs.Usage()
		return ExitUsage
	}

	_, _ = io.WriteString(stdout, "kyn check: scaffold ready\n")
	return ExitOK
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

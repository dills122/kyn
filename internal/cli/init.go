package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type initOptions struct {
	ConfigPath string
	Cwd        string
	Force      bool
	Preset     string
}

func newInitCommand() *cobra.Command {
	opts := initOptions{}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a starter Kyn config file",
		Long: strings.TrimSpace(`
Generate a starter Kyn config file.

This command creates a version 2 starter config with clear defaults so teams can
run 'kyn check' quickly and then adapt rules/families to their repo.
`),
		Example: strings.TrimSpace(`
  # Create default config in current directory
  kyn init

  # Write to explicit path
  kyn init --config .kyn/kyn.config.yaml

  # Overwrite existing config
  kyn init --force
`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, err := resolveCWD(opts.Cwd)
			if err != nil {
				return usageError("invalid --cwd: %v", err)
			}

			if strings.TrimSpace(opts.Preset) == "" {
				opts.Preset = "web-ui"
			}
			if opts.Preset != "web-ui" {
				return usageError("invalid --preset %q; expected web-ui", opts.Preset)
			}

			configPath := strings.TrimSpace(opts.ConfigPath)
			if configPath == "" {
				configPath = "kyn.config.yaml"
			}
			target := filepath.Join(cwd, filepath.FromSlash(configPath))

			if !opts.Force {
				if _, err := os.Stat(target); err == nil {
					return usageError("config already exists at %s (use --force to overwrite)", target)
				} else if !errors.Is(err, os.ErrNotExist) {
					return runtimeError("unable to access %s: %v", target, err)
				}
			}

			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return runtimeError("unable to create config directory: %v", err)
			}

			content := starterConfigWebUI()
			if err := os.WriteFile(target, []byte(content), 0o600); err != nil {
				return runtimeError("unable to write config: %v", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", target)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Next step: kyn check -c %s\n", configPath)
			return nil
		},
	}
	cmd.SilenceUsage = true

	cmd.Flags().StringVarP(&opts.ConfigPath, "config", "c", "kyn.config.yaml", "Path to write config file")
	cmd.Flags().StringVar(&opts.Cwd, "cwd", ".", "Working directory")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Overwrite config file if it already exists")
	cmd.Flags().StringVar(&opts.Preset, "preset", "web-ui", "Starter preset template")

	return cmd
}

func starterConfigWebUI() string {
	return strings.TrimSpace(`
version: 2

families:
  - id: web-component
    groups:
      source:
        include:
          - "src/**/*.component.ts"
          - "src/**/*.component.html"
      story:
        include:
          - "src/**/*.stories.ts"
      tests:
        include:
          - "src/**/*.spec.ts"
    baseName:
      stripSuffixes:
        - ".component"
    kin:
      story: "{dir}/{base}.stories.ts"
      spec: "{dir}/{base}.spec.ts"

rules:
  - id: story-sync
    family: web-component
    severity: error
    if:
      changedAny: [source]
      kinExists: [story]
    assert:
      kinChanged: [story]
    message: "Source changed but story did not."

  - id: tests-sync
    family: web-component
    severity: warn
    if:
      changedAny: [source]
      kinExists: [spec]
    assert:
      kinChanged: [spec]
    message: "Source changed but test did not."
`) + "\n"
}

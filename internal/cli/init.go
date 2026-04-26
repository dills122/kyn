package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

  # Generate an API starter config
  kyn init --preset api

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
			content, err := starterConfig(opts.Preset)
			if err != nil {
				return usageError("%v", err)
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
	cmd.Flags().StringVar(&opts.Preset, "preset", "web-ui", "Starter preset template: web-ui|api|proto|iac")

	return cmd
}

func starterConfig(preset string) (string, error) {
	switch preset {
	case "web-ui":
		return starterConfigWebUI(), nil
	case "api":
		return starterConfigAPI(), nil
	case "proto":
		return starterConfigProto(), nil
	case "iac":
		return starterConfigIAC(), nil
	default:
		return "", fmt.Errorf("invalid --preset %q; expected one of %s", preset, strings.Join(validPresets(), ", "))
	}
}

func validPresets() []string {
	presets := []string{"api", "iac", "proto", "web-ui"}
	sort.Strings(presets)
	return presets
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

func starterConfigAPI() string {
	return strings.TrimSpace(`
version: 2

families:
  - id: go-api-handler
    groups:
      source:
        include:
          - "internal/**/*handler.go"
          - "internal/**/*service.go"
      tests:
        include:
          - "internal/**/*_test.go"
    baseName:
      stripSuffixes:
        - "_handler"
        - "_service"
    kin:
      test: "{dir}/{name}_test.go"

rules:
  - id: api-tests-sync
    family: go-api-handler
    severity: error
    if:
      changedAny: [source]
      kinExists: [test]
    assert:
      kinChanged: [test]
    message: "API source changed but test did not."
`) + "\n"
}

func starterConfigProto() string {
	return strings.TrimSpace(`
version: 2

families:
  - id: proto-contract
    groups:
      source:
        include:
          - "proto/**/*.proto"
      generated:
        include:
          - "gen/**/*.pb.go"
    kin:
      generatedGo: "gen/{dir}/{name}.pb.go"

rules:
  - id: proto-generated-sync
    family: proto-contract
    severity: error
    if:
      changedAny: [source]
      kinExists: [generatedGo]
    assert:
      kinChanged: [generatedGo]
    message: "Proto contract changed but generated Go output did not."
`) + "\n"
}

func starterConfigIAC() string {
	return strings.TrimSpace(`
version: 2

families:
  - id: terraform-module
    groups:
      source:
        include:
          - "terraform/**/*.tf"
      docs:
        include:
          - "terraform/**/README.md"
    kin:
      readme: "{dir}/README.md"

rules:
  - id: terraform-docs-sync
    family: terraform-module
    severity: warn
    if:
      changedAny: [source]
      kinExists: [readme]
    assert:
      kinChanged: [readme]
    message: "Terraform module changed but module README did not."
`) + "\n"
}

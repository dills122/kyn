package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"kyn/internal/config"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type migrateOptions struct {
	ConfigPath string
	Cwd        string
	From       string
	To         string
	Out        string
	InPlace    bool
	Backup     bool
	Force      bool
}

func newConfigMigrateCommand() *cobra.Command {
	opts := migrateOptions{}
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate config schema versions safely",
		Long: strings.TrimSpace(`
Migrate config schema versions with validation safeguards.

Current supported migration:
  v1 -> v2

Safety behavior:
  - validates input config before migration
  - validates migrated output before writing
  - writes to a separate output file by default
  - when using --in-place, creates a backup by default
`),
		Example: strings.TrimSpace(`
  # Safe default: write side-by-side output
  kyn config migrate --config kyn.config.yaml --from v1 --to v2

  # Explicit output path
  kyn config migrate --config kyn.config.yaml --from v1 --to v2 --out kyn.v2.yaml

  # In-place migration with backup
  kyn config migrate --config kyn.config.yaml --from v1 --to v2 --in-place
`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(opts.From) != "v1" || strings.TrimSpace(opts.To) != "v2" {
				return usageError("unsupported migration path --from %q --to %q; only v1 -> v2 is supported", opts.From, opts.To)
			}
			if opts.InPlace && strings.TrimSpace(opts.Out) != "" {
				return usageError("--out cannot be used with --in-place")
			}

			cwd, err := resolveCWD(opts.Cwd)
			if err != nil {
				return usageError("invalid --cwd: %v", err)
			}

			inCfg, inPath, err := config.Load(cwd, opts.ConfigPath)
			if err != nil {
				return usageError("invalid config: %v", err)
			}
			if inCfg.Version != 1 {
				return usageError("input config version is %d; expected version 1 for v1 -> v2 migration", inCfg.Version)
			}

			outCfg, err := config.MigrateV1ToV2(inCfg)
			if err != nil {
				return runtimeError("migration failed: %v", err)
			}

			encoded, err := yaml.Marshal(outCfg)
			if err != nil {
				return runtimeError("failed to encode migrated config: %v", err)
			}

			targetPath := inPath
			if !opts.InPlace {
				if strings.TrimSpace(opts.Out) != "" {
					targetPath = resolveOutputPath(cwd, opts.Out)
				} else {
					targetPath = defaultMigratedPath(inPath)
				}
			}

			if !opts.InPlace {
				if err := ensureWritableTarget(targetPath, opts.Force); err != nil {
					return err
				}
			}

			backupPath := ""
			if opts.InPlace && opts.Backup {
				backupPath = inPath + ".bak"
				if !opts.Force {
					if _, err := os.Stat(backupPath); err == nil {
						return usageError("backup path already exists: %s (use --force to overwrite)", backupPath)
					} else if !errors.Is(err, os.ErrNotExist) {
						return runtimeError("unable to access backup path %s: %v", backupPath, err)
					}
				}
				if err := copyFile(inPath, backupPath); err != nil {
					return runtimeError("failed to create backup %s: %v", backupPath, err)
				}
			}

			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return runtimeError("failed to create output directory: %v", err)
			}
			if err := os.WriteFile(targetPath, encoded, 0o600); err != nil {
				return runtimeError("failed to write migrated config: %v", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Migrated config: %s -> %s\n", inPath, targetPath)
			if backupPath != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Backup created: %s\n", backupPath)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Next step: kyn check -c %s\n", pathForHint(cwd, targetPath))
			return nil
		},
	}
	cmd.SilenceUsage = true

	cmd.Flags().StringVarP(&opts.ConfigPath, "config", "c", "", "Path to input config file")
	cmd.Flags().StringVar(&opts.Cwd, "cwd", ".", "Working directory")
	cmd.Flags().StringVar(&opts.From, "from", "v1", "Source config version")
	cmd.Flags().StringVar(&opts.To, "to", "v2", "Target config version")
	cmd.Flags().StringVar(&opts.Out, "out", "", "Output path for migrated config (default: <input>.v2.yaml)")
	cmd.Flags().BoolVar(&opts.InPlace, "in-place", false, "Write migrated config back to input file")
	cmd.Flags().BoolVar(&opts.Backup, "backup", true, "Create backup when using --in-place")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Overwrite output/backup paths if they exist")

	return cmd
}

func defaultMigratedPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	switch ext {
	case ".yaml", ".yml":
		return base + ".v2" + ext
	default:
		return inputPath + ".v2.yaml"
	}
}

func resolveOutputPath(cwd string, out string) string {
	if filepath.IsAbs(out) {
		return filepath.Clean(out)
	}
	return filepath.Clean(filepath.Join(cwd, out))
}

func ensureWritableTarget(targetPath string, force bool) error {
	if _, err := os.Stat(targetPath); err == nil {
		if !force {
			return usageError("output path already exists: %s (use --force to overwrite)", targetPath)
		}
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return runtimeError("unable to access output path %s: %v", targetPath, err)
	}
	return nil
}

func copyFile(src string, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o600)
}

func pathForHint(cwd string, path string) string {
	rel, err := filepath.Rel(cwd, path)
	if err != nil {
		return path
	}
	if strings.HasPrefix(rel, "..") {
		return path
	}
	return rel
}

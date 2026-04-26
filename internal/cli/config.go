package cli

import "github.com/spf13/cobra"

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Kyn configuration files",
	}
	cmd.SilenceUsage = true

	cmd.AddCommand(newConfigMigrateCommand())
	return cmd
}

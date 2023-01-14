package cmd

import "github.com/spf13/cobra"

func criCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package",
		Short: "Install package",
	}
	cmd.AddCommand(criInstallCmd())

	return cmd
}

func criInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Download package from OCI registry and install",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

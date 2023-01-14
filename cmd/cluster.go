package cmd

import "github.com/spf13/cobra"

func clusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster [command]",
		Short: "Create Kubernetes cluster",
	}

	cmd.AddCommand(clusterInitCmd(), clusterJoinCmd())

	return cmd
}

func clusterInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Init a Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}

func clusterJoinCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join",
		Short: "Join node to the Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

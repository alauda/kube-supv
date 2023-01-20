package cmd

import (
	"os"

	"github.com/alauda/kube-supv/pkg/server"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func serverCmd() *cobra.Command {
	var kubeconfig string
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start server",
		Run: func(cmd *cobra.Command, args []string) {
			if kubeconfig != "" {
				os.Setenv(clientcmd.RecommendedConfigPathEnvVar, kubeconfig)
			}
			PrintError(server.Start)
		},
	}
	cmd.Flags().StringVar(&kubeconfig, config.KubeconfigFlagName, "", "Paths to a kubeconfig. Only required if out-of-cluster.")
	return cmd
}

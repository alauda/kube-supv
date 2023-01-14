package cmd

import (
	"github.com/alauda/kube-supv/pkg/log"
	"github.com/alauda/kube-supv/version"
	"github.com/spf13/cobra"
)

var subCmds = []func() *cobra.Command{
	clusterCmd,
	machineCmd,
	criCmd,
	certCmd,
}

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:              "kubesupv",
		Long:             "kubesupv is a tool for deploying and managing Kubernetes cluster.",
		SilenceUsage:     true,
		SilenceErrors:    true,
		TraverseChildren: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version: version.BuildVersion,
	}

	for _, c := range subCmds {
		root.AddCommand(c())
	}
	return root
}

func PrintError(f func() error) {
	err := f()
	if err != nil {
		log.Errorf(`%v`, err)
	}
}

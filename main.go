package main

import (
	"github.com/alauda/kube-supv/pkg/log"

	"github.com/alauda/kube-supv/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("%v", err)
	}
}

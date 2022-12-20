package main

import (
	"github.com/alauda/kube-supv/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

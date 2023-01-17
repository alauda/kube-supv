package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alauda/kube-supv/pkg/output"
	"github.com/alauda/kube-supv/pkg/unpack"
	"github.com/alauda/kube-supv/pkg/utils"
	"github.com/alauda/kube-supv/pkg/utils/registry"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func packageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package",
		Short: "Install package",
	}
	cmd.AddCommand(packageInstallCmd(), packageListCmd(), packageDescribeCmd(), packageRevmoeCmd())

	return cmd
}

func packageInstallCmd() *cobra.Command {
	var (
		tmpDir     string
		rootDir    string
		username   string
		password   string
		valuesPath string
		keep       bool
	)
	cmd := &cobra.Command{
		Use:   "install IMAGE",
		Short: "Download package from OCI registry and install",
		Run: func(cmd *cobra.Command, args []string) {
			PrintError(func() error {
				if len(args) < 1 {
					return fmt.Errorf(`need image`)
				}
				if len(args) > 1 {
					return fmt.Errorf(`too many arguments`)
				}
				if err := utils.MakeDir(tmpDir); err != nil {
					return err
				}
				packageDir, err := os.MkdirTemp(tmpDir, "kubesupv-package-")
				if err != nil {
					return err
				}
				if keep {
					fmt.Fprintf(cmd.OutOrStdout(), "package temporary directory: %s\n", packageDir)
				} else {
					defer os.RemoveAll(packageDir)
				}

				imageRef := args[0]
				if err := registry.PullImageLayerToLocal(imageRef, packageDir, username, password); err != nil {
					return err
				}

				var values map[string]interface{}
				{
					var in io.Reader
					if valuesPath != "" {
						if valuesPath == "-" {
							in = cmd.InOrStdin()
							valuesPath = "STDIN"
						} else {
							valuesPath = filepath.FromSlash(valuesPath)
							valuesFile, err := os.Open(valuesPath)
							if err != nil {
								return err
							}
							defer valuesFile.Close()
							in = valuesFile
						}
						var err error
						values, err = unpack.ReadValues(in)
						if err != nil {
							return errors.Wrapf(err, `read values from %s`, valuesPath)
						}
					}
				}

				if err := unpack.InstallOrUpgrade(packageDir, rootDir, unpack.DefaultRecordDir, imageRef, values); err != nil {
					return err
				}
				return nil
			})

		},
	}

	cmd.Flags().StringVar(&tmpDir, "tmp", "/tmp", "temporary directory")
	cmd.Flags().StringVar(&rootDir, "root", "/", "root directory")
	cmd.Flags().StringVar(&username, "username", "", "username used to pull image from OCI registry")
	cmd.Flags().StringVar(&password, "password", "", "password used to pull image from OCI registry")
	cmd.Flags().StringVar(&valuesPath, "values", "", "values file")
	cmd.Flags().BoolVar(&keep, "keep", false, "keep downloaded package in temporary directory")

	return cmd
}

func packageListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed packages",
		Run: func(cmd *cobra.Command, args []string) {
			PrintError(func() error {
				if len(args) > 0 {
					return fmt.Errorf(`unknown arguments`)
				}
				packages, err := unpack.ListRecords(unpack.DefaultRecordDir)
				if err != nil {
					return err
				}
				table, err := output.NewTable(cmd.OutOrStdout(), "NAME", "VERSION", "PHASE")
				if err != nil {
					return err
				}

				for _, pkg := range packages {
					if err := table.Write(pkg.Name, pkg.Version, string(pkg.Phase)); err != nil {
						return err
					}
				}
				if err := table.Flush(); err != nil {
					return err
				}
				return nil
			})
		},
	}
	return cmd
}

func packageDescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe PACKAGE",
		Short: "Describe package detail",
		Run: func(cmd *cobra.Command, args []string) {
			PrintError(func() error {
				if len(args) < 1 {
					return fmt.Errorf(`need image`)
				}
				if len(args) > 1 {
					return fmt.Errorf(`too many arguments`)
				}
				name := args[0]
				record, exist, err := unpack.LoadInstallRecord(unpack.DefaultRecordDir, name)
				if err != nil {
					return err
				}

				if !exist {
					return fmt.Errorf(`package "%s" does not exist`, name)
				}

				if err := record.Encode(cmd.OutOrStdout()); err != nil {
					return err
				}

				return nil
			})
		},
	}
	return cmd
}

func packageRevmoeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove PACKAGE1 [PACKAGE1 ...]",
		Short: "Uninstall packages",
		Run: func(cmd *cobra.Command, args []string) {
			PrintError(func() error {
				if len(args) == 0 {
					return fmt.Errorf(`need package name`)
				}
				for _, name := range args {
					if err := unpack.Uninstall(unpack.DefaultRecordDir, name); err != nil {
						return err
					}
				}
				return nil
			})
		},
	}
	return cmd
}

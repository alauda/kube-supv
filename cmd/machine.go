package cmd

import (
	"fmt"
	"strconv"

	"github.com/alauda/kube-supv/pkg/machine"
	"github.com/alauda/kube-supv/pkg/mock"
	"github.com/alauda/kube-supv/pkg/ping"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func machineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "machine",
		Short: "Collect machine informations",
	}

	cmd.AddCommand(machineInfoCmd(), machineMockCmd(), machinePingCmd(), machineConnectCmd())

	return cmd
}

func machineInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Collect machine informations to check deploying conditions",
		RunE: func(cmd *cobra.Command, args []string) error {
			mi, err := machine.CollectMachineInfo()
			if err != nil {
				return err
			}
			if err := mi.WriteJSON(cmd.OutOrStdout()); err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}

func machineMockCmd() *cobra.Command {
	var duration int
	cmd := &cobra.Command{
		Use:   "mock port1 [port2 ...]",
		Short: "Mock listeners for checking network conection",
		RunE: func(cmd *cobra.Command, args []string) error {
			var ports []int
			for _, arg := range args {
				port, err := strconv.Atoi(arg)
				if err != nil {
					return errors.Wrapf(err, `convert %s to int`, arg)
				}
				ports = append(ports, port)
			}
			if len(ports) == 0 {
				return cmd.Usage()
			}
			if err := mock.ValidatePorts(ports...); err != nil {
				return err
			}
			if duration <= 0 {
				return fmt.Errorf("--duration must > 0")
			}
			return mock.ListenTCP(duration, ports...)
		},
	}
	cmd.Flags().IntVar(&duration, "duration", 10, "listen duration in seconds")

	return cmd
}

func machinePingCmd() *cobra.Command {
	var timeout int
	cmd := &cobra.Command{
		Use:   "ping ip:port [ip:port2 ...]",
		Short: "Check network connection to mock server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if timeout <= 0 {
				return fmt.Errorf("--timeout must > 0")
			}
			return ping.PingTCPs(timeout, args...)
		},
	}
	cmd.Flags().IntVar(&timeout, "timeout", 10, "connection timeout in seconds")

	return cmd
}

func machineConnectCmd() *cobra.Command {
	var timeout int
	cmd := &cobra.Command{
		Use:   "connect ip:port [ip:port2 ...]",
		Short: "Check network connection",
		RunE: func(cmd *cobra.Command, args []string) error {
			if timeout <= 0 {
				return fmt.Errorf("--timeout must > 0")
			}
			return ping.ConnectTCPs(timeout, args...)
		},
	}
	cmd.Flags().IntVar(&timeout, "timeout", 10, "connection timeout in seconds")

	return cmd
}

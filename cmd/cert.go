package cmd

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/alauda/kube-supv/pkg/kube"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func certCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cert",
		Short: "Generate certificates",
	}

	cmd.AddCommand(certGenerateCmd(), certRenewCmd())

	return cmd
}

func certGenerateCmd() *cobra.Command {
	var (
		days              int
		caDays            int
		apiserverCertSANs string
		etcdCertSANs      string
		apiServerEndpoint string
		nodeName          string
		nodeIPStr         string
		clusterName       string
	)
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Generate all certificates and kubeconfigs for Kubernetes cluster",
		Run: func(cmd *cobra.Command, args []string) {
			PrintError(func() error {
				if nodeName == "" {
					var err error
					nodeName, err = os.Hostname()
					if err != nil {
						return errors.Wrap(err, `can not get hostname`)
					}
				}

				nodeIP := net.ParseIP(nodeIPStr)
				if nodeIP == nil {
					return fmt.Errorf(`can not parse node ip "%s"`, nodeIPStr)
				}

				config := &kube.PKIConfig{
					Days:              days,
					CADays:            caDays,
					APIServerEndpoint: apiServerEndpoint,
					NodeName:          nodeName,
					NodeIP:            nodeIP,
					ClusterName:       clusterName,
					APIServerCertSANs: strings.Split(apiserverCertSANs, ","),
					ETCDCertSANs:      strings.Split(etcdCertSANs, ","),
					Out:               cmd.OutOrStdout(),
				}
				return config.GenerateAll()
			})
		},
	}

	cmd.Flags().IntVar(&caDays, "cadays", 365*30, "CA expiry days")
	cmd.Flags().IntVar(&days, "days", 365*30, "certificate expiry days")
	cmd.Flags().StringVar(&apiServerEndpoint, "apiserver-endpoint", "https://127.0.0.1:6443", "apiserver endpoint")
	cmd.Flags().StringVar(&nodeName, "node-name", "", "node name, defualt use hostname")
	cmd.Flags().StringVar(&nodeIPStr, "node-ip", "127.0.0.1", "node ip")
	cmd.Flags().StringVar(&clusterName, "cluster-name", "cluster.local", "apiserver endpoint")
	cmd.Flags().StringVar(&etcdCertSANs, "etcd-cert-sans", "", "etcd certificate SANs, connected with comma")
	cmd.Flags().StringVar(&apiserverCertSANs, "apiserver-cert-sans", "", "apiserver certificate SANs, connected with comma")

	return cmd
}

func certRenewCmd() *cobra.Command {
	var (
		days              int
		apiserverCertSANs string
		etcdCertSANs      string
		apiServerEndpoint string
		nodeName          string
		nodeIPStr         string
		clusterName       string
	)
	cmd := &cobra.Command{
		Use:   "renew",
		Short: "renew certificates and kubeconfigs",
		Run: func(cmd *cobra.Command, args []string) {
			PrintError(func() error {
				var nodeIP net.IP
				if nodeIPStr != "" {
					nodeIP := net.ParseIP(nodeIPStr)
					if nodeIP == nil {
						return fmt.Errorf(`can not parse node ip "%s"`, nodeIPStr)
					}
				}

				config := &kube.PKIConfig{
					Days:              days,
					APIServerEndpoint: apiServerEndpoint,
					NodeName:          nodeName,
					NodeIP:            nodeIP,
					ClusterName:       clusterName,
					APIServerCertSANs: strings.Split(apiserverCertSANs, ","),
					ETCDCertSANs:      strings.Split(etcdCertSANs, ","),
					Out:               cmd.OutOrStdout(),
				}
				return config.Renew()
			})
		},
	}

	cmd.Flags().IntVar(&days, "days", 365*30, "certificate expiry days")
	cmd.Flags().StringVar(&apiServerEndpoint, "apiserver-endpoint", "", "apiserver endpoint")
	cmd.Flags().StringVar(&nodeName, "node-name", "", "node name")
	cmd.Flags().StringVar(&nodeIPStr, "node-ip", "", "node ip")
	cmd.Flags().StringVar(&clusterName, "cluster-name", "", "apiserver endpoint")
	cmd.Flags().StringVar(&etcdCertSANs, "etcd-cert-sans", "", "etcd certificate SANs, connected with comma")
	cmd.Flags().StringVar(&apiserverCertSANs, "apiserver-cert-sans", "", "apiserver certificate SANs, connected with comma")

	return cmd
}

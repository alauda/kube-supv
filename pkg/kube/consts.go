package kube

import (
	"io/fs"
	"path/filepath"
)

var (
	KubernetesDir = filepath.FromSlash("/etc/kubernetes/")

	KubernetesManifestsDir          = filepath.Join(KubernetesDir, "manifests")
	KubernetesAdminConf             = filepath.Join(KubernetesDir, "admin.conf")
	KubernetesControllerManagerConf = filepath.Join(KubernetesDir, "controller-manager.conf")
	KubernetesSchedulerConf         = filepath.Join(KubernetesDir, "scheduler.conf")
	KubernetesKubeletConf           = filepath.Join(KubernetesDir, "kubelet.conf")

	KubernetesPKIDir = filepath.Join(KubernetesDir, "pki")

	KubernetesPKICACert = filepath.Join(KubernetesPKIDir, "ca.crt")
	KubernetesPKICAKey  = filepath.Join(KubernetesPKIDir, "ca.key")

	KubernetesPKIAPIServerCert = filepath.Join(KubernetesPKIDir, "apiserver.crt")
	KubernetesPKIAPIServerKey  = filepath.Join(KubernetesPKIDir, "apiserver.key")

	KubernetesPKIAPIServerETCDClientCert = filepath.Join(KubernetesPKIDir, "apiserver-etcd-client.crt")
	KubernetesPKIAPIServerETCDClientKey  = filepath.Join(KubernetesPKIDir, "apiserver-etcd-client.key")

	KubernetesPKIAPIServerKubeletClientCert = filepath.Join(KubernetesPKIDir, "apiserver-kubelet-client.crt")
	KubernetesPKIAPIServerKubeletClientKey  = filepath.Join(KubernetesPKIDir, "apiserver-kubelet-client.key")

	KubernetesPKIFrontProxyCACert = filepath.Join(KubernetesPKIDir, "front-proxy-ca.crt")
	KubernetesPKIFrontProxyCAKey  = filepath.Join(KubernetesPKIDir, "front-proxy-ca.key")

	KubernetesPKIFrontProxyClientCert = filepath.Join(KubernetesPKIDir, "front-proxy-client.crt")
	KubernetesPKIFrontProxyClientKey  = filepath.Join(KubernetesPKIDir, "front-proxy-client.key")

	KubernetesPKIKubeletCert = filepath.Join(KubernetesPKIDir, "kubelet.cert")
	KubernetesPKIKubeletKey  = filepath.Join(KubernetesPKIDir, "kubelet.key")

	KubernetesPKISAPub = filepath.Join(KubernetesPKIDir, "sa.pub")
	KubernetesPKISAKey = filepath.Join(KubernetesPKIDir, "sa.key")

	KubernetesPKIETCDDir = filepath.Join(KubernetesPKIDir, "etcd")

	KubernetesPKIETCDCACert = filepath.Join(KubernetesPKIETCDDir, "ca.crt")
	KubernetesPKIETCDCAKey  = filepath.Join(KubernetesPKIETCDDir, "ca.key")

	KubernetesPKIETCDServerCert = filepath.Join(KubernetesPKIETCDDir, "server.crt")
	KubernetesPKIETCDServerKey  = filepath.Join(KubernetesPKIETCDDir, "server.key")

	KubernetesPKIETCDPeerCert = filepath.Join(KubernetesPKIETCDDir, "peer.crt")
	KubernetesPKIETCDPeerKey  = filepath.Join(KubernetesPKIETCDDir, "peer.key")

	KubernetesPKIETCDHealthCheckClientCert = filepath.Join(KubernetesPKIETCDDir, "healthcheck-client.crt")
	KubernetesPKIETCDHealthCheckClientKey  = filepath.Join(KubernetesPKIETCDDir, "healthcheck-client.key")

	KubeletDir        = filepath.FromSlash("/var/lib/kubelet")
	KubeletConfigYaml = filepath.Join(KubeletDir, "config.yaml")

	KubeletPKIDir                     = filepath.Join(KubeletDir, "pki")
	KubeletPKIKubeletClientCurrentPem = filepath.Join(KubeletPKIDir, "kubelet-client-current.pem")

	RootKubeConfig = filepath.FromSlash("/root/.kube/config")
)

const (
	KubernetesCACN              = "kubernetes"
	KubeAPIServerCN             = "kube-apiserver"
	FrontProxyCACN              = "front-proxy-ca"
	FrontProxyClientCN          = "front-proxy-client"
	ETCDCACN                    = "etcd-ca"
	KubeAPIServerEtcdClientCN   = "kube-apiserver-etcd-client"
	EtcdServerCN                = "etcd-server"
	KubeETCDHealthCheckClientCN = "kube-etcd-healthcheck-client"
	KubeletCN                   = "kubelet"
	APIServerKubeletClientCN    = "apiserver-kubelet-client"

	SystemMastersOrg            = "system:masters"
	SystemNodesOrg              = "system:nodes"
	SystemKubeControllerManager = "system:kube-controller-manager"
	SystemKubeScheduler         = "system:scheduler"
	KubernetesAdmin             = "kubernetes-admin"
	Admin                       = "admin"
)

const (
	ConfFileMode = fs.FileMode(0600)
)

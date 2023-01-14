package kube

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io"
	"net"

	"github.com/alauda/kube-supv/pkg/cert"
	"github.com/alauda/kube-supv/pkg/utils"
	"github.com/pkg/errors"
)

type PKIConfig struct {
	CADays            int
	Days              int
	APIServerEndpoint string
	NodeName          string
	NodeIP            net.IP
	ClusterName       string
	APIServerCertSANs []string
	ETCDCertSANs      []string
	renew             bool
	Out               io.Writer
}

func (c *PKIConfig) GenerateAll() error {
	for _, fun := range []func() error{
		c.KubernetesCA,
		c.ETCDCA,
		c.FrontProxyCA,
		c.SA,
		c.APIServer,
		c.ControllerManagerConf,
		c.SchedulerConf,
		c.AdminConf,
		c.RootKubeConf,
		c.KubeletConf,
		c.Kubelet,
		c.APIServerETCDClient,
		c.APIServerKubeletClient,
		c.FrontProxyClient,
		c.ETCDServer,
		c.ETCDPeer,
		c.ETCDHealthCheckClient,
	} {
		if err := fun(); err != nil {
			return err
		}
	}
	return nil
}

func (c *PKIConfig) Renew() error {
	c.renew = true
	for _, fun := range []func() error{
		c.APIServer,
		c.ControllerManagerConf,
		c.SchedulerConf,
		c.AdminConf,
		c.RootKubeConf,
		c.KubeletConf,
		c.Kubelet,
		c.APIServerETCDClient,
		c.APIServerKubeletClient,
		c.FrontProxyClient,
		c.ETCDServer,
		c.ETCDPeer,
		c.ETCDHealthCheckClient,
	} {
		if err := fun(); err != nil {
			return err
		}
	}
	return nil
}

func (c *PKIConfig) KubernetesCA() error {
	canLoad, err := cert.CanLoadFiles(KubernetesPKICACert, KubernetesPKICAKey)
	if err != nil {
		return err
	}

	if canLoad {
		if err := c.LoadCertKeyPair(KubernetesPKICACert, KubernetesPKICAKey); err != nil {
			return err
		}
		return nil
	}

	if err := c.generateCAToSave(KubernetesCACN, KubernetesPKICACert, KubernetesPKICAKey); err != nil {
		return err
	}

	return nil
}

func (c *PKIConfig) APIServer() error {
	caCert, caKey, err := cert.LoadCertKeyPair(KubernetesPKICACert, KubernetesPKICAKey)
	if err != nil {
		return err
	}
	loaded, err := c.tryLoadCertKey(KubernetesPKIAPIServerCert, KubernetesPKIAPIServerKey, caCert)
	if !c.renew {
		if err != nil {
			return err
		}
		if loaded != nil {
			return nil
		}
	} else {
		if loaded != nil {
			if len(c.APIServerCertSANs) == 0 {
				c.APIServerCertSANs = append(c.APIServerCertSANs, loaded.DNSNames...)
				for _, ip := range loaded.IPAddresses {
					if ip != nil {
						c.APIServerCertSANs = append(c.APIServerCertSANs, ip.String())
					}
				}
			}
		}
	}

	allterNames := append(c.localAlterNames(),
		"kubernetes",
		"kubernetes.default",
		"kubernetes.default.svc",
		"kubernetes.default.svc.cluster.local",
	)
	allterNames = append(allterNames, c.APIServerCertSANs...)

	req := cert.CertRequest{
		CommonName:  KubeAPIServerCN,
		AlterNames:  allterNames,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	if err := c.generateCertToSave(req, caCert, caKey, KubernetesPKIAPIServerCert, KubernetesPKIAPIServerKey); err != nil {
		return err
	}
	return nil
}

func (c *PKIConfig) AdminConf() error {
	user := KubernetesAdmin
	confPath := KubernetesAdminConf
	org := []string{SystemMastersOrg}

	if err := c.loadOrGenerateConfToSave(user, c.apiserverEndpoint(), confPath, "", org); err != nil {
		return err
	}
	return nil
}

func (c *PKIConfig) RootKubeConf() error {
	user := Admin
	confPath := RootKubeConfig
	org := []string{SystemMastersOrg}

	if err := c.loadOrGenerateConfToSave(user, c.localEndpoint(), confPath, "", org); err != nil {
		return err
	}
	return nil
}

func (c *PKIConfig) ControllerManagerConf() error {
	user := SystemKubeControllerManager
	confPath := KubernetesControllerManagerConf

	if err := c.loadOrGenerateConfToSave(user, c.localEndpoint(), confPath, "", nil); err != nil {
		return err
	}
	return nil
}

func (c *PKIConfig) SchedulerConf() error {
	user := SystemKubeScheduler
	confPath := KubernetesSchedulerConf

	if err := c.loadOrGenerateConfToSave(user, c.localEndpoint(), confPath, "", nil); err != nil {
		return err
	}
	return nil
}

func (c *PKIConfig) KubeletConf() error {
	user := fmt.Sprintf("system:node:%s", c.NodeName)
	confPath := KubernetesKubeletConf
	pemPath := KubeletPKIKubeletClientCurrentPem
	org := []string{SystemNodesOrg}

	if err := c.loadOrGenerateConfToSave(user, c.apiserverEndpoint(), confPath, pemPath, org); err != nil {
		return err
	}
	return nil
}

func (c *PKIConfig) FrontProxyCA() error {
	canLoad, err := cert.CanLoadFiles(KubernetesPKIFrontProxyCACert, KubernetesPKIFrontProxyCAKey)
	if err != nil {
		return err
	}

	if canLoad {
		if _, _, err := cert.LoadCertKeyPair(KubernetesPKIFrontProxyCACert, KubernetesPKIFrontProxyCAKey); err != nil {
			return err
		}
		return nil
	}

	if err := c.generateCAToSave(FrontProxyCACN, KubernetesPKIFrontProxyCACert, KubernetesPKIFrontProxyCAKey); err != nil {
		return err
	}
	return nil
}

func (c *PKIConfig) FrontProxyClient() error {
	caCert, caKey, err := cert.LoadCertKeyPair(KubernetesPKIFrontProxyCACert, KubernetesPKIFrontProxyCAKey)
	if err != nil {
		return err
	}
	loaded, err := c.tryLoadCertKey(KubernetesPKIFrontProxyClientCert, KubernetesPKIFrontProxyClientKey, caCert)
	if !c.renew {
		if err != nil {
			return err
		}
		if loaded != nil {
			return nil
		}
	}

	req := cert.CertRequest{
		CommonName:  FrontProxyClientCN,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	if err := c.generateCertToSave(req, caCert, caKey, KubernetesPKIFrontProxyClientCert, KubernetesPKIFrontProxyClientKey); err != nil {
		return err
	}
	return nil
}

func (c *PKIConfig) SA() error {
	canLoad, err := cert.CanLoadFiles(KubernetesPKISAPub, KubernetesPKISAKey)
	if err != nil {
		return err
	}

	if canLoad {
		if _, _, err := cert.LoadRSAKeyPair(KubernetesPKISAPub, KubernetesPKISAKey); err != nil {
			return err
		}
		c.out(`write %s`, KubernetesPKISAKey)
		c.out(`write %s`, KubernetesPKISAPub)
		return nil
	}

	privateKey, err := cert.GenerateRSAKey(cert.DefaultRSAKeyBits)
	if err != nil {
		return errors.Wrapf(err, `generate rsa private key for "%s"`, KubernetesPKISAKey)
	}

	if err := cert.SaveRSAPrivateKey(privateKey, KubernetesPKISAKey); err != nil {
		return err
	}
	c.out(`write %s`, KubernetesPKISAKey)

	if err := cert.SavePublickKey(privateKey, KubernetesPKISAPub); err != nil {
		return err
	}
	c.out(`write %s`, KubernetesPKISAPub)

	return nil
}

func (c *PKIConfig) Kubelet() error {
	caCert, caKey, err := cert.LoadCertKeyPair(KubernetesPKIETCDCACert, KubernetesPKIETCDCAKey)
	if err != nil {
		return err
	}

	loaded, err := c.tryLoadCertKey(KubernetesPKIKubeletCert, KubernetesPKIKubeletKey, caCert)
	if !c.renew {
		if err != nil {
			return err
		}
		if loaded != nil {
			return nil
		}
	}

	allterNames := c.localAlterNames()
	if c.renew {
		if c.NodeIP == nil {
			allterNames = append(allterNames, loaded.DNSNames...)
			for _, ip := range loaded.IPAddresses {
				if ip != nil {
					allterNames = append(allterNames, ip.String())
				}
			}
		}
	} else {
		if c.NodeIP != nil {
			allterNames = append(allterNames, c.NodeIP.String())
		}
	}

	req := cert.CertRequest{
		CommonName:  KubeletCN,
		AlterNames:  allterNames,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	}

	if err := c.generateCertToSave(req, caCert, caKey, KubernetesPKIKubeletCert, KubernetesPKIKubeletKey); err != nil {
		return err
	}
	return nil
}

func (c *PKIConfig) APIServerKubeletClient() error {
	caCert, caKey, err := cert.LoadCertKeyPair(KubernetesPKICACert, KubernetesPKICAKey)
	if err != nil {
		return err
	}

	loaded, err := c.tryLoadCertKey(KubernetesPKIAPIServerKubeletClientCert, KubernetesPKIAPIServerKubeletClientKey, caCert)
	if !c.renew {
		if err != nil {
			return err
		}
		if loaded != nil {
			return nil
		}
	}

	req := cert.CertRequest{
		CommonName:   KubeAPIServerEtcdClientCN,
		Organization: []string{SystemMastersOrg},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	if err := c.generateCertToSave(req, caCert, caKey, KubernetesPKIAPIServerKubeletClientCert, KubernetesPKIAPIServerKubeletClientKey); err != nil {
		return err
	}
	return nil
}

func (c *PKIConfig) ETCDCA() error {
	canLoad, err := cert.CanLoadFiles(KubernetesPKIETCDCACert, KubernetesPKIETCDCAKey)
	if err != nil {
		return err
	}

	if canLoad {
		if err := c.LoadCertKeyPair(KubernetesPKIETCDCACert, KubernetesPKIETCDCAKey); err != nil {
			return err
		}
		return nil
	}

	if err := c.generateCAToSave(ETCDCACN, KubernetesPKIETCDCACert, KubernetesPKIETCDCAKey); err != nil {
		return err
	}
	return nil
}

func (c *PKIConfig) ETCDServer() error {
	caCert, caKey, err := cert.LoadCertKeyPair(KubernetesPKIETCDCACert, KubernetesPKIETCDCAKey)
	if err != nil {
		return err
	}

	loaded, err := c.tryLoadCertKey(KubernetesPKIETCDServerCert, KubernetesPKIETCDServerKey, caCert)
	if !c.renew {
		if err != nil {
			return err
		}
		if loaded != nil {
			return nil
		}
	} else {
		if loaded != nil {
			if len(c.ETCDCertSANs) == 0 {
				c.ETCDCertSANs = append(c.APIServerCertSANs, loaded.DNSNames...)
				for _, ip := range loaded.IPAddresses {
					if ip != nil {
						c.ETCDCertSANs = append(c.ETCDCertSANs, ip.String())
					}
				}
			}
		}
	}

	allterNames := c.localAlterNames()
	if c.NodeIP != nil {
		allterNames = append(allterNames, c.NodeIP.String())
	}
	allterNames = append(allterNames, c.ETCDCertSANs...)

	req := cert.CertRequest{
		CommonName:  c.NodeIP.String(),
		AlterNames:  allterNames,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	}

	if err := c.generateCertToSave(req, caCert, caKey, KubernetesPKIETCDServerCert, KubernetesPKIETCDServerKey); err != nil {
		return err
	}
	return nil
}

func (c *PKIConfig) ETCDPeer() error {
	caCert, caKey, err := cert.LoadCertKeyPair(KubernetesPKIETCDCACert, KubernetesPKIETCDCAKey)
	if err != nil {
		return err
	}

	loaded, err := c.tryLoadCertKey(KubernetesPKIETCDPeerCert, KubernetesPKIETCDPeerKey, caCert)
	if !c.renew {
		if err != nil {
			return err
		}
		if loaded != nil {
			return nil
		}
	}

	allterNames := c.localAlterNames()
	if c.renew {
		if c.NodeIP == nil {
			allterNames = append(allterNames, loaded.DNSNames...)
			for _, ip := range loaded.IPAddresses {
				if ip != nil {
					allterNames = append(allterNames, ip.String())
				}
			}
		}
	} else {
		if c.NodeIP != nil {
			allterNames = append(allterNames, c.NodeIP.String())
		}
	}

	req := cert.CertRequest{
		CommonName:  c.NodeIP.String(),
		AlterNames:  allterNames,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	}

	if err := c.generateCertToSave(req, caCert, caKey, KubernetesPKIETCDPeerCert, KubernetesPKIETCDPeerKey); err != nil {
		return err
	}
	return nil
}

func (c *PKIConfig) ETCDHealthCheckClient() error {
	caCert, caKey, err := cert.LoadCertKeyPair(KubernetesPKIETCDCACert, KubernetesPKIETCDCAKey)
	if err != nil {
		return err
	}

	loaded, err := c.tryLoadCertKey(KubernetesPKIETCDHealthCheckClientCert, KubernetesPKIETCDHealthCheckClientKey, caCert)
	if !c.renew {
		if err != nil {
			return err
		}
		if loaded != nil {
			return nil
		}
	}

	req := cert.CertRequest{
		CommonName:   KubeETCDHealthCheckClientCN,
		Organization: []string{SystemMastersOrg},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	if err := c.generateCertToSave(req, caCert, caKey, KubernetesPKIETCDHealthCheckClientCert, KubernetesPKIETCDHealthCheckClientKey); err != nil {
		return err
	}
	return nil
}

func (c *PKIConfig) APIServerETCDClient() error {
	caCert, caKey, err := cert.LoadCertKeyPair(KubernetesPKIETCDCACert, KubernetesPKIETCDCAKey)
	if err != nil {
		return err
	}

	loaded, err := c.tryLoadCertKey(KubernetesPKIAPIServerETCDClientCert, KubernetesPKIAPIServerETCDClientKey, caCert)
	if !c.renew {
		if err != nil {
			return err
		}
		if loaded != nil {
			return nil
		}
	}

	req := cert.CertRequest{
		CommonName:   KubeAPIServerEtcdClientCN,
		Organization: []string{SystemMastersOrg},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	if err := c.generateCertToSave(req, caCert, caKey, KubernetesPKIAPIServerETCDClientCert, KubernetesPKIAPIServerETCDClientKey); err != nil {
		return err
	}
	return nil
}

func (c *PKIConfig) tryLoadCertKey(certPath, keyPath string, caCert *x509.Certificate) (*x509.Certificate, error) {
	canLoad, err := cert.CanLoadFiles(certPath, keyPath)
	if err != nil {
		return nil, err
	}
	if canLoad {
		loadCert, _, err := cert.LoadCertKeyPair(certPath, keyPath)
		if err != nil {
			return nil, err
		}

		if err := cert.ValidateCertWithCA(loadCert, caCert); err != nil {
			return nil, errors.Wrapf(err, `validate "%s" with CA`, certPath)
		}
		if !c.renew {
			c.out(`load %s`, certPath)
			c.out(`load %s`, keyPath)
		}
		return loadCert, nil
	}
	return nil, nil
}

func (c *PKIConfig) generateCertToSave(req cert.CertRequest, caCert *x509.Certificate, caKey *rsa.PrivateKey, certPath, keyPath string) error {
	req.Days = c.Days
	newCert, newKey, err := cert.SignRSACert(req, caCert, caKey)
	if err != nil {
		return errors.Wrapf(err, `sign certificate "%s"`, req.CommonName)
	}
	if err := cert.SaveCert(newCert, certPath); err != nil {
		return err
	}
	if err := cert.SaveRSAPrivateKey(newKey, keyPath); err != nil {
		return err
	}
	c.out(`write %s`, certPath)
	c.out(`write %s`, keyPath)
	return nil
}

func (c *PKIConfig) generateCertData(req cert.CertRequest, caCert *x509.Certificate, caKey *rsa.PrivateKey) ([]byte, []byte, []byte, error) {
	req.Days = c.Days
	newCert, newKey, err := cert.SignRSACert(req, caCert, caKey)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, `sign certificate "%s"`, req.CommonName)
	}

	certData, err := cert.CertEncodeToBytes(newCert)
	if err != nil {
		return nil, nil, nil, err
	}

	keyData, err := cert.RSAPrivateKeyEncodeToBytes(newKey)
	if err != nil {
		return nil, nil, nil, err
	}

	caData, err := cert.CertEncodeToBytes(caCert)
	if err != nil {
		return nil, nil, nil, err
	}

	return certData, keyData, caData, nil
}

func (c *PKIConfig) generateKubeConfigToSave(certData, keyData, caData []byte, server, userName, confPath, pemPath string) error {
	kubeconfig := NewKubeConfig()
	kubeconfig.CurrentContext = fmt.Sprintf("%s@%s", userName, c.ClusterName)
	kubeconfig.Clusters = []NamedCluster{
		{
			Name: c.ClusterName,
			Cluster: Cluster{
				CertificateAuthorityData: caData,
				Server:                   server,
			},
		},
	}
	if pemPath == "" {
		kubeconfig.AuthInfos = []NamedAuthInfo{
			{
				Name: userName,
				AuthInfo: AuthInfo{
					ClientCertificateData: certData,
					ClientKeyData:         keyData,
				},
			},
		}
	} else {
		kubeconfig.AuthInfos = []NamedAuthInfo{
			{
				Name: userName,
				AuthInfo: AuthInfo{
					ClientCertificate: pemPath,
					ClientKey:         pemPath,
				},
			},
		}

		file, err := utils.OpenFileToWrite(pemPath, cert.KeyFileMode)
		if err != nil {
			return err
		}

		if _, err := file.Write(certData); err != nil {
			file.Close()
			return errors.Wrapf(err, `write "%s"`, pemPath)
		}

		if _, err := file.Write(keyData); err != nil {
			file.Close()
			return errors.Wrapf(err, `write "%s"`, pemPath)
		}

		if err := file.Close(); err != nil {
			return errors.Wrapf(err, `close "%s"`, pemPath)
		}
	}

	kubeconfig.Contexts = []NamedContext{
		{
			Name: kubeconfig.CurrentContext,
			Context: Context{
				Cluster:  c.ClusterName,
				AuthInfo: userName,
			},
		},
	}

	if err := SaveKubeConfig(kubeconfig, confPath); err != nil {
		return err
	}
	c.out(`write %s`, confPath)
	return nil
}

func (c *PKIConfig) loadOrGenerateConfToSave(user, server, confPath, pemPath string, org []string) error {
	exist, err := utils.IsFileExist(confPath)
	if err != nil {
		return err
	}
	var kubeconfig *KubeConfig
	if exist {
		kubeconfig, err = LoadKubeConfig(confPath)
		if !c.renew {
			if err != nil {
				return err
			}
			c.out(`load %s`, confPath)
			return nil
		}
	}

	if server == "" {
		if kubeconfig != nil && len(kubeconfig.Clusters) > 0 {
			server = kubeconfig.Clusters[0].Cluster.Server
		}
	}

	caCert, caKey, err := cert.LoadCertKeyPair(KubernetesPKICACert, KubernetesPKICAKey)
	if err != nil {
		return err
	}
	req := cert.CertRequest{
		CommonName:   user,
		Organization: org,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	certData, keyData, caData, err := c.generateCertData(req, caCert, caKey)
	if err != nil {
		return errors.Wrapf(err, `generate "%s"`, confPath)
	}

	if err := c.generateKubeConfigToSave(certData, keyData, caData, server, user, confPath, pemPath); err != nil {
		return err
	}
	return nil
}

func (c *PKIConfig) generateCAToSave(cn string, certPath, keyPath string) error {
	caCert, caKey, err := cert.GenerateSelfSignedRSACA(cn, c.CADays)
	if err != nil {
		return errors.Wrapf(err, `generate CA "%s"`, cn)
	}
	if err := cert.SaveCert(caCert, certPath); err != nil {
		return err
	}
	if err := cert.SaveRSAPrivateKey(caKey, keyPath); err != nil {
		return err
	}

	c.out(`write %s`, certPath)
	c.out(`write %s`, keyPath)
	return nil
}

func (c *PKIConfig) LoadCertKeyPair(certPath, keyPath string) error {
	_, _, err := cert.LoadCertKeyPair(certPath, keyPath)
	if err != nil {
		c.out(`load %s`, certPath)
		c.out(`load %s`, keyPath)
	}
	return err
}

func (c *PKIConfig) out(format string, args ...interface{}) {
	if c.Out != nil {
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintln(c.Out, msg)
	}
}

func (c *PKIConfig) localAlterNames() []string {
	return []string{
		"localhost",
		"127.0.0.1",
		"::1",
	}
}

func (c *PKIConfig) localEndpoint() string {
	nodeIP := "127.0.0.1"
	if c.NodeIP == nil {
		if !c.NodeIP.To4().Equal(c.NodeIP) { // ipv6
			nodeIP = fmt.Sprintf("[%s]", c.NodeIP.String())
		} else {
			nodeIP = c.NodeIP.String()
		}
	}
	return fmt.Sprintf("https://%s:6443", nodeIP)
}

func (c *PKIConfig) apiserverEndpoint() string {
	if c.APIServerEndpoint != "" {
		return fmt.Sprintf("https://%s", c.APIServerEndpoint)
	}
	return ""
}

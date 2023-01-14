package kube

import (
	"encoding/json"
	"os"

	"github.com/alauda/kube-supv/pkg/utils"
	"github.com/alauda/kube-supv/pkg/utils/yaml/types"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type KubeConfig struct {
	Kind           string          `json:"kind,omitempty" yaml:"kind,omitempty"`
	APIVersion     string          `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Preferences    Preferences     `json:"preferences" yaml:"preferences"`
	Clusters       []NamedCluster  `json:"clusters" yaml:"clusters"`
	AuthInfos      []NamedAuthInfo `json:"users" yaml:"users"`
	Contexts       []NamedContext  `json:"contexts" yaml:"contexts"`
	CurrentContext string          `json:"current-context" yaml:"current-context"`
}

func NewKubeConfig() *KubeConfig {
	return &KubeConfig{
		Kind:       "Config",
		APIVersion: "v1",
	}
}

func LoadKubeConfig(path string) (*KubeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, `load cert from "%s"`, path)
	}
	kubeConfig := KubeConfig{}
	if err = json.Unmarshal(data, &kubeConfig); err != nil {
		err = yaml.Unmarshal(data, &kubeConfig)
	}
	if err != nil {
		return nil, errors.Wrapf(err, `load kubeconfig from "%s"`, path)
	}
	return &kubeConfig, nil
}

func SaveKubeConfig(kubeConfig *KubeConfig, path string) error {
	file, err := utils.OpenFileToWrite(path, ConfFileMode)
	if err != nil {
		return err
	}
	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	if err := encoder.Encode(kubeConfig); err != nil {
		file.Close()
		return errors.Wrapf(err, `marshal kubeconfig to "%s"`, path)
	}
	if err := file.Close(); err != nil {
		return errors.Wrapf(err, `close "%s"`, path)
	}
	return nil
}

type Preferences struct {
	Colors bool `json:"colors,omitempty" yaml:"colors,omitempty"`
}

type NamedCluster struct {
	Name    string  `json:"name" yaml:"name"`
	Cluster Cluster `json:"cluster" yaml:"cluster"`
}

type Cluster struct {
	Server                   string      `json:"server" yaml:"server"`
	TLSServerName            string      `json:"tls-server-name,omitempty" yaml:"tls-server-name,omitempty"`
	InsecureSkipTLSVerify    bool        `json:"insecure-skip-tls-verify,omitempty" yaml:"insecure-skip-tls-verify,omitempty"`
	CertificateAuthority     string      `json:"certificate-authority,omitempty" yaml:"certificate-authority,omitempty"`
	CertificateAuthorityData types.Bytes `json:"certificate-authority-data,omitempty" yaml:"certificate-authority-data,omitempty"`
	ProxyURL                 string      `json:"proxy-url,omitempty" yaml:"proxy-url,omitempty"`
	DisableCompression       bool        `json:"disable-compression,omitempty" yaml:"disable-compression,omitempty"`
}

type NamedAuthInfo struct {
	Name     string   `json:"name" yaml:"name"`
	AuthInfo AuthInfo `json:"user" yaml:"user"`
}

type AuthInfo struct {
	ClientCertificate     string              `json:"client-certificate,omitempty" yaml:"client-certificate,omitempty"`
	ClientCertificateData types.Bytes         `json:"client-certificate-data,omitempty" yaml:"client-certificate-data,omitempty"`
	ClientKey             string              `json:"client-key,omitempty" yaml:"client-key,omitempty"`
	ClientKeyData         types.Bytes         `json:"client-key-data,omitempty" yaml:"client-key-data,omitempty"`
	Token                 string              `json:"token,omitempty" yaml:"token,omitempty"`
	TokenFile             string              `json:"tokenFile,omitempty" yaml:"tokenFile,omitempty"`
	Impersonate           string              `json:"act-as,omitempty" yaml:"act-as,omitempty"`
	ImpersonateUID        string              `json:"act-as-uid,omitempty" yaml:"act-as-uid,omitempty"`
	ImpersonateGroups     []string            `json:"act-as-groups,omitempty" yaml:"act-as-groups,omitempty"`
	ImpersonateUserExtra  map[string][]string `json:"act-as-user-extra,omitempty" yaml:"act-as-user-extra,omitempty"`
	Username              string              `json:"username,omitempty" yaml:"username,omitempty"`
	Password              string              `json:"password,omitempty" yaml:"password,omitempty"`
	AuthProvider          *AuthProviderConfig `json:"auth-provider,omitempty" yaml:"auth-provider,omitempty"`
	Exec                  *ExecConfig         `json:"exec,omitempty" yaml:"exec,omitempty"`
}

type AuthProviderConfig struct {
	Name   string            `json:"name" yaml:"name"`
	Config map[string]string `json:"config,omitempty" yaml:"config,omitempty"`
}

type ExecConfig struct {
	Command            string       `json:"command" yaml:"command"`
	Args               []string     `json:"args" yaml:"args"`
	Env                []ExecEnvVar `json:"env" yaml:"env"`
	APIVersion         string       `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	InstallHint        string       `json:"installHint,omitempty" yaml:"installHint,omitempty"`
	ProvideClusterInfo bool         `json:"provideClusterInfo" yaml:"provideClusterInfo"`
}

type ExecEnvVar struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

type NamedContext struct {
	Name    string  `json:"name" yaml:"name"`
	Context Context `json:"context" yaml:"context"`
}

type Context struct {
	Cluster   string `json:"cluster" yaml:"cluster"`
	AuthInfo  string `json:"user" yaml:"user"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

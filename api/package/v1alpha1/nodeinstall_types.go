package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeInstallSpec struct {
}

type InstallStatus string

const (
	Unknown       InstallStatus = "Unknown"
	Installing    InstallStatus = "Installing"
	Upgrading     InstallStatus = "Upgrading"
	Reparing      InstallStatus = "Reparing"
	InstallFailed InstallStatus = "Failed"
	InstallReady  InstallStatus = "Ready"
)

type PackageInstall struct {
	// package image
	Image string `json:"image"`

	// package image
	Version string `json:"version"`

	// package install values
	PackageValues Values `json:"packageValues"`

	// node special values
	// +kubebuilder:pruning:PreserveUnknownFields
	NodeValues Values `json:"nodeValues"`

	// package install time
	InstallTime metav1.Time `json:"installTime"`

	// package need repair
	NeedRepair bool `json:"needRepair"`

	// package need upgrade
	NeedUpgrade bool `json:"needUpgrade"`

	// package conditions
	Conditions []metav1.Condition `json:"conditions"`

	// package status
	Status InstallStatus `json:"status"`
}

type NodeInstallPhase string

const (
	NodeInstallUpdating NodeInstallPhase = "Updating"
	NodeInstallUpdated  NodeInstallPhase = "Updated"
	NodeInstallFailed   NodeInstallPhase = "Failed"
	NodeInstallUnknown  NodeInstallPhase = "Unknown"
)

type NodeInstallStatus struct {
	// installed packages on this node
	InstalledPackages map[string]PackageInstall `json:"installedPackages"`

	// latest check time
	LatestCheckTime metav1.Time `json:"latestCheckTime"`

	// next check time
	NextCheckTime metav1.Time `json:"nextCheckTime"`

	Message string `json:"message"`

	Phase NodeInstallPhase `json:"phase"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName="nist",singular="nodeinstall"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="LatestCheck",type=string,JSONPath=`.status.latestCheckTime`
// +kubebuilder:printcolumn:name="NextCheck",type=string,JSONPath=`.status.nextCheckTime`
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.message`
type NodeInstall struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeInstallSpec   `json:"spec,omitempty"`
	Status NodeInstallStatus `json:"status,omitempty"`
}

func (c *NodeInstall) SpecInterface() interface{} {
	return c.Spec
}

// +kubebuilder:object:root=true
type NodeInstallList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Package `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeInstall{}, &NodeInstallList{})
}

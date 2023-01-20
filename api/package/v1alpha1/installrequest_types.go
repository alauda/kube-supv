package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type PackageReference struct {
	// package name
	Name string `json:"name,omitempty"`
	// package uid
	// +optional
	UID types.UID `json:"uid,omitempty"`
}

type InstallRequestSpec struct {
	// package reference
	PackageRef PackageReference `json:"packageRef"`
	// node special values
	// +kubebuilder:pruning:PreserveUnknownFields
	NodeValues Values `json:"nodelValues"`
}

type InstallRequestStatus struct {
	// install conditions
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// install phase
	Phase InstallStatus `json:"phase"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName="ireq",singular="installrequest"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Package",type=string,JSONPath=`.spec.packageRef.name`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
type InstallRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeInstallSpec   `json:"spec,omitempty"`
	Status NodeInstallStatus `json:"status,omitempty"`
}

func (c *InstallRequest) SpecInterface() interface{} {
	return c.Spec
}

// +kubebuilder:object:root=true
type InstallRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Package `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InstallRequest{}, &InstallRequestList{})
}

/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

type Values struct {
	runtime.RawExtension `json:",inline"`
}

func (v *Values) Decode() (map[string]interface{}, error) {
	var r map[string]interface{}
	if err := json.Unmarshal(v.Raw, &r); err != nil {
		return nil, err
	}
	return r, nil
}

func (v *Values) Encode(in map[string]interface{}) error {
	raw, err := json.Marshal(in)
	if err != nil {
		return err
	}
	v.Raw = raw
	return nil
}

type NodeConfig struct {
	NodeSelector metav1.LabelSelector `json:"nodeSelector"`
	Skip         bool                 `json:"skip"`
	NodeValues   Values               `json:"nodeValues"`
}

type PackageSpec struct {
	// package image
	Image string `json:"image"`
	// package vesion
	Version string `json:"version"`
	// package install values
	// +kubebuilder:pruning:PreserveUnknownFields
	Values Values `json:"values"`
	// parallelism for install or upgrade
	Parallelism int `json:"parallelism"`
	// node specific configs
	NodeConfigs []NodeConfig `json:"nodeConfigs"`
	// automatically repair damaged package
	AutoRepair bool `json:"autoRepair"`
	// automatically upgrade if necessary
	AutoUpgrade bool `json:"autoUpgrade"`
}

type NodeStatistic struct {
	Matched    int `json:"matched"`
	Ready      int `json:"ready"`
	Processing int `json:"processing"`
	Waiting    int `json:"waiting"`
}

type PackageStatus struct {
	Statistic NodeStatistic `json:"statisic"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName="pakg",singular="package"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`
// +kubebuilder:printcolumn:name="Matched",type=string,JSONPath=`.status.statisic.matched`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.statisic.ready`
// +kubebuilder:printcolumn:name="Processing",type=string,JSONPath=`.status.statisic.processing`
// +kubebuilder:printcolumn:name="Waiting",type=string,JSONPath=`.status.statisic.waiting`
type Package struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PackageSpec   `json:"spec,omitempty"`
	Status PackageStatus `json:"status,omitempty"`
}

func (c *Package) SpecInterface() interface{} {
	return c.Spec
}

// +kubebuilder:object:root=true
type PackageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Package `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Package{}, &PackageList{})
}

/*
Copyright 2025.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CivoClusterSpec defines the desired state of CivoCluster.
type CivoClusterSpec struct {
	Region               string   `json:"region"`
	KubernetesVersion    string   `json:"kubernetesVersion"`
	ControlPlaneEndpoint string   `json:"controlPlaneEndpoint,omitempty"`
	CIDRBlock            string   `json:"cidrBlock,omitempty"`
	NetworkID            string   `json:"networkID,omitempty"`
	DefaultMachineType   string   `json:"defaultMachineType"`
	NumControlPlaneNodes int      `json:"numControlPlaneNodes"`
	NumWorkerNodes       int      `json:"numWorkerNodes"`
	FirewallRules        []string `json:"firewallRules,omitempty"`
	Addons               []string `json:"addons,omitempty"`
}

// CivoClusterStatus defines the observed state of CivoCluster.
type CivoClusterStatus struct {
	Ready                bool                 `json:"ready"`
	FailureReason        *string              `json:"failureReason,omitempty"`
	ControlPlaneEndpoint string               `json:"controlPlaneEndpoint,omitempty"`
	ControlPlaneReady    bool                 `json:"controlPlaneReady"`
	NodeCount            int                  `json:"nodeCount,omitempty"`
	WorkerNodesReady     int                  `json:"workerNodesReady,omitempty"`
	NetworkID            string               `json:"networkID,omitempty"`
	FirewallID           string               `json:"firewallID,omitempty"`
	Conditions           clusterv1.Conditions `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CivoCluster is the Schema for the civoclusters API.
type CivoCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CivoClusterSpec   `json:"spec,omitempty"`
	Status CivoClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CivoClusterList contains a list of CivoCluster.
type CivoClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CivoCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CivoCluster{}, &CivoClusterList{})
}

/*
Copyright 2024 zncdatadev.

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
	commonsv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/commons/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HbaseClusterSpec defines the desired state of HbaseCluster
type HbaseClusterSpec struct {
	// +kubebuilder:validation:Optional
	Image *ImageSpec `json:"image,omitempty"`

	// +kubebuilder:validation:Required
	ClusterConfigSpec *ClusterConfigSpec `json:"clusterConfig,omitempty"`

	// +kubebuilder:validation:Optional
	ClusterOperationSpec *commonsv1alpha1.ClusterOperationSpec `json:"clusterOperation,omitempty"`

	// +kubebuilder:validation:Optional
	MasterSpec *MasterSpec `json:"master,omitempty"`

	// +kubebuilder:validation:Optional
	RegionServerSpec *RegionServerSpec `json:"regionServer,omitempty"`

	// +kubebuilder:validation:Optional
	RestServerSpec *RestServerSpec `json:"restServer,omitempty"`
}

type ClusterConfigSpec struct {

	// +kubebuilder:validation:Required
	ZookeeperConfigMapName string `json:"zookeeperConfigMapName,omitempty"`

	// +kubebuilder:validation:Required
	HdfsConfigMapName string `json:"hdfsConfigMapName,omitempty"`

	// +kubebuilder:validation:Optional
	ListenerClass string `json:"listenerClass,omitempty"`

	// +kubebuilder:validation:Optional
	Authentication *AuthenticationSpec `json:"authentication,omitempty"`

	// +kubebuilder:validation:Optional
	VectorAggregatorConfigMapName string `json:"vectorAggregatorConfigMapName,omitempty"`
}

type AuthenticationSpec struct {
	// +kubebuilder:validation:Optional
	AuthenticationClass string `json:"authenticationClass,omitempty"`

	// +kubebuilder:validation:Optional
	Oidc *OidcSpec `json:"oidc,omitempty"`

	// +kubebuilder:validation:Optional
	KerberosSecretClass string `json:"kerberosSecretClass,omitempty"`

	// +kubebuilder:validation:Optional
	TlsSecretClass string `json:"tlsSecretClass,omitempty"`
}

// OidcSpec defines the OIDC spec.
type OidcSpec struct {
	// OIDC client credentials secret. It must contain the following keys:
	//   - `CLIENT_ID`: The client ID of the OIDC client.
	//   - `CLIENT_SECRET`: The client secret of the OIDC client.
	// credentials will omit to pod environment variables.
	// +kubebuilder:validation:Required
	ClientCredentialsSecret string `json:"clientCredentialsSecret"`

	// +kubebuilder:validation:Optional
	ExtraScopes []string `json:"extraScopes,omitempty"`
}

// HbaseClusterStatus defines the observed state of HbaseCluster
type HbaseClusterStatus struct {
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// HbaseCluster is the Schema for the hbaseclusters API
type HbaseCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HbaseClusterSpec   `json:"spec,omitempty"`
	Status HbaseClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HbaseClusterList contains a list of HbaseCluster
type HbaseClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HbaseCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HbaseCluster{}, &HbaseClusterList{})
}

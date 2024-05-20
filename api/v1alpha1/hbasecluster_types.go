/*
Copyright 2024 zncdata-labs.

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
	apiv1alpha1 "github.com/zncdata-labs/hbase-operator/pkg/apis/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HbaseClusterSpec defines the desired state of HbaseCluster
type HbaseClusterSpec struct {
	// +kubebuilder:validation:Required
	Image *ImageSpec `json:"image"`

	// +kubebuilder:validation:Required
	ClusterConfigSpec *ClusterConfigSpec `json:"clusterConfig,omitempty"`

	// +kubebuilder:validation:Required
	ClusterOperationSpec *apiv1alpha1.ClusterOperationSpec `json:"clusterOperation,omitempty"`

	// +kubebuilder:validation:Optional
	MasterSpec *MasterSpec `json:"master,omitempty"`

	// +kubebuilder:validation:Optional
	RegionServerSpec *RegionServerSpec `json:"regionServer,omitempty"`

	// +kubebuilder:validation:Optional
	RestServerSpec *RestServerSpec `json:"restServer,omitempty"`
}

type ClusterConfigSpec struct {

	// +kubebuilder:validation:Required
	ZookeeperConfigMap string `json:"zookeeperConfigMap,omitempty"`

	// +kubebuilder:validation:Required
	HdfsConfigMap string `json:"hdfsConfigMap,omitempty"`

	// +kubebuilder:validation:Optional
	ListenerClass string `json:"listenerClass,omitempty"`
}

// HbaseClusterStatus defines the observed state of HbaseCluster
type HbaseClusterStatus struct {
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// HbaseCluster is the Schema for the hbaseclusters API
type HbaseCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HbaseClusterSpec   `json:"spec,omitempty"`
	Status HbaseClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HbaseClusterList contains a list of HbaseCluster
type HbaseClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HbaseCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HbaseCluster{}, &HbaseClusterList{})
}

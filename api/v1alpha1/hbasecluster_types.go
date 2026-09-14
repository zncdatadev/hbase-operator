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

	"github.com/zncdatadev/hbase-operator/internal/util/version"
)

// Role names of the HBase cluster. They are the keys of the generic Roles map produced by
// ToGenericSpec, and therefore become segments of every resource name
// ("<cluster>-<role>-<group>") and the value of the app.kubernetes.io/component label.
const (
	MasterRole       = "master"
	RegionServerRole = "regionserver"
	RestServerRole   = "restserver"
)

// HbaseClusterSpec defines the desired state of HbaseCluster
type HbaseClusterSpec struct {
	// +kubebuilder:validation:Optional
	// +default:value={"repo": "quay.io/zncdatadev", "pullPolicy": "IfNotPresent"}
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
	commonsv1alpha1.GenericClusterStatus `json:",inline"`
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

// GetSpec implements common.ClusterInterface: it bridges the typed role fields into the
// framework's generic Roles map.
func (r *HbaseCluster) GetSpec() *commonsv1alpha1.GenericClusterSpec {
	return r.Spec.ToGenericSpec()
}

// GetStatus implements common.ClusterInterface. The framework mutates the generic status
// through the returned pointer, which is why it points into the CR.
func (r *HbaseCluster) GetStatus() *commonsv1alpha1.GenericClusterStatus {
	return &r.Status.GenericClusterStatus
}

// VectorAggregatorConfigMapName implements reconciler.VectorAggregatorProvider so the framework
// owns vector.yaml generation: when a role group enables the Vector agent, the GenericReconciler
// resolves the aggregator address from this ConfigMap and renders vector.yaml into the role group
// ConfigMap. Returns "" when unset (the framework then omits vector.yaml).
func (r *HbaseCluster) VectorAggregatorConfigMapName() string {
	if r.Spec.ClusterConfigSpec == nil {
		return ""
	}
	return r.Spec.ClusterConfigSpec.VectorAggregatorConfigMapName
}

// ToGenericSpec adapts HbaseClusterSpec to the framework's GenericClusterSpec.
func (s *HbaseClusterSpec) ToGenericSpec() *commonsv1alpha1.GenericClusterSpec {
	result := &commonsv1alpha1.GenericClusterSpec{
		Image:            s.toGenericImage(),
		ClusterOperation: s.ClusterOperationSpec,
		Roles:            map[string]commonsv1alpha1.RoleSpec{},
	}

	if s.MasterSpec != nil {
		groups := make(map[string]commonsv1alpha1.RoleGroupSpec, len(s.MasterSpec.RoleGroups))
		for name, rg := range s.MasterSpec.RoleGroups {
			var cfg *commonsv1alpha1.RoleGroupConfigSpec
			if rg.Config != nil {
				cfg = rg.Config.RoleGroupConfigSpec
			}
			groups[name] = genericRoleGroup(rg.Replicas, cfg, rg.OverridesSpec)
		}
		var cfg *commonsv1alpha1.RoleGroupConfigSpec
		if s.MasterSpec.Config != nil {
			cfg = s.MasterSpec.Config.RoleGroupConfigSpec
		}
		result.Roles[MasterRole] = genericRole(s.MasterSpec.RoleConfig, cfg, s.MasterSpec.OverridesSpec, groups)
	}

	if s.RegionServerSpec != nil {
		groups := make(map[string]commonsv1alpha1.RoleGroupSpec, len(s.RegionServerSpec.RoleGroups))
		for name, rg := range s.RegionServerSpec.RoleGroups {
			var cfg *commonsv1alpha1.RoleGroupConfigSpec
			if rg.Config != nil {
				cfg = rg.Config.RoleGroupConfigSpec
			}
			groups[name] = genericRoleGroup(rg.Replicas, cfg, rg.OverridesSpec)
		}
		var cfg *commonsv1alpha1.RoleGroupConfigSpec
		if s.RegionServerSpec.Config != nil {
			cfg = s.RegionServerSpec.Config.RoleGroupConfigSpec
		}
		result.Roles[RegionServerRole] = genericRole(s.RegionServerSpec.RoleConfig, cfg, s.RegionServerSpec.OverridesSpec, groups)
	}

	if s.RestServerSpec != nil {
		groups := make(map[string]commonsv1alpha1.RoleGroupSpec, len(s.RestServerSpec.RoleGroups))
		for name, rg := range s.RestServerSpec.RoleGroups {
			var cfg *commonsv1alpha1.RoleGroupConfigSpec
			if rg.Config != nil {
				cfg = rg.Config.RoleGroupConfigSpec
			}
			groups[name] = genericRoleGroup(rg.Replicas, cfg, rg.OverridesSpec)
		}
		var cfg *commonsv1alpha1.RoleGroupConfigSpec
		if s.RestServerSpec.Config != nil {
			cfg = s.RestServerSpec.Config.RoleGroupConfigSpec
		}
		result.Roles[RestServerRole] = genericRole(s.RestServerSpec.RoleConfig, cfg, s.RestServerSpec.OverridesSpec, groups)
	}

	return result
}

// toGenericImage adapts the product ImageSpec, applying the operator's code-level defaults so
// the framework resolves exactly the image the pre-Gen3 operator ran:
// "{repo}/hbase:{productVersion}-kubedoop{kubedoopVersion}". ProductVersion falls back to
// DefaultProductVersion and KubedoopVersion to the operator build version, mirroring the old
// cluster reconciler's GetImage.
func (s *HbaseClusterSpec) toGenericImage() *commonsv1alpha1.ImageSpec {
	image := s.Image
	if image == nil {
		image = &ImageSpec{}
	}

	repo := image.Repo
	if repo == "" {
		repo = DefaultRepository
	}
	productVersion := image.ProductVersion
	if productVersion == "" {
		productVersion = DefaultProductVersion
	}
	kubedoopVersion := image.KubedoopVersion
	if kubedoopVersion == "" {
		kubedoopVersion = version.BuildVersion
	}

	return &commonsv1alpha1.ImageSpec{
		Custom:          image.Custom,
		Repo:            repo,
		ProductVersion:  productVersion,
		KubedoopVersion: kubedoopVersion,
		PullPolicy:      image.PullPolicy,
	}
}

// genericRole assembles a generic RoleSpec from the typed role's flattened fields.
func genericRole(
	roleConfig *commonsv1alpha1.RoleConfigSpec,
	config *commonsv1alpha1.RoleGroupConfigSpec,
	overrides *commonsv1alpha1.OverridesSpec,
	groups map[string]commonsv1alpha1.RoleGroupSpec,
) commonsv1alpha1.RoleSpec {
	role := commonsv1alpha1.RoleSpec{
		RoleConfig: roleConfig,
		Config:     config,
		RoleGroups: groups,
	}
	if overrides != nil {
		role.ConfigOverrides = overrides.ConfigOverrides
		role.EnvOverrides = overrides.EnvOverrides
		role.CliOverrides = overrides.CliOverrides
		role.PodOverrides = overrides.PodOverrides
	}
	return role
}

// genericRoleGroup assembles a generic RoleGroupSpec from a typed role group's flattened fields.
func genericRoleGroup(
	replicas *int32,
	config *commonsv1alpha1.RoleGroupConfigSpec,
	overrides *commonsv1alpha1.OverridesSpec,
) commonsv1alpha1.RoleGroupSpec {
	group := commonsv1alpha1.RoleGroupSpec{
		Replicas: replicas,
		Config:   config,
	}
	if overrides != nil {
		group.ConfigOverrides = overrides.ConfigOverrides
		group.EnvOverrides = overrides.EnvOverrides
		group.CliOverrides = overrides.CliOverrides
		group.PodOverrides = overrides.PodOverrides
	}
	return group
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

package v1alpha1

import (
	"testing"

	commonsv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/commons/v1alpha1"
	"k8s.io/utils/ptr"

	"github.com/zncdatadev/hbase-operator/internal/util/version"
)

// defaultGroup is the role group name every fixture below uses.
const defaultGroup = "default"

func TestToGenericSpecBridgesRoles(t *testing.T) {
	spec := &HbaseClusterSpec{
		MasterSpec: &MasterSpec{
			RoleGroups: map[string]MasterRoleGroupSpec{
				defaultGroup: {Replicas: ptr.To(int32(2))},
			},
			OverridesSpec: &commonsv1alpha1.OverridesSpec{
				EnvOverrides: map[string]string{"FOO": "bar"},
			},
		},
		RegionServerSpec: &RegionServerSpec{
			RoleGroups: map[string]RegionServerRoleGroupSpec{
				defaultGroup: {Replicas: ptr.To(int32(1))},
			},
		},
		RestServerSpec: &RestServerSpec{
			RoleGroups: map[string]RestServerRoleGroupSpec{
				defaultGroup: {Replicas: ptr.To(int32(1))},
			},
		},
	}

	generic := spec.ToGenericSpec()

	for _, role := range []string{MasterRole, RegionServerRole, RestServerRole} {
		roleSpec, ok := generic.Roles[role]
		if !ok {
			t.Fatalf("role %q missing from generic spec", role)
		}
		if _, ok := roleSpec.RoleGroups[defaultGroup]; !ok {
			t.Fatalf("role %q lost its default role group", role)
		}
	}

	master := generic.Roles[MasterRole]
	if got := master.RoleGroups[defaultGroup].Replicas; got == nil || *got != 2 {
		t.Fatalf("master replicas not bridged, got %v", got)
	}
	if master.EnvOverrides["FOO"] != "bar" {
		t.Fatalf("master env overrides not bridged, got %v", master.EnvOverrides)
	}

	// A role absent from the CR must not appear in the generic spec.
	spec.RestServerSpec = nil
	if _, ok := spec.ToGenericSpec().Roles[RestServerRole]; ok {
		t.Fatal("absent restServer role leaked into generic spec")
	}
}

func TestToGenericImageAppliesOperatorDefaults(t *testing.T) {
	// Empty spec.image resolves to the same image the pre-Gen3 operator ran.
	spec := &HbaseClusterSpec{}
	image := spec.ToGenericSpec().Image

	if image.Repo != DefaultRepository {
		t.Fatalf("repo default not applied: %q", image.Repo)
	}
	if image.ProductVersion != DefaultProductVersion {
		t.Fatalf("product version default not applied: %q", image.ProductVersion)
	}
	if image.KubedoopVersion != version.BuildVersion {
		t.Fatalf("kubedoop version default not applied: %q", image.KubedoopVersion)
	}

	want := DefaultRepository + "/" + DefaultProductName + ":" + DefaultProductVersion + "-kubedoop" + version.BuildVersion
	if got := image.GetImage(DefaultProductName); got != want {
		t.Fatalf("image resolution mismatch: got %q want %q", got, want)
	}

	// Explicit values win over defaults.
	spec.Image = &ImageSpec{ProductVersion: "2.6.2", Repo: "example.com/repo", KubedoopVersion: "1.2.3"}
	image = spec.ToGenericSpec().Image
	if got := image.GetImage(DefaultProductName); got != "example.com/repo/hbase:2.6.2-kubedoop1.2.3" {
		t.Fatalf("explicit image fields not honored: %q", got)
	}

	// Custom short-circuits everything.
	spec.Image = &ImageSpec{Custom: "example.com/custom-hbase:latest"}
	if got := spec.ToGenericSpec().Image.GetImage(DefaultProductName); got != "example.com/custom-hbase:latest" {
		t.Fatalf("custom image not honored: %q", got)
	}
}

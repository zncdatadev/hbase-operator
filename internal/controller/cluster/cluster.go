package cluster

import (
	hbasev1alpha1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/pkg/builder"
	"github.com/zncdata-labs/hbase-operator/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ExcludeFields = []string{"ConfigOverrides", "ClusterConfig", "PodOverride", "Image"}
)

const (
	MasterRoleName = "Master"
)

var _ builder.ResourceBuilder = &HbaseClusterBuilder{}

type HbaseClusterBuilder struct {
	builder.BaseResourceBuilder

	HbaseClusterSpec *hbasev1alpha1.HbaseClusterSpec
}

func (h *HbaseClusterBuilder) Build() (client.Object, error) {
	return nil, nil
}

var _ handler.ResourceHandler = &HbaseClusterreconciler{}

type HbaseClusterreconciler struct {
	handler.BaseReconciler
}

func (h *HbaseClusterreconciler) Register() error {

	roles, err := h.Builder.GetSubBuilder()

	if err != nil {
		return err

	}

	for _, role := range roles {
		reconciler, err := handler.NewRoleReconciler(h.Client, role)
		if err != nil {
			return err
		}

		if err := reconciler.Register(); err != nil {
			return err
		}

		if err := h.AddReconciler(reconciler); err != nil {
			return err
		}
	}

	return nil
}

func NewHbaseClusterreconciler(
	Client builder.Client,
	Builder builder.ResourceBuilder,
) (*HbaseClusterreconciler, error) {
	obj := &HbaseClusterreconciler{
		BaseReconciler: handler.BaseReconciler{
			Client:  Client,
			Builder: Builder,
		},
	}

	if err := obj.Register(); err != nil {
		return nil, err
	}

	return obj, nil
}

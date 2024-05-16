package reconciler

import (
	"context"
	"reflect"

	apiv1alpha1 "github.com/zncdata-labs/hbase-operator/pkg/apis/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/pkg/client"
	"github.com/zncdata-labs/hbase-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("common").WithName("reconciler")
)

type ClusterReconciler interface {
	Reconciler
	GetClusterOperation() *apiv1alpha1.ClusterOperationSpec
	GetImage() util.Image
	GetResources() []Reconciler
	AddResource(resource Reconciler)
	RegisterResources(ctx context.Context) error
}

type BaseClusterReconciler[T AnySpec] struct {
	BaseReconciler[T]
	ClusterInfo ClusterInfo
	resources   []Reconciler
}

func NewBaseClusterReconciler[T AnySpec](
	client client.ResourceClient,
	clusterInfo ClusterInfo,
	spec T,
) *BaseClusterReconciler[T] {
	return &BaseClusterReconciler[T]{
		BaseReconciler: BaseReconciler[T]{
			Client: client,
			Name:   clusterInfo.Name,
			Spec:   spec,
		},
		ClusterInfo: clusterInfo,
	}
}

func (r *BaseClusterReconciler[T]) GetResources() []Reconciler {
	return r.resources
}

func (r *BaseClusterReconciler[T]) AddResource(resource Reconciler) {
	r.resources = append(r.resources, resource)
}

func (r *BaseClusterReconciler[T]) RegisterResources(ctx context.Context) error {
	panic("unimplemented")
}

func (r *BaseClusterReconciler[T]) Reconcile(ctx context.Context) Result {
	for _, resource := range r.resources {
		result := resource.Reconcile(ctx)
		if result.RequeueOrNot() {
			return result
		}
	}
	return NewResult(false, 0, nil)
}

type RoleReconciler interface {
	ClusterReconciler
}

var _ RoleReconciler = &BaseRoleReconciler[AnySpec]{}

type BaseRoleReconciler[T AnySpec] struct {
	BaseClusterReconciler[T]

	RoleInfo RoleInfo
}

// MergeRoleGroupSpec
// merge right to left, if field of right not exist in left, add it to left.
// else skip it.
// merge will modify left, so left must be a pointer.
func (b *BaseRoleReconciler[T]) MergeRoleGroupSpec(roleGroup any) {
	leftValue := reflect.ValueOf(roleGroup)
	rightValue := reflect.ValueOf(b.Spec)

	if leftValue.Kind() == reflect.Ptr {
		leftValue = leftValue.Elem()
	} else {
		panic("roleGroup is not a pointer")
	}

	if rightValue.Kind() == reflect.Ptr {
		rightValue = rightValue.Elem()
	}

	for i := 0; i < rightValue.NumField(); i++ {
		rightField := rightValue.Field(i)
		rightFieldName := rightValue.Type().Field(i).Name

		if rightField.IsZero() {
			continue
		}

		leftField := leftValue.FieldByName(rightFieldName)

		if !leftField.IsValid() {
			leftValue.Field(i).Set(rightField)
			logger.V(5).Info("Merge role group", "field", rightFieldName, "value", rightField)
		}
	}
}

func (b *BaseRoleReconciler[T]) GetClusterOperation() *apiv1alpha1.ClusterOperationSpec {
	return b.RoleInfo.ClusterOperation
}

func (b *BaseRoleReconciler[T]) GetImage() util.Image {
	return b.RoleInfo.Image
}

func (b *BaseRoleReconciler[T]) Ready() Result {
	for _, resource := range b.resources {
		if result := resource.Ready(); result.RequeueOrNot() {
			return result
		}
	}
	return NewResult(false, 0, nil)
}

func NewBaseRoleReconciler[T AnySpec](
	client client.ResourceClient,
	roleInfo RoleInfo,
	spec T,
) *BaseRoleReconciler[T] {

	client.AddLabels(
		map[string]string{
			"app.kubernetes.io/component": roleInfo.Name,
		},
		false,
	)

	return &BaseRoleReconciler[T]{
		BaseClusterReconciler: *NewBaseClusterReconciler[T](
			client,
			roleInfo.ClusterInfo,
			spec,
		),
		RoleInfo: roleInfo,
	}
}

type ClusterInfo struct {
	Name             string
	Namespace        string
	ClusterOperation *apiv1alpha1.ClusterOperationSpec
	Image            util.Image
}

type RoleInfo struct {
	ClusterInfo
	Name string
}

func (r *RoleInfo) GetFullName() string {
	return r.ClusterInfo.Name + "-" + r.Name
}

type RoleGroupInfo struct {
	RoleInfo
	Name     string
	Replicas *int32

	PodDisruptionBudget *apiv1alpha1.PodDisruptionBudgetSpec
	CommandOverrides    []string
	EnvOverrides        map[string]string
	PodOverrides        *corev1.PodTemplateSpec
}

func (r *RoleGroupInfo) GetFullName() string {
	return r.RoleInfo.GetFullName() + "-" + r.Name
}

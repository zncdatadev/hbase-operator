package master

import (
	"time"

	hbasev1alph1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
	"github.com/zncdatadev/hbase-operator/internal/controller/common"
	"github.com/zncdatadev/operator-go/pkg/builder"
	"github.com/zncdatadev/operator-go/pkg/client"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	"github.com/zncdatadev/operator-go/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

func NewStatefulSetReconciler(
	client *client.Client,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	roleGroupInfo reconciler.RoleGroupInfo,
	ports []corev1.ContainerPort,
	image *util.Image,
	stopped bool,
	spec *hbasev1alph1.MasterRoleGroupSpec,
) (reconciler.ResourceReconciler[builder.StatefulSetBuilder], error) {

	options := builder.WorkloadOptions{
		Options: builder.Options{
			ClusterName:   roleGroupInfo.GetClusterName(),
			RoleName:      roleGroupInfo.GetRoleName(),
			RoleGroupName: roleGroupInfo.GetGroupName(),
			Labels:        roleGroupInfo.GetLabels(),
			Annotations:   roleGroupInfo.GetAnnotations(),
		},
		PodOverrides:     spec.PodOverrides,
		CommandOverrides: spec.CommandOverrides,
		EnvOverrides:     spec.EnvOverrides,
	}

	if spec.Config != nil {

		var gracefulShutdownTimeout time.Duration
		var err error

		if spec.Config.GracefulShutdownTimeout != nil {
			gracefulShutdownTimeout, err = time.ParseDuration(*spec.Config.GracefulShutdownTimeout)
			if err != nil {
				return nil, err
			}
		}

		options.TerminationGracePeriod = &gracefulShutdownTimeout
		options.Resource = spec.Config.Resources
		options.Affinity = spec.Config.Affinity
	}

	krb5SecretClass, tlsSecretClass := "", ""

	if clusterConfig.Authentication != nil {
		krb5SecretClass = clusterConfig.Authentication.KerberosSecretClass
		tlsSecretClass = clusterConfig.Authentication.TlsSecretClass
	}

	stsBuilder := common.NewStatefulSetBuilder(
		client,
		roleGroupInfo.GetFullName(),
		clusterConfig,
		spec.Replicas,
		ports,
		image,
		krb5SecretClass,
		tlsSecretClass,
		options,
	)

	return reconciler.NewStatefulSet(
		client,
		roleGroupInfo.GetFullName(),
		stsBuilder,
		stopped,
	), nil
}

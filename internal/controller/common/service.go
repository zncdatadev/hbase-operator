package common

import (
	"fmt"
	"strconv"

	"github.com/zncdatadev/hbase-operator/internal/constant"
	"github.com/zncdatadev/hbase-operator/internal/util"
	"github.com/zncdatadev/operator-go/pkg/builder"
	"github.com/zncdatadev/operator-go/pkg/client"
	opconstants "github.com/zncdatadev/operator-go/pkg/constants"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	corev1 "k8s.io/api/core/v1"
)

// NewRoleGroupMetricsService creates a metrics service reconciler using a simple function approach
// This creates a headless service for metrics with Prometheus labels and annotations
func NewRoleGroupMetricsService(
	client *client.Client,
	roleGroupInfo *reconciler.RoleGroupInfo,
) reconciler.Reconciler {
	roleName := roleGroupInfo.GetRoleName()
	// Get metrics port
	metricsPort, err := GetMetricsPort(roleName)
	if err != nil {
		// Return empty reconciler on error - should not happen
		panic("role " + roleName + " failed to get metrics port: " + err.Error())
	}

	// Create service ports
	servicePorts := []corev1.ContainerPort{
		{
			Name:          constant.HBASE_METRICS_PORT_NAME,
			ContainerPort: metricsPort,
			Protocol:      corev1.ProtocolTCP,
		},
	}

	// Create service name with -metrics suffix
	serviceName := util.CreateServiceMetricsName(roleGroupInfo)

	// Determine scheme based on TLS configuration
	scheme := "http" // TODO: HTTPS should be used if TLS is enabled
	// if IsTlsEnabled(hdfs.Spec.ClusterConfig) {
	// 	scheme = "https"
	// }

	// Prepare labels (copy from roleGroupInfo and add metrics labels)
	labels := make(map[string]string)
	for k, v := range roleGroupInfo.GetLabels() {
		labels[k] = v
	}
	labels["prometheus.io/scrape"] = "true"

	// Prepare annotations (copy from roleGroupInfo and add Prometheus annotations)
	annotations := make(map[string]string)
	for k, v := range roleGroupInfo.GetAnnotations() {
		annotations[k] = v
	}
	annotations["prometheus.io/scrape"] = "true"
	annotations["prometheus.io/path"] = "/prom"
	annotations["prometheus.io/port"] = strconv.Itoa(int(metricsPort))
	annotations["prometheus.io/scheme"] = scheme

	// Create base service builder
	baseBuilder := builder.NewServiceBuilder(
		client,
		serviceName,
		servicePorts,
		func(sbo *builder.ServiceBuilderOptions) {
			sbo.Headless = true
			sbo.ListenerClass = opconstants.ClusterInternal
			sbo.Labels = labels
			sbo.MatchingLabels = roleGroupInfo.GetLabels() // Use original labels for matching
			sbo.Annotations = annotations
		},
	)

	return reconciler.NewGenericResourceReconciler(
		client,
		baseBuilder,
	)
}

func GetMetricsPort(roleName string) (int32, error) {
	switch roleName {
	case constant.MASTER_ROLE:
		return constant.HBASE_MASTER_METRICS_PORT, nil
	case constant.REGIONSERVER_ROLE:
		return constant.HBASE_REGIONSERVER_METRICS_PORT, nil
	case constant.RESTSERVER_ROLE:
		return constant.HBASE_REST_METRICS_PORT, nil
	default:
		return 0, fmt.Errorf("unknown role name: %s", roleName)
	}
}

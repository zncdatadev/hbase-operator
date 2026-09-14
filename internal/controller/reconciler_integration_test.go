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

package controller

import (
	"context"
	"reflect"
	"strings"
	"testing"

	commonsv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/commons/v1alpha1"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	"github.com/zncdatadev/operator-go/pkg/vector"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	hbasev1alpha1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
)

func reconcileTestCluster(
	t *testing.T,
	cr *hbasev1alpha1.HbaseCluster,
	extra ...client.Object,
) (client.Client, error) {
	t.Helper()
	k8sClient, r := newTestReconciler(t, cr, extra...)
	_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)})
	return k8sClient, err
}

func newTestReconciler(
	t *testing.T,
	cr *hbasev1alpha1.HbaseCluster,
	extra ...client.Object,
) (client.Client, *reconciler.GenericReconciler[*hbasev1alpha1.HbaseCluster]) {
	t.Helper()

	scheme := testScheme(t)
	seed := make([]client.Object, 0, 3+len(extra))
	seed = append(seed,
		cr,
		znodeConfigMap(),
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: testHdfsCM, Namespace: testNamespace}},
	)
	seed = append(seed, extra...)
	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&hbasev1alpha1.HbaseCluster{}).
		WithObjects(seed...).
		Build()
	handler := NewHbaseRoleGroupHandler(scheme)
	r, err := reconciler.NewGenericReconciler(&reconciler.GenericReconcilerConfig[*hbasev1alpha1.HbaseCluster]{
		Client:            k8sClient,
		APIReader:         k8sClient,
		Scheme:            scheme,
		Recorder:          record.NewFakeRecorder(100),
		RoleGroupHandler:  handler,
		RoleProvider:      handler,
		RoleGroupResolver: handler,
		ImageResolution: reconciler.ImageResolution{
			ProductName: hbasev1alpha1.DefaultProductName,
			Defaults: commonsv1alpha1.ImageSpec{
				Repo:            hbasev1alpha1.DefaultRepository,
				ProductVersion:  hbasev1alpha1.DefaultProductVersion,
				KubedoopVersion: "0.4.0-dev",
			},
		},
		Prototype: &hbasev1alpha1.HbaseCluster{},
	})
	if err != nil {
		t.Fatalf("NewGenericReconciler failed: %v", err)
	}
	return k8sClient, r
}

func TestGenericReconcilerSupportsOptionalRoles(t *testing.T) {
	cr := testCR(func(cr *hbasev1alpha1.HbaseCluster) {
		cr.Spec.RegionServerSpec = nil
		cr.Spec.RestServerSpec = nil
	})
	k8sClient, err := reconcileTestCluster(t, cr)
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	fetched := &hbasev1alpha1.HbaseCluster{}
	if err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(cr), fetched); err != nil {
		t.Fatal(err)
	}
	if got, want := fetched.Status.RoleGroups, map[string][]string{hbasev1alpha1.MasterRole: {testGroup}}; !reflect.DeepEqual(got, want) {
		t.Errorf("status roleGroups = %v, want %v", got, want)
	}

	var list appsv1.StatefulSetList
	if err := k8sClient.List(context.Background(), &list, client.InNamespace(testNamespace)); err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 1 || list.Items[0].Name != testMasterResource {
		t.Fatalf("StatefulSets = %v, want only hbase-master-default", objectNames(list.Items))
	}
}

func TestGenericReconcilerBuildsCompleteVectorPipeline(t *testing.T) {
	const aggregatorName = "vector-aggregator"
	cr := testCR(func(cr *hbasev1alpha1.HbaseCluster) {
		cr.Spec.RegionServerSpec = nil
		cr.Spec.RestServerSpec = nil
		cr.Spec.ClusterConfigSpec.VectorAggregatorConfigMapName = aggregatorName
		cr.Spec.MasterSpec.Config = &hbasev1alpha1.MasterConfigSpec{
			RoleGroupConfigSpec: &commonsv1alpha1.RoleGroupConfigSpec{
				Logging: &commonsv1alpha1.LoggingSpec{EnableVectorAgent: ptr.To(true)},
			},
		}
	})
	aggregator := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: aggregatorName, Namespace: testNamespace},
		Data:       map[string]string{vector.AddressKey: "vector-aggregator.default.svc:6000"},
	}
	k8sClient, err := reconcileTestCluster(t, cr, aggregator)
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	key := types.NamespacedName{Namespace: testNamespace, Name: testMasterResource}
	cm := &corev1.ConfigMap{}
	if err := k8sClient.Get(context.Background(), key, cm); err != nil {
		t.Fatal(err)
	}
	if got := cm.Data[vector.VectorConfigFileName]; !strings.Contains(got, "vector-aggregator.default.svc:6000") ||
		!strings.Contains(got, "/kubedoop/log/*/*.log4j.xml") {
		t.Errorf("vector.yaml does not connect the declared producer to the aggregator:\n%s", got)
	}
	if got := cm.Data["log4j.properties"]; !strings.Contains(got, "/kubedoop/log/master/master.log4j.xml") {
		t.Errorf("log4j.properties does not write the file Vector collects:\n%s", got)
	}

	sts := &appsv1.StatefulSet{}
	if err := k8sClient.Get(context.Background(), key, sts); err != nil {
		t.Fatal(err)
	}
	vectorContainer := findContainer(sts.Spec.Template.Spec.InitContainers, vector.VectorSidecarName)
	if vectorContainer == nil {
		t.Fatalf("Vector native sidecar missing: %v", containerNames(sts.Spec.Template.Spec.InitContainers))
	}
	if vectorContainer.RestartPolicy == nil || *vectorContainer.RestartPolicy != corev1.ContainerRestartPolicyAlways {
		t.Errorf("Vector restartPolicy = %v, want Always", vectorContainer.RestartPolicy)
	}
	if vectorContainer.Image != sts.Spec.Template.Spec.Containers[0].Image {
		t.Errorf("Vector image = %q, product image = %q", vectorContainer.Image, sts.Spec.Template.Spec.Containers[0].Image)
	}
	assertMount(t, vectorContainer.VolumeMounts, vector.VectorConfigVolumeName, true)
	assertMount(t, vectorContainer.VolumeMounts, vector.VectorDataVolumeName, false)
	assertMount(t, vectorContainer.VolumeMounts, vector.VectorLogVolumeName, false)
	assertMount(t, sts.Spec.Template.Spec.Containers[0].VolumeMounts, vector.VectorLogVolumeName, false)
	for _, volumeName := range []string{vector.VectorConfigVolumeName, vector.VectorDataVolumeName, vector.VectorLogVolumeName} {
		if findVolume(sts.Spec.Template.Spec.Volumes, volumeName) == nil {
			t.Errorf("volume %q missing", volumeName)
		}
	}
}

func TestGenericReconcilerRejectsEnabledVectorWithoutAggregator(t *testing.T) {
	cr := testCR(func(cr *hbasev1alpha1.HbaseCluster) {
		cr.Spec.RegionServerSpec = nil
		cr.Spec.RestServerSpec = nil
		cr.Spec.MasterSpec.Config = &hbasev1alpha1.MasterConfigSpec{
			RoleGroupConfigSpec: &commonsv1alpha1.RoleGroupConfigSpec{
				Logging: &commonsv1alpha1.LoggingSpec{EnableVectorAgent: ptr.To(true)},
			},
		}
	})
	_, err := reconcileTestCluster(t, cr)
	if err == nil || !strings.Contains(err.Error(), "vectorAggregatorConfigMapName is not configured") {
		t.Fatalf("Reconcile error = %v, want missing vector aggregator error", err)
	}
}

func TestGenericReconcilerLifecycleAndOwnedResourceDrift(t *testing.T) {
	ctx := context.Background()
	cr := testCR(func(cr *hbasev1alpha1.HbaseCluster) {
		cr.Spec.RegionServerSpec = nil
		cr.Spec.RestServerSpec = nil
	})
	k8sClient, r := newTestReconciler(t, cr)
	request := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	reconcileOnce := func() {
		t.Helper()
		if _, err := r.Reconcile(ctx, request); err != nil {
			t.Fatalf("Reconcile failed: %v", err)
		}
	}
	reconcileOnce()

	metricsKey := types.NamespacedName{Namespace: testNamespace, Name: "hbase-master-default-metrics"}
	metrics := &corev1.Service{}
	if err := k8sClient.Get(ctx, metricsKey, metrics); err != nil {
		t.Fatal(err)
	}
	metrics.Annotations["prometheus.io/path"] = "/drifted"
	if err := k8sClient.Update(ctx, metrics); err != nil {
		t.Fatal(err)
	}
	reconcileOnce()
	if err := k8sClient.Get(ctx, metricsKey, metrics); err != nil {
		t.Fatal(err)
	}
	if got := metrics.Annotations["prometheus.io/path"]; got != prometheusPath {
		t.Errorf("owned Service drift was not repaired: path = %q", got)
	}

	fetched := &hbasev1alpha1.HbaseCluster{}
	if err := k8sClient.Get(ctx, request.NamespacedName, fetched); err != nil {
		t.Fatal(err)
	}
	fetched.Spec.ClusterOperationSpec = &commonsv1alpha1.ClusterOperationSpec{ReconciliationPaused: true}
	if err := k8sClient.Update(ctx, fetched); err != nil {
		t.Fatal(err)
	}
	if err := k8sClient.Get(ctx, metricsKey, metrics); err != nil {
		t.Fatal(err)
	}
	metrics.Annotations["prometheus.io/path"] = "/paused-drift"
	if err := k8sClient.Update(ctx, metrics); err != nil {
		t.Fatal(err)
	}
	reconcileOnce()
	if err := k8sClient.Get(ctx, metricsKey, metrics); err != nil {
		t.Fatal(err)
	}
	if got := metrics.Annotations["prometheus.io/path"]; got != "/paused-drift" {
		t.Errorf("paused reconcile mutated the Service: path = %q", got)
	}
	if err := k8sClient.Get(ctx, request.NamespacedName, fetched); err != nil {
		t.Fatal(err)
	}
	paused := fetched.Status.GetCondition(commonsv1alpha1.ConditionPaused)
	if paused == nil || paused.Status != metav1.ConditionTrue {
		t.Errorf("Paused condition = %#v, want True", paused)
	}

	fetched.Spec.ClusterOperationSpec = &commonsv1alpha1.ClusterOperationSpec{Stopped: true}
	if err := k8sClient.Update(ctx, fetched); err != nil {
		t.Fatal(err)
	}
	reconcileOnce()
	sts := &appsv1.StatefulSet{}
	workloadKey := types.NamespacedName{Namespace: testNamespace, Name: testMasterResource}
	if err := k8sClient.Get(ctx, workloadKey, sts); err != nil {
		t.Fatal(err)
	}
	if sts.Spec.Replicas == nil || *sts.Spec.Replicas != 0 {
		t.Errorf("stopped replicas = %v, want 0", sts.Spec.Replicas)
	}
	if err := k8sClient.Get(ctx, metricsKey, metrics); err != nil {
		t.Fatal(err)
	}
	if got := metrics.Annotations["prometheus.io/path"]; got != prometheusPath {
		t.Errorf("unpausing did not resume drift repair: path = %q", got)
	}
	if err := k8sClient.Get(ctx, request.NamespacedName, fetched); err != nil {
		t.Fatal(err)
	}
	available := fetched.Status.GetCondition(commonsv1alpha1.ConditionAvailable)
	if available == nil || available.Status != metav1.ConditionFalse || available.Reason != commonsv1alpha1.ReasonStopped {
		t.Errorf("Available condition = %#v, want False/Stopped", available)
	}

	fetched.Spec.ClusterOperationSpec.Stopped = false
	if err := k8sClient.Update(ctx, fetched); err != nil {
		t.Fatal(err)
	}
	reconcileOnce()
	if err := k8sClient.Get(ctx, workloadKey, sts); err != nil {
		t.Fatal(err)
	}
	if sts.Spec.Replicas == nil || *sts.Spec.Replicas != 1 {
		t.Errorf("resumed replicas = %v, want 1", sts.Spec.Replicas)
	}
}

func objectNames(items []appsv1.StatefulSet) []string {
	names := make([]string, 0, len(items))
	for i := range items {
		names = append(names, items[i].Name)
	}
	return names
}

func containerNames(containers []corev1.Container) []string {
	names := make([]string, 0, len(containers))
	for i := range containers {
		names = append(names, containers[i].Name)
	}
	return names
}

func findContainer(containers []corev1.Container, name string) *corev1.Container {
	for i := range containers {
		if containers[i].Name == name {
			return &containers[i]
		}
	}
	return nil
}

func findVolume(volumes []corev1.Volume, name string) *corev1.Volume {
	for i := range volumes {
		if volumes[i].Name == name {
			return &volumes[i]
		}
	}
	return nil
}

func assertMount(t *testing.T, mounts []corev1.VolumeMount, name string, readOnly bool) {
	t.Helper()
	for i := range mounts {
		if mounts[i].Name == name {
			if mounts[i].ReadOnly != readOnly {
				t.Errorf("mount %q readOnly = %t, want %t", name, mounts[i].ReadOnly, readOnly)
			}
			return
		}
	}
	t.Errorf("mount %q missing", name)
}

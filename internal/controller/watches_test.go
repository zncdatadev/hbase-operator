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
	"testing"

	authv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/authentication/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	hbasev1alpha1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
)

func TestReferenceIndexes(t *testing.T) {
	cr := testCR(func(cr *hbasev1alpha1.HbaseCluster) {
		cr.Spec.ClusterConfigSpec.HdfsConfigMapName = testZnodeCM
		cr.Spec.ClusterConfigSpec.VectorAggregatorConfigMapName = "vector"
		cr.Spec.ClusterConfigSpec.Authentication = &hbasev1alpha1.AuthenticationSpec{
			AuthenticationClass: testOIDCAuthClass,
			Oidc: &hbasev1alpha1.OidcSpec{
				ClientCredentialsSecret: testOIDCCredentials,
			},
		}
	})

	if got, want := configMapReferences(cr), []string{testZnodeCM, "vector"}; !reflect.DeepEqual(got, want) {
		t.Errorf("config map references = %v, want %v", got, want)
	}
	if got, want := secretReferences(cr), []string{testOIDCCredentials}; !reflect.DeepEqual(got, want) {
		t.Errorf("secret references = %v, want %v", got, want)
	}
	if got, want := authenticationClassReferences(cr), []string{testOIDCAuthClass}; !reflect.DeepEqual(got, want) {
		t.Errorf("authentication class references = %v, want %v", got, want)
	}

	if got := configMapReferences(&corev1.ConfigMap{}); got != nil {
		t.Errorf("references for unrelated object = %v, want nil", got)
	}
	if got := configMapReferences(&hbasev1alpha1.HbaseCluster{}); got != nil {
		t.Errorf("references for cluster without config = %v, want nil", got)
	}
}

func TestReferenceRequests(t *testing.T) {
	scheme := testScheme(t)
	clusterA := testCR(func(cr *hbasev1alpha1.HbaseCluster) {
		cr.Name = "a"
		cr.Spec.ClusterConfigSpec.Authentication = &hbasev1alpha1.AuthenticationSpec{
			AuthenticationClass: "shared",
			Oidc: &hbasev1alpha1.OidcSpec{
				ClientCredentialsSecret: testOIDCCredentials,
			},
		}
	})
	clusterB := clusterA.DeepCopy()
	clusterB.Name = "b"
	clusterB.Namespace = "other"
	clusterB.Spec.ClusterConfigSpec.ZookeeperConfigMapName = "other-znode"

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(clusterB, clusterA).
		WithIndex(&hbasev1alpha1.HbaseCluster{}, configMapReferenceField, configMapReferences).
		WithIndex(&hbasev1alpha1.HbaseCluster{}, secretReferenceField, secretReferences).
		WithIndex(&hbasev1alpha1.HbaseCluster{}, authenticationClassReferenceField, authenticationClassReferences).
		Build()

	configMapRequests := referenceRequests(k8sClient, configMapReferenceField, true)(
		context.Background(),
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: testZnodeCM, Namespace: testNamespace}},
	)
	if want := []reconcile.Request{{NamespacedName: client.ObjectKeyFromObject(clusterA)}}; !reflect.DeepEqual(configMapRequests, want) {
		t.Errorf("ConfigMap requests = %v, want %v", configMapRequests, want)
	}
	secretRequests := referenceRequests(k8sClient, secretReferenceField, true)(
		context.Background(),
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: testOIDCCredentials, Namespace: testNamespace}},
	)
	if want := []reconcile.Request{{NamespacedName: client.ObjectKeyFromObject(clusterA)}}; !reflect.DeepEqual(secretRequests, want) {
		t.Errorf("Secret requests = %v, want %v", secretRequests, want)
	}

	authClassRequests := referenceRequests(k8sClient, authenticationClassReferenceField, false)(
		context.Background(),
		&authv1alpha1.AuthenticationClass{ObjectMeta: metav1.ObjectMeta{Name: "shared"}},
	)
	want := []reconcile.Request{
		{NamespacedName: client.ObjectKeyFromObject(clusterA)},
		{NamespacedName: client.ObjectKeyFromObject(clusterB)},
	}
	if !reflect.DeepEqual(authClassRequests, want) {
		t.Errorf("AuthenticationClass requests = %v, want %v", authClassRequests, want)
	}
}

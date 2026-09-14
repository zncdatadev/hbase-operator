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
	"fmt"
	"slices"

	authv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/authentication/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	hbasev1alpha1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
)

const (
	configMapReferenceField           = "hbase.spec.clusterConfig.configMapReferences"
	secretReferenceField              = "hbase.spec.clusterConfig.secretReferences"
	authenticationClassReferenceField = "hbase.spec.clusterConfig.authenticationClassReference"
)

// SetupReferenceIndexes indexes the product-specific object references which GenericReconciler
// cannot discover from its generic spec. The matching watches below use these indexes to avoid
// listing every HbaseCluster for every ConfigMap, Secret or AuthenticationClass event.
func SetupReferenceIndexes(ctx context.Context, indexer client.FieldIndexer) error {
	indexes := []struct {
		field   string
		extract client.IndexerFunc
	}{
		{configMapReferenceField, configMapReferences},
		{secretReferenceField, secretReferences},
		{authenticationClassReferenceField, authenticationClassReferences},
	}
	for _, index := range indexes {
		if err := indexer.IndexField(ctx, &hbasev1alpha1.HbaseCluster{}, index.field, index.extract); err != nil {
			return fmt.Errorf("indexing HbaseCluster field %q: %w", index.field, err)
		}
	}
	return nil
}

// ReferenceWatches returns the product-specific watches plugged into
// reconciler.SetupWithManagerOptions. GenericReconciler owns the reconciliation; these mappings
// only tell controller-runtime which HbaseClusters refer to an external object that changed.
func ReferenceWatches(k8sClient client.Client) []func(*builder.Builder) *builder.Builder {
	return []func(*builder.Builder) *builder.Builder{
		func(b *builder.Builder) *builder.Builder {
			return b.Watches(
				&corev1.ConfigMap{},
				handler.EnqueueRequestsFromMapFunc(referenceRequests(k8sClient, configMapReferenceField, true)),
			)
		},
		func(b *builder.Builder) *builder.Builder {
			return b.Watches(
				&corev1.Secret{},
				handler.EnqueueRequestsFromMapFunc(referenceRequests(k8sClient, secretReferenceField, true)),
			)
		},
		func(b *builder.Builder) *builder.Builder {
			return b.Watches(
				&authv1alpha1.AuthenticationClass{},
				handler.EnqueueRequestsFromMapFunc(referenceRequests(
					k8sClient,
					authenticationClassReferenceField,
					false,
				)),
			)
		},
	}
}

func referenceRequests(
	k8sClient client.Client,
	field string,
	namespaced bool,
) handler.MapFunc {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		list := &hbasev1alpha1.HbaseClusterList{}
		opts := []client.ListOption{client.MatchingFields{field: obj.GetName()}}
		if namespaced {
			opts = append(opts, client.InNamespace(obj.GetNamespace()))
		}
		if err := k8sClient.List(ctx, list, opts...); err != nil {
			log.FromContext(ctx).Error(err, "unable to map external dependency to HbaseClusters",
				"kind", obj.GetObjectKind().GroupVersionKind().Kind,
				"namespace", obj.GetNamespace(),
				"name", obj.GetName(),
				"index", field)
			return nil
		}

		requests := make([]reconcile.Request, 0, len(list.Items))
		for i := range list.Items {
			requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&list.Items[i])})
		}
		slices.SortFunc(requests, func(a, b reconcile.Request) int {
			if byNamespace := compareStrings(a.Namespace, b.Namespace); byNamespace != 0 {
				return byNamespace
			}
			return compareStrings(a.Name, b.Name)
		})
		return requests
	}
}

func compareStrings(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func configMapReferences(obj client.Object) []string {
	cr, ok := obj.(*hbasev1alpha1.HbaseCluster)
	if !ok || cr.Spec.ClusterConfigSpec == nil {
		return nil
	}
	cfg := cr.Spec.ClusterConfigSpec
	return uniqueNonEmpty(
		cfg.ZookeeperConfigMapName,
		cfg.HdfsConfigMapName,
		cfg.VectorAggregatorConfigMapName,
	)
}

func secretReferences(obj client.Object) []string {
	cr, ok := obj.(*hbasev1alpha1.HbaseCluster)
	if !ok || cr.Spec.ClusterConfigSpec == nil || cr.Spec.ClusterConfigSpec.Authentication == nil ||
		cr.Spec.ClusterConfigSpec.Authentication.Oidc == nil {
		return nil
	}
	return uniqueNonEmpty(cr.Spec.ClusterConfigSpec.Authentication.Oidc.ClientCredentialsSecret)
}

func authenticationClassReferences(obj client.Object) []string {
	cr, ok := obj.(*hbasev1alpha1.HbaseCluster)
	if !ok || cr.Spec.ClusterConfigSpec == nil || cr.Spec.ClusterConfigSpec.Authentication == nil {
		return nil
	}
	return uniqueNonEmpty(cr.Spec.ClusterConfigSpec.Authentication.AuthenticationClass)
}

func uniqueNonEmpty(names ...string) []string {
	seen := make(map[string]struct{}, len(names))
	result := make([]string, 0, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	return result
}

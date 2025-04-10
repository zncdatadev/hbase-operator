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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
// var hbaseclusterlog = logf.Log.WithName("hbasecluster-resource")

func (r *HbaseCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-hbase-kubedoop-dev-v1alpha1-hbasecluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=hbase.kubedoop.dev,resources=hbaseclusters,verbs=create;update,versions=v1alpha1,name=mhbasecluster.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &HbaseCluster{}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:path=/validate-hbase-kubedoop-dev-v1alpha1-hbasecluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=hbase.kubedoop.dev,resources=hbaseclusters,verbs=create;update,versions=v1alpha1,name=vhbasecluster.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &HbaseCluster{}

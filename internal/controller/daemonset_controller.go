/*
Copyright 2026.

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
	"sync"

	"github.com/rjbrown57/aboutController/internal/common"
	"github.com/rjbrown57/aboutController/pkg/propertybuilder"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// DaemonsetReconciler reconciles a Daemonset object
type DaemonsetReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	common.AboutControllerCommon
	Mu sync.Mutex
}

// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=daemonsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=daemonsets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Daemonset object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.3/pkg/reconcile
func (r *DaemonsetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	logger := logf.FromContext(ctx)

	workload := &appsv1.DaemonSet{}
	if err := r.Get(ctx, req.NamespacedName, workload); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// check if we have a property for this deployment
	if _, exists := r.ManagedProperties[workload.GetUID()]; !exists {
		logger.Info("Detected", "Deployment", workload.Name)

		Properties := propertybuilder.PropertiesFromAnnotations(workload.ObjectMeta, common.WatchedPrefix)

		for _, prop := range Properties.Items {
			logger.Info("Adding clusterProperty", "name", prop.Name, "value", prop.Spec.Value)
			if err := r.Create(ctx, &prop); err != nil {
				logger.Error(err, "Failed to create Property for", "Deployment", workload.Name)
			}
		}

	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DaemonsetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.DaemonSet{}, builder.WithPredicates(common.AnnotationPredicate())).
		Named("daemonset").
		Complete(r)
}

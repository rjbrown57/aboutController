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
	"fmt"
	"sync"

	"github.com/rjbrown57/aboutController/internal/common"
	"github.com/rjbrown57/aboutController/pkg/propertybuilder"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	aboutapi "sigs.k8s.io/about-api/pkg/apis/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// DeploymentReconciler reconciles a Deployment object
type DeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	ManagedProperties map[string]*aboutapi.ClusterProperty
	Mu                sync.Mutex
}

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update

// +kubebuilder:rbac:groups=about.k8s.io,resources=ClusterProperty,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=about.k8s.io,resources=ClusterProperty/status,verbs=update;patch;delete
// +kubebuilder:rbac:groups=about.k8s.io,resources=ClusterProperty/finalizers,verbs=update

// Reconcile will trigger on any deployment that has opted in and create a clusterProperty
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.3/pkg/reconcile
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, deployment); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// check if we have a property for this deployment
	if _, exists := r.ManagedProperties[fmt.Sprintf("%s/%s", deployment.Namespace, deployment.Name)]; !exists {
		logger.Info("Detected", "Deployment", deployment.Name)

		Properties := propertybuilder.PropertiesFromAnnotations(deployment.ObjectMeta, common.WatchedPrefix)

		for _, prop := range Properties.Items {
			logger.Info("Adding clusterProperty", "name", prop.Name, "value", prop.Spec.Value)
			if err := r.Create(ctx, &prop); err != nil {
				logger.Error(err, "Failed to create Property for", "Deployment", deployment.Name)
			}
		}

	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}, builder.WithPredicates(common.AnnotationPredicate())).
		Named("deployment").
		Complete(r)
}

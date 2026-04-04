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

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	aboutapi "sigs.k8s.io/about-api/pkg/apis/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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

		// Create a clusterProperty
		clusterProperty := aboutapi.ClusterProperty{
			ObjectMeta: metav1.ObjectMeta{
				Name: deployment.Name,
			},
			Spec: aboutapi.ClusterPropertySpec{
				Value: "detected",
			},
		}

		if err := r.Create(ctx, &clusterProperty); err != nil {
			logger.Error(err, "Failed to create Property for", "Deployment", deployment.Name)
		}

	}

	return ctrl.Result{}, nil
}

const watchedAnnotation = "rjbrown57.io/track"

func hasWatchedAnnotation(d *appsv1.Deployment) bool {
	if d == nil {
		return false
	}
	_, ok := d.Annotations[watchedAnnotation]
	return ok
}

func deploymentAnnotationPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			d, ok := e.Object.(*appsv1.Deployment)
			return ok && hasWatchedAnnotation(d)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldD, ok1 := e.ObjectOld.(*appsv1.Deployment)
			newD, ok2 := e.ObjectNew.(*appsv1.Deployment)
			if !ok1 || !ok2 {
				return false
			}

			oldVal, oldOk := oldD.Annotations[watchedAnnotation]
			newVal, newOk := newD.Annotations[watchedAnnotation]

			return oldOk != newOk || oldVal != newVal
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			d, ok := e.Object.(*appsv1.Deployment)
			return ok && hasWatchedAnnotation(d)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}, builder.WithPredicates(deploymentAnnotationPredicate())).
		Named("deployment").
		Complete(r)
}

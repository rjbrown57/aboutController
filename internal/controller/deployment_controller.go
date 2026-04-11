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

	"github.com/rjbrown57/aboutController/internal/common"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeploymentReconciler reconciles a Deployment object
type DeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	common.ControllerCommon
}

func NewDeploymentReconcilier(mgr ctrl.Manager) (*DeploymentReconciler, error) {
	dr := &DeploymentReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		ControllerCommon: common.ControllerCommon{
			Recorder: mgr.GetEventRecorder("deployment-controller"),
		},
	}

	if err := dr.SetupWithManager(mgr); err != nil {
		return nil, err
	}

	return dr, nil
}

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update

// +kubebuilder:rbac:groups=about.k8s.io,resources=clusterproperties,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=about.k8s.io,resources=clusterproperties/status,verbs=update;patch;delete
// +kubebuilder:rbac:groups=about.k8s.io,resources=clusterproperties/finalizers,verbs=update
// +kubebuilder:rbac:groups=events.k8s.io,resources=events,verbs=create;patch

// Reconcile will trigger on any deployment that has opted in and create a clusterProperty
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.3/pkg/reconcile
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	wr := NewWorkloadReconciler(&WorkloadReconcilerOptions{
		Ctx:      ctx,
		C:        r.Client,
		Req:      req,
		Scheme:   r.Scheme,
		Workload: &appsv1.Deployment{},
		ER:       r.Recorder,
	})

	return wr.Reconcile()
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}, builder.WithPredicates(common.AnnotationPredicate())).
		Named("deployment").
		Complete(r)
}

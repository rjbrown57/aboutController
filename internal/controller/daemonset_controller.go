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

// DaemonsetReconciler reconciles a Daemonset object
type DaemonsetReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	common.ControllerCommon
}

func NewDaemonsetReconcilier(mgr ctrl.Manager) (*DaemonsetReconciler, error) {

	dr := &DaemonsetReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		ControllerCommon: common.ControllerCommon{
			Recorder: mgr.GetEventRecorder("daemonset-controller"),
		},
	}

	if err := dr.SetupWithManager(mgr); err != nil {
		return nil, err
	}

	return dr, nil
}

// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=daemonsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=daemonsets/finalizers,verbs=update
// +kubebuilder:rbac:groups=events.k8s.io,resources=events,verbs=create;patch

func (r *DaemonsetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	wr := NewWorkloadReconciler(&WorkloadReconcilerOptions{
		Ctx:      ctx,
		C:        r.Client,
		Req:      req,
		Scheme:   r.Scheme,
		Workload: &appsv1.DaemonSet{},
		ER:       r.Recorder,
	})

	return wr.Reconcile()
}

// SetupWithManager sets up the controller with the Manager.
func (r *DaemonsetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.DaemonSet{}, builder.WithPredicates(common.AnnotationPredicate())).
		Named("daemonset").
		Complete(r)
}

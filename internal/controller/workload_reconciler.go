package controller

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/rjbrown57/aboutController/internal/common"
	"github.com/rjbrown57/aboutController/pkg/propertybuilder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	aboutapi "sigs.k8s.io/about-api/pkg/apis/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type WorkloadReconcilerOptions struct {
	Ctx      context.Context
	C        client.Client
	Req      ctrl.Request
	Scheme   *runtime.Scheme
	Workload client.Object
}

type WorkloadReconciler struct {
	*WorkloadReconcilerOptions
}

// NewWorkloadReconciler builds a single-use reconciler for one workload reconcile request.
func NewWorkloadReconciler(opts *WorkloadReconcilerOptions) *WorkloadReconciler {
	return &WorkloadReconciler{WorkloadReconcilerOptions: opts}
}

// ProcessFinalizer will add a finalizer to the workload that has opt-ed in if missing
func (r *WorkloadReconciler) ProcessFinalizers() (ctrl.Result, error) {
	// Patch the finalizer onto the current workload so we do not overwrite unrelated changes
	// made to the workload between the initial Get and this reconciliation step.
	if !controllerutil.ContainsFinalizer(r.Workload, common.Finalizer) {
		before := r.Workload.DeepCopyObject().(client.Object)
		controllerutil.AddFinalizer(r.Workload, common.Finalizer)
		if err := r.C.Patch(r.Ctx, r.Workload, client.MergeFrom(before)); err != nil {
			return ctrl.Result{RequeueAfter: time.Millisecond * 5}, err
		}
		return ctrl.Result{RequeueAfter: time.Millisecond * 1}, nil
	}

	return ctrl.Result{}, nil
}

// Reconcile runs the shared workload reconciliation flow for one reconcile request.
func (r *WorkloadReconciler) Reconcile() (ctrl.Result, error) {
	if err := r.loadWorkload(); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r.logger().Info("Detected workload", "kind", r.kind(), "name", r.Workload.GetName())

	if r.isDeleting() {
		return r.handleDelete()
	}

	if result, err := r.ProcessFinalizers(); err != nil || !result.IsZero() {
		return result, err
	}

	return r.reconcileProperties()
}

// loadWorkload fetches the current workload instance addressed by the reconcile request.
func (r *WorkloadReconciler) loadWorkload() error {
	return r.C.Get(r.Ctx, r.Req.NamespacedName, r.Workload)
}

// isDeleting reports whether Kubernetes has started deleting the workload.
func (r *WorkloadReconciler) isDeleting() bool {
	return !r.Workload.GetDeletionTimestamp().IsZero()
}

// logger returns the request-scoped logger from the reconciliation context.
func (r *WorkloadReconciler) logger() logr.Logger {
	return logf.FromContext(r.Ctx)
}

// kind resolves the Kubernetes kind for the current workload for logging.
func (r *WorkloadReconciler) kind() string {
	return workloadKind(r.Workload, r.Scheme)
}

// listManagedProperties returns the ClusterProperty objects currently associated with this workload.
func (r *WorkloadReconciler) listManagedProperties() (*aboutapi.ClusterPropertyList, error) {
	propList := &aboutapi.ClusterPropertyList{}
	if err := r.C.List(r.Ctx, propList, client.MatchingLabels{
		common.OwnerLabel: string(r.Workload.GetUID()),
	}); client.IgnoreNotFound(err) != nil {
		return nil, err
	}
	return propList, nil
}

// desiredProperties builds the desired ClusterProperty set from the workload annotations.
func (r *WorkloadReconciler) desiredProperties() aboutapi.ClusterPropertyList {
	return propertybuilder.PropertiesFromAnnotations(metav1.ObjectMeta{
		Annotations: r.Workload.GetAnnotations(),
		UID:         r.Workload.GetUID(),
	}, common.WatchedPrefix, common.OwnerLabel)
}

// reconcileProperties applies creates, updates, and deletions so managed properties match the workload annotations.
func (r *WorkloadReconciler) reconcileProperties() (ctrl.Result, error) {
	propList, err := r.listManagedProperties()
	if err != nil {
		return ctrl.Result{}, err
	}

	existingProperties := make(map[string]*aboutapi.ClusterProperty, len(propList.Items))
	for _, prop := range propList.Items {
		propCopy := prop
		existingProperties[prop.Name] = &propCopy
	}

	properties := r.desiredProperties()
	selectedProperties := make(map[string]string, len(properties.Items))
	for _, prop := range properties.Items {
		selectedProperties[prop.Name] = prop.Spec.Value
	}

	for _, prop := range properties.Items {
		existingProp, exists := existingProperties[prop.Name]
		if exists && existingProp.Spec.Value == prop.Spec.Value {
			continue
		}

		if exists {
			before := existingProp.DeepCopy()
			existingProp.Spec.Value = prop.Spec.Value
			if err := r.C.Patch(r.Ctx, existingProp, client.MergeFrom(before)); err != nil {
				r.logger().Error(err, "Failed to update Property for", "kind", r.kind(), "workload", r.Workload.GetName())
			}
			continue
		}

		r.logger().Info("Adding clusterProperty", "kind", r.kind(), "workload", r.Workload.GetName(), "name", prop.Name, "value", prop.Spec.Value)
		if err := r.C.Create(r.Ctx, &prop); err != nil {
			r.logger().Error(err, "Failed to create Property for", "kind", r.kind(), "workload", r.Workload.GetName())
		}
	}

	return r.deleteStaleProperties(propList, selectedProperties)
}

// deleteStaleProperties removes previously managed properties that are no longer desired.
func (r *WorkloadReconciler) deleteStaleProperties(
	propList *aboutapi.ClusterPropertyList,
	selectedProperties map[string]string,
) (ctrl.Result, error) {
	for _, prop := range propList.Items {
		if _, selected := selectedProperties[prop.Name]; selected {
			continue
		}

		r.logger().Info("Removing clusterProperty", "kind", r.kind(), "workload", r.Workload.GetName(), "name", prop.Name)
		if err := r.C.Delete(r.Ctx, &prop); client.IgnoreNotFound(err) != nil {
			r.logger().Error(err, "Failed to delete Property for", "kind", r.kind(), "workload", r.Workload.GetName(), "name", prop.Name)
		}
	}

	return ctrl.Result{}, nil
}

// handleDelete cleans up managed properties before allowing the workload finalizer to be removed.
func (r *WorkloadReconciler) handleDelete() (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(r.Workload, common.Finalizer) {
		return ctrl.Result{}, nil
	}

	propList := &aboutapi.ClusterPropertyList{}
	if err := r.C.List(r.Ctx, propList, client.MatchingLabels{
		common.OwnerLabel: string(r.Workload.GetUID()),
	}); client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	}

	// Delete all properties linked to this workload before allowing the workload to disappear.
	for _, prop := range propList.Items {
		r.logger().Info("Removing clusterProperty for deleted workload", "kind", r.kind(), "workload", r.Workload.GetName(), "name", prop.Name)
		if err := r.C.Delete(r.Ctx, &prop); client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
	}

	// Patch away the finalizer after cleanup so we only send the finalizer change back to the API
	// server and avoid clobbering unrelated workload updates during deletion.
	before := r.Workload.DeepCopyObject().(client.Object)
	controllerutil.RemoveFinalizer(r.Workload, common.Finalizer)
	if err := r.C.Patch(r.Ctx, r.Workload, client.MergeFrom(before)); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// workloadKind resolves a stable Kubernetes kind name for logging.
func workloadKind(workload client.Object, scheme *runtime.Scheme) string {
	if scheme == nil {
		return "workload"
	}

	gvk, err := apiutil.GVKForObject(workload, scheme)
	if err != nil || gvk.Kind == "" {
		return "workload"
	}

	return gvk.Kind
}

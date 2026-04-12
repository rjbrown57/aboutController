package common

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	aboutapi "sigs.k8s.io/about-api/pkg/apis/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func ClusterPropertyToWorkload(expectedKind string) handler.MapFunc {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		_ = ctx

		prop, ok := obj.(*aboutapi.ClusterProperty)
		if !ok {
			return nil
		}

		labels := prop.GetLabels()
		if labels[WorkloadKindLabel] != expectedKind {
			return nil
		}

		namespace := labels[WorkloadNamespaceLabel]
		name := labels[WorkloadNameLabel]
		if namespace == "" || name == "" {
			return nil
		}

		return []reconcile.Request{{
			NamespacedName: types.NamespacedName{
				Namespace: namespace,
				Name:      name,
			},
		}}
	}
}

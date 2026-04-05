package controller

import (
	"maps"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const watchedPrefix = "aboutcontroller.io/"

// hasWatchedAnnotation will find out target annotation
func hasWatchedAnnotation(annotations map[string]string) bool {
	if annotations == nil {
		return false
	}

	for annotation := range maps.Keys(annotations) {
		if strings.HasPrefix(annotation, watchedPrefix) {
			return true
		}
	}
	return false
}

func deploymentAnnotationPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			d, ok := e.Object.(*appsv1.Deployment)
			return ok && hasWatchedAnnotation(d.GetAnnotations())
		},
		// stub
		UpdateFunc: func(e event.UpdateEvent) bool {
			newD, ok2 := e.ObjectNew.(*appsv1.Deployment)
			if !ok2 {
				return false
			}

			newOk := hasWatchedAnnotation(newD.GetAnnotations())

			return newOk
		},
		// stub
		DeleteFunc: func(e event.DeleteEvent) bool {
			d, ok := e.Object.(*appsv1.Deployment)
			return ok && hasWatchedAnnotation(d.GetAnnotations())
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

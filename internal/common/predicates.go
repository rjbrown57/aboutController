package common

import (
	"maps"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// HasWatchedAnnotation will find out target annotation
func HasWatchedAnnotation(annotations map[string]string) bool {
	if annotations == nil {
		return false
	}

	for annotation := range maps.Keys(annotations) {
		if strings.HasPrefix(annotation, WatchedPrefix) {
			return true
		}
	}
	return false
}

func AnnotationPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return e.Object != nil && HasWatchedAnnotation(e.Object.GetAnnotations())
		},
		UpdateFunc: func(e event.UpdateEvent) bool {

			// if annotations are only present in old we need to clean up properties
			// if annotations are only present in new ew need to add properties
			return e.ObjectNew != nil &&
				HasWatchedAnnotation(e.ObjectOld.GetAnnotations()) ||
				HasWatchedAnnotation(e.ObjectNew.GetAnnotations())
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return e.Object != nil && HasWatchedAnnotation(e.Object.GetAnnotations())
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

package common

import (
	"maps"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// hasWatchedAnnotation will find out target annotation
func hasWatchedAnnotation(annotations map[string]string) bool {
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
			return e.Object != nil && hasWatchedAnnotation(e.Object.GetAnnotations())
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			//todo: refactor here because a certain diff indicates removal of opt-in annotation
			return e.ObjectNew != nil && hasWatchedAnnotation(e.ObjectNew.GetAnnotations())
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return e.Object != nil && hasWatchedAnnotation(e.Object.GetAnnotations())
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

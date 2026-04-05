package common

import (
	"maps"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var WatchedPrefix = "aboutcontroller.io/"

// init will allow us to override watchedPrefix via envVar if set
func init() {
	if s, ok := os.LookupEnv("aboutPrefix"); ok {
		// Normalize input so validation accepts either "example.io" or "example.io/".
		s = strings.TrimSuffix(s, "/")
		// Only accept prefixes that are valid Kubernetes-style annotation name prefixes.
		if len(validation.IsDNS1123Subdomain(s)) == 0 {
			// Keep the trailing slash so downstream prefix checks stay simple.
			WatchedPrefix = s + "/"
		}
	}
}

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
			//todo: refactor here because a certain diff indicates removal
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

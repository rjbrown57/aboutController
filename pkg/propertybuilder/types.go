package propertybuilder

import (
	"maps"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	aboutapi "sigs.k8s.io/about-api/pkg/apis/v1alpha1"
)

const watchedPrefix = "aboutcontroller.io"

// hasWatchedAnnotation will find out target annotation
func HasWatchedAnnotation(annotations map[string]string) bool {
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

func NewClusterProperty(name, value string) *aboutapi.ClusterProperty {
	prop := &aboutapi.ClusterProperty{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: aboutapi.ClusterPropertySpec{
			Value: value,
		},
	}

	return prop
}

func PropertiesFromAnnotations(annotations map[string]string, prefix string) aboutapi.ClusterPropertyList {
	list := aboutapi.ClusterPropertyList{
		Items: make([]aboutapi.ClusterProperty, 0),
	}

	for annotationKey, annotationValue := range annotations {
		if strings.HasPrefix(annotationKey, prefix) {
			list.Items = append(list.Items, *NewClusterProperty(strings.TrimPrefix(annotationKey, prefix), annotationValue))
		}
	}

	return list
}

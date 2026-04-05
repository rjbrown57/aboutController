package propertybuilder

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	aboutapi "sigs.k8s.io/about-api/pkg/apis/v1alpha1"
)

// hasWatchedAnnotation will find out target annotation
func HasWatchedAnnotation(annotations map[string]string, prefix string) bool {
	if annotations == nil {
		return false
	}

	for annotation := range annotations {
		if strings.HasPrefix(annotation, prefix) {
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

func PropertiesFromAnnotations(obj metav1.ObjectMeta, prefix string, ownerlabel string) aboutapi.ClusterPropertyList {
	list := aboutapi.ClusterPropertyList{
		Items: make([]aboutapi.ClusterProperty, 0),
	}

	for annotationKey, annotationValue := range obj.GetAnnotations() {
		if s, found := strings.CutPrefix(annotationKey, prefix); found {
			n := NewClusterProperty(s, annotationValue)

			n.Labels = map[string]string{
				ownerlabel: string(obj.GetUID()),
			}

			list.Items = append(list.Items, *n)
		}
	}

	return list
}

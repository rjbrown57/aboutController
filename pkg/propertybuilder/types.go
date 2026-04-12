package propertybuilder

import (
	"strings"

	"github.com/rjbrown57/aboutController/internal/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	aboutapi "sigs.k8s.io/about-api/pkg/apis/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func PropertiesFromAnnotations(obj client.Object, prefix string, ownerlabel string) aboutapi.ClusterPropertyList {
	list := aboutapi.ClusterPropertyList{
		Items: make([]aboutapi.ClusterProperty, 0),
	}

	for annotationKey, annotationValue := range obj.GetAnnotations() {
		if s, found := strings.CutPrefix(annotationKey, prefix); found {
			n := NewClusterProperty(s, annotationValue)

			n.Labels = GetPropLabels(obj)

			list.Items = append(list.Items, *n)
		}
	}

	return list
}

func GetPropLabels(obj client.Object) map[string]string {
	return map[string]string{
		common.WorkloadNamespaceLabel: obj.GetNamespace(),
		common.WorkloadNameLabel:      obj.GetName(),
		common.WorkloadKindLabel:      obj.GetObjectKind().GroupVersionKind().Kind,
	}
}

package propertybuilder

import (
	"fmt"
	"testing"

	"github.com/rjbrown57/aboutController/internal/common"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var testLabelPrefix = "aboutcontroller.io"
var testAnnotationPrefix = fmt.Sprintf("%s/", testLabelPrefix)

func TestHasWatchedAnnotation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		annotations map[string]string
		prefix      string
		want        bool
	}{
		{
			name:        "nil annotations",
			annotations: nil,
			prefix:      testAnnotationPrefix,
			want:        false,
		},
		{
			name: "no matching annotations",
			annotations: map[string]string{
				"example.com/version": "v1.0.0",
			},
			prefix: testAnnotationPrefix,
			want:   false,
		},
		{
			name: "matching annotation present",
			annotations: map[string]string{
				"example.com/version":                          "v1.0.0",
				fmt.Sprintf("%smy-prop", testAnnotationPrefix): "v2.0.0",
			},
			prefix: testAnnotationPrefix,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := HasWatchedAnnotation(tt.annotations, tt.prefix); got != tt.want {
				t.Fatalf("HasWatchedAnnotation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPropLabels(t *testing.T) {
	t.Parallel()

	obj := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "example-deployment",
		},
	}

	got := GetPropLabels(obj)

	want := map[string]string{
		common.WorkloadNamespaceLabel: "default",
		common.WorkloadNameLabel:      "example-deployment",
		common.WorkloadKindLabel:      "Deployment",
	}

	if len(got) != len(want) {
		t.Fatalf("GetPropLabels() returned %d labels, want %d: %#v", len(got), len(want), got)
	}

	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Fatalf("GetPropLabels()[%q] = %q, want %q", key, got[key], wantValue)
		}
	}
}

func TestPropertiesFromAnnotations(t *testing.T) {
	t.Parallel()

	obj := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "example-statefulset",
			Annotations: map[string]string{
				fmt.Sprintf("%sstatefulset-name", testAnnotationPrefix):    "example-db",
				fmt.Sprintf("%sstatefulset-version", testAnnotationPrefix): "v4.5.6",
				"example.com/ignored": "ignore-me",
			},
		},
	}

	got := PropertiesFromAnnotations(obj, testAnnotationPrefix, "unused-owner-label")

	if len(got.Items) != 2 {
		t.Fatalf("PropertiesFromAnnotations() returned %d properties, want 2", len(got.Items))
	}

	gotByName := make(map[string]string, len(got.Items))
	for _, item := range got.Items {
		namespaceLabel := item.Labels[common.WorkloadNamespaceLabel]
		if namespaceLabel != "default" {
			t.Fatalf(
				"property %q namespace label = %q, want %q",
				item.Name,
				namespaceLabel,
				"default",
			)
		}
		nameLabel := item.Labels[common.WorkloadNameLabel]
		if nameLabel != "example-statefulset" {
			t.Fatalf(
				"property %q name label = %q, want %q",
				item.Name,
				nameLabel,
				"example-statefulset",
			)
		}
		kindLabel := item.Labels[common.WorkloadKindLabel]
		if kindLabel != "StatefulSet" {
			t.Fatalf("property %q kind label = %q, want %q", item.Name, kindLabel, "StatefulSet")
		}

		gotByName[item.Name] = item.Spec.Value
	}

	wantByName := map[string]string{
		"statefulset-name":    "example-db",
		"statefulset-version": "v4.5.6",
	}

	if len(gotByName) != len(wantByName) {
		t.Fatalf("PropertiesFromAnnotations() produced %#v, want %#v", gotByName, wantByName)
	}

	for name, wantValue := range wantByName {
		if gotByName[name] != wantValue {
			t.Fatalf("property %q value = %q, want %q", name, gotByName[name], wantValue)
		}
	}
}

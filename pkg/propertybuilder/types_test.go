package propertybuilder

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
			prefix:      "aboutcontroller.io/",
			want:        false,
		},
		{
			name: "no matching annotations",
			annotations: map[string]string{
				"example.com/version": "v1.0.0",
			},
			prefix: "aboutcontroller.io/",
			want:   false,
		},
		{
			name: "matching annotation present",
			annotations: map[string]string{
				"example.com/version":         "v1.0.0",
				"aboutcontroller.io/my-prop": "v2.0.0",
			},
			prefix: "aboutcontroller.io/",
			want:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
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
		"aboutcontroller.io/namespace": "default",
		"aboutcontroller.io/name":      "example-deployment",
		"aboutcontroller.io/kind":      "Deployment",
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
				"aboutcontroller.io/statefulset-name":    "example-db",
				"aboutcontroller.io/statefulset-version": "v4.5.6",
				"example.com/ignored":                    "ignore-me",
			},
		},
	}

	got := PropertiesFromAnnotations(obj, "aboutcontroller.io/", "unused-owner-label")

	if len(got.Items) != 2 {
		t.Fatalf("PropertiesFromAnnotations() returned %d properties, want 2", len(got.Items))
	}

	gotByName := make(map[string]string, len(got.Items))
	for _, item := range got.Items {
		if item.Labels["aboutcontroller.io/namespace"] != "default" {
			t.Fatalf("property %q namespace label = %q, want %q", item.Name, item.Labels["aboutcontroller.io/namespace"], "default")
		}
		if item.Labels["aboutcontroller.io/name"] != "example-statefulset" {
			t.Fatalf("property %q name label = %q, want %q", item.Name, item.Labels["aboutcontroller.io/name"], "example-statefulset")
		}
		if item.Labels["aboutcontroller.io/kind"] != "StatefulSet" {
			t.Fatalf("property %q kind label = %q, want %q", item.Name, item.Labels["aboutcontroller.io/kind"], "StatefulSet")
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

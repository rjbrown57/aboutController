package common

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	aboutapi "sigs.k8s.io/about-api/pkg/apis/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestClusterPropertyToWorkload(t *testing.T) {
	t.Parallel()

	mapFunc := ClusterPropertyToWorkload("Deployment")

	tests := []struct {
		name string
		obj  client.Object
		want []types.NamespacedName
	}{
		{
			name: "matching property routes request",
			obj: &aboutapi.ClusterProperty{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						WorkloadNamespaceLabel: "default",
						WorkloadNameLabel:      "example-deployment",
						WorkloadKindLabel:      "Deployment",
					},
				},
			},
			want: []types.NamespacedName{{
				Namespace: "default",
				Name:      "example-deployment",
			}},
		},
		{
			name: "wrong kind returns no requests",
			obj: &aboutapi.ClusterProperty{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						WorkloadNamespaceLabel: "default",
						WorkloadNameLabel:      "example-deployment",
						WorkloadKindLabel:      "StatefulSet",
					},
				},
			},
		},
		{
			name: "missing namespace returns no requests",
			obj: &aboutapi.ClusterProperty{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						WorkloadNameLabel: "example-deployment",
						WorkloadKindLabel: "Deployment",
					},
				},
			},
		},
		{
			name: "missing name returns no requests",
			obj: &aboutapi.ClusterProperty{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						WorkloadNamespaceLabel: "default",
						WorkloadKindLabel:      "Deployment",
					},
				},
			},
		},
		{
			name: "non cluster property returns no requests",
			obj: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "example-deployment",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			requests := mapFunc(context.Background(), tt.obj)

			if len(requests) != len(tt.want) {
				t.Fatalf("got %d requests, want %d", len(requests), len(tt.want))
			}

			for i, want := range tt.want {
				if requests[i].NamespacedName != want {
					t.Fatalf("request %d = %#v, want %#v", i, requests[i].NamespacedName, want)
				}
			}
		})
	}
}

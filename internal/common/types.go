package common

import (
	"k8s.io/apimachinery/pkg/types"
	aboutapi "sigs.k8s.io/about-api/pkg/apis/v1alpha1"
)

type AboutControllerCommon struct {
	ManagedProperties map[types.UID][]*aboutapi.ClusterProperty
}

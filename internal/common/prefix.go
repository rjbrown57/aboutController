package common

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

// We need the trailing / here so we can extract our propertyNames cleanly.
var WatchedPrefix = "aboutcontroller.io/"
var OwnerLabel string
var Finalizer string

// init will allow us to override watchedPrefix via envVar if set.
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

	OwnerLabel = fmt.Sprintf("%s%s", WatchedPrefix, "owner")
	Finalizer = fmt.Sprintf("%s%s", WatchedPrefix, "finalizer")
}

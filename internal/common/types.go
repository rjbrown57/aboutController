package common

import (
	"k8s.io/client-go/tools/events"
)

// ControllerCommon is embedded into controllers for common features
type ControllerCommon struct {
	Recorder events.EventRecorder
}

package types

import "time"

type Event struct {
	Timestamp time.Time         `json:"timestamp"`
	Kind      EventKind         `json:"kind"`
	Namespace string            `json:"namespace"`
	Source    string            `json:"source"`
	Message   string            `json:"message"`
	Details   map[string]string `json:"details,omitempty"`
}

type EventKind string

const (
	PodCreated    EventKind = "PodCreated"
	PodReady      EventKind = "PodReady"
	PodUnready    EventKind = "PodUnready"
	PodTerminated EventKind = "PodTerminated"

	PodOOMKilled     EventKind = "PodOOMKilled"
	PodEvicted       EventKind = "PodEvicted"
	CrashLoopBackOff EventKind = "CrashLoopBackOff"
	ImagePullBackOff EventKind = "ImagePullBackOff"

	ProbeFailed EventKind = "ProbeFailed"

	DeploymentUpdate EventKind = "DeploymentUpdate"
	ReplicaSetScaled EventKind = "ReplicaSetScaled"

	EventWarning EventKind = "EventWarning"
	EventNormal  EventKind = "EventNormal"
)

type Timeline []Event

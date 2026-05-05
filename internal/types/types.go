// Package types defines shared data structures for kube-tale.
package types

import "time"

// Event is a single entry in a timeline.
type Event struct {
	Timestamp time.Time         `json:"timestamp"`
	Kind      EventKind         `json:"kind"`
	Namespace string            `json:"namespace"`
	Source    string            `json:"source"`
	Message   string            `json:"message"`
	Details   map[string]string `json:"details,omitempty"`
}

// EventKind classifies the type of event.
type EventKind string

// Event kind constants for pod lifecycle, failures, probes, and workload changes.
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

// Timeline is a sorted, deduplicated slice of Events (oldest first).
type Timeline []Event

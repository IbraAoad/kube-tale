package client

import (
	"context"
	"fmt"
	"time"

	"github.com/IbraAoad/kube-tale/internal/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// RealClient implements DataSource against a live Kubernetes cluster.
type RealClient struct {
	clientset kubernetes.Interface
}

// NewRealClient creates a RealClient from a kubeconfig path.
// Pass an empty string for in-cluster configuration.
func NewRealClient(kubeconfig string) (*RealClient, error) {
	cfg, err := restConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("build rest config: %w", err)
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create clientset: %w", err)
	}
	return &RealClient{clientset: cs}, nil
}

func restConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

// Events fetches core/v1 Events for a namespace within the given time range.
func (r *RealClient) Events(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	list, err := r.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("events: list: %w", err)
	}

	var events []types.Event
	for _, e := range list.Items {
		ts := e.LastTimestamp.Time
		if ts.IsZero() {
			ts = e.EventTime.Time
		}
		if ts.Before(since) || (!until.IsZero() && ts.After(until)) {
			continue
		}
		events = append(events, types.Event{
			Timestamp: ts,
			Kind:      mapEventKind(e.Type, e.Reason, e.Message),
			Namespace: e.Namespace,
			Source:    fmt.Sprintf("%s/%s", e.InvolvedObject.Kind, e.InvolvedObject.Name),
			Message:   e.Message,
		})
	}
	return events, nil
}

// PodHistory derives pod lifecycle events from pod and container statuses.
func (r *RealClient) PodHistory(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pods, err := r.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("podhistory: list: %w", err)
	}

	var events []types.Event
	for _, pod := range pods.Items {
		if pod.CreationTimestamp.Time.Before(since) {
			continue
		}
		events = append(events, podEvents(&pod)...)
	}
	return events, nil
}

func podEvents(pod *corev1.Pod) []types.Event {
	var events []types.Event
	source := fmt.Sprintf("pod/%s", pod.Name)

	created := types.Event{
		Timestamp: pod.CreationTimestamp.Time,
		Kind:      types.PodCreated,
		Namespace: pod.Namespace,
		Source:    source,
		Message:   fmt.Sprintf("Pod %s created", pod.Name),
	}
	events = append(events, created)

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Running != nil && cs.Ready {
			events = append(events, types.Event{
				Timestamp: cs.State.Running.StartedAt.Time,
				Kind:      types.PodReady,
				Namespace: pod.Namespace,
				Source:    source,
				Message:   fmt.Sprintf("Pod %s became ready", pod.Name),
			})
		}

		if cs.State.Waiting != nil {
			ts := time.Now()
			if cs.State.Running != nil {
				ts = cs.State.Running.StartedAt.Time
			}
			reason := cs.State.Waiting.Reason
			switch reason {
			case "CrashLoopBackOff":
				events = append(events, types.Event{
					Timestamp: ts,
					Kind:      types.CrashLoopBackOff,
					Namespace: pod.Namespace,
					Source:    source,
					Message:   fmt.Sprintf("Pod %s in CrashLoopBackOff: %s", pod.Name, cs.State.Waiting.Message),
				})
			case "ErrImagePull", "ImagePullBackOff":
				events = append(events, types.Event{
					Timestamp: ts,
					Kind:      types.ImagePullBackOff,
					Namespace: pod.Namespace,
					Source:    source,
					Message:   fmt.Sprintf("Pod %s image pull failed: %s", pod.Name, cs.State.Waiting.Message),
				})
			}
		}

		lastTerm := cs.LastTerminationState.Terminated
		if lastTerm != nil && lastTerm.Reason == "OOMKilled" {
			events = append(events, types.Event{
				Timestamp: lastTerm.FinishedAt.Time,
				Kind:      types.PodOOMKilled,
				Namespace: pod.Namespace,
				Source:    source,
				Message:   fmt.Sprintf("Pod %s was OOMKilled", pod.Name),
			})
		}
	}

	return events
}
// ReplicaSetHistory derives deployment update events from Deployments and their ReplicaSets.
func (r *RealClient) ReplicaSetHistory(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	deployments, err := r.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("replicasethistory: list deployments: %w", err)
	}

	var events []types.Event
	for _, dep := range deployments.Items {
		for _, cond := range dep.Status.Conditions {
			if cond.Type == "Available" || cond.Type == "Progressing" {
				ts := cond.LastUpdateTime.Time
				if ts.Before(since) || (!until.IsZero() && ts.After(until)) {
					continue
				}
				msg := fmt.Sprintf("Deployment %s: %s=%s — %s", dep.Name, cond.Type, cond.Status, cond.Message)
				events = append(events, types.Event{
					Timestamp: ts,
					Kind:      types.DeploymentUpdate,
					Namespace: dep.Namespace,
					Source:    fmt.Sprintf("deployment/%s", dep.Name),
					Message:   msg,
					Details: map[string]string{
						"condition": string(cond.Type),
						"status":    string(cond.Status),
					},
				})
			}
		}
	}
	return events, nil
}

func mapEventKind(eventType, reason, message string) types.EventKind {
	_ = message
	if eventType == "Normal" {
		return types.EventNormal
	}

	reasonMap := map[string]types.EventKind{
		"BackOff":             types.CrashLoopBackOff,
		"CrashLoopBackOff":    types.CrashLoopBackOff,
		"ErrImagePull":        types.ImagePullBackOff,
		"ImagePullBackOff":    types.ImagePullBackOff,
		"OOMKilling":          types.PodOOMKilled,
		"OOMKilled":           types.PodOOMKilled,
		"Unhealthy":           types.ProbeFailed,
		"Evicted":             types.PodEvicted,
		"FailedScheduling":    types.EventWarning,
		"FailedMount":         types.EventWarning,
		"FailedCreatePodSandBox": types.EventWarning,
	}

	if kind, ok := reasonMap[reason]; ok {
		return kind
	}

	if eventType == "Warning" {
		return types.EventWarning
	}
	return types.EventNormal
}

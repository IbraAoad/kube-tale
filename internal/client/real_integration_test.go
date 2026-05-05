//go:build integration

package client

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/IbraAoad/kube-tale/internal/types"
)

func kubeconfigPath() string {
	if p := os.Getenv("KUBECONFIG"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	path := home + "/.kube/config"
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

func newClientset(t *testing.T) *kubernetes.Clientset {
	t.Helper()
	path := kubeconfigPath()
	if path == "" {
		t.Skip("no kubeconfig found, skipping integration test")
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		t.Fatalf("build config: %v", err)
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("new clientset: %v", err)
	}
	return cs
}

func createTestNamespace(t *testing.T, cs kubernetes.Interface) string {
	ns := fmt.Sprintf("kube-tale-test-%d", time.Now().UnixNano())
	_, err := cs.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: ns},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create namespace %s: %v", ns, err)
	}
	t.Cleanup(func() {
		_ = cs.CoreV1().Namespaces().Delete(context.Background(), ns, metav1.DeleteOptions{})
	})
	return ns
}

func deployNginx(t *testing.T, cs kubernetes.Interface, ns, name string) {
	t.Helper()
	replicas := int32(1)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "nginx",
						Image: "nginx:alpine",
					}},
				},
			},
		},
	}
	_, err := cs.AppsV1().Deployments(ns).Create(context.Background(), dep, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create deployment %s: %v", name, err)
	}
	waitForPodsRunning(t, cs, ns, 45*time.Second)
	waitForContainersReady(t, cs, ns, 30*time.Second)
}

func deployCrashLoop(t *testing.T, cs kubernetes.Interface, ns, name string) {
	t.Helper()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{{
				Name:    "crash",
				Image:   "busybox",
				Command: []string{"sh", "-c", "exit 1"},
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("32Mi"),
					},
				},
			}},
		},
	}
	_, err := cs.CoreV1().Pods(ns).Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create pod %s: %v", name, err)
	}
}

func waitForPodsRunning(t *testing.T, cs kubernetes.Interface, ns string, timeout time.Duration) {
	t.Helper()
	deadline := time.After(timeout)
	for {
		select {
		case <-deadline:
			t.Fatal("timeout waiting for pods to be Running")
		default:
		}
		pods, err := cs.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			t.Fatalf("list pods: %v", err)
		}
		if len(pods.Items) == 0 {
			time.Sleep(2 * time.Second)
			continue
		}
		allRunning := true
		for _, pod := range pods.Items {
			if pod.Status.Phase != corev1.PodRunning {
				allRunning = false
				break
			}
		}
		if allRunning {
			return
		}
		time.Sleep(2 * time.Second)
	}
}

func waitForContainersReady(t *testing.T, cs kubernetes.Interface, ns string, timeout time.Duration) {
	t.Helper()
	deadline := time.After(timeout)
	for {
		select {
		case <-deadline:
			t.Fatal("timeout waiting for containers to be Ready")
		default:
		}
		pods, err := cs.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			t.Fatalf("list pods: %v", err)
		}
		allContainerReady := true
		for _, pod := range pods.Items {
			for _, cs := range pod.Status.ContainerStatuses {
				if !cs.Ready {
					allContainerReady = false
					break
				}
			}
			if !allContainerReady {
				break
			}
		}
		if allContainerReady {
			return
		}
		time.Sleep(2 * time.Second)
	}
}

func TestRealClient_PodHistory_SuccessfulStartup(t *testing.T) {
	cs := newClientset(t)
	ns := createTestNamespace(t, cs)
	deployNginx(t, cs, ns, "nginx-test")

	rc, err := NewRealClient(kubeconfigPath())
	if err != nil {
		t.Fatalf("new real client: %v", err)
	}

	events, err := rc.PodHistory(context.Background(), ns, time.Now().Add(-5*time.Minute), time.Time{})
	if err != nil {
		t.Fatalf("pod history: %v", err)
	}

	hasCreated := false
	hasReady := false
	for _, e := range events {
		if e.Kind == types.PodCreated {
			hasCreated = true
		}
		if e.Kind == types.PodReady {
			hasReady = true
		}
	}
	if !hasCreated {
		t.Error("expected PodCreated event, none found")
	}
	if !hasReady {
		t.Error("expected PodReady event, none found")
	}
}

func TestRealClient_PodHistory_CrashLoop(t *testing.T) {
	cs := newClientset(t)
	ns := createTestNamespace(t, cs)
	deployCrashLoop(t, cs, ns, "crash-test")

	time.Sleep(10 * time.Second)

	rc, err := NewRealClient(kubeconfigPath())
	if err != nil {
		t.Fatalf("new real client: %v", err)
	}

	events, err := rc.PodHistory(context.Background(), ns, time.Now().Add(-5*time.Minute), time.Time{})
	if err != nil {
		t.Fatalf("pod history: %v", err)
	}

	hasCreated := false
	for _, e := range events {
		if e.Kind == types.PodCreated {
			hasCreated = true
		}
	}
	if !hasCreated {
		t.Error("expected at least PodCreated event, none found")
	}
	if len(events) == 0 {
		t.Error("expected pod history events, got none")
	}
}

func TestRealClient_ReplicaSetHistory_DeploymentUpdate(t *testing.T) {
	cs := newClientset(t)
	ns := createTestNamespace(t, cs)
	deployNginx(t, cs, ns, "deploy-test")

	rc, err := NewRealClient(kubeconfigPath())
	if err != nil {
		t.Fatalf("new real client: %v", err)
	}

	events, err := rc.ReplicaSetHistory(context.Background(), ns, time.Now().Add(-5*time.Minute), time.Time{})
	if err != nil {
		t.Fatalf("replicaset history: %v", err)
	}

	if len(events) == 0 {
		t.Error("expected deployment update events, got none")
	}
	for _, e := range events {
		if e.Kind != types.DeploymentUpdate {
			t.Errorf("expected DeploymentUpdate, got %s", e.Kind)
		}
	}
}

func TestRealClient_Events_FetchAndFilter(t *testing.T) {
	cs := newClientset(t)
	ns := createTestNamespace(t, cs)
	deployNginx(t, cs, ns, "events-test")

	rc, err := NewRealClient(kubeconfigPath())
	if err != nil {
		t.Fatalf("new real client: %v", err)
	}

	events, err := rc.Events(context.Background(), ns, time.Now().Add(-5*time.Minute), time.Time{})
	if err != nil {
		t.Fatalf("events: %v", err)
	}

	if len(events) == 0 {
		t.Error("expected events, got none")
	}
	for _, e := range events {
		if e.Namespace != ns {
			t.Errorf("expected namespace %s, got %s", ns, e.Namespace)
		}
		if e.Source == "" {
			t.Error("expected non-empty Source")
		}
	}
}

func TestRealClient_EmptyNamespace(t *testing.T) {
	rc, err := NewRealClient(kubeconfigPath())
	if err != nil {
		t.Fatalf("new real client: %v", err)
	}

	ns := fmt.Sprintf("nonexistent-ns-%d", time.Now().UnixNano())
	since := time.Now().Add(-1 * time.Hour)

	t.Run("Events", func(t *testing.T) {
		events, err := rc.Events(context.Background(), ns, since, time.Time{})
		if err != nil && !strings.Contains(err.Error(), "not found") {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 0 {
			t.Errorf("expected 0 events, got %d", len(events))
		}
	})
	t.Run("PodHistory", func(t *testing.T) {
		events, err := rc.PodHistory(context.Background(), ns, since, time.Time{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 0 {
			t.Errorf("expected 0 events, got %d", len(events))
		}
	})
	t.Run("ReplicaSetHistory", func(t *testing.T) {
		events, err := rc.ReplicaSetHistory(context.Background(), ns, since, time.Time{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 0 {
			t.Errorf("expected 0 events, got %d", len(events))
		}
	})
}

func TestIntegration_OrphanedNamespaceCleanup(t *testing.T) {
	cs := newClientset(t)

	// List existing test namespaces
	nss, err := cs.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("list namespaces: %v", err)
	}

	for _, ns := range nss.Items {
		if strings.HasPrefix(ns.Name, "kube-tale-test-") {
			// Delete namespaces older than 1 hour
			if time.Since(ns.CreationTimestamp.Time) > 1*time.Hour {
				t.Logf("cleaning up orphaned namespace: %s", ns.Name)
				_ = cs.CoreV1().Namespaces().Delete(context.Background(), ns.Name, metav1.DeleteOptions{})
			}
		}
	}
}

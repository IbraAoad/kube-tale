//go:build integration

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

var binaryPath string

func TestMain(m *testing.M) {
	bin, err := os.MkdirTemp("", "kube-tale-e2e-*")
	if err != nil {
		os.Exit(1)
	}
	binaryPath = filepath.Join(bin, "kube-tale")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build: %v\n%s", err, out)
		os.Exit(1)
	}

	code := m.Run()

	_ = os.RemoveAll(bin)
	os.Exit(code)
}

func e2eKubeconfigPath() string {
	if p := os.Getenv("KUBECONFIG"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return home + "/.kube/config"
}

func newE2EClientset(t *testing.T) *kubernetes.Clientset {
	t.Helper()
	cfg, err := clientcmd.BuildConfigFromFlags("", e2eKubeconfigPath())
	if err != nil {
		t.Fatalf("build config: %v", err)
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("new clientset: %v", err)
	}
	return cs
}

func createE2ENamespace(t *testing.T, cs kubernetes.Interface) string {
	ns := fmt.Sprintf("kube-tale-e2e-%d", time.Now().UnixNano())
	_, err := cs.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: ns},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}
	t.Cleanup(func() {
		_ = cs.CoreV1().Namespaces().Delete(context.Background(), ns, metav1.DeleteOptions{})
	})
	return ns
}

func deployE2ENginx(t *testing.T, cs kubernetes.Interface, ns string) {
	replicas := int32(1)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "nginx", Namespace: ns},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "nginx"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "nginx"}},
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
		t.Fatalf("create deployment: %v", err)
	}

	deadline := time.After(45 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timeout waiting for nginx pod")
		default:
		}
		pods, err := cs.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			t.Fatalf("list pods: %v", err)
		}
		ready := false
		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodRunning {
				for _, c := range pod.Status.ContainerStatuses {
					if c.Ready {
						ready = true
						break
					}
				}
			}
		}
		if ready {
			return
		}
		time.Sleep(2 * time.Second)
	}
}

func deployE2ECrash(t *testing.T, cs kubernetes.Interface, ns string) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "crash-test", Namespace: ns},
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
		t.Fatalf("create pod: %v", err)
	}
}

func runCLI(args ...string) (string, string, error) {
	cmd := exec.Command(binaryPath, args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestCLI_Timeline_JSONOutput(t *testing.T) {
	cs := newE2EClientset(t)
	ns := createE2ENamespace(t, cs)
	deployE2ENginx(t, cs, ns)

	stdout, stderr, err := runCLI("timeline", "--namespace", ns, "--since", "5m", "--kubeconfig", e2eKubeconfigPath())
	if err != nil {
		t.Fatalf("timeline: %v\nstderr: %s", err, stderr)
	}

	var events []types.Event
	if err := json.Unmarshal([]byte(stdout), &events); err != nil {
		t.Fatalf("unmarshal timeline: %v\nstdout: %s", err, stdout)
	}

	if len(events) == 0 {
		t.Error("expected at least one event")
	}

	hasPodCreated := false
	hasPodReady := false
	for _, e := range events {
		if e.Kind == types.PodCreated {
			hasPodCreated = true
		}
		if e.Kind == types.PodReady {
			hasPodReady = true
		}
	}
	if !hasPodCreated {
		t.Error("expected PodCreated event")
	}
	if !hasPodReady {
		t.Error("expected PodReady event")
	}
}

func TestCLI_Story_Output(t *testing.T) {
	cs := newE2EClientset(t)
	ns := createE2ENamespace(t, cs)
	deployE2ENginx(t, cs, ns)

	stdout, stderr, err := runCLI("story", "--namespace", ns, "--since", "5m", "--kubeconfig", e2eKubeconfigPath())
	if err != nil {
		t.Fatalf("story: %v\nstderr: %s", err, stderr)
	}
	if stdout == "" {
		t.Error("expected non-empty story output")
	}
	if !strings.Contains(stdout, "nginx") {
		t.Errorf("story output should mention nginx, got: %s", stdout)
	}
}

func TestCLI_Story_CrashLoop(t *testing.T) {
	cs := newE2EClientset(t)
	ns := createE2ENamespace(t, cs)
	deployE2ECrash(t, cs, ns)

	time.Sleep(5 * time.Second)

	stdout, stderr, err := runCLI("story", "--namespace", ns, "--since", "2m", "--kubeconfig", e2eKubeconfigPath())
	if err != nil {
		t.Fatalf("story: %v\nstderr: %s", err, stderr)
	}
	if stdout == "" {
		t.Error("expected non-empty story output for crash pod")
	}
}

func TestCLI_Why_CrashLoop(t *testing.T) {
	cs := newE2EClientset(t)
	ns := createE2ENamespace(t, cs)
	deployE2ECrash(t, cs, ns)

	time.Sleep(5 * time.Second)

	stdout, stderr, err := runCLI("why", "--namespace", ns, "--since", "2m", "--kubeconfig", e2eKubeconfigPath())
	if err != nil {
		t.Fatalf("why: %v\nstderr: %s", err, stderr)
	}
	if stdout == "" {
		t.Error("expected non-empty why output for crash pod")
	}
}

func TestCLI_Diff_BeforeAfter(t *testing.T) {
	cs := newE2EClientset(t)
	ns := createE2ENamespace(t, cs)
	deployE2ENginx(t, cs, ns)

	stdout, stderr, err := runCLI("diff", "--namespace", ns, "--since", "10m", "--until", "0s", "--kubeconfig", e2eKubeconfigPath())
	if err != nil {
		t.Fatalf("diff: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "State change") {
		t.Errorf("expected 'State change' in diff output, got: %s", stdout)
	}
}

func TestCLI_NoKubeconfigFallback(t *testing.T) {
	_, stderr, err := runCLI("story", "--namespace", "default", "--since", "1h", "--kubeconfig", "/nonexistent")
	if err != nil {
		t.Fatalf("story with fake kubeconfig: %v", err)
	}
	if !strings.Contains(stderr, "falling back to mock data") {
		t.Errorf("expected fallback warning, got: %s", stderr)
	}
}

func TestCLI_HelpOutput(t *testing.T) {
	stdout, stderr, err := runCLI()
	if err == nil {
		t.Error("expected non-zero exit for no subcommand")
	}
	_ = stdout
	if !strings.Contains(stderr, "Usage:") {
		t.Errorf("expected 'Usage:' in stderr, got: %s", stderr)
	}
}

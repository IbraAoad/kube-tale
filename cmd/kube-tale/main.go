// Binary kube-tale correlates Kubernetes signals into incident narratives.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/IbraAoad/kube-tale/internal/client"
	"github.com/IbraAoad/kube-tale/internal/diff"
	"github.com/IbraAoad/kube-tale/internal/story"
	"github.com/IbraAoad/kube-tale/internal/timeline"
	"github.com/IbraAoad/kube-tale/internal/types"
	"github.com/IbraAoad/kube-tale/internal/why"
)

var version = "0.0.0-dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "timeline":
		cmdTimeline(os.Args[2:])
	case "story":
		cmdStory(os.Args[2:])
	case "why":
		cmdWhy(os.Args[2:])
	case "diff":
		cmdDiff(os.Args[2:])
	case "version":
		if _, err := fmt.Fprintf(os.Stdout, "kube-tale %s\n", version); err != nil {
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: kube-tale <command> [flags]

Commands:
  timeline    Print merged timeline as JSON
  story       Print human-readable incident narrative
  why         Print ranked root-cause hypotheses
  diff        Compare state between two time windows
  version     Print version

Flags:
  --namespace string    Kubernetes namespace (default "default")
  --since duration      Start of time window, relative to now (default 1h)
  --until duration      End of time window, relative to now (default 0s = now)
  --kubeconfig string   Path to kubeconfig file
`)
}

func parseSharedFlags(args []string) (namespace, kubeconfig string, since, until time.Duration) {
	fs := flag.NewFlagSet("shared", flag.ContinueOnError)
	fs.StringVar(&namespace, "namespace", "default", "Kubernetes namespace")
	fs.DurationVar(&since, "since", 1*time.Hour, "start of time window")
	fs.DurationVar(&until, "until", 0, "end of time window (0 = now)")
	fs.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	_ = fs.Parse(args)
	return
}

func parseDiffFlags(args []string) (before, after time.Duration) {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	fs.DurationVar(&before, "since", 2*time.Hour, "start of before window")
	fs.DurationVar(&after, "until", 1*time.Hour, "end of before window / start of after window")
	_ = fs.Parse(args)
	return
}

func newDataSource(kubeconfig string) client.DataSource {
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}
	if kubeconfig != "" {
		rc, err := client.NewRealClient(kubeconfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to create real client: %v, falling back to mock data\n", err)
		} else {
			return rc
		}
	}
	fmt.Fprintf(os.Stderr, "warning: no kubeconfig found, using mock data\n")
	return newMockDataSource()
}

func newMockDataSource() client.DataSource {
	mockEvents := []types.Event{
		{Timestamp: time.Date(2026, 5, 5, 10, 3, 0, 0, time.UTC), Kind: types.EventWarning, Source: "pod/api-7d9", Message: "Readiness probe failed (HTTP 500)"},
	}

	mockPodHistory := []types.Event{
		{Timestamp: time.Date(2026, 5, 5, 10, 2, 0, 0, time.UTC), Kind: types.PodCreated, Source: "pod/api-7d9", Message: "Pod api-7d9 created"},
		{Timestamp: time.Date(2026, 5, 5, 10, 4, 0, 0, time.UTC), Kind: types.CrashLoopBackOff, Source: "pod/api-7d9", Message: "Back-off restarting failed container"},
		{Timestamp: time.Date(2026, 5, 5, 10, 5, 0, 0, time.UTC), Kind: types.CrashLoopBackOff, Source: "pod/api-7d9", Message: "Back-off restarting failed container"},
		{Timestamp: time.Date(2026, 5, 5, 10, 6, 0, 0, time.UTC), Kind: types.CrashLoopBackOff, Source: "pod/api-7d9", Message: "Back-off restarting failed container"},
	}

	mockReplicaSetHistory := []types.Event{
		{Timestamp: time.Date(2026, 5, 5, 10, 1, 0, 0, time.UTC), Kind: types.DeploymentUpdate, Source: "deployment/api", Message: "Deployment api updated (image: v1.2 → v1.3)", Details: map[string]string{"old_image": "v1.2", "new_image": "v1.3"}},
	}

	return &client.MockClient{
		EventsFn: func(_ context.Context, _ string, _, _ time.Time) ([]types.Event, error) {
			return mockEvents, nil
		},
		PodHistoryFn: func(_ context.Context, _ string, _, _ time.Time) ([]types.Event, error) {
			return mockPodHistory, nil
		},
		ReplicaSetHistoryFn: func(_ context.Context, _ string, _, _ time.Time) ([]types.Event, error) {
			return mockReplicaSetHistory, nil
		},
	}
}

func cmdTimeline(args []string) {
	namespace, kubeconfig, since, until := parseSharedFlags(args)
	ds := newDataSource(kubeconfig)
	now := time.Now()
	tl, err := timeline.Build(context.Background(), ds, namespace, now.Add(-since), now.Add(-until))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	out, _ := json.MarshalIndent(tl, "", "  ")
	fmt.Println(string(out))
}

func cmdStory(args []string) {
	namespace, kubeconfig, since, until := parseSharedFlags(args)
	ds := newDataSource(kubeconfig)
	now := time.Now()
	tl, err := timeline.Build(context.Background(), ds, namespace, now.Add(-since), now.Add(-until))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(story.Generate(tl))
}

func cmdWhy(args []string) {
	namespace, kubeconfig, since, until := parseSharedFlags(args)
	ds := newDataSource(kubeconfig)
	now := time.Now()
	tl, err := timeline.Build(context.Background(), ds, namespace, now.Add(-since), now.Add(-until))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	results := why.Analyze(tl)
	if len(results) == 0 {
		fmt.Println("No root-cause hypotheses identified.")
		return
	}
	fmt.Println("Root cause hypotheses:")
	for i, h := range results {
		fmt.Printf("  %d. %s (confidence: %.2f)\n", i+1, h.Cause, h.Confidence)
		for _, ev := range h.Evidence {
			fmt.Printf("     → %s\n", ev)
		}
	}
}

func cmdDiff(args []string) {
	_ = flag.NewFlagSet("diff", flag.ContinueOnError)
	before, after := parseDiffFlags(args)
	_ = flag.NewFlagSet("diff", flag.ContinueOnError)
	namespace, kubeconfig, _, _ := parseSharedFlags(args)
	ds := newDataSource(kubeconfig)
	now := time.Now()

	beforeTL, err := timeline.Build(context.Background(), ds, namespace, now.Add(-before), now.Add(-after))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	afterTL, err := timeline.Build(context.Background(), ds, namespace, now.Add(-after), now)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	result := diff.Compare(beforeTL, afterTL)
	fmt.Printf("State change:\n")
	fmt.Printf("  Running pods:   %d → %d (%+d)\n", result.Before.RunningPods, result.After.RunningPods, result.Delta.RunningPods)
	fmt.Printf("  Crash loop pods: %d → %d (%+d)\n", result.Before.CrashLoopPods, result.After.CrashLoopPods, result.Delta.CrashLoopPods)
	fmt.Printf("  Total restarts:  %d → %d (%+d)\n", result.Before.TotalRestarts, result.After.TotalRestarts, result.Delta.TotalRestarts)
	fmt.Printf("  Total events:    %d → %d (%+d)\n", result.Before.EventCount, result.After.EventCount, result.Delta.EventCount)
	fmt.Printf("  Warnings:       %d → %d (%+d)\n", result.Before.WarningCount, result.After.WarningCount, result.Delta.WarningCount)
	fmt.Printf("  Errors:         %d → %d (%+d)\n", result.Before.ErrorCount, result.After.ErrorCount, result.Delta.ErrorCount)
}

// Binary kube-tale correlates Kubernetes signals into incident narratives.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/IbraAoad/kube-tale/internal/client"
	"github.com/IbraAoad/kube-tale/internal/diff"
	"github.com/IbraAoad/kube-tale/internal/story"
	"github.com/IbraAoad/kube-tale/internal/timeline"
	"github.com/IbraAoad/kube-tale/internal/types"
	"github.com/IbraAoad/kube-tale/internal/why"
)

var (
	version   = "0.0.0-dev"
	commit    = "unknown"
	buildDate = "unknown"
	goVersion = "unknown"
)

func init() {
	goVersion = runtime.Version()
}

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
		if _, err := fmt.Fprintf(os.Stdout, "kube-tale %s\n  git commit: %s\n  build date: %s\n  go version: %s\n", version, commit, buildDate, goVersion); err != nil {
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

func parseSharedFlags(args []string) (namespace, kubeconfig, output string, since, until time.Duration) {
	fs := flag.NewFlagSet("shared", flag.ContinueOnError)
	var sinceStr, untilStr string
	fs.StringVar(&namespace, "namespace", "default", "Kubernetes namespace")
	fs.StringVar(&sinceStr, "since", "1h", "start of time window (duration or RFC3339)")
	fs.StringVar(&untilStr, "until", "0s", "end of time window (duration or RFC3339, 0 = now)")
	fs.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	fs.StringVar(&output, "output", "", "output format: json, yaml, or text")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing flags: %v\n", err)
		os.Exit(1)
	}

	var err error
	since, err = parseTimeWindowFlag(sinceStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	until, err = parseTimeWindowFlag(untilStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	return
}

func parseDiffFlags(args []string) (before, after time.Duration) {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	fs.DurationVar(&before, "since", 2*time.Hour, "start of before window")
	fs.DurationVar(&after, "until", 1*time.Hour, "end of before window / start of after window")
	_ = fs.Parse(args)
	return
}

func parseTimeWindowFlag(s string) (time.Duration, error) {
	if s == "" || s == "0" {
		return 0, nil
	}

	d, err := time.ParseDuration(s)
	if err == nil {
		return d, nil
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return 0, fmt.Errorf("invalid time window value %q: must be a duration (e.g. 1h) or RFC3339 timestamp (e.g. 2026-05-05T10:00:00Z)", s)
	}

	return time.Since(t), nil
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
	var format string
	{
		fs := flag.NewFlagSet("timeline", flag.ContinueOnError)
		fs.StringVar(&format, "format", "json", "output format: json or text")
		_ = fs.Parse(args)
	}
	namespace, kubeconfig, output, since, until := parseSharedFlags(args)

	ds := newDataSource(kubeconfig)
	now := time.Now()
	tl, err := timeline.Build(context.Background(), ds, namespace, now.Add(-since), now.Add(-until))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	switch {
	case format == "text":
		fmt.Print(formatTimelineText(tl))
	case output == "yaml":
		fmt.Print(formatYAMLTimeline(tl))
	default:
		fmt.Print(formatJSONTimeline(tl))
	}
}

func cmdStory(args []string) {
	var verbose bool
	var filtered []string
	for _, a := range args {
		if a == "--verbose" || a == "-verbose" {
			verbose = true
		} else {
			filtered = append(filtered, a)
		}
	}

	namespace, kubeconfig, output, since, until := parseSharedFlags(filtered)
	ds := newDataSource(kubeconfig)
	now := time.Now()
	tl, err := timeline.Build(context.Background(), ds, namespace, now.Add(-since), now.Add(-until))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	var s string
	if verbose {
		s = story.GenerateVerbose(tl)
	} else {
		s = story.Generate(tl)
	}
	switch output {
	case "json":
		out, _ := json.Marshal(map[string]string{"story": s})
		fmt.Println(string(out))
	case "yaml":
		fmt.Printf("story: %q\n", s)
	default:
		fmt.Print(s)
	}
}

func cmdWhy(args []string) {
	namespace, kubeconfig, output, since, until := parseSharedFlags(args)
	ds := newDataSource(kubeconfig)
	now := time.Now()
	tl, err := timeline.Build(context.Background(), ds, namespace, now.Add(-since), now.Add(-until))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	results := why.Analyze(tl)
	switch output {
	case "json":
		out, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(out))
	case "yaml":
		if len(results) == 0 {
			fmt.Println("[]")
			return
		}
		for i, h := range results {
			fmt.Printf("- cause: %q\n", h.Cause)
			fmt.Printf("  confidence: %.2f\n", h.Confidence)
			fmt.Print("  evidence:\n")
			for _, e := range h.Evidence {
				fmt.Printf("    - %q\n", e)
			}
			if i < len(results)-1 {
				fmt.Println()
			}
		}
	default:
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
}

func cmdDiff(args []string) {
	_ = flag.NewFlagSet("diff", flag.ContinueOnError)
	before, after := parseDiffFlags(args)
	_ = flag.NewFlagSet("diff", flag.ContinueOnError)
	namespace, kubeconfig, output, _, _ := parseSharedFlags(args)
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
	switch output {
	case "json":
		out, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(out))
	case "yaml":
		fmt.Print(formatYAMLDiff(result))
	default:
		fmt.Printf("State change:\n")
		fmt.Printf("  Running pods:   %d → %d (%+d)\n", result.Before.RunningPods, result.After.RunningPods, result.Delta.RunningPods)
		fmt.Printf("  Crash loop pods: %d → %d (%+d)\n", result.Before.CrashLoopPods, result.After.CrashLoopPods, result.Delta.CrashLoopPods)
		fmt.Printf("  Total restarts:  %d → %d (%+d)\n", result.Before.TotalRestarts, result.After.TotalRestarts, result.Delta.TotalRestarts)
		fmt.Printf("  Total events:    %d → %d (%+d)\n", result.Before.EventCount, result.After.EventCount, result.Delta.EventCount)
		fmt.Printf("  Warnings:       %d → %d (%+d)\n", result.Before.WarningCount, result.After.WarningCount, result.Delta.WarningCount)
		fmt.Printf("  Errors:         %d → %d (%+d)\n", result.Before.ErrorCount, result.After.ErrorCount, result.Delta.ErrorCount)
	}
}

func formatJSONTimeline(tl types.Timeline) string {
	out, _ := json.MarshalIndent(tl, "", "  ")
	return string(out) + "\n"
}

func formatYAMLTimeline(tl types.Timeline) string {
	if len(tl) == 0 {
		return "[]\n"
	}
	var sb strings.Builder
	for i, e := range tl {
		fmt.Fprintf(&sb, "- timestamp: %q\n", e.Timestamp.Format(time.RFC3339))
		fmt.Fprintf(&sb, "  kind: %q\n", e.Kind)
		fmt.Fprintf(&sb, "  namespace: %q\n", e.Namespace)
		fmt.Fprintf(&sb, "  source: %q\n", e.Source)
		fmt.Fprintf(&sb, "  message: %q\n", e.Message)
		if len(e.Details) > 0 {
			sb.WriteString("  details:\n")
			for k, v := range e.Details {
				fmt.Fprintf(&sb, "    %s: %q\n", k, v)
			}
		}
		if i < len(tl)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func formatYAMLDiff(result diff.Result) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "before:\n")
	fmt.Fprintf(&sb, "  running_pods: %d\n", result.Before.RunningPods)
	fmt.Fprintf(&sb, "  crash_loop_pods: %d\n", result.Before.CrashLoopPods)
	fmt.Fprintf(&sb, "  total_restarts: %d\n", result.Before.TotalRestarts)
	fmt.Fprintf(&sb, "  event_count: %d\n", result.Before.EventCount)
	fmt.Fprintf(&sb, "  warning_count: %d\n", result.Before.WarningCount)
	fmt.Fprintf(&sb, "  error_count: %d\n", result.Before.ErrorCount)
	fmt.Fprintf(&sb, "after:\n")
	fmt.Fprintf(&sb, "  running_pods: %d\n", result.After.RunningPods)
	fmt.Fprintf(&sb, "  crash_loop_pods: %d\n", result.After.CrashLoopPods)
	fmt.Fprintf(&sb, "  total_restarts: %d\n", result.After.TotalRestarts)
	fmt.Fprintf(&sb, "  event_count: %d\n", result.After.EventCount)
	fmt.Fprintf(&sb, "  warning_count: %d\n", result.After.WarningCount)
	fmt.Fprintf(&sb, "  error_count: %d\n", result.After.ErrorCount)
	fmt.Fprintf(&sb, "delta:\n")
	fmt.Fprintf(&sb, "  running_pods: %+d\n", result.Delta.RunningPods)
	fmt.Fprintf(&sb, "  crash_loop_pods: %+d\n", result.Delta.CrashLoopPods)
	fmt.Fprintf(&sb, "  total_restarts: %+d\n", result.Delta.TotalRestarts)
	fmt.Fprintf(&sb, "  event_count: %+d\n", result.Delta.EventCount)
	fmt.Fprintf(&sb, "  warning_count: %+d\n", result.Delta.WarningCount)
	fmt.Fprintf(&sb, "  error_count: %+d\n", result.Delta.ErrorCount)
	return sb.String()
}

func formatTimelineText(tl types.Timeline) string {
	if len(tl) == 0 {
		return "(no events)\n"
	}

	maxTS := len("TIMESTAMP")
	maxKind := len("KIND")
	maxSrc := len("SOURCE")
	maxMsg := len("MESSAGE")

	for _, e := range tl {
		ts := e.Timestamp.Format(time.RFC3339)
		kind := string(e.Kind)
		if l := len(ts); l > maxTS {
			maxTS = l
		}
		if l := len(kind); l > maxKind {
			maxKind = l
		}
		if l := len(e.Source); l > maxSrc {
			maxSrc = l
		}
		if l := len(e.Message); l > maxMsg {
			maxMsg = l
		}
	}

	fmtStr := fmt.Sprintf("%%-%ds  %%-%ds  %%-%ds  %%s\n", maxTS, maxKind, maxSrc)

	var sb strings.Builder
	fmt.Fprintf(&sb, fmtStr, "TIMESTAMP", "KIND", "SOURCE", "MESSAGE")

	for _, e := range tl {
		ts := e.Timestamp.Format(time.RFC3339)
		kind := string(e.Kind)
		fmt.Fprintf(&sb, fmtStr, ts, kind, e.Source, e.Message)
	}

	return sb.String()
}

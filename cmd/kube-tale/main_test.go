package main

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/IbraAoad/kube-tale/internal/types"
)

func TestVersionOutputContainsBuildMetadata(t *testing.T) {
	oldArgs := os.Args
	oldStdout := os.Stdout
	defer func() {
		os.Args = oldArgs
		os.Stdout = oldStdout
	}()

	os.Args = []string{"kube-tale", "version"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	main()

	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	output := string(out)

	if !strings.Contains(output, "git commit:") {
		t.Errorf("expected 'git commit:' in version output, got: %s", output)
	}
	if !strings.Contains(output, "build date:") {
		t.Errorf("expected 'build date:' in version output, got: %s", output)
	}
	if !strings.Contains(output, "go version:") {
		t.Errorf("expected 'go version:' in version output, got: %s", output)
	}
	if !strings.Contains(output, "kube-tale") {
		t.Errorf("expected 'kube-tale' in version output, got: %s", output)
	}
}

func TestTimelineOutputYAML(t *testing.T) {
	oldArgs := os.Args
	oldStdout := os.Stdout
	defer func() {
		os.Args = oldArgs
		os.Stdout = oldStdout
	}()

	os.Args = []string{"kube-tale", "timeline", "--output", "yaml"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	main()

	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	output := string(out)

	if !strings.Contains(output, "- timestamp:") {
		t.Errorf("expected YAML output with '- timestamp:', got: %s", output)
	}
	if !strings.Contains(output, "kind:") {
		t.Errorf("expected YAML output with 'kind:', got: %s", output)
	}
}

func TestStoryOutputJSON(t *testing.T) {
	oldArgs := os.Args
	oldStdout := os.Stdout
	defer func() {
		os.Args = oldArgs
		os.Stdout = oldStdout
	}()

	os.Args = []string{"kube-tale", "story", "--output", "json"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	main()

	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	output := string(out)

	if !strings.Contains(output, `"story":`) {
		t.Errorf("expected JSON output with 'story' key, got: %s", output)
	}
}

func TestWhyOutputJSON(t *testing.T) {
	oldArgs := os.Args
	oldStdout := os.Stdout
	defer func() {
		os.Args = oldArgs
		os.Stdout = oldStdout
	}()

	os.Args = []string{"kube-tale", "why", "--output", "json"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	main()

	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	output := string(out)

	if !strings.Contains(output, `"cause":`) {
		t.Errorf("expected JSON output with 'cause' field, got: %s", output)
	}
}

func TestDiffOutputYAML(t *testing.T) {
	oldArgs := os.Args
	oldStdout := os.Stdout
	defer func() {
		os.Args = oldArgs
		os.Stdout = oldStdout
	}()

	os.Args = []string{"kube-tale", "diff", "--output", "yaml"}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	main()

	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	output := string(out)

	if !strings.Contains(output, "before:") {
		t.Errorf("expected YAML output with 'before:', got: %s", output)
	}
	if !strings.Contains(output, "after:") {
		t.Errorf("expected YAML output with 'after:', got: %s", output)
	}
	if !strings.Contains(output, "delta:") {
		t.Errorf("expected YAML output with 'delta:', got: %s", output)
	}
}

func TestFormatTimelineText_Empty(t *testing.T) {
	result := formatTimelineText(types.Timeline{})
	if result != "(no events)\n" {
		t.Errorf("empty timeline: got %q, want %q", result, "(no events)\n")
	}
}

func TestFormatTimelineText_SingleEvent(t *testing.T) {
	tl := types.Timeline{
		{
			Timestamp: time.Date(2026, 5, 5, 10, 1, 0, 0, time.UTC),
			Kind:      types.DeploymentUpdate,
			Source:    "deployment/api",
			Message:   "Deployment api updated",
		},
	}
	result := formatTimelineText(tl)
	if result == "" {
		t.Error("expected non-empty output")
	}
}

func TestFormatTimelineText_MultipleEvents(t *testing.T) {
	tl := types.Timeline{
		{
			Timestamp: time.Date(2026, 5, 5, 10, 1, 0, 0, time.UTC),
			Kind:      types.DeploymentUpdate,
			Source:    "deployment/api",
			Message:   "Deployment api updated",
		},
		{
			Timestamp: time.Date(2026, 5, 5, 10, 2, 0, 0, time.UTC),
			Kind:      types.CrashLoopBackOff,
			Source:    "pod/api-7d9",
			Message:   "Back-off restarting failed container",
		},
	}
	result := formatTimelineText(tl)
	if result == "" {
		t.Error("expected non-empty output")
	}
}

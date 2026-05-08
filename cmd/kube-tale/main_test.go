package main

import (
	"io"
	"os"
	"strings"
	"testing"
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

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

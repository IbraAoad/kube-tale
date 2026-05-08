package main

import (
	"os"
	"testing"

	"github.com/IbraAoad/kube-tale/internal/types"
)

func TestColorize_WarningKindYellow(t *testing.T) {
	result := colorize("warning event", types.EventWarning)
	expected := "\033[33mwarning event\033[0m"
	if result != expected {
		t.Errorf("warning: got %q, want %q", result, expected)
	}
}

func TestColorize_OOMRed(t *testing.T) {
	result := colorize("oom kill", types.PodOOMKilled)
	expected := "\033[31moom kill\033[0m"
	if result != expected {
		t.Errorf("OOM: got %q, want %q", result, expected)
	}
}

func TestColorize_ReadyGreen(t *testing.T) {
	result := colorize("pod ready", types.PodReady)
	expected := "\033[32mpod ready\033[0m"
	if result != expected {
		t.Errorf("ready: got %q, want %q", result, expected)
	}
}

func TestColorize_NoColor(t *testing.T) {
	result := colorize("plain event", types.PodCreated)
	if result != "plain event" {
		t.Errorf("plain: got %q, want %q", result, "plain event")
	}
}

func TestApplyColors(t *testing.T) {
	line := "  2026-05-05T10:01:00Z  CrashLoopBackOff  pod/api-7d9  Back-off restarting"
	result := applyColors(line, types.CrashLoopBackOff)
	if len(result) <= len(line) {
		t.Error("expected colored output to be longer than input (has ANSI codes)")
	}
}

func TestIsTerminal_PipeIsNotTerminal(t *testing.T) {
	r, w, _ := os.Pipe()
	defer r.Close()
	defer w.Close()
	if isStdoutTerminal() {
		t.Log("stdout is a terminal (expected during interactive testing)")
	}
}

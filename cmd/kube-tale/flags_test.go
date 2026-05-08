package main

import (
	"testing"
	"time"
)

func TestParseTimeWindowFlag_DurationString(t *testing.T) {
	d, err := parseTimeWindowFlag("1h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != time.Hour {
		t.Errorf("expected 1h, got %v", d)
	}
}

func TestParseTimeWindowFlag_RFC3339(t *testing.T) {
	now := time.Now()
	ref := now.Add(-30 * time.Minute)

	d, err := parseTimeWindowFlag(ref.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := now.Sub(ref)
	if d < expected-time.Second || d > expected+time.Second {
		t.Errorf("expected ~%v, got %v", expected, d)
	}
}

func TestParseTimeWindowFlag_InvalidFormat(t *testing.T) {
	_, err := parseTimeWindowFlag("not-a-time")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestParseTimeWindowFlag_RFC3339WithTimezone(t *testing.T) {
	ref, _ := time.Parse(time.RFC3339, "2026-05-05T10:00:00+02:00")
	d, err := parseTimeWindowFlag("2026-05-05T10:00:00+02:00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Since(ref)
	if d < expected-time.Second || d > expected+time.Second {
		t.Errorf("expected ~%v, got %v", expected, d)
	}
}

package why

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/IbraAoad/kube-tale/internal/types"
)

func loadWhyInput(t *testing.T, path string) types.Timeline {
	t.Helper()
	f, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read input %s: %v", path, err)
	}
	var tl types.Timeline
	if err := json.Unmarshal(f, &tl); err != nil {
		t.Fatalf("unmarshal input %s: %v", path, err)
	}
	return tl
}

func loadExpectedHypotheses(t *testing.T, path string) []Hypothesis {
	t.Helper()
	f, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read expected %s: %v", path, err)
	}
	content := string(f)
	if content == "null\n" || content == "null" {
		return nil
	}
	var h []Hypothesis
	if err := json.Unmarshal(f, &h); err != nil {
		t.Fatalf("unmarshal expected %s: %v", path, err)
	}
	return h
}

func TestAnalyze_OOM(t *testing.T) {
	tl := loadWhyInput(t, "testdata/input_oom.json")
	expected := loadExpectedHypotheses(t, "testdata/expected_oom.json")

	result := Analyze(tl)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("mismatch:\n got:  %+v\n want: %+v", result, expected)
	}
}

func TestAnalyze_BadImage(t *testing.T) {
	tl := loadWhyInput(t, "testdata/input_bad_image.json")
	expected := loadExpectedHypotheses(t, "testdata/expected_bad_image.json")

	result := Analyze(tl)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("mismatch:\n got:  %+v\n want: %+v", result, expected)
	}
}

func TestAnalyze_CrashLoop(t *testing.T) {
	tl := loadWhyInput(t, "testdata/input_crash_loop.json")
	expected := loadExpectedHypotheses(t, "testdata/expected_crash_loop.json")

	result := Analyze(tl)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("mismatch:\n got:  %+v\n want: %+v", result, expected)
	}
}

func TestAnalyze_ProbeFail(t *testing.T) {
	tl := loadWhyInput(t, "testdata/input_probe_fail.json")
	expected := loadExpectedHypotheses(t, "testdata/expected_probe_fail.json")

	result := Analyze(tl)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("mismatch:\n got:  %+v\n want: %+v", result, expected)
	}
}

func TestAnalyze_Normal(t *testing.T) {
	tl := loadWhyInput(t, "testdata/input_normal.json")
	expected := loadExpectedHypotheses(t, "testdata/expected_normal.json")

	result := Analyze(tl)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("mismatch:\n got:  %+v\n want: %+v", result, expected)
	}
}

func TestAnalyze_Multiple(t *testing.T) {
	tl := loadWhyInput(t, "testdata/input_multiple.json")
	expected := loadExpectedHypotheses(t, "testdata/expected_multiple.json")

	result := Analyze(tl)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("mismatch:\n got:  %+v\n want: %+v", result, expected)
	}
}

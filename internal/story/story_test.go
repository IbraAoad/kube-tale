package story

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/IbraAoad/kube-tale/internal/types"
)

func loadTimeline(t *testing.T, path string) types.Timeline {
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

func loadExpectedString(t *testing.T, path string) string {
	t.Helper()
	f, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read expected %s: %v", path, err)
	}
	return string(f)
}

func TestGenerate_CrashLoop(t *testing.T) {
	tl := loadTimeline(t, "testdata/input_crash_loop.json")
	expected := loadExpectedString(t, "testdata/expected_crash_loop.txt")

	result := Generate(tl)
	if result != expected {
		t.Errorf("mismatch:\n got:\n%s\n want:\n%s", result, expected)
	}
}

func TestGenerate_SuccessfulRollout(t *testing.T) {
	tl := loadTimeline(t, "testdata/input_successful_rollout.json")
	expected := loadExpectedString(t, "testdata/expected_successful_rollout.txt")

	result := Generate(tl)
	if result != expected {
		t.Errorf("mismatch:\n got:\n%s\n want:\n%s", result, expected)
	}
}

func TestGenerate_OOM(t *testing.T) {
	tl := loadTimeline(t, "testdata/input_oom.json")
	expected := loadExpectedString(t, "testdata/expected_oom.txt")

	result := Generate(tl)
	if result != expected {
		t.Errorf("mismatch:\n got:\n%s\n want:\n%s", result, expected)
	}
}

func TestGenerate_MixedPods(t *testing.T) {
	tl := loadTimeline(t, "testdata/input_mixed_pods.json")
	expected := loadExpectedString(t, "testdata/expected_mixed_pods.txt")

	result := Generate(tl)
	if result != expected {
		t.Errorf("mismatch:\n got:\n%s\n want:\n%s", result, expected)
	}
}

func TestGenerate_ProseRawCondition(t *testing.T) {
	tl := loadTimeline(t, "testdata/input_prose_raw.json")
	expected := loadExpectedString(t, "testdata/expected_prose_raw.txt")

	result := Generate(tl)
	if result != expected {
		t.Errorf("mismatch:\n got:\n%s\n want:\n%s", result, expected)
	}
}

func TestGenerate_ProseFailedRollout(t *testing.T) {
	tl := loadTimeline(t, "testdata/input_prose_failed.json")
	expected := loadExpectedString(t, "testdata/expected_prose_failed.txt")

	result := Generate(tl)
	if result != expected {
		t.Errorf("mismatch:\n got:\n%s\n want:\n%s", result, expected)
	}
}

func TestGenerate_CrashLoopVerbose(t *testing.T) {
	tl := loadTimeline(t, "testdata/input_verbose.json")
	expected := loadExpectedString(t, "testdata/expected_crash_loop_verbose.txt")

	result := GenerateVerbose(tl)
	if result != expected {
		t.Errorf("mismatch:\n got:\n%s\n want:\n%s", result, expected)
	}
}

func TestGenerate_SummaryFooter(t *testing.T) {
	tl := loadTimeline(t, "testdata/input_footer.json")
	expected := loadExpectedString(t, "testdata/expected_footer.txt")

	result := Generate(tl)
	if result != expected {
		t.Errorf("mismatch:\n got:\n%s\n want:\n%s", result, expected)
	}
}

func TestGenerate_Empty(t *testing.T) {
	tl := loadTimeline(t, "testdata/input_empty.json")
	expected := loadExpectedString(t, "testdata/expected_empty.txt")

	result := Generate(tl)
	if result != expected {
		t.Errorf("mismatch:\n got:\n%s\n want:\n%s", result, expected)
	}
}

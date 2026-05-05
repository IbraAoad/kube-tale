package diff

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/IbraAoad/kube-tale/internal/types"
)

func loadDiffInput(t *testing.T, path string) types.Timeline {
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

func loadExpectedResult(t *testing.T, path string) Result {
	t.Helper()
	f, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read expected %s: %v", path, err)
	}
	var r Result
	if err := json.Unmarshal(f, &r); err != nil {
		t.Fatalf("unmarshal expected %s: %v", path, err)
	}
	return r
}

func TestCompare_Basic(t *testing.T) {
	before := loadDiffInput(t, "testdata/input_before.json")
	after := loadDiffInput(t, "testdata/input_after.json")
	expected := loadExpectedResult(t, "testdata/expected_delta.json")

	result := Compare(before, after)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("mismatch:\n got:  %+v\n want: %+v", result, expected)
	}
}

func TestCompare_Empty(t *testing.T) {
	before := types.Timeline{}
	after := types.Timeline{}

	result := Compare(before, after)
	zero := StateSnapshot{}
	if result.Before != zero || result.After != zero || result.Delta != zero {
		t.Errorf("expected all zero, got %+v", result)
	}
}

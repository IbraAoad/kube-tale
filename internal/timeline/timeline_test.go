package timeline

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/IbraAoad/kube-tale/internal/client"
	"github.com/IbraAoad/kube-tale/internal/types"
)

type inputData struct {
	Events             []types.Event `json:"events"`
	PodHistory         []types.Event `json:"pod_history"`
	ReplicaSetHistory  []types.Event `json:"replicaset_history"`
}

func loadInput(t *testing.T, path string) inputData {
	t.Helper()
	f, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read input %s: %v", path, err)
	}
	var in inputData
	if err := json.Unmarshal(f, &in); err != nil {
		t.Fatalf("unmarshal input %s: %v", path, err)
	}
	return in
}

func loadExpected(t *testing.T, path string) types.Timeline {
	t.Helper()
	f, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read expected %s: %v", path, err)
	}
	var tl types.Timeline
	if err := json.Unmarshal(f, &tl); err != nil {
		t.Fatalf("unmarshal expected %s: %v", path, err)
	}
	return tl
}

func newMockClient(in inputData) *client.MockClient {
	return &client.MockClient{
		EventsFn: func(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error) {
			return in.Events, nil
		},
		PodHistoryFn: func(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error) {
			return in.PodHistory, nil
		},
		ReplicaSetHistoryFn: func(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error) {
			return in.ReplicaSetHistory, nil
		},
	}
}

func TestBuild_Basic(t *testing.T) {
	in := loadInput(t, "testdata/input_basic.json")
	expected := loadExpected(t, "testdata/expected_basic.json")
	mc := newMockClient(in)

	result, err := Build(context.Background(), mc, "default", time.Time{}, time.Now())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("mismatch:\n got:  %+v\n want: %+v", result, expected)
	}
}

func TestBuild_Empty(t *testing.T) {
	in := loadInput(t, "testdata/input_empty.json")
	expected := loadExpected(t, "testdata/expected_empty.json")
	mc := newMockClient(in)

	result, err := Build(context.Background(), mc, "default", time.Time{}, time.Now())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("mismatch:\n got:  %+v\n want: %+v", result, expected)
	}
}

func TestBuild_Overlap(t *testing.T) {
	in := loadInput(t, "testdata/input_overlap.json")
	expected := loadExpected(t, "testdata/expected_overlap.json")
	mc := newMockClient(in)

	result, err := Build(context.Background(), mc, "default", time.Time{}, time.Now())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("mismatch:\n got:  %+v\n want: %+v", result, expected)
	}
}

func TestBuild_Duplicates(t *testing.T) {
	in := loadInput(t, "testdata/input_duplicates.json")
	expected := loadExpected(t, "testdata/expected_duplicates.json")
	mc := newMockClient(in)

	result, err := Build(context.Background(), mc, "default", time.Time{}, time.Now())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("mismatch:\n got:  %+v\n want: %+v", result, expected)
	}
}

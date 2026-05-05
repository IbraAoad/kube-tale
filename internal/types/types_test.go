package types

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestEventJSONRoundTrip(t *testing.T) {
	original := Event{
		Timestamp: time.Date(2026, 5, 5, 10, 1, 0, 0, time.UTC),
		Kind:      PodCreated,
		Namespace: "default",
		Source:    "pod/api-7d9",
		Message:   "Pod api-7d9 created",
		Details:   map[string]string{"image": "v1.3"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var restored Event
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !reflect.DeepEqual(original, restored) {
		t.Errorf("round-trip mismatch:\n got:  %+v\n want: %+v", restored, original)
	}
}

func TestTimelineIsSlice(t *testing.T) {
	tl := Timeline{
		{Timestamp: time.Now(), Kind: PodReady, Message: "ready"},
		{Timestamp: time.Now(), Kind: CrashLoopBackOff, Message: "crash"},
	}
	if len(tl) != 2 {
		t.Fatalf("expected 2 events, got %d", len(tl))
	}
}

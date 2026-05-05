package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IbraAoad/kube-tale/internal/types"
)

func TestMockClientEventsDelegates(t *testing.T) {
	want := []types.Event{{Kind: types.PodReady, Message: "ok"}}
	mc := &MockClient{
		EventsFn: func(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error) {
			return want, nil
		},
	}
	got, err := mc.Events(context.Background(), "default", time.Time{}, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Message != "ok" {
		t.Fatalf("delegation failed: got %+v", got)
	}
}

func TestMockClientEventsNilFn(t *testing.T) {
	mc := &MockClient{}
	_, err := mc.Events(context.Background(), "default", time.Time{}, time.Now())
	if err == nil {
		t.Fatal("expected error for nil EventsFn, got nil")
	}
}

func TestMockClientPodHistoryDelegates(t *testing.T) {
	want := []types.Event{{Kind: types.CrashLoopBackOff, Message: "crash"}}
	mc := &MockClient{
		PodHistoryFn: func(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error) {
			return want, nil
		},
	}
	got, err := mc.PodHistory(context.Background(), "default", time.Time{}, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Kind != types.CrashLoopBackOff {
		t.Fatalf("delegation failed: got %+v", got)
	}
}

func TestMockClientPodHistoryNilFn(t *testing.T) {
	mc := &MockClient{}
	_, err := mc.PodHistory(context.Background(), "default", time.Time{}, time.Now())
	if err == nil {
		t.Fatal("expected error for nil PodHistoryFn, got nil")
	}
}

func TestMockClientReplicaSetHistoryDelegates(t *testing.T) {
	want := []types.Event{{Kind: types.DeploymentUpdate, Message: "updated"}}
	mc := &MockClient{
		ReplicaSetHistoryFn: func(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error) {
			return want, nil
		},
	}
	got, err := mc.ReplicaSetHistory(context.Background(), "default", time.Time{}, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Kind != types.DeploymentUpdate {
		t.Fatalf("delegation failed: got %+v", got)
	}
}

func TestMockClientReplicaSetHistoryNilFn(t *testing.T) {
	mc := &MockClient{}
	_, err := mc.ReplicaSetHistory(context.Background(), "default", time.Time{}, time.Now())
	if err == nil {
		t.Fatal("expected error for nil ReplicaSetHistoryFn, got nil")
	}
}

func TestMockClientErrorPropagation(t *testing.T) {
	testErr := errors.New("simulated failure")
	mc := &MockClient{
		EventsFn: func(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error) {
			return nil, testErr
		},
	}
	_, err := mc.Events(context.Background(), "default", time.Time{}, time.Now())
	if err != testErr {
		t.Fatalf("expected %v, got %v", testErr, err)
	}
}

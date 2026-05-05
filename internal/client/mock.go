package client

import (
	"context"
	"fmt"
	"time"

	"github.com/IbraAoad/kube-tale/internal/types"
)

type MockClient struct {
	EventsFn            func(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error)
	PodHistoryFn        func(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error)
	ReplicaSetHistoryFn func(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error)
}

func (m *MockClient) Events(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error) {
	if m.EventsFn == nil {
		return nil, fmt.Errorf("MockClient.EventsFn not set")
	}
	return m.EventsFn(ctx, namespace, since, until)
}

func (m *MockClient) PodHistory(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error) {
	if m.PodHistoryFn == nil {
		return nil, fmt.Errorf("MockClient.PodHistoryFn not set")
	}
	return m.PodHistoryFn(ctx, namespace, since, until)
}

func (m *MockClient) ReplicaSetHistory(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error) {
	if m.ReplicaSetHistoryFn == nil {
		return nil, fmt.Errorf("MockClient.ReplicaSetHistoryFn not set")
	}
	return m.ReplicaSetHistoryFn(ctx, namespace, since, until)
}

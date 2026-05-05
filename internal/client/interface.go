package client

import (
	"context"
	"time"

	"github.com/IbraAoad/kube-tale/internal/types"
)

type DataSource interface {
	Events(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error)
	PodHistory(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error)
	ReplicaSetHistory(ctx context.Context, namespace string, since, until time.Time) ([]types.Event, error)
}

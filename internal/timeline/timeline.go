// Package timeline merges events from multiple data sources into a sorted, deduplicated timeline.
package timeline

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/IbraAoad/kube-tale/internal/client"
	"github.com/IbraAoad/kube-tale/internal/types"
)

// Build fetches events from all three DataSource methods, merges them
// chronologically, and removes duplicates. Returns an empty slice when
// no events are found.
func Build(ctx context.Context, c client.DataSource, namespace string, since, until time.Time) (types.Timeline, error) {
	events, err := c.Events(ctx, namespace, since, until)
	if err != nil {
		return nil, fmt.Errorf("fetch events: %w", err)
	}

	podHist, err := c.PodHistory(ctx, namespace, since, until)
	if err != nil {
		return nil, fmt.Errorf("fetch pod history: %w", err)
	}

	rsHist, err := c.ReplicaSetHistory(ctx, namespace, since, until)
	if err != nil {
		return nil, fmt.Errorf("fetch replicaset history: %w", err)
	}

	all := make([]types.Event, 0, len(rsHist)+len(podHist)+len(events))
	all = append(all, rsHist...)
	all = append(all, podHist...)
	all = append(all, events...)

	all = sortByTimestamp(all)
	all = deduplicate(all)

	if all == nil {
		all = types.Timeline{}
	}

	return all, nil
}

func sortByTimestamp(events []types.Event) []types.Event {
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})
	return events
}

type dedupKey struct {
	timestamp time.Time
	kind      types.EventKind
	namespace string
	source    string
	message   string
}

func deduplicate(events []types.Event) []types.Event {
	seen := make(map[dedupKey]bool)
	result := make([]types.Event, 0, len(events))
	for _, e := range events {
		key := dedupKey{
			timestamp: e.Timestamp,
			kind:      e.Kind,
			namespace: e.Namespace,
			source:    e.Source,
			message:   e.Message,
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, e)
	}
	return result
}

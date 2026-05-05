// Package diff compares state snapshots between two points in time.
package diff

import "github.com/IbraAoad/kube-tale/internal/types"

// StateSnapshot aggregates key metrics from a timeline.
type StateSnapshot struct {
	RunningPods   int `json:"running_pods"`
	CrashLoopPods int `json:"crash_loop_pods"`
	TotalRestarts int `json:"total_restarts"`
	EventCount    int `json:"event_count"`
	WarningCount  int `json:"warning_count"`
	ErrorCount    int `json:"error_count"`
	ReadyReplicas int `json:"ready_replicas"`
}

// Result holds the before/after snapshots and their delta.
type Result struct {
	Before StateSnapshot `json:"before"`
	After  StateSnapshot `json:"after"`
	Delta  StateSnapshot `json:"delta"`
}

// Compare analyzes two timelines (before and after a point in time) and returns the state difference.
func Compare(before, after types.Timeline) Result {
	beforeSnap := snapshot(before)
	afterSnap := snapshot(after)
	return Result{
		Before: beforeSnap,
		After:  afterSnap,
		Delta: StateSnapshot{
			RunningPods:   afterSnap.RunningPods - beforeSnap.RunningPods,
			CrashLoopPods: afterSnap.CrashLoopPods - beforeSnap.CrashLoopPods,
			TotalRestarts: afterSnap.TotalRestarts - beforeSnap.TotalRestarts,
			EventCount:    afterSnap.EventCount - beforeSnap.EventCount,
			WarningCount:  afterSnap.WarningCount - beforeSnap.WarningCount,
			ErrorCount:    afterSnap.ErrorCount - beforeSnap.ErrorCount,
			ReadyReplicas: afterSnap.ReadyReplicas - beforeSnap.ReadyReplicas,
		},
	}
}

func snapshot(timeline types.Timeline) StateSnapshot {
	type podState struct {
		ready   bool
		crashed bool
		dead    bool
	}
	pods := make(map[string]*podState)

	crashPods := make(map[string]bool)
	restarts := 0
	warnings := 0
	errors := 0

	for _, e := range timeline {
		src := e.Source
		if _, ok := pods[src]; !ok {
			pods[src] = &podState{}
		}
		switch e.Kind {
		case types.PodReady:
			pods[src].ready = true
			pods[src].crashed = false
		case types.PodTerminated:
			pods[src].dead = true
			pods[src].ready = false
		case types.CrashLoopBackOff:
			crashPods[src] = true
			pods[src].crashed = true
			pods[src].ready = false
			restarts++
		case types.EventWarning:
			warnings++
		case types.PodOOMKilled, types.ImagePullBackOff, types.PodEvicted:
			errors++
		}
	}

	running := 0
	ready := 0
	for _, ps := range pods {
		if ps.ready && !ps.crashed && !ps.dead {
			running++
			ready++
		}
	}

	return StateSnapshot{
		RunningPods:   running,
		CrashLoopPods: len(crashPods),
		TotalRestarts: restarts,
		EventCount:    len(timeline),
		WarningCount:  warnings,
		ErrorCount:    errors,
		ReadyReplicas: ready,
	}
}

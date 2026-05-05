// Package why performs pattern-based root-cause analysis on a timeline.
package why

import (
	"sort"

	"github.com/IbraAoad/kube-tale/internal/types"
)

// Hypothesis represents a possible root cause with supporting evidence.
type Hypothesis struct {
	Cause      string   `json:"cause"`
	Confidence float64  `json:"confidence"`
	Evidence   []string `json:"evidence"`
}

type matcher func(timeline types.Timeline) (Hypothesis, bool)

var matchers = []matcher{
	matchOOMKilled,
	matchBadImage,
	matchCrashLoop,
	matchProbeFailure,
}

// Analyze scans the timeline and returns ranked hypotheses (highest confidence first).
func Analyze(timeline types.Timeline) []Hypothesis {
	if len(timeline) == 0 {
		return nil
	}

	var results []Hypothesis
	for _, m := range matchers {
		if h, ok := m(timeline); ok {
			results = append(results, h)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Confidence > results[j].Confidence
	})

	if len(results) == 0 {
		return nil
	}
	return results
}

func matchOOMKilled(timeline types.Timeline) (Hypothesis, bool) {
	for _, e := range timeline {
		if e.Kind == types.PodOOMKilled {
			return Hypothesis{
				Cause:      "Out of Memory Kill",
				Confidence: 0.9,
				Evidence:   []string{e.Message},
			}, true
		}
	}
	return Hypothesis{}, false
}

func matchBadImage(timeline types.Timeline) (Hypothesis, bool) {
	var evidence []string
	hasReady := false
	for _, e := range timeline {
		if e.Kind == types.ImagePullBackOff {
			evidence = append(evidence, e.Message)
		}
		if e.Kind == types.PodReady {
			hasReady = true
		}
	}
	if len(evidence) > 0 && !hasReady {
		return Hypothesis{
			Cause:      "Bad Image",
			Confidence: 0.85,
			Evidence:   evidence,
		}, true
	}
	return Hypothesis{}, false
}

func matchCrashLoop(timeline types.Timeline) (Hypothesis, bool) {
	var evidence []string
	hasReady := false
	for _, e := range timeline {
		if e.Kind == types.CrashLoopBackOff {
			evidence = append(evidence, e.Message)
		}
		if e.Kind == types.PodReady {
			hasReady = true
		}
	}
	if len(evidence) >= 3 && !hasReady {
		return Hypothesis{
			Cause:      "Crash Loop",
			Confidence: 0.7,
			Evidence:   evidence,
		}, true
	}
	return Hypothesis{}, false
}

func matchProbeFailure(timeline types.Timeline) (Hypothesis, bool) {
	var evidence []string
	hasReady := false
	for _, e := range timeline {
		if e.Kind == types.ProbeFailed {
			evidence = append(evidence, e.Message)
		}
		if e.Kind == types.PodReady {
			hasReady = true
		}
	}
	if len(evidence) >= 3 && !hasReady {
		return Hypothesis{
			Cause:      "Probe Failure",
			Confidence: 0.65,
			Evidence:   evidence,
		}, true
	}
	return Hypothesis{}, false
}

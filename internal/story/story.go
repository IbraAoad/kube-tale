// Package story compresses a timeline into a human-readable incident narrative.
package story

import (
	"fmt"
	"strings"

	"github.com/IbraAoad/kube-tale/internal/types"
)

// Generate produces a human-readable narrative string from a sorted timeline.
func Generate(timeline types.Timeline) string {
	if len(timeline) == 0 {
		return ""
	}

	deployment, podGroups := partition(timeline)
	var parts []string

	skipPods := false

	if deployment != nil {
		msg := humanizeDeploymentMessage(*deployment)
		if allPodsReady(podGroups) && len(podGroups) > 0 {
			msg += " and rolled out successfully."
			skipPods = true
		} else {
			msg += "."
		}
		parts = append(parts, msg)
	}

	if !skipPods {
		for _, grp := range podGroups {
			story := narratePod(grp)
			if story != "" {
				parts = append(parts, story)
			}
		}
	}

	return strings.Join(parts, "\n") + "\n"
}

// GenerateVerbose produces a detailed narrative with timestamp-prefixed events
// followed by the compressed summary.
func GenerateVerbose(timeline types.Timeline) string {
	if len(timeline) == 0 {
		return ""
	}

	deployment, podGroups := partition(timeline)
	var parts []string

	skipPods := false

	if deployment != nil {
		msg := humanizeDeploymentMessage(*deployment)
		if allPodsReady(podGroups) && len(podGroups) > 0 {
			msg += " and rolled out successfully."
			skipPods = true
		} else {
			msg += "."
		}
		parts = append(parts, msg)
	}

	crashCounts := make(map[string]int)
	for _, e := range timeline {
		if e.Kind == types.DeploymentUpdate {
			parts = append(parts, fmt.Sprintf("[%s] %s",
				e.Timestamp.Format("15:04"), humanizeDeploymentMessage(e)))
			continue
		}

		switch e.Kind {
		case types.PodCreated:
			parts = append(parts, fmt.Sprintf("[%s] Pod %s was created",
				e.Timestamp.Format("15:04"), podName(e.Source)))
		case types.PodReady:
			parts = append(parts, fmt.Sprintf("[%s] Pod %s became ready",
				e.Timestamp.Format("15:04"), podName(e.Source)))
		case types.CrashLoopBackOff:
			crashCounts[e.Source]++
			parts = append(parts, fmt.Sprintf("[%s] Pod %s entered CrashLoopBackOff (restart %d)",
				e.Timestamp.Format("15:04"), podName(e.Source), crashCounts[e.Source]))
		case types.PodOOMKilled:
			parts = append(parts, fmt.Sprintf("[%s] Pod %s was OOMKilled",
				e.Timestamp.Format("15:04"), podName(e.Source)))
		case types.ProbeFailed:
			parts = append(parts, fmt.Sprintf("[%s] Pod %s probe failed",
				e.Timestamp.Format("15:04"), podName(e.Source)))
		case types.ImagePullBackOff:
			parts = append(parts, fmt.Sprintf("[%s] Pod %s image pull backoff",
				e.Timestamp.Format("15:04"), podName(e.Source)))
		default:
			parts = append(parts, fmt.Sprintf("[%s] %s: %s",
				e.Timestamp.Format("15:04"), string(e.Kind), e.Message))
		}
	}

	parts = append(parts, "")

	if !skipPods {
		for _, grp := range podGroups {
			story := narratePod(grp)
			if story != "" {
				parts = append(parts, story)
			}
		}
	}

	return strings.Join(parts, "\n") + "\n"
}

func humanizeDeploymentMessage(e types.Event) string {
	msg := e.Message
	if e.Details != nil {
		if oldImg, ok := e.Details["old_image"]; ok {
			if newImg, ok := e.Details["new_image"]; ok {
				msg = fmt.Sprintf("Deployment %s was updated (image: %s → %s)",
					deploymentName(e.Source), oldImg, newImg)
			}
		}
	}
	return msg
}

type podGroup struct {
	name   string
	events []types.Event
}

func partition(timeline types.Timeline) (*types.Event, []podGroup) {
	var deployment *types.Event
	podMap := make(map[string][]types.Event)
	podOrder := []string{}

	for _, e := range timeline {
		if e.Kind == types.DeploymentUpdate {
			ev := e
			deployment = &ev
			continue
		}
		if strings.HasPrefix(e.Source, "pod/") {
			if _, ok := podMap[e.Source]; !ok {
				podOrder = append(podOrder, e.Source)
			}
			podMap[e.Source] = append(podMap[e.Source], e)
		}
	}

	groups := make([]podGroup, 0, len(podOrder))
	for _, src := range podOrder {
		groups = append(groups, podGroup{name: podName(src), events: podMap[src]})
	}

	return deployment, groups
}

func allPodsReady(groups []podGroup) bool {
	if len(groups) == 0 {
		return false
	}
	for _, g := range groups {
		if !hasReady(g.events) {
			return false
		}
	}
	return true
}

func narratePod(grp podGroup) string {
	hasCreated := hasKind(grp.events, types.PodCreated)
	hasReady := hasKind(grp.events, types.PodReady)
	crashCount := countKind(grp.events, types.CrashLoopBackOff)
	hasOOM := hasKind(grp.events, types.PodOOMKilled)
	hasProbeFail := hasKind(grp.events, types.ProbeFailed)

	switch {
	case hasOOM:
		return fmt.Sprintf("Pod %s was killed due to out-of-memory.", grp.name)
	case hasProbeFail:
		return fmt.Sprintf("Pod %s failed its health probe.", grp.name)
	case hasCreated && crashCount > 0:
		return fmt.Sprintf("Pod %s was created but entered a crash loop (%d restarts).", grp.name, crashCount)
	case crashCount > 0:
		return fmt.Sprintf("Pod %s entered a crash loop (%d restarts).", grp.name, crashCount)
	case hasCreated && hasReady:
		return fmt.Sprintf("Pod %s was created and became ready.", grp.name)
	default:
		return ""
	}
}

func hasKind(events []types.Event, kind types.EventKind) bool {
	for _, e := range events {
		if e.Kind == kind {
			return true
		}
	}
	return false
}

func hasReady(events []types.Event) bool {
	return hasKind(events, types.PodReady)
}

func countKind(events []types.Event, kind types.EventKind) int {
	n := 0
	for _, e := range events {
		if e.Kind == kind {
			n++
		}
	}
	return n
}

func deploymentName(source string) string {
	return strings.TrimPrefix(source, "deployment/")
}

func podName(source string) string {
	return strings.TrimPrefix(source, "pod/")
}

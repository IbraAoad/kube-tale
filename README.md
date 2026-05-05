# kube-tale

Turn scattered Kubernetes signals into structured incident narratives.

kube-tale correlates existing Kubernetes signals — events, pod statuses, ReplicaSet history — and turns them into human-readable stories about **what happened, when, and why**.

No new observability. Just smarter correlation of what you already have.

## Features

kube-tale is organized into four modules:

| Module | Description | Status |
|--------|-------------|--------|
| **timeline** | Merged, sorted sequence of everything that happened (deployments, pod lifecycle, events, restarts) | Planned |
| **story** | Compressed human narrative — "This broke after a deployment update due to startup failures leading to crash loops." | Planned |
| **why** | Probabilistic root-cause guess using pattern-based scoring (OOMKilled, CrashLoopBackOff, probe failures, etc.) | Planned |
| **diff** | State comparison between two points in time (before/after deployment or incident) | Planned |

## Why kube-tale?

Debugging Kubernetes incidents means context-switching between `kubectl describe`, `kubectl events`, `kubectl rollout status`, and various dashboards. kube-tale brings it all together in one command.

```
$ kube-tale timeline --deployment api --namespace default --since 1h

[10:01] Deployment api updated (image: v1.2 → v1.3)
[10:02] Pod api-7d9 restarted (container started)
[10:03] Readiness probe failed (HTTP 500)
[10:04] Pod api-7d9 restarted (CrashLoopBackOff)
[10:06] New pod created: api-8a1
[10:07] Pod api-8a1 became Ready
```

## Installation

_Coming soon._

```bash
go install github.com/<you>/kube-tale@latest
```

Or download binaries from [Releases](https://github.com/<you>/kube-tale/releases).

## Quick Start

```bash
# Show timeline for a deployment
kube-tale timeline --deployment api --namespace default --since 1h

# Show human-readable story
kube-tale story --deployment api --namespace default --since 1h

# Guess root cause
kube-tale why --deployment api --namespace default --since 1h

# Diff state before and after an incident
kube-tale diff --deployment api --namespace default --since 1h --until 30m
```

## Architecture

kube-tale uses only standard Kubernetes data sources:

- **core/v1 Events** — Kubernetes event stream
- **Pod status / container status** — `containerStatuses` including restart counts, last state, ready state
- **Deployments + ReplicaSet history** — deployment conditions, RS generation changes

Optional (future):
- **metrics-server** — CPU/memory signals for OOM correlation

## Development

```bash
# Build
go build -o kube-tale ./cmd/kube-tale

# Run tests
go test ./...
```

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

[Apache 2.0](LICENSE)

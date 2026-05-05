# kube-tale

Turn scattered Kubernetes signals into structured incident narratives.

kube-tale correlates existing Kubernetes signals — events, pod statuses, ReplicaSet history — and turns them into human-readable stories about **what happened, when, and why**.

No new observability. Just smarter correlation of what you already have.

## Features

kube-tale is organized into four modules:

| Module | Description | Status |
|--------|-------------|--------|
| **timeline** | Merged, sorted sequence of everything that happened (deployments, pod lifecycle, events, restarts) | Done |
| **story** | Compressed human narrative — "This broke after a deployment update due to startup failures leading to crash loops." | Done |
| **why** | Probabilistic root-cause guess using pattern-based scoring (OOMKilled, CrashLoopBackOff, probe failures, etc.) | Done |
| **diff** | State comparison between two points in time (before/after deployment or incident) | Done |

## Prerequisites

- **Go 1.26+** (or download pre-built binary from [Releases](https://github.com/IbraAoad/kube-tale/releases))
- **Kubernetes cluster** — any K8s distribution (k3s, kind, minikube, production)
- **kubectl** — optional, for manual verification

## Installation

```bash
# Install via Go
go install github.com/IbraAoad/kube-tale@latest

# Or build from source
git clone https://github.com/IbraAoad/kube-tale
cd kube-tale
go build -o kube-tale ./cmd/kube-tale
```

Or download pre-built binaries from [Releases](https://github.com/IbraAoad/kube-tale/releases).

## Quick Start

```bash
# Show timeline for a namespace (auto-detects kubeconfig)
kube-tale timeline --namespace default --since 1h

# Show human-readable story
kube-tale story --namespace default --since 1h

# Guess root cause
kube-tale why --namespace default --since 1h

# Diff state before and after a point in time
kube-tale diff --namespace default --since 2h --until 1h

# Explicit kubeconfig
kube-tale timeline --namespace kube-system --kubeconfig ~/.kube/config --since 30m
```

Example output:

```
$ kube-tale story --namespace default --since 1h
Deployment api was updated (image: v1.2 → v1.3).
Pod api-7d9 was created but entered a crash loop (3 restarts).
```

```
$ kube-tale why --namespace default --since 1h
Root cause hypotheses:
  1. Crash Loop (confidence: 0.70)
     → Back-off restarting failed container
     → Back-off restarting failed container
     → Back-off restarting failed container
```

## Architecture

kube-tale uses only standard Kubernetes data sources via `client-go`:

- **core/v1 Events** — Kubernetes event stream (with reason-to-kind mapping)
- **Pod status / container status** — `containerStatuses` including restart counts, last state, ready state
- **Deployments + ReplicaSet history** — deployment conditions and progress

All cluster access goes through a single `DataSource` interface, making the correlation engine fully testable without a real cluster.

## Development

```bash
# Build
go build -o kube-tale ./cmd/kube-tale

# Run unit tests
go test -race -cover ./...

# Run lint
golangci-lint run

# Integration tests (requires a running K8s cluster)
go test -race -v -tags=integration ./internal/client/ ./cmd/kube-tale/
```

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

[Apache 2.0](LICENSE)

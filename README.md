# kube-tale

Turn scattered Kubernetes signals into structured incident narratives.

kube-tale correlates existing K8s signals — events, pod status, ReplicaSet history — and
turns them into human-readable stories about **what happened, when, and why**. No new
observability. Just smarter correlation of what you already have.

## Features

| Command | What it does |
|---------|-------------|
| `timeline` | Merged chronological view of deployments, pod lifecycle, events, and restarts |
| `story` | Compressed human-readable incident narrative with event-by-event detail (`--verbose`) |
| `why` | Ranked root-cause hypotheses with confidence scores |
| `diff` | State comparison between two points in time |
| `version` | Go version, commit hash, and build date |

**Output:** JSON, YAML, or colored text tables. See [Usage](#usage) below.

**Summary footer:** Every `story` output ends with a tally of pods, errors, warnings, and
deployments — a quick at-a-glance assessment.

## Quick Start

```bash
# Install
go install github.com/IbraAoad/kube-tale@latest

# Timeline (JSON or text table)
kube-tale timeline --namespace default --since 1h --format text --no-color

# Incident story (with verbose detail)
kube-tale story --namespace default --since 1h --verbose

# Root-cause analysis
kube-tale why --namespace default --since 1h

# State diff
kube-tale diff --namespace default --since 2h --until 1h

# Version info
kube-tale version
```

Example output (`story`):

```
Deployment api was updated (image: v1.2 → v1.3) but 1 pod(s) entered a crash loop.
Pod api-7d9 was created but entered a crash loop (3 restarts).

Summary: 1 pods, 3 errors, 1 warnings, 1 deployment
```

Example output (`why`):

```
Root cause hypotheses:
  1. Crash Loop (confidence: 0.70)
     → Back-off restarting failed container
     → Back-off restarting failed container
     → Back-off restarting failed container
```

## Installation

```bash
go install github.com/IbraAoad/kube-tale@latest

# Or build from source
git clone https://github.com/IbraAoad/kube-tale
cd kube-tale
go build -o kube-tale ./cmd/kube-tale
```

Pre-built binaries are available on [Releases](https://github.com/IbraAoad/kube-tale/releases).

## Usage

All subcommands accept shared flags:

| Flag | Description | Default |
|------|-------------|---------|
| `--namespace` | Kubernetes namespace | `default` |
| `--since` | Start of time window (duration or RFC3339) | `1h` |
| `--until` | End of time window (duration or RFC3339) | `0s` (now) |
| `--kubeconfig` | Path to kubeconfig file | `$KUBECONFIG` |
| `--output` | Output format: `json`, `yaml`, or `text` | (varies by cmd) |

Story-specific flags:

| Flag | Description |
|------|-------------|
| `--verbose` | Show per-event timestamps alongside the compressed summary |

Timeline-specific flags:

| Flag | Description | Default |
|------|-------------|---------|
| `--format` | Output style: `json` or `text` | `json` |
| `--no-color` | Disable colored output | `false` |

## Architecture

All cluster access goes through a `DataSource` interface using `client-go`:

- **Events** — Kubernetes event stream with reason-to-kind mapping
- **Pod status** — container statuses, restart counts, ready state
- **Deployments + ReplicaSets** — conditions and rollout history

The correlation engine is fully testable without a real cluster.

## Development

```bash
go build -o kube-tale ./cmd/kube-tale
go test -race -cover ./...
golangci-lint run
```

Integration tests require a running K8s cluster:

```bash
go test -race -v -tags=integration ./internal/client/ ./cmd/kube-tale/
```

## Contributing

Pull requests are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) and [AGENTS.md](AGENTS.md)
for conventions (TDD, commit format, branching).

## License

[Apache 2.0](LICENSE)

# kube-tale — Agent Notes

## Project
- Go CLI that correlates Kubernetes signals (events, pod status, ReplicaSet history) into incident narratives.
- **Module:** `github.com/IbraAoad/kube-tale`
- **Go version:** 1.26.2 (go.mod declares `go 1.26.2`; supported by golangci-lint-action v9.2.0+)
- **Status:** M0 through M7 complete (real K8s client, integration tests).
- **Entrypoint:** `cmd/kube-tale`

## Environment
- Go binary at `/usr/local/go/bin/go` — may not be in default `$PATH`.
- Prefix commands with `export PATH="/usr/local/go/bin:$PATH"`.

## Commands
- Build:          `go build -o kube-tale ./cmd/kube-tale`
- Test (all):     `go test -race ./...`
- Test (single):  `go test -race ./internal/timeline/`
- Integration:    `go test -race -v -tags=integration ./internal/client/ ./cmd/kube-tale/`
- Vet:            `go vet ./...`
- Lint:           `golangci-lint run` (config: `.golangci.yml`, v2 format)
- Mod tidy:       `go mod tidy`

## CI
- `.github/workflows/ci.yml` — triggers on push to any branch and PR to `main`/`develop`. Runs lint → test (vet, race, coverage) → build → integration (kind cluster).
- `.github/workflows/release.yml` — triggers on tag `v*`. Runs lint + test gates, then goreleaser.
- All PRs must pass CI before merge.

## Branching & Versioning
- `main` — tagged releases only (v0.1.0+). Protected. No direct pushes.
- `develop` — integration branch. All feature PRs merge here.
- `feat/*` — short-lived branches off `develop`.
- Semver starting at `v0.1.0`. Tag triggers goreleaser release via `.goreleaser.yaml`.

## TDD Convention
- Strict Red → Green → Refactor cycle per module.
- Commit prefixes: `feat:`, `fix:`, `test:`, `refactor:`, `chore:`, `ci:`, `docs:`
- Red commit: `test(<module>): add failing test for X`
- Green commit: `feat(<module>): implement X`
- Golden files in each module's `testdata/` directory.

## Modules

| Module | Purpose | Status |
|--------|---------|--------|
| `types`      | Shared Event, EventKind, Timeline structs | Done |
| `client`     | DataSource interface + MockClient + RealClient (client-go) | Done |
| `timeline`   | Merged, sorted sequence of events | Done |
| `story`      | Compressed human-readable narrative | Done |
| `why`        | Pattern-based root-cause scoring | Done |
| `diff`       | State comparison between two points in time | Done |

## Architecture Rules
- All cluster access through `client.DataSource` interface. Never call `client-go` directly from logic packages.
- K8s dependencies (`client-go`, `api`, `apimachinery`) pinned to `v0.35.4` matching k3s cluster.
- `cmd/` uses `flag` package, not `cobra`.
- Events are immutable, self-contained value objects.
- `story.Generate()` is a pure function: `timeline → string`.
- Timeline is `[]Event` sorted by time — no tree or graph.

# kube-tale — Agent Notes

## Project
- Go CLI that correlates Kubernetes signals (events, pod status, ReplicaSet history) into incident narratives.
- **Module:** `github.com/IbraAoad/kube-tale`
- **Go version:** 1.26.2
- **Status:** Bootstrapped — M0 + M1 complete (types, mock client, CI).
- **Entrypoint:** `cmd/kube-tale`

## Environment
- Go binary at `/usr/local/go/bin/go` — may not be in default `$PATH`.
- Prefix commands with `export PATH="/usr/local/go/bin:$PATH"`.

## Commands
- Build:          `go build -o kube-tale ./cmd/kube-tale`
- Test (all):     `go test -race ./...`
- Test (single):  `go test -race ./internal/timeline/`
- Vet:            `go vet ./...`
- Lint:           `golangci-lint run` (config: `.golangci.yml`)
- Mod tidy:       `go mod tidy`

## CI
- `.github/workflows/ci.yml` — triggers on push to any branch and PR to `main`/`develop`. Runs lint → test (vet, race, coverage) → build.
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
| `client`     | DataSource interface + MockClient | Done |
| `timeline`   | Merged, sorted sequence of events | Done |
| `story`      | Compressed human-readable narrative | Planned |
| `why`        | Pattern-based root-cause scoring | Planned |
| `diff`       | State comparison between two points in time | Planned |

## Architecture Rules
- All cluster access through `client.DataSource` interface. Never call `client-go` directly from logic packages.
- No external dependencies beyond Go stdlib until `client-go` is needed (M7).
- `cmd/` uses `flag` package, not `cobra`.
- Events are immutable, self-contained value objects.
- `story.Generate()` is a pure function: `timeline → string`.
- Timeline is `[]Event` sorted by time — no tree or graph.

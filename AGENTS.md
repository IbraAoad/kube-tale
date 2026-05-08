# kube-tale — Agent Notes

## Project
- Go CLI that correlates Kubernetes signals (events, pod status, ReplicaSet history) into incident narratives.
- **Module:** `github.com/IbraAoad/kube-tale`
- **Go version:** 1.26.2 (go.mod declares `go 1.26.2`; supported by golangci-lint-action v9.2.0+)
- **Status:** M0 through M7 complete. Released v0.1.1.
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
- `.github/workflows/ci.yml` — triggers on PR to `main`/`develop` and push to `main`/`develop`. Runs lint → test (vet, race, coverage) → build → integration (kind cluster).
- `.github/workflows/release.yml` — triggers on tag `v*`. Runs lint + test gates, then goreleaser.
- All PRs must pass CI before merge.

## Branching & Versioning
- `main` — tagged releases only (v0.1.0+). Protected. No direct pushes.
- `develop` — integration branch. All feature PRs merge here.
- `feat/*` — short-lived branches off `develop`.
- Semver starting at `v0.1.0`. Tag triggers goreleaser release via `.goreleaser.yaml`.
- **Releases only trigger on `v*` tags.** Docs-only commits (README, AGENTS.md, CODEOWNERS) never trigger releases. No direct pushes to `main` — only `git merge --ff-only develop` after tagging.
- After tagging, fast-forward `main`: `git checkout main && git merge --ff-only develop && git push origin main`

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

## Tooling Notes

### goreleaser v2
- Config uses `version: 2`. Key differences from v1:
  - `archives.formats` (plural) — not `format` (singular, deprecated since v2.6)
  - `changelog.use: github` — cleaner changelog generated from GitHub API
  - `release.prerelease: auto` — auto-detects preview versions
  - `release.footer` — custom message at the bottom of each release
  - `archives.files` — bundles LICENSE + README in each archive
  - `before.hooks: [go mod tidy]` — ensures clean deps before build
  - Archive naming uses `x86_64` / `aarch64` (not `amd64` / `arm64`)

### golangci-lint v2
- Config requires `version: "2"` at the top
- `gofmt` and `goimports` moved from `linters` to `formatters` section
- `gosimple` merged into `staticcheck` — only enable `staticcheck`
- `issues` section renamed to `linters.exclusions`
- Action: `golangci/golangci-lint-action@v9.2.0` (supports Go 1.26)

### GitHub Actions Versions (latest stable)
| Action | Version |
|--------|---------|
| `actions/checkout` | `@v6` |
| `actions/setup-go` | `@v6.4.0` |
| `golangci-lint-action` | `@v9.2.0` |
| `goreleaser-action` | `@v7.2.1` |
| `helm/kind-action` | `@v1.14.0` |

### K8s Dependencies
- `k8s.io/client-go`, `k8s.io/api`, `k8s.io/apimachinery` pinned to `v0.35.4`
- Matches k3s cluster running Kubernetes v1.35.4
- All three modules must use the same minor version

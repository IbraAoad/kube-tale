# Contributing

kube-tale follows strict TDD (Red → Green → Refactor) with golden-file tests.

## Setup

```bash
git clone https://github.com/IbraAoad/kube-tale
export PATH="/usr/local/go/bin:$PATH"
go build -o kube-tale ./cmd/kube-tale
go test -race ./...
```

## Workflow

1. Branch off `develop`: `git checkout -b feat/your-feature develop`
2. Write a failing test, commit it as `test(<module>): description`
3. Implement the minimal fix, commit as `feat(<module>): description`
4. Open a PR targeting `develop`

See `AGENTS.md` for full conventions and architecture rules.

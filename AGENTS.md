# AGENTS

## Purpose

Steering guidelines for Codex contributors working on the Kyn Go CLI.

## Stack

- Language: Go (target `go 1.22+`)
- Binary entrypoint: `cmd/kyn/main.go`
- Internal packages under `internal/`

## Core Rules

1. Keep Kyn stateless and CLI-only for MVP.
2. Preserve deterministic behavior for output ordering and exit codes.
3. Keep path handling slash-normalized and relative to `--cwd`.
4. Prefer small, testable internal packages over large command handlers.

## Exit Codes

- `0`: success
- `1`: rule failure
- `2`: usage/config validation error
- `3`: runtime/provider error

## Coding Standards

1. Run `gofmt -w` on changed Go files.
2. Keep exported APIs minimal.
3. Return concrete errors with actionable context.
4. Prefer table-driven tests.
5. Avoid introducing non-MVP features.

## Suggested Commands

```bash
go test ./...
go vet ./...
go build ./cmd/kyn
```

## MVP Boundaries

Do not implement in MVP:

- daemon/watch modes
- PR comments/integrations
- plugin systems
- monorepo graph adapters (Nx/Rush/Turbo)


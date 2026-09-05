# Testing And Quality Gates

Testing must protect Kyn's behavior, CLI contracts, deterministic output, and integration
boundaries.

## Default Expectations

- Add or update focused table-driven tests for behavior changes.
- Cover parsing, validation, path normalization, exit codes, and provider error paths as relevant.
- Keep fixtures small and explicit.
- Update golden files only when the intended external contract changes, and review their diffs.
- Prefer deterministic assertions over timing-sensitive or environment-dependent checks.

## Before Finishing Work

Run the smallest reliable command that validates the changed area, then broaden in proportion to
risk:

- Format check: `test -z "$(gofmt -l cmd internal)"`
- Lint and vet: `make lint`
- Unit and integration tests: `go test ./...`
- Build: `go build ./cmd/kyn`
- Full local gate: `make ci`

If a command cannot run locally, document why and what risk remains.

## Quality Gates

- No known failing tests introduced by the change.
- No unrelated formatting churn.
- Output ordering and exit codes remain deterministic.
- CLI, configuration, and report contracts are updated deliberately and covered by tests.
- User-facing docs and migration/release notes are updated when behavior changes.
- Changed Go files are formatted with `gofmt -w`.

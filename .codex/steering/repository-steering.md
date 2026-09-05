# Repository Scope And Priorities

This repository builds Kyn, a stateless Go CLI that enforces related-file change policy locally
and in CI.

Primary deliverables:

- the `cmd/kyn` command-line binary;
- deterministic policy evaluation in `internal/`;
- stable human- and machine-readable reports plus supporting documentation.

Core priorities:

- deterministic output ordering and stable exit codes (`0` success, `1` rule failure,
  `2` usage/config validation, and `3` runtime/provider failure);
- slash-normalized paths relative to `--cwd`;
- small, testable internal packages with minimal exported APIs;
- compatible CLI, configuration, and report-format contracts;
- maintainable local and CI workflows.

## Active Boundaries

- `cmd/kyn` owns process startup and delegates behavior to the CLI package.
- `internal/cli` owns command parsing, validation, and exit-code mapping.
- `internal/config`, `internal/changes`, `internal/family`, `internal/matcher`, and
  `internal/rules` own their focused policy-domain responsibilities.
- `internal/report` owns deterministic output serialization and golden fixtures.
- `docs/` and `README.md` own user-facing behavior and workflow documentation.

## Safe Refactor Boundaries

Do not change these without explicit instruction and compatibility evidence:

- CLI flags, command names, and exit-code semantics;
- configuration schema and migration behavior;
- report schemas, field meanings, and deterministic ordering;
- path normalization relative to `--cwd`;
- golden outputs used as external-contract fixtures.

Safe default changes:

- focused improvements within an existing internal package;
- validation and actionable error-context improvements;
- table-driven tests and small explicit fixtures;
- documentation updates that keep canonical behavior synchronized.

## MVP Boundaries

Do not add daemon or watch modes, PR integrations, plugin systems, or monorepo graph adapters.

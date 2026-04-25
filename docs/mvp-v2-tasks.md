# Kyn MVP v2 Task Backlog

## Scope

Execution backlog for shipping v2 with breaking changes, ergonomics upgrades, and CI portability requirements.

## Milestone A: Schema & Engine Core

1. Define v2 config schema (`version: 2`) with `if/assert/actions`.
2. Add parser + validator for v2 schema.
3. Keep v1 parser for transition window.
4. Implement migration command: `kyn config migrate --from v1 --to v2`.
5. Add explicit group model under families.
6. Refactor rule evaluator to support `if/assert/actions`.
7. Remove `emitFlag` from assertions and support `actions.emit`.
8. Add change status model (`added|modified|deleted|renamed`) end-to-end.
9. Add new predicate support for status-aware checks.
10. Add deterministic sorting and output coverage for v2 path.

## Milestone B: CLI Ergonomics

1. Add `--strict-input-mode`.
2. Add auto input-mode behavior for default `kyn check`.
3. Add `--summary-only`.
4. Add `--dry-run-resolve`.
5. Add `kyn init` with starter config generation.
6. Add `kyn explain` for per-rule diagnostics.
7. Improve error messages with expected/observed and next command hints.
8. Rework `--help` output around happy path and examples.
9. Add command docs and examples for stdin and git-default flows.

## Milestone C: Reporting & Integrations

1. Add SARIF reporter.
2. Add reviewdog `rdjson` reporter.
3. Add checkstyle reporter.
4. Keep text/json as stable baselines.
5. Add reporter golden tests and schema fixtures.

## Milestone D: CI Portability & Distribution

1. Build and publish binaries for required targets:
   - linux-amd64 (glibc)
   - linux-arm64 (glibc)
   - linux-amd64 (musl)
   - linux-arm64 (musl)
   - darwin-amd64
   - darwin-arm64
   - windows-amd64
2. Generate and publish SHA256 checksums.
3. Publish container image for CI usage.
4. Add CI smoke tests on multiple Linux base images.
5. Add copy/paste pipeline snippets for all major CI providers.

## Milestone E: Presets & Adoption

1. Ship first preset packs:
   - `web-ui`
   - `api`
   - `proto`
   - `iac`
2. Add preset docs and examples.
3. Add migration guide from v1 to v2.
4. Add troubleshooting guide for common CI/runtime issues.

## Required Gates Before v2 Release

1. Ergonomics gate:
   - New-user time-to-first-pass <= 10 minutes.
2. Portability gate:
   - Verified on all required release targets.
3. Reliability gate:
   - Stable exit codes and deterministic output in all reporters.
4. Migration gate:
   - v1-to-v2 migration tooling works on real sample configs.

## Suggested Execution Order

1. Milestone A
2. Milestone B (in parallel with later A tasks where possible)
3. Milestone C
4. Milestone D
5. Milestone E

## High-Risk Areas

1. Status-aware matching semantics.
2. Migration correctness for complex v1 configs.
3. Reporter compatibility with downstream tooling.
4. Linux distro/runtime compatibility edge cases.

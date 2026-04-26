# Changelog

All notable changes to this project will be documented in this file.

## v0.1.0 - 2026-04-26

First public release of Kyn.

### Added

- `kyn check` for CI policy enforcement over changed file families
- v1 and v2 config support, including `if` / `assert` / `actions`
- family groups and status-aware matching with `changedStatusAny`
- `kyn explain` for per-rule diagnostics
- `kyn init` with starter presets for `web-ui`, `api`, `proto`, and `iac`
- `kyn config migrate --from v1 --to v2`
- text, json, SARIF, reviewdog `rdjson`, and checkstyle reporters
- auto git mode, `--strict-input-mode`, `--summary-only`, and `--dry-run-resolve`
- GitHub Actions CI, release workflow, GoReleaser config, and container packaging
- migration, troubleshooting, CI, release, and preset documentation

### Quality

- deterministic output coverage for CLI and reporter paths
- overall test coverage raised to 85.4%

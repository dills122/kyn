# Changelog

All notable changes to this project will be documented in this file.

## v0.1.3 - 2026-09-05

### Added

- first-run guidance that leads from reviewing a generated policy through previewing,
  enforcing, and diagnosing it, with shell-safe commands for POSIX shells and PowerShell
- practical installation, CI-adoption, and frontend, Go API, and Terraform recipes on the
  documentation site
- automated WinGet manifest generation for Windows `amd64` and `arm64` archives, while keeping
  the other release channels independent of WinGet credentials
- documented, explicitly unapproved candidates for future related-file policy capabilities

### Changed

- empty change sets pass by default; use `--fail-on-empty` when an empty input must fail
- verbose execution diagnostics are written to stderr so JSON, SARIF, RDJSON, and Checkstyle
  stdout remains machine-readable
- `--summary-only` now emits a genuinely compact JSON object and is limited to text and JSON;
  SARIF, RDJSON, and Checkstyle continue to require per-rule diagnostics
- configuration validation now rejects unsupported or conflicting v1/v2 clauses instead of
  silently accepting policy the evaluator cannot enforce
- ambiguous kin templates that resolve differently across source files now fail with an
  actionable error instead of depending on input order
- release containers use GoReleaser's multi-platform Docker v2 pipeline while preserving the
  version, `latest`, and architecture-specific image tags
- local `make ci` now includes the same lint and minimum-coverage gates enforced by CI
- release validation and publishing use the same pinned GoReleaser version, and CI tests both
  Go 1.22 and the current stable Go toolchain
- release artifacts are compiled with the latest patched stable Go toolchain while `go.mod`
  continues to declare Go 1.22 as Kyn's supported minimum
- live documentation now states the current named-group, change-status, and source-anchored
  evaluation boundaries

### Fixed

- expected-file reports now identify only the missing kin from multi-name assertions instead of
  over-reporting every requested kin
- absolute, parent-traversing, and symlink-escaping changed or resolved kin paths are rejected
  when they leave `--cwd`
- Git repository probes and diffs are bounded by execution time and captured output
- Git-provider failures now return runtime/provider exit code `3`, while invalid config,
  change input, and family resolution return usage/config exit code `2`
- end-to-end Git fixtures are isolated from developer commit-signing configuration
- every command renders multi-line help examples with consistent indentation
- documentation deployment also runs when its workflow or pinned dependencies change

## v0.1.2 - 2026-08-21

### Added

- native installation through `go install github.com/dills122/kyn/cmd/kyn@latest`
- a CI release gate that verifies the canonical Go module path

### Fixed

- corrected the module identity and release linker paths so tagged versions can be fetched and installed by the Go toolchain
- report the embedded module version for binaries installed with `go install`, which do not receive GoReleaser linker metadata

## v0.1.1 - 2026-08-21

### Added

- `kyn version` command and `--version` flag, with version/commit/date stamped via ldflags at build and release time
- LICENSE file (MIT), matching the license already advertised in the README badge
- Homebrew cask (`dills122/tap/kyn`) and Scoop bucket (`dills122/scoop-bucket`) publishing on release
- `.deb`/`.rpm`/`.apk` packages built via nfpm and attached to GitHub Releases, plus an optional free push to a Fury.io repo once `FURY_ACCOUNT`/`FURY_TOKEN` secrets are configured

### Changed

- improved release-note generation with grouped changelog categories and curated release header/footer
- refreshed README structure and visual polish, including a dedicated docs-site link
- added a focused GitHub Pages docs site scaffold for quick-start and CI/reference paths

### CI / Docs

- added GitHub Pages deployment workflow to publish static docs to `gh-pages`
- added strict docs validation in CI via `mkdocs build --strict`
- added release-note authoring guidance for consistent, meaningful changelog entries
- migrated `golangci-lint` config and CI job to v2
- updated `docs/decisions.md` to reflect v2 `if`/`assert`/`actions` rule naming

## v0.1.0 - 2026-04-26

First public release of Kyn.

### Added

- `kyn check` for CI policy enforcement over changed file families
- v1 and v2 config support, including `if` / `assert` / `actions`
- source-anchored family groups and status-aware matching for retained Git paths
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

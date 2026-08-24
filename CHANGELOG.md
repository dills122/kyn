# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

### Added

- documented unapproved candidates for deeper Storybook and related-file policy capabilities

### Changed

- clarified current named-group and change-status limitations in the live documentation
- migrated container publishing to GoReleaser's current multi-platform Docker v2 pipeline while preserving architecture-specific image aliases

### Fixed

- kept machine-readable stdout valid under `--verbose` by routing execution diagnostics to stderr
- made root help and post-`init` guidance lead through previewing policy before enforcement, including copyable commands for custom working directories

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

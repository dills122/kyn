# Kyn

[![CI](https://github.com/dills122/kyn/actions/workflows/ci.yml/badge.svg)](https://github.com/dills122/kyn/actions/workflows/ci.yml)
[![Release](https://github.com/dills122/kyn/actions/workflows/release.yml/badge.svg)](https://github.com/dills122/kyn/actions/workflows/release.yml)
[![Docs](https://img.shields.io/badge/docs-github_pages-0A7EA4.svg)](https://dills122.github.io/kyn/)
[![Go Version](https://img.shields.io/badge/go-1.22%2B-00ADD8.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-informational.svg)](LICENSE)

Kyn is a stateless Go CLI for enforcing related-file change policy in CI.

It answers questions like:
- If source changed, did tests/stories/specs change too?
- If a sidecar exists, should a CI flag be emitted?
- Are file-family rules enforced consistently across repos?

## Why Teams Use Kyn

- Deterministic output suitable for CI parsing and diffing
- Stable exit codes (`0` pass, `1` policy fail, `2` usage/config, `3` runtime)
- Fast local and CI runs with no daemon, service, or plugin system
- Config-driven policy instead of brittle shell glue

## Install / Build

```bash
go build -o ./bin/kyn ./cmd/kyn
./bin/kyn --help
```

## Quick Start

```bash
./bin/kyn check \
  --cwd testdata/angular \
  -c kyn.config.yaml \
  -f libs/ui/button/button.component.ts,libs/ui/button/button.component.html
```

## Core Commands

```bash
# CI baseline
kyn check -c kyn.config.yaml --base origin/main --head HEAD -o json

# Auto git mode (default when no input mode is provided)
kyn check -c kyn.config.yaml -o json

# Explain per-rule diagnostics
kyn explain -c kyn.config.yaml --base origin/main --head HEAD

# Bootstrap starter config
kyn init --preset web-ui

# Migrate config v1 -> v2 safely
kyn config migrate -c kyn.config.yaml --from v1 --to v2
```

## Output Formats

- `text`
- `json`
- `sarif`
- `rdjson`
- `checkstyle`

## Documentation

- [Docs Site](https://dills122.github.io/kyn/)
- [Docs Index](docs/README.md)
- [Specification](docs/spec.md)
- [CI Guide](docs/ci.md)
- [Release Guide](docs/release.md)
- [Presets](docs/presets.md)
- [Migration v1 -> v2](docs/migration-v1-to-v2.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Changelog](CHANGELOG.md)

## Project Scope

Kyn intentionally stays focused:
- Stateless CLI only
- Deterministic behavior and stable contracts
- No daemon/watch mode, no plugin system, no PR API integrations

## Development

```bash
make hooks
make fmt
make lint
make test
make vet
make build
make e2e
```

`make e2e` builds the real `kyn` binary and runs it against realistic
example projects under [`e2e/projects`](e2e/projects) — see
[`e2e/README.md`](e2e/README.md).

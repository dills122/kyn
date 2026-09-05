# Kyn

[![CI](https://github.com/dills122/kyn/actions/workflows/ci.yml/badge.svg)](https://github.com/dills122/kyn/actions/workflows/ci.yml)
[![Release](https://github.com/dills122/kyn/actions/workflows/release.yml/badge.svg)](https://github.com/dills122/kyn/actions/workflows/release.yml)
[![Docs](https://img.shields.io/badge/docs-github_pages-0A7EA4.svg)](https://dills122.github.io/kyn/)
[![Go Version](https://img.shields.io/badge/go-1.22%2B-00ADD8.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-informational.svg)](LICENSE)

Kyn is a stateless CLI that turns related-file review conventions into
deterministic local and CI policy.

It answers questions like:

- A component changed—did its existing Storybook story change too?
- An API handler changed—was its test included in the same change?
- A Terraform module changed—should a missing README update warn or block?
- A design sidecar exists—should CI emit a follow-up flag?

## Why teams use Kyn

- Replace repeated review comments and brittle shell scripts with readable YAML.
- Get the same sorted results and stable exit codes locally and in CI.
- Start with advisory warnings, then tighten policy when the signal proves useful.
- Use text, JSON, SARIF, RDJSON, or Checkstyle without a daemon or hosted service.

## Install a released version

Package channels prepared for **v0.1.3 (2026-09-05)**.

| Environment | Recommended command |
| --- | --- |
| macOS | `brew tap dills122/tap && brew install --cask dills122/tap/kyn` |
| Linux with Homebrew | `brew tap dills122/tap && brew install --cask dills122/tap/kyn` |
| Go 1.22+ | `go install github.com/dills122/kyn/cmd/kyn@v0.1.3` |
| Windows | Add the [Scoop bucket](https://github.com/dills122/scoop-bucket), then `scoop install kyn` |
| Container CI (GitHub Packages / GHCR) | `docker pull ghcr.io/dills122/kyn:0.1.3` |

Kyn also publishes checksummed
[GitHub Release](https://github.com/dills122/kyn/releases/latest) archives and
native DEB/RPM/APK packages through public Fury repositories. The permanent
WinGet ID is `DylanSteele.Kyn`; its
[initial catalog submission](https://github.com/microsoft/winget-pkgs/pull/422458)
is still under Microsoft review.

See **[Install Kyn](https://dills122.github.io/kyn/install/)** for every command,
upgrade instructions, architecture support, Fury's unsigned-repository caveat,
and the current recommendation for each platform.

## Quick start

Generate the closest starter policy:

```bash
cd your-repository
kyn init --preset web-ui
```

Available presets are `web-ui`, `api`, `proto`, and `iac`. Adjust the
generated globs to match your repository, then preview the relationships:

```bash
kyn check -c kyn.config.yaml \
  --files path/to/example.component.ts \
  --dry-run-resolve
```

Run policy against a real change set:

```bash
kyn check -c kyn.config.yaml \
  --files path/to/example.component.ts,path/to/example.stories.ts \
  --show-passes
```

Inside a Git repository, `kyn check -c kyn.config.yaml` automatically compares
`origin/main...HEAD`. CI can keep the refs explicit:

```bash
kyn check -c kyn.config.yaml \
  --base origin/main \
  --head HEAD \
  --format json
```

## How a policy reads

```yaml
version: 2

families:
  - id: go-handler
    groups:
      source:
        include: ["internal/**/*_handler.go"]
    kin:
      test: "{dir}/{name}_test.go"

rules:
  - id: handler-tests-sync
    family: go-handler
    severity: error
    if:
      changedAny: [source]
      kinExists: [test]
    assert:
      kinChanged: [test]
    message: "Handler changed but its existing test file did not."
```

Kyn matches changed handler files, resolves the corresponding test path, and
exits 1 when that existing test is missing from the change set. See
[Configure families and rules](https://dills122.github.io/kyn/config/) for the
complete model.

## Core commands

```bash
# Evaluate policy
kyn check -c kyn.config.yaml

# Explain why rules apply, pass, or fail
kyn explain -c kyn.config.yaml --base origin/main --head HEAD

# Generate a starter config
kyn init --preset api

# Migrate v1 YAML to v2
kyn config migrate -c kyn.config.yaml --from v1 --to v2
```

Kyn's stable process contract is:

- `0`: policy passed at the selected threshold
- `1`: rule failure
- `2`: invalid usage or config
- `3`: runtime/provider failure

## Documentation

- [Getting started](https://dills122.github.io/kyn/getting-started/)
- [How Kyn works](https://dills122.github.io/kyn/concepts/)
- [Configuration and rules](https://dills122.github.io/kyn/config/)
- [Real-world recipes](https://dills122.github.io/kyn/recipes/frontend/)
- [CI adoption](https://dills122.github.io/kyn/ci/)
- [CLI and exit codes](https://dills122.github.io/kyn/reference/)
- [Troubleshooting](https://dills122.github.io/kyn/troubleshooting/)
- [What's new in v0.1.3](https://dills122.github.io/kyn/releases/v0.1.3/)
- [Release runbook](docs/release.md)
- [Changelog](CHANGELOG.md)

## Project scope

Kyn intentionally stays focused:

- Stateless CLI only
- Deterministic behavior and stable contracts
- No daemon/watch mode, plugin system, or PR API integrations

## Development

```bash
make hooks
make ci
```

`make ci` runs formatting checks, `golangci-lint`, vet, unit tests, the
minimum-coverage check, a build, and the black-box scenarios under
[`e2e/projects`](e2e/projects) — the same checks the GitHub Actions `lint`,
`test`, and `coverage` jobs enforce, so a passing `make ci` locally means
those jobs will pass too. It does not cover the release-config, docs-site,
cross-platform build-matrix, or containerized smoke-test jobs, which need
`goreleaser`, `mkdocs`, and Docker/QEMU. Install
[`golangci-lint`](https://golangci-lint.run/welcome/install/) before running
`make ci` or `make lint`.

## License

Kyn is available under the [MIT License](LICENSE).

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

## Install a released version

Release channels last verified: **2026-08-21**.

| Channel | Platforms | Status |
| --- | --- | --- |
| [Homebrew tap](https://github.com/dills122/homebrew-tap) | macOS, Linux | Available |
| [GitHub Releases](https://github.com/dills122/kyn/releases/latest) | macOS, Linux, Windows | Available |
| [GHCR (GitHub Packages)](https://github.com/dills122/kyn/pkgs/container/kyn) | Linux containers (`amd64`, `arm64`) | Available |
| [Scoop bucket](https://github.com/dills122/scoop-bucket) | Windows | Available |
| Fury package repository | Debian, RPM, Alpine | Coming soon |

### Homebrew (macOS/Linux)

Kyn currently ships through the `dills122/tap` third-party tap:

```bash
brew tap dills122/tap
brew install --cask dills122/tap/kyn
kyn --version
```

### GitHub Releases

Each [GitHub Release](https://github.com/dills122/kyn/releases/latest) includes:

- `.tar.gz` archives for macOS and Linux (`amd64` and `arm64`)
- `.zip` archives for Windows (`amd64` and `arm64`)
- `.deb`, `.rpm`, and `.apk` Linux packages
- `checksums.txt` for SHA-256 verification

Download the archive for your platform and `checksums.txt`, then verify and
install it. This example uses the `v0.1.1` Apple Silicon archive; change `VERSION`,
`OS`, and `ARCH` for the release and platform you downloaded:

```bash
VERSION=0.1.1
OS=darwin
ARCH=arm64
ARCHIVE="kyn_${VERSION}_${OS}_${ARCH}.tar.gz"
CHECKSUM_LINE="$(grep "  ${ARCHIVE}$" checksums.txt)"

if [ -z "$CHECKSUM_LINE" ]; then
  echo "checksum not found for $ARCHIVE" >&2
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  printf '%s\n' "$CHECKSUM_LINE" | sha256sum -c -
else
  printf '%s\n' "$CHECKSUM_LINE" | shasum -a 256 -c -
fi

tar -xzf "$ARCHIVE"
sudo install -m 0755 kyn /usr/local/bin/kyn
kyn --version
```

Linux users can instead install the downloaded native package:

```bash
# Debian/Ubuntu
sudo dpkg -i kyn_*_linux_amd64.deb

# Fedora/RHEL
sudo rpm -i kyn_*_linux_amd64.rpm

# Alpine
sudo apk add --allow-untrusted kyn_*_linux_amd64.apk
```

### Container image (GHCR / GitHub Packages)

Use `latest` for the newest release or a version tag such as `0.1.1` for a
repeatable installation:

```bash
docker pull ghcr.io/dills122/kyn:latest
docker run --rm ghcr.io/dills122/kyn:latest --version

# Pin an exact release for reproducible CI usage.
docker run --rm ghcr.io/dills122/kyn:0.1.1 --version
```

### Scoop (Windows)

```powershell
scoop bucket add dills122 https://github.com/dills122/scoop-bucket
scoop install kyn
kyn --version
```

### Fury Linux repository (coming soon)

Hosted `apt`, `dnf`/`yum`, and `apk` repository instructions will be added after
the Fury publishing credentials are fixed and the packages are verified. Until
then, install the `.deb`, `.rpm`, or `.apk` attached to the GitHub Release.

### Build from source

This builds the current checkout rather than installing a published release:

```bash
go build -o ./bin/kyn ./cmd/kyn
./bin/kyn --version
```

## Quick Start

```bash
kyn check \
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

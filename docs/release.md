# Release and Installation

This document describes the v2 distribution story for Kyn.

## Release Outputs

Tagged releases publish:

- a Go module installable with `go install github.com/dills122/kyn/cmd/kyn@<version>`
- GitHub release archives for Linux, macOS, and Windows
- `checksums.txt` with SHA256 sums
- GHCR container images
- a Homebrew cask in `dills122/homebrew-tap`
- `.deb`, `.rpm`, and `.apk` packages to the public `dsteele` Fury account

## Linux Compatibility

Kyn is built as a static Go binary with `CGO_ENABLED=0`.

That means the Linux release artifacts are intended to work across:

- glibc-based distros such as Debian and Ubuntu
- musl-based distros such as Alpine

CI currently smoke-tests the Linux binary on:

- Debian (`linux/amd64`, `linux/arm64`)
- Alpine (`linux/amd64`, `linux/arm64`)

## Container Image

The container image is built from a distroless static base and exposes the CLI as the entrypoint:

```bash
docker run --rm ghcr.io/<owner>/kyn:latest --help
```

Example CI usage:

```bash
docker run --rm \
  -v "$PWD:/work" \
  -w /work \
  ghcr.io/<owner>/kyn:latest \
  check -c kyn.config.yaml --base origin/main --head HEAD --format json
```

## Installing a Release Binary

With Go 1.22 or newer, install a released version directly from the module:

```bash
go install github.com/dills122/kyn/cmd/kyn@latest
```

Use an explicit tag such as `@v0.1.2` for reproducible installation. This path
depends on `go.mod` declaring the public repository as the canonical module path;
the `make module-path` release gate prevents that identity from drifting.

To install a prebuilt archive instead:

Typical manual install flow:

```bash
curl -L -o kyn.tar.gz <release-archive-url>
curl -L -o checksums.txt <checksums-url>
sha256sum -c checksums.txt
tar -xzf kyn.tar.gz
./kyn --help
```

## Creating a Release

Releases are driven by Git tags:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow uses GoReleaser to:

- build archives
- generate checksums
- publish GitHub release artifacts
- publish GHCR images and manifests
- update the Homebrew tap

After GoReleaser succeeds, a separate job downloads the Linux packages from the
GitHub release and publishes them to Fury. Keeping Fury separate makes a failed
upload retryable without rebuilding or republishing an existing release.

## Package Manager Credentials

Configure these repository secrets under **Settings > Secrets and variables >
Actions**:

- `TAP_GITHUB_TOKEN`: a fine-grained GitHub personal access token with
  `Contents: Read and write` access to `dills122/homebrew-tap`.
- `FURY_ACCOUNT`: the exact Fury account username from the account's push URL.
- `FURY_TOKEN`: an upload-capable token assigned to that Fury account. A
  read-only deploy token cannot publish packages.

The Fury API authenticates with `FURY_ACCOUNT:FURY_TOKEN` and publishes to the
account named by `FURY_ACCOUNT`. A `403` response means that token is not allowed
to upload to that account; recreate an upload-capable token or correct the
account username.

## Retry a Fury Upload

Do not create a new version tag just to retry Fury. In GitHub:

1. Open **Actions > Release**.
2. Choose **Run workflow**.
3. Enter the existing release tag, for example `v0.1.1`.
4. Run the workflow.

The manual run downloads the existing `.deb`, `.rpm`, and `.apk` release assets
and publishes only those packages to Fury.

## Verify Homebrew

The Homebrew cask is published independently by GoReleaser. Verify it with:

```bash
brew tap dills122/tap
brew install --cask dills122/tap/kyn
kyn --version
```

## Verify Fury

The public repositories are:

- APT: `https://apt.fury.io/dsteele/`
- YUM/DNF: `https://yum.fury.io/dsteele/`
- Alpine: `https://alpine.fury.io/dsteele/`

Verify each release by installing `kyn` through real Debian, Fedora, and Alpine
clients and running `kyn --version`. Repository signing is not configured yet,
so clients must explicitly trust these repositories. The README contains the
current commands and calls out that security tradeoff.

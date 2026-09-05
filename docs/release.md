# Release and Installation

This document describes the v2 distribution story for Kyn.

## Release Outputs

Tagged releases publish:

- a Go module installable with `go install github.com/dills122/kyn/cmd/kyn@<version>`
- GitHub release archives for Linux, macOS, and Windows
- `checksums.txt` with SHA256 sums
- GHCR container images
- a Homebrew cask in `dills122/homebrew-tap`
- a Scoop manifest in `dills122/scoop-bucket`
- a WinGet manifest PR from `dills122/winget-pkgs` to `microsoft/winget-pkgs`
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
docker run --rm ghcr.io/dills122/kyn:0.1.2 --help
```

Example CI usage:

```bash
git diff --name-only origin/main...HEAD \
  | docker run --rm -i \
  -v "$PWD:/work" \
  -w /work \
  ghcr.io/dills122/kyn:0.1.2 \
  check -c kyn.config.yaml --stdin --format json
```

The distroless image does not include Git. Compute changed paths on the host,
mount the repository for config and kin existence checks, and use stdin,
`--files`, or `--files-from`.

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
VERSION=0.1.2
OS=darwin
ARCH=arm64
ARCHIVE="kyn_${VERSION}_${OS}_${ARCH}.tar.gz"

curl -LO "https://github.com/dills122/kyn/releases/download/v${VERSION}/${ARCHIVE}"
curl -LO "https://github.com/dills122/kyn/releases/download/v${VERSION}/checksums.txt"
grep "  ${ARCHIVE}$" checksums.txt | shasum -a 256 -c -

tar -xzf "${ARCHIVE}"
./kyn --help
```

On Linux, replace `shasum -a 256` with `sha256sum`. Filtering the checksum
manifest to the downloaded archive avoids false failures for release assets
that are not present locally.

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
- update the Scoop bucket
- generate WinGet manifests and open a community-catalog pull request when the
  WinGet token is configured

After GoReleaser succeeds, a separate job downloads the Linux packages from the
GitHub release and publishes them to Fury. Keeping Fury separate makes a failed
upload retryable without rebuilding or republishing an existing release.

## Package Manager Credentials

Configure these repository secrets under **Settings > Secrets and variables >
Actions**:

- `TAP_GITHUB_TOKEN`: a fine-grained GitHub personal access token with
  `Contents: Read and write` access to `dills122/homebrew-tap` and
  `dills122/scoop-bucket`.
- `WINGET_GITHUB_TOKEN`: a GitHub personal access token that can push version
  branches to the public `dills122/winget-pkgs` fork and open pull requests
  against `microsoft/winget-pkgs`. Keep this separate from the release
  repository's default `GITHUB_TOKEN`, which cannot write across repositories.
- `FURY_ACCOUNT`: the exact Fury account username from the account's push URL.
- `FURY_TOKEN`: an upload-capable token assigned to that Fury account. A
  read-only deploy token cannot publish packages.

The Fury API authenticates with `FURY_ACCOUNT:FURY_TOKEN` and publishes to the
account named by `FURY_ACCOUNT`. A `403` response means that token is not allowed
to upload to that account; recreate an upload-capable token or correct the
account username.

## WinGet Publishing

The permanent package identifier is `DylanSteele.Kyn`. The one-time setup is:

1. Fork `microsoft/winget-pkgs` to `dills122/winget-pkgs`.
2. Add the fork/PR-capable token as the `WINGET_GITHUB_TOKEN` repository secret.

The initial v0.1.2 catalog submission is
[microsoft/winget-pkgs#422458](https://github.com/microsoft/winget-pkgs/pull/422458).

On each stable tag, GoReleaser generates portable manifests from the Windows
`amd64` and `arm64` zip archives, pushes a version-specific branch to the fork,
and opens a pull request against Microsoft's `master` branch. If the secret is
absent, the other release channels continue and only the WinGet upload is
skipped. Microsoft validates each manifest and installer and may conduct a
manual review before the version appears in the community catalog.

After catalog admission, verify a release from Windows PowerShell:

```powershell
winget install --id DylanSteele.Kyn --exact
kyn --version
```

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

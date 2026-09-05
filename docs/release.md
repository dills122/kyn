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
docker run --rm ghcr.io/dills122/kyn:0.1.3 --help
```

Example CI usage:

```bash
git diff --name-only origin/main...HEAD \
  | docker run --rm -i \
  -v "$PWD:/work" \
  -w /work \
  ghcr.io/dills122/kyn:0.1.3 \
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

Use an explicit tag such as `@v0.1.3` for reproducible installation. This path
depends on `go.mod` declaring the public repository as the canonical module path;
the `make module-path` release gate prevents that identity from drifting.

To install a prebuilt archive instead:

Typical manual install flow:

```bash
VERSION=0.1.3
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

## Pre-Tag Checklist

Run the complete release gate from the release-preparation branch before merging:

```bash
make ci
goreleaser check
GITHUB_REPOSITORY_OWNER=dills122 goreleaser release --snapshot --clean
mkdocs build --strict --site-dir /tmp/kyn-site
actionlint .github/workflows/ci.yml .github/workflows/docs-pages.yml \
  .github/workflows/release.yml
git diff --check origin/main...HEAD
```

The snapshot must contain macOS, Linux, and Windows archives, `checksums.txt`, and
DEB, RPM, and APK packages. Snapshot mode must not publish a GitHub release, package,
container, tap, bucket, or catalog pull request.

Before tagging, also confirm:

- the version and date are closed in `CHANGELOG.md`, and the matching release-notes page
  describes upgrade-impacting behavior;
- the candidate PR is merged, required `main` checks are green, and local `main` exactly
  matches `origin/main` with no working-tree changes;
- the release workflow pins the latest patched stable Go toolchain (Go 1.27.1 for v0.1.3);
  check the official Go release history and vulnerability database before changing that pin.
  `go.mod` declares the supported minimum and must not select the release compiler;
- no `v0.1.3` tag or GitHub release already exists;
- `TAP_GITHUB_TOKEN`, `FURY_ACCOUNT`, and `FURY_TOKEN` are configured, and
  `WINGET_GITHUB_TOKEN` is configured if this release should open a WinGet PR;
- the current unsigned-repository warning for Fury and the pending WinGet admission status
  are still accurate.

GitHub does not reveal secret values after they are stored. Confirm their presence and least
required access without copying credentials into logs or issue comments.

## Creating a Release

Releases are driven by Git tags.

For v0.1.3, merge the release-preparation pull request and wait for every required
`main` check to pass. Then create the tag from the exact, clean `origin/main` commit:

```bash
git fetch --prune origin
git checkout main
git pull --ff-only origin main
test -z "$(git status --porcelain)"
test "$(git rev-parse HEAD)" = "$(git rev-parse origin/main)"
git tag v0.1.3
git push origin v0.1.3
```

Do not tag the release branch or move an existing published tag.

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

## Post-Release Verification

Treat the release as complete only after checking every configured channel:

1. Confirm the Release workflow completed and the GitHub release contains all archives,
   native packages, and `checksums.txt`.
2. Download one archive and verify its checksum; run `kyn --version` and `kyn --help`.
3. Install `github.com/dills122/kyn/cmd/kyn@v0.1.3` with Go 1.22 or newer and verify
   the reported version.
4. Pull the `0.1.3`, `0.1.3-amd64`, and `0.1.3-arm64` GHCR tags and smoke-test the
   matching architectures.
5. Verify the Homebrew cask, Scoop manifest, and public Fury packages report v0.1.3.
6. If WinGet publishing was enabled, confirm that a v0.1.3 catalog pull request was opened.
   Until the initial package is admitted, continue recommending Scoop in user documentation.

After those checks pass, change wording such as "prepared" to "verified" if any release page
still uses it and record the verification date.

## Failure and Recovery

- Do not blindly rerun a failed GoReleaser job. First inspect the GitHub release and every
  configured external channel to determine what was already published. If nothing was created,
  GitHub's **Re-run failed jobs** action is safe; if any asset or channel was published, do not
  rerun the whole GoReleaser job against the same immutable tag.
- If only Fury fails, use the retry procedure below; it reuses the existing GitHub release
  packages and does not republish other channels.
- If GoReleaser partially publishes otherwise valid artifacts, preserve the tag and release,
  record the affected channels, and repair only a channel with a documented, channel-specific
  procedure. If no such procedure exists, publish the correction in the next patch release.
- If a source, archive, checksum, or already-published package is wrong, do not overwrite the
  immutable v0.1.3 artifacts. Document the affected channel, fix forward on `main`, and publish
  the next patch version.
- If external package-manager review is delayed, keep the GitHub release available and document
  the delayed channel rather than rebuilding the same version.

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

The current v0.1.3 catalog submission is
[microsoft/winget-pkgs#430084](https://github.com/microsoft/winget-pkgs/pull/430084).
It supersedes the closed initial v0.1.2 submission.

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
3. Enter the existing release tag, for example `v0.1.3`.
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

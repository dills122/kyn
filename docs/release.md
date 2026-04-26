# Release and Installation

This document describes the v2 distribution story for Kyn.

## Release Outputs

Tagged releases publish:

- GitHub release archives for Linux, macOS, and Windows
- `checksums.txt` with SHA256 sums
- GHCR container images

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

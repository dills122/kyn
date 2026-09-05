# Release Operations

This page is for Kyn maintainers. To install Kyn, use the
[installation guide](install.md).

Releases publish:

- A Go module installable with `go install`
- Multi-OS archives
- SHA256 checksums
- GHCR images
- A Homebrew cask and Scoop manifest
- A WinGet community-catalog manifest PR
- DEB, RPM, and APK packages in the public `dsteele` Fury repositories

Releases are tag-driven:

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

For the current release's highlights and compatibility notes, see
[Kyn v0.1.3](releases/v0.1.3.md).

The workflow updates Homebrew and Scoop, opens a WinGet catalog PR when its
token is configured, publishes GHCR images and GitHub assets, then uploads Linux
packages to Fury. See the repository
[release runbook](https://github.com/dills122/kyn/blob/main/docs/release.md) for
credentials, retry behavior, and channel verification.

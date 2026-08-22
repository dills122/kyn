# Release

Releases publish:

- A Go module installable with `go install`
- Multi-OS archives
- SHA256 checksums
- GHCR images
- A Homebrew cask and Scoop manifest
- DEB, RPM, and APK packages

Tag-driven release:

```bash
git tag v0.1.2
git push origin v0.1.2
```

More details: [docs/release.md](../release.md)

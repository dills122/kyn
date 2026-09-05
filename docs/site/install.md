# Install Kyn

Kyn is a single executable. Choose the channel that fits how your machine or CI
environment already manages tools.

## Recommended by platform

| Environment | Recommended channel | Why |
| --- | --- | --- |
| macOS | Homebrew | Native upgrades and both Apple Silicon and Intel support |
| Linux with Homebrew | Homebrew | Native upgrades on `amd64` and `arm64` |
| Linux with Go 1.22+ | Pinned `go install` | Simple, reproducible, and does not add a repository |
| Windows | Scoop | Available now for x64 and ARM64; WinGet review is still pending |
| Container CI | Pinned GHCR image | No host installation and an immutable version tag |
| Locked-down environment | Checksummed GitHub Release | Inspect and verify the exact artifact before installation |

!!! tip "Pin releases in CI"
    Examples below use v0.1.3, the release verified on 2026-09-05. Replace it when you
    intentionally upgrade. Avoid `latest` in reproducible jobs.

## Homebrew

Recommended for macOS and for Linux machines that already use Homebrew:

```bash
brew tap dills122/tap
brew install --cask dills122/tap/kyn
kyn --version
```

Upgrade later with:

```bash
brew upgrade --cask dills122/tap/kyn
```

Kyn's tap contains release archives for macOS and Linux on both `amd64` and
`arm64`.

## Go toolchain

With Go 1.22 or newer:

```bash
go install github.com/dills122/kyn/cmd/kyn@v0.1.3
kyn --version
```

Go places the binary in `GOBIN`, or `GOPATH/bin` when `GOBIN` is unset.
Put that directory on your `PATH`. Use `@latest` only when you deliberately
want the newest release.

## Windows

### Scoop — recommended today

```powershell
scoop bucket add dills122 https://github.com/dills122/scoop-bucket
scoop install kyn
kyn --version
```

Upgrade with `scoop update kyn`.

### WinGet — catalog review pending

Kyn's permanent package ID is `DylanSteele.Kyn`. Its
[current community-catalog submission](https://github.com/microsoft/winget-pkgs/pull/430084)
is under Microsoft review. After that PR is admitted:

```powershell
winget install --id DylanSteele.Kyn --exact
winget upgrade --id DylanSteele.Kyn --exact
kyn --version
```

Until the catalog entry is live, use Scoop or a Windows zip from GitHub Releases.

## Container image (GitHub Packages / GHCR)

[GitHub Container Registry](https://github.com/dills122/kyn/pkgs/container/kyn)
publishes Linux images for `amd64` and `arm64`:

```bash
docker pull ghcr.io/dills122/kyn:0.1.3
docker run --rm ghcr.io/dills122/kyn:0.1.3 --version
```

To check the current repository, compute the diff on the host and pipe the paths
to Kyn:

```bash
git diff --name-only origin/main...HEAD \
  | docker run --rm -i \
  -v "$PWD:/work" \
  -w /work \
  ghcr.io/dills122/kyn:0.1.3 \
  check -c kyn.config.yaml --stdin
```

The image is deliberately minimal and does not contain Git. Mount the repository
for config and kin existence checks, then use stdin, `--files`, or a mounted
`--files-from` list.

## GitHub Release archives

Every [GitHub Release](https://github.com/dills122/kyn/releases/latest) includes:

- `.tar.gz` archives for macOS and Linux on `amd64` and `arm64`
- `.zip` archives for Windows on `amd64` and `arm64`
- DEB, RPM, and APK packages
- `checksums.txt` with SHA-256 hashes

Download the archive and `checksums.txt`, then verify the exact filename:

```bash
VERSION=0.1.3
OS=darwin
ARCH=arm64
ARCHIVE="kyn_${VERSION}_${OS}_${ARCH}.tar.gz"

curl -LO "https://github.com/dills122/kyn/releases/download/v${VERSION}/${ARCHIVE}"
curl -LO "https://github.com/dills122/kyn/releases/download/v${VERSION}/checksums.txt"
grep "  ${ARCHIVE}$" checksums.txt | shasum -a 256 -c -

tar -xzf "${ARCHIVE}"
sudo install -m 0755 kyn /usr/local/bin/kyn
kyn --version
```

On Linux, replace `shasum -a 256` with `sha256sum` when needed. On Windows,
use `Get-FileHash -Algorithm SHA256` and compare the result with
`checksums.txt`.

## Fury Linux repositories

Fury provides APT, DNF/YUM, and Alpine packages under the public `dsteele`
account. The repositories are currently unsigned, so these commands explicitly
trust them. Prefer a checksummed GitHub Release if your environment requires
signed repository metadata.

### Debian and Ubuntu

```bash
sudo apt-get update
sudo apt-get install -y ca-certificates
echo 'deb [trusted=yes] https://apt.fury.io/dsteele/ * *' \
  | sudo tee /etc/apt/sources.list.d/kyn-fury.list >/dev/null
sudo apt-get update
sudo apt-get install -y kyn
kyn --version
```

### Fedora and RHEL

```bash
sudo tee /etc/yum.repos.d/kyn-fury.repo >/dev/null <<'EOF'
[fury-kyn]
name=Gemfury Kyn
baseurl=https://yum.fury.io/dsteele/
enabled=1
gpgcheck=0
EOF
sudo dnf install -y kyn
kyn --version
```

### Alpine

```bash
echo 'https://alpine.fury.io/dsteele/' \
  | sudo tee -a /etc/apk/repositories >/dev/null
sudo apk update --allow-untrusted
sudo apk add --allow-untrusted kyn
kyn --version
```

## Build the current checkout

This is for contributors. It does not install an immutable release:

```bash
go build -o ./bin/kyn ./cmd/kyn
./bin/kyn --version
```

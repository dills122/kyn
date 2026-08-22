# Getting Started

## Install

With Go 1.22 or newer:

```bash
go install github.com/dills122/kyn/cmd/kyn@latest
kyn --version
```

For Homebrew on macOS or Linux:

```bash
brew tap dills122/tap
brew install --cask dills122/tap/kyn
kyn --version
```

See the repository [release and installation guide](../release.md) for GitHub
Release archives, containers, Scoop, checksums, and Linux packages.

## Build the current checkout

```bash
go build -o ./bin/kyn ./cmd/kyn
```

## First Run

```bash
./bin/kyn check \
  --cwd testdata/angular \
  -c kyn.config.yaml \
  -f libs/ui/button/button.component.ts,libs/ui/button/button.component.html
```

## Helpful Commands

```bash
kyn check --help
kyn explain --help
kyn init --help
kyn config migrate --help
```

## Presets

Generate starter config:

```bash
kyn init --preset web-ui
```

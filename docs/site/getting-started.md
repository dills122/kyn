# Getting Started

## Build

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

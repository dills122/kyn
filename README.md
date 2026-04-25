# Kyn

Kyn is a lightweight CLI for detecting related file changes and enforcing file relationship rules in CI.

## Current Status

Repository scaffold is in place with a basic `kyn check` command parser.

## Quick Start

```bash
go build ./cmd/kyn
./kyn check --files a.ts --config kyn.config.yaml
```

## Development

```bash
make fmt
make test
make vet
make build
```

## Project Layout

```txt
cmd/kyn            CLI binary entrypoint
internal/cli       Command parsing and validation
internal/*         Reserved packages for MVP engine components
testdata/          Fixture test inputs
```


# Kyn

Kyn is a lightweight CLI for detecting related file changes and enforcing file relationship rules in CI.

## Current Status

MVP foundation is implemented:

- `kyn check` command
- YAML config loading and validation
- changed-file providers (`--files`, `--files-from`, git diff)
- family resolution and kin template expansion
- rule evaluation (`when`, `require`, `emitFlag`)
- text and JSON reporting
- deterministic output ordering
- stable exit code behavior

## Quick Start

```bash
go build ./cmd/kyn
./kyn check -c testdata/angular/kyn.config.yaml -f libs/ui/button/button.component.ts,libs/ui/button/button.component.html --cwd testdata/angular
```

## Development

```bash
make fmt
make test
make vet
make build
```

## CLI

```bash
kyn check \
  -c, --config <path> \
  [-f, --files <csv> | --files-from <path> | --stdin | --base <ref> --head <ref>] \
  [--cwd <path>] \
  [-o, --format text|json] \
  [--fail-on error|warn] \
  [--fail-on-empty] \
  [--show-passes] \
  [--verbose]
```

Exactly one change input mode is allowed.

## Exit Codes

- `0`: success
- `1`: one or more rule failures
- `2`: invalid CLI usage or invalid config
- `3`: runtime/provider error (for example git execution failure)

## Example Commands

```bash
# explicit files
go run ./cmd/kyn check \
  --cwd testdata/angular \
  -c kyn.config.yaml \
  -f libs/ui/button/button.component.ts,libs/ui/button/button.component.html

# files-from input
go run ./cmd/kyn check \
  --cwd testdata/angular \
  -c kyn.config.yaml \
  --files-from changed-files.txt

# stdin input (for piped changed file lists)
git diff --name-only origin/main...HEAD | go run ./cmd/kyn check \
  --cwd . \
  -c kyn.config.yaml \
  --stdin

# json output
go run ./cmd/kyn check \
  --cwd testdata/angular \
  -c kyn.config.yaml \
  -f libs/ui/button/button.component.ts,libs/ui/button/button.component.html \
  -o json
```

## Project Layout

```txt
cmd/kyn            CLI binary entrypoint
internal/cli       Command parsing and validation
internal/*         Reserved packages for MVP engine components
testdata/          Fixture test inputs
```

## CI Example

```bash
go build -o ./bin/kyn ./cmd/kyn
./bin/kyn check -c ./kyn.config.yaml --base origin/main --head HEAD -o json
```

See DevOps-focused guidance: [CI Guide](/Users/dsteele/go/src/kyn/docs/ci.md)

# Kyn

Kyn is a stateless Go CLI that enforces related-file rules for changed files in CI and local workflows.

## Why Kyn

Kyn answers policy questions like:

- If a component changed, did its Storybook story change?
- If a component has a Figma sidecar, should a publish flag be emitted?
- If required sibling files exist, were they updated when source changed?

## Features

- `kyn check` command for policy evaluation
- YAML config with schema validation
- Multiple change input modes:
  - `--files`
  - `--files-from`
  - `--stdin` (alias for `--files-from -`)
  - `--base` + `--head` (git diff)
- Family/kin resolution via glob + templates
- Rule evaluation with `when` and `require` clauses
- Deterministic text and JSON output
- Stable exit codes for CI integration

## Install / Build

```bash
go build -o ./bin/kyn ./cmd/kyn
```

Run help:

```bash
./bin/kyn check --help
```

## Quick Start

Run against included fixture data:

```bash
./bin/kyn check \
  --cwd testdata/angular \
  -c kyn.config.yaml \
  -f libs/ui/button/button.component.ts,libs/ui/button/button.component.html
```

## CLI Usage

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

Exactly one change input mode is required.

### Common commands

```bash
# CI happy path (git refs)
kyn check -c kyn.config.yaml --base origin/main --head HEAD -o json

# Piped changed-file list
git diff --name-only origin/main...HEAD | kyn check -c kyn.config.yaml --stdin

# Explicit files
kyn check -c kyn.config.yaml -f path/a.ts,path/b.ts
```

## Exit Codes

- `0`: policy passed
- `1`: one or more rules failed
- `2`: invalid CLI usage or invalid config
- `3`: runtime/provider error (for example git failure)

## Configuration Model

Kyn config defines:

- `families`: related file groups detected by glob patterns
- `kin`: template-resolved sibling paths
- `rules`: policies bound to a family
  - `when`: gate for whether rule runs
  - `require`: pass/fail checks or `emitFlag`

Example config and full schema details are in [docs/spec.md](docs/spec.md).

## Output

Formats:

- `text`: human-readable summary
- `json`: machine-readable CI parsing

Behavior:

- Results and file lists are deterministic.
- Text output shows failures first.
- Passing rows are hidden by default; show with `--show-passes`.

## CI Integration

Default CI command:

```bash
./bin/kyn check -c kyn.config.yaml --base origin/main --head HEAD -o json
```

Detailed provider examples: [docs/ci.md](docs/ci.md)

## Documentation

- [docs/spec.md](docs/spec.md): full product + CLI spec
- [docs/decisions.md](docs/decisions.md): locked MVP decisions
- [docs/cli-validation-matrix.md](docs/cli-validation-matrix.md): valid/invalid flag combinations
- [docs/ci.md](docs/ci.md): DevOps and CI usage guide
- [docs/mvp-tasks.md](docs/mvp-tasks.md): original MVP backlog
- [docs/README.md](docs/README.md): docs index

## Development

```bash
make fmt
make test
make vet
make build
```

Project layout:

```txt
cmd/kyn         binary entrypoint
internal/       core implementation packages
testdata/       fixture inputs
docs/           project documentation
```

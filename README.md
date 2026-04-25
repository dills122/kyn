# Kyn

Kyn is a stateless Go CLI that helps DevOps and platform teams enforce change policy in CI without custom scripts.

## The Problem Kyn Solves

Most CI pipelines can tell you what changed, but not whether related updates happened together. That creates policy drift:

- Components change but stories/specs/docs are missed.
- Release sidecars exist (for example Figma metadata) but follow-up checks are skipped.
- Teams rely on brittle shell glue that is hard to maintain across repos.

Kyn turns those checks into deterministic, config-driven policy with stable exit codes and machine-readable output.

## Why DevOps Teams Use Kyn

Kyn gives you:

- One consistent policy gate across local runs and CI.
- Explicit failure categories (`1` policy fail, `2` usage/config, `3` runtime/provider).
- Deterministic output that is safe for CI parsing and diffing.
- A reusable rules model instead of one-off per-repo scripts.

At runtime, Kyn answers questions like:

- If a component changed, did its Storybook story change?
- If a component has a Figma sidecar, should a publish flag be emitted?
- If required sibling files exist, were they updated when source changed?

## Why You Need This

If your repo has "file families" (source + tests + docs + configs + generated artifacts), drift is inevitable without enforcement.

Typical failure pattern:

- Engineers update source files.
- Related files are forgotten because CI only validates syntax/tests, not relationship policy.
- Reviewers catch misses inconsistently.
- Release risk and maintenance burden rise over time.

Kyn closes that gap by making relationship rules explicit and enforceable in every PR.

## Problems Kyn Solves

- Prevents partial changes that break team conventions.
- Reduces reviewer overhead for repetitive checklist items.
- Catches policy drift early in CI before release.
- Standardizes rules across many repos/services.
- Replaces ad hoc bash logic with a declarative config model.

## Where Kyn Is Useful (Beyond Web UI)

Web/frontend is a strong fit, but the pattern applies broadly:

- React/Angular/Vue/Svelte:
  component changes should align with stories, tests, styles, snapshots, docs.
- Node/Go/Java/.NET services:
  handler/model changes should align with tests, API specs, migration notes, runbooks.
- API-first teams:
  OpenAPI/GraphQL/Proto changes should align with generated clients, schema docs, compatibility tests.
- Data/analytics:
  SQL/model changes should align with downstream tests, lineage docs, dashboard contract files.
- IaC/platform:
  Terraform/Helm/Kubernetes changes should align with policy files, env overlays, or release manifests.
- Mobile:
  view-model/component changes should align with UI tests, snapshots, localization/resource files.

## Example Policy Families

- `button.component.ts` + `button.stories.ts` + `button.spec.ts`
- `user_handler.go` + `user_handler_test.go` + `openapi.yaml`
- `service.proto` + generated stubs + compatibility tests
- `deployment.yaml` + `values-prod.yaml` + policy manifest
- `model.sql` + data quality tests + docs page

## Features

- `kyn check` command for policy evaluation
- YAML config with schema validation
- Multiple change input modes: `--files`
- Multiple change input modes: `--files-from`
- Multiple change input modes: `--stdin` (alias for `--files-from -`)
- Multiple change input modes: `--base` + `--head` (git diff)
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
- `when`: gate for whether a rule runs
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
make hooks
make fmt
make lint
make test
make vet
make build
```

`make hooks` configures repo-managed git hooks (`.githooks`) for:

- pre-commit: format staged Go files + `go vet ./...`
- pre-push: `go test ./...`

Project layout:

```txt
cmd/kyn         binary entrypoint
internal/       core implementation packages
testdata/       fixture inputs
docs/           project documentation
```

# Configure Families and Rules

Use config version 2 for new policies. Kyn still loads v1 configs, and
`kyn config migrate` can convert their rule shape safely.

## Start from a preset

```bash
kyn init --preset web-ui
```

Available presets are `web-ui`, `api`, `proto`, and `iac`. The command
writes `kyn.config.yaml` and refuses to overwrite an existing file unless you
pass `--force`.

Without `-c`, Kyn searches the working directory in this order:

1. `kyn.config.yaml`
2. `kyn.config.yml`
3. `.kyn.yaml`
4. `.kyn.yml`

Unknown YAML fields and invalid references are config errors, so misspellings
fail early instead of silently weakening policy.

## Complete v2 shape

```yaml
version: 2

families:
  - id: web-component
    groups:
      source:
        include:
          - "src/**/*.component.ts"
          - "src/**/*.component.html"
        exclude:
          - "src/**/generated/**"
      tests:
        include:
          - "src/**/*.spec.ts"
    baseName:
      stripSuffixes:
        - ".component"
    kin:
      story: "{dir}/{base}.stories.ts"
      spec: "{dir}/{base}.spec.ts"
      figma: "{dir}/figma.{base}.json"

rules:
  - id: story-sync
    description: "Keep Storybook coverage aligned with component changes."
    family: web-component
    severity: error
    if:
      changedAny: [source]
      kinExists: [story]
    assert:
      kinChanged: [story]
    message: "Component changed but its existing story did not."

  - id: design-review-signal
    family: web-component
    severity: info
    if:
      changedAny: [source]
      kinExists: [figma]
    actions:
      emit: [designReviewRequired]
    message: "A component with a Figma sidecar changed."
```

## Family fields

| Field | Required | Purpose |
| --- | --- | --- |
| `id` | Yes | Unique family identifier |
| `groups.source.include` | Yes for native v2 configs | Globs that create family instances |
| `groups.source.exclude` | No | Matching source paths to ignore |
| Other `groups` | No | Reserved definitions that current rules cannot reference |
| `baseName.stripSuffixes` | No | Normalizes names before resolving `{base}` |
| `kin` | Yes | Map of related-file names to path templates |

The `source` group is the family entry point. Related files are resolved by
kin templates and checked against the full change set.

For migration compatibility, a version 2 family with no `groups` map may still
use the legacy top-level `include` and `exclude` fields. Kyn treats those fields
as the source group. New and edited v2 configs should use `groups.source`; the
v1-to-v2 migrator emits that native shape.

Configuration validation rejects a family that sets both a `groups.source`
with `include`/`exclude` values and the legacy top-level `include`/`exclude`
fields — pick one, since only `groups.source` is consulted once it's set.

!!! note "Source-anchored evaluation"
    Runtime family resolution is source-anchored. Rules may reference only
    `changedAny: [source]`; configuration validation rejects non-source group
    references until those groups can be evaluated correctly.

### Glob behavior

Use slash-separated repository-relative paths, including on Windows:

```yaml
include:
  - "internal/**/*_handler.go"
exclude:
  - "internal/generated/**"
```

Changed-file inputs and resolved kin paths must remain inside `--cwd`. Kyn
rejects absolute paths, Windows drive paths, NUL bytes, and paths that escape
through `..`.

All source patterns in the family participate in matching. Excluded paths do
not create family instances.

### Base-name normalization

Given `internal/order/order_handler.go`:

```yaml
baseName:
  stripSuffixes: ["_handler", "_service"]
kin:
  test: "{dir}/{base}_test.go"
```

`{base}` becomes `order`, so the resolved kin is
`internal/order/order_test.go`. Use `{name}` when the suffix should remain.

Kin templates may also use `{file}` (the full changed path) and `{ext}` (the
file extension, including the dot). Avoid `{ext}`, `{file}`, or `{name}` when
a family's `source` group matches more than one extension for the same
`{dir}/{base}` (for example `*.component.ts` and `*.component.html` in the
same instance) — Kyn resolves each family instance's kin paths once, from one
of its changed source files, and rejects the config with an error if a
template would resolve to a different path depending on which source file
that happened to be.

## Rule fields

| Field | Required | Purpose |
| --- | --- | --- |
| `id` | Yes | Unique rule identifier |
| `description` | No | Longer intent for maintainers |
| `family` | Yes | Existing family ID |
| `severity` | Yes | `info`, `warn`, or `error` |
| `if` | No | Applicability conditions |
| `assert` | No | Conditions that can pass or fail |
| `actions.emit` | No | Informational flags for machine consumers |
| `message` | Yes | Actionable result text |

### Conditions and assertions

| Clause | Allowed in | Meaning |
| --- | --- | --- |
| `changedAny` | `if` | A configured source group has changed |
| `changedStatusAny` | `if` | Source status matches an allowed configured value |
| `kinExists` | `if`, `assert` | Named kin exists on disk |
| `kinMissing` | `if`, `assert` | Named kin does not exist on disk |
| `kinChanged` | `assert` | Named kin is in the change set |
| `kinUnchanged` | `assert` | Named kin is absent from the change set |

Every kin name must exist in the family's `kin` map. `changedAny` currently
accepts only the required `source` group.

Configuration validation rejects `changedAny` and `changedStatusAny` under
`assert`; keep both change-selection clauses under `if`.

Configuration validation also rejects a rule that sets both `if` and the
legacy v1 `when`, or both `assert` and the legacy v1 `require`. Pick one
clause pair per rule — mixing them is not allowed to silently resolve in
`if`/`assert`'s favor.

!!! note "Use Git input for status rules"
    Git input preserves `added`, `modified`, and rename-destination status.
    Deleted paths are excluded, so configuration validation accepts only
    `added`, `modified`, and `renamed`. Explicit `--files`, file, and stdin lists
    record supplied paths as `modified`.

## Design a useful policy

Start with a review comment your team repeats:

1. Identify the source files that trigger the concern.
2. Define one deterministic kin template.
3. Apply the rule only when that kin exists, unless creating it is the policy.
4. Start at `warn` when rollout noise is uncertain.
5. Use `kyn check --dry-run-resolve` to inspect matching.
6. Use `kyn explain` to inspect each clause.
7. Promote to `error` or run with `--fail-on warn` after the policy proves useful.

The [real-world recipes](recipes/frontend.md) show this process with complete
configs and expected outcomes.

## Migrate a v1 config

```bash
kyn config migrate \
  -c kyn.config.yaml \
  --from v1 \
  --to v2
```

The migrator writes a separate v2 file by default. Review it, preview resolution,
and compare results before replacing the original.

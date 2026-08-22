# How Kyn Works

Kyn evaluates a change set against relationships you define in YAML. It does
not inspect language syntax, call a hosted service, or modify files.

## The evaluation pipeline

```text
changed files
  → family matching
  → kin path resolution
  → rule evaluation
  → deterministic report and exit code
```

That small model is enough to express policies such as “when a handler changes,
its existing test must be part of the same pull request.”

## Change set

Kyn can collect changed files in four ways:

| Mode | Example | Best for |
| --- | --- | --- |
| Automatic Git diff | `kyn check` | Normal local and CI use |
| Explicit Git refs | `--base origin/main --head HEAD` | Reproducible CI |
| Explicit paths | `--files a.go,b.go` | Fast local testing |
| File or stdin | `--files-from changed.txt`, `--stdin` | Existing change detectors |

Only one mode can be active. Automatic mode uses `origin/main...HEAD` inside
a Git repository. `KYN_BASE_REF` and `KYN_HEAD_REF` override those defaults.

Git diff mode also knows whether each path was added, modified, deleted, or
renamed. Rules that use `changedStatusAny` should use Git input; explicit path
lists do not carry status metadata.

## Families

A family describes one logical unit in the repository. This frontend family
makes component source files the entry point:

```yaml
families:
  - id: web-component
    groups:
      source:
        include:
          - "src/**/*.component.ts"
          - "src/**/*.component.html"
    baseName:
      stripSuffixes: [".component"]
    kin:
      story: "{dir}/{base}.stories.ts"
```

For `src/button/button.component.ts`, Kyn derives:

- directory: `src/button`
- name: `button.component`
- base after suffix stripping: `button`
- family instance: `src/button/button`
- `story` kin: `src/button/button.stories.ts`

Paths are slash-normalized on every operating system.

## Kin

Kin are named related paths. Templates can use:

| Variable | Meaning |
| --- | --- |
| `{dir}` | Directory containing the matched source |
| `{file}` | Full matched source path |
| `{name}` | Filename without its final extension |
| `{base}` | Name after configured suffix stripping |
| `{ext}` | Final extension, including the dot |

Kin do not have to exist. A rule can apply only when kin exists, require it to
exist, require it to change, or require it to remain unchanged.

## Rules

A v2 rule has three parts:

```yaml
- id: story-sync
  family: web-component
  severity: error
  if:
    changedAny: [source]
    kinExists: [story]
  assert:
    kinChanged: [story]
  message: "Component changed but story did not."
```

1. `if` decides whether the rule applies.
2. `assert` decides whether the applicable rule passes.
3. `severity` and `--fail-on` decide whether a failed result blocks.

Populated clauses within `if` are combined with AND semantics. Assertions are
also combined with AND semantics.

Rules can emit flags instead of asserting:

```yaml
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

Flags are sorted and deduplicated in machine output, which makes them safe input
for later CI steps.

## Severity versus process failure

`info`, `warn`, and `error` describe rule results. The command's threshold
is separate:

| Result | Default `--fail-on error` | `--fail-on warn` |
| --- | --- | --- |
| Failed info rule | Does not block | Does not block |
| Failed warning rule | Does not block | Blocks |
| Failed error rule | Blocks | Blocks |

This means a report can contain `Status: fail` for a warning while the command
still exits 0. Use `--fail-on warn` when warnings should be policy gates.

## Determinism

Kyn sorts changed files, family instances, rule results, expected paths, and
flags. Given the same repository state, config, and change set, local and CI
runs produce the same ordering and exit behavior.

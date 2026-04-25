# Kyn MVP Decisions

This file captures product and implementation decisions locked for MVP.

## CLI Input Modes

Exactly one change input mode is allowed per run:

1. `--files <csv>`
2. `--files-from <path>` (supports `-` to read from stdin)
3. `--stdin` (alias for `--files-from -`)
4. `--base <ref> --head <ref>`

Invalid combinations are CLI usage errors (`exit 2`).

If git mode is selected, both `--base` and `--head` are required.

## Rule Semantics

- `when` predicates are `AND` across keys.
- `require` predicates are `AND` across keys.
- For list-based predicates, all listed kin names must satisfy the predicate.
- `changedAny` is `ANY` across listed change groups.

## Flags in Rules

`require.emitFlag` is informational and adds a flag string to summary output.

`emitFlag` itself does not introduce a failure condition.

## Family Instance Deduplication

Family instances are deduplicated by default for MVP.

## No-Match Behavior

If no family instances match, command succeeds with empty results by default.

`--fail-on-empty` changes this behavior to fail the run.

## Config Validation

If a rule references a non-existent `family` id, config validation fails (`exit 2`).

## Git Change Detection (MVP)

Use:

```bash
git -C <cwd> diff --name-status -M <base>...<head>
```

Flattening policy:

- Include `A` and `M` paths in changed-set evaluation.
- Include rename destination (`R*` new path) in changed-set evaluation.
- Exclude `D` paths from changed-set evaluation.

Deleted paths may be retained as metadata for future enhancements, but are not evaluated as changed files in MVP rules.

## Deterministic Output

Output should be deterministic:

- Stable ordering of results.
- Stable ordering of `changedFiles`, `expectedFiles`, and `flags`.
- Text output groups failures first.
- Passing results are hidden by default in text output and can be shown with `--show-passes`.

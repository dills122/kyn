# v3 grammar

Status: proposed. This is the current shape; the history of how it got here is
in [`00-workplan.md`](00-workplan.md). Nothing is implemented.

The goal is a config a developer can read aloud: *when files like this change,
that file is expected to be touched too.*

## The common rule

```yaml
version: 3

rules:
  handler-test:
    match: "internal/**/{name}_handler.go"
    related: "{dir}/{name}_handler_test.go"
    expect: touched
    skipIf: missing
```

> When a handler changes, its test is expected to be touched in the same
> change. Handlers that have no test yet are skipped.

Drop `skipIf` and the rule is strict: every handler must have a test, and it
must be touched.

## Fields

| Field | Required | Default | Meaning |
| --- | --- | --- | --- |
| `match` | yes | | Glob, or list of globs, selecting source files. May bind variables |
| `exclude` | no | | Glob, or list, removed from `match` |
| `related` | yes, unless `expect: none` | | Path template for the related file |
| `expect` | yes | | What must be true of the related file |
| `skipIf` | no | never skip | Skip the rule based on whether the related file is there |
| `changeType` | no | all | Only apply when a matched source was `added`, `modified` or `renamed` |
| `severity` | no | `error` | `info`, `warn` or `error` |
| `message` | no | generated | Shown when the rule fails |
| `why` | no | | Why the rule exists. Carried to SARIF `fullDescription` and `explain` |
| `flags` | no | | Names added to the report's `flags` list whenever the rule applies |

Rule IDs are the mapping keys under `rules`. Every list-valued field also
accepts a single string.

### `match` binds variables

A `{name}` in a glob captures what it matches, and `related` reuses it.

| Syntax | Matches | Example |
| --- | --- | --- |
| `*`, `**`, `{a,b}` | standard glob wildcards and alternation, not captured | `src/**/*.{ts,html}` |
| `{name}` | one or more characters within a single path segment | `{name}.component.ts` |
| `{**name}` | one or more whole path segments | `src/main/java/{**pkg}/{name}.java` |
| `{dir}` | built in: the directory of the matched file | used in `related` only |

A brace group containing a comma is alternation; one without is a capture.

```yaml
story-sync:
  match: "src/**/{name}.component.{ts,html}"
  related: "{dir}/{name}.stories.ts"
```

`button.component.ts` and `button.component.html` both bind `name=button`, so
they are one instance demanding one story.

`{**name}` is what makes mirrored trees expressible, which v2 could not do:

```yaml
java-test:
  match: "src/main/java/{**pkg}/{name}.java"
  related: "src/test/java/{pkg}/{name}Test.java"
```

Captures replace v2's `{base}`, `{name}`, `{ext}`, `{file}` and
`baseName.stripSuffixes`. The variable is named where it is matched, so there
is no second place for it to disagree — the defect behind OS11.

### `expect`

| Value | True when the related path |
| --- | --- |
| `touched` | is part of this change (added, modified, or a rename destination) |
| `untouched` | is not part of this change |
| `exists` | is present in the working tree |
| `missing` | is not present in the working tree |
| `none` | no assertion; the rule is informational |

A list is AND-combined. Kyn observes paths, never content: `touched` means the
path is in the change, not that its content is correct.

Plain `expect: touched` is the form to reach for. `[exists, touched]` differs
from it only when the diff and the working tree disagree (the file is in the
diff but not on disk), so it is rarely needed.

`untouched` passes for a file this change deleted, because a deleted path is
not part of the change set. To require "present and unmodified", write
`expect: [exists, untouched]`.

### `skipIf`

| Value | The rule is skipped when the related file |
| --- | --- |
| `missing` | was not there before this change and is not there now |
| `exists` | is there, or was removed by this change |

A file **deleted or renamed away by the change is never "missing"**. That is
what stops a rule being silenced by deleting the file it protects (OS3, OS8).
The three underlying states are defined in
[`02-semantics.md`](02-semantics.md).

`skipIf: exists` is for the inverse rule — act only when the related file is
absent:

```yaml
needs-story:
  match: "src/**/{name}.component.ts"
  related: "{dir}/{name}.stories.ts"
  expect: none
  skipIf: exists
  flags: [needsStory]
```

Skipped results are reported; see [`03-instances-and-reports.md`](03-instances-and-reports.md).

### Informational rules

`expect: none` makes a rule informational: it never fails, reports status
`info`, and adds its `flags`.

```yaml
payments-review:
  match: "internal/payments/**"
  expect: none
  flags: [paymentsReviewRequired]
  message: "Payments code changed."
```

`related` is optional here, because "this code was touched" needs no second
file. `expect` stays required so that a rule cannot become informational by
losing a line.

## Validation

All of these fail with exit 2 and name the rule and field.

| # | Rule |
| --- | --- |
| 1 | Unknown fields are rejected. Duplicate rule IDs are rejected by the YAML decoder, with line numbers |
| 2 | `match` has at least one glob; every glob in one rule binds the same capture names |
| 3 | A glob has at most one `{**name}`; capture names are unique within a glob; `dir` is reserved |
| 4 | `related` uses only `{dir}` and names that `match` captures |
| 5 | `related` uses at most one variable that can contain `/` (`{dir}` or the `{**name}` capture) |
| 6 | `expect` is required; `none` cannot be combined with other values; lists are deduplicated and sorted |
| 7 | `related` is required unless `expect: none`; `skipIf` requires `related` |
| 8 | `expect: none` requires `severity: info` (its default) and at least one of `message` or `flags` |
| 9 | A resolved `related` path that matches its own rule's `match` and is not excluded is an error. Detected at resolve time, when a concrete path exists |
| 10 | Resolved paths must stay inside `--cwd` |
| 11 | Rule IDs are sorted before every validation and resolution pass, so the first error reported is stable |

Rule 5 keeps instance names unique; the reasoning is in
[`03-instances-and-reports.md`](03-instances-and-reports.md). Rule 9's error
tells the author what to do:

```text
rule "go-test": related path "src/a_test.go" also matches this rule's own
match; add an exclude or narrow the match
```

Rule 11 is not style. OS7 measured the current code reporting three different
errors across thirty identical runs.

## Sharing globs

There is no `patterns` block and no rule-to-rule reference. Two rules over the
same sources repeat the `match` line, which with captures is one line.

If a long glob list must be shared, YAML anchors already work through Kyn's
strict loader (OS15):

```yaml
rules:
  story-sync:
    match: &components ["src/**/{name}.component.ts", "lib/**/{name}.component.ts"]
    related: "{dir}/{name}.stories.ts"
    expect: touched
  spec-sync:
    match: *components
    related: "{dir}/{name}.spec.ts"
    expect: touched
```

## What v3 cannot express

Two v2 capabilities are not carried over. No configuration in this repository
uses either (four presets, three recipes, `docs/site/config.md`,
`docs/site/concepts.md`, `testdata/`, `e2e/projects/`).

| Limit | v2 form | Why it is out |
| --- | --- | --- |
| Atomic multi-path rules | `if.kinExists: [story, spec]` with `assert.kinChanged: [story, spec]`: skipped as one unit if either is missing | One `related` per rule. Two rules evaluate independently |
| Gating on one path, asserting on another | `if.kinExists: [figma]` with `assert.kinChanged: [story]` | `skipIf` reads the rule's own `related` path |

These are the limits found so far, by three reviews. The list has been wrong
twice, so it is not claimed to be complete. `expect` is list-shaped and
`related` could become list-shaped later, so the first limit can be lifted
additively.

One capability is gained: mirrored source and test trees, via `{**name}`.

## Naming

`version: 3` is the config schema version and is unrelated to the product
version. Prose calls this the rule-centric format. The number stays `3` even
though v1 and v2 are removed, because restarting at `1` would collide with the
original v1 in history and documentation.

## JSON Schema

A schema ships once this grammar freezes, generated from or CI-checked against
the loader. It covers shape and enums; cross-field rules (2 to 8) and the
resolve-time rule (9) stay with the loader, which remains the authority.

## Worked scenarios

[`experiments/sim/scenarios.yaml`](experiments/sim/scenarios.yaml) holds 21
realistic configurations with change sets, run by a small simulator of these
semantics:

```bash
go run docs/design/v3/experiments/sim/main.go docs/design/v3/experiments/sim/scenarios.yaml
```

The simulator is a model of the design, not Kyn code. It exists so the shape
can be tried before anything is built, and it is what surfaced optional
`related` for informational rules and the `{**name}` capture.

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
| `exclude` | no | | Glob, or list, removed from `match`. No variables |
| `related` | yes, unless `expect: none` | | Path template for the related file |
| `expect` | yes | | What must be true of the related file |
| `skipIf` | no | never skip | Skip the rule based on the state of the related file |
| `changeType` | no | all | Apply only when a matched **source** file was `added`, `modified` or `renamed`. It filters the source, not the related file |
| `severity` | no | `error`; `info` when `expect: none` | `info`, `warn` or `error` |
| `message` | no | generated | Shown when the rule fails |
| `why` | no | | Why the rule exists. Carried to SARIF `fullDescription` and `explain` |
| `flags` | no | | Names added to the report's `flags` list whenever the rule applies |

Rule IDs are the mapping keys under `rules`. Every list-valued field also
accepts a single string.

## Matching

Paths are slash-normalized and relative to `--cwd`. Matching is case-sensitive.
Every rule below is checked when the config loads, except the two marked
resolve-time.

### Globs

| Syntax | Matches |
| --- | --- |
| `*` | any characters within one path segment, including none |
| `?` | exactly one character within a segment |
| `**/` | zero or more whole directories |
| a trailing `**` | everything below the directory |
| `{a,b}` | alternation of literal alternatives, not captured |

There is no way to match a literal `{` or `}` in v3.0.

### Captures

A brace group **without a comma** is a capture.

| Syntax | Matches | Notes |
| --- | --- | --- |
| `{name}` | one or more characters within a single segment, including a leading dot | Greedy: it takes as much as it can while the rest of the glob still matches |
| `{**name}` | zero or more whole path segments | Must start a segment and fill it. Captured without the surrounding slashes; zero segments captures the empty string. A **trailing** `{**name}` matches one or more segments, since a path must end in a file |
| `{dir}` | built in, used in `related` only: the directory of the matched source | Empty at the repository root. Cannot be captured |

A capture name is letters, digits and underscore, not starting with a digit.

```yaml
story-sync:
  match: "src/**/{name}.component.{ts,html}"
  related: "{dir}/{name}.stories.ts"
```

`button.component.ts` and `button.component.html` both bind `name=button`, so
they are one instance demanding one story.

Greedy matching only matters when a literal repeats. In
`src/{name}.{kind}.ts`, the path `user.profile.component.ts` binds
`name=user.profile` and `kind=component`: the first capture takes as much as it
can, so it does not split at the first dot. A glob whose captures are separated
by literals that cannot repeat avoids the question.

`{**name}` is what makes mirrored trees expressible, which v2 could not do. It
matches zero segments, so top-level files are covered:

```yaml
py-test:
  match: "src/{**pkg}/{name}.py"
  related: "tests/{pkg}/test_{name}.py"
```

`src/acme/billing/invoice.py` binds `pkg=acme/billing`; `src/app.py` binds
`pkg=` and resolves to `tests/test_app.py`, because empty segments are removed
when the path is cleaned.

### Several globs in one rule

Every glob in a rule binds the **same set of capture names**. A source path that
matches more than one glob must bind them to the same values. If it does not,
the config is rejected at resolve time, naming the path. Order never decides,
so reordering the list cannot change a policy.

```yaml
match: ["internal/**/{name}_{handler,service}.go", "internal/**/{name}.go"]
```

This is rejected the first time `order_handler.go` changes: the first glob binds
`name=order`, the second binds `name=order_handler`. Write globs that do not
overlap.

### Exclude

`exclude` globs contain no captures. They remove paths from `match`, and they
also count when a resolved related path is checked against its own rule (see
rule 10 below).

### `related`

A template for exactly one path:

- relative, with no `..` segment and no glob characters (`* ? [ ]`);
- variables are `{dir}` and the names `match` captures, nothing else;
- at most one variable that can contain `/`: `{dir}` or the `{**name}` capture;
- every capture must be used.

The result is cleaned, so an empty `{dir}` or `{**name}` leaves no stray slash.

## `expect`

| Value | True when the related path |
| --- | --- |
| `touched` | is part of this change (added, modified, or a rename destination) |
| `untouched` | is unaffected by this change: not touched, and not removed by it |
| `exists` | is present in the working tree |
| `missing` | is not present in the working tree |
| `none` | no assertion; the rule is informational |

A list is AND-combined. Kyn observes paths, never content: `touched` means the
path is in the change, not that its content is correct.

Plain `expect: touched` is the form to reach for. `[exists, touched]` differs
from it only when the diff and the working tree disagree (the file is in the
diff but not on disk), so it is rarely needed.

## `skipIf`

The related file is in one of three states: `exists`, `missing`, or `removed`
(it was deleted or renamed away by this change; see
[`02-semantics.md`](02-semantics.md)).

| `skipIf` | The rule is skipped when the related file is |
| --- | --- |
| `missing` | `missing`: it was not there before and is not there now |
| `exists` | `exists` |

**A removal never causes a skip.** A `removed` file is neither "missing" nor
"exists" for `skipIf`, so the rule applies under both values. That is what
stops a rule being silenced by deleting the file it protects.

`skipIf: exists` is for the inverse rule: act only when the related file is not
there.

```yaml
needs-story:
  match: "src/**/{name}.component.ts"
  related: "{dir}/{name}.stories.ts"
  expect: none
  skipIf: exists
  flags: [needsStory]
```

Skipped results are reported; see
[`03-instances-and-reports.md`](03-instances-and-reports.md).

The protection is same-change only. If a test was deleted in an earlier change,
it is `missing` now and `skipIf: missing` skips. A team that wants every handler
to have a test always should omit `skipIf`.

## Informational rules

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
file. `expect` stays required so a rule cannot become informational by losing a
line. `none` reads as "no assertion", not "expect no such file".

## Validation

All of these fail with exit 2 and name the rule and the field.

### At load

| # | Rule |
| --- | --- |
| 1 | Unknown fields are rejected, and so are duplicate rule IDs, by the YAML decoder with line numbers |
| 2 | `match` has at least one glob. Every glob binds the same capture names. Capture names are valid, unique within a glob, and not `dir`. No adjacent captures. At most one `{**name}` per glob, filling whole segments. No captures in `exclude` |
| 3 | `related` is relative, has no `..` and no glob characters, uses only `{dir}` and captured names, and uses at most one variable that can contain `/` |
| 4 | Every capture is used in `related`. This also catches `{ts}` written where `{ts,html}` was meant |
| 5 | `expect` is required with valid values. `none` stands alone. `touched` with `untouched`, and `exists` with `missing`, contradict and are rejected |
| 6 | `related` is required unless `expect: none`. `skipIf` requires `related`. `skipIf` and `changeType` take only their listed values |
| 7 | `skipIf: exists` with an `expect` that includes `exists` can never pass, since the rule only applies when the file is not there, and is rejected |
| 8 | `expect: none` requires `severity: info` and at least one of `message` or `flags` |

### At resolve time

These need a concrete path, so they fire only when a matching file changes.

| # | Rule |
| --- | --- |
| 9 | A path matching two `match` globs with different bindings is an error |
| 10 | A resolved `related` path that matches its own rule's `match` and is not excluded is an error |

```text
rule "go-test": related path "src/a_test.go" also matches this rule's own
match; add an exclude or narrow the match
```

Because 9 and 10 depend on which files change, a config can load cleanly and
fail on someone's unrelated change. To exercise every rule against the whole
tree before that happens, resolve every tracked file:

```bash
git ls-files | kyn check --files-from - --dry-run-resolve
```

### Always

Rule IDs are processed in sorted order and lists are deduplicated and sorted, so
the first error reported is stable. These are not errors. OS7 measured the
current code reporting three different errors across thirty identical runs.

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

Limits found so far, by four reviews and the author's scenarios. The list has
been wrong before, so it is not claimed to be complete. None is used by a
configuration in this repository (four presets, three recipes,
`docs/site/config.md`, `docs/site/concepts.md`, `testdata/`, `e2e/projects/`).

| Limit | Example | Why |
| --- | --- | --- |
| Atomic multi-path rules | v2 `if.kinExists: [story, spec]` with `assert.kinChanged: [story, spec]`, skipped as one unit | One `related` per rule; two rules evaluate independently |
| Gating on one path, asserting on another | v2 `if.kinExists: [figma]` with `assert.kinChanged: [story]` | `skipIf` reads the rule's own `related` |
| A related path that is a glob | "a schema change needs *some* new migration"; "a source change needs a `.changeset/*.md`" | `related` names exactly one path |
| A capture limited to alternatives | v2 `{ext}` over `*.test.{ts,tsx}`; v3 writes one rule per extension | Alternation inside braces is not captured |
| Guaranteeing one failure per related file | `related: "docs/{a}-{b}.md"` over `src/{a}_{b}.go`: `x-y_z.go` and `x_y-z.go` are different instances demanding the same `docs/x-y-z.md` | Instances are keyed on variable values, and distinct values can resolve to one path when literal text between variables can also occur inside them. Rare, and reported as two failures, not hidden. Scenario 39 |
| Forbidding edits with no related path | "files under `db/migrations/` must never be modified" | `expect: untouched` needs a related path, and a rule whose related file is its own match is rejected by rule 10 |

`expect` is list-shaped and `related` could become list-shaped later, so the
first and third limits could be lifted additively.

What v3 gains: mirrored source and test trees (`{**name}`); per-package and
per-directory paths (`packages/{pkg}/CHANGELOG.md`); stripping a prefix or
suffix with literal text (`test_{name}.py`); and one failure, not one per
source file, for a constant related path.

## Naming

`version: 3` is the config schema version and is unrelated to the product
version. Prose calls this the rule-centric format. The number stays `3` even
though v1 and v2 are removed, because restarting at `1` would collide with the
original v1 in history and documentation.

## JSON Schema

A schema ships once this grammar freezes, generated from or CI-checked against
the loader. It covers shape and enums; the cross-field rules (2 to 8) and the
resolve-time rules (9 and 10) stay with the loader, which remains the
authority.

## Worked scenarios

[`experiments/sim/scenarios.yaml`](experiments/sim/scenarios.yaml) holds 51
scenarios: realistic policies with change sets, and one for each validation
rule. Each states its expected outcome, and the simulator fails on a mismatch:

```bash
go run docs/design/v3/experiments/sim/main.go docs/design/v3/experiments/sim/scenarios.yaml
```

The simulator is a model of this design, not Kyn code, and was written by the
same author as the scenarios. A green run shows the model is internally
consistent; it is not independent evidence the design is right. It was
mutation-tested: reverting each of five decisions in the model turns the
scenarios that guard them red.

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
Every rule below is checked when the config loads, except the three marked
resolve-time.

### Globs

v3 has its own small glob dialect. It is not the matcher v2 uses (doublestar),
and it does not inherit that matcher's extra syntax: anything outside the table
is rejected at load, by name, instead of silently never matching.

| Syntax | Matches |
| --- | --- |
| `*` | zero or more characters within one path segment |
| `?` | exactly one character within a segment |
| `**` as a whole segment, not last | zero or more directories: `src/**/x.go` matches `src/x.go` and `src/a/b/x.go` |
| `**` as the last segment | one or more segments below the directory: `src/**` matches `src/a`, not `src` |
| `{a,b}` | alternation of literal text, not captured |

A glob is relative and made of non-empty segments. These are rejected:

| Rejected | Why it matters |
| --- | --- |
| `[ab]`, `[!x]` | v2 treated them as character classes. v3 has none, and silently reading `[` as a literal would change what a ported glob matches |
| `\` escapes | the same |
| nested braces, or `*`, `?`, `/` or an empty alternative inside `{a,b}` | alternatives are literal text so that alternation cannot decide a binding |
| `**` that is not a whole segment (`a**b`) | doublestar reads it as `*` and does not cross a `/`; the simulator once read it as "anything". Requiring a whole segment leaves one meaning |
| a leading `/`, a trailing `/`, an empty segment, a `.` or `..` segment | a glob names repository-relative paths |
| an unmatched `{` or `}` | |

There is no way to match a literal `{`, `}`, `[`, `]` or `\` in v3.0. No
configuration in this repository uses any rejected form (the four presets, the
recipes, `docs/site/`, `testdata/`, `e2e/projects/`), and v2's matcher also let
`src/**` match a file named `src`, which v3 does not.

### Captures

A brace group **without a comma** is a capture.

| Syntax | Matches | Notes |
| --- | --- | --- |
| `{name}` | one or more characters within a single segment, including a leading dot | See binding below |
| `{**name}` | zero or more whole path segments | Must be the whole segment. Captured without the surrounding slashes; zero segments captures the empty string. A **trailing** `{**name}` matches one or more segments, since a path must end in a file |
| `{dir}` | built in, used in `related` only: the directory of the matched source | Empty at the repository root. Cannot be captured |

A capture name is letters, digits and underscore, not starting with a digit.

```yaml
story-sync:
  match: "src/**/{name}.component.{ts,html}"
  related: "{dir}/{name}.stories.ts"
```

`button.component.ts` and `button.component.html` both bind `name=button`, so
they are one instance demanding one story.

**Binding.** A glob can sometimes match one path in several ways. The binding is
the way whose captures are **longest**, comparing the first capture first, then
the second, and so on. Nothing else has a preference: `*`, `**` and the choice of
alternative do not influence which way is chosen, and the order of alternatives
and of globs never matters. In `src/{name}.{kind}.ts`, the path
`user.profile.component.ts` binds `name=user.profile` and `kind=component`,
because a longer `name` wins.

To keep the rule easy to predict, a `{name}` capture must be bordered by literal
text, a `/`, or an end of the glob. It may not touch another capture, `*`, `?` or
an alternation. `src/*{n}.go` is rejected; in a regular-expression matcher it
would bind `n=c` on `src/abc.go`, which no one means.

The longest-first rule is the **definition**, and the simulator implements it by
enumerating every parse. An implementation can use any algorithm that returns
the same bindings; the layer 0 conformance cases check that
([`07-conformance.md`](07-conformance.md)).

`{**name}` is what makes mirrored trees expressible, which v2 could not do. It
matches zero segments, so top-level files are covered:

```yaml
py-test:
  match: "src/{**pkg}/{name}.py"
  related: "tests/{pkg}/test_{name}.py"
```

`src/acme/billing/invoice.py` binds `pkg=acme/billing`; `src/app.py` binds
`pkg=` and resolves to `tests/test_app.py`, because empty segments are removed
when the path is built.

### Several globs in one rule

Every glob in a rule binds the **same set of capture names, each of the same
kind**: a name is `{x}` in every glob or `{**x}` in every glob, never one of each.
A source path that matches more than one glob must bind them to the same values. If it does not,
the config is rejected at resolve time, naming the path. Order never decides,
so reordering the list cannot change a policy.

```yaml
match: ["internal/**/{name}_{handler,service}.go", "internal/**/{name}.go"]
```

This is rejected the first time `order_handler.go` changes: the first glob binds
`name=order`, the second binds `name=order_handler`. Write globs that do not
overlap.

### Exclude

`exclude` globs contain no captures and use the same dialect. They remove paths
from `match`, and they also count when a resolved related path is checked against
its own rule (see rule 10 below). An `exclude` identical to a `match` glob is
rejected, since it would remove every match.

### `related`

A template for exactly one path:

- non-empty, relative, not ending in `/`, with no `.` or `..` segment, no `\` and
  no glob characters (`* ? [ ]`);
- variables are `{dir}` and the names `match` captures, nothing else, and two
  variables may not touch (`{a}{b}`);
- at most one distinct variable that can contain `/`: `{dir}` or the `{**name}`
  capture;
- every capture must be used.

Resolution substitutes every variable **in one pass**, so a captured value is
never scanned for variables, then drops empty segments, so an empty `{dir}` or
`{**name}` leaves no stray slash and the result is never absolute. The resolved
path must be non-empty and have no `.` or `..` segment. A captured value can be
`..` (a file named `..x.go` matched by `{n}x.go`), and without this check
`{n}/{n}/x.txt` would resolve outside `--cwd`.

A `related` with no variables is a fixed path, so it is checked against the
rule's own `match` when the config loads.

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
| 2 | `match` has at least one glob, and every `match` and `exclude` glob is in the dialect above: nothing from the rejected table, `**` only as a whole segment, alternatives literal and non-empty. Captures: valid names, unique within a glob, not `dir`; `{**name}` fills a whole segment, at most one per glob; a `{name}` capture does not touch another capture, `*`, `?` or an alternation; no captures in `exclude`; an `exclude` is not identical to a `match` glob |
| 3 | Every glob binds the same capture names, each of the same kind (`{x}` or `{**x}`). `related` is non-empty and relative, has no `.`/`..` segment, `\`, trailing `/` or glob character, uses only `{dir}` and captured names, has no two adjacent variables, and uses at most one distinct variable that can contain `/`. A `related` with no variables must not match its own rule |
| 4 | Every capture is used in `related`. This also catches `{ts}` written where `{ts,html}` was meant |
| 5 | `expect` is required with valid values. `none` stands alone. `touched` with `untouched`, and `exists` with `missing`, contradict and are rejected |
| 6 | `related` is required unless `expect: none`. `skipIf` requires `related`. `skipIf` and `changeType` take only their listed values |
| 7 | `skipIf: exists` with an `expect` that includes `exists` can never pass, and with one that includes `touched` can pass only when the diff and the working tree disagree. Both are rejected: the rule applies only when the file is not there |
| 8 | `expect: none` requires `severity: info` and at least one of `message` or `flags` |

### At resolve time

These need a concrete path, so they fire only when a matching file changes. They
run for **every instance, before `changeType` or `skipIf` can skip it**, so a
skipped instance cannot hide a broken rule.

| # | Rule |
| --- | --- |
| 9 | A path matching two `match` globs with different bindings is an error |
| 10 | A resolved `related` path that matches its own rule's `match` and is not excluded is an error |
| 11 | A resolved `related` path that is empty, or has a `.` or `..` segment, is an error |

```text
rule "go-test": related path "src/a_test.go" also matches this rule's own
match; add an exclude or narrow the match
```

Because 9 to 11 depend on which files change, a config can load cleanly and fail
on someone's unrelated change. To exercise them against every existing tracked
path before that happens, resolve all of them:

```bash
git -c core.quotePath=false ls-files | kyn check --files-from - --dry-run-resolve
```

This samples the paths that exist today. It cannot know a source file someone
will add tomorrow, so it narrows the risk rather than removing it. The
`-c core.quotePath=false` is needed because git otherwise quotes non-ASCII paths,
which `--files-from` does not accept.

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

Limits found so far, by five reviews and the author's scenarios. The list has
been wrong before, so it is not claimed to be complete. None is used by a
configuration in this repository (four presets, three recipes,
`docs/site/config.md`, `docs/site/concepts.md`, `testdata/`, `e2e/projects/`).
Glob constructs v3 rejects are listed under Globs, not here.

| Limit | Example | Why |
| --- | --- | --- |
| Atomic multi-path rules | v2 `if.kinExists: [story, spec]` with `assert.kinChanged: [story, spec]`, skipped as one unit | One `related` per rule; two rules evaluate independently |
| Gating on one path, asserting on another | v2 `if.kinExists: [figma]` with `assert.kinChanged: [story]` | `skipIf` reads the rule's own `related` |
| A related path that is a glob | "a schema change needs *some* new migration"; "a source change needs a `.changeset/*.md`" | `related` names exactly one path |
| A capture limited to alternatives | v2 `{ext}` over `*.test.{ts,tsx}`; v3 writes one rule per extension | Alternation inside braces is not captured |
| Guaranteeing one failure per related file | `related: "docs/{a}-{b}.md"` over `src/{a}_{b}.go`: `x-y_z.go` and `x_y-z.go` are different instances demanding the same `docs/x-y-z.md` | Instances are keyed on variable values, and distinct values can resolve to one path when literal text between variables can also occur inside them. Rare, and reported as two failures, not hidden. Scenario 39 |
| Forbidding edits with no related path | "files under `db/migrations/` must never be modified" | `expect: untouched` needs a related path, and a rule whose related file is its own match is rejected by rule 10 |

Things Kyn does not detect, and says so rather than guess:

- An `exclude` that happens to cover everything `match` selects. The rule simply
  has no instances; `--fail-on-empty` reports a run where no rule had any.
- A `related` that resolves to a directory. It is `exists` and can never be
  `touched`, since only files appear in a change.
- A case-only rename on a case-insensitive filesystem. Matching and the change
  set are case-sensitive, so the old and new names look like different files
  while the working tree sees one.

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
resolve-time rules (9 to 11) stay with the loader, which remains the
authority.

## Worked scenarios

[`experiments/sim/scenarios.yaml`](experiments/sim/scenarios.yaml) holds 87
scenarios: realistic policies with change sets, one for each validation rule,
and one for each finding of the fifth review. Each states its expected outcome, and the simulator fails on a mismatch:

```bash
go run docs/design/v3/experiments/sim/main.go docs/design/v3/experiments/sim/scenarios.yaml
```

The simulator is a model of this design, not Kyn code, and was written by the
same author as the scenarios. A green run shows the model is internally
consistent; it is not independent evidence the design is right. It was
mutation-tested: reverting a decision in the model turns the scenarios that
guard it red (five decisions at first, nine more after review 5).

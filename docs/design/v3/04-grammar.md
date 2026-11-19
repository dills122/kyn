# v3 grammar

Status: proposed. This is the current shape; the history of how it got here is
in [`notes.md`](notes.md). Nothing is implemented.

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

## The document

A config is a YAML mapping with exactly two top-level keys.

| Key | Rule |
| --- | --- |
| `version` | required; must be the integer `3`. Any other value is an error naming the version this binary reads (`this kyn reads config version 3; the file says 2`) and, for a lower one, saying to run a kyn release that reads it |
| `rules` | required; a non-empty mapping. Absent, `null`, `{}`, a list or a scalar is an error, so a file holding only `version: 3` is rejected instead of passing with nothing checked |

Any other top-level key is an error. A rule ID, the key under `rules`, matches
`[A-Za-z0-9][A-Za-z0-9._-]*`: it is printed in reports, sorted byte-wise, and
used in SARIF, so it carries no spaces or control characters. Duplicate IDs are
rejected by the decoder (rule 1).

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

Paths are slash-normalized and relative to `--cwd`. Matching is case-sensitive
and works on Unicode code points: `?` is one code point, "longest" counts code
points, and nothing is normalized, so a name in NFD form does not match an NFC
glob. `*` and `?` match a leading dot.

A **blank character** is a Unicode `White_Space` character or a control character
(category Cc). The byte order mark is rejected anywhere, separately. **One predicate** is used for globs, `related` templates, input
paths and resolved related paths, and everything else says "clean path" and
points here. A path is **clean** when it is non-empty valid UTF-8, relative, has
no empty, `.` or `..` segment, no `\`, no control character or byte order mark
anywhere, no segment that begins or ends with a blank character, and no first
segment that begins with an ASCII letter and a colon. Interior spaces are
fine, so `my file.go` is a legal name, and so are names with format characters
(zero-width joiners in an emoji sequence, a tag-sequence flag, a trailing
right-to-left mark): format characters (Unicode Cf) are **not** blank. Globs and templates additionally allow no blank character
inside `{}`, and a glob's first segment may not begin with an ASCII letter and a
colon, since no clean path can match it.

Only `White_Space` and Cc are detected, so a character that merely looks blank is
accepted: a zero-width space or joiner, U+2800 (Braille blank), U+3164, U+115F,
U+034F, a lone U+FE0F. A glob with one pasted at a segment edge or inside `{}`
silently never matches. That is a known limit, chosen over rejecting real file
names that end in a format character (a tag-sequence flag, a right-to-left
mark). `White_Space` is taken from the Unicode version of the Go release that
builds Kyn, which a conformance suite must pin. Paths read from `--files-from` and `--files`
additionally may not be in git's quoted form (a leading `"`); git mode reads paths
with quoting disabled, so there a leading `"` is part of a real name.

The predicate runs on a changed path that a rule's `match` selects as a source,
on every resolved related path (rule 11), and, for `--files` and `--files-from`,
on any path in git's quoted form. A changed path that no rule matches is never
evaluated and is not checked. A path that is only **removed** (a `D` path or a
rename source) is a state key and is never matched, so it is exempt, and the
change that renames a bad name away passes.

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
| `**` that is not a whole segment (`a**b`) | doublestar reads it as `*` and does not cross a `/`. Requiring a whole segment leaves one meaning |
| a leading `/`, a trailing `/`, an empty segment, a `.` or `..` segment, or an alternative that can make one (`{..,a}`, `{.,a}`) | a glob names repository-relative paths |
| a blank character at the edge of a segment or inside `{}` (`{ts, html}` is the likely typo: the alternative would be `" html"`), or a control character or BOM anywhere | it can never match what was meant |
| a first segment that begins with an ASCII letter and a colon (`c:{n}.go`, `c:/{n}`, `c:*`) | a drive prefix; no clean path can match it |
| a `/` inside `{}`, or `*` in a `{**x,a}` | alternatives and capture names cannot hold a separator or a wildcard |
| an unmatched `{` or `}`, or invalid UTF-8 | |

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

A capture name is ASCII letters, digits and underscore, not starting with a digit.

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

That is deterministic only if no two ways tie on every length while binding
different values, and a tie needs a capture that can slide: `src/**/{n}/**` on
`src/auth/user/a.ts` could bind `n=auth` or `n=user`, both four characters.
Sliding also makes the binding depend on directory-name length, so renaming an
unrelated directory could move a file to another instance. A capture, `{name}` or
`{**name}`, therefore **may not have `**` or `{**name}` on both sides of it** in
one glob; the glob is rejected, naming the capture. `src/**/{n}.go` and
`src/{**pkg}/{name}.py` are fine. The rule is deliberately a little wider than the
ties it prevents: `src/{**pkg}/{name}/**` and `src/**/{name}/{**rest}` do not tie
(the longer capture wins) but would still bind on name length, and are rejected
for that stability reason. With the rule the comparison is total. A throwaway
simulator checked that by generating random globs, keeping the accepted ones, and
looking for a path with two longest parses that bind different values: two million
globs tried, about one million accepted, none found, and thousands found with the
rule removed. It is in the history of PR #50, not in the tree.

A named capture is greedy and `*` is not a capture, so they behave differently
where they meet: `src/*-{n}.go` binds `n=b-c` on `a-b-c.go`, while
`src/{m}-{n}.go` binds `m=a-b`, `n=c`.

To keep the rule easy to predict, a `{name}` capture must be bordered by literal
text, a `/`, or an end of the glob. It may not touch another capture, `*`, `?` or
an alternation. `src/*{n}.go` is rejected; in a regular-expression matcher it
would bind `n=c` on `src/abc.go`, which no one means.

The longest-first rule is the **definition**, and a reference matcher implements it by
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
rejected as pointless: delete that match glob instead. A constant `related` is
checked against the rule's `match` and `exclude` together.

### `related`

A template for exactly one path:

- non-empty, relative, not ending in `/`, with no `.` or `..` segment, no `\` and
  no glob characters (`* ? [ ]`);
- a clean path once its variables are resolved: the same blank-character and
  drive-prefix rules apply to the text (see Matching), and no `{}` holds a blank;
- `{dir}` and a `{**name}` capture fill a whole segment, because either can be
  empty: `{dir}.md` would become `.md` at the root and is rejected. A template
  made only of such variables is rejected too, since it could resolve to nothing;
- variables are `{dir}` and the names `match` captures, nothing else, and two
  variables may not touch (`{a}{b}`);
- at most one distinct variable that can contain `/`: `{dir}` or the `{**name}`
  capture;
- every capture must be used.

Resolution substitutes every variable **in one pass**, so a captured value is
never scanned for variables, then drops empty segments, so an empty `{dir}` or
`{**name}` leaves no stray slash and the result is never absolute. The resolved
path must be non-empty, have no `.` or `..` segment, contain no `\`, and not
start with a Windows drive, meaning a first segment that begins with an ASCII
letter and a colon. The resolved path must be a clean path (the definition under
Matching above), and input paths pass the same check
([`02-semantics.md`](02-semantics.md) §6). A captured value can be `..` (a file named
`..x.go` matched by `{n}x.go`), and without this check `{n}/{n}/x.txt` would
resolve outside `--cwd`. The check is run-level: one such file name stops the
whole run with exit 2 rather than one instance. Rules 9 and 10 also stop the whole
run.

A `related` with no variables is a fixed path, so it is checked against the
rule's own `match` when the config loads.

## `expect`

| Value | True when the related path |
| --- | --- |
| `touched` | is part of this change (added, modified, or a rename destination) |
| `untouched` | is unaffected by this change: not touched, and not removed by it |
| `exists` | is in state `exists`: present in the working tree and not removed by this change |
| `missing` | is in state `removed` or `missing`: not there after this change |
| `none` | no assertion; the rule is informational |

A list is AND-combined. Kyn observes paths, never content: `touched` means the
path is in the change, not that its content is correct.

Plain `expect: touched` is the form to reach for. `[exists, touched]` differs
from it only in plain-path input, where a listed file need not be on disk, so it
is rarely needed.

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
| 2 | `match` has at least one glob, and every `match` and `exclude` glob is in the dialect above: nothing from the rejected table, `**` only as a whole segment, alternatives literal, non-empty and unable to spell `.` or `..`. Captures: valid names, unique within a glob, not `dir`; `{**name}` fills a whole segment, at most one per glob; a `{name}` capture does not touch another capture, `*`, `?` or an alternation; no capture has `**` or `{**name}` on both sides; no captures in `exclude`; an `exclude` is not identical to a `match` glob |
| 3 | Every glob binds the same capture names, each of the same kind (`{x}` or `{**x}`). `related` is non-empty and relative, has no `.`/`..` segment, `\`, trailing `/` or glob character, uses only `{dir}` and captured names, has no two adjacent variables, and uses at most one distinct variable that can contain `/`, and gives `{dir}` and `{**name}` a whole segment each with at least one other segment present. A `related` with no variables must not match its own rule. A list-valued `related` is an error saying `related` names exactly one path and that either-or rules are a stated limit |
| 4 | Every capture is used in `related`; a rule with no `related` therefore has no captures. This also catches `{ts}` written where `{ts,html}` was meant |
| 5 | `expect` is required with valid values. `none` stands alone. `touched` with `untouched`, and `exists` with `missing`, contradict and are rejected |
| 6 | `related` is required unless `expect: none`. `skipIf` requires `related`. `skipIf` and `changeType` take only their listed values |
| 7 | The rule must be able to pass and able to fail in a state that can really occur, unless it is informational. Kyn enumerates the four reachable (state, touched) pairs of [`02-semantics.md`](02-semantics.md) §4, drops the pairs `skipIf` skips, and rejects a rule whose `expect` is false on every remaining pair (it can never pass) or true on every remaining pair (it can never fail). `expect: none` is exempt: it is true on every pair by design. Examples: `skipIf: exists` with `exists` or `touched` (never passes) or with `missing` (never fails; use `none` with `flags`); `skipIf: missing` with `[touched, missing]` (never passes); an unconditional `[touched, missing]` (never passes). In plain modes nothing is `removed` and a listed related file need not be on disk, so two accepted combinations are vacuous there: `skipIf: missing` with `expect: missing` can never pass, and `skipIf: missing` with `expect: exists` can never fail. Git mode needs both, so they are not rejected |
| 8 | `expect: none` requires `severity: info` and at least one of `message` or `flags` |

### At resolve time

These need a concrete path, so they fire only when a matching file changes. They
run for **every instance, before `changeType` or `skipIf` can skip it**, so a
skipped instance cannot hide a broken rule.

| # | Rule |
| --- | --- |
| 9 | A path matching two `match` globs with different bindings is an error |
| 10 | A resolved `related` path that matches its own rule's `match` and is not excluded is an error |
| 11 | A resolved `related` path that is not a clean path (defined under Matching) is an error. Input paths use the same predicate |

```text
rule "go-test": related path "src/a_test.go" also matches this rule's own
match; add an exclude or narrow the match
```

Because 9 to 11 depend on which files change, a config can load cleanly and fail
on someone's unrelated change. To exercise them against every existing tracked
path before that happens, resolve all of them (run it in CI on pull requests that change the config; the natural rule `src/**/{name}.ts` with `{dir}/{name}.spec.ts` and no `exclude` fails rule 10 only on a change that touches a `.spec.ts`, so it is the easiest way to learn about it before someone else does):

```bash
git ls-files | kyn check --stdin --dry-run-resolve --fail-on-empty
```

This samples the **tracked paths that are on disk today**: plain input drops
listed sources that are not on disk, and no related path is ever listed as
removed. It cannot know a source file someone will add tomorrow, so it narrows
the risk rather than removing it. `--fail-on-empty` makes a run that resolved
nothing fail (exit 1), which is what happens when git runs in another directory
than `--cwd`. It is a global guard, so the preview also prints a count for every
rule, zeros included (`03` §9): a misspelled rule shows as `0` while another rule
matches. Git still quotes a non-ASCII name in `ls-files` output unless run with
`-c core.quotePath=false`, and quotes a name containing a tab, a newline, `"` or
`\` regardless; those lines are rejected by name, not skipped, so the run stops
and tells you. Use `git -c core.quotePath=false ls-files` in a repository with
non-ASCII names.

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

Limits found so far, by nine reviews and the author's scenarios. The list has
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
| Guaranteeing one failure per related file | `related: "docs/{a}-{b}.md"` over `src/{a}_{b}.go`: `x-y_z.go` and `x_y-z.go` are different instances demanding the same `docs/x-y-z.md` | Instances are keyed on variable values, and distinct values can resolve to one path when literal text between variables can also occur inside them. Rare, and reported as two failures, not hidden. |
| Forbidding edits with no related path | "files under `db/migrations/` must never be modified" | `expect: untouched` needs a related path, and a rule whose related file is its own match is rejected by rule 10 |

Things Kyn does not detect, and says so rather than guess:

- An `exclude` that happens to cover everything `match` selects. The rule simply
  has no instances; `--fail-on-empty` reports a run where no rule had any.
- That `exists` is a working-tree fact: an untracked, ignored or generated file
  counts (a `gen/*.pb.go` built earlier in CI), and a sparse checkout hides files.
  Git mode requires the head to be checked out ([`02-semantics.md`](02-semantics.md) §6).
- A `related` that resolves to a directory. It is `exists` and can never be
  `touched`, since only files appear in a change.
- Anything that depends on the operating system's file naming. Matching and the
  change set are case-sensitive and compare code points, but the working tree
  that decides `exists` is the machine's. On a case-insensitive filesystem, a
  related template `docs/Readme.md` finds an on-disk `README.md` (`exists`, never
  `touched`), a case-only rename looks like two files, and NFD and NFC names can
  be one file. Windows trailing dots and spaces are not handled either. Backslash
  and drive prefixes are rejected (rule 11), so those cannot reach a path.
- A file whose name makes a capture `.`, `..`, or a segment with a blank at its
  edge (`.x.go` under `{n}x.go`; `a ` under `{n}.go`), so the
  resolved path is not clean. Rule 11 stops the whole run, since the alternative
  is a per-instance error kind that no other rule needs. It is rare, and the
  message names the path.

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

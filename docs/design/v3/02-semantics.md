# Semantics

Status: proposed. Current state only; history is in
[`notes.md`](notes.md). Claims about today's behavior cite an observation (OS1,
OS2, ...) listed there.

## 1. What Kyn observes

For each related path Kyn knows two independent things.

| Fact | Source |
| --- | --- |
| Is it **touched**: part of this change? | The change set: added, modified, and rename-destination paths |
| What **state** is it in? | The working tree plus the removals in the change |

Kyn never reads content. `touched` means the path is in the change, nothing
more (OS1: an assertion on membership passes even for a path that does not
exist on disk, when the diff and the tree disagree).

## 2. The three states of a related path

| State | Condition |
| --- | --- |
| `removed` | **the change record says** this change deleted it or renamed it away, whatever the working tree holds |
| `exists` | not `removed`, and present in the working tree |
| `missing` | neither: not recorded as removed, and not in the tree |

The three are checked in that order, so `removed` wins. The change record is
authoritative because the working tree can be a different checkout from the one
the diff describes: with `--base main --head origin/feature` run from a checkout
that still has the file the feature branch deleted, the file is `removed`, not
`exists`. Otherwise the same diff would fail on one machine and pass on another,
and `expect: untouched` would pass for a file the change deleted. `skipIf` reads
the same state, so a removal is never skipped from a stale tree either.

Today Kyn has only "exists on disk or not", so a deleted test and a test that
never existed look identical, and the rule protecting the test is skipped
either way (OS3). Renaming the test away does the same (OS8).

The `removed` state costs nothing to compute. `git diff --name-status -M`,
which Kyn already runs, prints `D path` and `R100 old new`; today the `D` line
and the `old` half of a rename are parsed and discarded
(`internal/changes/git.go:63-64` and `:48-53`). Retaining them is the whole
change.

Removed paths do **not** join the change set. If they did, `expect: touched`
would pass for a file the change deleted.

## 3. `skipIf` over the three states

A removal never causes a skip: each value skips on exactly the state of the same
name, and `removed` falls through to the rule in both cases.

| `skipIf` | `exists` | `removed` | `missing` |
| --- | --- | --- | --- |
| *(omitted)* | applies | applies | applies |
| `missing` | applies | applies | skipped |
| `exists` | skipped | applies | applies |

For `expect: touched`:

| Scenario | State | Touched | Today | `skipIf: missing` |
| --- | --- | --- | --- | --- |
| related untouched | `exists` | no | FAIL | FAIL |
| related updated | `exists` | yes | pass | pass |
| related **deleted** | `removed` | no | skipped | **FAIL** |
| related **renamed away** | `removed` | no | skipped | **FAIL** |
| related never existed | `missing` | no | skipped | skipped |

There is no `skipIf: exists` column: that gate applies only when the related file
is not there, so `expect: touched` under it is rejected at load (validation rule
7 in [`04-grammar.md`](04-grammar.md)). `skipIf: exists` is for `none` and
`untouched`; with `missing` it could never fail, so validation rejects it.

Against today's behavior, `skipIf: missing` changes exactly the two rows it is
meant to. The never-existed case still skips.

The state alone decides a skip. A related path that is in the diff but not in the
working tree (the `missing`, touched row of §4) is `missing`, so `skipIf: missing`
skips it even though it is touched.

v2's `if.kinMissing` fires on a file deleted in this change, because it tests
only whether the file is on disk. `skipIf: exists` does the same, so that
behavior is preserved.

## 4. `expect` over state and membership

| `expect` | Passes when |
| --- | --- |
| `touched` | the path is in the change set |
| `untouched` | the path is not in the change set **and** is not `removed` |
| `exists` | state is `exists` |
| `missing` | state is `removed` or `missing` |
| `none` | always; the result is `info` |

`untouched` means "this change did not affect the path". A deleted file was
affected, so a rule meant to keep a file frozen fails if the file is deleted as
well as if it is edited. Today's `kinUnchanged` passes for a deleted file and
its trace prints "unchanged" for a path that no longer exists (OS3).

State and membership are independent except for one constraint: a `removed`
path is never touched. That leaves five reachable combinations.

| State | Touched | Reachable? |
| --- | --- | --- |
| `exists` | yes / no | yes |
| `removed` | no | yes (deleted or renamed away, whether or not the tree still has it) |
| `removed` | yes | no |
| `missing` | no | yes |
| `missing` | **yes** | **yes**: the diff contains the path but the working tree does not, e.g. `--base main --head origin/feature` from another checkout (OS1) |

The last row is the only one that separates `expect: touched` from
`expect: [exists, touched]`, so the conformance suite must include it, and so the
stale-checkout deletion above ([`07-conformance.md`](07-conformance.md)).

## 5. Result status

| Status | When |
| --- | --- |
| `skipped` | `skipIf` excluded the instance, or `changeType` excluded it |
| `fail` | the rule applied and an expectation was false |
| `pass` | the rule applied and every expectation was true |
| `info` | the rule applied and `expect` is `none` |

A `skipped` or `fail` result carries a stable **reason code** next to the
message, so a custom `message` (which says what the rule is for) never hides what
actually went wrong. A preset message such as "Component changed but its story
did not" is the wrong advice when the story was deleted; the reason says so.

| Status | `reason` | Text shown |
| --- | --- | --- |
| `skipped` | `related_missing` | `related is missing` |
| `skipped` | `related_exists` | `related exists` |
| `skipped` | `change_type` | `no source has a listed changeType` |
| `fail` | `removed`, `missing`, `not_touched`, `touched`, `exists` | the last column of the table below |

`pass` and `info` carry no reason. The text report prints the authored (or
generated) message, then the reason text; JSON has both fields.

`flags` are added whenever the rule applies, pass or fail, and never when it is
skipped. This matches v2 (`internal/rules/evaluator.go:47`) and is why
`skipIf: exists` exists: "flag only when the file is absent" cannot be written
as an assertion.

When a failure has several causes, the message names the most fundamental, in
this order:

| Cause | Message |
| --- | --- |
| the path was removed by this change | `removed by this change` |
| the path is not there and was never there | `does not exist` |
| the path is there but is not in the change | `not touched` |
| the path is in the change but must not be | `was touched` |
| the path exists but `expect: missing` | `exists but must not` |

## 6. Input modes

| Mode | Change set | Removal information | Source statuses |
| --- | --- | --- | --- |
| Git (`--base`/`--head`, or automatic) | from the diff | yes | yes |
| `--files-from` or `--stdin` with `git diff --name-status -M` lines | from the lines | yes | yes |
| `--files`, or `--files-from` / `--stdin` with plain paths | the listed paths | **no** | all `modified` |

`--stdin` is `--files-from -`: the same reader and the same grammar, not a second
format. **Enforcement in CI should always feed the status form**
(`git diff --name-status -M base...head | kyn check --stdin`), including the
container job, because plain paths lose deletion protection (below). `--files`
is the quick way to try a rule locally. The help text and every example in the
docs follow this.

### `--files-from` format

Each line is read on its own: a line containing a TAB is a name-status line, and
a line with none is a plain path, so one file may mix both. The status field is
exactly `A`, `M`, `D` or `T` (a type change, which counts as `modified` and
touched) with one more field (the path), or `R` followed by
a score from 0 to 100, as git prints it (`R100`, `R095`), with two more (the old
and new paths). Anything else is an error naming the line: another status, a bare
`R`, a score above 100, extra fields. `C` and `X` are errors because Kyn never
asks git for copies, and no other status is one it expects; `U` (an unmerged
path) is an error because a conflicted change set is not one to enforce, and the
message says to resolve the merge first.

A line with no TAB that **starts with a status token and a space** (`D src/a.go`,
`M  src/b.go`, `?? src/c.go`, `AM src/d.go`, `R100 a b`) is rejected too, naming
the line: it is a name-status line whose TAB became spaces, or `git status`
output, and reading it as a file called `D src/a.go` would silently evaluate
nothing. The tokens are one letter from `A M D T U C X R`; any two letters from
`M T A D R C U` (`AM`, `MM`, `RM`, and the unmerged pairs `UU AA DD AU UA DU UD`);
`R` or `C` with digits; `??` and `!!`. The unmerged ones get the "resolve the
merge first" message. The error names the way out: a tracked file whose name
really begins that way is fed through the status form (`M<TAB>M foo.go`). `--files`
has no status form, so each of its comma-separated entries gets the same check and
the way out is `--files-from`. Both halves of a rename matter: the destination
joins the change set, the source is what makes a related path `removed`.

```text
M	src/a.go
D	src/a_test.go
R100	src/old.go	src/new.go
src/plain-path.go
```

Plain single-path lines mean `modified`. Empty and whitespace-only lines are
ignored, and a trailing CR is removed, so CRLF input works. Every path that is added,
modified, or a rename destination, in every input mode including `--files`, must
be a **clean path** or the run fails naming it: the one predicate defined in
[`04-grammar.md`](04-grammar.md) under Matching, which rule 11 also applies to
resolved paths. A path that is only **removed** (a `D` path, a rename source)
is a state key and is never matched, so it is exempt: the change that fixes a bad
file name must be able to pass. For `--files-from` and `--files` (not git mode, which
reads paths with quoting disabled) every path, removed ones included, also may not
be in git's quoted form. This is fatal for every changed path, matched or not, so a
repository that tracks a name like `Q: and A.md`, `old /x.txt` or `a\b.txt`
(legal on Linux and macOS) fails every run whose change touches it
([`06-cutover.md`](06-cutover.md)). Nothing is
trimmed, converted or cleaned for the user: v2 turned `\` into `/` and stripped
`./`, and v3 rejects both ([`06-cutover.md`](06-cutover.md)). A path may appear on
several lines only if the lines are identical; otherwise (`D x` and `A x`, or a
rename destination that is also listed alone) the run fails naming the path,
because no order of precedence is right. Git never emits such a diff. The
statuses are exactly the five above; every other status is an error.

The `git ls-files` recipe ([`04-grammar.md`](04-grammar.md)) feeds plain paths,
so in a repository that tracks a name starting with a status token and a space
(`M Notes.md`) it fails with the status-like error; use `--files-from` with the
status form there.

Paths that git would quote (non-ASCII or special characters) are not supported in
the line-oriented stream. `git -c core.quotePath=false` fixes non-ASCII names but
still quotes a name containing a tab, newline, `"` or `\`; those lines are
rejected, not skipped. Git mode does not have this limit: Kyn reads git with
`git diff --name-status -z -M`, which prints every name literally (measured), so
a `"` in a name is legal there. v2 reads quoted output, which is OS16, and the
cutover fixes it ([`06-cutover.md`](06-cutover.md)).

### What plain-path modes lose

Plain paths carry no removal information and no statuses, so a policy tried with
`--files` can pass locally and fail in git-mode CI. The differences are not a
list to memorize: the table below follows from three facts about a plain-path
line, and a difference that does not follow from them is a bug in this section.

1. The listed paths are the whole change set, and every one is `modified`.
2. Nothing is `removed`: a path that is not on disk is `missing`.
3. A listed source that is not on disk is dropped. This one is a decision, not a
   consequence: a deleted source and a mistyped path look the same in a plain
   list, and evaluating a typo would fail every run. It is limited to sources;
   dropping is never applied to a related path (see below).

The facts apply per line, so in a mixed file only the plain lines are affected.

| # | Difference | Git mode | `--files`, or plain `--files-from` |
| --- | --- | --- | --- |
| 1 | A related file deleted by the change and **not listed** | state `removed`: `skipIf: missing` applies, `expect: untouched` fails | state `missing`: `skipIf: missing` **skips**, `expect: untouched` **passes** |
| 2 | A related file deleted by the change and **listed** (what `git diff --name-only` prints) | `removed`, not touched: `expect: touched` **fails**, `untouched` fails | `missing` and **touched**: `expect: touched` **passes**, `untouched` fails "was touched"; `skipIf: missing` skips |
| 3 | Source statuses | `added`, `modified`, `renamed` | all `modified`, so `changeType: [added]` never matches |
| 4 | A listed **source** that is not on disk | `D`, or the old name of a rename: no instance. `A` or `M` but absent from disk (a checkout without the file, OS1): an instance is created, its related path is evaluated, and rules 9 to 11 run for it | dropped, whatever it was, because a deleted source and a typo look the same; **no instance, so no resolve-time check, runs for it** (the same config can be fine in plain mode and stop a git-mode run under rule 10) |

Row 2 reopens the delete-to-silence hole for a policy run on `--name-only`
output: deleting the test makes `expect: touched` pass. Feed `--files-from` the
`--name-status` form to close it. `skipIf: exists` and `expect: missing` behave
the same in both. The tutorial must say so where `--files` is introduced.
Inferring deletion from "listed but not on disk" for a related path was
considered and rejected: it would turn typos into deletions. Fact 3 applies to
sources only, and the recipe in [`04-grammar.md`](04-grammar.md) relies on it.

## 7. `changeType`

`changeType` reads the statuses of the matched **source** files; it says nothing
about the related file. A rule with `changeType: [added]` applies to an instance
if **any** of its sources was added. Instances it excludes are reported as
`skipped`, not dropped, so a rule that never applies is visible.

`changeType` is meaningful in the first two input modes. In plain-path modes
every source is `modified` (§6, row 3). It filters instances after the resolve-time
checks have run, so a rule whose `changeType` excludes an instance still reports a
broken `related` ([`04-grammar.md`](04-grammar.md), rules 9 to 11).

Deleted source files never create instances: a removed source has no
working-tree path to match or resolve from. A source in the diff as added or
modified but absent from the working tree does create one in git mode (§6, row
4). "Delete the handler, delete its
test" is therefore not expressible in v3.0.

## 8. Severity and exit code

Unchanged: a run fails (exit 1) when a failing result's severity is at or above
`--fail-on`, which defaults to `error`.

`expect: none` rules must be `severity: info`. Measured (OS14): a v2
informational rule with `severity: error` reports `errors: 1` beside
`failed: 0` and `ok: true`, so a rule that cannot fail inflates the error
count.

`--fail-on-empty` fails the run (exit 1) when no rule produced any instance,
skipped or not. It applies to `--dry-run-resolve` too, which today exits 0 with
zero matches because preview returns before the check. It is a global guard: it
cannot see one misspelled rule while another matches, which is what the per-rule
counts in [`03-instances-and-reports.md`](03-instances-and-reports.md) §9 are for.

## 9. Generated messages

`message` is optional. The generated text is a versioned, fixture-tested
contract containing only the rule ID and the resolved related path. It never
contains a config path or an absolute path, which would leak the CI checkout
root into review comments.

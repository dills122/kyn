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
exist on disk, in plain-path input, where a listed file need not be on disk).

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

The state alone decides a skip. In plain-path input a related path that is listed
but not on disk is `missing` and touched (§4), so `skipIf: missing` skips it even
though it is touched.

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
path is never touched. In git mode and `--name-status` mode that leaves **four**
reachable combinations, because a changed path that is not in the working tree is
an error (§6).

| State | Touched | Reachable in git mode? |
| --- | --- | --- |
| `exists` | yes / no | yes |
| `removed` | no | yes (deleted or renamed away, whether or not the tree still has it) |
| `removed` | yes | no |
| `missing` | no | yes |
| `missing` | yes | **no**: the working tree must contain every added or modified path |

Plain-path input adds a fifth, `missing` and touched: a related path that is
listed but not on disk. Validation (rule 7) uses the four git-mode pairs; the fifth
exists only where removal information does not.

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

`changeType` is applied before `skipIf`, so an instance both exclude reports `change_type`. `pass` and `info` carry no reason. The text report prints the authored (or
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
| `--files-from` or `--stdin` with `--name-status` | from the status lines | yes | yes |
| `--files`, or `--files-from` / `--stdin` without `--name-status` | the listed paths | **no** | all `modified` |

`--stdin` is `--files-from -`: the same reader and the same grammar, not a second
format. **The format is declared, never guessed.** Without `--name-status` every
line is a plain path, and a line containing a TAB is an error saying to pass
`--name-status`. With it every line must be a status line, and a plain path is an
error. A CI step that changes `git diff --name-status` to `--name-only` therefore
fails closed instead of quietly losing deletion protection. The report always says
which it read: `inputMode` is `git`, `name-status` or `paths`, and `removalInfo`
is false for `paths` ([`03-instances-and-reports.md`](03-instances-and-reports.md)
§8).

**Enforcement in CI always feeds the status form**
(`git -c core.quotePath=false diff --relative --name-status -M base...head | kyn check --stdin --name-status`),
including the container job, because plain paths lose deletion protection
(below). `--files` is the quick way to try a rule locally. The help text and every
example in the docs follow this.

### Git mode and the working tree

Git mode runs `git diff --relative --name-status -z -M`. `-z` prints every name
literally, so no path is quoted. `--relative` makes paths relative to `--cwd` and
drops paths outside it; without it a `--cwd` below the repository root receives
root-relative paths, matches nothing and passes. `exists` and `missing` are facts
about the working tree, so **the head being evaluated must be checked out**: in git
mode and in `--name-status` mode, a path the change adds, modifies, or renames to
that is not in the working tree is an error naming the path (exit 3 for git mode,
2 for `--name-status` input): `src/a.go is added in base...head but is not in the
working tree; check out head`. This is deliberate. Without it `exists`, `missing`
and both `skipIf` values would depend on which checkout ran the diff, in the
direction of an addition as well as a deletion. `removed` has no such check
because the change record, not the tree, decides it (§2). Plain input has no
change record to compare against; see below.

`exists` is a working-tree fact in every sense: an untracked, ignored or
generated file counts (a `gen/*.pb.go` that CI builds first), and a sparse
checkout hides files. A related path whose symlink components resolve outside
`--cwd` is a runtime error, as in v2.

### `--name-status` format

The status field is exactly `A`, `M`, `D` or `T` (a type change, which counts as
`modified` and touched) with one more field (the path), or `R` followed by a
score from 0 to 100, as git prints it (`R100`, `R095`), with two more (the old and
new paths). Fields are separated by TAB. Anything else is an error naming the
line: another status, a bare `R`, a score above 100, extra fields. `C` and `X` are
errors because Kyn never asks git for copies, and no other status is one it
expects; `U` (an unmerged path) is an error because a conflicted change set is not
one to enforce, and the message says to resolve the merge first. Both halves of a
rename matter: the destination joins the change set, the source is what makes a
related path `removed`.

```text
M	src/a.go
D	src/a_test.go
R100	src/old.go	src/new.go
```

Without `--name-status`, input is one path per line, so a file named `M foo.go`
is an ordinary name. If someone pipes `git status --short` into it, the lines are
read as names, are not on disk, and the sources are dropped (fact 3 below); the
report's `inputMode: "paths"` and `--fail-on-empty` are the guards.

Empty and whitespace-only lines are ignored, and a trailing CR is removed, so
CRLF input works. A path that a rule matches as a source, a resolved related path
(rule 11), and every path read from `--files` or `--files-from` that is in git's
quoted form (a leading `"`) must be a **clean path**: the one predicate defined in
[`04-grammar.md`](04-grammar.md) under Matching, and the run fails naming the
path otherwise. A changed path that no rule matches is never evaluated, so it is
not checked: a repository that tracks `Q: and A.md` or `a\b.txt` (legal on Linux
and macOS) fails only runs where a rule would match that file
([`06-cutover.md`](06-cutover.md)). A path that is only **removed** (a `D` path, a
rename source) is a state key and is never matched, so it is exempt: the change
that fixes a bad file name must be able to pass. Nothing is trimmed, converted or
cleaned for the user: v2 turned `\` into `/` and stripped `./`, and v3 rejects
both ([`06-cutover.md`](06-cutover.md)). A path may appear on several lines only
if the lines are identical; otherwise (`D x` and `A x`, or a rename destination
that is also listed alone) the run fails naming the path, because no order of
precedence is right. Git never emits such a diff.

Errors in input the user supplied (`--files`, `--files-from`, `--stdin`) exit 2.
Errors in what git printed (a bad path, a path missing from the tree) exit 3.

The `git ls-files` recipe ([`04-grammar.md`](04-grammar.md)) feeds plain paths.
Plain paths cannot carry a tab or newline, and `git -c core.quotePath=false`
fixes non-ASCII names but still quotes a name containing `"` or `\`; those lines
are rejected, not skipped. Git mode does not have these limits (`-z`). v2 reads
quoted output, which is OS16; v2 now reads `-z` too, and the cutover keeps it
([`06-cutover.md`](06-cutover.md)).

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

| # | Difference | Git mode | `--files`, or plain `--files-from` |
| --- | --- | --- | --- |
| 1 | A related file deleted by the change and **not listed** | state `removed`: `skipIf: missing` applies, `expect: untouched` fails | state `missing`: `skipIf: missing` **skips**, `expect: untouched` **passes** |
| 2 | A related file deleted by the change and **listed** (what `git diff --name-only` prints) | `removed`, not touched: `expect: touched` **fails**, `untouched` fails | `missing` and **touched**: `expect: touched` **passes**, `untouched` fails "was touched"; `skipIf: missing` skips |
| 3 | Source statuses | `added`, `modified`, `renamed` | all `modified`, so `changeType: [added]` never matches |
| 4 | A listed **source** that is not on disk | `D`, or the old name of a rename: no instance. `A`, `M` or a rename destination that is absent is an error (above) | dropped, whatever it was, because a deleted source and a typo look the same; **no instance, so no resolve-time check, runs for it** (the same config can be fine in plain mode and stop a git-mode run under rule 10) |

Row 2 reopens the delete-to-silence hole for a policy run on `--name-only`
output: deleting the test makes `expect: touched` pass. Feed the `--name-status`
form to close it. `skipIf: exists` and `expect: missing` behave
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
zero matches because preview returns before the check. `explain` takes no
`--fail-on-empty` and never exits 1 for a result. It is a global guard: it
cannot see one misspelled rule while another matches, which is what the per-rule
counts in [`03-instances-and-reports.md`](03-instances-and-reports.md) §9 are for.

## 9. Generated messages

`message` is optional. The generated text is a versioned, fixture-tested
contract containing only the rule ID and the resolved related path. It never
contains a config path or an absolute path, which would leak the CI checkout
root into review comments. The exact strings, by `reason`:

| `reason` | Generated message |
| --- | --- |
| `removed` | `<ruleId>: <related> was removed by this change` |
| `missing` | `<ruleId>: <related> does not exist` |
| `not_touched` | `<ruleId>: <related> was not touched` |
| `touched` | `<ruleId>: <related> was touched` |
| `exists` | `<ruleId>: <related> exists but must not` |

An authored `message` replaces this text in the `message` field; the `reason` code
and the reason text stay beside it.

# Semantics

Status: proposed. Current state only; history is in
[`00-workplan.md`](00-workplan.md). Claims about today's behavior cite an
observation in [`01-observed-semantics.md`](01-observed-semantics.md).

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
| `exists` | present in the working tree |
| `removed` | not in the tree, and this change deleted it or renamed it away |
| `missing` | neither: it was not there before and is not there now |

Today Kyn has only "exists on disk or not", so a deleted test and a test that
never existed look identical, and the rule protecting the test is skipped
either way (OS3). Renaming the test away does the same (OS8).

The `removed` state costs nothing to compute. `git diff --name-status -M`,
which Kyn already runs, prints `D path` and `R100 old new`; today the `D` line
and the `old` half of a rename are parsed and discarded
(`internal/changes/git.go:45`, `:60`). Retaining them is the whole change.

Removed paths do **not** join the change set. If they did, `expect: touched`
would pass for a file the change deleted.

## 3. `skipIf` over the three states

| `skipIf` | `exists` | `removed` | `missing` |
| --- | --- | --- | --- |
| *(omitted)* | applies | applies | applies |
| `missing` | applies | **applies** | skipped |
| `exists` | skipped | skipped | applies |

Measured against real git output by
[`e9-gate-prototype.sh`](experiments/e9-gate-prototype.sh), for
`expect: touched`:

| Scenario | State | Touched | Today | `skipIf: missing` | `skipIf: exists` |
| --- | --- | --- | --- | --- | --- |
| related untouched | `exists` | no | FAIL | FAIL | skipped |
| related updated | `exists` | yes | pass | pass | skipped |
| related **deleted** | `removed` | no | skipped | **FAIL** | skipped |
| related **renamed away** | `removed` | no | skipped | **FAIL** | skipped |
| related never existed | `missing` | no | skipped | skipped | FAIL |

Against today's behavior, `skipIf: missing` changes exactly the two rows it is
meant to. The never-existed case still skips.

## 4. `expect` over state and membership

| `expect` | Passes when |
| --- | --- |
| `touched` | the path is in the change set |
| `untouched` | the path is not in the change set |
| `exists` | state is `exists` |
| `missing` | state is `removed` or `missing` |
| `none` | always; the result is `info` |

State and membership are independent except for one constraint: a `removed`
path is never touched. That leaves six reachable combinations.

| State | Touched | Reachable? |
| --- | --- | --- |
| `exists` | yes / no | yes |
| `removed` | no | yes (deleted, or renamed away) |
| `removed` | yes | no |
| `missing` | no | yes |
| `missing` | **yes** | **yes**: the diff contains the path but the working tree does not, e.g. `--base main --head origin/feature` from another checkout (OS1) |

The last row is the only one that separates `expect: touched` from
`expect: [exists, touched]`, so the conformance suite must include it
([`07-conformance.md`](07-conformance.md)).

## 5. Result status

| Status | When |
| --- | --- |
| `skipped` | `skipIf` excluded the instance |
| `fail` | the rule applied and an expectation was false |
| `pass` | the rule applied and every expectation was true |
| `info` | the rule applied and `expect` is `none` |

`flags` are added whenever the rule applies, pass or fail, and never when it is
skipped. This matches v2 (`internal/rules/evaluator.go:47`) and is why
`skipIf: exists` exists: "flag only when the file is absent" cannot be written
as an assertion.

When a failure has several causes, the message names the most fundamental:
"does not exist" before "not touched".

## 6. Input modes

| Mode | Change set | Removal information | Source statuses |
| --- | --- | --- | --- |
| Git (`--base`/`--head`, or automatic) | from the diff | yes | yes |
| `--files-from` with `git diff --name-status` lines | from the lines | yes | yes |
| `--files`, or `--files-from` with plain paths | the listed paths | **no** | all `modified` |

`--files-from` accepts git's own format verbatim, including the three-field
rename line. Both halves of a rename matter: the destination joins the change
set, the source is what makes a related path `removed`.

```text
M	src/a.go
D	src/a_test.go
R100	src/old.go	src/new.go
```

Plain single-path lines still mean `modified`.

Without removal information the `removed` state cannot be told from `missing`.
The consequences per gate:

| `skipIf` | In a plain-path mode, a deleted related file |
| --- | --- |
| *(omitted)* | behaves the same as in git mode |
| `missing` | is **skipped** (today's behavior), where git mode would apply the rule |
| `exists` | **applies** the rule, where git mode would skip it |

So a policy tried with `--files` can pass locally and fail in git-mode CI, or
raise a flag locally that CI does not. The tutorial must say this where
`--files` is introduced. Inferring deletion from "listed but not on disk" was
considered and rejected: it would turn typos into deletions.

`changeType` reads source statuses, so it is meaningful in the first two modes
and matches only `modified` in the third.

## 7. `changeType`

A rule with `changeType` applies to an instance if **any** of its matched
source files has a listed status. Deleted source files never create instances:
a removed source has no working-tree path to match or resolve from. "Delete the
handler, delete its test" is therefore not expressible in v3.0.

## 8. Severity and exit code

Unchanged: a run fails (exit 1) when a failing result's severity is at or above
`--fail-on`, which defaults to `error`.

`expect: none` rules must be `severity: info`. Measured (OS14): a v2
informational rule with `severity: error` reports `errors: 1` beside
`failed: 0` and `ok: true`, so a rule that cannot fail inflates the error
count.

## 9. Generated messages

`message` is optional. The generated text is a versioned, fixture-tested
contract containing only the rule ID and the resolved related path. It never
contains a config path or an absolute path, which would leak the CI checkout
root into review comments.

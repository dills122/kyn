# Observed v2 semantics

Status: measured ground truth. Not a proposal.

Statements here were produced by running the `kyn` binary built from this
branch against a generated fixture. Reproduction scripts live in
[`experiments/`](experiments/) and each observation cites the one that produced
it. Two are exceptions and say so: OS2 is an analysis of OS1's table, and OS10
is a search of the source.

The purpose is to give the v3 design a factual baseline. It also turned up
defects in shipped v0.1.3 that are independent of v3: non-deterministic error
selection (OS7), a policy gate silenced by deleting or renaming the file it
protects (OS3, OS8), a preset that fails on an ordinary change set (OS11), a
`PASS` headline above a failure count (OS12), and non-ASCII filenames that never
match any rule (OS16).

## Method

```bash
go build -o /tmp/kyn ./cmd/kyn
bash docs/design/v3/experiments/<script>.sh
```

Scripts take the binary path from `$KYN` or default to a sibling build. They
create disposable fixtures under a scratch directory and never touch the repo.

## OS1 — Expectation truth table

Source: [`e3-truthtable.sh`](experiments/e3-truthtable.sh)

Each candidate v3 expectation, evaluated against every combination of *related
path exists on disk* and *related path present in the change set*, in `--files`
mode. `STATUS` is `explain`'s per-result status; `EXIT` is `check`'s exit code.

| Assertion | V2 clauses | exists | in change set | status | exit |
| --- | --- | --- | --- | --- | --- |
| `if.kinExists` then `kinChanged` | `if.kinExists` + `assert.kinChanged` | Y | Y | pass | 0 |
| | | Y | N | fail | 1 |
| | | N | Y | **skipped** | 0 |
| | | N | N | **skipped** | 0 |
| `kinExists` and `kinChanged` | `assert.kinExists` + `assert.kinChanged` | Y | Y | pass | 0 |
| | | Y | N | fail | 1 |
| | | N | Y | fail | 1 |
| | | N | N | fail | 1 |
| `kinChanged` | `assert.kinChanged` | Y | Y | pass | 0 |
| | | Y | N | fail | 1 |
| | | N | Y | **pass** | 0 |
| | | N | N | fail | 1 |
| `kinUnchanged` | `assert.kinUnchanged` | Y | Y | fail | 1 |
| | | Y | N | pass | 0 |
| | | N | Y | fail | 1 |
| | | N | N | pass | 0 |
| `kinExists` | `assert.kinExists` | Y | Y | pass | 0 |
| | | Y | N | pass | 0 |
| | | N | Y | fail | 1 |
| | | N | N | fail | 1 |
| `kinMissing` | `assert.kinMissing` | Y | Y | fail | 1 |
| | | Y | N | fail | 1 |
| | | N | Y | pass | 0 |
| | | N | N | pass | 0 |

## OS2 — Existence and change-set membership are orthogonal axes

OS1 reduces cleanly. Reading the table by which axis each expectation consults:

| Assertion | reads *exists* | reads *in change set* |
| --- | --- | --- |
| `kinChanged` | no | yes |
| `kinUnchanged` | no | yes (inverted) |
| `kinExists` | yes | no |
| `kinMissing` | yes (inverted) | no |
| `if.kinExists` then `kinChanged` | yes (as a gate) | yes |
| `kinExists` and `kinChanged` | yes (as an assertion) | yes |

Four of the six read exactly one axis. The two compound forms differ from each
other only in whether the existence check *gates* the rule (producing `skipped`)
or *asserts* it (producing `fail`). This is the whole of the vocabulary, and it
is a 2x2 with one axis optionally promoted to a gate — not an open-ended enum.

`kinExists` with `kinChanged`, and `kinChanged` alone, differ in exactly one
cell: `exists=N, in-change-set=Y`. That state is reachable when the selected diff and the working
tree disagree, for example `--base main --head origin/feature` evaluated from a
checkout that does not contain the added file. The distinction is real but
narrow, and any documentation must say so explicitly or users will treat the two
as synonyms.

## OS3 — Kyn cannot observe deletion, and the default shape is silenced by it

Source: [`e4-git-delete.sh`](experiments/e4-git-delete.sh)

`docs/decisions.md` excludes `D` paths from the change set, and
`internal/changes/git.go:63-64` discards them. The consequence is
larger than a vocabulary problem. In a real git repository where the source file
was modified and its related test was **deleted in the same commit**:

| Assertion | result | exit |
| --- | --- | --- |
| `if.kinExists` then `kinChanged` | **skipped** | 0 |
| `kinUnchanged` | **pass** | 0 |
| `kinChanged` | fail | 1 |

`explain` renders the `unchanged` case as:

```text
Assert:
  - assert.kinUnchanged: pass (rel unchanged (src/a_test.go))
```

The file does not exist. The trace states that it is unchanged.

`if.kinExists` then `kinChanged` is the shape every `kyn init` preset ships. Deleting the related test
silences the rule that exists to protect it, and `check` exits 0.

OS8 shows the same hole for renames. No current rule form can express "the
related file was deleted or renamed away in this change", because neither event
enters the change set. `assert.kinExists` catches the
after-effect — the path is gone from the working tree — but cannot attribute it
to this diff.

This is a capability gap, not a naming problem. Renaming `kinChanged` makes the
vocabulary honest; it does not make deletion observable.

The gap is closed in [`02-semantics.md`](02-semantics.md) §2 and §3.

## OS8 — Renaming the related file away is equally invisible

Source: [`e8-rename-blindness.sh`](experiments/e8-rename-blindness.sh)

`fromGitDiff` handles `R` by recording `fields[2]`, the rename destination, and
discarding `fields[1]`, the source
([`internal/changes/git.go:48-53`](../../../internal/changes/git.go)). The path the
policy was watching simply disappears from Kyn's view.

Source modified, related test renamed away in the same commit:

```text
--- diff under evaluation ---
M	src/a.go
R100	src/a_test.go	src/renamed_test.go
--- what kyn collected ---
  - src/a.go
  - src/renamed_test.go
--- result ---
Status: skipped
If:
  - if.kinExists: fail (rel missing (src/a_test.go))
check exit=0
```

Identical outcome to OS3. From the rule's perspective a rename-away and a
deletion are the same event, and both are silent. Any fix for OS3 that only
handles `D` leaves this half open.

## OS4 — `stripSuffixes` is a grouping input, not a source-shape detail

Source: [`e1-stripsuffix.sh`](experiments/e1-stripsuffix.sh)

The family instance key is `familyID + {dir}/{base}`, where `base` is the file
name with `stripSuffixes` removed
([`internal/family/resolver.go:53`](../../../internal/family/resolver.go),
[`internal/family/resolver.go:151`](../../../internal/family/resolver.go)).
`stripSuffixes` therefore decides how many instances exist, how many results are
emitted, and which related paths are demanded.

Given the shipped `api` preset shape — sources `order_handler.go` and
`order_service.go` changed together:

| Config | instances | results | failed | expected files |
| --- | --- | --- | --- | --- |
| with `stripSuffixes: ["_handler", "_service"]` | 1 | 1 | 1 | `internal/order/order_test.go` |
| without | 2 | 2 | 2 | `internal/order/order_handler_test.go`, `internal/order/order_service_test.go` |

Different policy, different result count, different demanded paths.

So `stripSuffixes` is not a presentation detail of the template: it decides the
policy. It has to be designed together with the template that consumes it.

## OS5 — The common form self-matches, and OS3 hides it

Source: [`e7-selfmatch.sh`](experiments/e7-selfmatch.sh)

`match: ["src/**/*.go"]` with `related: "{dir}/{name}_test.go"` and no `exclude`
matches the related file too. `src/a.go` and `src/a_test.go` changed together
produce **two** instances:

```text
[fam] src/a
Kin:
  - rel: src/a_test.go

[fam] src/a_test
Kin:
  - rel: src/a_test_test.go
```

The phantom instance demands `src/a_test_test.go`. It does not fail, because
`src/a_test_test.go` does not exist, so the `if.kinExists` gate skips it —
the OS3 silent skip conceals the OS5 mistake, and `check` exits 0 with `PASS`.

Both hazards come from the natural shape of a rule: a broad source glob and a
related template that lands inside it. Measured against real layouts, an
`exclude` is needed in the common case, and omitting it fails silently rather
than loudly.

## OS6 — Non-`source` groups are inert

Source: [`e5-inert-groups.sh`](experiments/e5-inert-groups.sh)

`Family.SourceInclude` / `SourceExclude` consult only `groups["source"]`
([`internal/config/config.go:57`](../../../internal/config/config.go)). Nothing
reads any other group. Validation nevertheless *requires* every declared group
to carry a non-empty `include`
([`internal/config/validate.go:73`](../../../internal/config/validate.go)), and
`kyn init` emits `groups.story` and `groups.tests` in its presets
([`internal/cli/init.go:160`](../../../internal/cli/init.go)).

Deleting those groups from the generated `web-ui` preset produces byte-identical
output in every mode:

| Mode | result |
| --- | --- |
| `text` | identical |
| `json` | identical |
| `sarif` | identical |
| `rdjson` | identical |
| `checkstyle` | identical |
| `--dry-run-resolve` | identical |

`if.changedAny: [source]` is likewise a no-op: `evalWhen` only tests
`len(inst.SourceFiles) == 0`
([`internal/rules/evaluator.go:110`](../../../internal/rules/evaluator.go)), and
an instance cannot exist with zero source files. Validation rejects any group
name other than `source` in `changedAny`
([`internal/config/validate.go:196`](../../../internal/config/validate.go)).

Every starter config therefore teaches two constructs that do nothing. This is
evidence for the v3 pivot and it settles the migration question for these
groups: they can be dropped with proven zero behavior change.

## OS13 — The instance key over-partitions when the template ignores the file

Source: [`e14-template-footprint.sh`](experiments/e14-template-footprint.sh)

The instance key is `{dir}/{base}`, derived from the **source** file. The
demanded path comes from the **template**. When the template does not use the
per-file variables, the two disagree and one file gets demanded many times.

Three changed sources, three templates:

| `related` template | instances | failures | distinct paths demanded |
| --- | --- | --- | --- |
| `{dir}/{name}_test.go` | 3 | 3 | 3 |
| `{dir}/README.md` | 3 | 3 | **1** |
| `CHANGELOG.md` | 3 | 3 | **1** |

Row 1 is right. Rows 2 and 3 report the same missing file three times.

Row 2 is the shipped `iac` preset. Row 3 — "any change under `src/` requires a
changelog entry" — is an obvious policy to want, and it produces one failure per
changed file.

The mismatch is mechanical: the key is derived from the source, but the
partition that matters is the one the template induces.

## OS12 — A report can say `PASS` and `Rules failed: 1` at once

Source: [`e13-severity-failon.sh`](experiments/e13-severity-failon.sh)

`summarize` counts every failing result in `Failed`, but only clears `OK` for
results at or above the `--fail-on` threshold
([`internal/rules/engine.go:32`](../../../internal/rules/engine.go)). With the
default `--fail-on error`, a failing `severity: warn` rule produces:

```text
kyn check

PASS

Rules failed: 1
Warnings: 1

[WARN] r
Status: fail
Message: Test not updated.
EXIT=0
```

The exit code is defensible — `warn` is not meant to block. The headline is not:
`PASS` sits directly above `Rules failed: 1`, and the two contradict.

This is not a corner case. The shipped `web-ui` preset's `tests-sync` rule is
`severity: warn`, so the most common starter config prints this whenever a spec
is not updated alongside its component.

Same family as OS3 and OS5: the tool reports success while a policy was
violated. Here nothing is hidden — the detail is right there — but the word a
reader scans for is wrong.

Fix belongs in v3's report design: reserve `PASS` for `Failed == 0`, and give
non-blocking failures their own headline so the summary word and the counts
agree.

## OS11 — The shipped `api` preset fails on an ordinary change set

Source: [`e12-preset-api-defect.sh`](experiments/e12-preset-api-defect.sh)

`kyn init --preset api` generates a family that sets

```yaml
    baseName:
      stripSuffixes: ["_handler", "_service"]
    kin:
      test: "{dir}/{name}_test.go"
```

These two lines contradict each other. `stripSuffixes` collapses
`order_handler.go` and `order_service.go` into the single instance
`internal/order/order` (OS4), while `{name}` resolves differently for each of
them. Changing a handler and a service in the same directory — both matched by
the preset's own globs — trips the kin-agreement guard:

```text
family resolution failed: family "go-api-handler" instance "internal/order/order":
kin "test" template "{dir}/{name}_test.go" resolves to different paths for
different source files in this instance
```

`check` exits 2. Changing either file alone works, so the defect only appears
once a plausible multi-file change arrives — likely in CI rather than locally.

Two separate problems:

1. **The preset is wrong.** For its evident intent — each handler and service
   has its own `_test.go` — the fix is to drop `stripSuffixes`, giving the
   two-instance behavior OS4 measured. Using `{base}` instead would be the other
   coherent reading: one shared `order_test.go`.
2. **The error's hint is wrong for this case.** It says the template "uses
   `{ext}`, `{file}`, or `{name}` and the instance's source files have different
   extensions". Here both files are `.go`; the collapse came from
   `stripSuffixes`. The guard fires correctly, but its explanation sends the user
   looking at extensions.

## OS9 — Result cardinality is rules × instances

Source: [`e10-shared-instance.sh`](experiments/e10-shared-instance.sh)

One family, two rules, three changed source files of which two collapse into
one instance:

```text
[web-component] src/a     sources: a.component.html, a.component.ts
[web-component] src/b     sources: b.component.ts
```

produces four results, one per (rule, instance), sorted by `familyId`, then
`familyName`, then `ruleId`:

| ruleId | familyId | familyName |
| --- | --- | --- |
| `spec-sync` | `web-component` | `src/a` |
| `story-sync` | `web-component` | `src/a` |
| `spec-sync` | `web-component` | `src/b` |
| `story-sync` | `web-component` | `src/b` |

Several source files resolving to one instance yield one result listing all of
them, not one result each.

## OS16 — Non-ASCII filenames never match any rule

Source: [`e17-nonascii-paths.sh`](experiments/e17-nonascii-paths.sh)

Git quotes unusual filenames in `--name-status` output unless told not to. Kyn
runs `git diff --name-status -M` without `-c core.quotePath=false` or `-z` and
does not unquote (`internal/changes/git.go:12`). For a change adding
`src/café.go` and `src/café_test.go`:

```text
git prints:   A	"src/caf\303\251.go"
kyn collects: "src/caf/303/251.go"
```

The octal escapes are normalized into path separators, so the file never matches
a glob and no rule sees it. The run reports `PASS` for a source file that has no
updated test. It fails open, with no error.

This is independent of v3. v3 must read git output with quoting disabled (or
`-z`), and `--files-from` must either do the same or reject quoted lines
([`02-semantics.md`](02-semantics.md) §6).

## OS14 — Informational rules, and their effect on the severity counters

Source: [`e15-emit-only.sh`](experiments/e15-emit-only.sh)

A v2 rule with `actions.emit` and no assertion is valid. It reports
`status: info`, never fails, and adds its flag. It is specified behavior
(`docs/spec.md:480`) and ships in `docs/site/recipes/frontend.md`.

Its declared `severity` still feeds the severity counters:

| `severity` | `ok` | `failed` | `infos` | `warnings` | `errors` | result status |
| --- | --- | --- | --- | --- | --- | --- |
| `info` | true | 0 | 1 | 0 | 0 | `info` |
| `warn` | true | 0 | 0 | 1 | 0 | `info` |
| `error` | true | 0 | 0 | 0 | 1 | `info` |

A rule that cannot fail reports `errors: 1`, and with `severity: error` a
result whose status is `info` is counted under `errors` while `infos` stays 0.
Same family as OS12: counters that do not mean what they appear to.

Flags are added whenever a rule applies, pass or fail
(`internal/rules/evaluator.go:47`), so "raise a flag only when the related file
is absent" needs an applicability gate; it cannot be written as an assertion.

## OS15 — YAML anchors pass through the strict loader

Source: [`e16-yaml-anchors.sh`](experiments/e16-yaml-anchors.sh)

A glob list declared once with `&anchor` and reused with `*alias` loads and
resolves correctly under `KnownFields(true)`. Sharing a list between
declarations needs no Kyn feature.

## OS10 — `rule.description` is parsed and never read

`Rule.Description` is declared
([`internal/config/config.go:33`](../../../internal/config/config.go)) and, via
`KnownFields(true)`, is an accepted field — a config using it loads
([`05-compatibility-matrix.md`](05-compatibility-matrix.md), row 13).
Nothing reads it. `grep -rn '\.Description' internal/report internal/rules
internal/cli` returns no hits.

SARIF has the natural home for it and fills both slots with the message instead
([`internal/report/sarif.go:86`](../../../internal/report/sarif.go)):

```go
ShortDescription: sarifMessage{Text: result.Message},
FullDescription:  sarifMessage{Text: result.Message},
```

That makes three constructs v2 accepts that do nothing:

1. non-`source` `groups` (OS6)
2. `if.changedAny: [source]` (OS6)
3. `rule.description` (this observation)

Two of the three are taught by the generated starter configs. This is the
strongest single argument for the v3 pivot, and it is also a small standalone
improvement: routing `Description` into SARIF `fullDescription` costs one line
and makes the field mean something.

## OS7 — Error selection is already non-deterministic (shipping defect)

Source: [`e6-determinism.sh`](experiments/e6-determinism.sh)

`AGENTS.md` core rule 2 and the repository steering both require deterministic
behavior. Two code paths select an error by ranging over a Go map and returning
the first hit, so *which* error the user sees varies between identical runs.

`checkKinAgreement` ranges `fam.Kin`
([`internal/family/resolver.go:117`](../../../internal/family/resolver.go)).
With three kin templates that all conflict, 30 identical invocations reported:

```text
  23 kin "alpha"
   5 kin "bravo"
   2 kin "charlie"
```

`Validate` ranges `fam.Kin` for template validation
([`internal/config/validate.go:89`](../../../internal/config/validate.go)). With
three invalid templates, 30 identical invocations reported:

```text
  22 {bogusA}
   5 {bogusC}
   3 {bogusB}
```

Exit codes are stable (`2` in both cases); only the message varies. The impact
is flaky golden fixtures, CI logs that differ run to run, and a user who fixes
one reported problem only to be shown a different one.

`Validate` ranges `fam.Groups` the same way
([`internal/config/validate.go:74`](../../../internal/config/validate.go)); the
same class applies.

This is a v0.1.x defect, fixable independently of v3 by sorting keys before
each validation and resolution pass. It also sets a hard constraint on v3: the
normalized policy model must be an **ordered slice sorted by ID**, never a Go
map, or every v3 rule and pattern inherits this defect.

## Where each observation is addressed

| Observation | Addressed in |
| --- | --- |
| OS1, OS2 — what the assertions actually read | [`02-semantics.md`](02-semantics.md) §1, §4 |
| OS3, OS8 — deletion and rename blindness | [`02-semantics.md`](02-semantics.md) §2, §3 |
| OS4, OS11 — `stripSuffixes` decides grouping; the `api` preset contradicts itself | Captures in `match`, [`04-grammar.md`](04-grammar.md); instance keys, [`03-instances-and-reports.md`](03-instances-and-reports.md) §1 |
| OS5 — the common form matches its own related file | Validation rule 9, [`04-grammar.md`](04-grammar.md) |
| OS6, OS10 — constructs that are accepted and do nothing | Removed, or made functional (`why`), [`04-grammar.md`](04-grammar.md) |
| OS7 — non-deterministic error selection | First item in [`06-cutover.md`](06-cutover.md) and [`07-conformance.md`](07-conformance.md) |
| OS9 — rules × instances | [`03-instances-and-reports.md`](03-instances-and-reports.md) §1 |
| OS12 — `PASS` above a failure count | [`03-instances-and-reports.md`](03-instances-and-reports.md) §6 |
| OS13 — one file demanded once per source | [`03-instances-and-reports.md`](03-instances-and-reports.md) §1 |
| OS14 — informational rules and counters | [`02-semantics.md`](02-semantics.md) §5, §8 |
| OS15 — YAML anchors work | [`04-grammar.md`](04-grammar.md), Sharing globs |
| OS16 — non-ASCII filenames never match | [`02-semantics.md`](02-semantics.md) §6; [`06-cutover.md`](06-cutover.md), change collection |

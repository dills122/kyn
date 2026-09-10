# Observed v2 semantics

Status: measured ground truth. Not a proposal.

Every statement here was produced by running the `kyn` binary built from this
branch against a generated fixture, not by reading source. Reproduction scripts
live in [`experiments/`](experiments/) and each observation cites the one that
produced it.

The purpose is to give the v3 design a factual baseline. Several proposal and
review claims turned out to be understated, and two shipping defects surfaced
that are independent of v3.

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

| Expectation (v3 name) | V2 clauses | exists | in change set | status | exit |
| --- | --- | --- | --- | --- | --- |
| `changed-if-present` | `if.kinExists` + `assert.kinChanged` | Y | Y | pass | 0 |
| | | Y | N | fail | 1 |
| | | N | Y | **skipped** | 0 |
| | | N | N | **skipped** | 0 |
| `exists-and-changed` | `assert.kinExists` + `assert.kinChanged` | Y | Y | pass | 0 |
| | | Y | N | fail | 1 |
| | | N | Y | fail | 1 |
| | | N | N | fail | 1 |
| `changed` | `assert.kinChanged` | Y | Y | pass | 0 |
| | | Y | N | fail | 1 |
| | | N | Y | **pass** | 0 |
| | | N | N | fail | 1 |
| `unchanged` | `assert.kinUnchanged` | Y | Y | fail | 1 |
| | | Y | N | pass | 0 |
| | | N | Y | fail | 1 |
| | | N | N | pass | 0 |
| `exists` | `assert.kinExists` | Y | Y | pass | 0 |
| | | Y | N | pass | 0 |
| | | N | Y | fail | 1 |
| | | N | N | fail | 1 |
| `missing` | `assert.kinMissing` | Y | Y | fail | 1 |
| | | Y | N | fail | 1 |
| | | N | Y | pass | 0 |
| | | N | N | pass | 0 |

## OS2 — Existence and change-set membership are orthogonal axes

OS1 reduces cleanly. Reading the table by which axis each expectation consults:

| Expectation | reads *exists* | reads *in change set* |
| --- | --- | --- |
| `changed` | no | yes |
| `unchanged` | no | yes (inverted) |
| `exists` | yes | no |
| `missing` | yes (inverted) | no |
| `changed-if-present` | yes (as a gate) | yes |
| `exists-and-changed` | yes (as an assertion) | yes |

Four of the six read exactly one axis. The two compound forms differ from each
other only in whether the existence check *gates* the rule (producing `skipped`)
or *asserts* it (producing `fail`). This is the whole of the vocabulary, and it
is a 2x2 with one axis optionally promoted to a gate — not an open-ended enum.

`exists-and-changed` and `changed` differ in exactly one cell: `exists=N,
in-change-set=Y`. That state is reachable when the selected diff and the working
tree disagree, for example `--base main --head origin/feature` evaluated from a
checkout that does not contain the added file. The distinction is real but
narrow, and any v3 documentation must say so explicitly or users will treat the
two as synonyms.

## OS3 — Kyn cannot observe deletion, and the default shape is silenced by it

Source: [`e4-git-delete.sh`](experiments/e4-git-delete.sh)

`docs/decisions.md` excludes `D` paths from the change set. The consequence is
larger than a vocabulary problem. In a real git repository where the source file
was modified and its related test was **deleted in the same commit**:

| Expectation | result | exit |
| --- | --- | --- |
| `changed-if-present` | **skipped** | 0 |
| `unchanged` | **pass** | 0 |
| `changed` | fail | 1 |

`explain` renders the `unchanged` case as:

```text
Assert:
  - assert.kinUnchanged: pass (rel unchanged (src/a_test.go))
```

The file does not exist. The trace states that it is unchanged.

`changed-if-present` is the shape every `kyn init` preset ships and the shape
the v3 proposal nominates as the common default. Deleting the related test
silences the rule that exists to protect it, and `check` exits 0.

No current rule form can express "the related file was deleted in this change",
because deleted paths never enter the change set. `assert.kinExists` catches the
after-effect — the path is gone from the working tree — but cannot attribute it
to this diff.

This is a capability gap, not a naming problem. Renaming `changed` to
`in-change-set` makes the vocabulary honest; it does not make deletion
observable. v3 must decide whether to close the gap or document it.

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

The v3 proposal places `stripSuffixes` in `patterns` only, classified under
"source-shape concerns". A v3 inline rule consequently has no way to express the
shipped `api` preset, and any migration that drops the pattern indirection
changes behavior.

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
`src/a_test_test.go` does not exist, so the `changed-if-present` gate skips it —
the OS3 silent skip conceals the OS5 mistake, and `check` exits 0 with `PASS`.

Both hazards are properties of the shape the v3 proposal nominates as the
default and describes as needing `exclude` only "unless needed". Measured
against real layouts, `exclude` is needed in the common case, and omitting it
fails silently rather than loudly.

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

## Consequences for the v3 design

| Observation | Consequence |
| --- | --- |
| OS1, OS2 | The expectation vocabulary is a small closed grid, not an open enum. The proposal's combinatorial-growth risk is smaller than feared. |
| OS3 | Deletion blindness must be an explicit product decision before the vocabulary is frozen. Naming alone cannot fix it. |
| OS4 | `stripSuffixes` must be available to the inline form, or the inline form must be declared single-shape-only and migration must refuse to flatten multi-shape families. |
| OS5 | v3 must either auto-exclude resolved related paths from source matching, or make an unmatched/self-matched source a loud error. |
| OS6 | Migration drops non-`source` groups with proven zero behavior change. `kyn init` should stop emitting them now, independently of v3. |
| OS7 | Normalized model is an ordered slice. Fix the existing defect first so v3 differential tests have a stable baseline. |

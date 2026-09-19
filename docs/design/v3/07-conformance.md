# G5 — Conformance suite

Status: proposed. Closes gate G5 in [`00-workplan.md`](00-workplan.md).

> Formerly a *differential* harness comparing a v2 policy against its migrated
> v3 equivalent. The
> [scope change](00-workplan.md#scope-change-2026-09-10) removed the migration,
> so this is now a straight conformance suite for v3 semantics. It is no longer
> blocked on a stable v2 baseline and can be built alongside the implementation.

## The oracle

There is no v2 side to compare against, so the oracle is the frozen semantics
themselves:

1. Build a fixture and a change set.
2. Run the v3 config.
3. Assert every observable against the truth table in
   [`02-semantics.md`](02-semantics.md) §6, computed independently of the
   implementation.

Structural comparison of normalized policies is **not** sufficient. Two policies
can normalize identically and still render different text, and the text is what
users and CI consume. Assert on rendered output.

The determinism guard below is what makes any of this meaningful: OS7 measured
the current binary producing three different error messages across thirty
identical runs, and that code is reused by v3.

## Observables

All of them, every run:

| Observable | Why it can differ independently |
| --- | --- |
| `check` exit code | severity mapping, `--fail-on`, gate changes |
| `check` text | ordering, `Shape:` / `Instance:` labels, pass suppression |
| `check` json | field values, result count |
| `check` sarif | rule-level dedup, `properties`, level mapping |
| `check` rdjson | instance identity is baked into `message` prose (IR1) |
| `check` checkstyle | groups by path, reuses the RDJSON message renderer |
| `explain` text | `When:` / `Expect:` section labels, clause traces |
| `explain` json | `clause` strings, `skipped` counts |
| `--dry-run-resolve` text and json | instance list, resolved related paths |
| result counts | `passed`, `failed`, `infos`, `skipped`, `warnings`, `errors` |
| emitted flags | ordering and membership |
| result ordering | `shapeId`, then `instanceName`, then `ruleId` (OS9) |

`--show-passes` must be on. Passing results are hidden by default in text
output, so a regression that flips a fail to a pass would otherwise show up only
as a missing line.

## Invariants that were sanctioned differences

The differential version of this document carried three sanctioned exceptions.
Two disappear with the migration; one becomes an ordinary assertion.

| Was | Now |
| --- | --- |
| S1 — a v3 rule fails where v2 skipped, on delete/rename-away | Not an exception. The gate is universal ([`02-semantics.md`](02-semantics.md) §4), so this is simply required behavior, asserted in layer 1. |
| S2 — `explain` uses v3 clause vocabulary | Not an exception. One vocabulary exists. |
| S3 — unreferenced kin vanish from the resolve report | Gone with `kin`. |

What replaces them is a positive assertion set. The delete and rename-away
scenarios from [`e9-gate-prototype.sh`](experiments/e9-gate-prototype.sh) must
**fail**, and the never-existed scenario must **skip**. A suite that only
asserted "no crash" would pass while the OS3 hole was reopened.

## Scenario matrix

The full cross product is not worth running. Two layers instead.

### Layer 1 — semantics core, exhaustive

The dimensions that interact:

| Dimension | Values |
| --- | --- |
| `when` | `always`, `related-existed`, `related-absent` |
| `expect` | `in-change-set`, `not-in-change-set`, `exists`, `missing`, `[exists, in-change-set]` |
| `relatedState` | `present`, `vanished` (deleted), `vanished` (renamed away), `absent` |
| related in change set | yes, no |

An earlier draft counted 3 × 5 × 2 × 2 × 3 = 180 cells by treating existence and
vanishing as independent axes. They are not — review 2's point about reachable
cells. The tri-state in [`02-semantics.md`](02-semantics.md) §3 collapses them,
and the remaining combinations are further constrained:

- a `vanished` or `absent` path cannot be in the change set, since deleted paths
  never enter it and a rename records only the destination
- so `in change set` is free only for `present`

That leaves 3 × 5 × (2 `present` rows + 1 each for deleted, renamed, absent) =
**3 × 5 × 5 = 75 reachable cells**, asserted on `explain --format json` plus the
exit code. This is the direct successor to
[`e3-truthtable.sh`](experiments/e3-truthtable.sh) and should replace it as a
committed Go test — the experiment script produced the v2 baseline, and this
produces the v3 contract.

### Layer 2 — projection and shape, sampled

Applied to a sampled subset of layer 1, but across every output mode:

| Dimension | Values |
| --- | --- |
| input mode | `--files`, `--files-from` 1-column, `--files-from` 2-column (D9), git |
| source files per instance | one, several (multi-extension) |
| rules per source shape | one (inline), several (pattern) |
| `on` filter | absent, matching, non-matching, mixed statuses within one instance |
| severity | `info`, `warn`, `error` |
| output mode | all seven |

Layer 2 is where IR2's grouping concerns and IR1's projection concerns actually
get exercised.

## Corpus

Real configs, not only synthetic ones:

- the four rewritten v3 presets, including the corrected `api` ([`06-cutover.md`](06-cutover.md))
- the configs embedded in `docs/site/recipes/*.md`, rewritten for v3
- one config per validation rule in [`04-grammar.md`](04-grammar.md), asserted to
  be **rejected** with the right message and exit 2

That last item matters as much as the passing cases. Validation rules 7 and 10 —
unused pattern, self-match — are new and unproven, and a suite that only runs
valid configs would never exercise them.

## Form

Per `AGENTS.md`: table-driven Go tests, small explicit fixtures, golden files
alongside the existing `internal/report` goldens.

The existing v1→v2 end-to-end test (`e2e/workflows_test.go:95`) is the right
*shape* to copy even though its subject is being deleted: build a fixture, run
the real binary, assert on rendered output and exit code.

Golden files are external-contract fixtures under the repository steering, so a
diff in one is a deliberate review event, not a test to re-record. Record them
only after the `iac` deduplication question in [`06-cutover.md`](06-cutover.md)
is settled, since that changes result counts.

## Exit criteria for G5

1. Layer 1 green across all 75 reachable cells, with the delete and
   rename-away scenarios asserted to fail under `related-existed`, and
   never-existed asserted to skip under it and fire under `related-absent`.
2. Layer 2 green across all seven output modes.
3. Every invalid corpus config rejected with the expected message and exit 2.
4. The suite passes repeatedly — a determinism guard that runs the same input
   several times and requires identical bytes, closing OS7 permanently. This is
   the one criterion that must be written first, because it is the only one that
   can fail intermittently.

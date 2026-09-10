# G5 — Differential conformance harness

Status: proposed. Closes gate G5 in [`00-workplan.md`](00-workplan.md).

The harness that proves a migrated policy behaves identically — and proves the
sanctioned exceptions are the *only* differences.

## Precondition: G0 first

A differential harness compares outputs across two runs. OS7 measured the
current binary producing three different error messages across thirty identical
runs. Until G0 lands, a red result cannot be distinguished from a coin flip.

**Do not build this harness before the determinism fix.**

## The oracle

For each config in the corpus:

1. Migrate v2 → v3 (G4).
2. Run both configs against the same fixture and the same change set.
3. Compare every observable.
4. Assert the difference set is empty, or is exactly a sanctioned exception.

Structural comparison of normalized policies is **not** sufficient. Two policies
can normalize identically and still render different text, and the text is what
users and CI consume. Compare rendered output.

## Observables

All of them, every run:

| Observable | Why it can differ independently |
| --- | --- |
| `check` exit code | severity mapping, `--fail-on`, gate changes |
| `check` text | ordering, `Family:` / `Instance:` labels, pass suppression |
| `check` json | field values, result count |
| `check` sarif | rule-level dedup, `properties`, level mapping |
| `check` rdjson | family identity is baked into `message` prose (IR1) |
| `check` checkstyle | groups by path, reuses the RDJSON message renderer |
| `explain` text | `If:` / `Assert:` section labels, clause traces |
| `explain` json | `clause` strings, `skipped` counts |
| `--dry-run-resolve` text and json | instance list, kin map |
| result counts | `passed`, `failed`, `infos`, `skipped`, `warnings`, `errors` |
| emitted flags | ordering and membership |
| result ordering | `familyId`, then `familyName`, then `ruleId` (OS9) |

`--show-passes` must be on. Passing results are hidden by default in text
output, so a regression that flips a fail to a pass would otherwise show up only
as a missing line.

## Sanctioned differences

A naive harness flags all three of these as regressions. They are the design,
and each must be asserted **positively** — the harness should fail if the
difference is *absent*, not merely tolerate it when present.

| # | Difference | Where | Authority |
| --- | --- | --- | --- |
| S1 | A v3 rule fails where v2 skipped, when the related path was deleted or renamed away in the change under evaluation | exit code, status, counts | D5, [`02-semantics.md`](02-semantics.md) §4 |
| S2 | `explain` clause strings and text section labels use v3 vocabulary (`when.related-existed`, `expect.in-change-set`, `When:` / `Expect:`) | `explain` only | D13, [`03-instance-and-identity.md`](03-instance-and-identity.md) §7 |
| S3 | Kin unreferenced by any rule disappear from the resolve report | `--dry-run-resolve` only | [`06-migration.md`](06-migration.md) |

Everything else is a regression. In particular `familyId`, `familyName`, the
`(family instance: …)` message suffix, and result ordering must be **byte
identical** (D10).

S1 is the one with teeth: [`e9-gate-prototype.sh`](experiments/e9-gate-prototype.sh)
established it changes exactly 2 of 5 scenarios. The harness must confirm both
that those two change and that the other three do not.

## Scenario matrix

The full cross product is not worth running. Two layers instead.

### Layer 1 — semantics core, exhaustive

The dimensions that interact:

| Dimension | Values |
| --- | --- |
| `when` | `always`, `related-existed`, `related-absent` |
| `expect` | `in-change-set`, `not-in-change-set`, `exists`, `missing`, `[exists, in-change-set]` |
| related exists | yes, no |
| related in change set | yes, no |
| related vanished | no, deleted, renamed away |

3 × 5 × 2 × 2 × 3 = 180 cells, minus the unreachable combinations (a path cannot
both exist and have vanished in the same change). Run in git mode, compared on
`explain --format json` plus the exit code. This is the direct successor to
[`e3-truthtable.sh`](experiments/e3-truthtable.sh) and should replace it as a
committed Go test.

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

- the four `kyn init` presets — post-G0 fix for `api` (OS11)
- the configs embedded in `docs/site/recipes/*.md`
- the existing `internal/config`, `internal/family` and `internal/rules` fixtures
- one config per refusal class R1–R3, asserted to be **refused** rather than
  migrated

That last item matters as much as the passing cases. A migration that silently
approximates R1 would pass every equivalence test that only runs migratable
configs.

## Form

Per `AGENTS.md`: table-driven Go tests, small explicit fixtures, golden files
alongside the existing `internal/report` goldens.

The existing v1→v2 end-to-end test (`e2e/workflows_test.go:95`) is the right
shape and the right standard — it covers one scenario, and this expands it to a
matrix.

Golden files are external-contract fixtures under the repository steering, so a
diff in one is a deliberate review event, not a test to re-record.

## Exit criteria for G5

1. Layer 1 green, with S1 asserted positively in both directions.
2. Layer 2 green across all seven output modes.
3. Every corpus config either migrates with only sanctioned differences, or is
   refused with a cited R-number.
4. The suite passes repeatedly — a determinism guard that runs the same input
   several times and requires identical bytes, closing OS7 permanently.

# v3 configuration: workplan and decision log

Status: design only. Nothing is implemented and implementation is not
authorized.

This document is the history. The other documents in this directory describe
the current design and nothing else.

## Where things stand

| Area | State | Document |
| --- | --- | --- |
| Measured v2 baseline | Done; 15 observations, each with a script | [`01`](01-observed-semantics.md) |
| Semantics | Settled | [`02`](02-semantics.md) |
| Instances and reports | Settled | [`03`](03-instances-and-reports.md) |
| Grammar and wording | Revised 2026-11-13; **not yet independently reviewed** | [`04`](04-grammar.md) |
| Cutover inventory | Measured | [`06`](06-cutover.md) |
| Conformance suite | Specified | [`07`](07-conformance.md) |
| v1/v2 compatibility matrix | Historical evidence only | [`05`](05-compatibility-matrix.md) |

Next gate: an independent review of the revised shape. Three reviews have been
run, but the 2026-11-13 revision changed the grammar after the third, so the
capture syntax, the removal of `patterns`, and the new wording have been tested
only by the author's own scenarios.

## Two changes of direction

**2026-09-10 — no users.** Kyn has no real users yet. A bad bug is fixed in
place; anything breaking goes into v3, and a full cutover is acceptable. This
removed every decision that existed only to protect compatibility: version
gating, migration, and keeping the old report field names.

**2026-11-13 — developer delight.** Three review rounds had made the design
correct and the config unpleasant: to write the common rule an author needed
`when: related-existed`, `expect: in-change-set`, `patterns`, `use`,
`stripSuffixes`, and the difference between `{base}` and `{name}`, backed by 15
validation rules. The shape was revised against one test: can a developer read
the rule aloud. The four presets went from 39/27/22/22 non-blank lines in v2 to
13/7/7/8. The semantics from the earlier rounds were kept; the surface changed.

## Decisions in force

| ID | Decision | Basis |
| --- | --- | --- |
| D1 | The normalized model is ordered slices sorted by ID, never Go maps | OS7 |
| D3 | `rules` is a YAML mapping keyed by ID | `yaml.v3` already rejects duplicate keys with line numbers |
| D5 | The deletion gap is closed, for every config | OS3, OS8; the data is already in the git output Kyn runs |
| D7 | Generated messages are a versioned, fixture-tested contract containing only the rule ID and resolved related path | Absolute paths would leak the CI checkout root |
| D8 | A related path that matches its own rule's `match` is a config error, detected at resolve time, exit 2 | OS5; a static check is glob intersection |
| D15 | v3.0 has one `related` per rule; atomic multi-path rules are out | No config uses one |
| D17 | A rule's instance key is the values of the variables its `related` template uses | OS13; also makes OS11 impossible |
| D18 | Headlines are `FAIL`, `NON-BLOCKING`, `PASS`; `PASS` means no failures | OS12 |
| D19 | `docs/decisions.md` and `docs/migration-v1-to-v2.md` are bannered as superseded, not deleted | This set cites them |
| D20 | A JSON Schema ships once the grammar freezes, generated from or CI-checked against the loader | A drifting schema approves configs the binary rejects |
| D23 | A related path is in one of three states: `exists`, `removed` (by this change), `missing` | Review 2: the earlier boolean was true for a file added in the change |
| D24 | `--files-from` accepts `git diff --name-status -M` lines verbatim, rename lines included | Review 2: two columns cannot hold a rename |
| D25 | Experiment scripts never `cd`; every fixture git call is scoped and checked; `set -e` is not used | Review 2: a failed fixture could commit to the invoking repository |
| D26 | Informational rules are written `expect: none`; `severity` must be `info`; at least one of `message` or `flags` | Review 2; OS14 |
| D27 | `match` binds variables: `{name}` within a segment, `{**name}` across segments; `{dir}` is the only built-in. `stripSuffixes`, `{base}`, `{name}`, `{ext}`, `{file}` are gone | 2026-11-13. Names the variable where it is matched; a 50-line prototype resolved every preset shape; `{**name}` adds mirrored trees, which v2 could not express |
| D28 | No `patterns` block and no `use`. Rules repeat their `match` | 2026-11-13. Patterns gave only de-duplication and a grouping label; every v2 capability maps without them; YAML anchors already work (OS15); it removed 9 of 15 validation rules |
| D29 | Wording: `expect: touched` / `untouched`; `skipIf: missing` / `exists` in place of `when`; `changeType` in place of `on`; `flags` in place of `emit`; `why` in place of `description` | 2026-11-13. `skipIf` stays three-way because flags are added whenever a rule applies (OS14), so "flag only when absent" needs a gate |
| D30 | `related` is optional when `expect: none`; `skipIf` requires `related` | 2026-11-13. Scenario 11: a flag-only rule had to invent a related file, which then matched its own glob |
| D31 | Instances belong to a rule. `instanceName` is the variable values joined with `/` in the order they appear in `related`; `(all)` with no variables; `(root)` for the repository root. `related` may use at most one variable that can contain `/` | 2026-11-13, review 3. Keeps names unique within a rule |
| D32 | `check` reports skipped results: a count in text and JSON always, entries in JSON, entries in text with `--show-passes` | Review 3: skips were invisible outside `explain` |
| D33 | Capability accounting names what was found and does not claim completeness. Two limits (atomic multi-path; gating on one path while asserting on another), one gain (mirrored trees) | Review 3: the list had been wrong twice |
| D34 | The conformance suite extends `e2e/e2e_test.go` and `e2e/projects`; it does not build a second harness | Review 3 |
| D35 | Reports drop `shapeId`/`familyId`; results carry `ruleId` and `instanceName` and sort by them. `explain` clauses use the field names (`skipIf.missing`, `expect.touched`) | 2026-11-13, following D28 |
| D36 | `expect` stays required; there is no default assertion | 2026-11-13. Explicit over implicit |

D29's `why` and `changeType`, and D36, were confirmed by the maintainer on
2026-11-13.

## Superseded decisions

| ID | Was | Replaced by |
| --- | --- | --- |
| D2 | Migration drops inert groups | No migration |
| D4, D6 | Vocabulary `in-change-set` etc. with a `when` gate | D29. The two-field split and "no lenient default" survive |
| D9 | `--files-from` two-column form | D24 |
| D10 | Keep `familyId`/`familyName` on the wire | D16, then D35 |
| D11 | Instances keyed on a shared source shape | D31 |
| D12 | `stripSuffixes` legal inline; a pattern is a named shape | D27, D28 |
| D13 | Clause vocabulary follows the config version | D35; one version exists |
| D14 | `description` survives; version stays `3` | D29 renames it `why`; the version number stands |
| D16 | Rename to `shapeId` / `instanceName` | D35 |
| D21 | Rule and pattern IDs share one namespace | D28; there are no patterns |
| D22 | Uniform footprint per pattern; `stripSuffixes` must be used | D27, D28, D31 |

## Reviews

### Independent review 1

On the original proposal. Recorded in
[`../../reviews/v3-config-independent-review-1.md`](../../reviews/v3-config-independent-review-1.md).

| Finding | Disposition |
| --- | --- |
| Normalized model cannot preserve report identity | D35 |
| Flat rules omit instance grouping semantics | D17, D31 |
| Migration claim broader than representable | No migration; capability limits in D33 |
| Expectation names overstate what Kyn observes | D29 |
| No compatibility matrix | Measured in `05`; no longer a constraint |
| Map-key validation not deterministic | D1, D3 |

### Independent review 2 (2026-09-19)

A comment on [PR #50](https://github.com/dills122/kyn/pull/50).

| Finding | Disposition |
| --- | --- |
| Rule and pattern IDs collide | Moot under D28 |
| `instanceName` undefined for shared patterns with constant paths | D31 |
| The existence predicate was misnamed; second gate unspecified | D23; `02` §3 |
| Informational rules lost | D26, D30 |
| Two columns cannot encode a rename | D24 |
| Experiment scripts hide fixture failures | D25 |

### Independent review 3 (2026-09-19)

A fresh-context review of commit `8efbfb2`. Verdict: ready to advance,
conditional on five P1 findings. It ran all thirteen scripts then present and
confirmed every table in `01` reproduces. Its findings were mostly decisions
that had not propagated to the cutover and conformance documents, which is why
those documents now describe only the current state.

| Sev | Finding | Disposition |
| --- | --- | --- |
| P1 | Reachable-cell count wrong; the excluded cell is the one separating `touched` from `[exists, touched]` | `02` §4; `07` layer 1 is 108 cells including it |
| P1 | Conformance document not updated for review 2 | `07` rewritten |
| P1 | Gating on one path while asserting on another is a second capability loss | D33; `04` |
| P1 | `check` cannot represent a skipped result | D32 |
| P1 | Cutover inventory incomplete; `e2e/projects` overlaps the proposed harness | `06` re-measured; D34 |
| P2 | OS9 cited but never written | Written into `01` |
| P2 | `instanceName` empty components; root-level files | D31 |
| P2 | Gate prototype still tested the old boolean | `e9` updated; both gates shown |
| P2 | Informational-rule claim had no script | `e15`, OS14 |
| P2 | Plain-path degradation specified for one gate only | `02` §6 |
| P2 | Deduplication decided, deferred and untracked at once | D17 stands; contradictions removed by the rewrite |
| P2 | Superseded decisions cited as delivered | This document separates in-force from superseded |
| P2 | No dimension exercised the headline | `07` layer 2 crosses severity with `--fail-on` |

The review also cleared three things it was asked to attack: the instance-key
model is sound, report identity was complete, and findings from reviews 1 and 2
had not been accepted too readily.

## Open items

1. An independent review of the 2026-11-13 revision. Scope is frozen in
   [`../../reviews/v3-config-review-4-bootstrap.md`](../../reviews/v3-config-review-4-bootstrap.md).
2. The `api` preset must choose between one shared test per handler/service
   pair and one test per file ([`06`](06-cutover.md)).
3. `docs/site/getting-started.md` teaches v2 (PR #49) and will need a v3
   rewrite at cutover.

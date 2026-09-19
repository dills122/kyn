# v3 configuration: design workplan

Status: active. Design only — no implementation is authorized.

One question is open (item 12, informational rules). Independent review 2 is
recorded below; the PR body's earlier claim that no questions remained is
superseded by it.

This tracks the findings raised against
[`../../reviews/v3-config-proposal.md`](../../reviews/v3-config-proposal.md) by
[independent review 1](../../reviews/v3-config-independent-review-1.md) and by a
follow-up review, plus findings produced by the experiments recorded in
[`01-observed-semantics.md`](01-observed-semantics.md).

## Scope change (2026-09-10)

Maintainer input: **Kyn has no real users yet.** The working rule is now — a bad
bug gets fixed in place; anything that would be a breaking change gets lifted
into v3 instead, and a full cutover to v3 is acceptable.

This invalidates two decisions that rested on compatibility alone:

| Decision | Original rationale | Now |
| --- | --- | --- |
| **D5** version-gated the deletion-gate fix | preserve a shipped contract | Gate is universal. The "first and only intentional divergence" disappears, and with it the S1 exception in G5. |
| **D10** keep `familyId` / `familyName` on the wire | renaming breaks consumers at migration | No consumers. Renamed to `shapeId` / `instanceName` (D16), including the `(family instance: …)` prose in RDJSON and Checkstyle. |

Two more go moot, both presupposing several live config versions: **D2**
(migration drops inert groups) and **D13** (clause vocabulary follows config
version).

Ten decisions are unaffected — D1, D3, D4, D6, D7, D8, D9, D11, D12, D14 — none
rested on compatibility. D11 is worth re-justifying on merit: shape-keyed
instances are not a compatibility artifact, they are what lets two rules share
one pattern's instances.

Consequences for the plan:

- **G0 mostly dissolves.** Only OS7 stays urgent, because the determinism defect
  lives in `resolver.go` and `validate.go`, code v3 reuses. The rest folds into
  building v3. The reject-vs-warn question about non-`source` groups is moot —
  `groups` does not exist in v3.
- **G3's compatibility matrix** becomes measured evidence rather than a
  constraint. Three findings survive; see that document.
- **G4 becomes a cutover**, not a migration. IR3 — a P1 — largely dissolves.
- **G5 loses its differential framing** and is no longer blocked on a stable v2
  baseline.

## Finding register

`IR` = independent review 1. `CR` = follow-up review. `OS` = surfaced by
experiment; see [`01-observed-semantics.md`](01-observed-semantics.md).

| ID | Finding | Sev | Gate | Status |
| --- | --- | --- | --- | --- |
| IR1 | Normalized model cannot preserve report identity | P1 | G2 | Settled — D16 (was D10) |
| IR2 | Flat rules omit source-instance grouping semantics | P1 | G2 | Settled — D11 |
| IR3 | Migration claim broader than the representable subset | P1 | G4 | Largely dissolved by the scope change; residue in [`06-cutover.md`](06-cutover.md) |
| IR4 | Expectation names overstate what Kyn observes | P2 | G1 | Settled — D6 |
| IR5 | Version-specific decoding lacks a compatibility matrix | P2 | G3 | Measured; superseded as a constraint by the scope change |
| IR6 | Map-key validation not fully deterministic | P2 | G0/G2 | Settled — D3, D1 |
| CR1 | `stripSuffixes` is a grouping input the inline form cannot express | P1 | G2 | Settled — D12 |
| CR2 | Migration has no answer for declared-but-inert v2 groups | P1 | G4 | Moot — no migration; `groups` deleted |
| CR3 | Resolve-time error selection becomes order-dependent under maps | P2 | G0 | Confirmed by OS7 |
| CR4 | Generated default messages are an unowned output contract | P2 | G1 | Settled — D7 |
| CR5 | `explain` clause names are a JSON contract the shorthand must map to | P2 | G2 | Settled — one vocabulary, no versioning needed |
| CR6 | "v3 config" collides with product versioning; `description` dropped | nit | G3 | Settled — D14 |
| OS3 | Kyn cannot observe deletion; the default shape is silenced by it | **P0** | G1 | Settled — D5 |
| OS8 | Renaming the related file away is equally invisible | **P0** | G1 | Settled — D5 |
| OS5 | The proposed common form self-matches; OS3 hides the mistake | P1 | G2 | Settled — D8 |
| OS7 | Error selection is already non-deterministic (ships today) | P1 | G0 | Specified — v0.1.x patch, not yet written |
| OS10 | `rule.description` is parsed and never read | P2 | G0/G3 | Settled — D14 |
| OS11 | The shipped `api` preset fails with exit 2 on an ordinary change set | P1 | G4 | Settled — corrected in the preset rewrite |
| OS12 | A report can print `PASS` above `Rules failed: 1` | P1 | G2 | Settled — D18 |
| OS13 | The instance key over-partitions when the template ignores the file | P1 | G2 | Settled — D17 |

### IR1 is wider than reported

Independent review 1 cited `familyId` / `familyName` in check, explain, SARIF
and dry-run. Measured, family identity reaches **every** output mode, because
`renderRDJSONMessage` appends it to the message string
([`internal/report/rdjson.go:136`](../../../internal/report/rdjson.go)) and
Checkstyle reuses that renderer
([`internal/report/checkstyle.go:83`](../../../internal/report/checkstyle.go)):

```json
"message": "Include the test. (family instance: src/a)"
```

```xml
<error ... message="Include the test. (family instance: src/a) Expected: src/a_test.go" source="kyn.r"/>
```

This is not a structured field a consumer can ignore. It is prose that lands in
reviewdog pull-request comments and CI annotations. A v3 rule has no family, so
every one of those strings changes unless the normalized model supplies a
substitute identity. Treat IR1 as the widest blocker, not a JSON-only concern.

### OS3 is raised to P0

Deleting the related file makes `changed-if-present` report `skipped` and exit
`0`, and makes `unchanged` print `pass (rel unchanged (src/a_test.go))` for a
file that no longer exists. `changed-if-present` is the shape every `kyn init`
preset ships and the shape the proposal nominates as the v3 default.

A policy tool whose default rule is disabled by deleting the file it protects
has a correctness problem, not a vocabulary problem. IR4 remains valid but is
now downstream of this: renaming `changed` to `in-change-set` makes the words
honest without making deletion observable.

## Independent review 2 (2026-09-19)

Raised on [PR #50](https://github.com/dills122/kyn/pull/50). All five design
points and the test-hygiene point were verified and accepted; two of them found
things the design set had asserted wrongly rather than merely left vague.

| ID | Finding | Status |
| --- | --- | --- |
| R2-1 | Rule and pattern IDs collide in one shape-ID space | Settled — D21 |
| R2-2 | `instanceName` undefined when rules sharing a pattern have different constant `related` paths | Settled — D22 |
| R2-3 | `existsNow OR vanishedInChange` is not existence at base; `related-absent` never specified | Settled — D23 |
| R2-4 | v2 emit-only informational rules are a second capability loss | **Open** — see open questions |
| R2-5 | `status<TAB>path` cannot encode a git rename | Settled — D24 |
| R2-6 | Git-backed experiment scripts hide fixture failures | Settled — D25 |

Two corrections to claims this design set had made:

- **R2-3 was a naming error, not an omission.** `existedAtBase` was computed as
  `existsNow OR vanishedInChange`, which is true for a file *added* in the
  change under evaluation — a file that never existed at base. Replaced by an
  explicit tri-state.
- **R2-4 falsified "atomic multi-path rules are the only loss."** Emit-only
  informational rules are specified behavior (`docs/spec.md:480`), produce
  `status: info` and a `flags` array, and ship in a recipe
  (`docs/site/recipes/frontend.md:61`). The grammar made `expect` required, so
  they were unexpressible and the loss went unrecorded.

## Gates

Ordered. A gate may not open until the one before it is closed.

### G0 — Stabilize the baseline — MOSTLY DISSOLVED

The scope change removed most of this gate. What was a v2 patch is now either
moot or folded into building v3.

| Item | Fate |
| --- | --- |
| OS7 — non-deterministic error selection | **Still urgent.** Lives in `resolver.go` and `validate.go`, which v3 reuses. Fix before the G5 determinism guard is written. |
| Zero-value error messages (matrix rows 15, 17) | Carries into v3 if the code is reused. Fix during the v3 build. |
| OS10 — `description` never read | Fix during the v3 build; the field survives (D14). |
| OS11 — broken `api` preset | Folded into the preset rewrite ([`06-cutover.md`](06-cutover.md)). Do not fix and then rewrite. |
| Kin-agreement hint text blames extensions | Carries into v3 — the same guard applies to `related`. Fix during the v3 build. |
| OS6 — `kyn init` emits inert groups | Moot; `groups` is deleted. |
| Reject vs warn on non-`source` groups | Moot; `groups` is deleted. |

### G1 — Freeze semantics — CLOSED

Delivered in [`02-semantics.md`](02-semantics.md).

- OS3 + OS8 closed by D5.
- Expectation grid frozen by D6.
- Generated messages constrained by D7.
- **Still open**: self-match policy for OS5 — auto-exclude resolved related
  paths, error, or warn. D5 makes the phantom instance's skip legible but does
  not stop it being created. Carried into G2, where the instance key is defined.

### G2 — Instance and identity model — CLOSED

Delivered in [`03-instance-and-identity.md`](03-instance-and-identity.md).

- Instance key rebased onto the source shape (D11), which is what preserves the
  measured cardinality and ordering contract (OS9).
- `stripSuffixes` moved to the rule (D12), so the inline form can express the
  shipped `api` preset.
- Report identity resolved by keeping the wire format unchanged (D10).
- `explain` clause vocabulary follows the config version (D13).
- Self-match rejected at resolve time (D8).

### G3 — Grammar and compatibility matrix — CLOSED

Delivered in [`04-grammar.md`](04-grammar.md) and
[`05-compatibility-matrix.md`](05-compatibility-matrix.md).

- Compatibility matrix measured, not guessed: 20 shapes probed, 11 load, 9
  rejected, each marked contractual or incidental.
- Grammar written against the frozen semantics and instance model.
- `description` survives and becomes functional (D14).
- All ten of the proposal's open questions answered.

### G4 — Cutover — CLOSED

Delivered in [`06-cutover.md`](06-cutover.md), which replaced `06-migration.md`.

- Removal list for v1/v2 structs, `MigrateV1ToV2`, `kyn config migrate`,
  `internal/family`, and the v1→v2 end-to-end test.
- Four presets rewritten by hand, with `api` corrected rather than transcribed.
- One capability limit accepted on evidence: atomic multi-path rules, used by no
  config anywhere in the repository.

### G5 — Conformance suite — SPECIFIED

Delivered in [`07-conformance.md`](07-conformance.md). No longer differential and
no longer blocked on a stable v2 baseline; it can be built alongside the
implementation.

- Two-layer matrix — exhaustive on semantics, sampled across output modes.
- The three former sanctioned differences collapse into ordinary assertions.
- Corpus includes one config per validation rule, asserted rejected — rules 7 and
  10 are new and otherwise unexercised.
- The determinism guard must be written first; it is the only criterion that can
  fail intermittently.

### G6 — Implementation

Not authorized. Requires maintainer approval of the G1–G4 design set as a
specification. With the cutover replacing migration, G5 is now built *alongside*
the implementation rather than before it — but the OS7 fix and the determinism
guard still come first.

## Decisions taken

Recorded here as they close; each is backed by an experiment or a code citation.

| ID | Decision | Basis |
| --- | --- | --- |
| D1 | The normalized policy model is an **ordered slice sorted by ID**, never a Go map. | OS7 shows map iteration already produces non-deterministic error selection. A v3 keyed by mapping would inherit it per rule and per pattern. |
| D2 | Migration drops non-`source` groups, as a documented no-op. | OS6 proves output is byte-identical without them across all six modes. |
| D3 | `rules` as a YAML mapping keyed by ID is safe; duplicate detection is free. | `yaml.v3` with `KnownFields(true)` already rejects duplicate keys with line numbers: `line 11: mapping key "test" already defined at line 10`, exit 2. Answers proposal open question 2 and closes half of IR6. |
| D4 | The expectation vocabulary is a closed 2x2 grid (existence × change-set membership) with existence optionally promoted to a gate — not an open enum. | OS2. Reduces the author's stated combinatorial-growth risk. |
| D5 | Close the deletion gap. The applicability gate becomes "existed at base", derived from `D` and rename-source entries Kyn already parses and discards. ~~Version-gated to v3~~ — universal, per the scope change. | Maintainer decision 2026-09-09. Validated by `e9-gate-prototype.sh`: fixes OS3 and OS8, changes 2 of 5 scenarios, adds no false positives, needs no new git call. See [`02-semantics.md`](02-semantics.md) §4. |
| D6 | The expectation vocabulary is two orthogonal fields — `when` and `expect` — not fused enum names. `when` defaults to `always`, never to a lenient gate. | OS1/OS2 plus IR4. Removes the concealed skip, and makes enum growth additive instead of multiplicative. See [`02-semantics.md`](02-semantics.md) §6. |
| D7 | `message` may be generated, but the generated text is a versioned, fixture-tested contract containing only the rule ID and resolved related path — never a config path or absolute path. | CR4. Absolute paths would leak the CI checkout root into reviewdog comments. |
| D8 | A resolved `related` path that matches its own rule's `match`/`exclude` is a config error, reported at resolve time with exit 2. | Maintainer decision 2026-09-09. A sound static check is glob intersection; the resolve-time check is exact and cheap. See [`03-instance-and-identity.md`](03-instance-and-identity.md) §5. |
| D9 | `--files-from` gains an optional `status<TAB>path` form so explicit mode can express deletion. Single-column lines still mean `modified`. | Maintainer decision 2026-09-09. Without it a policy designed with `--files` passes locally and fails in git-mode CI. See [`02-semantics.md`](02-semantics.md) §5. |
| ~~D10~~ | ~~Report wire format is unchanged across config versions.~~ **Superseded by D16.** | Rested entirely on not breaking consumers. The scope change removed that constraint. |
| D11 | Instances are keyed on the **source shape** — the (`match`, `exclude`, `stripSuffixes`) triple — not on the rule. A pattern is a named source shape; an inline rule owns an anonymous one identified by its rule ID. | OS9 shows rule-keyed instances would regroup and reorder every report. Shape-keying maps a v2 family one-for-one and preserves the contract. |
| D12 | `match`, `exclude` and `stripSuffixes` are all legal inline on a rule. A pattern is a named bundle of exactly those three and carries no extra capability. | OS4/CR1. Confining `stripSuffixes` to patterns left the inline form unable to express the shipped `api` preset. Also makes inline and pattern-backed forms normalize identically by construction. |
| D13 | Structural identity fields stay stable across config versions; vocabulary that echoes the user's own config (`explain`'s `clause`) follows the config version. | CR5. Reporting `if.kinExists` for a config containing neither `if` nor `kinExists` would be the actual defect. |
| D14 | `description` survives into v3 and becomes functional (SARIF `fullDescription`, `explain`). `version:` is a config-schema version; prose calls v3 "the rule-centric format", and the number stays `3` rather than restarting. | OS10 + CR6. Kept on merit rather than compatibility: a policy file benefits from recording *why* a rule exists, separately from the failure message, and SARIF has the slot. |
| D15 | v3.0 ships **without** atomic multi-path rules. `expect` stays list-shaped so a future list-shaped `related` is additive. | Maintainer decision 2026-09-10. Measured: every kin clause in the four presets, `docs/site/recipes/*` and `docs/site/config.md` names exactly one kin. The only multi-path syntax is `kinChangedAny` in an explicitly unapproved exploration doc. |
| D16 | Reports rename `familyId` → `shapeId` and `familyName` → `instanceName`; the RDJSON/Checkstyle message suffix becomes `(instance: src/a)`; text output says `Shape:`. Supersedes D10. | Maintainer decision 2026-09-10. With no consumers the fields get honest names. Sort structure is unchanged, so the OS9 ordering contract survives. |
| D17 | A rule's instance key is the source shape plus **exactly the variables its `related` template uses**. Refines D11. | Maintainer decision 2026-09-10 (dedupe). OS13 measured the over-partition: `{dir}/README.md` and `CHANGELOG.md` each demand one path but produce one failure per source file. Footprint keying deduplicates structurally, makes the template-agreement error unreachable, and makes OS11 impossible. See [`03-instance-and-identity.md`](03-instance-and-identity.md) §4. |
| D18 | Three report headlines: `FAIL` when the run blocked, `NON-BLOCKING` when failures exist but did not block, `PASS` only when `Failed == 0`. | Maintainer decision 2026-09-10. OS12: the current headline says `PASS` above `Rules failed: 1`, and the shipped `web-ui` preset triggers it. |
| D19 | `docs/decisions.md` and `docs/migration-v1-to-v2.md` are annotated as superseded rather than deleted. | Maintainer decision 2026-09-10. This design set cites `decisions.md` as the source of the `D`-exclusion rule D5 overturns, so deleting it would break the citations. |
| D20 | A JSON Schema ships once the grammar freezes, generated from or CI-checked against the loader. | Maintainer decision 2026-09-10. A schema that drifts from the validator green-lights configs the binary rejects. It cannot express the cross-field and resolve-time rules; the loader stays authoritative. |
| D21 | Rule and pattern IDs share **one namespace** and must be globally unique. | R2-1, maintainer decision 2026-09-19. Kind-qualified shape references would avoid the collision but put a prefix in every `shapeId` in every report, and a config where a rule and pattern share a name confuses a reader regardless. |
| D22 | All rules using one shape must share a footprint signature; `instanceName` is the footprint values joined `/`, or `(all)` when empty, derived from the source side only; the internal key is `shapeID` + NUL + NUL-joined values. Carries a new rule that `stripSuffixes` is an error when no rule on that shape uses `{base}`. | R2-2, maintainer decision 2026-09-19 after a deliberate re-check. Verified: rendering is injective within a signature, no existing config is broken, the restriction is vacuous for inline rules, and it is statically detectable. See [`03-instance-and-identity.md`](03-instance-and-identity.md) §4. |
| D23 | The gate is defined over a tri-state `relatedState` ∈ `present` \| `vanished` \| `absent`, replacing the misnamed `existedAtBase` boolean. `related-existed` fires on present and vanished; `related-absent` fires on absent. Refines D5. | R2-3. The boolean was true for a file added in the change under evaluation, which never existed at base, and left `related-absent` undefined. |
| D24 | `--files-from` accepts `git diff --name-status -M` output verbatim, three-field rename form included. Supersedes D9's two-column form. | R2-5. A rename carries two paths that both matter — the destination joins the change set, the source makes the related path `vanished` — and two columns cannot hold them. Git's own format needs no second grammar. |
| D25 | Experiment scripts route every fixture git call through a helper that requires and validates `$W`; `cd` is banned; `source lib.sh` is guarded. `set -e` is deliberately not used. | R2-6. The scripts could run `git add -A && git commit` against the invoking repository when a fixture failed. `set -e` was tried and aborted 10 of 13 scripts on the non-zero kyn exits they exist to measure. |

## Status

| Gate | State |
| --- | --- |
| G0 — baseline | Mostly dissolved by the scope change. OS7 still urgent. |
| G1 — semantics | Closed; revised by review 2 (D23, D24) — [`02-semantics.md`](02-semantics.md) |
| G2 — instance and identity | Closed; revised by review 2 (D21, D22) — [`03-instance-and-identity.md`](03-instance-and-identity.md) |
| G3 — grammar | Closed — [`04-grammar.md`](04-grammar.md); [`05`](05-compatibility-matrix.md) superseded as a constraint |
| G4 — cutover | Closed — [`06-cutover.md`](06-cutover.md) |
| G5 — conformance suite | Specified; matrix corrected to 75 reachable cells — [`07-conformance.md`](07-conformance.md) |
| G6 — implementation | Not authorized |

## Open questions for the maintainer

Answered 2026-09-09:

1. ~~OS3 / deletion~~ — close the gap. Recorded as D5.
2. ~~G0 timing~~ — separate v0.1.x patch, before the v3 design continues.

Currently open:

3. ~~OS5 / self-match~~ — reject at resolve time. Recorded as D8.
4. ~~`--files-from` two-column form~~ — add it. Recorded as D9, superseded by D24 (git's own `--name-status` format).

Currently open:

5. **Non-`source` groups in v2 validation (G0).** Now that OS6 proves they are
   inert, should `kyn check` reject them, warn, or keep accepting them silently?
   Rejecting is a breaking change to a shape that loads today (matrix row 8).
   Warning fits the G0 "make inert things visible" theme.
6. **Should G0 ship before the design set is approved?** Everything in G0 is a
   v2 improvement on its own merits — determinism, a broken preset, an inert
   field, two zero-value error messages. It does not depend on v3 being
   approved, and G5 cannot start without it.

# v3 configuration: design workplan

Status: active. Design only — no implementation is authorized.

This tracks the findings raised against
[`../../reviews/v3-config-proposal.md`](../../reviews/v3-config-proposal.md) by
[independent review 1](../../reviews/v3-config-independent-review-1.md) and by a
follow-up review, plus findings produced by the experiments recorded in
[`01-observed-semantics.md`](01-observed-semantics.md).

## Finding register

`IR` = independent review 1. `CR` = follow-up review. `OS` = surfaced by
experiment; see [`01-observed-semantics.md`](01-observed-semantics.md).

| ID | Finding | Sev | Gate | Status |
| --- | --- | --- | --- | --- |
| IR1 | Normalized model cannot preserve report identity | P1 | G2 | Settled — D10 |
| IR2 | Flat rules omit source-instance grouping semantics | P1 | G2 | Settled — D11 |
| IR3 | Migration claim broader than the representable subset | P1 | G4 | Open |
| IR4 | Expectation names overstate what Kyn observes | P2 | G1 | Settled — D6 |
| IR5 | Version-specific decoding lacks a compatibility matrix | P2 | G3 | Settled — measured, 05-compatibility-matrix.md |
| IR6 | Map-key validation not fully deterministic | P2 | G0/G2 | Settled — D3, D1 |
| CR1 | `stripSuffixes` is a grouping input the inline form cannot express | P1 | G2 | Settled — D12 |
| CR2 | Migration has no answer for declared-but-inert v2 groups | P1 | G4 | Settled — D2 |
| CR3 | Resolve-time error selection becomes order-dependent under maps | P2 | G0 | Confirmed by OS7 |
| CR4 | Generated default messages are an unowned output contract | P2 | G1 | Settled — D7 |
| CR5 | `explain` clause names are a JSON contract the shorthand must map to | P2 | G2 | Settled — D13 |
| CR6 | "v3 config" collides with product versioning; `description` dropped | nit | G3 | Settled — D14 |
| OS3 | Kyn cannot observe deletion; the default shape is silenced by it | **P0** | G1 | Settled — D5 |
| OS8 | Renaming the related file away is equally invisible | **P0** | G1 | Settled — D5 |
| OS5 | The proposed common form self-matches; OS3 hides the mistake | P1 | G2 | Settled — D8 |
| OS7 | Error selection is already non-deterministic (ships today) | P1 | G0 | Specified — v0.1.x patch, not yet written |
| OS10 | `rule.description` is parsed and never read | P2 | G0/G3 | Settled — D14 |

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

## Gates

Ordered. A gate may not open until the one before it is closed.

### G0 — Stabilize the baseline (independent of v3)

Differential conformance testing is worthless against a non-deterministic
baseline, so this comes first and ships on its own schedule. Everything here is
a v2 improvement that stands on its own merits.

- Fix OS7: sort map keys before every validation and resolution pass, so error
  selection is stable. Add tests that assert *which* error is reported.
- Fix OS6: stop `kyn init` emitting inert `groups.story` / `groups.tests`.
- Fix OS10: route `rule.description` into SARIF `fullDescription` and `explain`,
  so an accepted field stops being inert.
- Fix the zero-value leaks in two validation messages: a missing `version`
  reports `unsupported config version 0`, and a missing `severity` reports
  `has invalid severity ""` (matrix rows 15 and 17).
- Decide whether validation should reject or warn on non-`source` groups.

Deliverable: a v0.1.x patch. No v3 content.

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

### G4 — Migration

- IR3: formal migratable normal form; refuse everything outside it.
- CR2 is settled (D2) and becomes a documented no-op with a differential test.
- Deprecation policy for configs that cannot migrate.

Deliverable: `06-migration.md`.

### G5 — Differential conformance harness

Compare check and explain text and JSON, SARIF, RDJSON, Checkstyle, dry-run,
flags, result counts, ordering, and exit codes across equivalent configs. The
existing v1→v2 end-to-end test (`e2e/workflows_test.go:95`) is the standard but
covers one scenario.

### G6 — Implementation

Not authorized. Requires G0–G5 closed and an approved specification.

## Decisions taken

Recorded here as they close; each is backed by an experiment or a code citation.

| ID | Decision | Basis |
| --- | --- | --- |
| D1 | The normalized policy model is an **ordered slice sorted by ID**, never a Go map. | OS7 shows map iteration already produces non-deterministic error selection. A v3 keyed by mapping would inherit it per rule and per pattern. |
| D2 | Migration drops non-`source` groups, as a documented no-op. | OS6 proves output is byte-identical without them across all six modes. |
| D3 | `rules` as a YAML mapping keyed by ID is safe; duplicate detection is free. | `yaml.v3` with `KnownFields(true)` already rejects duplicate keys with line numbers: `line 11: mapping key "test" already defined at line 10`, exit 2. Answers proposal open question 2 and closes half of IR6. |
| D4 | The expectation vocabulary is a closed 2x2 grid (existence × change-set membership) with existence optionally promoted to a gate — not an open enum. | OS2. Reduces the author's stated combinatorial-growth risk. |
| D5 | Close the deletion gap. The applicability gate becomes "existed at base", derived from `D` and rename-source entries Kyn already parses and discards. Version-gated to v3. | Maintainer decision 2026-09-09. Validated by `e9-gate-prototype.sh`: fixes OS3 and OS8, changes 2 of 5 scenarios, adds no false positives, needs no new git call. See [`02-semantics.md`](02-semantics.md) §4. |
| D6 | The expectation vocabulary is two orthogonal fields — `when` and `expect` — not fused enum names. `when` defaults to `always`, never to a lenient gate. | OS1/OS2 plus IR4. Removes the concealed skip, and makes enum growth additive instead of multiplicative. See [`02-semantics.md`](02-semantics.md) §6. |
| D7 | `message` may be generated, but the generated text is a versioned, fixture-tested contract containing only the rule ID and resolved related path — never a config path or absolute path. | CR4. Absolute paths would leak the CI checkout root into reviewdog comments. |
| D8 | A resolved `related` path that matches its own rule's `match`/`exclude` is a config error, reported at resolve time with exit 2. | Maintainer decision 2026-09-09. A sound static check is glob intersection; the resolve-time check is exact and cheap. See [`03-instance-and-identity.md`](03-instance-and-identity.md) §5. |
| D9 | `--files-from` gains an optional `status<TAB>path` form so explicit mode can express deletion. Single-column lines still mean `modified`. | Maintainer decision 2026-09-09. Without it a policy designed with `--files` passes locally and fails in git-mode CI. See [`02-semantics.md`](02-semantics.md) §5. |
| D10 | Report wire format is unchanged across config versions: `familyId` carries the source-shape ID, `familyName` the instance name. | IR1. The proposal's own rule forbids changing report schemas because the input config version changed; renaming would break every consumer at the moment a repo migrates. See [`03-instance-and-identity.md`](03-instance-and-identity.md) §6. |
| D11 | Instances are keyed on the **source shape** — the (`match`, `exclude`, `stripSuffixes`) triple — not on the rule. A pattern is a named source shape; an inline rule owns an anonymous one identified by its rule ID. | OS9 shows rule-keyed instances would regroup and reorder every report. Shape-keying maps a v2 family one-for-one and preserves the contract. |
| D12 | `match`, `exclude` and `stripSuffixes` are all legal inline on a rule. A pattern is a named bundle of exactly those three and carries no extra capability. | OS4/CR1. Confining `stripSuffixes` to patterns left the inline form unable to express the shipped `api` preset. Also makes inline and pattern-backed forms normalize identically by construction. |
| D13 | Structural identity fields stay stable across config versions; vocabulary that echoes the user's own config (`explain`'s `clause`) follows the config version. | CR5. Reporting `if.kinExists` for a config containing neither `if` nor `kinExists` would be the actual defect. |
| D14 | `description` survives into v3 and becomes functional (SARIF `fullDescription`, `explain`). `version:` is a config-schema version; prose calls v3 "the rule-centric format". | OS10 + CR6. Dropping `description` would break configs that already set it, and the fix is one line. |

## Open questions for the maintainer

Answered 2026-09-09:

1. ~~OS3 / deletion~~ — close the gap. Recorded as D5.
2. ~~G0 timing~~ — separate v0.1.x patch, before the v3 design continues.

Currently open:

3. **OS5 / self-match.** Should v3 auto-exclude resolved related paths from
   source matching, or reject a config whose `related` template can match its
   own `match` globs? Auto-exclude is quieter; rejection is louder and matches
   the repository's strict-validation habit. Decided in G2.
4. **`--files-from` two-column form.** [`02-semantics.md`](02-semantics.md) §5
   proposes `status<TAB>path` so explicit mode can express deletion. This adds
   CLI surface, which the steering doc guards. Worth it, or is degrading to
   current behavior in explicit mode acceptable?

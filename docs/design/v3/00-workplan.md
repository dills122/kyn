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
| IR1 | Normalized model cannot preserve report identity | P1 | G2 | Open — widened, see below |
| IR2 | Flat rules omit source-instance grouping semantics | P1 | G2 | Open |
| IR3 | Migration claim broader than the representable subset | P1 | G4 | Open |
| IR4 | Expectation names overstate what Kyn observes | P2 | G1 | Open — reframed by OS3 |
| IR5 | Version-specific decoding lacks a compatibility matrix | P2 | G3 | Open |
| IR6 | Map-key validation not fully deterministic | P2 | G0/G2 | Partly settled — D3, D1 |
| CR1 | `stripSuffixes` is a grouping input the inline form cannot express | P1 | G2 | Confirmed by OS4 |
| CR2 | Migration has no answer for declared-but-inert v2 groups | P1 | G4 | Settled — D2 |
| CR3 | Resolve-time error selection becomes order-dependent under maps | P2 | G0 | Confirmed by OS7 |
| CR4 | Generated default messages are an unowned output contract | P2 | G1 | Open |
| CR5 | `explain` clause names are a JSON contract the shorthand must map to | P2 | G2 | Open |
| CR6 | "v3 config" collides with product versioning; `description` dropped | nit | G3 | Open |
| OS3 | Kyn cannot observe deletion; the default shape is silenced by it | **P0** | G1 | Open — needs product decision |
| OS5 | The proposed common form self-matches; OS3 hides the mistake | P1 | G1 | Open |
| OS7 | Error selection is already non-deterministic (ships today) | P1 | G0 | Open — fix independently |

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
baseline, so this comes first and ships on its own schedule.

- Fix OS7: sort map keys before every validation and resolution pass, so error
  selection is stable. Add tests that assert *which* error is reported.
- Fix OS6: stop `kyn init` emitting inert `groups.story` / `groups.tests`.
- Decide whether validation should reject or warn on non-`source` groups.

Deliverable: a v0.1.x patch. No v3 content.

### G1 — Freeze semantics

- Product decision on OS3 (deletion observability). Blocks everything else in
  this gate.
- Self-match policy for OS5: auto-exclude resolved related paths, error, or warn.
- Expectation grid frozen against the OS1 truth table, with names chosen after
  the OS3 decision.
- CR4: decide whether `message` may be generated, and if so version the text.

Deliverable: `02-semantics.md` with a truth table per final expectation name.

### G2 — Instance and identity model

- IR2 + CR1: instance-key derivation, `stripSuffixes` placement, source
  aggregation, template-context agreement, status aggregation.
- IR1: normalized identity and its projection into all seven output modes.
- CR5: clause naming in `explain` traces.
- D1 applies throughout: ordered slices, never maps.

Deliverable: `03-instance-and-identity.md`, including a normalized Go type
sketch and a field-by-field report projection table.

### G3 — Grammar and compatibility matrix

- IR5: inventory every v1/v2 shape that loads today, mark each contractual or
  incidental.
- v3 grammar, informed by G1 and G2.
- CR6: naming, and whether `description` survives.

Deliverable: `04-grammar.md`, `05-compatibility-matrix.md`.

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

## Open questions for the maintainer

1. **OS3 / deletion.** Close the capability gap, or document the blindness?
   Closing it means revisiting the `docs/decisions.md` rule that excludes `D`
   paths from the change set, which is a change-set contract change.
2. **G0 timing.** Ship the determinism and inert-group fixes as a v0.1.x patch
   now, or fold them into the v3 branch?

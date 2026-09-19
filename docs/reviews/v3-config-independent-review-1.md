# Independent review: v3 configuration proposal

Review instance: 1 of 3

Reviewed target: `docs/reviews/v3-config-proposal.md`

Verdict: **Not ready for specification approval or implementation.** The flat
rule-centric form and constrained pattern reuse are coherent enough to keep
refining; the blockers are missing semantic contracts, not a reason to abandon
the direction.

## Findings

No P0 findings.

### P1 — The normalized model cannot preserve existing report identity

The proposed normalized representation omits family and instance identity while
also saying pattern names should normally disappear at runtime. Current
contracts expose `familyId` and `familyName` in check and explain results, use
them for sorting, include them in SARIF, and expose family instances with named
kin in dry-run output.

Failure mode: migrating one v2 family into one pattern plus multiple v3 rules
leaves no specified value for `familyId`, no stable shared instance identity,
and no named kin map for `--dry-run-resolve`. Using the rule ID changes grouping
and ordering; using the pattern ID fails for inline rules; leaving values empty
changes field meaning and machine output.

Required fix: define the normalized instance and relation identity model before
approving the grammar, including its exact projection into check, explain,
SARIF, RDJSON, Checkstyle, and dry-run reports. For migrated policies, preserve
legacy family/kin identity as compatibility metadata or explicitly version the
affected report contracts.

Evidence:

- `docs/reviews/v3-config-proposal.md:197`
- `internal/rules/result.go:19`
- `internal/rules/evaluator.go:83`
- `internal/report/sarif.go:83`
- `internal/report/dryrun.go:22`

### P1 — Flat rules omit source-instance grouping semantics

Current resolution groups matched source files by family ID plus
`{dir}/{base}`, evaluates each rule once per grouped instance, and aggregates
its sorted source files. It rejects templates that resolve differently across
source files in the same instance. Status applicability uses “any matching
source within the instance” semantics.

The proposal retains matching, suffix normalization, templates, and source
statuses but does not define the instance key or aggregation behavior.

Failure mode: a `.component.ts` and `.component.html` change could produce two
rule results instead of one. A template containing `{name}`, `{file}`, or
`{ext}` could silently resolve to two paths rather than raising the current
ambiguity error. `on: [renamed]` is undefined when only one of several grouped
source files was renamed.

Required fix: add normalized semantics for instance-key derivation, source
aggregation, template-context agreement, evaluation cardinality, and status
aggregation. Differential tests should cover multi-extension sources, reordered
input, ambiguous templates, and mixed statuses.

Evidence:

- `internal/family/resolver.go:53`
- `internal/family/resolver.go:117`
- `internal/family/resolver_test.go:153`
- `internal/rules/evaluator.go:137`

### P1 — The migration claim is broader than the representable subset

The proposal says a v2 family with several kin and rules can become a pattern
plus one rule per related path. V2 allows AND-combined multi-kin applicability
and assertions, cross-kin clauses, several assertion types, and emitted actions.
Existing explain tests exercise multiple assertions as one atomic result.

For example, splitting this v2 rule is not equivalent:

```yaml
if:
  kinExists: [story, spec]
assert:
  kinChanged: [story, spec]
actions:
  emit: [reviewRequired]
```

If `story` exists and `spec` does not, v2 skips the entire rule and emits
nothing. Two `changed-if-present` v3 rules would evaluate the story rule,
potentially fail it, alter result counts, and emit a flag. Cross-kin policies
such as “if metadata exists, include documentation” cannot fit a single
`related` field.

Required fix:

1. Define a formal migratable normal form, likely limited to one logical kin
   per rule, same-kin shorthand combinations, supported source status filters,
   and explicitly handled actions.
2. Reject everything outside that subset before producing migration output.
3. If v3 launches without full parity, keep v2 fully supported until an
   equivalent advanced form exists.
4. Prove equivalence through differential check, explain, report, and exit-code
   tests, not only normalized-structure comparisons.

Evidence:

- `docs/reviews/v3-config-proposal.md:225`
- `docs/site/config.md:158`
- `internal/rules/explain_test.go:141`

### P2 — Expectation names overstate what Kyn observes

`kinChanged` checks membership in the collected path set, while `kinUnchanged`
checks absence from that set. Existence is checked separately against the
current checkout. Git collection excludes deleted paths, and explicit inputs
mark supplied paths as modified.

Consequences:

- `unchanged` can pass when the related file was deleted.
- `changed` can pass for a nonexistent path supplied manually.
- `missing` means missing from the current checkout, not deleted in the selected
  diff.
- `changed-if-present` silently becomes a skipped rule when the current path is
  absent.

Required fix: specify a complete truth table over existence, change-set
membership, input mode, and source status. Prefer precise vocabulary such as
`in-change-set`, `not-in-change-set`, and `in-change-set-if-existing`, or
document equally exact names. Define canonical explain traces and actionable
messages for positive and negative expectations.

Evidence:

- `docs/reviews/v3-config-proposal.md:129`
- `internal/rules/evaluator.go:157`
- `docs/site/concepts.md:30`

### P2 — Version-specific decoding lacks a compatibility matrix

Separate document types are the correct direction, but “continue loading v1
and v2” does not define the existing accepted grammar. Current v2 compatibility
includes top-level family `include` and `exclude` when `groups` is absent, plus
legacy `when` and `require` aliases when they are not mixed with `if` and
`assert`.

Failure mode: clean v1/v2 structs could reject configurations that load today,
despite the proposal promising continued support.

Required fix: inventory accepted native and compatibility shapes by version,
decide which are contractual, and turn that matrix into loader and normalization
fixtures before replacing the shared decoder.

Evidence:

- `docs/reviews/v3-config-proposal.md:197`
- `docs/site/config.md:88`
- `docs/site/config.md:179`
- `internal/config/config.go:31`

### P2 — Map-key validation is not fully deterministic yet

Mapping-key IDs are reasonable, but sorting only before normalization and
reporting leaves structural and semantic validation vulnerable to map iteration
order. Once duplicate YAML keys are decoded into an ordinary map, reliable
duplicate detection and source-position diagnostics may be lost.

Failure mode: a file containing several invalid rules or patterns could report
a different first error between runs, weakening deterministic CLI behavior and
making migrations harder to diagnose.

Required fix: detect duplicate keys while parsing the YAML node structure,
retain source positions, sort rule and pattern IDs before every semantic
validation pass, and test that multiple-invalid-input diagnostics consistently
select the same field.

Evidence:

- `docs/reviews/v3-config-proposal.md:177`
- `internal/family/resolver_test.go:137`

## Plan review

Before implementation, add these ordered gates:

1. Freeze expectation truth tables and advanced-rule scope.
2. Define normalized policy, instance, relation, clause-provenance, and
   report-projection types.
3. Define the versioned grammar and compatibility matrix.
4. Specify migration eligibility and the non-parity support/deprecation policy.
5. Build a differential conformance matrix across existence, membership,
   status, grouping, and every report mode.
6. Only then implement loaders, normalization, evaluation adaptation,
   migration, starters, and documentation.

Migration tests should compare exact check and explain text and JSON, machine
reports, dry-run output, flags, result counts, ordering, and exit codes. The
current v1-to-v2 end-to-end test demonstrates the expected standard but covers
only one scenario (`e2e/workflows_test.go:95`).

## Notable strengths

- The inline common form materially reduces first-policy ceremony.
- Patterns are narrowly bounded: one reference, no inheritance, no composition,
  no overrides, and no imports.
- Version-specific documents feeding a normalized model are cleaner than adding
  more aliases to the current shared structs.
- Lossy migration is explicitly prohibited.
- The proposal accurately identifies its unresolved advanced semantics rather
  than presenting them as complete.

## Author-claim reconciliation

| Author claim | Status | Consequence |
| --- | --- | --- |
| Inline and pattern-backed forms can behave identically | Unverified | Plausible, but needs one normalized representation and differential fixtures. |
| Pattern reuse is optional and constrained | Confirmed by proposed grammar | This is the strongest part of the design and does not recreate general inheritance. |
| Existing report contracts remain stable | Contradicted as currently specified | Family, instance, kin identity, and dry-run projection are missing. |
| Migration will never weaken policy | Intent confirmed; mechanism unverified | A formal migratable subset and behavioral equivalence harness are required. |
| A normalized IR is a substantial refactor | Confirmed | Current CLI, resolver, evaluator, and reports directly use `config.Rule` and `family.Instance`. |
| Partial parity may strand advanced users | Confirmed | V2 cannot enter deprecation until parity or a durable compatibility policy exists. |

## Verification performed

- Verified branch `codex/usability-onboarding-review` at
  `0ce1bd58cfa2fe45194b0179e96e34528fc05f1a`.
- Verified the dirty state matched the bootstrap and excluded unrelated
  onboarding/site edits.
- Inspected repository steering, canonical documentation, configuration
  loading, validation and migration, family resolution, rule evaluation and
  explanation, reports, starter configurations, and relevant tests.
- `go test ./internal/config ./internal/family ./internal/rules ./internal/report`
  passed.
- `go test ./e2e` passed.
- No files were modified by the reviewer.

These tests establish current v1/v2 behavior only. There is no v3 parser,
normalizer, migration, or executable proposal evidence.

## Open questions and residual risks

- Should native v3 results retain synthetic family terminology or deliberately
  version report vocabulary?
- Does `description` remain supported? It exists in v2 but is absent from the
  proposed v3 rule fields and normalized model.
- Are `patterns` and `use` clearer than `sources` and `source`?
- Should the expanded form preserve atomic multi-related-path rules or leave
  them in v2?
- How should migration dry-run output differ from ordinary side-by-side output,
  and how should users compare behavioral equivalence?

## Recommended next actions

1. Resolve the three P1 findings in the proposal.
2. Add a normalized-model sketch containing instance and report identities.
3. Add explicit expectation and migration truth tables.
4. Define the v1/v2 compatibility matrix and v3 support/deprecation policy.
5. Submit the revised proposal for another independent review only if those
   changes materially close the blockers.

# Author Explanation

## Intent And Success Criteria

The proposal attempts to make Kyn’s common configuration read as one direct
source-to-related-path rule while retaining optional, constrained reuse for
repeated source matchers. The intended user is a developer adopting Kyn locally
or in CI who should not need to learn the evaluator’s family/group/kin graph for
the first policy.

Success means an inline rule and a pattern-backed rule can express the same
behavior; reuse is optional; validation stays strict; migration never weakens a
policy; and existing runtime, report, path, determinism, and exit contracts do
not change solely because of v3 syntax.

## Plan-To-Implementation Traceability

There is no implementation. The current deliverable is the exploratory design
in `docs/reviews/v3-config-proposal.md`. The requested sequence is independent
review first, then a separate Claude critique, then human reconciliation before
a formal specification or implementation plan.

Implemented: none.

Proposed but unresolved:

- final schema vocabulary;
- expectation enumeration;
- advanced v2 expressive parity;
- signals and multiple-kin behavior;
- migration eligibility and deprecation window;
- generated-message stability;
- local-only versus eventual external pattern reuse.

## Technical Approach And Flow

The proposed end-to-end flow is:

1. Decode YAML into a version-specific document type.
2. Validate structural rules, IDs, patterns, paths, and templates.
3. Expand an inline matcher or one referenced local pattern into an independent
   rule definition.
4. Translate expectation shorthand into explicit normalized applicability and
   assertions.
5. Feed a version-neutral policy model to the existing evaluation pipeline.
6. Preserve current report sorting, machine contracts, and exit mapping.

Patterns are intended as compile-time configuration reuse, not runtime entities.

## Changed-Component Walkthrough

No production component changed. Review artifacts only:

- `v3-config-proposal.md` owns the proposed schema, constraints, migration
  boundaries, alternatives, and open questions.
- `v3-config-review-bootstrap.md` freezes the independent review scope.
- `v3-config-author-explanation.md` records this author rationale separately so
  the reviewer can inspect the proposal first.

If approved later, likely implementation ownership would span version-specific
config decoding/normalization, validation, migration, starters, tests, and user
documentation. Those tasks are not yet planned or authorized.

## Decisions And Rejected Alternatives

Author preference: a flat rule-centric common form with one related path per
rule. Repetition of source matching is accepted for simple configs. When it
becomes material, users may promote only matching and name normalization into a
same-file pattern.

Rejected or deferred:

- Mandatory named source/family definitions: too much indirection for first use.
- Nested target maps: reduce repetition but recreate hierarchy.
- Arbitrary rule-template inheritance: complex merge and provenance semantics.
- YAML anchors as product API: textual rather than domain-aware reuse.
- External imports in v3.0: unresolved resolution, cycle, precedence, and
  security concerns.

## Invariants And Boundary Conditions

The design must preserve:

- exit codes 0 success, 1 rule failure, 2 usage/config validation, and 3
  runtime/provider failure;
- deterministic ordering independent of YAML map iteration;
- slash-normalized paths relative to `--cwd`;
- containment and symlink safety for resolved paths;
- current machine-report contracts unless separately versioned;
- explicit, non-lossy migration behavior;
- a stateless CLI-only product boundary.

Pattern-specific intended invariants:

- `use` versus inline `match` is exclusive;
- one pattern reference per rule;
- no pattern inheritance or composition;
- patterns contain source-shape fields only;
- initial pattern-backed rules cannot override pattern fields.

## Verification Performed And Results

The author inspected current configuration structs, validation, v1-to-v2
migration, starter generators, canonical configuration documentation, and the
v0.1.3 onboarding findings. No v3 parser, migration, normalization prototype, or
behavioral test exists.

Earlier documentation work passed `go test ./...` and `mkdocs build --strict`,
but those checks do not validate the v3 proposal and should not be credited as
v3 evidence.

## Risks, Tradeoffs, And Maintenance Costs

- Expectation enums may hide rather than remove policy complexity and may grow
  combinatorially.
- One related path per rule can produce repeated matcher blocks and duplicated
  resolution work unless normalization or evaluation caches it safely.
- Named patterns may be a renamed subset of v2 families rather than a true
  simplification.
- Map-key rule IDs improve visual economy but require duplicate-key detection
  and deterministic sorting.
- Generated messages could change human-output expectations or become vague.
- A normalized IR is architecturally cleaner but is a substantial refactor with
  migration and explain-diagnostic consequences.
- Partial v2 expressive parity could strand advanced users or force two mental
  models to coexist indefinitely.
- Pattern fields may need future captures or transforms that challenge the
  initial strict boundary.

## Deviations, Deferrals, And Known Gaps

The proposal deliberately does not settle how v3 represents:

- informational emit-only rules;
- applicability based on related-path existence or absence outside the provided
  shorthand;
- multiple related paths in one assertion;
- combinations of several assertions;
- reverse or bidirectional relationships;
- migration of every valid v2 rule;
- imports or shared pattern packages.

It also lacks a formal JSON schema, grammar, normalized Go type design,
prototype migration, complexity measurements from real configs, and user
testing. These omissions should materially affect the reviewer’s verdict on
specification readiness.

## Challenge Points For The Reviewer

Apply extra skepticism to:

1. Whether `patterns`/`use` materially improves on families or simply renames
   them.
2. Whether `expect` values are intuitive, orthogonal, and extensible.
3. Whether skip behavior is too hidden in `changed-if-present`.
4. Whether one related path per rule loses useful atomic grouping.
5. Whether map-key IDs harm diagnostics, ordering, or YAML tooling.
6. Whether source matching can be duplicated safely without performance or
   inconsistent-resolution costs.
7. Whether a version-neutral normalized model can preserve clause-level explain
   output across versions.
8. Whether partial v2 parity and selective migration are acceptable product
   boundaries.
9. Whether local pattern reuse needs override, composition, or parameterization
   sooner than expected.
10. Whether a smaller v2 shorthand could achieve the benefit without a new
    config version.

The author does not provide a readiness verdict. The reviewer should return one
of the skill’s allowed verdicts based on the proposal’s fitness to advance into
a formal specification, not implementation release readiness.

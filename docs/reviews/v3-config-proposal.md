# Kyn v3 configuration proposal

Status: exploratory proposal frozen for independent review on 2026-09-05.
This is not an approved specification and does not change shipped behavior.

## Objective

Make the common related-file policy read as one direct relationship instead of
requiring users to assemble `families`, `groups.source`, named `kin`, and a
separate `rules` entry. Preserve an optional reuse mechanism for repositories
where several rules share the same source matching and name normalization.

The proposal is intentionally a configuration-language pivot, not a request for
new runtime policy capabilities. Existing evaluation ordering, path safety,
reports, machine schemas, and exit codes should remain stable.

## User problem

Version 2 exposes several evaluator concepts before a user can state the common
policy “when this source path changes, include that related path.” A typical
single relationship currently requires:

1. A family ID.
2. A `groups.source` matcher.
3. A named kin template.
4. A rule that references the family ID.
5. Conditions and assertions that reference both `source` and the kin name.

That structure is expressive, but the indirection and indentation make the
first useful policy harder to author and explain. The v0.1.3 onboarding review
also found that users must understand applicability versus assertion semantics
early because starter rules use `if.kinExists`.

## Design principles

1. The simplest rule must not require a reusable declaration.
2. One rule should connect one source matcher directly to one related path.
3. Reuse should be an explicit promotion of repeated source matching, not a
   prerequisite for ordinary rules.
4. A small amount of repetition is preferable to hidden inheritance or
   cross-referenced object graphs.
5. Configuration versions should compile into a normalized internal policy
   model instead of accumulating version aliases in shared structs.
6. V3 must not silently weaken a v2 policy during migration.
7. Deterministic results must not depend on YAML map iteration order.

## Proposed common form

Rule IDs are mapping keys, removing the `- id:` sequence ceremony while keeping
stable identifiers for reports and machine integrations.

```yaml
version: 3

rules:
  handler-test:
    match:
      - "internal/**/*_handler.go"
    related: "{dir}/{name}_test.go"
    expect: changed-if-present
```

The intended reading is:

> When a path matching `internal/**/*_handler.go` changes, its related test path
> must be part of the change set if that test currently exists.

Proposed defaults:

- A source change is the implicit trigger.
- `severity` defaults to `error`.
- `message` is optional; Kyn produces an actionable default containing the rule
  ID and resolved expected path.
- `exclude`, `stripSuffixes`, source statuses, and emitted flags are absent
  unless needed.
- Each rule owns exactly one `related` path in the common form.

## Optional local pattern reuse

Repeated source matching and name normalization can be promoted into a named
pattern in the same file:

```yaml
version: 3

patterns:
  web-component:
    match:
      - "src/**/*.component.ts"
      - "src/**/*.component.html"
    exclude:
      - "src/**/generated/**"
    stripSuffixes:
      - ".component"

rules:
  story-sync:
    use: web-component
    related: "{dir}/{base}.stories.ts"
    expect: changed-if-present

  test-sync:
    use: web-component
    related: "{dir}/{base}.spec.ts"
    expect: exists-and-changed
    severity: warn
    message: "Include the component test in this change."
```

Patterns own only source-shape concerns:

- `match`
- `exclude`
- `stripSuffixes`
- Potential future source-path capture or normalization options

Rules own policy concerns:

- `related`
- `expect`
- source change-status filters
- `severity`
- `message`
- emitted flags

This boundary is intended to prevent patterns from becoming general rule
templates or an inheritance system.

## Proposed expectation vocabulary

The following table is a candidate, not a final enumeration. Names should be
reviewed for accuracy because Kyn observes path membership and existence rather
than content quality.

| V3 expectation | V2 semantic expansion |
| --- | --- |
| `changed-if-present` | `if.kinExists` plus `assert.kinChanged` |
| `exists-and-changed` | `assert.kinExists` plus `assert.kinChanged` |
| `changed` | `assert.kinChanged` without an existence precondition |
| `unchanged` | `assert.kinUnchanged` |
| `exists` | `assert.kinExists` |
| `missing` | `assert.kinMissing` |

Questions for review:

- Does an enum make common intent clearer, or merely hide combinations users
  eventually need to understand?
- Should names use `included` rather than `changed` to avoid implying content
  inspection?
- Are `unchanged`, `exists`, and `missing` meaningful without explicit
  applicability controls?
- Does `changed-if-present` conceal the important skip behavior too much?

## Candidate optional fields

An expanded rule could remain flat:

```yaml
rules:
  generated-go:
    match:
      - "proto/**/*.proto"
    exclude:
      - "proto/vendor/**"
    related: "gen/{dir}/{name}.pb.go"
    expect: exists-and-changed
    on: [added, modified, renamed]
    severity: error
    message: "Regenerate the Go output for this protobuf change."
```

The name and exact semantics of `on` remain open. Explicit file lists record
paths as modified today, while Git input preserves supported change statuses.
V3 must retain an actionable validation or documentation story for that
difference.

## Pattern validation rules

Candidate constraints:

1. `use` and inline `match` are mutually exclusive.
2. A rule can use exactly one pattern.
3. Patterns cannot reference or extend other patterns.
4. A rule using a pattern cannot override `match`, `exclude`, or
   `stripSuffixes` in the initial v3 release.
5. Unknown pattern references are errors.
6. Duplicate rule or pattern IDs are impossible at the YAML mapping level and
   must still be rejected if a decoder exposes duplicates.
7. Unused patterns should be errors or warnings; the policy is unresolved.
8. Empty match lists, path traversal, invalid templates, and incompatible
   multi-extension templates retain strict validation.
9. Pattern and rule IDs are sorted before normalization and reporting.

For a legitimate source variant, the initial design prefers a second explicit
pattern over inheritance or local merging.

## Normalization architecture

Do not add v3 fields to the current `config.Config`, `config.Family`, and
`config.Rule` structs, which already carry v1 and v2 compatibility aliases.
Use version-specific document types and compile them into a version-neutral
policy representation:

```text
v1 YAML ──→ v1 document ──┐
v2 YAML ──→ v2 document ──┼─→ normalized policy model ─→ evaluator ─→ reports
v3 YAML ──→ v3 document ──┘
```

The normalized representation needs to retain at least:

- stable rule ID;
- source includes and excludes;
- base-name normalization;
- resolved related-path definition;
- source status applicability;
- existence/change assertions;
- severity, message, and emitted flags;
- enough clause identity to keep `explain` useful and deterministic.

Pattern expansion is a compile-time configuration step. Runtime results should
identify the rule and resolved path; pattern names should appear only in
configuration diagnostics unless they prove useful elsewhere.

## Compatibility and migration

Proposed boundaries:

- Continue loading v1 and v2 during an explicitly defined deprecation window.
- Add `kyn config migrate --from v2 --to v3` only after the v3 grammar and
  normalized semantics are approved.
- Migrate only representations that can be proven equivalent.
- Refuse or preserve v2 for configurations that cannot be represented without
  semantic loss.
- Never change machine report schemas or exit behavior solely because the input
  config version changed.
- Provide a dry-run migration that shows the normalized rules and any defaults
  inserted by v3.

The common v2 family containing several kin and rules can become one v3 pattern
plus one v3 rule per related path. A v2 family or rule that uses combinations
outside the v3 expectation vocabulary may need an expanded v3 form or may remain
v2 until that design exists.

## Signals and advanced conditions

V2 supports informational rules with `actions.emit`, existence/missing
applicability, status applicability, and multiple kin references. The flat v3
common form does not yet define a complete replacement for every combination.

Candidate strategies for later decision:

1. Keep v3 deliberately focused on related-path enforcement and let advanced
   configurations remain v2.
2. Add an explicit expanded rule form that exposes `when`, `assert`, and
   `emit` without families or named kin.
3. Grow the expectation enumeration and add separate flat applicability fields.

The proposal does not select one yet. An independent reviewer should treat this
as a central completeness and migration risk, not a documentation detail.

## Alternatives considered

### Keep v2 and add documentation only

Rejected as the primary direction because the onboarding walkthrough can teach
the model, but it cannot remove the indirection users must maintain.

### Source-centric definitions with nested targets

This removes repeated match patterns but makes named source declarations
mandatory and brings back hierarchy similar to families and kin. It may be
useful as an advanced representation, but it should not be the first-use form.

### Arbitrary reusable rule templates

Rejected because merge precedence, local overrides, nested inheritance, and
diagnostic provenance would create a second configuration language.

### YAML anchors and merge keys

Rejected as the official reuse mechanism because they perform textual merging,
are difficult to validate as domain concepts, and make generated migrations and
diagnostics harder to understand.

### External pattern imports in v3.0

Deferred. Imports introduce file-resolution scope, cycles, precedence,
provenance, and security considerations. Same-file patterns should first prove
that the reuse boundary is valuable. A future `imports` key may be reserved, but
it should not be accepted before semantics are designed.

## Success criteria for an approved design

1. A single common relationship is expressible without a pattern, family,
   group, or named kin reference.
2. Two rules can reuse one local source pattern without inheritance or merge
   semantics.
3. Inline and pattern-backed forms normalize to identical policy behavior.
4. Validation errors identify the rule or pattern and exact invalid field.
5. Every supported v2-to-v3 migration is behaviorally equivalent across check,
   explain, text, and machine reports.
6. Unsupported v2 combinations fail migration explicitly rather than weakening
   policy.
7. Evaluation order, report order, path normalization, path containment, and
   exit codes remain deterministic and compatible.
8. Starter configurations and the first-run tutorial no longer need to teach
   families, source groups, named kin, `if`, and `assert` for the common case.

## Open questions

1. Is `patterns`/`use` the right vocabulary, or would
   `sources`/`source` be more domain-specific?
2. Should `rules` be a mapping keyed by ID or remain a sequence with explicit
   `id` fields?
3. Is one related path per common rule the right deliberate duplication trade?
4. Which expectation names communicate path semantics without implying content
   analysis?
5. Should unused patterns be errors, warnings, or allowed?
6. Does v3 need full v2 expressive parity before release?
7. What expanded form, if any, should handle emitted flags, multiple kin, and
   advanced applicability?
8. Should rule messages be generated by default, and can that be done without
   destabilizing human-readable output expectations?
9. How long should older versions remain supported, and should v1-to-v3 migrate
   directly or only through v2?
10. Should external pattern imports be explicitly reserved or omitted until a
    concrete use case exists?

## Explicit non-goals

- No daemon or watch mode.
- No PR integration or hosted service.
- No plugin or remote policy registry.
- No monorepo graph adapter.
- No content or language-syntax inspection.
- No implementation work before the proposal and migration semantics are
  reviewed and approved.

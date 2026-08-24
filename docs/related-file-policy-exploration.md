# Related-File Policy Feature Exploration

## Status

- Status: exploratory; not approved for implementation
- Decision owner: Kyn maintainer
- Last updated: 2026-08-24
- Repository baseline: `71b8cf79fe78883fd4d5e23e0c129ba9225365d2`
- Evidence environment: Go 1.23.4 on darwin/arm64

This document retains candidate features for review. Its examples are illustrative, not proposed
configuration contracts. Approving a problem statement does not approve its example syntax or an
implementation plan.

## Decision Question

Which tightly related capabilities should Kyn add after v0.1.2 so that it better enforces Storybook
and other related-file change contracts without becoming a framework-specific integration tool?

The next gate is to select one or more problems for deeper research. Implementation should begin
only after the relevant behavior, compatibility expectations, and acceptance criteria have been
approved in a feature specification.

## Executive Direction

The proposed product framing is:

> Kyn enforces change contracts between files that are expected to evolve together.

Storybook is the flagship example: when a component changes, its story may need to be created,
reviewed, renamed, or deleted in the same change. The same mechanism should remain useful for tests,
configuration examples, schemas, generated files, module documentation, and similar repository
relationships.

Kyn should continue to determine whether related files participated in a change set. It should not
claim that the content of a story, test, or generated artifact is semantically correct.

## Working Assumptions

1. Kyn remains a stateless, deterministic, CLI-only tool for local and CI use.
2. Configuration remains the source of relationship policy.
3. Storybook receives strong defaults and examples, while the rule engine remains framework-neutral.
4. Existing exit-code, ordering, path-normalization, and report contracts remain stable unless a
   separately approved change explicitly replaces them.
5. New behavior should improve generic related-file policy rather than add Storybook AST parsing,
   publishing, automatic file generation, or PR-provider integration.

## Current-State Evidence

### Documented facts

- The v2 config model defines named family groups and accepts group names in `changedAny`.
  See `internal/config/config.go` and `internal/config/validate.go`.
- The current family resolver obtains include and exclude patterns through `SourceInclude` and
  `SourceExclude`, then stores matches in `SourceFiles`. See `internal/family/resolver.go`.
- The rule evaluator treats a non-empty `changedAny` condition as a check that `SourceFiles` is
  non-empty; it does not select files by the configured group names. See
  `internal/rules/evaluator.go`.
- Validation accepts `changedAny` and `changedStatusAny` in assertions, but the assertion evaluator
  currently evaluates only kin existence/change predicates.
- Each named kin currently resolves to one exact path because `KinMap` is `map[string]string`.
- Git collection retains added and modified paths, retains only the destination of a rename, and
  excludes deleted paths from evaluation. See `internal/changes/git.go`.
- The `web-ui` preset requires a story change only when the story already exists. Components without
  a story do not fail that rule. See `internal/cli/init.go`.

### Observations

- The current source-to-existing-story policy is usable today and represents the primary product
  use case.
- Named non-source groups do not yet provide distinct runtime behavior even though the configuration
  suggests that they do.
- Exact single-path kin work for one naming convention at a time, but common repositories may use
  different story extensions or locations.
- Lifecycle rules for deletions and rename pairs cannot be expressed from the current flattened
  change metadata.

### Inferences

- Completing group-aware evaluation is a correctness and contract-clarity issue, not merely an
  optional new predicate.
- Candidate-path and lifecycle support would deepen the original related-file use case without
  broadening Kyn into an integration platform.
- Explicit family identity is likely necessary for config bundles and many-to-one relationships,
  but it introduces more design risk than the Storybook-focused capabilities.

### Unknowns

- Whether family instances should remain source-anchored or be discoverable from changes in any
  named group.
- Which Storybook filename and directory conventions should be first-class examples.
- Whether candidate kin belong in the family schema or should be composed from additional rule
  predicates.
- How much lifecycle information manual input modes should support.
- Whether exceptions and relationship discovery are valuable enough to justify additional config
  and CLI surface area.

## Candidate 1: True Group-Aware Change Tracking

### Problem

Named groups currently describe intent but are not preserved in resolved family instances. As a
result, `changedAny: [source]`, `changedAny: [story]`, and `changedAny: [tests]` do not yet select
different changed files at evaluation time.

### Hypothesis

Retaining changed files and statuses per group will make the existing rule language truthful and
will provide the foundation for bidirectional and lifecycle-aware policies.

An internal family instance might conceptually retain:

```text
changedByGroup:
  source: [src/Button.tsx]
  story: [src/Button.stories.tsx]
  tests: []
```

The existing config could then behave literally:

```yaml
if:
  changedAny: [source]
assert:
  changedAny: [story]
```

The example above is not approval to make `changedAny` valid in assertions; the research must first
decide whether group assertions should complement or replace kin-change assertions.

The same investigation must explicitly dispose of assertion-side `changedStatusAny`. It should
either receive defined, group-aware assertion semantics, be rejected during config validation, or
be deprecated with a migration path. Continuing to accept it while evaluating no behavior is not a
valid outcome.

### Options to compare

1. **Source-anchored instances:** resolve an instance only from source changes, then classify other
   changed paths into that instance. This is simpler and preserves the current one-way mental model.
2. **All-group instances:** resolve an instance from a change in any group. This enables story-to-source
   and other bidirectional policies but requires group-specific identity normalization.

### Decision criteria

- A group name has one precise meaning in conditions, assertions, explanations, and reports.
- Existing source-to-kin configurations continue to behave predictably.
- Resolution and output remain deterministic.
- Story-only and test-only changes have explicitly documented behavior.
- Invalid or ambiguous group identities produce actionable config errors.
- Assertion-side `changedAny` and `changedStatusAny` are either evaluated with documented semantics
  or rejected by validation, with focused tests for the selected behavior.

### Research needed

- Build fixtures for source-only, story-only, source-plus-story, and conflicting group matches.
- Determine whether group-specific suffix stripping is required for all-group resolution.
- Decide whether reports expose group membership and whether doing so changes machine schemas.
- Review migration behavior for v2 configs that currently name non-source groups.
- Trace assertion-side `changedAny` and `changedStatusAny` through validation, migration, evaluation,
  explanation, and tests before selecting whether to implement or reject them.

## Candidate 2: Story Creation and Existing-Story Policies

### Problem

The current preset protects an existing story but silently permits a new component with no story.
Teams may want either policy, so Kyn should make the distinction obvious.

### Hypothesis

The initial improvement may require only better preset rules and documentation rather than new
engine behavior.

Illustrative rules using current capabilities:

```yaml
- id: existing-story-sync
  family: web-component
  severity: error
  if:
    changedAny: [source]
    kinExists: [story]
  assert:
    kinChanged: [story]
  message: "Source changed but its existing story did not."

- id: new-component-needs-story
  family: web-component
  severity: error
  if:
    changedAny: [source]
    changedStatusAny: [added]
  assert:
    kinExists: [story]
    kinChanged: [story]
  message: "New component must include a Storybook story."
```

### Decision criteria

- The default preset states whether stories are optional or required.
- Existing adopters can choose strict creation enforcement without duplicating an entire preset.
- A newly added secondary source file does not incorrectly make an existing component appear new.
- `kyn explain` clearly distinguishes a skipped optional-story rule from a failed required-story rule.

### Research needed

- Test multi-file component families where only one source file is added.
- Compare optional, warn-first, and required defaults for a new `web-ui` configuration.
- Decide whether this is a preset-only improvement or exposes a status/group precision gap.

## Candidate 3: Candidate Kin and Relationship Cardinality

### Problem

A kin currently resolves to one exact path. Storybook projects commonly need to accept more than one
valid extension or location. Current list predicates can already require every separately named kin,
but they cannot model a named candidate set in which any one acceptable path satisfies the policy.

### Hypothesis

First-class `anyOf` candidate relationships may express these cases more clearly than duplicating
families or rules. A new `allOf` form would not add basic all-required behavior—current list
predicates already provide that across named kin—and should be considered only if nested candidate
sets or clearer diagnostics justify the additional schema.

One possible family-oriented syntax is:

```yaml
kin:
  story:
    anyOf:
      - "{dir}/{base}.stories.tsx"
      - "{dir}/{base}.stories.ts"
      - "{dir}/{base}/index.stories.tsx"
```

Another option is to retain scalar kin and add rule-level predicates across several names:

```yaml
kin:
  storyTsx: "{dir}/{base}.stories.tsx"
  storyTs: "{dir}/{base}.stories.ts"
assert:
  kinChangedAny: [storyTsx, storyTs]
```

Neither syntax is selected by this document.

### Decision criteria

- Users can express "one acceptable story changed" without weakening the existing ability to
  require all separately named generated outputs.
- Diagnostics identify the accepted, missing, and expected candidate paths without noisy output.
- Existing scalar kin remains valid or has a straightforward migration.
- Candidate resolution is deterministic when multiple acceptable paths exist.
- File existence and change-set membership have unambiguous semantics for `anyOf`.

### Research needed

- Collect representative Storybook layouts from official framework examples and maintained sample
  repositories.
- Compare family-level cardinality with rule-level `Any`/`All` predicates.
- Establish whether an explicit `allOf` form adds value beyond current all-listed predicate semantics.
- Define behavior when two candidate story files both exist but only one changes.
- Determine how candidates appear in text, JSON, SARIF, rdjson, and checkstyle reports.

## Candidate 4: Lifecycle Synchronization

### Problem

Related files should often follow source files through addition, modification, rename, and deletion.
The current change model cannot validate complete rename pairs or deletions.

### Hypothesis

Preserving richer change records will support policies such as:

- added component requires an added story;
- renamed component requires its story to be renamed or otherwise updated;
- deleted component requires its story to be deleted;
- modified component requires its story to participate in the change.

Potential rule concepts include `kinStatusAny`, `kinStatusMatches`, or lifecycle-specific assertions.
Syntax should be selected only after the underlying change model is agreed.

### Decision criteria

- Rename records preserve old and new slash-normalized paths.
- Deleted paths are evaluated without relying on their existence in the head worktree.
- Git and manual input modes have explicitly documented capability differences.
- Lifecycle assertions report the observed and expected status.
- Existing flattened-file consumers and report formats remain compatible or receive a deliberate
  versioned change.

### Research needed

- Model add, modify, delete, and rename records through collection, resolution, evaluation, and output.
- Test rename detection thresholds and cases Git reports as delete-plus-add.
- Decide whether manual input accepts a name-status format in addition to the current path list.
- Define behavior for deleted source paths whose old directory or basename drives kin resolution.

## Candidate 5: Explicit Family Identity

### Problem

Directory plus basename is a good implicit identity for components, but not for every related-file
contract. Several configuration fragments may share one generated output, and module-level docs may
relate to many source files in one directory.

### Hypothesis

An explicit instance key could support sibling configs and many-to-one relationships without adding
repository graph adapters.

Illustrative concept:

```yaml
families:
  - id: service-config
    instanceKey: "{dir}"
    groups:
      source:
        include: ["services/**/config/*.yaml"]
    kin:
      example: "{dir}/example.yaml"
      docs: "{dir}/README.md"
```

### Options to compare

1. A general `instanceKey` template.
2. A constrained enum such as `groupBy: base | directory`.
3. Explicit anchor patterns that identify one canonical file per family.

### Decision criteria

- Multiple source paths can intentionally resolve to one stable family instance.
- Accidental identity collisions are detected or clearly explained.
- The feature covers config bundles, module docs, and generated artifacts with understandable YAML.
- The design does not become a monorepo project-graph language.

### Research needed

- Create fixtures for one-to-one, many-to-one, and one-to-many relationships.
- Compare the expressiveness and validation complexity of the three identity options.
- Determine how template variables behave after several files merge into one instance.
- Measure whether identity configuration remains approachable in `kyn init` and `kyn explain`.

## Candidate 6: Auditable Exceptions

### Problem

Strict related-file rules can be difficult to adopt in legacy areas. Plain excludes are easy to add
but do not communicate which rule is waived or why.

### Hypothesis

Rule-scoped exceptions with required reasons may improve adoption while keeping policy debt visible.

Illustrative concept:

```yaml
exceptions:
  - rule: story-sync
    match: "src/legacy/**"
    reason: "Legacy components are being migrated separately."
```

An expiry date is intentionally not assumed. Reading the wall clock during evaluation could violate
Kyn's deterministic behavior unless time becomes an explicit input.

### Decision criteria

- An exception identifies its rule, scope, and reason.
- Applied exceptions are visible in explain or summary output.
- Output for the same config and change set is deterministic.
- Exceptions remain easier to audit than general exclude patterns.

### Research needed

- Determine whether existing family excludes already cover enough of the need.
- Compare config exceptions with repository-owned ignore files.
- Decide whether expiry belongs outside Kyn, is an explicit CLI input, or should not be supported.
- Establish whether exception output is informational, warning-level, or opt-in.

## Candidate 7: Relationship Discovery

### Problem

Users may understand the policy they want but struggle to encode repository-specific suffixes,
locations, and identities.

### Hypothesis

A deterministic, read-only `kyn suggest` command could inspect repository paths and produce candidate
families or config on stdout without turning Kyn into an autofix system.

Possible output could include:

- common component-to-story and source-to-test suffixes;
- components with and without likely stories;
- ambiguous relationships requiring user selection;
- a proposed config fragment.

### Decision criteria

- Suggestions are deterministic and sorted.
- The command does not edit files by default.
- Confidence and ambiguity are visible; suggestions are not presented as verified truth.
- Scanning remains fast and bounded on large repositories.
- The result materially improves time-to-first-use over `kyn init --preset` alone.

### Research needed

- Define a bounded file inventory and exclusion strategy.
- Test heuristics against small React, Angular, Vue, and non-UI fixtures.
- Compare a new command with an `init --detect` mode.
- Establish measurable precision and onboarding improvement thresholds.

## Preliminary Ordering

This ordering is a research recommendation, not a release commitment.

1. Investigate and correct group-aware runtime semantics.
2. Validate Storybook creation versus optional-existing-story preset behavior.
3. Compare candidate-kin/cardinality designs.
4. Extend change metadata only if lifecycle use cases justify the compatibility cost.
5. Investigate explicit identity after group and candidate semantics stabilize.
6. Consider exceptions and discovery only after the core relationship model is clear.

Group tracking is foundational. Candidate kin is the strongest immediately visible Storybook
improvement. Lifecycle and explicit identity add broader value but carry more data-model and
compatibility risk.

## Comparison Criteria

Each candidate should be scored consistently against:

1. Closeness to Kyn's related-file mission.
2. Value to the Storybook flagship use case.
3. Reuse across tests, configs, schemas, generated files, and docs.
4. Deterministic behavior and diagnostic clarity.
5. Configuration simplicity and migration cost.
6. Implementation and report-contract risk.
7. Ability to validate the feature with small, explicit fixtures.

A candidate should not advance merely because it is technically possible. It should solve a common,
observable policy gap more clearly than existing configuration can.

## Investigation Method

For each selected candidate:

1. Write one narrow decision question and name the decision owner.
2. Gather primary evidence, favoring official Storybook/framework documentation and reproducible
   repository fixtures.
3. Label conclusions as documented facts, observations, inferences, or unknowns.
4. Compare at least two credible designs using the criteria above.
5. Record what evidence would change the recommendation.
6. Produce a proposed specification only after the problem and design are reviewed.

No external repository should be modified during research. Any controlled experiment should record
its commit, environment, inputs, commands, and retained fixture or output path.

## Expected Implementation Surface

If approved, the candidates are likely to affect these existing areas:

- `internal/config`: schema, validation, and migration behavior;
- `internal/family`: identity, matching, and resolved group/kin metadata;
- `internal/changes`: status and rename/delete collection;
- `internal/rules`: condition and assertion semantics;
- `internal/report`: deterministic diagnostics and external report contracts;
- `internal/cli`: preset, explain, or discovery behavior;
- `docs`: user-facing configuration, migration, and examples.

This list identifies review scope only; it is not an implementation plan.

## Testing and Quality Expectations

Any approved implementation should use table-driven tests and small fixtures covering:

- group matching and ambiguous membership;
- path normalization and family identity;
- added, modified, renamed, and deleted status behavior as relevant;
- candidate-set pass/fail diagnostics, including `anyOf` and any separately justified `allOf` form;
- deterministic text and machine-readable output;
- config validation and migration errors;
- stable exit-code classification.

Repository verification commands remain:

```bash
test -z "$(gofmt -l cmd internal)"
make lint
go test ./...
go build ./cmd/kyn
make ci
```

## Boundaries

### Always

- Preserve deterministic ordering, slash-normalized relative paths, and stable exit codes.
- Keep rule semantics explicit and testable.
- Update user documentation and report fixtures when a public contract changes.
- Separate repository evidence, inference, and unverified product hypotheses.

### Ask first

- Change config schema or migration behavior.
- Version a machine-readable report schema.
- Introduce time, network access, or framework parsing into evaluation.
- Change how existing v2 configurations interpret named groups.

### Never within this feature set

- Add daemon or watch behavior.
- Add PR-provider comments or API integrations.
- Add a plugin system or monorepo graph adapter.
- Automatically publish Storybook or create related files.
- Claim semantic correctness merely because two files changed together.

## Review Gate

Before any candidate advances to planning, reviewers should confirm:

- the problem is real and cannot be expressed clearly today;
- its acceptance criteria are measurable;
- the proposed scope respects Kyn's product boundaries;
- compatibility and report impacts are understood;
- unresolved questions have an owner and a bounded research path;
- the feature has a reviewed specification separate from this exploration document.

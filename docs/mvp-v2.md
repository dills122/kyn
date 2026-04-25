# Kyn MVP v2 Proposal

## Status

Draft design document for breaking-change v2 before first public release.

## Goal

Improve usability, reduce semantic ambiguity, and expand policy power while keeping Kyn a fast, stateless CLI.

## Required v2 Outcome

For v2 to be considered complete:

1. A new user can get Kyn running in 10 minutes or less.
2. CI integration works across major CI platforms with first-party examples.
3. Release artifacts are published for major Linux/macOS/Windows targets.
4. Breaking schema changes include a practical migration path.

## Why v2 Now

No public release has been made yet, so this is the right time to:

- simplify the rule model
- introduce stronger change metadata
- make CLI defaults friendlier
- keep future extensibility clean

## Design Principles

1. Rules should read like policy, not implementation detail.
2. Validation and side effects should be separated.
3. CI output must stay deterministic and machine-consumable.
4. Defaults should optimize for common CI usage, with strict modes available.

## Breaking Changes

## 1) Rule Model Rename: `when/require` -> `if/assert`

Old:

```yaml
when:
  changedAny: [source]
require:
  kinChanged: [story]
```

New:

```yaml
if:
  changedAny: [source]
assert:
  kinChanged: [story]
```

Rationale:

- `if` clearly means applicability.
- `assert` clearly means pass/fail policy checks.

## 2) `emitFlag` moves out of assertions into `actions`

Old:

```yaml
require:
  emitFlag: figmaPublishRequired
```

New:

```yaml
actions:
  emit:
    - figmaPublishRequired
```

Rationale:

- Assertions should only validate.
- Emission is a side effect.

## 3) First-class change status support

v2 tracks and exposes change status (`added`, `modified`, `deleted`, `renamed`) instead of flattening away details.

New filter examples:

```yaml
if:
  changedStatusAny: [modified, renamed]
```

## 4) Family groups become explicit

Families can define named groups to reduce hardcoded assumptions like `source`.

```yaml
families:
  - id: angular-component
    groups:
      source:
        include: ["libs/**/*.component.ts", "libs/**/*.component.html"]
      test:
        include: ["libs/**/*.spec.ts"]
      story:
        include: ["libs/**/*.stories.ts"]
```

## 5) CLI input behavior update

v2 adds optional auto mode:

- If no change mode is passed and git context exists, default to configured git refs.
- Keep strict enforcement available via `--strict-input-mode`.

This ergonomics behavior is required for v2, not optional.

## Proposed v2 Config Schema

```yaml
version: 2

families:
  - id: angular-component
    groups:
      source:
        include:
          - "libs/**/*.component.ts"
          - "libs/**/*.component.html"
      story:
        include:
          - "libs/**/*.stories.ts"
    baseName:
      stripSuffixes:
        - ".component"
    kin:
      story: "{dir}/{base}.stories.ts"
      figma: "{dir}/figma.{base}.json"

rules:
  - id: story-sync
    family: angular-component
    severity: error
    if:
      changedAny: [source]
      kinExists: [story]
      changedStatusAny: [modified, renamed]
    assert:
      kinChanged: [story]
    message: "Component changed but its Storybook file did not change."

  - id: figma-publish-signal
    family: angular-component
    severity: warn
    if:
      changedAny: [source]
      kinExists: [figma]
    actions:
      emit:
        - figmaPublishRequired
    message: "Component changed and has Figma metadata."
```

## Rule Semantics (v2)

- `if`: `AND` across keys.
- `assert`: `AND` across keys.
- List predicates require all listed kin names unless explicitly `Any`.
- `actions` never fail the run on their own.

## CLI Contract (v2 Direction)

Keep:

- `kyn check`
- `--config`, `--format`, `--fail-on`, `--fail-on-empty`

Keep input modes:

- `--files`
- `--files-from`
- `--stdin`
- `--base --head`

Add:

- `--strict-input-mode` (enforce explicit single mode)
- `--summary-only`
- `--dry-run-resolve` (show resolved families/kin without policy eval)

Planned commands:

- `kyn explain` (detailed per-rule/per-family reasoning)
- `kyn init` (bootstrap config from repo patterns)

## Output & Integrations

v2 targets additional machine reporters:

- SARIF
- reviewdog rdjson
- checkstyle XML

Text output remains deterministic, failures-first, and concise by default.

## Ergonomics & Portability Requirements (Required)

v2 must satisfy the contract in [ergonomics.md](ergonomics.md), including:

- 10-minute time-to-first-pass onboarding target
- clear happy-path CLI experience
- strong error usability
- multi-platform CI compatibility targets

## Migration Plan (v1 -> v2)

1. Add versioned config parser for `version: 1` and `version: 2`.
2. Add migration helper:
   - `kyn config migrate --from v1 --to v2`
3. Auto-map:
   - `when` -> `if`
   - `require` -> `assert`
   - `require.emitFlag` -> `actions.emit`
4. Warn on v1 usage with deprecation notice before v1 removal.

## Implementation Roadmap

## Phase 1: Schema & Engine Core

1. Implement v2 config structs + validation.
2. Implement dual-parser support (`version: 1` and `version: 2`).
3. Refactor evaluator to support `if/assert/actions`.
4. Add change status model end-to-end.

## Phase 2: CLI & UX

1. Add `--strict-input-mode`.
2. Add auto mode defaults for CI-friendly runs.
3. Add `--summary-only` and `--dry-run-resolve`.
4. Improve error messages with mode and rule/family context.

## Phase 3: Tooling & Adoption

1. Add `kyn explain`.
2. Add `kyn init`.
3. Add output adapters (SARIF/reviewdog/checkstyle).
4. Add preset rule packs for common domains.

## Success Criteria

- Teams can express policy with less config ambiguity.
- Fewer false positives from status-aware evaluation.
- Faster adoption through init/explain and better defaults.
- Strong CI integration via richer output adapters.
- New-user setup and first passing run can be achieved in <= 10 minutes.

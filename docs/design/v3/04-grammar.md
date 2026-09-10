# G3 — v3 grammar

Status: proposed. Closes gate G3 with
[`05-compatibility-matrix.md`](05-compatibility-matrix.md).

Follows from the semantics frozen in [`02-semantics.md`](02-semantics.md) and
the instance model in [`03-instance-and-identity.md`](03-instance-and-identity.md).
Nothing here is safe to implement before the G5 conformance suite exists.

## Common form

```yaml
version: 3

rules:
  handler-test:
    match: ["internal/**/*_handler.go"]
    related: "{dir}/{name}_test.go"
    expect: in-change-set
```

> When a path matching `internal/**/*_handler.go` changes, its related test path
> must be part of the change set.

No pattern, no family, no group, no named kin — the proposal's success
criterion 1.

Note what changed from the proposal's version of this example: `expect:
changed-if-present` became `expect: in-change-set`, with **no** `when`. Per D6
the rule now fires unconditionally rather than silently skipping when the test
does not exist. A user who wants the lenient behavior writes it down:

```yaml
    when: related-existed
    expect: in-change-set
```

## Reuse form

```yaml
version: 3

patterns:
  web-component:
    match:
      - "src/**/*.component.ts"
      - "src/**/*.component.html"
    exclude: ["src/**/generated/**"]
    stripSuffixes: [".component"]

rules:
  story-sync:
    use: web-component
    related: "{dir}/{base}.stories.ts"
    when: related-existed
    expect: in-change-set

  test-sync:
    use: web-component
    related: "{dir}/{base}.spec.ts"
    expect: [exists, in-change-set]
    severity: warn
    message: "Include the component test in this change."
```

Both rules share one source shape, so they share instances and reproduce the
measured cardinality contract (OS9): two rules × two instances = four results,
grouped by shape and sorted `shapeId`, `instanceName`, `ruleId`.

## Field reference

### `patterns.<id>` — a named source shape

| Field | Required | Notes |
| --- | --- | --- |
| `match` | yes | ≥1 glob |
| `exclude` | no | |
| `stripSuffixes` | no | feeds the instance key, not just the template |

A pattern is a named bundle of exactly these three fields and carries no
capability a rule lacks (D12).

### `rules.<id>`

| Field | Required | Default | Notes |
| --- | --- | --- | --- |
| `match` / `exclude` / `stripSuffixes` | one of these or `use` | | inline source shape |
| `use` | one of these or inline | | pattern reference |
| `related` | yes | | path template |
| `expect` | yes | | assertion, or AND-combined list |
| `when` | no | `always` | applicability gate |
| `on` | no | all statuses | source change statuses |
| `severity` | no | `error` | `info` \| `warn` \| `error` |
| `message` | no | generated (D7) | |
| `emit` | no | | informational flags; flat, replacing `actions.emit` |
| `description` | no | | see below |

`when` ∈ `always` \| `related-existed` \| `related-absent`.
`expect` ∈ `in-change-set` \| `not-in-change-set` \| `exists` \| `missing`.

Dropped from v2 with reason:

- **`changedAny`** — OS6 proved `if.changedAny: [source]` is a no-op, and
  validation already rejects every other group name. A source change is the
  implicit trigger.
- **`groups`** — only `groups.source` was ever read (OS6). It becomes the rule's
  or pattern's `match`/`exclude`.
- **`families`, `kin`** — replaced by the source shape plus `related`.

### `description` survives, and becomes functional

OS10 found `description` is accepted, validated and never read; SARIF fills both
its description slots with the message instead.

Keep the field and wire it into SARIF `fullDescription` and `explain`. The
original argument was compatibility — dropping it would break configs that set
it. With no such configs the field is kept on merit instead: a policy file
benefits from recording *why* a rule exists, separately from the message shown
when it fails, and SARIF already has the slot for it.

## Validation rules

Revised from the proposal's list, with the measured results folded in.

| # | Rule | Source |
| --- | --- | --- |
| 1 | `use` and inline `match`/`exclude`/`stripSuffixes` are mutually exclusive | proposal |
| 2 | Exactly one pattern reference per rule | proposal |
| 3 | Patterns cannot reference or extend patterns | proposal |
| 4 | A rule cannot override any field of the pattern it uses | proposal |
| 5 | Unknown pattern reference is an error | proposal |
| 6 | Duplicate rule or pattern IDs — **already free** | D3; `yaml.v3` rejects duplicate mapping keys with line numbers |
| 7 | An unused pattern is an **error**, not a warning | OS6: v2's lesson is that inert-but-valid constructs mislead for years |
| 8 | Empty match lists, path traversal, invalid template variables, multi-extension template disagreement all retain strict validation | proposal, `internal/family/resolver.go:117` |
| 9 | Rule and pattern IDs sorted before every normalization, validation and resolution pass | D1; OS7 shows the unsorted version is already non-deterministic |
| 10 | **New** — a resolved `related` path matching its own rule's source shape is an error, exit 2 | D8, OS5 |
| 11 | **New** — `expect` lists are deduplicated and sorted during normalization | D1 |

Rule 9 is not a style preference. OS7 measured the current unsorted behavior
producing three different error messages across thirty identical runs.

## Known capability limit

v3.0 deliberately cannot express a rule whose applicability or assertions span
**several** related paths atomically. The v2 shape:

```yaml
if:     { kinExists: [story, spec] }
assert: { kinChanged: [story, spec] }
actions: { emit: [reviewRequired] }
```

skips as one unit when either path is absent, and emits once. Two v3 rules
evaluate independently, so counts and flags differ.

Originally this had an escape hatch — such configs stayed on v2. The
[scope change](00-workplan.md#scope-change-2026-09-10) removed it, so this is now
a real capability limit rather than a migration boundary.

Accepted on evidence: **no config uses it.** Every kin clause across the four
presets, `docs/site/recipes/*`, and `docs/site/config.md` names exactly one kin.
The only multi-path syntax anywhere in the repository is `kinChangedAny` in
`docs/related-file-policy-exploration.md`, which is explicitly unapproved and was
never implemented.

Reserve the space rather than filling it: `expect` is already list-shaped, so
adding a list-shaped `related` later is additive and does not break the grammar.
Do not add it in v3.0.

## Naming (CR6)

`version: 3` is a **configuration schema** version and is unrelated to the
product version, which is `v0.1.x`. Once the product reaches 1.0 the phrase
"kyn v3" becomes ambiguous.

Convention: the file says `version: 3`; prose calls it **the rule-centric
format**. Reserve "v1/v2/v3" for the `version:` key alone, and say "Kyn 0.1.4"
for releases.

Keep the number at `3` even though v1 and v2 are being deleted. Restarting at
`1` would collide with the original v1 in git history, in `docs/decisions.md`,
and in `docs/migration-v1-to-v2.md`, and would make `version: 1` ambiguous
forever. A gap in the sequence costs nothing.

## Proposal open questions — answers

| # | Question | Answer |
| --- | --- | --- |
| 1 | `patterns`/`use` or `sources`/`source`? | `patterns`/`use`. `source` is already overloaded — `groups.source`, `changedAny: [source]`, `SourceFiles` — and reusing it would collide with report vocabulary. |
| 2 | `rules` as mapping or sequence? | Mapping (D3). Duplicate detection and line numbers are already free. |
| 3 | One related path per rule? | Yes. `expect` takes a list of assertions; multiple related *paths* means multiple rules. Atomic multi-path rules are a known capability limit, accepted because no config uses one. |
| 4 | Which expectation names? | D6: `in-change-set`, `not-in-change-set`, `exists`, `missing`, gated by `when`. |
| 5 | Unused patterns — error, warning, or allowed? | Error (validation rule 7). |
| 6 | Full v2 parity before release? | No. v1 and v2 are deleted rather than deprecated, and the one gap — atomic multi-path rules — is unused. |
| 7 | What expanded form handles emit, multiple kin, advanced applicability? | None in v3.0. `emit` is flat and per-rule; multi-path is reserved, not built. |
| 8 | Generated messages by default? | Yes, constrained by D7. |
| 9 | Deprecation window; v1→v3 direct? | Neither. v1 and v2 are removed outright; the four presets are rewritten by hand. |
| 10 | Reserve `imports`? | Omit entirely. Reserving a key with no semantics invites the same inert-construct problem OS6 and OS10 document. |

## JSON Schema

Approved as a deliverable, to be written once this grammar freezes.

It buys editor completion and inline validation for a format whose whole premise
is that the common case should be obvious to write. Two constraints:

- **Generated from the same source as the loader**, or checked against it in CI.
  A schema that drifts from the validator is worse than none, because it
  green-lights configs the binary rejects.
- **It cannot express every rule.** Validation rules 7 and 10 — unused pattern,
  self-match — are cross-field and resolve-time respectively. The schema covers
  shape and enums; the loader remains the authority.

## What G3 does not settle

- The conformance suite — gate G5.

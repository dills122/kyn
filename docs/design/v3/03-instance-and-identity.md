# G2 — Instance and identity model

Status: proposed. Closes gate G2 in [`00-workplan.md`](00-workplan.md).

Resolves IR1, IR2, CR1, CR5 and the OS5 item carried over from G1. Builds on the
semantics frozen in [`02-semantics.md`](02-semantics.md).

## 1. The contract this must preserve

Source: [`e10-shared-instance.sh`](experiments/e10-shared-instance.sh)

One v2 family, two rules, three changed source files where two collapse into one
instance:

```text
[web-component] src/a     sources: a.component.html, a.component.ts
[web-component] src/b     sources: b.component.ts
```

produces **four** results — rules × instances — each carrying the same
`familyId`, sorted by `familyId`, then `familyName`, then `ruleId`:

| ruleId | familyId | familyName |
| --- | --- | --- |
| `spec-sync` | `web-component` | `src/a` |
| `story-sync` | `web-component` | `src/a` |
| `spec-sync` | `web-component` | `src/b` |
| `story-sync` | `web-component` | `src/b` |

Any v3 model must reproduce these counts, this grouping, and this ordering.

## 2. Instances belong to a source shape, not to a rule

The decisive question for IR2 is what the instance key is keyed *on*, now that
families are gone.

Keying instances on the **rule** breaks the contract above. Each rule would
resolve its own instances, `familyId` would become `story-sync` / `spec-sync`,
and results would group by rule instead of by relationship — changing the sort
order of every report.

So the normalized model needs an entity that survives when families do not. It
already exists implicitly: the triple that decides which files match and how
their name is normalized.

> **Source shape** = (`match`, `exclude`, `stripSuffixes`)

A `pattern` is a *named* source shape. An inline rule carries an *anonymous* one
that belongs to exactly one rule.

```text
sourceShapeID = pattern name        (when the rule uses `use:`)
              = rule ID             (when the rule declares match/exclude inline)

instanceKey   = sourceShapeID + "|" + instanceName
instanceName  = {dir}/{base}        (unchanged from v2)
```

This maps a v2 family onto a v3 pattern one-for-one, so `familyId` →
`sourceShapeID` and `familyName` → `instanceName` preserve §1 exactly.

## 3. `stripSuffixes` moves to the rule (CR1, OS4)

OS4 proved `stripSuffixes` decides instance cardinality, not just how a template
renders. The proposal confined it to `patterns`, which left the inline form
unable to express the shipped `api` preset at all.

Decision: **`match`, `exclude` and `stripSuffixes` are all legal on a rule
inline, and a pattern is a named bundle of exactly those three.**

```yaml
rules:
  api-tests-sync:
    match: ["internal/**/*_handler.go", "internal/**/*_service.go"]
    stripSuffixes: ["_handler", "_service"]
    related: "{dir}/{base}_test.go"
    expect: in-change-set
```

Consequences:

- Patterns carry no capability a rule lacks. They are pure de-duplication, which
  is what the proposal's principle 3 asked for and what keeps them from becoming
  families under a new name.
- Inline and pattern-backed rules normalize to the identical structure, which is
  the proposal's success criterion 3 — now true by construction rather than by
  assertion.
- `use:` remains mutually exclusive with inline `match`/`exclude`/
  `stripSuffixes`, and pattern fields still cannot be overridden.

## 4. Aggregation rules

Carried forward from v2 unchanged, because they are already correct and tested:

| Concern | Rule |
| --- | --- |
| Source aggregation | All source files resolving to one instance key are collected and sorted. |
| Template agreement | If `related` resolves differently across source files in one instance, error. `{dir}` and `{base}` agree by construction; `{ext}`, `{file}` and `{name}` can disagree (`internal/family/resolver.go:117`). |
| Status applicability | `on:` matches if **any** source file in the instance carries an allowed status. |
| Evaluation cardinality | One result per (rule, instance). |

### Vanished source paths do not create instances

D5 retains deleted and renamed-away paths. It would now be *possible* to resolve
instances from a vanished **source** file, expressing "delete the handler, delete
its test".

Deferred deliberately. Creating instances from vanished sources would change
instance counts for every migrated policy, making it a second intentional v2→v3
divergence. [`02-semantics.md`](02-semantics.md) §4 commits to keeping that count
at one. Vanished paths stay available to the *gate* and to nothing else in v3.0.

## 5. Self-match is rejected at resolve time (OS5, D8)

Maintainer decision: reject rather than auto-exclude.

A fully static check is not sound — deciding whether a variable-substituted
template can intersect a glob set is glob-intersection in general. The exact
check is cheap at resolve time, where concrete paths exist:

> When a rule resolves a related path, test it against that rule's own source
> `match`/`exclude`. If it matches, fail with exit 2.

```text
rule "test-sync": related path "src/a_test.go" also matches this rule's own
match globs ["src/**/*.go"]; add an exclude such as "src/**/*_test.go", or
narrow the match pattern
```

Exit 2, because this is a configuration defect, not a policy failure — even
though it surfaces during resolution.

Tradeoff, stated plainly: the error only fires once a matching file actually
changes, so a latent self-match can sit in a config until it is exercised. This
is mitigated by `--dry-run-resolve`, which the tutorial already positions as the
step where you confirm a relationship before enforcing it. A conservative static
pre-check remains possible future work.

## 6. Report identity (IR1)

IR1 is the widest blocker because family identity reaches every output mode,
including RDJSON and Checkstyle as message prose.

With §2, the identity exists: `sourceShapeID` and `instanceName` are total for
v3 rules, inline ones included. So the question is only what goes on the wire.

**Decision: the wire format does not change.** `familyId` carries
`sourceShapeID`; `familyName` carries `instanceName`. Field names, types,
ordering and message strings are identical across v1, v2 and v3 configs.

The proposal itself supplies the rule: *"Never change machine report schemas or
exit behavior solely because the input config version changed."* Renaming
`familyId` to `sourceId` for v3 configs would break every consumer at the moment
a repository migrates — precisely the coupling that sentence forbids.

| Output | Field | v3 value |
| --- | --- | --- |
| `json` | `familyId` / `familyName` | `sourceShapeID` / `instanceName` |
| `text` | `Family:` / `Instance:` | same |
| `sarif` | `properties.familyId` / `familyName`, rule `properties.familyId` | same |
| `rdjson` | `message` suffix `(family instance: <name>)` | `instanceName` |
| `checkstyle` | `message` (reuses the RDJSON renderer) | `instanceName` |
| `--dry-run-resolve` | `instances[].familyId` / `.name` | same |
| `explain` | `familyId` / `familyName` | same |

### The `kin` map in `--dry-run-resolve`

`ResolveInstance.Kin` is `map[string]string` of kin name to resolved path
(`internal/report/dryrun.go:22`). A v3 rule has one `related` path and no kin
names.

Projection: the map keeps its shape, keyed by **rule ID**, since in v3 the rule
*is* the relationship. An instance shared by two rules lists both:

```text
[web-component] src/a
Kin:
  - spec-sync: src/a.spec.ts
  - story-sync: src/a.stories.ts
```

Structurally identical to the v2 output in §1, with rule IDs where kin names
were. For a migrated config the strings differ only if the user named their kin
differently from their rules — which the migration dry-run must show.

### Deferred: the "family" vocabulary

`(family instance: …)` in RDJSON and Checkstyle, and `Family:` in text output,
are v2 words that a v3 author never wrote. Changing them alters every message
string and breaks golden fixtures and downstream string matching.

Keep them. Revisit under a deliberate report-schema version bump, never as a
side effect of a config-version change.

## 7. `explain` clause naming (CR5)

`ClauseTrace.Clause` is a JSON string carrying values like `if.kinExists` and
`assert.kinChanged` (`internal/rules/explain.go:29`).

Distinguish two kinds of field:

- **Structural identity** — `familyId`, `familyName`, `ruleId`, `status`. Stable
  across config versions (§6).
- **Vocabulary that echoes the user's own config** — `clause`. Follows the
  version of the config the user wrote.

`clause` is the second kind. A v3 config's trace should name v3 fields:

```text
When:
  - when.related-existed: pass (existed at base (src/a_test.go))
Expect:
  - expect.in-change-set: fail (not in change set (src/a_test.go))
```

This is not a schema change: the field exists in both, is a string in both, and
its documented meaning — "which clause produced this result" — is unchanged.
Reporting `if.kinExists` for a config containing neither `if` nor `kinExists`
would be the actual defect.

The two-field grammar from D6 maps onto the existing `IfTrace` / `AssertTrace`
sections without inventing a third, so `ExplainResult` keeps its shape. Only the
section labels in **text** output change for v3 configs (`If:` → `When:`,
`Assert:` → `Expect:`); the JSON keys `if` and `assert` stay.

## 8. Normalized types

D1 governs: **ordered slices sorted by ID, never Go maps.** OS7 showed map
iteration already produces non-deterministic error selection; a v3 keyed by YAML
mappings would inherit that per rule and per pattern.

```go
// Version-neutral. v1, v2 and v3 documents all compile into this.
type Policy struct {
    Shapes []SourceShape   // sorted by ID
    Rules  []Rule          // sorted by ID
}

type SourceShape struct {
    ID            string   // pattern name, or the rule ID for an inline shape
    Inline        bool     // true when synthesized from a rule
    Match         []string
    Exclude       []string
    StripSuffixes []string
}

type Rule struct {
    ID       string
    ShapeID  string        // into Policy.Shapes
    Related  string        // template
    When     Gate          // always | related-existed | related-absent
    Expect   []Assertion   // AND-combined; sorted for determinism
    On       []changes.Status
    Severity string
    Message  string        // generated per D7 when empty
    Emit     []string
    Origin   Provenance    // config version, source position, pre-migration identity
}
```

`Origin` is what keeps `explain` honest across versions and lets the G4 migration
dry-run report exactly what changed. It is normalization input, never report
output.

Instance resolution iterates `Shapes` in sorted order, so both error selection
and result ordering are total functions of the input.

## 9. What G2 does not settle

- The concrete v3 grammar and the v1/v2 compatibility matrix — gate G3.
- The migratable normal form — gate G4. §6's kin-map projection and §3's
  `stripSuffixes` move both feed it directly.
- Whether `description` survives into v3 — gate G3.
- A sound static self-match pre-check (§5) — future work, not a v3.0 blocker.

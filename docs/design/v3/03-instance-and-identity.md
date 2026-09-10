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

instanceKey   = sourceShapeID + "|" + <template footprint>   (see §4)
```

This maps a v2 family onto a v3 pattern one-for-one, so the grouping and
ordering measured in §1 survive. The second half of the key — what v2 hardcoded
as `{dir}/{base}` — is derived in §4 instead.

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

## 4. The instance key is derived from the template's variable footprint

Maintainer decision: deduplicate results that demand the same path (item 8).
OS13 shows the cleanest place to do that is the instance key itself, not a
post-hoc merge of results.

### The problem

The key is `{dir}/{base}`, derived from the **source** file. The demanded path
comes from the **template**. When the template ignores the per-file variables
the two disagree, and one file is demanded once per source file (OS13):

| `related` | instances | distinct paths demanded |
| --- | --- | --- |
| `{dir}/{name}_test.go` | 3 | 3 |
| `{dir}/README.md` | 3 | **1** |
| `CHANGELOG.md` | 3 | **1** |

### The rule

> A rule's instance key is the source shape plus **exactly the template
> variables its `related` uses**, in the fixed order `dir`, `base`, `name`,
> `ext`, `file`.

| `related` | footprint | key | iac/changelog outcome |
| --- | --- | --- | --- |
| `{dir}/{base}.spec.ts` | dir, base | `shape\|dir\|base` | unchanged from v2 |
| `{dir}/{name}_test.go` | dir, name | `shape\|dir\|name` | unchanged |
| `{dir}/README.md` | dir | `shape\|dir` | **1 instance per directory** |
| `CHANGELOG.md` | — | `shape` | **1 instance total** |

`instanceName` is the rendered key. For an empty footprint it is the resolved
related path, which is constant by definition.

### Why this rather than merging results

Three properties fall out for free.

1. **Deduplication is structural.** There is no merge step, so there is no
   question of what `instanceName`, `changedFiles` or `shapeId` become on a
   merged row. Sources aggregate into one instance the same way they always did.
2. **The template-agreement error becomes unreachable.** The guard at
   `internal/family/resolver.go:117` exists because a template using `{ext}`,
   `{file}` or `{name}` can resolve differently across files sharing a
   `{dir}/{base}` key. Under footprint keying those variables are *in* the key,
   so files in one instance necessarily agree. The check stays as an assertion,
   but it can no longer fire.
3. **OS11 stops being possible.** The `api` preset's contradiction —
   `stripSuffixes` collapsing handler and service while `{name}` splits them —
   resolves automatically: `{name}` is in the footprint, so they are separate
   instances, each with its own test. `stripSuffixes` now affects grouping only
   when the template actually uses `{base}`, which is the only case where it
   *should*.

Property 3 means the preset fix in [`06-cutover.md`](06-cutover.md) is belt and
braces rather than load-bearing — worth keeping, since a config should not
depend on a subtle key derivation to be coherent.

### Relationship to D11

This refines D11, it does not contradict it. The **shape** still decides which
files group together; the footprint decides how finely that group is
partitioned. Two rules over one pattern whose templates have the same footprint
partition identically and share instances exactly as OS9 measured. They diverge
only when their templates genuinely address different granularities — which is
the correct behavior, not a regression.

### Aggregation rules

| Concern | Rule |
| --- | --- |
| Source aggregation | All source files resolving to one instance key are collected and sorted. |
| Template agreement | Guaranteed by the key. Retained as an assertion that must never fire. |
| Status applicability | `on:` matches if **any** source file in the instance carries an allowed status. |
| Evaluation cardinality | One result per (rule, instance). |

### Vanished source paths do not create instances

D5 retains deleted and renamed-away paths. It would now be *possible* to resolve
instances from a vanished **source** file, expressing "delete the handler, delete
its test".

Deferred deliberately on scope grounds. A vanished source has no working-tree
file to derive `{ext}`, `{file}` or `{name}` from, so both template resolution
and the footprint key would need a second code path. Vanished paths stay
available to the *gate* and to nothing else in v3.0.

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

**Decision: rename the fields to say what they are.**

The original decision here was to keep `familyId` / `familyName` untouched,
because renaming would break every consumer at the moment a repository migrated.
The [scope change](00-workplan.md#scope-change-2026-09-10) removed that
constraint — there are no consumers — so the fields get honest names instead of
inherited ones.

| Old | New | Value |
| --- | --- | --- |
| `familyId` | `shapeId` | pattern name, or the rule ID for an inline shape |
| `familyName` | `instanceName` | `{dir}/{base}` |

| Output | Change |
| --- | --- |
| `json` | `familyId` → `shapeId`, `familyName` → `instanceName` |
| `text` | `Family:` → `Shape:`, `Instance:` unchanged |
| `sarif` | `properties.familyId` → `shapeId`, same for the rule-level property |
| `rdjson` | message suffix `(family instance: src/a)` → `(instance: src/a)` |
| `checkstyle` | inherits the RDJSON message renderer |
| `--dry-run-resolve` | `instances[].familyId` → `.shapeId`, `.name` → `.instanceName` |
| `explain` | same as `json` |

Sorting is unchanged in structure — `shapeId`, then `instanceName`, then
`ruleId` — so the OS9 ordering contract survives the rename intact.

For an inline rule `shapeId` equals the rule ID, which reads as redundant in
reports. That redundancy is deliberate: it keeps one grouping key across inline
and pattern-backed rules, so consumers do not need to branch on which form the
author used.

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

### Report headline (OS12)

`summarize` counts every failing result in `Failed` but only clears `OK` for
results at or above `--fail-on`
(`internal/rules/engine.go:32`), so a failing `severity: warn` rule prints
`PASS` directly above `Rules failed: 1` and exits 0. The `web-ui` preset's
`tests-sync` rule is `severity: warn`, so this is the common case.

The exit code is right — `warn` is not meant to block. The headline is not.
Three states, so three words:

| Condition | Headline | Exit |
| --- | --- | --- |
| `OK == false` | `FAIL` | 1 |
| `OK == true` and `Failed > 0` | `NON-BLOCKING` | 0 |
| `Failed == 0` | `PASS` | 0 |

`PASS` becomes a claim that nothing failed, which is what a reader scanning the
first line assumes it already means. `NON-BLOCKING` says the run did not block
without saying nothing happened.

This changes the first line of every text golden, so it must land before the G5
fixtures are recorded.

## 7. `explain` clause naming (CR5)

`ClauseTrace.Clause` is a JSON string carrying values like `if.kinExists` and
`assert.kinChanged` (`internal/rules/explain.go:29`).

With v1 and v2 retired there is only one vocabulary, so this reduces to naming
the clauses after the fields the user actually writes:

```text
When:
  - when.related-existed: pass (existed at base (src/a_test.go))
Expect:
  - expect.in-change-set: fail (not in change set (src/a_test.go))
```

`ExplainResult` keeps its shape — the two-field grammar from D6 maps onto the
existing `IfTrace` / `AssertTrace` slots without needing a third. The JSON keys
stay `if` and `assert` only if that reads honestly; since nothing depends on
them any more, rename them to `when` and `expect` alongside the §6 rename and
keep one vocabulary end to end.

The distinction D13 originally drew — structural identity stable across config
versions, user-facing vocabulary following the config version — is moot with a
single version. It is worth keeping as a principle for any future version bump.

## 8. Normalized types

D1 governs: **ordered slices sorted by ID, never Go maps.** OS7 showed map
iteration already produces non-deterministic error selection; a v3 keyed by YAML
mappings would inherit that per rule and per pattern.

```go
// The normalized policy. The v3 document compiles into this.
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
    Origin   Provenance    // source position: file, line, column
}
```

`Origin` carries YAML source position so validation errors can point at the
line the user wrote. It is normalization input, never report output. With v1 and
v2 retired it no longer needs to carry a config version or a pre-migration
identity.

Instance resolution iterates `Shapes` in sorted order, so both error selection
and result ordering are total functions of the input.

## 9. What G2 does not settle

- The concrete v3 grammar — gate G3.
- Nothing outstanding. The deduplication question from the workplan is answered
  by §4.
- A sound static self-match pre-check (§5) — future work, not a v3.0 blocker.

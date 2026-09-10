# G4 — Migration

Status: proposed. Closes gate G4 in [`00-workplan.md`](00-workplan.md).

Answers IR3. Depends on [`02-semantics.md`](02-semantics.md),
[`03-instance-and-identity.md`](03-instance-and-identity.md),
[`04-grammar.md`](04-grammar.md) and
[`05-compatibility-matrix.md`](05-compatibility-matrix.md).

IR3's objection: the proposal claims a v2 family with several kin and rules
becomes a pattern plus one rule per related path, but v2 permits AND-combined
multi-kin applicability, cross-kin clauses and atomic multi-assertion rules that
this cannot represent. The answer is a formal subset plus an explicit refusal
set — never an approximation.

## The migratable normal form

A v2 **rule** migrates if and only if all of the following hold.

| # | Condition |
| --- | --- |
| M1 | Every kin name it references — across `if` and `assert` — is the **same single** name. |
| M2 | Its `if` clauses use only `changedAny: [source]`, `changedStatusAny`, and at most one of `kinExists` / `kinMissing`. |
| M3 | Its `assert` clauses reference only that one kin name, in any combination of `kinChanged`, `kinUnchanged`, `kinExists`, `kinMissing`. |
| M4 | Its assertion combination maps onto the D6 grid — see the table below. |

A v2 **family** migrates if every rule referencing it migrates.

### Clause mapping

| v2 | v3 |
| --- | --- |
| `if.changedAny: [source]` | dropped — proven no-op (OS6) |
| `if.kinExists: [k]` | `when: related-existed` |
| `if.kinMissing: [k]` | `when: related-absent` |
| *(no `if`)* | `when: always` |
| `if.changedStatusAny: [...]` | `on: [...]` |
| `assert.kinChanged: [k]` | `expect: in-change-set` |
| `assert.kinUnchanged: [k]` | `expect: not-in-change-set` |
| `assert.kinExists: [k]` | `expect: exists` |
| `assert.kinMissing: [k]` | `expect: missing` |
| several of the above on the same `k` | `expect: [...]`, sorted |
| `actions.emit`, `require.emitFlag` | `emit: [...]` |
| `description`, `severity`, `message` | carried verbatim |

### Family mapping

| v2 | v3 |
| --- | --- |
| `groups.source.include` / `.exclude` | `match` / `exclude` |
| top-level `include` / `exclude` (matrix row 5) | `match` / `exclude` |
| `baseName.stripSuffixes` | `stripSuffixes` — on the rule or pattern (D12) |
| non-`source` groups | **dropped**, proven no-op (D2, OS6) |
| kin referenced by exactly one rule | that rule's `related` |
| kin referenced by no rule | **dropped** — see below |

A family used by one rule becomes an inline source shape. A family used by two
or more rules becomes a `pattern`, which is what preserves the shared-instance
contract (D11, OS9).

## The refusal set

Refused, not approximated. Each stays on v2.

| # | Shape | Why it cannot be represented |
| --- | --- | --- |
| R1 | `if` or `assert` naming **several** kin | v2 evaluates them atomically; two v3 rules change result counts and can emit when v2 would not |
| R2 | Cross-kin rules — `if` names kin A, `assert` names kin B | v3 has one `related` per rule |
| R3 | A rule whose kin template disagrees across its instance's source files | already fails at resolve time (OS11); migrating it would carry the defect forward silently |

R1 is IR3's example, and it is worth restating in full because it is the case an
approximate migration would get wrong:

```yaml
if:      { kinExists: [story, spec] }
assert:  { kinChanged: [story, spec] }
actions: { emit: [reviewRequired] }
```

If `story` exists and `spec` does not, v2 skips the whole rule and emits
nothing. Two `when: related-existed` v3 rules would evaluate the story rule,
potentially fail it, and emit the flag. Different exit code, different result
count, different flags.

## Coverage against the shipped presets

The honest test of a subset is whether it covers the configs Kyn itself
generates. All four `kyn init` presets migrate.

| Preset | v2 shape | v3 shape | Notes |
| --- | --- | --- | --- |
| `web-ui` | 1 family, 2 kin, 2 rules | 1 pattern + 2 rules | The family is shared, so it must become a pattern — see D11 |
| `api` | 1 family, 1 kin, 1 rule | 1 inline rule | Needs `stripSuffixes` inline, which is exactly why D12 moved it |
| `proto` | 1 family, 1 kin, 1 rule | 1 inline rule | |
| `iac` | 1 family, 1 kin, 1 rule | 1 inline rule | |

Two caveats the coverage table would otherwise hide:

**`api` must be fixed before it is migrated.** OS11 shows the shipped preset
contradicts itself — `stripSuffixes` collapses handler and service into one
instance while `{name}` resolves differently for each — and fails with exit 2 on
an ordinary change set. Migration preserves behavior, so migrating the broken
preset produces a broken v3 config. Fix the preset in G0 first.

**`iac` demands the same file twice.** Its `related` template
`{dir}/README.md` contains no `{base}`, so two `.tf` files in one directory
produce two instances that both demand the same README:

```text
[terraform-module] terraform/vpc/main       readme: terraform/vpc/README.md
[terraform-module] terraform/vpc/variables  readme: terraform/vpc/README.md
```

Two failures, one file. v3 preserves this exactly. Deduplicating demands in
report output is a reasonable future improvement, but it is a report change, not
a migration change, and must not ride along here.

## Unreferenced kin are a reportable loss

A kin no rule references still appears in `--dry-run-resolve` output. Under the
D10 projection, v3's kin map is keyed by rule ID, so an unreferenced kin has
nowhere to go.

This is the one place where dropping something is *not* a proven no-op: the
resolve report changes. Migration must list every dropped kin by name rather
than removing it silently, and the G5 harness must expect a dry-run difference
here — the only sanctioned one.

## Dry-run output

`kyn config migrate --from v2 --to v3 --dry-run` must show all five:

1. The normalized v3 rules.
2. Every default v3 inserted — `severity: error`, `when: always`, generated
   `message` (D7).
3. **The D5 gate divergence**, per rule: *this rule will additionally fail when
   the related file is deleted or renamed away*. This is the one intentional
   behavior change between a v2 policy and its v3 equivalent
   ([`02-semantics.md`](02-semantics.md) §4), and burying it would be the worst
   possible outcome of this whole design.
4. Everything dropped: non-`source` groups, `changedAny: [source]`,
   unreferenced kin.
5. Every refusal, with its R-number and the specific clause that triggered it.

## Deprecation policy

- **v2 is supported until parity exists.** The refusal set is non-empty, so v2
  cannot enter deprecation. This is the proposal's own rule that migration must
  never weaken a policy, applied to the product timeline.
- **v1 → v3 is a distinct path, not a shortcut through v2.** Matrix row 2
  measured that v1 rejects `groups` entirely, so a v1 family's top-level
  `include` maps straight to `match`. Routing v1 through the v2 migrator first
  is still the simpler implementation and produces the same result; the point is
  that "v1 is v2 with extra steps" is not something to assume.
- **No deprecation window is announced until the refusal set is empty** or a
  durable dual-support policy is published.

## What G4 does not settle

- The differential conformance harness that proves each migration equivalent —
  gate G5. Structural comparison is not enough; the harness must compare
  rendered output.
- Whether an expanded v3 form should eventually absorb R1 and R2. Deliberately
  out of scope for v3.0 ([`04-grammar.md`](04-grammar.md), Known non-parity).

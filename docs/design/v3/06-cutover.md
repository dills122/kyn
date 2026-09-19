# G4 — Cutover

Status: proposed. Closes gate G4 in [`00-workplan.md`](00-workplan.md).

> This document replaced `06-migration.md`. The
> [scope change](00-workplan.md#scope-change-2026-09-10) retires v1 and v2
> outright rather than migrating off them, so the gate changed from "prove every
> migration equivalent" to "delete the old versions and rewrite what depended on
> them". The migration analysis that is still load-bearing is preserved below.

## What IR3 asked, and what remains of it

IR3's objection was that the proposal claimed a v2 family becomes a pattern plus
one rule per related path, when v2 permits multi-kin applicability, cross-kin
clauses and atomic multi-assertion rules that this cannot represent.

The objection was correct, and the answer is no longer a migration subset. There
is nothing to migrate. What survives is the **capability question**: does v3
lose something v2 could express?

Yes, in exactly one place — atomic multi-path rules — and it is unused. Measured
across the four `kyn init` presets, `docs/site/recipes/*` and
`docs/site/config.md`, every kin clause names exactly one kin. The only
multi-path syntax anywhere in the repository is `kinChangedAny` in
`docs/related-file-policy-exploration.md`, which is explicitly unapproved and
was never implemented. See
[`04-grammar.md`](04-grammar.md#known-capability-limit).

## Removal list

| Component | Action |
| --- | --- |
| `internal/config` v1/v2 structs, aliases, `Validate` | Replace with a v3 document type plus the normalized `Policy` from [`03-instance-and-identity.md`](03-instance-and-identity.md) §8 |
| `internal/config/migrate.go`, `MigrateV1ToV2` | Delete |
| `internal/cli/config_migrate.go`, `kyn config migrate` | Delete |
| `internal/family` — `Family`, `KinMap`, `GroupMap` | Replace with source shapes and a single `related` template |
| `docs/migration-v1-to-v2.md` | Retire to historical, or delete |
| `docs/site/config.md` | Rewrite for the v3 grammar |
| `e2e/workflows_test.go:95` v1→v2 end-to-end test | Delete; its role is taken by the G5 suite |

Deleting `kyn config migrate` is a CLI surface removal, which the repository
steering guards. It is justified here only because the command's sole purpose
was moving between versions that no longer exist.

## Preset rewrite

All four presets become v3 by hand. Three are direct; one must be fixed first.

| Preset | v3 shape | Note |
| --- | --- | --- |
| `web-ui` | 1 pattern + 2 rules | Two rules share a shape, so the pattern earns its place — this is the example that justifies `patterns` existing at all |
| `api` | 1 inline rule | **Fix OS11 first**, see below |
| `proto` | 1 inline rule | |
| `iac` | 1 inline rule | Demands one file from many instances, see below |

### `api` must be fixed, not transcribed

OS11: the shipped preset sets `stripSuffixes: ["_handler", "_service"]` while
templating with `{name}`. The two contradict — `stripSuffixes` collapses handler
and service into one instance, `{name}` resolves differently for each — and an
ordinary change set fails with exit 2.

Transcribing it into v3 would carry the defect across. The evident intent is
that each handler and service has its own `_test.go`, so the v3 preset drops
`stripSuffixes` entirely:

```yaml
version: 3

rules:
  api-tests-sync:
    match: ["internal/**/*_handler.go", "internal/**/*_service.go"]
    related: "{dir}/{name}_test.go"
    when: related-existed
    expect: in-change-set
    message: "API source changed but its test did not."
```

This is also the clearest argument for D12: had `stripSuffixes` and the template
been adjacent in one rule rather than split across a family and its kin map, the
contradiction would have been visible when the preset was written.

### `iac` demands one file from many instances

Its template `{dir}/README.md` contains no `{base}`, so two `.tf` files in one
directory produce two instances that both demand the same README — two failures
for one file:

```text
[terraform-module] terraform/vpc/main       readme: terraform/vpc/README.md
[terraform-module] terraform/vpc/variables  readme: terraform/vpc/README.md
```

Previously this had to be preserved exactly, because migration could not change
behavior. That constraint is gone, so it becomes a live question: should reports
deduplicate results whose resolved `related` path is identical within a shape?

Deferred, not decided — it is a report-shaping question, tracked as an open item
in the workplan. Note that deduplicating changes result *counts*, so it must be
settled before the G5 goldens are recorded, not after.

## Starter-config policy

The presets are the first v3 most people will read, so they are specification by
example. Two rules for their rewrite:

1. **No inert constructs.** OS6 and OS10 measured three things v2 accepts that do
   nothing, two of which the presets taught. Nothing in a v3 preset may be
   decorative.
2. **`when` is written explicitly wherever the intent is lenient.** D6 removed
   the silent skip by defaulting `when: always`. A preset that wants
   `related-existed` must say so, since the preset is where users learn what the
   field means.

## What G4 does not settle

- The conformance suite — gate G5.
- Whether `--dry-run-resolve` deduplicates identical related paths.
- Whether `docs/migration-v1-to-v2.md` and `docs/decisions.md` are retired or
  annotated. Both describe behavior that will no longer exist, and
  `docs/decisions.md` in particular is cited throughout this design set as the
  source of the `D`-exclusion rule that D5 overturns.

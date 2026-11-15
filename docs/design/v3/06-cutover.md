# Cutover

Status: proposed. Kyn has no users yet, so v1 and v2 are removed outright
rather than migrated. There is no `v2 → v3` converter and no deprecation
window.

## Inventory

Measured by searching the repository, not listed from memory.

### Code

| Component | Action |
| --- | --- |
| `internal/config`: v1/v2 structs, aliases, `Validate`, `MigrateV1ToV2` | Replace with the v3 document and the normalized model in [`03-instances-and-reports.md`](03-instances-and-reports.md) §7 |
| `internal/family` | Replace: captures, template resolution, instance keys |
| `internal/matcher` | Extend: capture-aware globs (`{name}`, `{**name}`) |
| `internal/changes` | Retain deleted paths and rename sources; accept `--name-status` lines in `--files-from`; read git output with quoting disabled (`-c core.quotePath=false`, or `-z`), which also fixes OS16 for v2-era behavior |
| `internal/rules`, `internal/report` | The three states, `skipIf`, skipped status, headline, renamed fields |
| `internal/cli/config_migrate.go`, `kyn config migrate` | Delete. A CLI surface removal, justified because both ends of the migration are gone |
| `internal/cli/init.go` | Four presets rewritten; the `kin` and `families` wording in its help and next-steps text |
| `internal/cli/check.go`, `explain.go` | Help text and flag descriptions that say "family" and "kin": command summaries, `--fail-on-empty` ("no family instances match"), `--dry-run-resolve` ("Resolve families/kin only") |

### Fixtures and tests

| Item | Count |
| --- | --- |
| `e2e/projects/*/kyn.config.yaml` | 4 |
| `e2e/projects/*/testdata/*.golden` | 8 |
| `e2e/projects/*/scenarios.json` | 4 |
| `internal/report/testdata/*.golden` | 13 |
| `testdata/angular/kyn.config.yaml` | 1 |
| Inline v1/v2 configs in `internal/cli/execute_test.go` | 16 |
| Inline in `internal/cli/output_stream_test.go` | 2 |
| Inline in `internal/config/load_test.go` | 3 |
| `e2e/workflows_test.go` v1→v2 migration test | delete |
| Struct-based tests of the packages being replaced: `internal/config/validate_test.go`, `internal/config/migrate_test.go`, `internal/family/resolver_test.go`, `internal/rules/evaluator_test.go`, `internal/rules/explain_test.go` | rewritten or deleted with their packages |

Every golden changes. One example,
`e2e/projects/terraform-iac/testdata/module-changed-readme-not-updated-warns-only.golden`,
changes three ways at once: `PASS` becomes `NON-BLOCKING`, the `Family:` line
goes, and `Instance: terraform/network/main` becomes `terraform/network`. That
file currently records the OS12 defect as a contract fixture.

Goldens are external-contract fixtures under the repository steering.
Re-recording them is one deliberate, reviewed event, done once the report
decisions here are implemented, not incrementally.

### Documentation

| File | Action |
| --- | --- |
| `README.md`, `docs/site/config.md`, `docs/site/concepts.md`, `docs/site/troubleshooting.md`, `docs/site/recipes/*.md` | Rewrite for v3 |
| `docs/site/getting-started.md` | Rewrite for v3 (the v2 walkthrough is PR #49) |
| `docs/spec.md`, `docs/mvp-v2.md`, `docs/mvp-v2-tasks.md`, `docs/mvp-tasks.md`, `docs/related-file-policy-exploration.md` | Historical; add a superseded banner |
| `docs/migration-v1-to-v2.md`, `docs/decisions.md` | Already bannered |
| `docs/site/releases/v0.1.3.md` | Leave; it describes a past release |

## `e2e/projects` becomes the conformance corpus

`e2e/e2e_test.go` already walks each project, runs the real binary, and
compares rendered output and exit codes across output modes. That is the
harness [`07-conformance.md`](07-conformance.md) needs, and it is extended, not
duplicated.

It is not enough as it stands. Its projects are static directories driven by
`--files`, with no git support, and the conformance matrix needs repositories
with deletions, renames, and a checkout missing a file that exists on the head
ref. The extension is a small helper that builds those repositories from a
description, so each cell is data and not a hand-made directory. The four
existing projects are rewritten for v3 and keep their static form.

## Presets

All four are rewritten by hand.

```yaml
# web-ui
rules:
  story-sync:
    match: "src/**/{name}.component.{ts,html}"
    related: "{dir}/{name}.stories.ts"
    expect: touched
    skipIf: missing
    message: "Component changed but its story did not."
  tests-sync:
    match: "src/**/{name}.component.{ts,html}"
    related: "{dir}/{name}.spec.ts"
    expect: touched
    skipIf: missing
    severity: warn
    message: "Component changed but its test did not."
```

```yaml
# api
rules:
  api-tests-sync:
    match: "internal/**/{name}_{handler,service}.go"
    related: "{dir}/{name}_test.go"
    expect: touched
    skipIf: missing
```

The v2 `api` preset contradicts itself and fails with exit 2 when a handler and
a service change together (OS11). It is not transcribed. The v3 form above
states one reading plainly: a handler and service that share a name share one
test. The other reading, one test per file, is `match:
"internal/**/{name}.go"` with an exclude for `*_test.go`. The preset must pick
one and say which in a comment.

`proto` and `iac` are direct: `proto/**/{name}.proto` to
`gen/{dir}/{name}.pb.go`, and `terraform/**/*.tf` to `{dir}/README.md`.

Preset rules:

1. Nothing decorative. v2 accepted three constructs that do nothing (OS6,
   OS10), and the presets taught two of them.
2. `skipIf: missing` is written out wherever a preset is lenient, because the
   preset is where people learn what it means.
3. Each preset carries a one-line comment per rule saying what it enforces.

## Order of work

1. Fix OS7 (sorted iteration) and add the determinism guard. Everything else is
   tested on top of it.
2. Capture-aware matcher and instance keys.
3. Change collection: removals, rename sources, `--name-status` input.
4. Evaluation and reports.
5. Presets, e2e projects, goldens re-recorded once.
6. Documentation.

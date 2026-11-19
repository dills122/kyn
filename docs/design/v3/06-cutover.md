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
| `internal/matcher` | Replace doublestar with the v3 dialect in [`04-grammar.md`](04-grammar.md): captures, longest-first binding (with the rule that a capture may not sit between two `**`), code-point matching, and rejection of the constructs v2 accepted (`[ ]`, `\`, nested braces, wildcards in alternatives, partial-segment `**`, `.`/`..`/empty segments, whitespace-padded globs). Measured against this repository, no configuration uses a rejected form; `src/**` no longer matches a file named `src`. The reference matcher in the simulator is the oracle |
| `internal/changes` | Retain deleted paths and rename sources; accept `--name-status` lines in `--files-from`; read git with `git diff --name-status -z -M`, which prints every name literally (`-c core.quotePath=false` alone still quotes a name containing `"`), which also fixes OS16 for v2-era behavior; count git `T` as `modified` and touched; accept `--stdin` as `--files-from -`; drop plain-mode sources that are not on disk, but keep git-mode sources that are absent from disk; reject conflicting `--files-from` lines and non-clean input paths ([`02-semantics.md`](02-semantics.md) §6). **Compatibility break:** `NormalizePath` today trims whitespace, turns `\` into `/`, strips `./` and cleans the path; v3 rejects all of that with exit 2, for `--files` and `--files-from` alike. Blank lines and CRLF are tolerated. `--files` is comma-separated today, so a path containing a comma cannot be passed; use `--files-from` for it. Removing the cleaning is deliberate: a path Kyn rewrites silently is a path it may match differently than git does. The check is fatal for every changed path, matched or not: a repository that tracks a name with an edge blank, a backslash, a control or zero-width character, or a first segment like `Q:` (legal on Linux and macOS) fails every run whose change adds, modifies or renames to it with exit 2, naming the path. The same goes for a file whose name starts with a status token and a space (`M Notes.md`) when fed as a plain path, and for a name with a leading `"` given to `--files-from` or `--files` (git mode accepts it). Deleted paths and rename sources are exempt, so the change that fixes a bad name can pass. Names with a format character at a segment edge (a tag-sequence flag, a trailing right-to-left mark) are rejected too, as a consequence of the Cf rule No file in this repository is affected |
| `internal/rules`, `internal/report` | The three states, `skipIf`, skipped status, headline, renamed fields |
| `internal/cli/config_migrate.go`, `kyn config migrate` | Delete. A CLI surface removal, justified because both ends of the migration are gone |
| `internal/cli/init.go` | Four presets rewritten; the `kin` and `families` wording in its help and next-steps text |
| `internal/cli/check.go` | `--dry-run-resolve` returns before the empty check today, so `--fail-on-empty` exits 0 with zero matches: v3 applies it to the preview, prints per-rule counts, and has the headline say when nothing matched or everything skipped ([`03`](03-instances-and-reports.md) §8, §9) |
| `internal/rules`, `internal/changes`: the `removed` state | The change record decides it before the working tree is read ([`02`](02-semantics.md) §2), so a stale checkout cannot hide a deletion |
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
| `docs/site/ci.md` (the container job pipes `git diff --name-only` into `--stdin`), CLI help examples | Switch every enforcement example to `git diff --name-status -M ... | kyn check --stdin`; keep `--files` as the local quick start. Plain paths lose deletion protection ([`02`](02-semantics.md) §6) |
| `README.md`, `docs/site/config.md`, `docs/site/concepts.md`, `docs/site/troubleshooting.md`, `docs/site/recipes/*.md` | Rewrite for v3 |
| `docs/site/getting-started.md` | Rewrite for v3 (the v2 walkthrough is PR #49) |
| `docs/spec.md`, `docs/mvp-v2.md`, `docs/mvp-v2-tasks.md`, `docs/mvp-tasks.md`, `docs/related-file-policy-exploration.md` | Historical; add a superseded banner |
| `docs/migration-v1-to-v2.md`, `docs/decisions.md` | Bannered as slated for removal. The v1 to v2 guide stays the live workflow until the release that removes it ships |
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

The v2 `api` preset contradicts itself and fails with exit 2 when a handler and
a service change together (OS11). It is not transcribed. The v3 preset uses
**focused tests per source**, matching the published Go recipe: one narrow rule
for handlers and one for services, so `order_handler.go` needs
`order_handler_test.go` and `order_service.go` needs `order_service_test.go`.

```yaml
# api
rules:
  handler-test:
    match: "internal/**/{name}_handler.go"
    related: "{dir}/{name}_handler_test.go"
    expect: touched
    skipIf: missing
  service-test:
    match: "internal/**/{name}_service.go"
    related: "{dir}/{name}_service_test.go"
    expect: touched
    skipIf: missing
```

The shared-test form (`{name}_{handler,service}.go` to `{dir}/{name}_test.go`)
is documented as an explicit variant, not the default. Acceptance verifies that a
change touching both sources resolves both test paths without error.

`proto` and `iac` are direct: `proto/**/{name}.proto` to
`gen/{dir}/{name}.pb.go`, and `terraform/**/*.tf` to `{dir}/README.md`.

Preset rules:

1. Nothing decorative. v2 accepted three constructs that do nothing (OS6,
   OS10), and the presets taught two of them.
2. `skipIf: missing` is written out wherever a preset is lenient, because the
   preset is where people learn what it means.
3. Each preset carries a one-line comment per rule saying what it enforces.

## Order of work

1. OS7 (sorted iteration) is fixed in v2 (PR #51); keep its determinism guard
   and build everything else on it.
2. The v3 matcher (dialect, captures, binding) with layer 0 passing, then instance keys and resolved-path validation.
3. Change collection: removals, rename sources, `--name-status` input.
4. Evaluation and reports.
5. Presets, e2e projects, goldens re-recorded once.
6. Documentation.

## Release acceptance

Beyond the conformance suite, a v3 release is accepted when these work end to
end, because the matrix alone cannot prove them:

1. **First run.** `kyn init --preset <p>`, preview with `--dry-run-resolve`,
   change only a source and see it fail, add the related file and see it pass.
   The tutorial introduces `skipIf: missing` as an opt-in: starting with it can
   make a first run skip every rule.
2. **Container workflow.** The `docs/site/ci.md` distroless pipeline, fed
   `git diff --name-status -M` through `--stdin`, fails when a test is deleted.
3. **Stale checkout.** A diff that deletes a related file, evaluated from a tree
   that still has it, gives the same results as a fresh checkout.
4. **Upgrade from a pinned v2 setup.** A v3 binary given a v2 config exits 2 with
   the unsupported-version message; the docs show pinning the binary and config
   together and staying on the last v2 release during an upgrade, and list the
   report field changes (`familyId` to `instanceName`, `reason`, `summary`).
5. **Failure messages.** A failing result names the source that triggered the
   rule, the exact related path, what happened to it (`reason`), and what would
   satisfy the rule. An overlapping-glob error (rule 9) shows the path and both
   bindings.

Documentation for cutover also carries working recipes with expected output for
mirrored trees (`{**name}`), exclusions, and one failure for a shared changelog,
and states the limits early: "any new migration", "any changeset" and
deletion-triggered source rules are not expressible.

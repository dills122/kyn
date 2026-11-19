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
| `internal/matcher` | Replace doublestar with the v3 dialect in [`04-grammar.md`](04-grammar.md): captures, longest-first binding (with the rule that a capture may not sit between two `**`), code-point matching, and rejection of the constructs v2 accepted (`[ ]`, `\`, nested braces, wildcards in alternatives, partial-segment `**`, `.`/`..`/empty segments, whitespace-padded globs). Measured against this repository, no configuration uses a rejected form; `src/**` no longer matches a file named `src`. The oracle is the reference algorithm in [`07-conformance.md`](07-conformance.md) |
| `internal/changes` | Retain deleted paths and rename sources. Git mode runs `git diff --relative --name-status -z -M`: `-z` prints every name literally (`-c core.quotePath=false` alone still quotes a name containing `"`), which also keeps the OS16 fix; `--relative` makes a `--cwd` below the repository root work (today it matches nothing and passes). Count git `T` as `modified` and touched. Add the `--name-status` flag: `--files-from` and `--stdin` (the same thing as `--files-from -`) read plain paths without it and status lines with it, never a guess; a TAB line without the flag, and a plain line with it, are errors. In git and `--name-status` mode an added, modified or renamed-to path absent from the working tree is an error (exit 3 for git, 2 for input). Drop plain-mode sources that are not on disk. Reject conflicting lines and non-clean paths ([`02-semantics.md`](02-semantics.md) §6). **Compatibility break:** `NormalizePath` today trims whitespace, turns `\` into `/`, strips `./` and cleans the path; v3 rejects all of that with exit 2, for `--files` and `--files-from` alike. Blank lines and CRLF are tolerated. `--files` is comma-separated today, so a path containing a comma cannot be passed; use `--files-from` for it. Removing the cleaning is deliberate: a path Kyn rewrites silently is a path it may match differently than git does. The clean-path check is fatal only for a path a rule would match as a source, for resolved related paths, and for a leading `"` in `--files`/`--files-from`: a repository that tracks a name with an edge blank, a backslash, a control or zero-width character, or a first segment like `Q:` (legal on Linux and macOS) fails only runs where a rule matches it, with exit 2, naming the path. Deleted paths and rename sources are exempt, so the change that fixes a bad name can pass. Names with a format character at a segment edge (a tag-sequence flag, a trailing right-to-left mark) are rejected as a consequence of the Cf rule. No file in this repository is affected |
| `internal/rules`, `internal/report` | The three states, `skipIf`, skipped status, headline, renamed fields |
| `internal/cli/config_migrate.go`, `kyn config migrate`, `internal/cli/config.go` | Delete the command; the `config` parent has no other subcommand and goes with it. A CLI surface removal, justified because both ends of the migration are gone |
| `internal/cli/run_context.go` | The `families=%d` verbose line and the `family resolution failed` error text |
| `internal/config` load path | A step of its own, before the matcher: the document root, `version` handling, rules 1 to 8 ([`04`](04-grammar.md)) |
| `internal/cli/init.go` | Four presets rewritten; the `kin` and `families` wording in its help and next-steps text |
| `internal/cli/check.go` | `--dry-run-resolve` returns before the empty check today, so `--fail-on-empty` exits 0 with zero matches: v3 applies it to the preview, prints per-rule counts, and has the headline say when nothing matched or everything skipped ([`03`](03-instances-and-reports.md) §8, §9) |
| `internal/rules`, `internal/changes`: the `removed` state | The change record decides it before the working tree is read ([`02`](02-semantics.md) §2), so a stale checkout cannot hide a deletion |
| `internal/cli/check.go`, `explain.go` | Help text and flag descriptions that say "family" and "kin": command summaries, `--fail-on-empty` ("no family instances match"), `--dry-run-resolve` ("Resolve families/kin only") |

### Fixtures and tests

| Item | Count |
| --- | --- |
| `e2e/projects/*/kyn.config.yaml` | 4. `figma-publish-required` is an emit-only rule with `severity: warn`; v3 rule 8 requires `info`, so the rewrite changes that golden |
| `e2e/workflows_test.go` | the migration test is deleted; the preset-bootstrap and empty-input tests assert the v2 JSON shape (`"ok": true`) and are rewritten |
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

Found by sweeping the repository for `family`, `kin`, `groups`, `version: 2`,
`--name-only` and `config migrate`; the table is the inventory, not a sample.

| File | Action |
| --- | --- |
| `docs/site/ci.md` (the container job pipes `git diff --name-only` into `--stdin`), CLI help examples | Switch every enforcement example to `git diff --name-status -M ... | kyn check --stdin`; keep `--files` as the local quick start. Plain paths lose deletion protection ([`02`](02-semantics.md) §6) |
| `README.md`, `docs/site/config.md`, `docs/site/concepts.md`, `docs/site/troubleshooting.md`, `docs/site/recipes/*.md` | Rewrite for v3 |
| `docs/site/getting-started.md` | Rewrite for v3 (the v2 walkthrough is PR #49) |
| `docs/site/install.md`, `docs/site/index.md`, `docs/site/presets.md`, `docs/site/reference.md`, `mkdocs.yml`, `e2e/README.md`, `.codex/steering/repository-steering.md` | Rewrite or update for v3 wording; `install.md` and `docs/release.md` carry the `--name-only | --stdin` container job and switch to the status form |
| `docs/ci.md`, `docs/presets.md`, `docs/troubleshooting.md`, `docs/ergonomics.md`, `docs/cli-validation-matrix.md`, `docs/release.md` | Update, or banner as historical like `docs/spec.md` |
| `CHANGELOG.md` | The v3 entry leads with the break and the pin instructions |
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

```yaml
# proto
rules:
  proto-gen:
    match: "proto/**/{name}.proto"
    related: "gen/{dir}/{name}.pb.go"
    expect: touched
    skipIf: missing
```

```yaml
# iac
rules:
  module-readme:
    match: "terraform/**/*.tf"
    related: "{dir}/README.md"
    expect: touched
    skipIf: missing
    severity: warn
    message: "Module changed but its README did not."
```

Preset rules:

1. Nothing decorative. v2 accepted three constructs that do nothing (OS6,
   OS10), and the presets taught two of them.
2. Presets are **lenient**: every rule has `skipIf: missing`, so a repository
   that has no story, test or README for a file is not failed for it, and the
   preset is where people learn what `skipIf` means. Removing the line makes a
   rule strict. A first run therefore needs a related file to exist.
3. Each preset carries a one-line comment per rule saying what it enforces.

## Order of work

1. OS7 (sorted iteration) is fixed in v2 (PR #51); keep its determinism guard
   and build everything else on it.
2. The v3 matcher (dialect, captures, binding) with layer 0 passing.
3. The config document and load-time validation (rules 1 to 8), then instance keys and resolve-time rules 9 to 11.
4. Change collection: removals, rename sources, `--relative`, `--name-status` input, the working-tree check.
5. Evaluation and reports.
6. Presets, e2e projects, goldens re-recorded once.
7. Documentation and release notes.

## Release acceptance

Beyond the conformance suite, a v3 release is accepted when these work end to
end, because the matrix alone cannot prove them:

1. **First run.** In a repository that already has the related file,
   `kyn init --preset <p>`, preview with `--dry-run-resolve`, change only the
   source and see it fail, touch both and see it pass, delete the related file
   and see it fail. The tutorial then shows what `skipIf: missing` does by
   removing the related file from a copy: presets are lenient, so a repository
   with no related files skips every rule, which the report shows as skips.
2. **Container workflow.** The `docs/site/ci.md` distroless pipeline, fed
   `git diff --name-status -M` through `--stdin --name-status`, fails when a test
   is deleted; the same step fed `--name-only` fails with the declared-format error.
3. **Stale checkout.** A diff that deletes a related file, evaluated from a tree
   that still has it, gives the same results as a fresh checkout. A diff that
   adds a file, evaluated from a tree that lacks it, exits with the
   "check out head" error instead of a machine-dependent result.
4. **Upgrade from a pinned v2 setup.** A v3 binary given a v2 config exits 2 with
   the unsupported-version message; the docs show pinning the binary and config
   together and staying on the last v2 release during an upgrade, and list the
   report field changes below. Unpinned users (`brew upgrade`, `:latest`,
   `go install @latest`) get exit 2 and that message on their v2 config; the
   release notes lead with it. The channels (Homebrew, WinGet, Fury packages,
   container tags, `go install`) are each smoke-tested with a v2 config.
5. **Failure messages.** A failing result names the source that triggered the
   rule, the exact related path, what happened to it (`reason`), and what would
   satisfy the rule (the `Expected:` line). An overlapping-glob error (rule 9) shows the path and both
   bindings.

Documentation for cutover also carries working recipes with expected output for
mirrored trees (`{**name}`), exclusions, and one failure for a shared changelog,
and states the limits early: "any new migration", "any changeset" and
deletion-triggered source rules are not expressible.

## Field and meaning changes to document

JSON `check`: `ok` kept; `infos` becomes `summary.info`; `passed`, `failed`,
`warnings`, `errors` move into `summary` and now count failures by severity;
`flags` stays top level (including `--summary-only`); per result `familyId`,
`familyName`, `changedFiles`, `expectedFiles` become `instanceName`, `sources`,
`related`; new `headline`, `inputMode`, `removalInfo`, `reason`, `relatedState`,
`rulesWithoutInstances`. Text loses `Family:`. The `--dry-run-resolve` JSON keeps
`changedCount` and `changedFiles` and gains `inputMode` and per-rule counts.

`{name}` changes meaning. In v2 it is the whole file stem (`order_handler`); in
v3 it is whatever the glob captures, so `{name}_handler.go` gives `order`. Every
recipe that used `{name}` is rewritten, and there is no converter to do it.

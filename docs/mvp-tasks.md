# Kyn MVP Task Backlog

## Ticket Backlog

| ID | Task | Est. | Depends On |
|---|---|---:|---|
| MVP-001 | Create `docs/decisions.md` with locked MVP rules (input modes, git handling, dedupe, fail semantics) | 0.5d | - |
| MVP-002 | Patch spec: rename `require.flag` -> `require.emitFlag` and define evaluator semantics | 0.5d | MVP-001 |
| MVP-003 | Define CLI validation matrix (all valid/invalid flag combinations) in docs | 0.5d | MVP-001 |
| MVP-004 | Initialize Go module + baseline folder structure | 0.5d | - |
| MVP-005 | Add `cobra` root + `check` command skeleton with `--help` | 0.5d | MVP-004 |
| MVP-006 | Implement centralized exit code handling (`0/1/2/3`) | 0.5d | MVP-005 |
| MVP-007 | Implement flag parsing + enum validation (`--format`, `--fail-on`) | 0.5d | MVP-005 |
| MVP-008 | Implement strict input-mode validation (exactly one mode; base/head pair required) | 0.5d | MVP-007, MVP-003 |
| MVP-009 | Implement config discovery (`kyn.config.yaml`, etc.) + explicit `--config` path behavior | 0.5d | MVP-004 |
| MVP-010 | Implement YAML config loading (`yaml.v3`) | 0.5d | MVP-009 |
| MVP-011 | Implement config schema validation (version, ids, enums, references) | 1d | MVP-010 |
| MVP-012 | Implement template token validation (`{dir,file,name,base,ext}`) | 0.5d | MVP-011 |
| MVP-013 | Implement path normalization utility (slash, trim, strip `./`) | 0.5d | MVP-004 |
| MVP-014 | Implement manual changes provider: `--files` CSV parsing | 0.5d | MVP-013, MVP-008 |
| MVP-015 | Implement file-based changes provider: `--files-from` | 0.5d | MVP-013, MVP-008 |
| MVP-016 | Implement git changes provider with `git diff --name-status -M <base>...<head>` | 1d | MVP-013, MVP-008 |
| MVP-017 | Implement git status flattening policy (include A/M/R-new, exclude D) | 0.5d | MVP-016, MVP-001 |
| MVP-018 | Implement glob matcher (`doublestar`) include/exclude logic | 1d | MVP-013, MVP-011 |
| MVP-019 | Implement family instance resolver + `baseName.stripSuffixes` + kin template expansion | 1d | MVP-018, MVP-012 |
| MVP-020 | Implement family instance de-duplication (default on) | 0.5d | MVP-019, MVP-001 |
| MVP-021 | Implement rule engine `when` operators (`changedAny`, `kinExists`, `kinMissing`) | 1d | MVP-019 |
| MVP-022 | Implement rule engine `require` operators (`kinChanged`, `kinUnchanged`, `kinExists`, `kinMissing`, `emitFlag`) | 1d | MVP-021, MVP-002 |
| MVP-023 | Implement reporters (text/json), deterministic sorting, summary counts, `--fail-on-empty` | 1d | MVP-022 |
| MVP-024 | Build tests: unit + fixture + CLI validation matrix + golden outputs | 2d | MVP-023 |
| MVP-025 | README quickstart + sample config/testdata + CI usage examples | 0.5d | MVP-023 |

## Milestones

1. Foundation: `MVP-001` to `MVP-012`
2. Core Engine: `MVP-013` to `MVP-022`
3. Quality + Docs: `MVP-023` to `MVP-025`

## Critical Path

`MVP-001 -> MVP-002 -> MVP-011 -> MVP-018 -> MVP-019 -> MVP-021 -> MVP-022 -> MVP-023 -> MVP-024`

## Total Estimate

~16.5 engineer-days

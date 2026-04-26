# Kyn v1 to v2 Migration Guide

Kyn v2 keeps the same core behavior, but clarifies rule semantics and adds safer CLI defaults.

## Why Migrate

v2 improves:

- rule readability with `if` / `assert`
- side-effect clarity with `actions.emit`
- family structure with explicit groups
- status-aware matching with `changedStatusAny`
- better CLI ergonomics and diagnostics

## Automatic Migration

Safe side-by-side migration:

```bash
kyn config migrate -c kyn.config.yaml --from v1 --to v2
```

This writes `<input>.v2.yaml` by default and validates the migrated output before writing it.

In-place migration with backup:

```bash
kyn config migrate -c kyn.config.yaml --from v1 --to v2 --in-place
```

This creates `<input>.bak` by default.

## Field Mapping

The built-in migrator applies these transformations:

- `version: 1` -> `version: 2`
- `when` -> `if`
- `require` -> `assert`
- `require.emitFlag` -> `actions.emit`
- family `include` / `exclude` -> `groups.source.include` / `groups.source.exclude`

It also clears the legacy v1-only fields after conversion so the resulting config is a clean v2 document.

## Example

v1:

```yaml
version: 1

families:
  - id: angular-component
    include:
      - "libs/**/*.component.ts"
    kin:
      story: "{dir}/{base}.stories.ts"

rules:
  - id: story-sync
    family: angular-component
    severity: error
    when:
      changedAny: [source]
    require:
      kinChanged: [story]
      emitFlag: figmaPublishRequired
    message: "Story was not updated."
```

v2:

```yaml
version: 2

families:
  - id: angular-component
    groups:
      source:
        include:
          - "libs/**/*.component.ts"
    kin:
      story: "{dir}/{base}.stories.ts"

rules:
  - id: story-sync
    family: angular-component
    severity: error
    if:
      changedAny: [source]
    assert:
      kinChanged: [story]
    actions:
      emit:
        - figmaPublishRequired
    message: "Story was not updated."
```

## Recommended Post-Migration Checks

1. Run `kyn check -c kyn.config.v2.yaml --dry-run-resolve`
2. Run `kyn explain -c kyn.config.v2.yaml --summary-only`
3. Compare behavior on a few real PR change sets before switching CI over

## Breaking Changes to Watch

- v2 prefers explicit family groups over implicit source patterns
- `actions.emit` is informational and separate from assertions
- auto git mode means `kyn check` without input flags may now work by default in CI git worktrees

If you want strict explicit behavior in pipelines, keep using:

```bash
kyn check --strict-input-mode --base origin/main --head HEAD
```

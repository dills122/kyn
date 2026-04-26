# Troubleshooting

## No Input Mode Selected

Error example:

```txt
auto input mode unavailable: no explicit mode provided and --cwd is not a git repository.
```

Use one of:

```bash
kyn check --files path/a.ts,path/b.ts
kyn check --files-from changed.txt
kyn check --stdin
kyn check --base origin/main --head HEAD
```

## Strict Mode Failure

Error example:

```txt
invalid input mode: expected exactly one mode, observed none.
```

If your pipeline is meant to stay explicit, use:

```bash
kyn check --strict-input-mode --base origin/main --head HEAD
```

## Git Diff Failure

Error example:

```txt
git change detection failed
```

Common causes:

- base ref is not fetched in CI
- shallow clone does not contain the merge base
- `origin/main` is not the correct default branch

Typical fix:

```bash
git fetch origin main --depth=100
kyn check --base origin/main --head HEAD
```

Or set:

- `KYN_BASE_REF`
- `KYN_HEAD_REF`

for auto mode.

## No Matches Found

By default, Kyn returns success with empty results when no family instances match.

If you want CI to fail in that case, use:

```bash
kyn check --fail-on-empty
```

## Migrated Config Looks Wrong

Use the safe migrator first:

```bash
kyn config migrate -c kyn.config.yaml --from v1 --to v2
```

Then inspect behavior with:

```bash
kyn check -c kyn.config.v2.yaml --dry-run-resolve
kyn explain -c kyn.config.v2.yaml
```

## Reviewdog / SARIF / Checkstyle Output Not Accepted

Check that you are using the right format for the integration:

- `rdjson` for reviewdog
- `sarif` for GitHub code scanning and SARIF consumers
- `checkstyle` for checkstyle XML parsers

Examples are in [ci.md](ci.md).

## Windows Path Confusion

Kyn normalizes internal matching to slash-separated paths.

Use config globs and kin templates like:

```yaml
include:
  - "src/**/*.go"
```

not backslash-separated Windows path patterns.

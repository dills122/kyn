# Kyn CLI Validation Matrix (MVP)

This matrix defines valid and invalid combinations for change input flags.

## Valid

1. `kyn check --files a.ts`
2. `kyn check --files-from changed.txt`
3. `kyn check --stdin`
4. `kyn check --base origin/main --head HEAD`
5. `git diff --name-only origin/main...HEAD | kyn check --files-from -`
6. `git diff --name-only origin/main...HEAD | kyn check --stdin`

## Invalid (Exit 2)

1. `kyn check` (no change input mode)
2. `kyn check --files a.ts --files-from changed.txt`
3. `kyn check --files a.ts --base origin/main --head HEAD`
4. `kyn check --files-from changed.txt --base origin/main --head HEAD`
5. `kyn check --base origin/main` (missing `--head`)
6. `kyn check --head HEAD` (missing `--base`)
7. `kyn check --base origin/main --head HEAD --files a.ts`
8. `kyn check --base origin/main --head HEAD --files-from changed.txt`
9. `kyn check --stdin --files-from changed.txt`
10. `kyn check --stdin --files a.ts`

## Other Validation Rules

1. `--format` must be `text` or `json`.
2. `--fail-on` must be `error` or `warn`.
3. `--show-passes` only affects text output rendering.
4. Unknown positional arguments after flags are invalid usage.

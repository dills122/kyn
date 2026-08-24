# Troubleshooting

## No input mode outside Git

```text
auto input mode unavailable: no explicit mode provided and --cwd is not a git repository
```

Choose one explicit mode:

```bash
kyn check --files path/a.ts,path/b.ts
kyn check --files-from changed.txt
kyn check --stdin
kyn check --base origin/main --head HEAD
```

## Base ref is missing in CI

A shallow checkout often does not contain `origin/main` or the merge base.
Fetch enough history before running Kyn:

```bash
git fetch origin main --depth=100
kyn check -c kyn.config.yaml --base origin/main --head HEAD
```

Use your repository's real default branch. You can also set
`KYN_BASE_REF` and `KYN_HEAD_REF` for automatic mode.

## A rule never appears

Rules only appear when their `if` clauses apply. Inspect matching first:

```bash
kyn check -c kyn.config.yaml \
  --files path/to/source.go \
  --dry-run-resolve
```

Then inspect clause diagnostics:

```bash
kyn explain -c kyn.config.yaml --files path/to/source.go
```

Check the source glob, suffix stripping, resolved kin path, and whether
`kinExists` intentionally filters the rule.

## A warning says fail but CI passed

`Status: fail` means the assertion failed. The default process threshold is
`error`, so failed warning rules are reported without blocking. Make warnings
blocking with:

```bash
kyn check -c kyn.config.yaml --fail-on warn
```

## No family instances matched

No matches are allowed by default. This is useful when a change does not touch a
configured family. To make an unexpectedly empty scope a pipeline failure:

```bash
kyn check -c kyn.config.yaml --fail-on-empty
```

## Windows paths do not match

Write config globs and kin templates with forward slashes:

```yaml
include:
  - "src/**/*.go"
```

Kyn normalizes paths internally on every operating system.

## Config migration needs review

```bash
kyn config migrate -c kyn.config.yaml --from v1 --to v2
kyn check -c kyn.config.v2.yaml --dry-run-resolve --files path/to/source
kyn explain -c kyn.config.v2.yaml --files path/to/source
```

Compare resolution and results before replacing the v1 file.

## A machine-output integration rejects the report

Match the consumer:

- reviewdog: `--format rdjson`
- GitHub code scanning: `--format sarif`
- Checkstyle consumers: `--format checkstyle`
- custom scripts: `--format json`

Only `check` supports SARIF, RDJSON, and Checkstyle. `explain` supports text
and JSON.

When retaining verbose diagnostics with a machine report, redirect the streams
separately:

```bash
kyn check --format json --verbose > kyn.json 2> kyn-debug.log
```

## Explain displays FAIL but exits 0

This is intentional. `kyn explain` is diagnostic and returns success when
evaluation and rendering complete, even when the policy summary is `FAIL`.
Run `kyn check` with the same input when policy failure must return exit 1.

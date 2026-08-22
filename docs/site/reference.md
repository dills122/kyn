# CLI and Exit-Code Reference

## Commands

| Command | Purpose |
| --- | --- |
| `kyn check` | Evaluate policy and return a CI-ready exit code |
| `kyn explain` | Show per-rule diagnostics without returning rule-failure exit 1 |
| `kyn init` | Generate a v2 starter config |
| `kyn config migrate` | Convert a v1 config to v2 |
| `kyn version` or `kyn --version` | Print build and release metadata |

Run `kyn <command> --help` for the complete flag list.

## Check input modes

Exactly one explicit mode may be selected:

```bash
# Comma-separated paths
kyn check -c kyn.config.yaml --files src/a.go,src/a_test.go

# One path per line
kyn check -c kyn.config.yaml --files-from changed-files.txt

# Standard input
git diff --name-only origin/main...HEAD | kyn check -c kyn.config.yaml --stdin

# Git diff
kyn check -c kyn.config.yaml --base origin/main --head HEAD
```

With no explicit mode, Kyn uses Git diff with base `origin/main` and head
`HEAD`. Set `--strict-input-mode` when a pipeline should reject that
automatic behavior.

## Common check flags

| Flag | Default | Effect |
| --- | --- | --- |
| `-c, --config` | Search defaults | Config path |
| `--cwd` | `.` | Repository root for config, Git, existence checks, and paths |
| `-o, --format` | `text` | `text`, `json`, `sarif`, `rdjson`, or `checkstyle` |
| `--fail-on` | `error` | Blocking threshold: `error` or `warn` |
| `--fail-on-empty` | false | Fail if no family instance matches |
| `--show-passes` | false | Include passing rules in text output |
| `--summary-only` | false | Omit per-rule detail |
| `--dry-run-resolve` | false | Resolve families and kin without evaluating rules |
| `--verbose` | false | Print config and input diagnostics |

`explain` shares the input and diagnostic flags but supports only text and JSON.

## Output formats

| Format | Use |
| --- | --- |
| `text` | Developers and CI logs |
| `json` | General automation and custom steps |
| `sarif` | GitHub code scanning and other SARIF consumers |
| `rdjson` | reviewdog |
| `checkstyle` | CI systems that ingest Checkstyle XML |

Machine outputs go to standard output. Operational and validation errors go to
standard error.

## Exit codes

| Code | Meaning | Typical response |
| --- | --- | --- |
| 0 | Policy passed at the selected threshold | Continue |
| 1 | One or more applicable rules blocked | Update kin or review the policy |
| 2 | CLI usage or config validation error | Fix the command or YAML |
| 3 | Runtime/provider error, such as Git failure | Fix checkout/ref/runtime state |

These codes are stable API. CI should distinguish policy failures from broken
tool setup instead of treating every nonzero result as the same problem.

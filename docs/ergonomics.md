# Kyn Ergonomics & Adoption Contract

## Objective

A new user should discover Kyn, run it, and integrate it into CI in **10 minutes or less**.

## First-Run Experience (Required)

1. Install one binary.
2. Run `kyn init` (or quick-start config template).
3. Run `kyn check`.
4. See immediately actionable output.
5. Paste one CI command and gate PRs.

## Golden Path Commands

Primary command:

```bash
kyn check
```

Expected behavior:

- Auto-discover config from default paths.
- Auto-select git mode when in a git repo (unless strict mode is enabled).
- Produce deterministic output.
- Exit with stable CI-safe codes.

Strict mode for explicit pipelines:

```bash
kyn check --strict-input-mode --base origin/main --head HEAD
```

## CLI Design Rules

1. Keep common flags minimal and memorable.
2. Keep advanced flags available but de-emphasized.
3. Every error must explain what failed and what to do next.
4. Help output must include copy/paste happy-path examples.

## Flag Tiers

Core flags (highlight in help):

- `-c, --config`
- `-o, --format`
- `--fail-on`

Input mode flags:

- `-f, --files`
- `--files-from`
- `--stdin`
- `--base --head`

Advanced flags:

- `--strict-input-mode`
- `--summary-only`
- `--dry-run-resolve`
- `--show-passes`
- `--verbose`
- `--cwd`

## Error Message Contract

Errors should include:

1. The offending mode/rule/family context.
2. What was expected vs observed.
3. A concrete next step command.

Example:

```txt
invalid input mode: selected files + git.
Choose exactly one: --files | --files-from | --stdin | --base+--head.
Try: kyn check --strict-input-mode --base origin/main --head HEAD
```

## CI Portability Requirements (Linux Distros/Platforms)

Kyn must be usable across major CI runners and Linux distributions.

Required distribution targets:

- `linux-amd64` (glibc)
- `linux-arm64` (glibc)
- `linux-amd64` (musl)
- `linux-arm64` (musl)
- `darwin-amd64`
- `darwin-arm64`
- `windows-amd64`

Required release artifacts:

1. GitHub Release binaries.
2. SHA256 checksums.
3. Container image (GHCR) for runner/container workflows.
4. Versioned changelog entry.

## CI Provider Coverage (Required Docs + Snippets)

Provide tested snippets for:

- GitHub Actions
- GitLab CI
- Jenkins
- CircleCI
- Buildkite
- Azure Pipelines

## Adoption Metrics

Track:

1. Time-to-first-pass (`TTP`) median under 10 minutes.
2. CI integration median under 15 minutes.
3. Runtime portability issues under 2% of support tickets.
4. Percentage of users succeeding with only `init`, `check`, `explain`.

## Non-Goals

- Advanced plugin marketplace in v2.
- Non-CI daemon/watch workflows in v2.

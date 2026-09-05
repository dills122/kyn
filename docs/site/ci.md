# Adopt Kyn in CI

The safest rollout starts with one proven local policy, reports it in CI, and
tightens the blocking threshold only after the team trusts the signal.

## 1. Prove the policy locally

Use explicit paths while tuning:

```bash
kyn check -c kyn.config.yaml \
  --files path/to/source,path/to/related-file \
  --show-passes
```

Preview family and kin resolution with `--dry-run-resolve`. Diagnose rule
applicability with `kyn explain`.

## 2. Choose the rollout threshold

| Rollout | Config/flags | Behavior |
| --- | --- | --- |
| Observe | Use `warn` rules and default `--fail-on error` | Reports warning failures without blocking |
| Enforce critical rules | Mix `warn` and `error` severities | Only failed errors block |
| Enforce everything | Add `--fail-on warn` | Failed warnings and errors block |
| Detect a broken scope | Add `--fail-on-empty` | No matched family instances blocks |

## 3. Pin the tool

CI should use an immutable release:

```bash
go install github.com/dills122/kyn/cmd/kyn@v0.1.3
```

If the job should not install Go, use the pinned
`ghcr.io/dills122/kyn:0.1.3` image or a checksummed release archive. See
[Install Kyn](install.md) for every channel.

## 4. Make the base ref available

The canonical gate is:

```bash
kyn check -c kyn.config.yaml \
  --base origin/main \
  --head HEAD \
  --format json
```

Shallow checkouts frequently omit the base or merge history. Fetch the default
branch before Kyn runs, or configure the checkout step to retain full history.
Replace `main` when your default branch has a different name.

## GitHub Actions

```yaml
name: related-file-policy

on:
  pull_request:

jobs:
  kyn:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: "1.22.x"

      - name: Install Kyn
        run: go install github.com/dills122/kyn/cmd/kyn@v0.1.3

      - name: Check related-file policy
        run: kyn check -c kyn.config.yaml --base origin/main --head HEAD --format text
```

Use JSON for a downstream script, SARIF for code scanning, or RDJSON for
reviewdog.

## GitLab CI

```yaml
kyn:
  image: golang:1.22
  variables:
    GIT_DEPTH: "0"
  before_script:
    - go install github.com/dills122/kyn/cmd/kyn@v0.1.3
  script:
    - kyn check -c kyn.config.yaml --base origin/main --head HEAD --format text
```

## Container job

The distroless Kyn image does not contain Git. Compute changed paths on the host,
mount the repository for config and existence checks, and pipe the list:

```bash
git diff --name-only origin/main...HEAD \
  | docker run --rm -i \
  -v "$PWD:/work" \
  -w /work \
  ghcr.io/dills122/kyn:0.1.3 \
  check -c kyn.config.yaml --stdin --format json
```

## Existing change detector

If CI already computes changed paths, do not make Kyn run a second Git diff:

```bash
your-change-detector | kyn check \
  -c kyn.config.yaml \
  --stdin \
  --format json
```

Explicit path streams record every supplied path as `modified`; they do not
preserve added or renamed status, and deleted paths are not currently evaluated.
Rules that depend on added or renamed status should keep Git input mode.

## Machine-output integrations

### GitHub code scanning

```yaml
- name: Run Kyn as SARIF
  id: kyn
  continue-on-error: true
  run: kyn check -c kyn.config.yaml --base origin/main --head HEAD --format sarif > kyn.sarif

- name: Upload Kyn SARIF
  if: always()
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: kyn.sarif
```

`continue-on-error` lets the upload step run after a policy failure. Add a
final gate appropriate to your workflow if code-scanning ingestion is not
itself the required check.

### reviewdog

```bash
kyn check -c kyn.config.yaml \
  --base origin/main \
  --head HEAD \
  --format rdjson \
  | reviewdog -f=rdjson -reporter=github-pr-review
```

### Checkstyle

```bash
kyn check -c kyn.config.yaml \
  --base origin/main \
  --head HEAD \
  --format checkstyle > kyn-checkstyle.xml
```

## Handle exit codes deliberately

| Code | CI classification |
| --- | --- |
| 0 | Policy passed at the chosen threshold |
| 1 | Code/policy failure |
| 2 | Command or config is invalid |
| 3 | Checkout, Git, filesystem, or runtime setup failed |

Codes 2 and 3 indicate a broken pipeline or tool setup, not a developer choosing
the wrong related files.

## Debug a failed job

Reproduce the exact refs locally, then switch to human-readable diagnostics:

```bash
kyn check -c kyn.config.yaml \
  --base origin/main \
  --head HEAD \
  --format text \
  --verbose \
  --show-passes

kyn explain -c kyn.config.yaml --base origin/main --head HEAD
```

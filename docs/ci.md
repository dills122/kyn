# Kyn in CI (DevOps Guide)

This guide focuses on predictable CI integration, clear failure behavior, and low-friction setup.

## Canonical CI Command

Use this as the default gate in CI:

```bash
kyn check -c kyn.config.yaml --base origin/main --head HEAD --format json
```

## Failure Semantics

- `0`: policy passed
- `1`: policy failed (rule violations)
- `2`: usage/config error (pipeline/config issue)
- `3`: runtime/provider error (for example git invocation failed)

Recommended policy:

- Treat `1` as code quality/policy failure.
- Treat `2` and `3` as pipeline/tooling failures.

## Policy Matrix

| Team Policy | Recommended Flags | Behavior |
|---|---|---|
| Error-only blocking | `--fail-on error` | Only error severity failures block |
| Strict blocking | `--fail-on warn` | Warn and error failures block |
| Ensure scope matched | `--fail-on-empty` | Fails when no family instances matched |
| Debug reruns | `--verbose --show-passes` | More context in text output |

## CI Provider Snippets

### GitHub Actions

```yaml
- name: Kyn policy check
  run: |
    go build -o ./bin/kyn ./cmd/kyn
    ./bin/kyn check -c kyn.config.yaml --base origin/main --head HEAD --format json
```

### GitLab CI

```yaml
kyn_check:
  image: golang:1.22
  script:
    - go build -o ./bin/kyn ./cmd/kyn
    - ./bin/kyn check -c kyn.config.yaml --base origin/main --head HEAD --format json
```

### Jenkins (Declarative)

```groovy
stage('Kyn Check') {
  steps {
    sh 'go build -o ./bin/kyn ./cmd/kyn'
    sh './bin/kyn check -c kyn.config.yaml --base origin/main --head HEAD --format json'
  }
}
```

### Azure Pipelines

```yaml
- script: go build -o ./bin/kyn ./cmd/kyn
  displayName: Build Kyn

- script: ./bin/kyn check -c kyn.config.yaml --base origin/main --head HEAD --format json
  displayName: Kyn Check
```

## Provider-Agnostic Piped Mode

If CI already computes changed files:

```bash
your-change-detector-command | kyn check -c kyn.config.yaml --stdin --format json
```

Equivalent explicit form:

```bash
your-change-detector-command | kyn check -c kyn.config.yaml --files-from - --format json
```

## Debug Workflow

1. Run default gate command with `--format json`.
2. Re-run failed job with `--format text --verbose --show-passes`.
3. Use explicit files mode for targeted local repro:
   `kyn check -c kyn.config.yaml -f path/a.ts,path/b.ts`

# CI Guide

Recommended CI command:

```bash
kyn check -c kyn.config.yaml --base origin/main --head HEAD --format json
```

Use `--fail-on warn` when warnings should block.

Use machine formats as needed:

- `--format sarif`
- `--format rdjson`
- `--format checkstyle`

For full provider-specific snippets, see [docs/ci.md](../ci.md).

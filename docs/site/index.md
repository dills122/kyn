# Kyn

Kyn is a focused Go CLI that enforces related-file policy in CI with deterministic output and stable exit codes.

## Highlights

- Stateless CLI workflow
- Deterministic text and machine outputs
- Strong CI ergonomics (`check`, `explain`, `init`, `config migrate`)
- v1 + v2 config support with migration path

## Core Contract

- `0`: success
- `1`: rule failure
- `2`: usage/config validation error
- `3`: runtime/provider error

## Quick Example

```bash
kyn check -c kyn.config.yaml --base origin/main --head HEAD -o json
```

## Learn More

- [Getting Started](getting-started.md)
- [CI Guide](ci.md)
- [Config & Rules](config.md)
- [Presets](presets.md)
- [Release](release.md)

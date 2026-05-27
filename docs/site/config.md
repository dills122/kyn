# Config and Rules

Kyn config defines:

- `families`
- `kin`
- `rules`

v2 rule model:

- `if`: applicability
- `assert`: pass/fail policy checks
- `actions.emit`: informational flags

Key behavior:

- Deterministic ordering of outputs
- `if` and `assert` use AND semantics across keys
- Stable exit codes across all runs

See full schema and semantics in [docs/spec.md](../spec.md).

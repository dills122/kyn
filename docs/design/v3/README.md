# v3 configuration design

Working area for the v3 configuration design. Nothing here is approved and
nothing here changes shipped behavior.

The exploratory proposal and its reviews live in
[`../../reviews/`](../../reviews/). This directory holds the design work that
follows from them.

Start with the workplan. Everything else is downstream of the measured
baseline in `01`.

| Document | Gate | Contents |
| --- | --- | --- |
| [`00-workplan.md`](00-workplan.md) | — | Finding register, gates, decisions D1–D14, open questions |
| [`01-observed-semantics.md`](01-observed-semantics.md) | — | Measured v2 behavior — the factual baseline for every decision |
| [`02-semantics.md`](02-semantics.md) | G1 | Deletion gap closed; expectation grid frozen as `when` + `expect` |
| [`03-instance-and-identity.md`](03-instance-and-identity.md) | G2 | Instances keyed on the source shape; report identity preserved |
| [`04-grammar.md`](04-grammar.md) | G3 | The v3 grammar, validation rules, and answers to the proposal's open questions |
| [`05-compatibility-matrix.md`](05-compatibility-matrix.md) | G3 | Twenty v1/v2 shapes probed against the shipped binary |
| [`06-migration.md`](06-migration.md) | G4 | Migratable normal form, refusal set, preset coverage |
| [`07-conformance.md`](07-conformance.md) | G5 | Differential harness, sanctioned differences, scenario matrix |
| [`experiments/`](experiments/) | — | Reproduction scripts for every observation in `01` |

Nothing is implemented. G0 — a set of v2 fixes that stand on their own merits —
is specified in the workplan and blocks the conformance harness.

## Running the experiments

```bash
go build -o /tmp/kyn ./cmd/kyn
KYN=/tmp/kyn bash docs/design/v3/experiments/e3-truthtable.sh
```

Each script builds a disposable fixture under `$TMPDIR` and leaves the
repository untouched.

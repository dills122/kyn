# v3 configuration design

Working area for the v3 configuration design. Nothing here is approved and
nothing here changes shipped behavior.

The exploratory proposal and its reviews live in
[`../../reviews/`](../../reviews/). This directory holds the design work that
follows from them.

| Document | Contents |
| --- | --- |
| [`00-workplan.md`](00-workplan.md) | Finding register, gates, decisions taken, open questions |
| [`01-observed-semantics.md`](01-observed-semantics.md) | Measured v2 behavior — the factual baseline for every decision |
| [`experiments/`](experiments/) | Reproduction scripts for everything in `01` |

Later gates add `02-semantics.md`, `03-instance-and-identity.md`,
`04-grammar.md`, `05-compatibility-matrix.md`, and `06-migration.md`. See the
gate list in the workplan.

## Running the experiments

```bash
go build -o /tmp/kyn ./cmd/kyn
KYN=/tmp/kyn bash docs/design/v3/experiments/e3-truthtable.sh
```

Each script builds a disposable fixture under `$TMPDIR` and leaves the
repository untouched.

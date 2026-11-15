# v3 configuration design

Design for the v3, rule-centric configuration format. Nothing here is approved
or implemented, and nothing here changes shipped behavior.

```yaml
version: 3

rules:
  handler-test:
    match: "internal/**/{name}_handler.go"
    related: "{dir}/{name}_handler_test.go"
    expect: touched
    skipIf: missing
```

Start with the grammar. Each document describes the current design only; how it
got here is in the workplan.

| Document | Contents |
| --- | --- |
| [`04-grammar.md`](04-grammar.md) | The format: fields, captures, validation, limits |
| [`02-semantics.md`](02-semantics.md) | What Kyn observes, the three related-path states, `skipIf`, `expect`, input modes |
| [`03-instances-and-reports.md`](03-instances-and-reports.md) | Instances, names, ordering, report fields, skipped results, headline |
| [`06-cutover.md`](06-cutover.md) | Removing v1/v2: measured inventory, presets, order of work |
| [`07-conformance.md`](07-conformance.md) | The suite that proves an implementation matches |
| [`01-observed-semantics.md`](01-observed-semantics.md) | Measured v2 behavior, the factual baseline |
| [`05-compatibility-matrix.md`](05-compatibility-matrix.md) | v1/v2 shapes probed; historical |
| [`00-workplan.md`](00-workplan.md) | Decision log, superseded decisions, review dispositions, open items |
| [`experiments/`](experiments/) | Reproduction scripts and the scenario simulator |
| [`../../reviews/`](../../reviews/) | The nine independent reviews, recorded verbatim, and their frozen scopes |

## Running the experiments

```bash
go build -o /tmp/kyn ./cmd/kyn
KYN=/tmp/kyn bash docs/design/v3/experiments/e3-truthtable.sh
```

Each script builds a disposable fixture under `$TMPDIR` and leaves the
repository untouched. They measure the shipped v2 binary.

The scenario simulator models the v3 design instead:

```bash
go run docs/design/v3/experiments/sim/main.go docs/design/v3/experiments/sim/scenarios.yaml
```

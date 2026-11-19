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
got here is in the notes.

| Document | Contents |
| --- | --- |
| [`04-grammar.md`](04-grammar.md) | The format: fields, captures, validation, limits |
| [`02-semantics.md`](02-semantics.md) | What Kyn observes, the three related-path states, `skipIf`, `expect`, input modes |
| [`03-instances-and-reports.md`](03-instances-and-reports.md) | Instances, names, ordering, report fields, skipped results, headline |
| [`06-cutover.md`](06-cutover.md) | Removing v1/v2: measured inventory, presets, order of work |
| [`07-conformance.md`](07-conformance.md) | The suite that proves an implementation matches |
| [`notes.md`](notes.md) | Measured v2 behavior (OS1 to OS16), decisions that shaped the specs, review history, open items |

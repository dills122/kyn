# Conformance suite

Status: proposed. Current state only; history is in
[`00-workplan.md`](00-workplan.md).

The suite that proves an implementation matches
[`02-semantics.md`](02-semantics.md), [`03-instances-and-reports.md`](03-instances-and-reports.md)
and [`04-grammar.md`](04-grammar.md). There is no v2 side to compare against;
the oracle is the tables in those documents, computed independently of the
implementation.

## First: determinism

OS7 measured today's code reporting three different errors across thirty
identical runs, and v3 reuses that code. The first test written runs the same
input repeatedly and requires byte-identical output, for valid and invalid
configs. It is the only criterion that can fail intermittently, so nothing else
is trustworthy until it passes.

## Harness

`e2e/e2e_test.go` already walks `e2e/projects/*`, runs the real binary and
compares rendered output and exit codes. The suite extends that harness and
adds projects; it does not build a second one. Tests are table-driven and
goldens sit beside the existing ones.

Assert on rendered output. Two policies can normalize identically and still
print different text, and the text is what people and CI consume.

## Observables

| Observable | Notes |
| --- | --- |
| `check` exit code | under each `--fail-on` value |
| `check` text | headline, counts including `Skipped`, `Instance:` lines |
| `check` JSON | `ruleId`, `instanceName`, statuses including `skipped`, counters, `flags` |
| SARIF, RDJSON, Checkstyle | failures only; `(instance: …)` message suffix; `why` in SARIF |
| `explain` text and JSON | `SkipIf:` / `Expect:` sections, clause strings, the state a `skipIf` saw |
| `--dry-run-resolve` | per rule: instances, sources, related path |
| ordering | `ruleId`, then `instanceName` |

`--show-passes` is on, so a failure flipping to a pass or a skip shows up as a
changed line rather than a missing one.

## Layer 1: semantics, exhaustive

| Dimension | Values | Count |
| --- | --- | --- |
| `skipIf` | omitted, `missing`, `exists` | 3 |
| `expect` | `touched`, `untouched`, `exists`, `missing`, `none`, `[exists, touched]` | 6 |
| related path | `exists`+touched, `exists`+untouched, `removed` by delete, `removed` by rename, `missing`+untouched, `missing`+touched | 6 |

3 × 6 × 6 = **108 cells**, run in git mode, asserted on `explain` JSON, `check`
JSON and the exit code.

The related-path rows are the six reachable combinations from
[`02-semantics.md`](02-semantics.md) §4. `missing`+touched needs the diff and
the working tree to disagree: the fixture commits the related file on the head
ref and evaluates from a checkout without it. It is the only row that separates
`touched` from `[exists, touched]`, and an earlier draft of this document
wrongly excluded it.

Required assertions, stated positively so the suite fails if the behavior is
absent:

- `removed` rows **fail** under `skipIf: missing` with `expect: touched`.
  Reopening the delete-to-silence hole (OS3, OS8) must turn the suite red.
- `missing`+untouched **skips** under `skipIf: missing` and **applies** under
  `skipIf: exists`.
- every `expect: none` cell reports `info`, never fails, and adds its flags
  only when the rule applies.

Plus the related-less informational rule, which has no related-path dimension.

## Layer 2: shape and projection, sampled

Applied to a sample of layer 1 across every output mode.

| Dimension | Values |
| --- | --- |
| input mode | git; `--files-from` with name-status lines including a rename; `--files-from` plain; `--files` |
| instance shape | per file, per directory, per capture only, `(all)`, `(root)`, `{**name}` mirrored tree |
| sources per instance | one, several |
| `changeType` | absent, matching, non-matching, mixed within one instance |
| severity × `--fail-on` | `info`/`warn`/`error` × `error`/`warn` |
| output mode | all seven |

Severity crossed with `--fail-on` is what reaches all three headlines; each is
asserted against its condition. Input mode is where the plain-path degradation
in [`02-semantics.md`](02-semantics.md) §6 is pinned down: a deleted related
file skips under `skipIf: missing` with `--files` and fails in git mode.

## Corpus

- the four rewritten presets and the documented recipes
- the 21 scenarios in
  [`experiments/sim/scenarios.yaml`](experiments/sim/scenarios.yaml), promoted
  to real projects; the simulator's output is a draft of the expected results,
  not an oracle
- one invalid config per validation rule in [`04-grammar.md`](04-grammar.md),
  asserted to be rejected with exit 2 and the expected message

The invalid configs matter as much as the valid ones. All eleven validation
rules are new or changed and none has been exercised by real code.

## Exit criteria

1. The determinism guard passes.
2. Layer 1 passes for all 108 cells, with the positive assertions above.
3. Layer 2 passes across all seven output modes and all three headlines.
4. Every invalid corpus config is rejected with the expected message.
5. Goldens are recorded once, after 1 to 4, as a reviewed change.

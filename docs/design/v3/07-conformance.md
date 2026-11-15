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
compares rendered output and exit codes. The suite extends that harness; it does
not build a second one. Tests are table-driven and goldens sit beside the
existing ones.

The existing harness cannot run layer 1 as it stands: its projects are static
directories driven by `--files`, with no git. The extension is a helper that
builds a repository from a description (commits, deletions, renames, and a
checkout that lacks a file present on the head ref), so each cell is a row of
data. See [`06-cutover.md`](06-cutover.md).

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

## Layer 0: the matcher

The capture grammar is the largest new mechanism and the one with no existing
code to lean on, so it gets its own table-driven layer, run on the matcher in
isolation. Each row is a glob, a path, and either the expected bindings or
no-match.

| Group | Cases |
| --- | --- |
| Captures | `{name}` within a segment; leading dot; dotted values; greedy with a repeated literal (`{name}.{kind}.ts` on `user.profile.component.ts`) |
| `{**name}` | zero segments; one; many; trailing form needs one or more; a file at the root |
| Alternation | `{ts,html}` is not captured; `{ts}` is a capture |
| Several globs | agreeing bindings; **disagreeing bindings are an error**; reordering the list changes nothing |
| Exclude | removes a source; also stops a related path counting as a self-match |
| Instance names | one case per row of [`03-instances-and-reports.md`](03-instances-and-reports.md) §2, including a root file next to a directory literally named `(root)` |
| Invalid globs | one case per load-time error in [`04-grammar.md`](04-grammar.md): unclosed brace, invalid name, `dir`, duplicate capture, adjacent captures, partial-segment `{**x}`, capture in exclude, two `{**x}` |
| Invalid templates | glob characters, `..`, absolute path, unknown variable, two variables that can contain `/`, unused capture |

The simulator's scenarios 14 to 58 are the starting point. They are a draft of
the expected results, not an oracle.

## Layer 1: semantics, exhaustive

| Dimension | Values | Count |
| --- | --- | --- |
| `skipIf` | omitted, `missing`, `exists` | 3 |
| `expect` | `touched`, `untouched`, `exists`, `missing`, `none`, `[exists, touched]` | 6 |
| related path | `exists`+touched, `exists`+untouched, `removed` by delete, `removed` by rename, `missing`+untouched, `missing`+touched | 6 |

3 × 6 × 6 = 108, less the 12 cells with `skipIf: exists` and an `expect` that
includes `exists`, which validation rejects as unable to pass, leaves **96
cells**. Each runs in git mode and is asserted on `explain` JSON, `check` JSON
and the exit code.

The related-path rows are the six reachable combinations from
[`02-semantics.md`](02-semantics.md) §4. `missing`+touched needs the diff and
the working tree to disagree: the fixture commits the related file on the head
ref and evaluates from a checkout without it. It is the only row that separates
`touched` from `[exists, touched]`, and an earlier draft of this document
wrongly excluded it.

Required assertions, stated positively so the suite fails if the behavior is
absent:

- `removed` rows **apply** under both `skipIf` values, and **fail** under
  `skipIf: missing` with `expect: touched`. Reopening the delete-to-silence hole
  (OS3, OS8) must turn the suite red.
- `skipIf: missing` skips only `missing`; `skipIf: exists` skips only `exists`.
- `expect: untouched` **fails** on both `removed` rows.
- every `expect: none` cell reports `info`, never fails, and adds its flags only
  when the rule applies; flags are also added when a non-`none` rule fails or
  passes, and never when it is skipped.
- the failure message names the most fundamental cause, in the order in
  [`02-semantics.md`](02-semantics.md) §5.

Plus the related-less informational rule, which has no related-path dimension.

## Layer 2: shape and projection, sampled

Applied to a sample of layer 1 across every output surface.

| Dimension | Values |
| --- | --- |
| input mode | git; `--files-from` with name-status lines including a rename; `--files-from` plain; `--files` |
| instance shape | per file, per directory, per capture only, `(all)`, `.`, `{**name}` mirrored tree |
| sources per instance | one, several |
| `changeType` | absent, matching, non-matching (reported as `skipped`), mixed within one instance |
| severity × `--fail-on` | `info`/`warn`/`error` × `error`/`warn` |
| output surface | `check` text, JSON, SARIF, RDJSON, Checkstyle; `explain` text, JSON; `--dry-run-resolve` |

That is eight surfaces: five `check` formats, two `explain` formats, and the
resolve preview.

Severity crossed with `--fail-on` is what reaches all three headlines; each is
asserted against its condition. Input mode is where the plain-path loss in
[`02-semantics.md`](02-semantics.md) §6 is pinned down: with `--files`, a deleted
related file is skipped under `skipIf: missing` and passes `untouched`, where git
mode fails both; `skipIf: exists` and `expect: missing` behave the same in both.

Also covered here: `--files-from` line parsing (a tab means name-status; an
unknown status is an error naming the line) and a non-ASCII filename, which must
match its rule (OS16).

## Corpus

- the four rewritten presets and the documented recipes
- the 51 scenarios in
  [`experiments/sim/scenarios.yaml`](experiments/sim/scenarios.yaml), promoted
  to real projects; each already states its expected outcome, but the simulator
  and the scenarios share an author, so they are a draft, not an oracle
- one invalid config per validation rule in [`04-grammar.md`](04-grammar.md),
  asserted to be rejected with exit 2 and the expected message

The invalid configs matter as much as the valid ones. Every validation rule is new or changed and none has been exercised by real
code.

## Exit criteria

1. The determinism guard passes.
2. Layer 0 passes, including every invalid-glob and invalid-template case.
3. Layer 1 passes for all 96 cells, with the positive assertions above.
4. Layer 2 passes across all eight surfaces and all three headlines.
5. Every invalid corpus config is rejected with the expected message.
6. Goldens are recorded once, after 1 to 5, as a reviewed change.

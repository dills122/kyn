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
| Binding | the longest-first rule: `{name}.{kind}.ts` on `user.profile.component.ts`; the same with a `**/` before it; length counted in code points; a capture with `**` or `{**x}` on both sides is rejected, and a random-glob run (`-ties`) finds no two longest parses that bind different values; every row cross-checked against an enumerating reference matcher, which is the definition |
| Captures | `{name}` within a segment; leading dot; dotted values; a value of `.` or `..`; a lazy `*` next to a greedy capture (`src/*-{n}.go`) |
| Characters | a non-ASCII literal; `?` against one multi-byte character; NFD input against an NFC glob does not match; `*` and `?` match a leading dot |
| `{**name}` | zero segments; one; many; trailing form needs one or more; a file at the root |
| Alternation | `{ts,html}` is not captured; `{ts}` is a capture; order of alternatives never changes a binding |
| `**` | `src/**` does not match the file `src`; `**/x` matches `x`; a bare `**` matches every path |
| Several globs | agreeing bindings; **disagreeing bindings are an error**; the same name as `{x}` and `{**x}` is an error in either order; reordering the list changes nothing |
| Exclude | removes a source; also stops a related path counting as a self-match |
| Instance names | one case per row of [`03-instances-and-reports.md`](03-instances-and-reports.md) §2, including a root file next to a directory literally named `(root)` |
| Invalid globs | one case per load-time error in [`04-grammar.md`](04-grammar.md): unclosed or unmatched brace, nested braces, `[`, `\`, wildcard or empty alternative, partial-segment `**` or `{**x}`, `.`/`..`/empty segment, `{..,a}` and `{.,a}`, leading or trailing `/`, a `/` inside `{}`, whitespace inside `{}` (`{ts, html}`), edge whitespace, a control character or BOM, invalid name, `dir`, duplicate capture, a capture touching another capture, `*`, `?` or an alternation, capture in exclude, two `{**x}`, exclude identical to a match |
| Invalid templates | empty, glob characters, `\`, `.` or `..`, absolute, trailing `/`, unknown variable, adjacent variables, two variables that can contain `/`, `{dir}` or `{**x}` in part of a segment, a template made only of those, unused capture, a constant path inside its own match or inside its match minus its exclude |
| Resolution | one-pass substitution (a value of `{dir}` stays literal); empty segments dropped; a resolved path that is empty, has a `.`/`..` segment, a `\` or a drive prefix is an error |

The simulator's scenarios are the starting point. They are a draft of
the expected results, not an oracle.

## Layer 1: semantics, exhaustive

| Dimension | Values | Count |
| --- | --- | --- |
| `skipIf` | omitted, `missing`, `exists` | 3 |
| `expect` | `touched`, `untouched`, `exists`, `missing`, `none`, `[exists, touched]` | 6 |
| related path | `exists`+touched, `exists`+untouched, `removed` by delete, `removed` by rename, `missing`+untouched, `missing`+touched | 6 |

3 × 6 × 6 = 108, less the 24 cells that rule 7 rejects, leaves **84 cells**. The
rejected combinations are `skipIf: exists` with `touched`, `exists` or
`[exists, touched]` (the rule could never pass) and with `missing` (it could never
fail); the other `expect` values chosen here reach no rejected combination. Rule 7 itself gets its own table: every `skipIf` × every `expect`
list, checked against an independent enumeration of the five reachable pairs. Each runs in git mode and is asserted on `explain` JSON, `check` JSON
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
[`02-semantics.md`](02-semantics.md) §6 is pinned down, both halves: a deleted
related file that is **not listed** is skipped under `skipIf: missing` and passes
`untouched`, where git mode fails both; one that **is listed** passes `touched`
and fails `untouched`, where git mode fails both. `skipIf: exists` and
`expect: missing` behave the same in both.

Also covered here: `--files-from` line parsing (a tab means name-status; an
unknown status, including `C`, is an error naming the line; identical repeated
lines are harmless; conflicting lines for one path are errors; so are a `..`
segment, a backslash, a drive prefix, surrounding whitespace, invalid UTF-8 and a
quoted path; blank lines and CRLF are tolerated; a name with a space; a TAB-less line starting with a status token and a space is rejected; a rename score above 100 is rejected; a zero-width character at a segment edge is rejected and one in the middle is legal) and a
non-ASCII filename, which must match its rule (OS16).

Resolve-time rules 9 to 11 get a case each where `changeType` or `skipIf` would
skip the instance, because they must still fire. The four plain-path
differences in [`02-semantics.md`](02-semantics.md) §6 are asserted as a table,
including a plain-mode source that is not on disk creating no instance while the
same source in git mode creates one.

## Corpus

- the four rewritten presets and the documented recipes
- the 150 scenarios in
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
3. Layer 1 passes for all 84 cells, with the positive assertions above.
4. Layer 2 passes across all eight surfaces and all three headlines.
5. Every invalid corpus config is rejected with the expected message.
6. Goldens are recorded once, after 1 to 5, as a reviewed change.

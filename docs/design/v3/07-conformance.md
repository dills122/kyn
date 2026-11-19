# Conformance suite

Status: proposed. Current state only; history is in
[`notes.md`](notes.md).

The suite that proves an implementation matches
[`02-semantics.md`](02-semantics.md), [`03-instances-and-reports.md`](03-instances-and-reports.md)
and [`04-grammar.md`](04-grammar.md). There is no v2 side to compare against;
the oracle is the tables in those documents, computed independently of the
implementation.

## First: determinism

OS7 measured v2 reporting three different errors across thirty identical runs.
It is fixed in v2 (PR #51, sorted iteration) and v3 reuses that code, so the
guard stays. The first test written runs the same
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
| `check` text | the golden-grade example in [`03`](03-instances-and-reports.md) §8: headline, `Instance:`, `Source:`, `Related:`, `Expected:`, `Reason:`, `Message:`, counts including `Skipped`, the `Input:` line; the exact generated messages of `02` §9 |
| `check` JSON | every key of the `03` §8 examples: `ok`, `headline`, `inputMode`, `removalInfo`, `summary` (and what `evaluated`, `errors`, `warnings` count), top-level `flags` (also under `--summary-only`), `rulesWithoutInstances`, per result `reason`, `sources`, `related`, `relatedState`, `touched`; `--summary-only`; the empty run |
| SARIF, RDJSON, Checkstyle | failures only; `(instance: …, <reason text>)` suffix; `reason`, `relatedState`, `expectedFiles` and `relatedLocations` in SARIF; `why` as rule `fullDescription`; the synthetic `kyn/fail-on-empty` diagnostic |
| `explain` text and JSON | `SkipIf:` / `Expect:` sections, clause strings, the state a `skipIf` saw, the `trace` object, list-valued `expect`, a `changeType` skip |
| exit codes | 0, 1, 2 and 3 each reached: a bad git ref and a path missing from the tree (3), a bad config and bad user input (2) |
| `--dry-run-resolve` | per rule: count (zeros included), instances, sources, related path, state; exit code with `--fail-on-empty` |
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
| Document | `version` as a string, a float, `2` or absent; `rules` absent, `null`, `{}`, a list or a scalar; an unknown top-level key; a bad rule ID; a related-less rule with a capture; a list-valued `related`; YAML anchors (OS15) |
| Resolution | one-pass substitution (a value of `{dir}` stays literal); empty segments dropped; a resolved path that is empty, has a `.`/`..` segment, a `\` or a drive prefix is an error |

## Reference algorithm

The oracle for layer 0 and layer 1 is this, written independently of the
implementation. (A throwaway simulator that modelled matching existed at commit
`3918223`; it predates the `removed` and working-tree rules and is not the oracle.)

1. **Parses.** A glob compiles to segments: literal text, `*`, `?`, `{a,b}`, a
   capture, `**`, `{**name}`. A path matches by enumerating **every** way the
   segments cover its code points; each way is a parse.
2. **Binding.** Among a path's parses, keep those whose capture values are
   longest comparing the first capture first, then the second. Totality follows
   from the load-time rejection of a capture between two `**`. Several globs must
   agree on the binding of a path (rule 9).
3. **State.** `removed` if the change record deleted or renamed away the path;
   else `exists` if it is in the working tree; else `missing`. `touched` is
   membership in the change set (added, modified, rename destination).
4. **Instance.** Group source files by the values of exactly the variables
   `related` uses; resolve `related` in one pass; apply rules 9 to 11.
5. **Evaluate.** `changeType` first, then `skipIf` over the state, then each
   `expect` over (state, touched), AND-combined; the reason is the most
   fundamental cause in the `02` §5 order.

## Layer 1: semantics, exhaustive

| Dimension | Values | Count |
| --- | --- | --- |
| `skipIf` | omitted, `missing`, `exists` | 3 |
| `expect` | `touched`, `untouched`, `exists`, `missing`, `none`, `[exists, touched]` | 6 |
| related path | `exists`+touched, `exists`+untouched, `removed` by delete, `removed` by rename, `missing`+untouched | 5 |

3 × 6 × 5 = 90, less the 20 cells that rule 7 rejects, leaves **70 cells**. The
rejected combinations are `skipIf: exists` with `touched`, `exists` or
`[exists, touched]` (the rule could never pass) and with `missing` (it could never
fail), each across the five related-path rows; the other `expect` values chosen
here reach no rejected combination, and `none` is exempt. Rule 7 itself gets its
own table: every `skipIf` × every `expect` list (including `none` and
`[touched, missing]`), checked against an independent enumeration of the four
reachable pairs. Each runs in git mode and is asserted on `explain` JSON, `check`
JSON and the exit code.

The related-path rows are the reachable combinations of git mode from
[`02-semantics.md`](02-semantics.md) §4. `missing`+touched is unreachable there:
the fixtures for it are the **error** cases (below), and a plain-path layer covers
the one place it exists.

Fixtures outside the matrix, each its own test:

- **Stale checkout, deletion.** The same diff deleting a related file, evaluated
  from a tree that has it and from one that does not, gives identical results for
  every `expect` value and both `skipIf` values.
- **Stale checkout, addition.** The same diff adding a file, evaluated from a tree
  that lacks it, exits 3 (git) or 2 (`--name-status`) with the "check out head"
  error naming the path, for a source and for a related file.
- **Subdirectory `--cwd`.** A monorepo with the config in `svc/`: git mode with
  `--relative` matches; a `--cwd` below the root without it is the failure the
  flag exists to prevent.
- **Plain paths.** The four rows of `02` §6, with `missing`+touched for a listed
  related file that is not on disk.

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

Also covered here: input parsing. Without `--name-status`, a TAB line is an error
and `M foo.go` is an ordinary name; with it, a plain line is an error. Status
lines: an unknown status, including `C`, is an error naming the line; `U` says to
resolve the merge; a rename score above 100 is rejected; identical repeated lines
are harmless and conflicting lines for one path are errors. Paths: a `..`
segment, a backslash, a drive prefix, surrounding whitespace, invalid UTF-8 and a
quoted path are rejected **when a rule matches the file or a related path
resolves to it**, and ignored otherwise; a deleted path or rename source with an
edge blank is accepted and a destination with one is not; a leading `"` is
rejected for `--files-from` and `--files` and accepted in git mode; blank lines
and CRLF are tolerated; a name with a space; a zero-width character is legal anywhere
in a name, and a glob with one at an edge never matches (documented, not rejected). A non-ASCII filename must match
its rule (OS16).

Resolve-time rules 9 to 11 get a case each where `changeType` or `skipIf` would
skip the instance, because they must still fire. The four plain-path
differences in [`02-semantics.md`](02-semantics.md) §6 are asserted as a table,
including a plain-mode source that is not on disk creating no instance while the
same source in git mode creates one.

## Corpus

- the four rewritten presets and the documented recipes
- one invalid config per validation rule in [`04-grammar.md`](04-grammar.md),
  asserted to be rejected with exit 2 and the expected message

The invalid configs matter as much as the valid ones. Every validation rule is new or changed and none has been exercised by real
code.

## Exit criteria

1. The determinism guard passes.
2. Layer 0 passes, including every invalid-glob and invalid-template case.
3. Layer 1 passes for all 70 cells, with the positive assertions above, plus the
   stale-checkout fixture: the same diff deleting a related file, evaluated from
   a tree that has it and from one that does not, gives identical results for
   `touched`, `untouched`, `exists`, `missing` and both `skipIf` values.
4. Layer 2 passes across all eight surfaces and all three headlines.
5. Every invalid corpus config is rejected with the expected message.
6. The status-bearing container pipeline (`git diff --name-status -M | kyn check
   --stdin --name-status`) fails on a deleted test; the same pipeline with
   `--name-only` and no flag fails with the declared-format error; `--stdin` and
   `--files-from -` give byte-identical output. A name containing `"` is read
   correctly in git mode.
7. The document root cases are rejected with exit 2: no `rules`, `rules: null`,
   `rules: {}`, an unknown top-level key, a bad rule ID, `version: 2`.
8. The empty-run and preview contracts of `03` §8 and §9, including
   `--dry-run-resolve --fail-on-empty` exiting 1 on zero matches.
9. Release acceptance in [`06-cutover.md`](06-cutover.md) passes.
10. Goldens are recorded once, after 1 to 9, as a reviewed change.

# Fresh Review Bootstrap — v3 matching section, round 4

Review instance: 8. A narrow review of what review 7's fixes introduced, and a
freeze-readiness verdict.

## Why this review exists

Review 7 found no P0 and one P1 (a plain-path row contradicting the derivation
that was meant to win) and four P2s. Closing them: replaced the earliest-start
binding tie-break with a rejection (a capture, `{name}` or `{**name}`, may not
have `**` or `{**name}` on both sides in one glob); rewrote
`02-semantics.md` §6 around three stated facts; made rule 7 also reject a rule
that can never fail (layer 1 is now 84 cells, not 90); introduced one
blank-character predicate shared by globs, `related`, input paths and resolved
paths; and made the `--files-from` line grammar exact. Every review so far
found that the previous round's fixes introduced defects, and the author also
wrote the simulator that "confirms" them. If this pass finds only P3s, the
matching section will be frozen.

## Repository And Branch

- Branch: `design/v3-config`, open as draft PR #50
- Design state under review: `582442e` (this bootstrap is the commit after it)
- Review read-only. Scratch work goes in a fresh directory under $TMPDIR.

## Primary Targets

1. `docs/design/v3/04-grammar.md`: Matching intro (blank character), Globs and
   the rejected table, Captures and **Binding** (including the new rejection),
   Several globs, Exclude, `related`, Validation rules 1 to 11, the recipe, limits
2. `docs/design/v3/02-semantics.md` §3, §4, §5, §6 (line grammar, the three
   facts, the four rows) and §7
3. `docs/design/v3/experiments/sim/main.go` and `scenarios.yaml` (134
   scenarios). `parses`/`bind` are the claimed reference; `passability` is rule 7;
   `badPath` is the shared predicate; `ties` (run with `-ties N`) is the claimed
   evidence that binding is total
4. `docs/design/v3/00-workplan.md` D66 to D70 and the review 7 disposition table
5. `docs/reviews/v3-config-independent-review-7.md`, to confirm each finding is
   really closed and not reworded
6. Compatibility: `internal/matcher/`, `internal/changes/`, the four presets,
   `e2e/projects/*/kyn.config.yaml`, `testdata/`, `docs/site/`

Everything else in `docs/design/v3/` is context.

## Method Expected

**Attack the specification, not the prose.**

1. Read `04-grammar.md` and `02-semantics.md` first. Without opening the
   simulator, write down what an implementer must still decide, then write at
   least twenty adversarial cases (globs, change sets, input files).
2. Run them through the simulator (copy it and your scenarios into scratch;
   `go run` from the repository root so it finds its YAML dependency).
3. **Independent binding check.** Write your own rune-based enumerating matcher
   from the Binding paragraph alone and diff it against `bind` on generated
   globs and paths. Separately, **attack the claim that the new rejection makes
   binding total**: find a glob the simulator accepts that has two longest parses
   binding different values, by construction and by your own generator, not by
   running `-ties`. Also check the rejection is not wider than needed (a glob a
   developer would reasonably write that it rejects).
4. **Independent rule 7 and predicate checks.** Rewrite rule 7 from the text and
   diff it against `passability`, including the new can-never-fail clause and the
   84-cell claim. Rewrite the clean-path predicate from the text and diff it
   against `badPath` and the glob/related checks on generated strings, including
   Unicode white space, NBSP, BOM, zero-width characters, combining marks, and
   control characters.
5. Treat every scenario as the author's expectation, not evidence.

## Specific Scepticism Requested

1. **Is binding total, and is the rejection right?** Any accepted glob with two
   longest parses binding different values? A capture between a `**` and a
   `{**x}` in either order, `*` or `?` interacting, several `{name}` captures in
   one segment, alternation, a glob list where each glob is fine alone. Is the
   rule too wide, rejecting `src/**/{name}/{**rest}`-style globs people write?
2. **The three facts.** Derive every plain-mode behavior yourself from the facts.
   Does any row still contradict them? Is fact 3 (drop listed sources not on
   disk) coherent with mixed files, renames, `changeType`, and rule 9 to 11? Is
   "applies per line" actually what the simulator does?
3. **Rule 7 now rejects can-never-fail.** Is the rejection right for every
   `skipIf` and `expect` list, or does it wrongly reject something useful (for
   example a rule whose point is a flag)? Does the 84-cell count hold? Does the
   plain-mode vacuity sentence match the model?
4. **The blank-character predicate.** Is it one predicate or several that drifted?
   Names with interior spaces, NBSP, ideographic space, zero-width joiner,
   variation selectors, a name that is legal on Linux/macOS but now rejected. Is
   rejecting these a regression for real repositories, and is it in `06`?
5. **The line grammar.** Status `R` with scores (`R0`, `R1000`), `T`, a path
   containing a tab in plain mode, mixed files, duplicate and conflicting lines
   across status and plain forms, a rename whose destination is also a plain line,
   CRLF and a lone CR.
6. **Claimed closures.** For each of D66 to D70 and each row of the review 7
   disposition table, does the text and the simulator deliver it? Look for a
   disposition that says fixed where it is not.
7. **Overclaims.** Any "all", "never", "exactly" or "complete" in the changed
   text. Test it.
8. **Freeze readiness.** Say plainly whether the matching section could be
   frozen as is, frozen after specific P3 edits, or not, and what would have to be
   true.

## Out Of Scope

Report formats, the conformance suite layout, the cutover inventory beyond the
compatibility claim, wording of field names, and implementation.

## Verdict Expected

A readiness verdict for the matching section only. Findings ranked P0/P1/P2/P3,
each with a concrete input that demonstrates it, the required fix, and file:line
evidence; labeled *wrong*, *unspecified*, or *matter of taste*. Say plainly where
an area holds up, including any review 7 finding you judge genuinely closed. List
the cases you tried, including those that behaved correctly, and report the
results of all differential tests.

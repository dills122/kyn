# Fresh Review Bootstrap — v3 matching section, round 5 (freeze review)

Review instance: 9. A narrow review of what review 8's fixes introduced, and
the freeze decision for the matching section.

## Why this review exists

Review 8 found no P0 or P1, two P2s and P3 wording. Closing them: a TAB-less
line that starts with a status token and a space (`D x`, `M  x`, `?? x`,
`R100 a b`) is now rejected in plain and mixed input; format characters (Unicode
Cf, such as the zero-width space) are blank at a segment edge and inside `{}`
and legal in the middle of a segment; a rename score is 0 to 100; the "clean
path" predicate is defined once in `04-grammar.md` (Matching) and referenced
elsewhere; the whole-run abort on odd-but-legal tracked names is disclosed in
`02` and `06`; the capture-between-`**` rejection is kept and reworded; rule 7
and fact 3 wording was corrected. This is meant to be the last review. If it
returns only P3s the matching section is frozen.

## Repository And Branch

- Branch: `design/v3-config`, open as draft PR #50
- Design state under review: `f7c43d9` (this bootstrap is the commit after it)
- Review read-only. Scratch work goes in a fresh directory under $TMPDIR.

## Primary Targets

1. `docs/design/v3/04-grammar.md`: Matching (clean path, blank character),
   Globs and the rejected table, Captures and Binding, `related`, Validation
   rules 1 to 11, the recipe, limits
2. `docs/design/v3/02-semantics.md` §3 to §7, especially the `--files-from`
   grammar, the status-like-line rule, the three facts and row 4
3. `docs/design/v3/06-cutover.md` row `internal/changes` (compatibility claims,
   including "no file in this repository is affected")
4. `docs/design/v3/experiments/sim/main.go` and `scenarios.yaml` (150 scenarios,
   116 onward are new): `isBlank`, `badPath`, `statusLike`, the rename-score check
5. `docs/design/v3/00-workplan.md` D71 to D75 and the review 8 table
6. `docs/reviews/v3-config-independent-review-8.md`, to confirm each finding is
   closed and not reworded

Everything else in `docs/design/v3/` is context.

## Method Expected

**Attack the specification, not the prose.** You have the benefit of eight
earlier rounds; do not re-derive what they established. Spend effort where the
newest edits are.

1. Read `04-grammar.md` and `02-semantics.md` first; write down what an
   implementer must still decide.
2. **Attack the new rejections for false positives and false negatives.** The
   status-like rule rejects plain lines starting with `A `, `M `, `D `, `T `,
   `U `, `C `, `X `, `R<digits> `, `?? `, `!! `: find a realistic tracked path
   that is now wrongly rejected, and a status-looking line it still lets through
   (other whitespace after the token, lower case, a quoted path, `git diff
   --stat`, `git ls-files -s`, `git log --name-status` output, a line with a
   leading space). Attack Cf-as-blank: a real name or glob it wrongly rejects
   (emoji ZWJ sequences, Arabic, Indic scripts, Mongolian, right-to-left marks,
   soft hyphen) and an invisible character it still lets through (Mn variation
   selectors, tag characters U+E0000, bidi isolates, U+2800 braille blank, U+115F
   fillers, U+3164, U+00AD, U+034F).
3. **Check the single definition.** Is "clean path" really one predicate across
   `04`, `02`, rule 11 and the model? Diff the model's `badPath`, `isBlank` and
   `hasHidden` against a predicate you write from the text, on generated strings.
4. **Re-check totality and the 84-cell count quickly** with your own code; they
   were closed in review 8 and should not have changed.
5. **Check every changed number and claim**: 150 scenarios, thirteen then seven
   mutations, "two million tried, about one million accepted", "no file in this
   repository is affected", the scenario numbers cited in the workplan.
6. Treat every scenario as the author's expectation, not evidence.

## Specific Scepticism Requested

1. Does the status-like rejection create a new silent failure or a new false
   alarm that is worse than what it fixed?
2. Is the Cf rule consistent for globs, `related`, input paths and resolved
   paths, including inside `{}`?
3. Is anything in the matching section still contradictory between `02`, `04`,
   `06`, `07` and the workplan after eight rounds of edits (stale counts, stale
   rule numbers, stale cross references, a sentence that describes an earlier
   design)?
4. Is the disclosed whole-run abort acceptable as a stated limit, or does it hide
   a worse case?
5. **Freeze readiness.** State plainly: freeze as is, freeze after listed P3
   edits, or not ready and why. If not ready, say the smallest change that would
   make it ready.

## Out Of Scope

Report formats, the conformance suite layout, the cutover inventory beyond the
compatibility claims above, wording of field names, and implementation.

## Verdict Expected

A freeze decision for the matching section, with findings ranked P0/P1/P2/P3,
each with a concrete input, the required fix and file:line evidence, labeled
*wrong*, *unspecified* or *matter of taste*. List the cases you tried, including
those that behaved correctly, and the results of your differential tests.

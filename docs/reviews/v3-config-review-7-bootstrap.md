# Fresh Review Bootstrap — v3 matching section, round 3

Review instance: 7. A narrow review of what review 6's fixes introduced.

## Why this review exists

Review 6 found no P0 but two P1s (binding not total with a bare `**` around a
capture; the plain-path "whole list" wrong a third time) and nine P2s. Closing
them added an earliest-start binding tie-break, a derivation (not a list) for
plain-path modes with four rows, a general rule 7 checked by enumerating the five
reachable (state, touched) pairs, code-point matching, more dialect rejections,
whole-segment `{dir}`/`{**x}` in `related`, rule 11 rejecting `\` and a drive
prefix, and a stated compatibility break in input cleaning. Every review so far
found that the previous round's fixes introduced defects, and the author of these
fixes also wrote the simulator that "confirms" them. This pass is scoped to that
new material.

## Repository And Branch

- Branch: `design/v3-config`, open as draft PR #50
- Design state under review: `580896e` (this bootstrap is the commit after it)
- Review read-only. Scratch work goes in a fresh directory under $TMPDIR.

## Primary Targets

1. `docs/design/v3/04-grammar.md`: Matching intro, Globs and the rejected table,
   Captures and **Binding**, Several globs, Exclude, `related`, Validation rules
   1 to 11, the `git ls-files` recipe, and the limits list
2. `docs/design/v3/02-semantics.md` §3, §4, §6 (input cleaning, statuses, the
   plain-path derivation and its four rows) and §7
3. `docs/design/v3/experiments/sim/main.go` and `scenarios.yaml` (115
   scenarios, 93 onward are new), as the author's model. `bind` is claimed to be
   the normative reference; `passability` is claimed to implement rule 7
4. `docs/design/v3/00-workplan.md` D59 to D65 and the review 6 disposition table,
   to check each claimed closure against the text and the model
5. `docs/reviews/v3-config-independent-review-6.md`, to confirm each finding is
   really closed and not just reworded
6. Today's input handling for the compatibility claims: `internal/matcher/`,
   `internal/changes/`, and the glob strings in the four presets,
   `e2e/projects/*/kyn.config.yaml`, `testdata/`, `docs/site/`

Everything else in `docs/design/v3/` is context.

## Method Expected

**Attack the specification, not the prose.**

1. Read `04-grammar.md` and `02-semantics.md` first. Without opening the
   simulator, write down what an implementer must still decide, then write at
   least twenty adversarial (glob, path), (config, change set) and (input file)
   cases.
2. Run them through the simulator (copy it and your scenarios into a scratch
   directory; `go run` from the repository root so it finds its YAML
   dependency). Where the model does something the text does not license, or the
   text promises something the model does not do, that is a finding.
3. **Write your own independent matcher** from the Binding paragraph alone (no
   reading `bind`), rune-based, and diff it against the simulator on generated
   globs and paths, now including non-ASCII, `{**p}`, and bare `**`. Also write an
   independent checker for rule 7 from the text and diff it against
   `passability` over every `skipIf` and every `expect` list.
4. Treat every scenario as the author's expectation, not evidence.

## Specific Scepticism Requested

1. **Is binding now total?** Equal lengths then earliest start, first capture
   first. Can two different parses tie on both and bind different values? Does
   the comparison order (all lengths before any start) ever surprise? Does
   `{**p}` with an empty value, or a capture inside a glob list, break it?
2. **Is the plain-path derivation correct?** Two facts (the list is the whole
   change set, all `modified`; nothing is `removed`). Derive behavior yourself for
   rename, a related file that is a rename destination and also listed, a listed
   source absent from disk, and `changeType`. Does any row contradict the
   derivation, or does the derivation miss a case? Is row 4 (git keeps absent
   sources, plain drops them) coherent with `changeType` and rule 9 to 11?
3. **Rule 7 by enumeration.** Is the five-pair set really the reachable set? Is
   "can never pass" or "passes only on the disagreement pair" the right test, or
   are there other rules that are vacuous, always pass, or contradict
   `skipIf`? Does the 90-cell layer 1 count still hold?
4. **Input cleaning.** Is the "clean path" predicate consistent between input
   paths and resolved paths? What does a Windows-style or macOS-specific name do?
   Is rejecting `\` and a drive prefix a regression for any real repository, and
   is the compatibility break complete in `06-cutover.md`? CRLF, blank lines,
   statuses, whitespace, a name with a tab.
5. **Code points.** Is "code points, no normalization" fully specified (invalid
   UTF-8 in a glob, in input, combining marks, case)? Does the model do what the
   text says for `?`, `*`, length, and literals?
6. **New rejections.** `{..,a}`, whitespace, `/` inside braces, whole-segment
   `{dir}`: anything wrongly rejected that a real config uses, or still accepted
   that never matches? Is each message accurate?
7. **Claimed closures.** For each of D59 to D65 and each row of the review 6
   disposition table, does the text and the simulator actually deliver it? Look
   for a disposition that says fixed where it is not.
8. **Overclaims.** The author has now been wrong several times about "the whole
   list" and "complete". Find any such claim in the changed text and test it.

## Out Of Scope

Report formats, the conformance suite layout, the cutover inventory beyond the
compatibility claim, wording of field names, and implementation.

## Verdict Expected

A readiness verdict for the matching section only. Findings ranked P0/P1/P2,
each with a concrete input that demonstrates it, the required fix, and file:line
evidence; labeled *wrong*, *unspecified*, or *matter of taste*. Say plainly where
an area holds up, including any review 6 finding you judge genuinely closed. List
the cases you tried, including those that behaved correctly, and report the
results of both differential tests.

# Fresh Review Bootstrap — v3 configuration design, revised shape

Review instance: 4. Reviews 1 to 3 covered earlier shapes; the grammar was
revised on 2026-11-13, after review 3, and that revision has had no independent
review.

## Review Objective

Assess whether the revised v3 design is ready to be approved as a specification
and enter implementation, judged on two things equally:

1. **Is it correct and complete enough to build from?**
2. **Is it pleasant and unambiguous to write?** The revision's stated goal is
   that a developer can read a rule aloud. Judge whether it achieves that.

This is a design review. There is no v3 implementation.

## Repository And Branch

- Branch: `design/v3-config`, open as draft PR #50
- Design state under review: `172ebcf6ac9b35d1d1585ba715d3fa8145025926` (this bootstrap is the commit after it)
- Review read-only.

## Primary Targets

`docs/design/v3/`: `04-grammar.md` first, then `02-semantics.md`,
`03-instances-and-reports.md`, `06-cutover.md`, `07-conformance.md`,
`01-observed-semantics.md`, `00-workplan.md`, and `experiments/` including
`experiments/sim/`.

## Method Expected

- Run the experiment scripts; they measure the shipped v2 binary.
- Run the simulator and read its source. It is the author's model of the v3
  design, not an oracle. Where the simulator and the documents disagree, or
  where a scenario passes only because the author wrote both, that is a finding.
- **Write configs cold.** Before reading the scenarios, take five or six
  policies a real team would want and write them from `04-grammar.md` alone.
  Report where you hesitated, guessed, or got it wrong.

## Specific Scepticism Requested

1. **Captures in `match`.** `{name}` versus `{a,b}` share brace syntax and
   are told apart by a comma. Greedy matching with adjacent captures, several
   globs binding the same names, whether `{**name}` should match zero
   segments, names containing dots. Is the capture grammar specified well
   enough to implement without guessing?
2. **Removing `patterns`.** The set claims every v2 capability maps without
   it. Try to find one that does not.
3. **`skipIf`.** Three states, two values. Is "a file removed by this change
   is never missing" stated clearly enough that an author would predict the
   behavior? Is `skipIf: exists` with a removed file right?
4. **Instance names.** The uniqueness argument rests on validation rule 5. Try
   to break it.
5. **Capability accounting.** The set has claimed a complete list of losses
   twice and been wrong twice. It now claims two limits and one gain without
   claiming completeness. Look for more of either.
6. **Consistency.** Review 3's P1s were mostly decisions that had not reached
   every document. The set was rewritten so each document states only the
   current design. Check that no document contradicts another, and that the
   decision log matches the documents.
7. **The conformance suite.** 108 cells. Is the count right this time, and
   would the suite catch a regression in each decision in force?
8. **Wording.** `touched`, `untouched`, `skipIf`, `changeType`, `flags`,
   `why`. Anything misleading, overloaded, or likely to be misread.

## Out Of Scope

The onboarding rewrite (PR #49) and implementation.

## Verdict Expected

A readiness verdict, findings ranked P0/P1/P2 with a concrete failure scenario,
required fix and file:line evidence each, and a clear separation of *wrong*,
*unspecified*, and *matter of taste*. Say so plainly where an area holds up.

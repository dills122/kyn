# Fresh Review Bootstrap — v3 matching section, round 2

Review instance: 6. A narrow review of what review 5's fixes introduced.

## Why this review exists

Review 5 found the matching section could not be implemented from the text
(seven P1s). Closing them replaced the glob matcher with a defined dialect,
defined capture binding formally, added resolve-time rule 11, reordered the
resolve-time checks, changed the plain-path rules, and added several load-time
rejections. Reviews 4 and 5 each found that the previous round's fixes
introduced new defects, and the author of these fixes also wrote the simulator
that "confirms" them. This pass is scoped to that new material.

## Repository And Branch

- Branch: `design/v3-config`, open as draft PR #50
- Design state under review: `6737d73` (this bootstrap is the commit after it)
- Review read-only. Scratch work goes in a fresh directory under $TMPDIR.

## Primary Targets

1. `docs/design/v3/04-grammar.md`: Globs, Captures (including **Binding**),
   Several globs, Exclude, `related`, and the Validation tables (rules 1 to 11)
2. `docs/design/v3/02-semantics.md` §3 and §6 to §7 (input modes, `--files-from`,
   the plain-path differences)
3. `docs/design/v3/experiments/sim/main.go` and `scenarios.yaml`, as the
   author's model. The `bind` method is claimed to be the normative reference.
4. `docs/design/v3/00-workplan.md` D51 to D58 and the review 5 disposition table,
   to check each claimed closure against the text and the model
5. Today's matcher, for compatibility: `internal/matcher/` and
   `github.com/bmatcuk/doublestar/v4`, and the glob strings in the four presets,
   `e2e/projects/*/kyn.config.yaml`, `testdata/`, `docs/site/`

Everything else in `docs/design/v3/` is context. Review 5 itself is in
`docs/reviews/v3-config-independent-review-5.md`.

## Method Expected

**Attack the specification, not the prose.**

1. Read **only** `04-grammar.md` first. Without opening the simulator, write down
   what an implementer must still decide, then write at least twenty
   adversarial (glob, path), (glob, related) and (config, change set) cases.
2. Run them through the simulator (copy it and your scenarios into a scratch
   directory; `go run` from the repository root so it finds its YAML
   dependency). Where the model does something the text does not license, or the
   text promises something the model does not do, that is a finding.
3. **Write your own independent matcher** from the Binding paragraph alone (no
   reading `bind`), and diff it against the simulator on generated globs and
   paths. Disagreement means the text is ambiguous or the model is wrong.
4. Treat every scenario as the author's expectation, not evidence.

## Specific Scepticism Requested

1. **Is the longest-first binding rule total and deterministic?** Ties, `{**name}`
   zero versus many segments, a capture bordered by `/` on both sides, a
   capture with literal text that can match several positions, `**` before or
   after a capture. Does "longest, first capture first" ever disagree with what
   a developer would expect, and does the border restriction actually remove
   every case where wildcards influence the binding?
2. **Is the dialect list complete and each rejection justified?** Anything
   doublestar accepts that v3 now neither supports nor rejects (silent
   narrowing remaining)? Anything rejected that a real config in the repository
   uses? Unicode, case, a leading dot, a glob that is a single `*`, `{a,b}` at a
   segment boundary, a capture with an empty value.
3. **Rule 11 and one-pass resolution.** Can a resolved path still be wrong,
   escape `--cwd`, collide with another instance's path, or depend on the
   operating system (backslash, case, trailing dots, `:` on Windows)? Are
   empty-segment dropping and the `.`/`..` check in the right order?
4. **Resolve-time ordering.** Rules 9 to 11 now run before `changeType` and
   `skipIf`. Is the claim "a skipped instance cannot hide a broken rule" true?
   What about instances that exist only in sources dropped in plain mode, and the
   `git ls-files` recipe: what exactly does it exercise now?
5. **Plain-path differences.** Is the list of three complete and correct? Try a
   rename, a copy, a source listed but not on disk with a related file listed
   too, a path on conflicting lines, and an identical repeated line.
6. **Validation completeness and redundancy.** A config that passes rules 1 to 8
   and still cannot be evaluated or can never apply; a rule that is now
   redundant or contradicts another; the new `skipIf: exists` with `touched`
   rejection (is it right for every state?); "exclude identical to a match glob"
   (what about near-identical?).
7. **Claimed closures.** For each of D51 to D58, does the text and the simulator
   actually deliver it? Look for a disposition that says fixed where the text
   does not.

## Out Of Scope

Report formats, the conformance suite layout, the cutover inventory, wording of
field names, and implementation.

## Verdict Expected

A readiness verdict for the matching section only. Findings ranked P0/P1/P2,
each with a concrete input that demonstrates it, the required fix, and file:line
evidence; labeled *wrong*, *unspecified*, or *matter of taste*. Say plainly where
an area holds up, including any review 5 finding you judge genuinely closed.
List the cases you tried, including those that behaved correctly, and report the
result of your differential test.

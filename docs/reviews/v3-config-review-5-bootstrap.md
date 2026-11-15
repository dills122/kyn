# Fresh Review Bootstrap — v3 matching section

Review instance: 5. A narrow review of what review 4's fixes introduced.

## Why this review exists

Review 4 found the capture grammar could not be implemented without guessing,
and two semantics were wrong. Closing those findings added a normative matching
section and changed `skipIf` and `untouched`. All of that was written and
checked by one author, and the simulator that "confirms" it has the same author.
Independent reviews have found the most defects in exactly this kind of
freshly-written, self-checked material, so this pass is scoped to it.

## Repository And Branch

- Branch: `design/v3-config`, open as draft PR #50
- Design state under review: `f9b2914` (this bootstrap is the commit after it)
- Review read-only. Scratch work goes in a fresh directory under $TMPDIR.

## Primary Targets

1. `docs/design/v3/04-grammar.md`, the **Matching** section and the
   **Validation** tables
2. `docs/design/v3/02-semantics.md` §2 to §7
3. `docs/design/v3/03-instances-and-reports.md` §1 and §2 (instance keys and
   names)
4. `docs/design/v3/experiments/sim/main.go` and `scenarios.yaml`, as the
   author's model of the above
5. Today's matcher, for compatibility: `internal/matcher/` and
   `github.com/bmatcuk/doublestar/v4`

Everything else in `docs/design/v3/` is context.

## Method Expected

**Attack the specification, not the prose.** The goal is to find inputs on which
the specification is ambiguous, self-contradictory, or gives an answer a
developer would not predict.

1. Read **only** `04-grammar.md` Matching and Validation first. Without opening
   the simulator, write down for each rule what an implementer must decide that
   the text does not say. Then write at least twenty adversarial
   (glob, path) and (glob, related) cases of your own.
2. Run them through the simulator (copy it and your scenarios into a scratch
   directory; `go run` it from the repository root so it finds its YAML
   dependency). Where the simulator does something the text does not license, or
   the text promises something the simulator does not do, that is a finding.
3. Treat the simulator as the author's model, not an oracle. A scenario that
   passes because the same person wrote the model and the expectation shows
   nothing.

## Specific Scepticism Requested

1. **Is "greedy" well defined?** It is specified for captures. What about a
   capture next to `*`, `?`, `**`, or an alternation (`{name}*.go`,
   `{name}.{ts,tsx}`)? Is the result independent of how an implementer builds
   the matcher (backtracking regexp versus a hand-written matcher)?
2. **Dialect compatibility.** Today's matcher is doublestar, which supports
   character classes, backslash escapes, and wildcards inside brace
   alternatives. The new grammar documents a narrower set and says a literal
   brace cannot be matched. Does that silently narrow or change any glob a
   current config could contain? Is each narrowing stated?
3. **Uniqueness of instance names.** Argue for or against the claim in
   `03-instances-and-reports.md` §2 that the name is unique within a rule. Try
   to construct a collision, including with empty values, `.`, `(all)`, a
   trailing slash, and values containing the join character.
4. **`{**name}` zero segments.** Cleaning removes the empty segment. Check
   every place a zero-segment capture can appear in `related` (leading,
   middle, trailing, doubled, next to a literal) and whether the resolved path
   is what a developer would expect and can never escape `--cwd`.
5. **The three states and `skipIf`.** Is every combination of state, change
   membership, and input mode accounted for, including a file that is both
   deleted and re-added, a rename onto an existing path, a copy, and a
   case-only rename on a case-insensitive filesystem?
6. **Validation completeness.** Is there a config that passes every load-time
   rule and still cannot be evaluated, evaluates ambiguously, or can never
   apply? Is any rule redundant or contradictory with another?
7. **Resolve-time errors.** Two errors only fire when a matching file changes.
   Is the suggested `git ls-files | kyn check --files-from - --dry-run-resolve`
   check sufficient to surface them, given plain-path mode?

## Out Of Scope

Report formats, the conformance suite, the cutover inventory, wording of field
names, and implementation.

## Verdict Expected

A readiness verdict for the matching section only. Findings ranked P0/P1/P2,
each with a concrete input that demonstrates it, the required fix, and file:line
evidence; labeled *wrong*, *unspecified*, or *matter of taste*. Say plainly
where an area holds up. List the cases you tried, including those that behaved
correctly, so the coverage can be judged.

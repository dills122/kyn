# Fresh Review Bootstrap — v3 configuration design

Review instance: 3 of 3.

## Review Objective

Assess whether the v3 configuration **design set** is ready to be approved as a
specification and enter implementation. This is a design review; there is no v3
implementation to approve.

Reviews 1 and 2 are recorded and every finding from both is settled. The
question for review 3 is whether the resulting design is coherent, complete
enough to build from, and correctly grounded in the measured behavior it claims.

## Repository And Branch

- Branch: `design/v3-config`
- HEAD at freeze: `ab29520c71b5225db75d2a4b5bb156ed6a2b6783`
- Open as [PR #50](https://github.com/dills122/kyn/pull/50), draft.
- The checkout is clean. Review it read-only.

## Primary Targets

`docs/design/v3/`, in dependency order:

1. `00-workplan.md` — finding register, gates, decisions D1–D26
2. `01-observed-semantics.md` — measured v2 baseline, OS1–OS13
3. `02-semantics.md` — G1, expectation grid and the deletion gate
4. `03-instance-and-identity.md` — G2, instance key and report identity
5. `04-grammar.md` — G3, grammar and validation rules
6. `05-compatibility-matrix.md` — G3, superseded as a constraint
7. `06-cutover.md` — G4, v1/v2 removal and preset rewrite
8. `07-conformance.md` — G5, conformance suite
9. `experiments/` — 13 reproduction scripts

## Context, Not Targets

- `docs/reviews/v3-config-proposal.md` — the original exploratory proposal.
  Superseded in most particulars; read it for intent, not for current design.
- `docs/reviews/v3-config-independent-review-1.md` — review 1.
- Review 2 is a comment thread on PR #50, summarised in `00-workplan.md`.

## Repository Evidence To Inspect

`AGENTS.md`, `.codex/steering/`, `docs/decisions.md`, `docs/spec.md`,
`internal/config/`, `internal/family/`, `internal/rules/`, `internal/report/`,
`internal/changes/`, `internal/cli/init.go`, and the corresponding tests.

## Method Expected

Every claim about *current* behavior in the design set is supposed to be backed
by a script in `experiments/`. Run them rather than trusting the prose:

```bash
go build -o /tmp/kyn ./cmd/kyn
KYN=/tmp/kyn bash docs/design/v3/experiments/e3-truthtable.sh
```

Treat a claim whose script does not reproduce as a finding.

## Specific Scepticism Requested

Weight these over general commentary:

1. **Footprint-derived instance keys (D17, D22)** are the newest and least
   exercised idea in the set. Is the uniform-footprint restriction sound, is the
   rendering genuinely unique, and does the model stay coherent for templates
   nobody has tried?
2. **The tri-state gate (D23)** replaced a boolean that turned out to be
   misnamed. Is the replacement complete across all three states, both gates,
   and both input modes?
3. **Report identity (D16, D21)** renames wire fields and asserts sort order is
   preserved. Is anything still projecting a value that no longer exists?
4. **Capability loss.** The set has twice claimed a complete list of what v3
   cannot express, and been wrong twice — atomic multi-path rules, then
   informational rules. Look for a third.
5. **The G5 conformance suite (07)** is the only thing standing between this
   design and a silent regression. Would it actually catch one? The 75-cell
   count was corrected once already.
6. **Self-assessment risk.** Reviews 1 and 2 were both answered by the same
   author who wrote the design. Findings were accepted readily; check whether
   any were accepted too readily, or resolved in a way that suits the existing
   design rather than the problem.

## Out Of Scope

- The onboarding documentation rewrite, which is [PR #49](https://github.com/dills122/kyn/pull/49).
- Implementation. It is not authorized and no gate assumes it is.

## Verdict Expected

A readiness verdict on advancing this design set into a formal specification and
implementation plan — not on release readiness.

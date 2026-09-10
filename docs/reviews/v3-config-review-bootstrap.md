# Fresh Review Bootstrap

Review instance: 1 of 3.

## Review Objective

Independently assess whether the exploratory Kyn v3 configuration pivot is a
sound, sufficiently complete direction for a future specification. Review the
flat rule-centric common form, optional local pattern reuse, version-specific
normalization architecture, compatibility boundaries, and unresolved advanced
semantics. Identify architectural traps, semantic gaps, migration risks,
misleading vocabulary, and simpler or stronger alternatives.

This is a design review. There is no v3 implementation to approve.

## Repository And Worktree

- Saved repository: `/Users/dsteele/go/src/kyn`
- Source branch: `codex/usability-onboarding-review`
- HEAD at freeze: `0ce1bd58cfa2fe45194b0179e96e34528fc05f1a`
- The fresh review task is expected to run in an isolated Codex worktree created
  from the source working tree so the uncommitted proposal is present.
- Review the task's current checkout read-only. Do not modify it or create an
  additional worktree.

## Base, Head, Branch, And Dirty State

- Base and head for committed repository behavior are both the frozen HEAD
  above.
- The proposal is an explicit working-tree review target rather than a commit
  diff.
- The checkout is dirty. At preparation time it contained modified
  `docs/README.md`, `docs/site/getting-started.md`, `docs/site/index.md`, and
  `mkdocs.yml`, plus untracked content under `docs/reviews/`.
- Do not assume every dirty path belongs to this v3 review.

## In-Scope Commits And Paths

Primary proposal target:

- `docs/reviews/v3-config-proposal.md`

Repository evidence to inspect:

- `AGENTS.md`
- `.codex/steering/repository-steering.md`
- `docs/reviews/v0.1.3-onboarding.md`
- `docs/site/config.md`
- `docs/site/concepts.md`
- `docs/decisions.md`
- `docs/spec.md`
- `docs/mvp-v2.md`
- `internal/config/config.go`
- `internal/config/validate.go`
- `internal/config/migrate.go`
- `internal/family/`
- `internal/rules/`
- `internal/report/`
- representative starter configs in `internal/cli/init.go`
- relevant config, evaluator, resolver, migration, report, and end-to-end tests

The proposal must be judged against current behavior and repository boundaries,
not against the author’s recollection.

## Canonical Requirements And Plan

- `AGENTS.md` and `.codex/steering/repository-steering.md` define current product
  and compatibility boundaries.
- `docs/site/config.md` and `docs/site/concepts.md` are canonical shipped user
  documentation.
- Deterministic ordering, slash-normalized repository-relative paths, stable
  report contracts, and exit codes 0/1/2/3 are compatibility invariants.
- `docs/reviews/v3-config-proposal.md` is the exploratory plan under review.
- The proposal is not approved and intentionally contains open decisions.

## Explicit Exclusions

- Do not review or approve the onboarding documentation prototype as part of
  this instance.
- Do not treat the existing v0.1.3 onboarding review as a v3 implementation.
- Do not implement v3, edit documentation, stage changes, create commits, or
  start another review instance.
- Do not broaden v3 into daemon, PR integration, plugin, remote registry,
  content-analysis, or monorepo-graph work.
- Do not issue release readiness for a product implementation that does not yet
  exist.

## Verification Commands Available To Reviewer

Use proportionate non-mutating inspection. Useful commands include:

```bash
git status --short
git branch --show-current
git rev-parse HEAD
git diff -- docs/reviews/v3-config-proposal.md
go test ./internal/config ./internal/family ./internal/rules ./internal/report
go test ./e2e
```

Tests establish current behavior only; they cannot validate an unimplemented v3
design by themselves.

## Author Explanation Location Or Delivery Step

Perform and record a preliminary review of the proposal and repository evidence
before reading:

`docs/reviews/v3-config-author-explanation.md`

Then reconcile the author’s claims against the evidence.

Use `$independent-review` in reviewer mode. This is review instance 1 of 3. Work
from the Fresh Review Bootstrap first and record a preliminary review before
reading the Author Explanation. Then verify the explanation against the
repository, review both the implementation and its plan, run proportionate
non-mutating checks, and return an evidence-backed verdict. Do not implement
fixes, create further review instances, or split the work into new workstreams.

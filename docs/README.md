# Kyn Documentation Index

Repository documentation for Kyn users, maintainers, and historical design work.

## Live Docs Site

- [GitHub Pages](https://dills122.github.io/kyn/) is the canonical user
  documentation.
- The published site is built only from [`docs/site/`](site). Keep installation,
  concepts, recipes, CLI reference, and troubleshooting current there.

## Maintainer Guides

- [release-notes.md](release-notes.md): commit and PR conventions for meaningful generated release notes
- [release.md](release.md): release artifacts, container image, and install flow
- [migration-v1-to-v2.md](migration-v1-to-v2.md): migration strategy and command usage

## Active Design Work

Unapproved. Nothing in this section describes shipped behavior.

- [design/v3/](design/v3/): v3 configuration design — workplan, gates, and measured baseline
- [design/v3/01-observed-semantics.md](design/v3/01-observed-semantics.md): experimentally measured v2 semantics, including two shipping defects
- [v3 configuration proposal](reviews/v3-config-proposal.md): the exploratory proposal under review
- [v3 independent review 1](reviews/v3-config-independent-review-1.md): first independent readiness verdict
- [v3 review 3 bootstrap](reviews/v3-config-review-3-bootstrap.md): frozen scope for the third independent pass

## Historical Design and Planning

- [v0.1.3 onboarding review](reviews/v0.1.3-onboarding.md): usability findings, reproduced first-run behavior, and proposed documentation improvements
- [related-file-policy-exploration.md](related-file-policy-exploration.md): unapproved candidates for deeper related-file policy capabilities
- [spec.md](spec.md): original product specification; not the current CLI reference
- [ci.md](ci.md): earlier CI examples retained for repository history
- [presets.md](presets.md): earlier preset adoption notes
- [troubleshooting.md](troubleshooting.md): earlier troubleshooting notes
- [decisions.md](decisions.md): locked MVP decisions
- [cli-validation-matrix.md](cli-validation-matrix.md): earlier input-mode validation notes, predating auto git mode and the additional output formats
- [mvp-tasks.md](mvp-tasks.md): original MVP backlog
- [mvp-v2.md](mvp-v2.md): v2 proposal and design direction
- [ergonomics.md](ergonomics.md): v2 adoption/ergonomics contract referenced by mvp-v2.md; not all of it shipped (see issue tracker for gaps)
- [mvp-v2-tasks.md](mvp-v2-tasks.md): execution backlog for v2
- [mvp-v2-finish-tasks.md](mvp-v2-finish-tasks.md): final v2 closeout checklist

## Working norm

Update `docs/site/` from executable behavior and tests. Treat design documents
and completed plans as historical evidence, not shipped behavior.

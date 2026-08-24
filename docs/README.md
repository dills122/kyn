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

## Historical Design and Planning

- [related-file-policy-exploration.md](related-file-policy-exploration.md): unapproved candidates for deeper related-file policy capabilities
- [spec.md](spec.md): original product specification; not the current CLI reference
- [ci.md](ci.md): earlier CI examples retained for repository history
- [presets.md](presets.md): earlier preset adoption notes
- [troubleshooting.md](troubleshooting.md): earlier troubleshooting notes
- [decisions.md](decisions.md): locked MVP decisions
- [cli-validation-matrix.md](cli-validation-matrix.md): valid and invalid input mode combinations
- [mvp-tasks.md](mvp-tasks.md): original MVP backlog
- [mvp-v2.md](mvp-v2.md): v2 proposal and design direction
- [mvp-v2-tasks.md](mvp-v2-tasks.md): execution backlog for v2
- [mvp-v2-finish-tasks.md](mvp-v2-finish-tasks.md): final v2 closeout checklist

## Working norm

Update `docs/site/` from executable behavior and tests. Treat design documents
and completed plans as historical evidence, not shipped behavior.

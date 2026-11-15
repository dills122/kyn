# Independent review 2 of the v3 configuration design

| | |
| --- | --- |
| Source | a comment on [PR #50](https://github.com/dills122/kyn/pull/50), 2026-09-19 |
| Reviewed | `ad03945`, branch `design/v3-config` |
| Dispositions | `docs/design/v3/00-workplan.md`, section Reviews |

Written by the maintainer after an independent pass. Reproduced as posted.

---

## Review text

This remains a draft specification. Review 2 found open contracts, so the PR body's statement that G1–G4 are closed, G5 is specified, and no questions remain is now stale. Please keep this PR unmerged until these decisions are recorded:

1. **Shape identity:** decide whether rule and pattern IDs share a namespace or internal shape references carry a kind. A same-named inline rule and pattern currently collide.
2. **Instance identity:** define one stable `instanceName` when rules sharing a pattern have different constant `related` paths, and specify how footprint values form the key and sort order.
3. **Existence gate:** specify added, deleted, renamed-away, and never-present related paths for both `related-existed` and `related-absent`. `existsNow OR vanishedInChange` is not literal existence at base.
4. **Cutover capability:** decide whether v3 supports v2 informational, emit-only rules with no assertion or explicitly retires them. Atomic multi-path rules are not the only currently identified loss.
5. **Manual rename input:** decide how `--files-from` represents both old and new paths; `status<TAB>path` alone cannot encode a Git rename.

Then update the design and G5 conformance oracle, including the actual reachable truth-table cells and report fixtures. The Git-backed experiment scripts should also fail when fixture setup fails; otherwise a failed commit can produce a misleading exit 0. Implementation remains unauthorized until the design is approved.

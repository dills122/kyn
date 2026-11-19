# v3 design notes

Status: design only; nothing is implemented. This file is the condensed record
behind the specs. The long form (decision log D1 to D80, nine independent reviews
recorded verbatim, 17 measurement scripts, a 172-scenario simulator) is in the git
history of PR #50 (`design/v3-config`, last full commit `3918223`).

## Measured v2 behavior (OS)

Observed on the shipped binary. The specs cite these ids.

| Id | Finding | Where addressed |
| --- | --- | --- |
| OS1, OS2 | Assertions read two orthogonal axes: exists-on-disk and in-the-change-set | `02` §1, §4 |
| OS3, OS8 | Deleting or renaming away the related file is invisible, so a rule is silenced by the change that should trip it | `02` §2, §3 |
| OS4, OS11 | `stripSuffixes` decides grouping; the shipped `api` preset contradicts itself | `04` captures, `03` §1 |
| OS5 | The common form matches its own related file, hidden by OS3 | `04` rules 3 and 10 |
| OS6, OS10 | Non-`source` groups are inert; `description` is parsed and never read | removed, or made functional (`why`) |
| OS7 | Error selection was non-deterministic; **fixed in v2** (PR #51), determinism guard kept | `06`, `07` |
| OS9 | Result count is rules × instances | `03` §1 |
| OS12 | A report can print `PASS` above `Rules failed: 1` | `03` §6 |
| OS13 | The instance key over-partitions when `related` ignores the file name | `03` §1 |
| OS14 | Informational rules affect severity counters | `02` §5, §8 |
| OS15 | YAML anchors work through the strict loader | `04` sharing globs |
| OS16 | Non-ASCII filenames never matched any rule (git quoting); **fixed in v2** (PR #52, `-z`) | `02` §6, `06` |

## Decisions that shape the specs

- **Rule-centric.** `rules:` is a mapping keyed by ID (D3). No `patterns`, no
  `use`; rules repeat their `match`, and YAML anchors cover sharing (D28).
- **`match` binds variables.** `{name}` within a segment, `{**name}` across
  segments, `{dir}` built in. `stripSuffixes`, `{base}`, `{ext}`, `{file}` are
  gone (D27). `expect` is required (D36).
- **Instance key = the variables `related` uses** (D17), so a constant related
  path gives one failure, not one per source; this makes OS11 impossible and fixes
  OS13. Names are the values joined by `/`; `.` at the root; `(all)` for none
  (D31, D44).
- **Three related states: `exists`, `removed`, `missing`** (D23). A removal never
  causes a skip (D40); `untouched` fails on a removed file (D43). Deletion data
  comes from git output Kyn already runs (D5).
- **Input.** `--files-from` accepts `git diff --name-status -M` lines (D24); git
  mode reads paths with quoting disabled (D48).
- **Reports.** `ruleId` + `instanceName`; skips are visible (D32, D46); headlines
  `FAIL` / `NON-BLOCKING` / `PASS` (D18). Messages are a fixture-tested contract
  (D7).
- **One `related` per rule** (D15). Atomic multi-path rules, a glob-valued
  `related`, and forbidding edits with no related path are out; the grammar's
  limits list is the honest accounting (D33, D47).
- **Own glob dialect, narrow and loud** (D51). It is not doublestar. Anything
  outside the table is rejected at load by name rather than silently narrowing.
  Matching is on Unicode code points, no normalization (D61).
- **Binding is total by rejection** (D66, D75): longest captures win, first capture
  first; a capture between two `**` is rejected, so no tie exists.
- **The change record decides `removed`** (review: Codex, after review 9): a
  path the diff deletes or renames away is `removed` whatever the working tree
  holds, so a stale checkout cannot hide a deletion; `skipIf` reads the same
  state.
- **CI feeds status lines.** `--stdin` is `--files-from -`; enforcement examples
  use `git diff --name-status -M`; plain paths are for local trials. Git mode
  reads `git diff --name-status -z -M` (lossless); `T` counts as `modified`.
- **Results carry a stable `reason`** beside the authored message; the preview
  prints a count per rule, zeros included, and `--fail-on-empty` applies to it.
- **Resolve-time rules 9 to 11 run before `changeType` and `skipIf`** (D55), and
  rule 11 (resolved path must be clean) stops the whole run.
- **Rule 7 is general** (D62, D68): enumerate the four reachable (state, touched)
  pairs, reject a rule that can never pass or never fail. Layer 1 conformance is
  70 cells (90 less 20).
- **Plain-path modes follow three facts** (D60, D67): the list is the whole change
  set, all `modified`; nothing is `removed`; a listed source not on disk is dropped.
- **One clean-path predicate** (D69, D72): blank = White_Space, Cc, Cf (Unicode
  15.0); applied to globs, `related`, input paths and resolved paths. Input
  cleaning is a stated compatibility break (D64); a bad changed path aborts the
  run only where a rule would match it or a related path resolves to it (D74,
  narrowed by review 10), except removed paths and rename sources (D77).
- **Input format is declared** (review 10, replacing D71 and D76): `--name-status`
  says a stream is status lines; without it every line is a path and a TAB line is
  an error. No sniffing, no status-token list. The report records `inputMode`.
- **Git mode requires the head to be checked out**: an added or modified path
  absent from the tree is an error, and git is read with `--relative -z`.

## Review history

| Review | Verdict | Main finding |
| --- | --- | --- |
| 1, 2, 3 | advance, with conditions | decisions had not reached every doc; three related states; skips invisible |
| 4 | not approvable until 5 P1s closed | capture grammar under-specified; `skipIf: exists` contradicted its prose |
| 5 | not implementable from the text (7 P1) | binding undefined, silent glob narrowing, capture binding `..` escaped `--cwd` |
| 6 | 2 P1, 9 P2 | binding not total; plain-path differences wrong a third time |
| 7 | 1 P1, 4 P2 | plain-path row contradicted its derivation; loose line grammar |
| 8 | no P0/P1, 2 P2 | status line with spaces read as a path; zero-width chars not blank |
| 9 | no P0/P1, 1 P2 | two-letter git status codes still read as paths; fix-PR blocked by abort |
| 10 (whole design) | no P0/P1, 11 P2 | added related file still machine-dependent; subdirectory `--cwd` silent pass; format sniffing; rule 7 text rejected `none`; report fields dropped; cutover inventory incomplete |

Review 10 also suggested cuts that are **not taken**: the Cf and Unicode-pinned
blank rules, list-valued `expect`, and rule 7 as a forbidden table. Each removes
complexity but narrows what v3 can say; decide them before implementation, not by
default.

Recurring failure mode: "exactly N" or "complete" claims, and fixes that opened
new gaps. The last fixes (D76 to D80) were not independently re-reviewed or
mutation-tested.

## Open items

1. OS7 and OS16 are fixed in v2 (PRs #51 and #52). OS11 (`api` preset) and OS12 (`PASS` above a failure count) remain in v2 and are resolved by the cutover.
2. The `api` preset uses focused tests per source (`06`).
3. A glob-valued `related` is deferred (D47).
4. `docs/site/getting-started.md` teaches v2 and needs a v3 rewrite at cutover.
5. A JSON Schema ships once the grammar freezes (D20).

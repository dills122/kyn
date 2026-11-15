# v3 configuration: workplan and decision log

Status: design only. Nothing is implemented and implementation is not
authorized.

This document is the history. The other documents in this directory describe
the current design and nothing else.

## Where things stand

| Area | State | Document |
| --- | --- | --- |
| Measured v2 baseline | 16 observations; 14 have a script, OS2 is derived from OS1 and OS10 is a source search | [`01`](01-observed-semantics.md) |
| Semantics | Revised after review 5 (input modes) | [`02`](02-semantics.md) |
| Instances and reports | Revised after review 4 | [`03`](03-instances-and-reports.md) |
| Grammar and matching | Rewritten after review 5 with a defined dialect and binding; **not yet independently reviewed** | [`04`](04-grammar.md) |
| Cutover inventory | Measured and extended | [`06`](06-cutover.md) |
| Conformance suite | Specified, with a matcher layer; 90 semantic cells | [`07`](07-conformance.md) |
| Scenario simulator | 87 scenarios, each asserting its outcome; matcher is an enumerating reference; mutation-tested | [`experiments/sim/`](experiments/sim/) |
| v1/v2 compatibility matrix | Historical evidence only | [`05`](05-compatibility-matrix.md) |

Five independent reviews have run (recorded in
[`../../reviews/`](../../reviews/)). Review 4 judged the revised shape ready to
advance once five P1 findings were closed, and closing them added the matching
section. Review 5 attacked that section and found it was not implementable from
the text: seven P1 findings, all reproduced. They are closed below (D51 to D58)
and the matching section was rewritten around a defined dialect. Closing them
changed the grammar again, so the next gate is a sixth review of the revised
section.

## Two changes of direction

**2026-09-10 — no users.** Kyn has no real users yet. A bad bug is fixed in
place; anything breaking goes into v3, and a full cutover is acceptable. This
removed every decision that existed only to protect compatibility: version
gating, migration, and keeping the old report field names.

**2026-11-13 — developer delight.** Three review rounds had made the design
correct and the config unpleasant: to write the common rule an author needed
`when: related-existed`, `expect: in-change-set`, `patterns`, `use`,
`stripSuffixes`, and the difference between `{base}` and `{name}`, backed by 15
validation rules. The shape was revised against one test: can a developer read
the rule aloud. The four presets went from 39/27/22/22 non-blank lines in v2 to
13/7/7/8. The semantics from the earlier rounds were kept; the surface changed.

## Decisions in force

| ID | Decision | Basis |
| --- | --- | --- |
| D1 | The normalized model is ordered slices sorted by ID, never Go maps | OS7 |
| D3 | `rules` is a YAML mapping keyed by ID | `yaml.v3` already rejects duplicate keys with line numbers |
| D5 | The deletion gap is closed, for every config | OS3, OS8; the data is already in the git output Kyn runs |
| D7 | Generated messages are a versioned, fixture-tested contract containing only the rule ID and resolved related path | Absolute paths would leak the CI checkout root |
| D8 | A related path that matches its own rule's `match` is a config error, detected at resolve time, exit 2 | OS5; a static check is glob intersection |
| D15 | v3.0 has one `related` per rule; atomic multi-path rules are out | No config uses one |
| D17 | A rule's instance key is the values of the variables its `related` template uses | OS13; also makes OS11 impossible |
| D18 | Headlines are `FAIL`, `NON-BLOCKING`, `PASS`; `PASS` means no failures | OS12 |
| D19 | `docs/decisions.md` and `docs/migration-v1-to-v2.md` are bannered as superseded, not deleted | This set cites them |
| D20 | A JSON Schema ships once the grammar freezes, generated from or CI-checked against the loader | A drifting schema approves configs the binary rejects |
| D23 | A related path is in one of three states: `exists`, `removed` (by this change), `missing` | Review 2: the earlier boolean was true for a file added in the change |
| D24 | `--files-from` accepts `git diff --name-status -M` lines verbatim, rename lines included | Review 2: two columns cannot hold a rename |
| D25 | Experiment scripts never `cd`; every fixture git call is scoped and checked; `set -e` is not used | Review 2: a failed fixture could commit to the invoking repository |
| D26 | Informational rules are written `expect: none`; `severity` must be `info`; at least one of `message` or `flags` | Review 2; OS14 |
| D27 | `match` binds variables: `{name}` within a segment, `{**name}` across segments; `{dir}` is the only built-in. `stripSuffixes`, `{base}`, `{name}`, `{ext}`, `{file}` are gone | 2026-11-13. Names the variable where it is matched; a 50-line prototype resolved every preset shape; `{**name}` adds mirrored trees, which v2 could not express |
| D28 | No `patterns` block and no `use`. Rules repeat their `match` | 2026-11-13. Patterns gave only de-duplication and a grouping label; every v2 capability maps without them; YAML anchors already work (OS15); it removed 9 of 15 validation rules |
| D29 | Wording: `expect: touched` / `untouched`; `skipIf: missing` / `exists` in place of `when`; `changeType` in place of `on`; `flags` in place of `emit`; `why` in place of `description`. (The meaning of `skipIf: exists` on a removed file was revised by D40) | 2026-11-13. `skipIf` stays three-way because flags are added whenever a rule applies (OS14), so "flag only when absent" needs a gate |
| D30 | `related` is optional when `expect: none`; `skipIf` requires `related` | 2026-11-13. Scenario 11: a flag-only rule had to invent a related file, which then matched its own glob |
| D31 | Instances belong to a rule. `instanceName` is the variable values joined with `/` in the order they appear in `related`; `(all)` with no variables. `related` may use at most one variable that can contain `/`. (The root rendering was revised by D44) | 2026-11-13, review 3. Keeps names unique within a rule |
| D32 | `check` reports skipped results: a count in text and JSON always, entries in JSON, entries in text with `--show-passes`. (Extended to `changeType` by D46) | Review 3: skips were invisible outside `explain` |
| D33 | Capability accounting names what was found and does not claim completeness. (The list grew in D47) | Review 3: the list had been wrong twice |
| D34 | The conformance suite extends `e2e/e2e_test.go` and `e2e/projects`; it does not build a second harness | Review 3 |
| D35 | Reports drop `shapeId`/`familyId`; results carry `ruleId` and `instanceName` and sort by them. `explain` clauses use the field names (`skipIf.missing`, `expect.touched`) | 2026-11-13, following D28 |
| D36 | `expect` stays required; there is no default assertion | 2026-11-13. Explicit over implicit |
| D37 | Matching is specified normatively in `04`: captures are greedy, single-segment, with names `[A-Za-z_][A-Za-z0-9_]*` and `dir` reserved; `{**name}` fills whole segments; adjacent captures, captures in `exclude`, glob characters, `..` and absolute paths in `related` are rejected; a literal brace cannot be matched | Review 4 P1-1: the capture grammar could not be implemented without guessing |
| D38 | A path matching two `match` globs with different bindings is a resolve-time error. List order never decides | Review 4 P1-1: first-match-wins made reordering a list change a policy |
| D39 | A capture that `related` never uses is a validation error | Review 4 P1-2: `{ts}` is a literal in today's matcher and a capture in v3, so trimming `{ts,html}` to `{ts}` would silently change a rule. Also catches typos |
| D40 | **A removal never causes a skip.** `skipIf: missing` skips only `missing`; `skipIf: exists` skips only `exists`; `removed` applies under both. Supersedes the part of D29 under which `skipIf: exists` skipped a removed file | Review 4 P1-3: the prose said "act only when absent" and the table skipped a removed file. Each value now matches the state of the same name; v2's `if.kinMissing` behavior on a just-deleted file is preserved; one plain-path divergence disappears |
| D41 | `{**name}` matches zero segments, as `**/` does; a trailing `{**name}` matches one or more | Review 4 P1-4: one-or-more left top-level files silently unchecked, and the only workaround was forbidden |
| D42 | The conformance suite has a matcher layer (layer 0) and 96 semantic cells | Review 4 P1-5. The count was 108 less the 12 cells validation rejected; D58 makes it 90 |
| D43 | `expect: untouched` means unaffected: it fails when the related file was removed by the change | Review 4 P2-5, accepted. Deleting a "frozen" file breaks it; also removes the `[exists, untouched]` workaround |
| D44 | `instanceName` joins the non-empty variable values with `/`; `.` if none remain; `(all)` for a rule with no variables. The root renders as nothing, not `(root)` | Review 4 P2-1: `(root)` collides with a real directory of that name. Supersedes the root rendering in D31 |
| D45 | `related` may not contain `..` or be absolute | Review 4 P2-2: re-creates OS13 by reaching one file from two directories |
| D46 | Instances excluded by `changeType` are reported as `skipped`, with a reason | Review 4 P2-8: otherwise a configured rule that never applies is silent, which is what D32 fixed for `skipIf` |
| D47 | Further capability limits recorded: a glob-valued `related`; a capture limited to alternatives; forbidding edits with no related path; one failure per related file when `related` joins variables with text that can occur in their values. A glob-valued `related` is deferred, not rejected | Review 4 P2-4 and the author's own scenario 39. Decided with the maintainer 2026-11-14 |
| D48 | `--files-from`: a line containing a tab is a name-status line, an unknown status is an error, git-quoted paths are unsupported there. Git mode reads with quoting disabled | Review 4 P2-8 and OS16: today non-ASCII filenames are mangled and never match |
| D49 | Reviews 2, 3 and 4 are recorded verbatim in `docs/reviews/`. The simulator asserts every scenario's outcome and is mutation-tested. `e9` states that it derives state from real git but computes verdicts itself | Review 4 P2-6, P2-10 and its note that review 3 had no record |
| D50 | `changeType` keeps its name and is documented as filtering the matched source, not the related file | Review 4: `expect: touched` beside `changeType: added` can read as "the related file was added". Taste; maintainer confirmed the name on 2026-11-13 |
| D51 | **v3 has its own glob dialect, narrow and loud.** Rejected at load, by name: `[ ]`, `\`, nested braces, wildcards, `/` or empty alternatives inside `{a,b}`, `**` that is not a whole segment, and leading/trailing `/`, empty, `.` or `..` segments. A trailing `/**` needs at least one segment, so `src/**` no longer matches a file named `src`. Decided with the maintainer 2026-11-15 | Review 5 P1-4: doublestar accepts all of these; v3 narrowed them silently, so a ported glob could stop matching without any error. No configuration in this repository uses a rejected form |
| D52 | Binding is defined: among all parses, the one whose captures are longest, first capture first. `*`, `**`, alternation and list order carry no preference. A `{name}` capture may not touch another capture, `*`, `?` or an alternation. The simulator enumerates parses and is the reference | Review 5 P1-1: "greedy" was undefined beside wildcards; `src/*{n}.go` bound `n=c` and alternation order changed bindings. Refines D37 |
| D53 | A capture name has one kind across all globs of a rule: `{x}` everywhere or `{**x}` everywhere | Review 5 P1-2: reordering `match` flipped a config between rejected and accepted, and the accepted order gave a non-unique instance name. Refines D38 |
| D54 | `related` resolves in one pass; empty segments are dropped; the result must be non-empty with no `.` or `..` segment (resolve-time rule 11). A related with no variables is checked against its own match at load | Review 5 P1-3 and P2: a capture can bind `..`, so `{n}/{n}/x.txt` left `--cwd`; sequential substitution rescanned values; an empty `{dir}` left an absolute path |
| D55 | Resolve-time rules 9 to 11 run for every instance before `changeType` and `skipIf` can skip it. The `git ls-files` recipe is documented as a sample of existing paths, with `-c core.quotePath=false` | Review 5 P1-5: a skipped instance hid a rule 10 error from the recipe meant to find it |
| D56 | Plain-path modes drop listed sources that are not on disk. The plain-path differences are three, listed in `02` §6 | Review 5 P1-6: "exactly two differences" was wrong; a deleted source listed in plain mode created a failing instance that git mode does not |
| D57 | `--files-from`: identical repeated lines are harmless; a path on conflicting lines, or a path that is not relative and clean, is an error naming it | Review 5 P2: precedence was order-dependent (`D x` then `A x` gave removed and touched) |
| D58 | More load-time rejections: `skipIf: exists` with `expect: touched` (rule 7, as with `exists`); empty `related` or `match`; `\`, trailing `/` or adjacent variables in `related`; an `exclude` identical to a `match` glob. Layer 1 is therefore 90 cells, not 96 (108 less 18). Undetected and stated: an exclude that covers everything, a `related` that is a directory, case-only renames on case-insensitive filesystems | Review 5 P2. Supersedes the count in D42 |

D29's `why` and `changeType`, and D36, were confirmed by the maintainer on
2026-11-13.

## Superseded decisions

| ID | Was | Replaced by |
| --- | --- | --- |
| D2 | Migration drops inert groups | No migration |
| D4, D6 | Vocabulary `in-change-set` etc. with a `when` gate | D29. The two-field split and "no lenient default" survive |
| D9 | `--files-from` two-column form | D24 |
| D10 | Keep `familyId`/`familyName` on the wire | D16, then D35 |
| D11 | Instances keyed on a shared source shape | D31 |
| D12 | `stripSuffixes` legal inline; a pattern is a named shape | D27, D28 |
| D13 | Clause vocabulary follows the config version | D35; one version exists |
| D14 | `description` survives; version stays `3` | D29 renames it `why`; the version number stands |
| D16 | Rename to `shapeId` / `instanceName` | D35 |
| D21 | Rule and pattern IDs share one namespace | D28; there are no patterns |
| D22 | Uniform footprint per pattern; `stripSuffixes` must be used | D27, D28, D31 |

## Reviews

### Independent review 1

On the original proposal. Recorded in
[`../../reviews/v3-config-independent-review-1.md`](../../reviews/v3-config-independent-review-1.md).

| Finding | Disposition |
| --- | --- |
| Normalized model cannot preserve report identity | D35 |
| Flat rules omit instance grouping semantics | D17, D31 |
| Migration claim broader than representable | No migration; capability limits in D33 |
| Expectation names overstate what Kyn observes | D29 |
| No compatibility matrix | Measured in `05`; no longer a constraint |
| Map-key validation not deterministic | D1, D3 |

### Independent review 2 (2026-09-19)

A comment on [PR #50](https://github.com/dills122/kyn/pull/50), recorded in
[`../../reviews/v3-config-independent-review-2.md`](../../reviews/v3-config-independent-review-2.md).

| Finding | Disposition |
| --- | --- |
| Rule and pattern IDs collide | Moot under D28 |
| `instanceName` undefined for shared patterns with constant paths | D31 |
| The existence predicate was misnamed; second gate unspecified | D23; `02` §3 |
| Informational rules lost | D26, D30 |
| Two columns cannot encode a rename | D24 |
| Experiment scripts hide fixture failures | D25 |

### Independent review 3 (2026-09-19)

A fresh-context review of commit `8efbfb2`, recorded verbatim in
[`../../reviews/v3-config-independent-review-3.md`](../../reviews/v3-config-independent-review-3.md). Verdict: ready to advance,
conditional on five P1 findings. It ran all thirteen scripts then present and
confirmed every table in `01` reproduces. Its findings were mostly decisions
that had not propagated to the cutover and conformance documents, which is why
those documents now describe only the current state.

| Sev | Finding | Disposition |
| --- | --- | --- |
| P1 | Reachable-cell count wrong; the excluded cell is the one separating `touched` from `[exists, touched]` | `02` §4; `07` layer 1 is 108 cells including it |
| P1 | Conformance document not updated for review 2 | `07` rewritten |
| P1 | Gating on one path while asserting on another is a second capability loss | D33; `04` |
| P1 | `check` cannot represent a skipped result | D32 |
| P1 | Cutover inventory incomplete; `e2e/projects` overlaps the proposed harness | `06` re-measured; D34 |
| P2 | OS9 cited but never written | Written into `01` |
| P2 | `instanceName` empty components; root-level files | D31 |
| P2 | Gate prototype still tested the old boolean | `e9` updated; both gates shown |
| P2 | Informational-rule claim had no script | `e15`, OS14 |
| P2 | Plain-path degradation specified for one gate only | `02` §6 |
| P2 | Deduplication decided, deferred and untracked at once | D17 stands; contradictions removed by the rewrite |
| P2 | Superseded decisions cited as delivered | This document separates in-force from superseded |
| P2 | No dimension exercised the headline | `07` layer 2 crosses severity with `--fail-on` |

The review also cleared three things it was asked to attack: the instance-key
model is sound, report identity was complete, and findings from reviews 1 and 2
had not been accepted too readily.

### Independent review 4 (2026-11-14)

A fresh-context review of the revised shape at `73f33dc`, recorded verbatim in
[`../../reviews/v3-config-independent-review-4.md`](../../reviews/v3-config-independent-review-4.md).
It wrote six configs cold from `04` alone before reading the scenarios; three
came out right, one was silently partial, and two were inexpressible. Verdict:
not approvable as a spec yet, no P0, ready to advance once five P1 findings are
closed, all document-level.

| Sev | Finding | Disposition |
| --- | --- | --- |
| P1 | Capture matching not specified well enough to implement | D37, D38; `04` Matching |
| P1 | `{x}` versus `{a,b}` silently reinterprets an existing glob | D39 |
| P1 | `skipIf: exists` skips a removed file, contradicting its prose | D40 |
| P1 | `{**name}` needs one segment, leaving top-level files unchecked | D41 |
| P1 | Conformance suite has no matcher layer | D42; `07` layer 0 |
| P2 | `(root)` breaks instance-name uniqueness | D44; scenario 21 |
| P2 | `..` in `related` re-creates OS13 | D45 |
| P2 | Stale content survived the rewrite (`01`, `05`, citations, severity default, rule numbering) | Fixed |
| P2 | Capability accounting short in both directions | D47 |
| P2 | `untouched` passes for a deleted file; scenario 13 fell into it | D43; scenarios 13b, 13c |
| P2 | Simulator thinner than described: no expected outputs, most validation rules unimplemented, keyed on rendered name | Rewritten: 51 asserted scenarios, every load-time rule, keyed on values; mutation-tested |
| P2 | The named harness cannot run layer 1 | `06` and `07` say what the extension is |
| P2 | Smaller unspecified points (`changeType` invisibility, contradictory lists, removed-failure wording, `--fail-on-empty`, `--files-from` parsing, resolve-time-only errors) | D46, D48; validation rules 5 to 7; `02` §5 to §8; the `git ls-files` check in `04` |
| P2 | Cutover inventory gaps | `06` extended |
| P2 | Evidence overclaims (`e9`; two observations without scripts) | `e9` and `01` reworded; D49 |

Review 4 also reported that it could not verify the dispositions of reviews 2
and 3, because only a summary existed. They are now recorded in `docs/reviews/`.

Findings from the author's own work while closing these: OS16, non-ASCII
filenames never match any rule in shipped v0.1.3 (`e17`); and scenario 39, two
instances demanding one file when `related` joins variables with text that can
occur inside their values (D47).

### Independent review 5 (2026-11-15)

A fresh-context review of the matching section at `f9b2914`, scope frozen in
[`../../reviews/v3-config-review-5-bootstrap.md`](../../reviews/v3-config-review-5-bootstrap.md),
recorded verbatim in
[`../../reviews/v3-config-independent-review-5.md`](../../reviews/v3-config-independent-review-5.md).
Verdict: not ready to implement from the text. Seven P1 findings, each
reproduced against the simulator before it was accepted. It held up the
instance-name uniqueness argument for normalized inputs, zero-segment
`{**name}`, the `skipIf` table, and the messages for rules 3 to 6 and 8.

| Sev | Finding | Disposition |
| --- | --- | --- |
| P1 | "Greedy" undefined beside wildcards and alternation; order changed bindings | D52 |
| P1 | Rules 2 and 3 compared names, not kinds; reordering flipped validity | D53 |
| P1 | A capture can bind `..` and escape `--cwd` | D54, rule 11 |
| P1 | v3 silently narrows doublestar (`[ ]`, `\`, nested braces, wildcards in alternatives, `src/**` on a file) | D51; `04` Globs, `06` |
| P1 | `changeType` skips before rule 10; the `ls-files` recipe cannot exercise it | D55 |
| P1 | "Exactly two plain-path differences" is wrong | D56; `02` §6 |
| P1 | Ordering and "exercise every rule" overclaims | D55; recipe reworded |
| P2 | Conflicting or duplicate `--files-from` lines | D57 |
| P2 | Validation gaps: `skipIf: exists` with `touched`, constant related in own match, exclude covering match, `../x` globs, empty `related`, adjacent variables, backslash | D58, D51 |
| P2 | Sequential substitution; empty leading variable stays absolute | D54 |
| P2 | `git ls-files` quoting | D55 |
| P2 | Case-only renames; related touched but missing under `skipIf: missing` | Stated: `04` limits, `02` §3 |

The simulator grew from 51 to 87 scenarios, one or more per finding, and the
matcher was replaced by an enumerating reference. Nine mutations of the new
decisions each turn their scenarios red.

## Open items

1. **A sixth independent review.** Review 5's fixes added a dialect, a formal
   binding rule, rule 11 and the input-mode changes (D51 to D58), and the last two
   rounds each found that fixes introduced new gaps. Scope it to `04` Globs,
   Captures, `related` and Validation, and `02` §3 to §7.
2. **Two shipped defects, per the rule that a bad bug is fixed in place:** OS7
   (non-deterministic errors) and OS16 (non-ASCII filenames never match).
   Neither depends on v3, and both are in code v3 reuses. Decide whether they
   are fixed in v2 now or at the cutover.
3. The `api` preset must choose between one shared test per handler/service pair
   and one test per file ([`06`](06-cutover.md)).
4. A glob-valued `related` is deferred (D47). Revisit if "some migration" and
   "some changeset" policies turn out to be common.
5. `docs/site/getting-started.md` teaches v2 (PR #49) and will need a v3 rewrite
   at cutover.

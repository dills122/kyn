# v3 configuration: workplan and decision log

Status: design only. Nothing is implemented and implementation is not
authorized.

This document is the history. The other documents in this directory describe
the current design and nothing else.

## Where things stand

| Area | State | Document |
| --- | --- | --- |
| Measured v2 baseline | 16 observations; 14 have a script, OS2 is derived from OS1 and OS10 is a source search | [`01`](01-observed-semantics.md) |
| Semantics | Revised after review 8 (status-like lines, rename score) | [`02`](02-semantics.md) |
| Instances and reports | Revised after review 4 | [`03`](03-instances-and-reports.md) |
| Grammar and matching | Rewritten after review 5, tightened after reviews 6 to 8 (total binding by rejection, code points, validation by enumeration, one blank-character predicate defined once); **the review-8 fixes are not yet independently reviewed** | [`04`](04-grammar.md) |
| Cutover inventory | Measured and extended | [`06`](06-cutover.md) |
| Conformance suite | Specified, with a matcher layer; 84 semantic cells | [`07`](07-conformance.md) |
| Scenario simulator | 150 scenarios, each asserting its outcome; matcher is an enumerating reference over code points; a random-glob run (`-ties`) checks binding is total; mutation-tested | [`experiments/sim/`](experiments/sim/) |
| v1/v2 compatibility matrix | Historical evidence only | [`05`](05-compatibility-matrix.md) |

Eight independent reviews have run (recorded in
[`../../reviews/`](../../reviews/)). Review 4 judged the revised shape ready to
advance once five P1 findings were closed, and closing them added the matching
section. Review 5 attacked that section and found it was not implementable from
the text: seven P1 findings, all reproduced. They are closed below (D51 to D58)
and the matching section was rewritten around a defined dialect. Review 6
found the dialect implementable and no P0, but two P1s (binding not total with
`**` around a capture; the plain-path "whole list" wrong a third time) and nine
P2s, closed below (D59 to D65). Review 7 found one P1 (a plain-path row
contradicting its own derivation) and four P2s, closed as D66 to D70. Review 8
found no P0 or P1, two P2s (a status line with spaces read as a path; zero-width
characters not treated as blank) and a set of P3s, closed as D71 to D75. The next
gate is a ninth, and last, review.

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
| D59 | Binding is total: longest captures first, then **earliest start**, each first capture first. A named capture is greedy and `*` is not a capture (one sentence in `04`). Refines D52 | Review 6 P1-1: `src/**/{n}/**` tied on `auth`/`user` and the simulator silently chose by enumeration order. Rejecting the shape was the alternative; the tie-break is cheaper and `src/**/{feature}/**` is a realistic glob |
| D60 | Plain modes are **derived**, not listed: the list is the whole change set, every path `modified`, nothing `removed`. Four rows follow in `02` §6, including a deleted related file that is also listed (`touched` passes, `untouched` fails) and a git-mode source absent from disk, which is **kept** (an instance is created) where plain mode drops it. Supersedes the list in D56 | Review 6 P1-2 and P2-1: "the whole list" was wrong a third time, and `--name-only` output reopened the delete-to-silence hole |
| D61 | Matching is in Unicode code points: `?` is one, "longest" counts them, no normalization, `*` and `?` match a leading dot, capture names are ASCII. Input must be valid UTF-8 | Review 6 P2-3: the model compared bytes, so non-ASCII literals and `?` never matched |
| D62 | Rule 7 is general: enumerate the five reachable (state, touched) pairs, drop those `skipIf` skips, and reject a rule that can never pass or can pass only on the disagreement pair. The `skipIf: exists` column of the `02` §3 table is removed. Layer 1 was 90 cells; D68 makes it 84. Supersedes the case list in D58 | Review 6 P2-4: `skipIf: missing` with `[touched, missing]` could never pass and was accepted |
| D63 | More dialect rejections: an alternative that can spell `.` or `..`, whitespace-padded globs and segments, `/` inside `{}` (named, not "unclosed"). In `related`, `{dir}` and `{**x}` fill whole segments and a template of only those is rejected. Resolved paths may not contain `\` or start with a drive prefix. Refines D51 and D54 | Review 6 P2-5, P2-6, P2-8 |
| D64 | **Input cleaning is a stated compatibility break.** v3 rejects, with exit 2, input paths with whitespace, `\`, `./`, a drive prefix, invalid UTF-8 or git quoting, for `--files` and `--files-from` alike; CRLF and blank lines are tolerated; `C`, `U` and `X` statuses are errors. The `ls-files` recipe gains `--fail-on-empty` and says it samples tracked, on-disk sources. Refines D57 | Review 6 P2-2, P2-9 |
| D65 | A file name that makes a capture `.` or `..` (`.x.go` under `{n}x.go`) keeps rule 11 **run-level**, stated as a limit. OS-dependent behavior (case-insensitive filesystems, NFD, Windows trailing dots) is listed as undefended. `exclude` identical to a `match` glob stays rejected, reworded as pointless | Maintainer chose run-level on 2026-11-15; Review 6 P2-7, P2-8, P3 |
| D66 | **Binding has no tie-break; ties are removed by rejection.** A capture, `{name}` or `{**name}`, may not have `**` or `{**name}` on both sides in one glob. Supersedes the earliest-start rule of D59. `-ties` in the simulator generates random globs and finds no two longest parses that bind different values in two million accepted globs, and finds thousands with the rejection removed | Review 7 P2-4: `src/**/{feature}/**` bound on directory-name length, so renaming an unrelated directory moved a file between instances. The earliest-start comparison never decided any accepted glob once the rejection existed, so it was dropped rather than kept untested. D59 was wrong that this was a realistic glob |
| D67 | Plain-path differences follow from **three facts**: the list is the whole change set, all `modified`; nothing is `removed`; a listed source not on disk is dropped (a decision, stated as one, limited to sources). The "derivation wins" sentence is deleted. Refines D60 | Review 7 P1-1: row 4 contradicted the derivation, and the simulator followed the row |
| D68 | Rule 7 also rejects a rule that **can never fail** (`skipIf: exists` with `missing`), so layer 1 is **84 cells**, not 90 (108 less 24). `02` stops recommending that combination. The five pairs are git-mode; plain-mode vacuity is stated, not rejected. Supersedes the count in D62 | Review 7 P2-2: an always-pass rule is the same defect class as an informational rule with an error severity |
| D69 | One **blank-character predicate** (white space, BOM, control) for globs, `related`, and input paths: controls and BOM anywhere, no blank at a segment edge, none inside `{}` (`{ts, html}`), and rule 11 applies the same clean-path check to resolved paths. A first segment of a letter and a colon is rejected in `related` at load. `{**,a}` gets its own message. Refines D63 and D64 | Review 7 P2-1, P2-3 and P3s |
| D70 | `--files-from` grammar is exact: each line stands alone (a file may mix status and plain lines); status is `A`, `M`, `D`, `T` with two fields or `R<1-3 digits>` with three; extra fields, `C`, `X` and `U` are errors; whitespace-only lines are ignored; a BOM is an error, not a silent never-match. `--files` cannot carry a comma (`06`). Refines D57 and D64 | Review 7 P2-3 |
| D71 | A TAB-less line that starts with a status token and a space (`D x`, `M  x`, `?? x`, `R100 a b`) is **rejected**, naming the line, in plain and mixed input. A file whose name really starts that way goes through the status form. Refines D70 | Review 8 P2-1: `git status` output or a status line whose TAB became spaces was read as a file called `D src/a.go` and silently evaluated nothing, reopening the delete-to-silence hole for people who thought they were feeding removals |
| D72 | Format characters (Unicode Cf: zero-width space, joiner, BOM) are **blank** at a segment edge and inside `{}`, and legal in the middle of a segment. The predicate is **defined once** in `04` (Matching) as "clean path", and `02`, rule 11 and `06` refer to it. The property is named: `White_Space`. Refines D69 | Review 8 P2-2 and P3-6: a pasted zero-width space made a glob that never matches, and the BOM was the only Cf character rejected. The text phrased the predicate three ways |
| D73 | A rename score is 0 to 100, as git prints it; `C`, `X` and every other status are errors for the stated reason that Kyn expects nothing else. Refines D70 | Review 8 P3-10 and P3-7: `R1000` was accepted, and the `T`/type-change wording was wrong |
| D74 | **Disclosed, not narrowed:** the clean-path check is fatal for every changed path, matched or not, so a tracked name legal on Linux and macOS (`Q: and A.md`, `old /x.txt`, `a\b.txt`) fails every run that touches it. Stated in `02` §6 and `06`; no file in this repository is affected (measured). Maintainer chose disclosure on 2026-11-15 | Review 8 P3-9. Narrowing the check to resolved paths would weaken the guarantee for a rare case |
| D75 | **The capture-between-`**` rule is kept as is**, reworded: it is a little wider than the ties it prevents, and `src/{**pkg}/{name}/**` and `src/**/{name}/{**rest}` are rejected for name-length stability, not because they tie. Rule 7's vacuity sentences list both plain-mode cases (`skipIf: missing` with `expect: missing` never passes, with `expect: exists` never fails); fact 3 states that no resolve-time check runs for a dropped source; the `-ties` count is stated as two million tried, about one million accepted. Maintainer chose to keep the rule on 2026-11-15 | Review 8 P3-1 to P3-5, P3-8. Relaxing it was supported by a random run, not a proof, and would have needed another review |

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

### Independent review 6 (2026-11-15)

A fresh-context review of the matching section at `6737d73`, scope frozen in
[`../../reviews/v3-config-review-6-bootstrap.md`](../../reviews/v3-config-review-6-bootstrap.md),
recorded verbatim in
[`../../reviews/v3-config-independent-review-6.md`](../../reviews/v3-config-independent-review-6.md).
Verdict: not ready to freeze, much closer than review 5. No P0, two P1s. It held
up the dialect table, one-pass resolution with rule 11, the rule 9 to 11 ordering
claim, D53, and the claim that no repository configuration uses a rejected form.
Its own rune-based matcher agreed with the simulator on about 25 million
(glob, path) pairs wherever the answer was unique; the only ties had a bare `**`.

| Sev | Finding | Disposition |
| --- | --- | --- |
| P1 | Binding not total when a bare `**` brackets a capture | D59 |
| P1 | Plain-path "whole list" wrong again; a listed deleted related file passes `touched` | D60; `02` §6 |
| P2 | A git-mode source absent from disk vs plain drop | D60, row 4 |
| P2 | `git ls-files` recipe can pass vacuously; git still quotes tab, `"`, `\`, newline; model split names with spaces | D64; model fixed |
| P2 | Model compared bytes, not code points | D61; model fixed |
| P2 | Validation gaps from enumerating state pairs; stale `skipIf: exists` column | D62 |
| P2 | `{..,a}`, `{.,a}`, whitespace globs, `{a,b/c}` message | D63 |
| P2 | `{dir}.md` and `related: "{dir}"` | D63 |
| P2 | A dotfile aborts the run (rule 11 run-level) | D65, maintainer's choice |
| P2 | OS dependence: case, `\`, drive prefix, NFD | D63, D65 |
| P2 | Input cleaning is a compatibility break; `--files`, CRLF, blank lines | D64; `06` |
| P3 | Exclude-identical wording; lazy `*` vs greedy capture; ASCII names; scenario counts | D65, D59, D61; counts now stated without ranges |

The simulator grew from 87 to 115 scenarios and was rewritten for code points,
tab-separated name-status lines, the tie-break and rule 7 by enumeration. Ten
mutations of the new decisions each turn scenarios red (the first attempt at the
tie-break mutation stayed green, because enumeration order already preferred the
earlier start; the scenario now needs the mutation to prefer a later one).

### Independent review 7 (2026-11-15)

A fresh-context review at `154ffaf` (design `580896e`), scope frozen in
[`../../reviews/v3-config-review-7-bootstrap.md`](../../reviews/v3-config-review-7-bootstrap.md),
recorded verbatim in
[`../../reviews/v3-config-independent-review-7.md`](../../reviews/v3-config-independent-review-7.md).
Verdict: not ready to freeze, but close. No P0, one P1. Its own rune-based
matcher agreed with the simulator on about 20 million pairs (including non-ASCII
and the tie-break), and its independent rule 7 checker agreed on every
combination. It closed D59, D61, D65 and most of D64.

| Sev | Finding | Disposition |
| --- | --- | --- |
| P1 | Plain-path row 4 contradicts the derivation that "wins" | D67 |
| P2 | Whitespace rejection misses `{ts, html}`, `related`, BOM, controls | D69 |
| P2 | `skipIf: exists` with `expect: missing` can never fail, and `02` recommends it | D68 |
| P2 | `--files-from` grammar loose; model laxer than text (mixed files, statuses, blank lines, BOM, `U`, clean-path scope) | D70, D69 |
| P2 | `src/**/{feature}/**` binds on directory-name length | D66 |
| P3 | Stale row reference; `expect: missing` message not in the table; dead `\` check; `[untouched, missing]` mutation survivor; `C:{n}` static; `{**,a}` message; `--files` comma; plain-mode rule 7; empty `{**p}` start | Fixed or stated: `02`, `04`, `06`, scenarios 114c, 113f. The `\` check is now part of one shared predicate, so it is not separately testable |

Two of the author's own claims failed while closing these, and were caught by
the new `-ties` run before the review could: a first version of the rejection
only covered `{name}`, and `{**p}` slid the same way; and the tie-break the
author kept, to cover the rest, turned out never to decide any accepted glob.
The simulator is at 134 scenarios and thirteen mutations of the new decisions
each turn scenarios red (two survivors found on the way, a missing segment-edge
whitespace case and a mutation that only failed to compile, were fixed).

### Independent review 8 (2026-11-15)

A fresh-context review at `300a06e` (design `582442e`), scope frozen in
[`../../reviews/v3-config-review-8-bootstrap.md`](../../reviews/v3-config-review-8-bootstrap.md),
recorded verbatim in
[`../../reviews/v3-config-independent-review-8.md`](../../reviews/v3-config-independent-review-8.md).
Verdict: not freezable as it stood; freezable after two small P2 edits and P3 text
fixes. No P0, no P1. Its own matcher, rule 7 checker and clean-path predicate
agreed with the simulator (about 27 million binding pairs, 2 million random
strings, every `skipIf` and `expect` list), its attack on totality found nothing,
and it re-derived the three facts and the 84-cell count.

| Sev | Finding | Disposition |
| --- | --- | --- |
| P2 | A name-status line with spaces is read as a plain path | D71 |
| P2 | Zero-width (Cf) characters are not blank | D72 |
| P3 | Rejection wider than needed for `{**x}` before a capture | D75, kept and reworded |
| P3 | "two million accepted" overclaim | D75, fixed |
| P3 | Rule 7 sentences: plain-mode vacuity incomplete, example labels wrong | D75, fixed |
| P3 | "as rules 9 and 10 do" ambiguous; rule 11 list drifts from the predicate | `04`; D72 |
| P3 | `C`/`X` wording; row 4 omits resolve-time checks | D73, D75 |
| P3 | Whole-run abort on legal names not disclosed | D74 |
| P3 | `R1000` mutation survivor; uncompiled mutations not counted | D73; scenarios 117 to 117c; mutations redone with a compile check |

The simulator is at 150 scenarios. Seven mutations of the new decisions were
run with a compile check and a panic check; one survivor (the rename-score cap,
which `R1000` alone could not reach because the pattern already rejects four
digits) needed scenario 117c.

## Open items

1. **A ninth, and last, independent review.** Review 8's fixes added a status-like
   line rejection, Cf as blank, a rename-score cap, and a single definition of the
   clean path (D71 to D75). It should confirm those and report whether the matching
   section can be frozen. If it returns only P3s, freeze.
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

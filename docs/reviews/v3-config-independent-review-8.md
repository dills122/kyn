# Independent review 8 of the v3 matching section

| | |
| --- | --- |
| Source | a fresh-context subagent given `docs/reviews/v3-config-review-8-bootstrap.md` and no prior context, 2026-11-15 |
| Reviewed | `582442e` (scope frozen in `300a06e`), branch `design/v3-config` |
| Dispositions | `docs/design/v3/00-workplan.md`, section Reviews |

Verbatim as delivered. Scope was the matching section only. Line citations refer to that commit. The reviewer's scratch files were under its own `$TMPDIR` and are not preserved; the cases it describes are reproducible from the text.

---

# Independent review 8 of the v3 matching section

State reviewed: `300a06e` (design `582442e`), branch `design/v3-config`. Read-only; the repository is untouched. Scratch is `/var/folders/lg/j56llg991d1gpjvtjh8l9gbh0000gn/T/tmp.dBf9PJiuYF`. Subdirectories:
- `s2/` holds a modified copy of the simulator plus my independent `mine.go`.
- `sc/` holds three scenario files (`a.json`, `b.json`, `c.json`, 63 cases in all).
- `mut/` holds the mutation runs.

The author's run is green: 134 scenarios, 0 mismatched.

## Verdict

**The matching section is not freezable exactly as it stands. It is freezable after two small P2 edits and a handful of P3 text edits.** There is no P0 and no P1.

- **Binding totality holds.** I attacked it with my own matcher and generator, with the rejection removed, and with a relaxed rule. I found no accepted glob with two longest parses that bind different values.
- **The three facts hold.** Every plain-mode row derives from them, and "applies per line" is what the simulator does.
- **Rule 7 holds.** The 84-cell count is correct.
- **The line grammar and the blank predicate leave two holes**, and both are silent never-matches, the class the earlier reviews insisted on closing:
  - A name-status line with spaces instead of a TAB is silently read as a plain path (P2-1).
  - A zero-width space (U+200B) at a segment edge is accepted, although the BOM, which is also category Cf, is rejected (P2-2).
- If the author's threshold ("only P3s freezes it") stands, these two P2s must be closed first. Both fixes are a few lines of text and model.

**Freeze condition.** Close P2-1 and P2-2, correct the P3 text (two-million claim, rule 7 sentences, rule 11 list, "as rules 9 and 10 do", "type changes"), and add a scenario for `R1000`. Then I would call the section frozen. No further review is needed unless the author changes the binding rule, for example by relaxing P3-1.

## Findings, ranked

### P2-1: A name-status line with spaces instead of a TAB is silently accepted as a plain path that never matches (*unspecified*)

- **Evidence.**
  - `02-semantics.md:153-156` says a line with no TAB is a plain path. `02-semantics.md:176-178` says interior spaces are fine in a path.
  - Simulator: `main.go:781-788`.
- **Inputs, plain/mixed mode, all accepted with no diagnostic:**
  - `D src/a_t.go` and `M src/a.go` (a TAB converted to spaces, for example by an editor or a pasted log). They become two files with those literal names. Result: PASS with zero instances (my case 11).
  - `git status --short` or `--porcelain` output: `M  src/a.go` and `?? src/b.go`. Result: PASS, nothing evaluated (case C).
- **Why it matters.** This is the same class as the BOM line that review 7 closed. A user who feeds status output, which does carry removals, believes the delete-to-silence hole is closed and gets a silent pass. `--fail-on-empty` only helps if every rule is empty.
- **Fix.** Reject, naming the line, a TAB-less line whose first space-delimited token is a status token (`A`, `M`, `D`, `T`, `U`, `C`, `X`, `R<digits>`, `??`, `!!`) followed by a space. Say so in `02` §6. A tracked file literally named `M foo.go` must then be fed through the status form, which costs one sentence. Alternatively, say plainly that plain lines are never checked and cannot be.

### P2-2: Zero-width and other format characters (Unicode category Cf) are not "blank" (*unspecified*; inconsistent with the BOM)

- **Evidence.**
  - `04-grammar.md:51-55` and `:81`.
  - Model: `main.go:385-400`. `isBlank` is White_Space, BOM or Cc. Its message calls the class "control or invisible character", which is wrong.
  - The BOM (U+FEFF) is itself Cf, so the predicate already special-cases one zero-width character and not its neighbours.
- **Inputs:**
  - `match: "src/{n}.go​"` is accepted. It can never match, and a copy-paste from a web page or chat produces exactly this. Result: PASS, no instances (case 29).
  - `src/{n}​.go` and a leading U+200D or U+2060 are likewise accepted.
  - An input path `src/a.go​` is accepted, so a name differing only by an invisible character is a different file (case F).
- **Fix.** Treat Cf as blank at a segment edge and inside `{}`, in globs, `related` and paths. Keep Cf legal in the interior, because ZWJ emoji sequences need it. Do not extend this to variation selectors or other Mn marks. Case D confirms a directory ending in U+FE0F (`✔️/x`) is accepted today. That is correct and must stay accepted, since it is a real name.
- **Also.** The text says "Unicode white space" without naming the property. Say "Unicode `White_Space`". Java's `isWhitespace`, for example, excludes NBSP.

### P3 findings

1. **The rejection is wider than needed when a `{**x}` is the "before" side (*matter of taste*).**
   - `src/{**pkg}/{name}/**` is rejected (`main.go:330-344`, `04-grammar.md:123-126`).
   - A developer would plausibly write it. It is tie-free: `pkg` is the first capture, so the longest `pkg` wins, and `name` is always the penultimate directory. It does not bind by directory-name length.
   - `**/{**p}/**` and `{**p}/{n}/**/{m}.go` are also rejected and tie-free (exhaustive over 6 words to depth 6: 0 ties).
   - Relaxation: stop counting `{**x}` as a "before" marker and do not check the `{**x}` capture itself. I ran that variant (env `RELAX`) with `-ties 1500000`: 768,025 accepted, 8.4M matching pairs, 0 ties. It accepts `src/{**pkg}/{name}/**`.
   - Keep rejecting `src/**/{name}/**`, which is a real tie, and `src/**/{name}/{**rest}`, which is not a tie because `rest` differs in length. That one is rejected only for the stability reason (it binds on directory-name length), and the text should say so. At `04-grammar.md:119-126` it implies a tie.
   - Relaxing is optional.
2. **Overclaim: "two million random accepted globs" (`04-grammar.md:126-128`, `00-workplan.md:106`) (*wrong*).**
   - `-ties 2000000` prints "2000000 globs tried, 1007980 accepted, 11014745 matching (glob, path) pairs, 0 …". About 1.0M were accepted.
   - The "thousands with the rule removed" claim reproduces: 1,991 hits at 300k tried.
   - Fix the wording.
3. **Plain-mode vacuity sentence is incomplete (`04-grammar.md:296`) (*unspecified*).**
   - It names `skipIf: missing` with `expect: missing` (can never pass in plain mode).
   - In plain mode, `skipIf: missing` with `expect: exists` can never fail: the only applicable pairs are `exists` ones. Git mode does not reject it either, because `removed` fails it. My plain-mode enumeration finds exactly these two vacuous combinations among the accepted ones.
   - Mention both, or reject neither.
4. **Rule 7 example labels (`04-grammar.md:296`) (*wrong*, minor).** `skipIf: exists` with `touched` is not "never pass". It passes only on (missing, touched), the disagreement case, and is rejected for that reason.
5. **"as rules 9 and 10 do" (`04-grammar.md:209`) (*wrong* or ambiguous).** Rules 9, 10 and 11 all stop the whole run in the simulator (`cfgErr`). The clause reads as if 9 and 10 fail one instance.
6. **Rule 11 list drifts from the predicate (`04-grammar.md:309`) (*unspecified*).**
   - It says "a control character" and omits the BOM and invalid UTF-8, both in `02-semantics.md:175-179` and `main.go:637-649`.
   - It is one predicate in the model and in `02`, but three phrasings in the text. State it once and refer to it.
7. **`02-semantics.md:158-160` says `C` and `X` are errors "because Kyn never asks git for copies or type changes".** `T` is the type-change status, and it is accepted. Reword, for example "copies, or any other status".
8. **Fact 3 also suppresses rules 9 to 11 (`02-semantics.md:210-215`, row 4) (*unspecified*).**
   - A dropped plain source never reaches resolve-time checks. A git-mode `M` source absent from disk does.
   - Config: `match: src/**/{n}.go`, `related: src/{n}.go`, `expect: exists`. With `src/a.go` absent from disk, plain gives PASS and `M\tsrc/a.go` gives exit 2 (rule 10). I ran both (cases A and B).
   - Row 4 should add "and no resolve-time check runs for it".
9. **Whole-run abort for real tracked file names is not disclosed in `06` (*unspecified*).**
   - `02-semantics.md:174-179` makes every changed path, matched or not, fatal for the entire run if it has an edge blank, a control, a backslash, or a first segment that is a letter plus a colon.
   - Confirmed:
     - A root file `Q: and A.md` aborts with "looks like a Windows drive" (case E).
     - `old /x.txt` aborts (case 24).
     - `a\b.txt` aborts (case G).
     - `x:y` aborts (case 25).
   - These names are legal on Linux and macOS.
   - `06-cutover.md:18` frames the break as Kyn no longer cleaning paths. It does not say a repository with such a tracked name now fails every PR touching it.
   - Either disclose this, or apply the drive and edge-blank checks only to resolved related paths.
10. **Mutation survivor (`main.go:773`).** Changing `R[0-9]{1,3}` to `R[0-9]+` leaves all 134 scenarios green, so nothing covers `R1000`. `R0`, `R000` and `R999` are accepted, and git emits at most `R100`. Add `R1000` and bare `R` scenarios, and decide whether to cap at 100.
11. **A mutation that did not compile does not count as guarded.** Three of my mutations (rule 11's `badPath` in resolve, the conflict check, the drive check in `related`) failed to compile. I could not test them this way, so I take the author's "thirteen mutations" as unverified for those three. Another mutation (whitespace-only lines) panicked instead of turning scenarios red. All the others turned scenarios red.

## What holds up

- **D66, binding totality: closed.** Details are in the differential results below.
- **D67, the three facts: closed.** I derived every row from the facts, and the model agrees:
  - Row 1: related deleted and not listed. `skipIf: missing` skips, and `untouched` passes (case L).
  - Row 2: related deleted and listed. `touched` passes, `untouched` fails "was touched", and `skipIf: missing` skips (cases 33, K).
  - Row 3: `changeType: added` never matches in plain mode.
  - Row 4: a plain source absent from disk is dropped, and a git `A` or `M` absent from disk is kept (cases 01, 02, 35, 36).
  - "Per line" holds: a mixed file keeps removal and status behaviour for status lines and applies the drop only to plain lines (case 03).
  - Mixed files with conflicting or duplicate lines behave as `02-semantics.md:182-186` says. Identical duplicate plain lines pass (case 05). A plain line plus `M` for the same path errors (case 04). A rename destination listed plain errors (case 06). `D x` plus a plain `x` errors (case J).
  - The phrase "derivation wins" no longer appears in any `docs/design/v3/*.md` file.
- **D68, can-never-fail: closed.**
  - 84 cells is correct. The 24 rejected cells are the four `expect` lists (`touched`, `exists`, `[exists, touched]`, `missing`) under `skipIf: exists`, times the six related-path rows.
  - `02` no longer recommends `skipIf: exists` with `missing`.
  - It does not reject anything useful: the only accepted-before, rejected-now combination is a rule that can never fail, and `expect: none` with `flags` covers the flag use.
- **D69 and D70: closed apart from P2-1 and P2-2.**
  - `{ts, html}`, `{ts,html }`, `related` edge blanks and BOM, and C0/C1 controls (including U+0085) are all rejected.
  - Interior NBSP, interior spaces, interior ZWJ and a trailing U+FE0F are accepted, as intended.
  - `{**,a}` has its own message ("alternatives must be literal text: * is not allowed inside {}").
  - `C:{n}` is rejected at load.
  - CRLF is tolerated, and a lone CR mid-line is a control and an error.
  - `U` has its own message.
  - Whitespace-only and tab-only lines are ignored.
  - A BOM is an error.
  - Extra fields, a bare `R` and `R1000` are errors.
  - `T` and `R0` are accepted.
- **"No configuration in this repository uses a rejected form" holds.** I extracted every glob string from the four `e2e/projects/*/kyn.config.yaml`, `testdata/angular/kyn.config.yaml`, `internal/cli/init.go`, `docs/site/*.md` and recipes, and `README.md`. All 38 distinct patterns are plain `*`, `**` or literal.
- **`--files` comma claim** (`06-cutover.md:18`): `internal/cli/flags.go:42` says "Comma-separated changed files". It holds.

## Closure check: D66 to D70

| ID | Verdict |
| --- | --- |
| D66, binding has no tie-break and ties are rejected | **Delivered.** Binding is total on every accepted glob I generated (0 ties in about 27M pairs across the tests below). The mutation turns 3 scenarios red. The rule is wider than needed (P3-1). The "two million accepted" count is wrong (P3-2). |
| D67, three facts | **Delivered.** Every row derives, and the model is per line. Gap: row 4 omits that rules 9 to 11 never run for a dropped source (P3-8). |
| D68, can-never-fail and 84 cells | **Delivered.** The plain-mode vacuity sentence is incomplete (P3-3), and the "never pass" label for `touched` is wrong (P3-4). |
| D69, one blank predicate | **Delivered as specified, but the specification omits Cf (P2-2).** It is one predicate in the model. The text phrases it three ways (P3-6). |
| D70, exact line grammar | **Delivered except the silent space-separated status line (P2-1).** The `R`-score bound has no scenario (P3-10). The `T`/`C`/`X` wording is off (P3-7). |

## Closure check: review 7 disposition table

| Finding | Verdict |
| --- | --- |
| P1 row 4 contradicts the derivation | **Genuinely closed.** The rows are no longer separate from the facts. The text states the drop as a decision. |
| P2 whitespace misses `{ts, html}`, `related`, BOM and controls | **Closed for the stated predicate. Reopened for Cf (P2-2).** |
| P2 `skipIf: exists` with `expect: missing` can never fail | **Closed.** |
| P2 `--files-from` grammar loose | **Closed, except space-separated status lines read as plain paths (P2-1).** |
| P2 `src/**/{feature}/**` binds on name length | **Closed**, by rejection. See P3-1 for over-width. |
| P3 stale row ref, `expect: missing` message, dead `\` check, `[untouched, missing]` survivor, `C:{n}`, `{**,a}` message, `--files` comma, plain-mode rule 7, empty `{**p}` start | **Closed**, with these notes:<br>• The `[untouched, missing]` survivor now turns 1 scenario red.<br>• The plain-mode rule 7 sentence is only half complete (P3-3).<br>• The `\` check is now part of the shared predicate.<br>• The `expect: missing` message "exists but must not" is in the `02` §5 table. |

## Overclaim check

- **"With that rule the comparison is total":** true.
- **"two million random accepted globs":** false (about 1.0M accepted). See P3-2.
- **"the same predicate is used for globs, `related` and input paths":** true in the model. In the text, rule 11 (`04-grammar.md:309`) lists a different set (P3-6).
- **"Every path, in every input mode including `--files`, must be clean":** true in the sense of fatal. This is the blast-radius point in P3-9.
- **"controls and the BOM are rejected anywhere":** true. The model's own message, "control or invisible character", claims more than the predicate does (P2-2).
- **"The statuses are exactly the five above":** true.
- **"The simulator checks … with the rule removed, it finds them":** true.
- **`02-semantics.md:223` "the recipe relies on" fact 3:** true.
- **"(rule 7) … five reachable pairs":** true.
- **"Git never emits such a diff" for conflicting lines:** not tested against real git, but plausible.

## Cases tried

63 scenarios in `sc/a.json`, `sc/b.json` and `sc/c.json`, run in verbose mode. Counts below are the ones that behaved correctly unless marked.

**Plain and mixed input**
- Plain absent source dropped; git `M` absent source evaluated; `A` absent source evaluated.
- Mixed `D` related plus plain source (`skipIf: missing` fails "removed by this change").
- Plain plus `M` conflict; identical plain duplicate; rename destination also plain (error); `D x` plus plain `x` (error).
- `R0` accepted; `R1000` and bare `R` errors; `R999` accepted; `T` accepted; `U` error.
- CRLF; lone CR mid-line; BOM first line; whitespace-only and tab-only lines ignored; interior TAB in a plain path (error).
- Space-separated status lines (**P2-1**); `git status` porcelain lines (**P2-1**).
- Related deleted and listed or unlisted, under `touched` and under `untouched`.
- Two sources of one instance, one absent from disk.
- Dropped-source versus rule 10 (case A versus case B).

**Globs and templates**
- `{ts, html}`, `*{n}`, `?{n}`, `{.,a}`, `{**,a}`, `C:{n}`, `{n}{n}`.
- `src/{**pkg}/{name}/**`, `src/**/{name}/**`, `src/**/{name}/{**rest}`.
- `{n}/**`, `**/**`, bare `**`, `{**p}/{n}.go`.
- Mirrored py (`src/{**pkg}/{name}.py`).
- Two overlapping globs (rule 9), alternation `{ts,tsx}` with related matching own match (rule 10), constant related inside own match.
- Trailing-space capture in related (rule 11, run-level).
- NBSP interior; ZWSP at glob edge (**P2-2**); ZWSP at path edge; ZWJ at a segment start; emoji directory ending in U+FE0F; NFD name against an NFC glob (no match); `e` plus combining mark.
- Root file `Q: and A.md`, `x:y`, `old /x.txt`, `a\b.txt` (abort the run).

**Rule 7**
- `skipIf: missing` with `expect: exists`, in plain mode (accepted, can never fail in plain).

## Differential tests

1. **Independent matcher and acceptance check (`s2/mine.go`).**
   - I wrote it from the Binding and Globs paragraphs alone: a segment-wise rune matcher that enumerates all parses, selects longest-first, and reports ties. It has its own glob validator from the dialect table.
   - Diffed against `compileGlob` and `bind` on random globs (1 to 4 segments, with `*`, `?`, three capture names, alternation, bare `**`, `{**p}`, blanks, NBSP, ideographic space, ZWSP, ZWJ, variation selector, BOM, TAB, `{dir}`, bad names, `[`, `{**p,a}`), and on random and glob-derived paths. About half of the paths were instantiated from the glob, with non-ASCII, combining marks and a 4-byte emoji.

   | Run | Globs | Accepted by both | Accept/reject diffs | Pairs | Binding diffs | Ties found by mine |
   | --- | --- | --- | --- | --- | --- | --- |
   | 1 | 100k | 19,967 | 0 | 599k | 0 | 0 |
   | 2 | 600k | 119,024 | 0 | 3.36M | 0 | 0 |
   | 3 (`HEAVY`: `**`, `{n}`, `*`, `?`, `-`, `a`) | 1.5M | 540,430 | 0 | 15.68M | 0 | 0 |

   - The first run showed 34,831 diffs, all from my harness handing an empty path to my matcher. I fixed the harness; the table is after the fix.
2. **Attack on totality.**
   - **By construction.** For a capture with equal-length parses at different starts, the longer extension always exists unless the capture cannot cross a segment boundary. That needs a flexible token (`**` or `{**x}`) on both sides of the capture. With `{**x}` as the first capture, it wins by length instead. I tried to build a counterexample with `*`/`?`/alternation separators and several `{name}` captures in one segment (for example `*-{n}-*`, `{a}-*-{b}`). None ties.
   - **Generator.**
     - My `HEAVY` run above.
     - A richer segment set in `s1/` (`-{x}-`, `-{x}.`, `{x}-`, `{a,b}`; 5 to 6 tokens per segment): `-ties 300000` gives 98,230 accepted, 1.54M pairs, 0 ties.
     - The author's `-ties 2000000` also gives 0 ties (1,007,980 accepted).
   - **Rejection removed.** `NOREJECT=1 -ties 300000` gives 1,991 ties, for example `**/{y}/**` on `bb.ab/é./é--/a.aé./b.b`.
   - **Exhaustive** (6 words, depth ≤ 6): `src/**/{name}/**` ties 4,914 times (`src/src/src/src`); `src/{**pkg}/{name}/**`, `src/**/{name}/{**rest}` and `**/{**p}/**` tie 0 times.
   - **Rejected but tie-free:** among 18,344 randomly generated globs the rule rejects, 1,148 had a tie in 300 random paths each, and the rest did not. The "no tie found" bucket is not a proof, because the paths are random. Case-by-case reasoning plus the exhaustive runs support P3-1.
3. **Rule 7.**
   - My checker from `02` §4 and rule 7, over 3 `skipIf` values × the 8 non-contradictory `expect` subsets (24 combos): 0 disagreements with `passability`.
   - Every scenario rule in `scenarios.yaml` was also compared in review 7. This time I re-derived the 84-cell count by hand: 24 rejected cells, 108 − 24 = 84.
   - Plain-mode variant with four pairs: exactly two accepted combinations are vacuous (P3-3).
4. **Clean-path predicate (`myCleanPath` against `badPath` and `badInputPath`).**
   - 43 hand-picked strings × 2 modes. They cover: NBSP, ideographic space, ZWSP, ZWJ, variation selector, BOM, combining mark, NEL, U+2028, U+180E, DEL, TAB, LF, CR, NUL, invalid UTF-8, a quote prefix, a drive prefix, a backslash, and edge-blank segments.
   - Plus 2,000,000 random strings × 2 modes over `a`, `b`, `/`, `.`, space, NBSP, U+3000, BOM, TAB, ZWSP, ZWJ, FE0F, `é`, `e`+combining, `\`, `:`, `"`, `C`, DEL, NEL, U+2028, U+180E and C0.
   - **0 diffs.** My predicate used `unicode.White_Space`, which is what Go's `unicode.IsSpace` uses for non-Latin1.
   - The same predicate is used for globs, `related` and input paths: 0 accept/reject diffs in test 1 includes the edge-blank and in-brace cases.
5. **Mutation tests on the author's model.** Nineteen attempted; sixteen ran:
   - Turned scenarios red (1 to 3 each): both-sides rejection, never-fail, disagree-only, plain drop, `U` message, CR strip, `badPath` edge blank, `hasHidden`, blank-in-braces, `related` edge blank, `untouched`/`removed`, input quote prefix, rule 9.
   - **Survived:** the `R[0-9]{1,3}` bound (P3-10).
   - Did not compile or panicked: whitespace-only lines, rule 11 in resolve, conflict check, drive in `related` (P3-11). One more mutation I wrote did not apply.

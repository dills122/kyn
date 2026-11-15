# Independent review 9 of the v3 matching section

| | |
| --- | --- |
| Source | a fresh-context subagent given `docs/reviews/v3-config-review-9-bootstrap.md` and no prior context, 2026-11-15 |
| Reviewed | `f7c43d9` (scope frozen in `8b882e8`), branch `design/v3-config` |
| Dispositions | `docs/design/v3/00-workplan.md`, section Reviews |

Verbatim as delivered. Scope was the matching section only. Line citations refer to that commit. The reviewer's scratch files were under its own `$TMPDIR` and are not preserved; the cases it describes are reproducible from the text.

---

# Independent review 9 of the v3 matching section

State reviewed: design `f7c43d9` plus the bootstrap commit, branch `design/v3-config`. Read-only: nothing in the repository was modified or committed. Scratch is `/var/folders/lg/j56llg991d1gpjvtjh8l9gbh0000gn/T/tmp.lqB940TNZE`. It holds `sim/` (a modified copy of the simulator plus my `diff.go`), `mut/`, `seed11/` and `seed12/`, and the scenario files `mine.yaml`, `m2.yaml` and `m3.yaml`. I did not read any `v3-config-independent-review-9*.md`.

## Verdict

**Freeze decision: not freezable exactly as it stands. Freeze after one small P2 edit and the listed P3 edits.** There is no P0 or P1.

- **The P2** (below) is an incomplete closure of D71, and I reproduced it. The status-like rejection catches single-letter `git status --short` codes but lets the two-letter codes through (`AM`, `MM`, `UU`, `RM`, and others). Those are silently read as a file name that evaluates nothing, which is the failure D71 set out to close.
- **The smallest change** that makes it ready is to widen the token set in `02-semantics.md:164-168` and `main.go:780`. The wider set must cover all of git's two-letter XY codes and `C<digits>`. Add one scenario each for `AM`, `MM`, `UU` and `C100`, and add a sentence saying what `--files` does with such a line. After that, only P3s remain.
- **The rest holds.** My own predicate, rule 7 enumeration and tie searches all agreed with the model. The 150, "two million tried, about one million accepted" and "no file in this repository is affected" claims all reproduce.

## Findings

### P2-1: The status-like rejection misses two-letter `git status` codes, so the D71 hole stays open for them (*unspecified*)

- **Evidence.** `02-semantics.md:164-169` lists the tokens as `A M D T U C X`, `R` with digits, `??` and `!!`. `main.go:780` implements `^([AMDTUCX]|R[0-9]+|\?\?|!!) `. The text justifies the rule as covering "`git status` output".
- **Inputs, plain mode and mixed mode** (all accepted, config `match: src/**/{n}.go`, `related: {dir}/{n}_test.go`):
  - `AM src/a.go`: a new file that was edited after `git add`. This is very common for a one-file change.
  - `MM src/a.go`, `UU src/a.go`, `AA`, `DD`, `RM src/a.go -> src/b.go`.
  - `C100 a b`, `C075 a b`: git's copy form. `R` is covered with digits and `C` is not.
  - Each gives `=> PASS map[]` with zero instances. The line is read as a file called `AM src/a.go`, which is not on disk, so fact 3 drops it silently.
- **Why it matters.** It is the same silent-never-match class as review 8 P2-1. Typical multi-line `git status --short` output contains some `M  ` or `?? ` line and so aborts. Output where every line is a two-letter code, such as a single `AM` file, still passes with nothing evaluated.
- **Fix.**
  - Reject a TAB-less line that starts with an exact git XY pair followed by a space. The pairs are `??`, `!!`, and any two letters from `{M,T,A,D,R,C,U}`. Also reject `C<digits>` followed by a space.
  - Give the unmerged pairs `UU`, `AA`, `DD`, `AU`, `UA`, `DU` and `UD` the same "resolve the merge first" message as `U`.
  - Cost: names such as `AD Notes.md` or `DM Plan.md` at the root now need the status form. That is the same trade-off D71 already accepted.
  - Add scenarios `AM`, `MM`, `UU`, `C100`.

### P3 findings, ranked

1. **The status-like rejection is a new false alarm that aborts the whole run, and part of its workaround does not exist (*unspecified*).**
   - Evidence: `02-semantics.md:168-169` says to use the status form. Simulator output:
     - `A Guide.md`, `M Notes.md`, `D Day.md`, `T Shirt/a.go`, `C Sharp.md`, `X Marks.md` all abort the run with exit 2. These are plausible root-level names in docs repositories and note vaults.
     - `D ` (a trailing space) also aborts.
   - Gap 1: `--files` is comma-separated and has no status form. The text only describes the rule for `--files-from` lines, but the model applies it to every plain input. State what `--files` does.
   - Gap 2: the `git ls-files | kyn check --files-from -` recipe (`04-grammar.md:~335`) cannot succeed in a repository that tracks such a name, and the recipe does not say so.
   - Gap 3: `06-cutover.md:18` lists the names that now abort a run. It omits status-like names and a leading `"`.
   - Fix: add one sentence each. The model's error message does not mention the status form either. Adding "use `M<TAB>path`" to the message is cheap.
2. **Cf is blank at a segment edge, which rejects real names, and the text only says emoji joiners are legal (*matter of taste*, but the whole-run abort amplifies it).**
   - Evidence: `04-grammar.md:53-63`.
   - Rejected with a whole-run abort:
     - The emoji tag sequence for the England flag, which ends in U+E007F (Cf). A directory `flags/🏴󠁧󠁢󠁥󠁮󠁧󠁿` is rejected.
     - A trailing or leading RLM U+200F, as in `src/ا<RLM>/x.go`.
     - Malayalam and Devanagari segments ending in ZWJ or ZWNJ, and U+0600 or U+180E at the start of a segment.
     - Emoji or Arabic text with the joiner inside `{}` in a glob (`{png,<ZWJ>x}`), which is rejected as "whitespace inside {}" although it is legal in a name.
   - Fix: either disclose these in `06`, or exempt tag sequences and the ZWJ and ZWNJ edges. Disclosure is enough to freeze.
3. **Invisible characters outside White_Space, Cc and Cf are still accepted anywhere (*matter of taste*, but undisclosed).**
   - Accepted, including as the whole segment: U+2800 braille blank, U+115F, U+3164, U+034F, a lone U+FE0F.
   - These accepted names can never match what was meant, which is the rationale used at `04-grammar.md:92`.
   - Fix: add one sentence saying only the three categories are detected.
4. **The legal-interior, illegal-edge rule can abort on a legal source through a capture (*unspecified*).**
   - Config: `match: src/x {n}.go` with `related: t/{n}.go`.
   - Input: the tracked file `src/x  a.go`, with two spaces, binds `n=" a"`. It resolves to `t/ a.go`, which is an edge blank, and rule 11 stops the whole run (`resolves to an invalid path`).
   - The same happens with an interior ZWJ at a capture boundary.
   - `04-grammar.md:219-225` and the limits list (`:405`) disclose only the `..` case. Fix: name edge blanks there.
5. **Globs that can never match a clean path are accepted (*unspecified*).**
   - `match: "c:{n}.go"`, `"c:/{n}"` and a glob starting with `"` all load and give PASS with zero instances.
   - Evidence: `04-grammar.md:53-56` says one predicate is used for globs, but the glob rejection table (`:92`) has no drive or quote row. Templates do have the drive check.
   - This is the same class as the ZWSP glob closed in D72. Fix: either reject a first glob segment of letter plus colon, or say the predicate does not apply to globs.
6. **The clean-path check also covers deleted paths and rename sources, so the PR that fixes a bad name cannot pass (*unspecified*).**
   - `D<TAB>old /x.txt` and `R100<TAB>old /x.txt<TAB>old/x.txt` both abort with "invalid input path", and so does the plain form.
   - A removed path only ever serves as a state key. A resolved related path is always clean, so it can never equal an unclean one. Checking removal-only paths therefore adds no protection and blocks the remedy.
   - D74 discloses "every changed path", but not this consequence.
   - Fix: exempt `D` and rename-source paths from the check, or add a sentence saying the fix PR is also blocked.
7. **A leading `"` aborts a git-mode run with a misleading message (*unspecified*).**
   - `M<TAB>"x".md` gives "in git's quoted form; feed git -c core.quotePath=false".
   - Git mode reads with quoting disabled, so a tracked file named `"x".md` is legal and not quoted.
   - Fix: apply the quote rule to `--files-from` and `--files` only.
8. **The mutation claim covers fewer decisions than it implies (*matter of taste*).** Seven decisions are guarded, but these mutations leave all 150 scenarios green:
   - Dropping any of `A`, `T`, `U`, `C` from the token class, or `!!`. Only `D`, `M`, `??` and `R` are guarded.
   - Changing `R[0-9]{1,3}` to `R[0-9]+`. This one is equivalent, because the `n > 100` cap catches it.
   - Suggested scenarios: one per token. `X` and `C` also survive.
9. **Stale text (*wrong*, trivial).**
   - `04-grammar.md:433` says scenarios cover the "fifth to seventh reviews". The 116 to 119 scenarios are review 8.
   - `00-workplan.md:108` (D66) still says "two million accepted globs". D75 says this was fixed. It is about 1.0M accepted.
   - `04-grammar.md:206` says "begins with a letter and a colon". `:218`, `:59` and the model say ASCII letter.
   - Related restatement: `04-grammar.md:204-220` restates the predicate (the `related` bullets and the resolution paragraph). That sits awkwardly with "defined once" (D72).
   - `scenarios.yaml:1864`: `117c` sits at the end of the file, after `119c`.
10. **The Unicode version is not pinned (*unspecified*).**
    - White_Space and Cf membership shift between Unicode versions (U+180E moved from Zs to Cf in 6.3).
    - Name a version in `04` or `07`, since the conformance suite will compare implementations. Go 1.22 is Unicode 15.0.
11. **`git log --name-status` header lines (*matter of taste*).** `commit abc123`, `Author: x <y>` and `Date:   Mon` are accepted as plain paths, which are dropped. Indented message lines abort the run. A lower-case `m src/a.go` and an NBSP after the token (`A<NBSP>src/a.go`) also pass, and a leading-space `" M x"` is rejected. All of these are acceptable.

## What holds up

- **Binding totality** is unchanged.
  - `-ties 2000000` gave 1,007,980 accepted, 11,014,745 pairs, 0 ties.
  - Two extra seeds, with extra tokens such as `{x}-`, `-{y}-` and `{z}.`, gave 600k tried and about 287k accepted each, about 6.5M pairs, 0 ties.
- **Rule 7 and the 84 cells** are unchanged.
  - The 24 rejected cells are the four `expect` lists (`touched`, `exists`, `[exists, touched]`, `missing`) under `skipIf: exists`, times the six related-path rows.
  - 108 − 24 = 84.
- **"No file in this repository is affected" holds.** I checked all 208 paths from `git ls-files`. None has a status-like start, a backslash, a quote, a drive prefix, an edge blank or any Cc or Cf character.
- **The numbers hold.**
  - 150 scenarios, and the run is green (0 mismatched).
  - "Two million tried, about one million accepted" is exact.
  - The scenario numbers cited in the workplan exist: 39, 113f, 114c and 117 to 117c.
- **Scepticism 3 (contradictions across `02`, `04`, `06`, `07` and the workplan).** After eight rounds there are only the stale items in P3-9.
- **Scepticism 4 (the whole-run abort).** It is acceptable as a stated limit, but it hides the worse cases in P3-1, P3-4 and P3-6.

## Closure check: D71 to D75

| ID | Verdict |
| --- | --- |
| D71, status-like rejection | **Partly closed.** Single-letter and `??` forms are rejected. Two-letter XY codes and `C<digits>` still pass silently (P2-1). The workaround does not exist for `--files` or `ls-files` (P3-1). |
| D72, Cf is blank, predicate defined once | **Closed as specified, and the sim matches the text** (differential below). Edge-Cf false positives (P3-2), undetected non-Cf invisibles (P3-3). The predicate is restated twice (P3-9). |
| D73, rename score 0 to 100 | **Closed.** `R101` and `R1000` are errors, and `R100` and `R095` are accepted. Scenario 117c guards the cap. |
| D74, disclosed whole-run abort | **Disclosed, with gaps.** The fix-PR consequence is undisclosed (P3-6), as are the status-like and leading-quote names (P3-1, P3-7). |
| D75, capture rule kept, wording, counts | **Closed.** The count and the rule 7 sentences are right. The one stale count is D66's (P3-9). |

## Closure check: review 8 table

| Review 8 finding | Verdict |
| --- | --- |
| P2 status line with spaces read as a path | **Reopened for the two-letter codes (P2-1).** |
| P2 Cf not blank | **Closed.** |
| P3 `{**x}` rejection wider than needed | Kept and reworded; the wording is accurate. |
| P3 "two million accepted" | **Fixed in `04`.** Still stale in D66 (P3-9). |
| P3 rule 7 sentences and example labels | **Fixed.** |
| P3 "as rules 9 and 10 do" | **Fixed.** `04:224` now says "Rules 9 and 10 also stop the whole run". |
| P3 rule 11 list drifts from the predicate | **Fixed.** Rule 11 points at the definition. |
| P3 `C`/`X` wording; row 4 resolve-time | **Fixed.** `02` fact 3 and row 4 say so. |
| P3 whole-run abort undisclosed | **Disclosed**, with gaps (P3-1, P3-4, P3-6, P3-7). |
| P3 `R1000` survivor and uncompiled mutations | **Fixed** by scenario 117c. Token-level survivors remain (P3-8). |

## Cases tried

About 90 inputs, run through the simulator in verbose mode.

**Rejected or accepted correctly:**
- Status-like: `D x`, `M  x`, `?? x`, `!! x`, `R100 a b`, `U x`, `C x`, `X x`, and `D<NBSP>x` (accepted, a path). Lower case, a leading space (rejected by the edge check), a quoted path, and `git ls-files -s` (rejected as an unknown status).
- `R100 a` and `R1000 a b` are rejected. `Rx a`, `R-1 a` and bare `R x` are accepted as paths.
- Formats: `M<TAB>M foo.go` (accepted), `M<TAB> src/a.go` (edge blank) and `M src/a.go<TAB>` (unknown status). BOM first line, whitespace-only lines.
- Cf and invisibles at the edge: ZWSP, ZWJ, ZWNJ, LRM and RLM, bidi isolates U+2066 and U+2069, U+2060, U+202E, U+061C, U+00AD, U+0600, U+180E, tag characters. All are rejected at the edge and accepted in the interior.
- Invisibles accepted: U+2800, U+115F, U+3164, U+034F and a lone U+FE0F (P3-3).
- A trailing U+FE0F on an emoji (a real name) is accepted, as intended.
- Resolve-time and glob cases: related resolving to an edge blank (P3-4), `D` of a bad name (P3-6), globs `c:{n}.go`, `c:/{n}` and a quote-leading glob (P3-5), a ZWJ interior to a glob (accepted), a ZWJ in `{}` (rejected, P3-2), a leading `"` in git mode (P3-7), and mixed-file lines.
- `git log` header lines (P3-11).
- A real-repo scan for names the new rejections would hit: none found.

## Differential tests

All of these ran from the repository root against the simulator copy, using my own predicate built from `unicode.White_Space`, `Cc` and `Cf` tables rather than `IsSpace` or `IsControl`.

1. **Clean-path predicate** (my `myClean` against `badPath` and `badInputPath`):
   - 3,000,000 random strings, drawn from 40 hostile characters including all the Cf and invisible cases above.
   - Every rune from U+0000 to U+10FFFF (excluding surrogates) in six contexts: interior, start, end, and inside a directory. 6,672,384 checks.
   - **0 diffs in either run, in both modes.**
2. **Glob blank rule:** 2,000,000 generated globs of benign pieces mixed with blank characters, compared with my edge-blank, in-`{}` and control-anywhere check.
   - 0 globs accepted that mine rejects.
   - 0 rejected for a blank reason that mine accepts.
3. **Template blank, control and drive checks:** 1,000,000 generated templates, **0 diffs**.
4. **Rule 7:** all 3 `skipIf` values by the non-contradictory `expect` subsets, derived from the `02` §4 truth table. **0 diffs against `passability`.**
5. **Binding totality:** 0 ties in about 17M pairs across three runs (seed 1 at 2M globs, seeds 11 and 12 at 600k each).
6. **Mutations** (each run with a compile and panic check; the first attempts failed to compile and were redone):
   - Turn scenarios red: status-like off, dropping `??`, dropping `R`, Cf not blank, blank in braces off, glob edge-blank off, input quote check off, BOM off, the `n > 100` cap, drive check off in `badPath` (scenario 107), drive check off in templates (scenario 113f), `mixed` handled as plain only.
   - Survive: dropping `A`, `T`, `U`, `C`, `X` or `!!` from the token class (P3-8), and `R[0-9]+` for `R[0-9]{1,3}` (equivalent).

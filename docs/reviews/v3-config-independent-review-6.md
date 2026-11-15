# Independent review 6 of the v3 matching section

| | |
| --- | --- |
| Source | a fresh-context subagent given `docs/reviews/v3-config-review-6-bootstrap.md` and no prior context, 2026-11-15 |
| Reviewed | `6737d73`, branch `design/v3-config` |
| Dispositions | `docs/design/v3/00-workplan.md`, section Reviews |

Verbatim as delivered. Scope was the matching section only. Line citations refer to that commit. The reviewer's scratch files were under its own `$TMPDIR` and are not preserved; the cases it describes are reproducible from the text.

---

# Independent review 6 of the v3 matching section (state 6737d73, bootstrap commit 306373e)

Read-only. Scratch work is in `$TMPDIR/rev6.eV4A` (my matcher, the differential harness, 4 scenario files with 70 adversarial cases, and a doublestar probe). Nothing in the repository was touched. The author's own `go run ... scenarios.yaml` is green: 87 scenarios, 0 mismatched.

## Verdict

**Not ready to freeze the matching section. It is much closer than review 5 found it.** No P0. There are 2 P1s, both about the model or the "this is the whole list" claims, not about the dialect.

- **Holds up.** The dialect table and its rejections are implementable. One-pass resolution with rule 11 is sound. The rule 9 to 11 ordering claim is true. The claim that no repository configuration uses a rejected form is true. My independent matcher agreed with the simulator on 100% of cases without `**`, and on every case that has a unique answer.
- **Fails.**
  1. "Longest, first capture first" is **not total** once a bare `**` is on both sides of a capture. The simulator resolves ties silently by enumeration order, and the text says `**` carries "no preference".
  2. The claim "there are three plain-path differences, and this is the whole list" is wrong again. A related file that the diff deleted and the user also listed behaves differently in plain mode, in a way row 1 contradicts.
- **Model faults.** The simulator compares byte lengths, not characters. It cannot match non-ASCII literals or `?` against non-ASCII text. The text is silent on all three.

## Findings, ranked

### P1-1: binding is not total when `**` brackets a capture (*unspecified*, with the model silently choosing)

- **Evidence.** `04-grammar.md:103-109` defines binding as longest captures, first capture first, and says "Nothing else has a preference: `*`, `**` ... do not influence". No scenario covers a tie. Scenario 63 anchors the capture on a literal, so it does not.
- **Input.**
  - Config: `match: "src/**/{n}/**"`, `related: "docs/{n}.md"`.
  - Path `src/auth/user/a.ts`: n=`auth` or `user`, both length 4. The simulator binds `auth`.
  - Path `src/user/auth/a.ts`: the simulator binds `user`.
  - So the answer is "outermost first", and nothing in the text says so. A longer name still wins, so `src/auth/login/a.ts` binds `login`.
- **Realistic form.** `src/**/{feature}/**/*.test.ts`, or any `**/{x}/**`. Two different directories with equal-length names give two valid parses with different values.
- **Differential result.** All 69 ties in 12M cases involved a bare `**`. Zero ties turned up in 12M cases without `**`, and in 12M cases with `{**p}` but no bare `**`. So `**` is the only source. `{n}` bordered by literals or `/` on one side is safe.
- **Impact.** An instance identity and the related path depend on directory depth. In multi-glob rules, rule 9 compares bindings that are themselves tie-broken arbitrarily.
- **Fix.** Either reject a `{name}` capture that has an unanchored `**` on both sides, or add a final tie-break such as "leftmost start wins" and add conformance rows for it. The second is cheaper. State it in the Binding paragraph and test it at layer 0.

### P1-2: the plain-path difference list is incomplete, and row 1 is wrong for a listed deleted file (*wrong*)

- **Evidence.** `02-semantics.md:175-189` says "There are three differences, and this is the whole list". Row 1 (`:181`) says a deleted related file is `missing` and "`expect: untouched` **passes**". The simulator treats every listed path as touched (`main.go:569-575`, with `isTouched` used at `main.go:~855`).
- **Input.** Plain mode, `match: "src/{n}.go"`, `exclude: "src/*_test.go"`, `related: "src/{n}_test.go"`. Disk is `src/a.go`. The list is `src/a.go` and `src/a_test.go`, where the test was deleted. This is what `git diff --name-only | kyn check --files-from -` produces.
  - `expect: touched` gives **PASS [missing]**. In git mode it is FAIL "removed by this change".
  - `expect: untouched` gives **FAIL "was touched"**. Row 1 says it passes.
- **Impact.** This is the silencing hole that the `removed` state exists to close, and plain mode reopens it on the most common feed. It is the third time "exactly N differences" has been wrong.
- **Fix.** Add a row: in plain modes a listed path counts as touched even if it is not on disk, so a deleted related file satisfies `touched`. Correct row 1's `untouched` claim to "passes only if the deleted file is not listed". Replace "this is the whole list" with a derivation: plain mode loses removal and status, and treats the list as the whole change set.

### P2-1: rows 3 and 90/90b hide a real git-mode versus plain-mode difference (*unspecified*)

- **Input.** Git mode, `A src/a.go` with the file absent from the working tree. This is the OS1 case, `--base main --head origin/feature` from another checkout. The simulator creates an instance (result "FAIL a does not exist"). Plain mode drops the source (`main.go:767`) and reports nothing.
- **Evidence.** The text conflates "deleted" with "not on disk": `02-semantics.md:183` and `:203-205`, and scenario 90's name "as in git mode". Only `D` is covered, by scenario 90b.
- **Fix.** Decide whether git mode drops sources that are not on disk. The cleaner choice is to keep sources present in the diff and add a fourth difference row.

### P2-2: the `git ls-files` recipe can pass vacuously (*unspecified*)

- **Evidence.** `04-grammar.md:286-292`, combined with D56.
- **Mechanism.** Plain mode drops every listed source that is not on disk. If the git command runs in a different directory from `--cwd`, or in a sparse checkout, a submodule, or with quoted names, everything is dropped. Exit is 0 with zero instances. Repro: plain mode with `disk: []` gives "PASS" with no results.
- **Quoting.** `core.quotePath=false` only affects non-ASCII. Git still quotes a tab, `"`, `\` or newline. The line is then read as a literal path with quotes and silently dropped. A name with a tab is read as a name-status line. A name with a space is split by the simulator's `strings.Fields` at `main.go:569`, a model-only defect.
- **Fix.** Add `--fail-on-empty` to the recipe. Prefer `-z` and say whether `--files-from` accepts NUL. Say what the recipe exercises: tracked, on-disk sources only.

### P2-3: the model is wrong for non-ASCII input (*wrong in the model*, silent in the text)

- **Literals.** `main.go:264` does `string(c)` on a byte. Any non-ASCII byte becomes a two-byte Latin-1 rune, so the literal never matches. Repro: `match: "src/é.go"` with path `src/é.go` yields no instance. doublestar matches it (probed).
- **`?`.** `?` consumes one byte, so `src/?.go` does not match `src/é.go`. doublestar does.
- **Length unit.** "Longest" compares bytes (`main.go:302`). With `src/**/{n}/**` and `src/é/ab/a.ts` the simulator binds `é` (2 bytes, a tie). By characters `ab` wins. The text never says whether length is bytes, runes or segments.
- **Not specified.** `?` and `*` units, Unicode normalization (macOS NFD versus NFC), and whether `*` matches a leading dot. The simulator and doublestar both let it match.
- **Fix.** Specify runes, and fix `compileGlob` to scan runes. Say `*` and `?` match dotfiles.

### P2-4: validation gaps that rule 7's own reasoning implies (*unspecified*)

- **Never passes, accepted.** `skipIf: missing` with `expect: [touched, missing]`, or with `[untouched, missing]`. The rule applies only in states `exists` or `removed`. `missing` needs state `removed` or `missing`. `touched` or `untouched` then excludes `removed`. Result: it can never pass. Simulator output: "FAIL a exists" and "FAIL b removed by this change".
- **Passes only on disagreement, accepted.** Unconditional `[touched, missing]` is the same defect rule 7 rejects for `skipIf: exists` with `touched`.
- **Fix.** Reject any state-set that is empty or reachable only on diff/tree disagreement, by enumeration over the 5 reachable (state, touched) pairs rather than by case.
- **Rule 7 itself is right.** `skipIf: exists` with `exists` never passes. With `touched` it passes only on disagreement. D58 closes it. The layer-1 count 108 minus 18 = 90 is correct.
- **Stale table.** The `skipIf: exists` column of the §3 table (`02-semantics.md:60-68`, `expect: touched`) shows results for a config that rule 7 now rejects. Remove the column or mark it rejected.

### P2-5: `{..,a}` and `{.,a}` slip past the `.`/`..` rejection (*unspecified*)

- **Input.** Both globs are accepted by `compileGlob` and create a `..` segment after expansion that can never match. That is the silent never-matches class D51 set out to remove. It is the same for a whitespace-only glob or one with trailing spaces (`" "` and `"src/** "` pass).
- **Message quality.** `{a,b/c}` is rejected, but with "unclosed {" rather than naming the `/`, even though `04:72` promises rejection "by name".
- **Fix.** Check expanded alternatives as segments. Reject leading or trailing whitespace. Name the `/` case.

### P2-6: `related` can resolve to something odd without a load-time check (*unspecified* / *matter of taste*)

- **Partial segments.** `{dir}` and `{**p}` need not fill a whole segment. `related: "{dir}.md"` gives `.md` at the repository root. `pre{p}post` with `a/b` gives `prea/bpost`, with an empty value `prepost`. Rule 3 only limits how many slash-capable variables appear.
- **Empty at runtime only.** `related: "{dir}"` for a root source resolves to empty, and rule 11 fires only when such a file changes. This is statically detectable when a template contains only slash-capable variables.
- **Fix.** Require `{dir}` and `{**x}` to fill whole segments in `related`, or document and test the rendering.

### P2-7: a legitimate file name aborts the run (*matter of taste*, verify the intent)

- **Input.** `match: "{n}x.go"`, `related: "docs/{n}/index.md"`, source `.x.go`. n=`.`, so rule 11 fails the whole run with exit 2. `..x.go` does the same. This is the correct fix for the escape (review 5 P1-3). The cost is that one dotfile aborts every rule's result.
- **Suggestion.** Consider a per-instance error instead of a run-level one.

### P2-8: OS dependence is not closed (*unspecified*)

- **Case.** "Matching is case-sensitive" (`04-grammar.md:48`), but state `exists` reads the working tree. On macOS or Windows a related template `docs/Readme.md` against an on-disk `README.md` is `exists` yet never `touched`. On Linux CI it is `missing`. The limits list covers only case-only renames.
- **Characters.** A captured value can carry `\` or `:` (a legal Linux file name, `a\b:c.go`, resolves to `docs/a\b:c.md`; the simulator accepts it). Rule 11 does not check them. On Windows `\` is a separator and `C:x` is a drive-relative path. Trailing dots and spaces and NFD normalization are also unaddressed.
- **Fix.** State which of these are not defended, as with the case-only rename item. Better, have rule 11 also reject a value containing `\` or a drive-letter prefix.

### P2-9: v3 input cleaning is a compatibility break (*unspecified*)

- **Evidence.** Today `NormalizePath` (`internal/matcher/normalize.go`) trims whitespace, turns `\` into `/`, strips `./` and cleans. v3 (`02-semantics.md:161-166`) rejects `./src/a.go`, `a//b` and `..` with exit 2.
- **Scope.** The rule appears only under `--files-from`. It is silent on `--files`, on CRLF line endings, on blank lines, and on `C` or `U` statuses. The simulator crashes on a blank line (index out of range at `main.go:569-575`), a model defect.
- **Fix.** List it in `06-cutover.md` or normalize leniently. Specify `--files`, CRLF and blank lines.

### P3

- **Exclude equals a match glob.** `match: ["src/**","lib/**"]` with `exclude: "src/**"` is rejected as "removes every match" (`main.go:712`). It removes only that glob's matches. The rule is stricter than its stated reason.
- **Constant `related` at load.** Whether `exclude` counts is unspecified at load but applied by the simulator.
- **`*` is lazy, a preceding named capture is greedy.** `src/*-{n}.go` on `a-b-c.go` binds `n=b-c`. `src/{m}-{n}.go` binds `m=a-b`, `n=c`. That is defined, but it surprises a regex-minded author. Worth one sentence.
- **Capture-name charset.** "Letters" is ambiguous. The simulator accepts ASCII only.
- **Counts.** `07-conformance.md` says "scenarios 14 to 92" and the grammar says 87. The numbering has gaps.

## Closure check, D51 to D58

| Decision | Verdict |
| --- | --- |
| D51 dialect | Mostly closed. Gaps: P2-5, P2-3. Repository-usage claim true. |
| D52 binding | **Not closed** (P1-1). Border restriction removes within-segment ambiguity but not `**` ties. |
| D53 kinds | Closed. Scenarios 64, 65. |
| D54 one pass, rule 11 | Closed for `.`/`..`. See P2-6, P2-7, P2-8. |
| D55 ordering, recipe | Ordering claim is true. Recipe weak (P2-2). |
| D56 plain drops | **Not closed** (P1-2, P2-1). |
| D57 duplicate lines | Closed in text. Model `Fields` bug (P2-2). |
| D58 validation | Closed for stated items. Gaps: P2-4, stale §3 table. |

## Specific scepticism, short answers

1. Longest-first is deterministic only without a bare `**` around the capture. It is surprising only in the `*` versus named-capture laziness case (P3).
2. Dialect complete? Gaps: P2-5, P2-3, whitespace. Nothing in the four presets, `e2e/projects/*`, `testdata/` or `docs/site` uses a rejected form (grepped).
3. Rule 11: ordering is right (substitute, drop empties, then check each segment). The OS issues are P2-8.
4. "A skipped instance cannot hide a broken rule" is **true** (scenarios 88, 89, and my P8/P9). What it does not cover is a source dropped in plain mode or a source absent from disk. The recipe exercises tracked, on-disk sources of every rule only.
5. Plain differences: rename, copy, source-only-listed and conflicting lines behave as written. Copy lines (`C100`) are an error, which is arguably friction (taste). The gap is P1-2.
6. See P2-4. "Exclude identical" is stricter than its justification (P3). Near-identical excludes that cover everything are accepted, as the text already says.
7. See the table above.

## Cases tried

All run through the author's simulator (`dt` binary built from a copy of `main.go` with `main` renamed).

- **Ties and unicode.** C1/C1b/C1c (the tie above), C2 (bytes versus characters), C3/C4 (non-ASCII literal and `?`: no match).
- **Syntax edge cases.** C5 dotfile matched by `*` (correct), C21 (`{..,a}`), C22 (single `*`), C23 (`src/**` on a file `src`: no match, as specified), P16 (`{src,lib}`), P17/P18 (alternation with prefixes).
- **Resolution.** C6/C7 (`.`/`..` values, rejected correctly), C8 and Q4 (`related: {dir}`), C9 (`\` and `:` in values accepted), C24/P14 (`{**p}` zero, root and trailing, correct), Q1/Q2 (partial-segment variables).
- **Binding behaviour.** C13/C13b (lazy `*` versus greedy capture), P10/P15 (empty and leading-dot captures).
- **Plain versus git mode.** C15/C15g/P1/P1g/P2 (deleted related, plain versus git), C16/C16p (source `A` absent from disk), P6 (source dropped), P3 (rename destination), C27 (rename away as related, correct: removed), P4 (conflicting lines, error), P5 (identical lines, harmless), P7 (`C` status, error), C25/Q5 (spaces and quoting).
- **Validation and ordering.** C17/P8/P9 (rule 9/10 hidden by `skipIf` or `changeType`: not hidden, correct), C18/C19/Q3 (never-pass combinations), C20 (exclude equals one glob), C26 (`{dir}` in different directories: two instances, correct), P12/P13 (constant related and exclude, `{**p}` spelling: rejected), P11 (case variants), Q6 (ties inside two globs).
- **Dialect table.** 61 glob strings through `compileGlob` against the dialect table. Everything matched the text except `{..,a}`, `{.,a}`, `" "` and the `{a,b/c}` message.
- **doublestar v4.10.0 probe.** Confirmed `src/**` matches `src`, `a**b` reads as `*`, `[a-z]` is a class, `*` matches dotfiles, `./src/*.go` does not match, and `?` and literals are rune-aware. The text's compatibility claims are right.

## Differential test

I wrote an independent rune-based matcher from the Binding paragraph alone, without reading `bind`, and diffed it against the simulator's `bind` on generated globs and paths.

- **Setup.** 4,000 + 30,000 + 30,000 random dialect-valid globs over an alphabet of `a b - . x`, `*`, `?`, three capture names, `{a,b}`, `**` and `{**p}`; random 1 to 4 segment paths. About 25M (glob, path) pairs, around 2.07M of them matching.
- **Match decision.** 0 disagreements.
- **Bindings.** 0 disagreements wherever the answer was unique.
- **Ties.** 69 ties in the first sample, all with a bare `**` around a capture. The simulator picks by enumeration order (first parse found), which the text does not state. Without `**`, no ties in 12M cases. With `{**p}` and no `**`, no ties in 12M cases.
- **Caveat.** ASCII only. My matcher and the simulator share the same reading of "longest", so agreement shows the text pins the answer when unique, not that "longest" is the right rule. The unicode divergences are in P2-3.

Files for reference: `docs/design/v3/04-grammar.md`, `docs/design/v3/02-semantics.md`, `docs/design/v3/experiments/sim/main.go` (lines 264, 302, 569, 767, 712), `docs/design/v3/experiments/sim/scenarios.yaml`.

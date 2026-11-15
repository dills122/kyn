# Independent review 5 of the v3 matching section

| | |
| --- | --- |
| Source | a fresh-context subagent given `docs/reviews/v3-config-review-5-bootstrap.md` and no prior context, 2026-11-14 |
| Reviewed | `f9b2914`, branch `design/v3-config` |
| Dispositions | `docs/design/v3/00-workplan.md`, section Reviews |

Verbatim as delivered. Scope was the matching section only. Line citations refer to that commit. The reviewer's scratch files were under its own `$TMPDIR` and are not preserved; the cases it describes are reproducible from the text.

---

# Review 5: v3 matching section (read-only)

Scratch work is in `/var/folders/lg/j56llg991d1gpjvtjh8l9gbh0000gn/T/kyn-r5-review/`. The sim copy there (`main.go`) differs from the original by one line that also prints the resolved related path. The case files are `mine.yaml` to `mine5.yaml`, plus `ds/` (doublestar probes) and `diff/` (differential test). The repository was not touched.

## Verdict

**Not ready to implement from the text.** The section is close, and the common cases are right. But I found seven P1s:

- The greedy rule is stated only for captures. For wildcards and alternation, the simulator's binding differs from what the text says.
- Order dependence leaks in through alternation order and through capture kind.
- Capture values can fabricate `..` segments.
- Glob features that work today silently stop matching.
- The suggested dry-run check cannot exercise rule 10 for `changeType` rules.
- The plain-path "exactly two places" claim is wrong.

Each is fixable in text plus a few simulator lines.

## Findings

### P1-1 *Wrong* (simulator contradicts the text) and *unspecified*: "greedy" is defined only for captures

Evidence: `04-grammar.md:70`, `:85-89`. Simulator: `main.go:107-115` (`**` and `*` become regexp-greedy), `:130-139` (alternation becomes a regexp alternation).

| Case | Text says | Simulator gives |
| --- | --- | --- |
| `src/*{n}.go` on `src/abc.go` (A02) | the capture takes as much as it can, so `n=abc` | `n=c`, because `*` took `ab` |
| `a/**/{**p}/f` on `a/b/c/f` (A09) | greedy capture, so `p=b/c` | `p=""`, because plain `**` took everything |
| `src/{x,xy}{n}` vs `src/{xy,x}{n}` on `src/xyz` (G01/G02) | `n=yz` for both, since the match set is identical | `n=yz` for the first, `n=z` for the second |

- A backtracking-regexp implementer and a hand-written matcher following the text would bind differently on the first two cases.
- On the third, reordering alternatives inside braces changes a binding. That breaks the "order never decides" principle at `04-grammar.md:108`.
- The text's claim that greedy "only matters when a literal repeats" (`:85`) is false for all of these.

Required fix:
- Define binding formally: among all parses of the path, choose the one that maximises capture lengths left to right. State that `*`, `?`, `**` and alternation carry no priority.
- Add conformance cases for `*{n}`, `**/` next to `{**p}`, and alternation order.
- Alternatively, reject globs where a capture borders a wildcard or alternation.
- Fix the simulator, or add an enumerating reference matcher.

### P1-2 *Wrong*: rules 2 and 3 compare capture names, not kinds, so reordering `match` changes validity

Evidence: `04-grammar.md:213`, `:214`; `main.go:401` (names only), `:424` (reads only `c.match[0].multi`).

Config (F01a/F01b): `match: ["src/{**p}/{n}.go", "lib/{p}/{n}.go"]`, `related: "{dir}/{p}/{n}.md"`.
- With the multi-segment glob first, the config is rejected (two variables can contain `/`).
- With the globs swapped, it loads. For `src/a/b/c.go` the instance name is `src/a/b/a/b/c`, which is the very ambiguity `03-instances-and-reports.md:72-73` says the one-slash-variable rule prevents.

So this breaks both the "order never decides" claim and the instance-name uniqueness proof.

Required fix:
- In rule 2, require each name to have the same kind (single segment or `**`) in every glob.
- Make the simulator check every glob, not just the first.
- Add a scenario.

### P1-3 *Unspecified* (and unsafe): captured `{name}` values can be `.` or `..`, which yields `..` segments after substitution

Evidence: only the template is checked (`04-grammar.md:129`, `:214`). `03-instances-and-reports.md:42` implies safety. Simulator: `main.go:545` runs `Clean` with no escape check.

| Case | Config | Result |
| --- | --- | --- |
| A18 | `match: lib/{n}.go`, `related: {n}/{n}/x.txt`, file `lib/...go` | `n=..`, resolves to `../../x.txt`, outside `--cwd`. The simulator emits it without error. |
| F03 | `related: o/.{n}/x`, file `lib/..go` | `n=.`, resolves to `x` |
| F12 | same `{n}.go` glob | `...go` gives `n=..`; `..go` gives `n=.` |

- The existing `NormalizeRelativePath` (`internal/matcher/normalize.go:21-35`) would reject the escaping path, but the spec never says to apply it. It also never says which exit code, 2 or 3, applies to an error caused by data.
- `{**name}` and `{dir}` cannot carry `..`, because inputs are normalised first. The spec does not cite that normalisation.

Required fix:
- Say that the resolved path is validated after substitution: relative, no `..`, non-empty, still under `--cwd`, with a named error and exit code.
- Alternatively, forbid a capture that binds a whole segment of only dots.
- Add conformance cases for both.

### P1-4 *Wrong* narrowing, not all stated: dialect compatibility with doublestar

Evidence: `04-grammar.md:52-62` lists a narrower set and states only the literal-brace loss. Probes are in `ds/` and `diff/`.

Constructs that doublestar v4.10.0 matches today, and what v3 does with them:

| Today (doublestar) | v3 text | Simulator |
| --- | --- | --- |
| `[ab]`, `[!a]` match | not mentioned | treated as literal text (diff corpus), so they never match |
| `\*`, `\{a\}` escapes | not mentioned | `\{` makes the glob a config error (B04) |
| `{a*,b}` wildcard inside alternative | "literal alternatives" | `*` is literal, so `x.ts` no longer matches (B02) |
| nested `{a,{b,c}}` | not mentioned | silently never matches (B01) |
| `{**,x}` | not mentioned | silently never matches |
| `src/**` matches the file `src` | "everything below the directory" | does not match |
| `*/**` matches `a.go`; `**/a/**` matches `a` | not mentioned | does not match |
| `a**b` acts as `*` | not mentioned | acts as `.*`, so it crosses segments (A16: `a**b/{n}` matches `ax/xb/f`) |
| `src/{foo}/x.go` matches literally | becomes a capture | loud error only via rule 4 (B08) |

- Rules 2 and 3 do not validate any of these in `match`. So a glob that matches today can silently never match in v3, and a policy tool then passes where it used to fail.
- The only stated narrowing is the literal brace (`:62`). `05-compatibility-matrix.md` and `06-cutover.md` do not mention globs.

Required fix: either reject `[`, `]`, `\`, nested braces, and `* ? /` inside alternatives at load, or support them. Say that `**` must be a whole segment. State each narrowing in `05` and `06`, including the `src/**` and `*/**` edge behaviour.

### P1-5 *Unspecified*: ordering of rule 10 against `changeType`, so the suggested dry-run check misses rule 10

Evidence: `04-grammar.md:235-241`. The simulator skips on `changeType` at `main.go:524-533`, before rule 10 at `:546`.

- Case H01: `match: src/{n}.go`, `related: src/{n}_x.go`, `changeType: [added]`, plain mode with `src/a.go` and `src/a_x.go`.
- Result: both instances are `SKIPPED`, with no error.
- The same rule without `changeType` (H02) correctly gets the rule 10 error.
- In plain mode every status is `modified`. So `git ls-files | kyn check --files-from - --dry-run-resolve` never exercises rule 10 for any rule using `changeType: [added]` or `[renamed]`.
- Rule 9 is checked before `changeType`, so it is exercised (H03).

Required fix: specify that rules 9 and 10 are evaluated for every matched instance before `changeType` and `skipIf`. Pin it in the conformance suite. Reword the claim at `04-grammar.md:236` ("exercise every rule"), because the check only samples existing tracked files.

### P1-6 *Wrong*: "exactly two places" in the plain-path table

Evidence: `02-semantics.md:161-175` against `:187-189` and `:136-140`.

- In git mode, deleted sources never create instances.
- In plain mode, `git diff --name-only` lists deleted files, and they are treated as `modified`.
- Case D09: `src/a.go` is deleted, `src/a_test.go` is untouched, and the plain list contains `src/a.go`. The result is `FAIL not touched`. Git mode gives no instance.
- So a third difference exists, and it flips pass to fail in the common CI pipeline.

Required fix: add the row, or specify that plain-mode sources not present on disk are dropped. Either way, state it in `§6`.

### P2-1 *Unspecified*: conflicting or duplicate `--files-from` lines

Evidence: `02-semantics.md:142-159` is silent. Simulator: `main.go:289-296` (last line wins).

- Case D01/D02: `A src/a.go` then `M src/a.go` skips under `changeType: [added]`. The reverse order applies the rule. The result depends on line order, which is a determinism hole.
- Case D03: `D o/a` plus `A o/a` with the file absent from the tree gives state `removed` and touched together. That contradicts `02-semantics.md:98-99`, which says removed and touched is unreachable.
- Fix: define the precedence, or reject duplicate paths.

### P2-2 *Unspecified* and *wrong*: validation gaps

- **`skipIf: exists` with `expect: touched`, and `[missing, touched]`** (C01, C02). Both load, and both can pass only in the diff-and-tree-disagree row (`02-semantics.md:101`). Rule 7 (`04-grammar.md:218`) catches only `exists`.
- **Constant `related` inside the rule's own `match`** (C04). Example: `match: src/**`, `related: src/CHANGELOG.md`. It is decidable at load but deferred to rule 10, and it then errors on every change, not "someone's unrelated change" (`:235`).
- **`exclude` covering all of `match`** (C03), and `match` globs like `../x/*.go`, `/abs/*`, `a//b` (B13). They load silently and can never apply.
- **`expect: none` with a capture and no `related`** (B14). Rule 4 rejects it. That is probably intended, but it is unstated, and `04-grammar.md:200` says `related` is optional.
- **`related` that resolves to empty, `.`, or a directory.** Example: `related: "{p}"` with `p` empty (C08) resolves to `.`, and the text is silent. A directory can never be `touched`.
- **`related: ""`** is accepted as absent when `expect: none` (C05).
- **"At most one variable that can contain `/`"** (`:131`) does not say whether a variable used twice (`t/{p}/{p}/x`, A19) counts once or twice. The simulator counts distinct names.
- **A variable glued to literal text** (`t/x{p}y/{n}`, A20) yields `t/xa/by/x` for a multi-segment value. Presumably intended, but unstated.

### P2-3 *Unspecified*: single-pass substitution

Evidence: simulator `main.go:542-544` uses sequential `ReplaceAll`.

- Case A23: source `src/{dir}.go`, `related: o/{a}/{dir}`. The simulator gives `o/src/src`. The right answer is `o/{dir}/src`.
- Case A24 is similar.
- The text says only that no literal brace can be matched in globs. Say that substitution is single-pass and that values may contain braces.

### P2-4 *Unspecified*: empty leading variable and "cleaned"

Evidence: `04-grammar.md:134`.

- `path.Clean("/README.md")` stays absolute.
- Today's `template.go:18-20` avoids this by substituting `.`. The existing `NormalizeRelativePath` rejects absolute paths.
- An implementer reusing today's helpers would reject root `{dir}/README.md` (`03-instances-and-reports.md:56`). The simulator trims the leading `/` at `main.go:545`.
- Say "drop empty components before cleaning" and add a conformance case.

### P2-5 *Unspecified*: case-only rename and touched-but-missing related

- **Case-only rename of a related file** (D07). The simulator, being case-sensitive, reports `removed by this change`. On a case-insensitive filesystem the existence check would say `exists`, so the outcome of `expect: missing` or `exists` differs between macOS and Linux CI. The text says only that "matching is case-sensitive" (`:48`).
- **Related in the diff but absent from the tree, with `skipIf: missing`** (D08). The rule is skipped as `related is missing` although the file is in the change. This follows the table but is surprising. It is a *matter of taste* unless the intent is stated.

### P2-6 *Unspecified*: the `git ls-files` recipe and quoting

Evidence: `04-grammar.md:239-241`, `02-semantics.md:157-159`, `01-observed-semantics.md:369-389`.

- `git ls-files` quotes non-ASCII paths by default. `--files-from` does not support quoted lines.
- So the recommended check silently fails open (OS16) on repositories with such paths.
- Give a safe form, such as `git -c core.quotePath=false ls-files`. Note the residual limit for paths with tab, quote, or backslash characters.
- Also say that the check does not flag a `match` that matches no tracked file.

### P2-7 *Unspecified*: backslashes in `related`

The simulator keeps `o\{n}` raw (C18). Today's `NormalizePath` turns `\` into `/`. Say which applies.

### Matter of taste

- `related` cannot name `[id]`-style files, which Next.js repositories have, or literal braces (`:129`).
- `{dir}` plus `{**p}` forces repeating the literal prefix (E05).

## The seven sceptical areas

1. **Greedy.** Not well defined across wildcards or alternation, and the result depends on how the matcher is built (P1-1). For two captures separated by a repeating literal, it held up: `{name}.{kind}.ts` gives `user.profile`/`component`, and `{a}_{b}_{c}` gives `x_y/z/w`.
2. **Dialect.** It does narrow silently, and only the brace loss is stated (P1-4).
3. **Uniqueness of instance names.** Held up, with one caveat. I tried to construct a collision with empty values, `.`, `(all)`, a trailing slash, and the join character:
   - Values `.` (E01, E02) and the literal `(all)` (E03) do not collide.
   - Single-segment values are non-empty and slash-free.
   - Inputs are normalised, so `.` cannot arise from `{**name}` or `{dir}`.
   - The only break is P1-2, where kind mismatch across globs defeats the one-slash-variable rule.
   - The claim should cite the input normalisation.
   - Related to uniqueness: distinct keys can resolve to one path through `.` values (`./x` and `x/.` both give `o/x`, E01). This is the same class as the documented `docs/{a}-{b}.md` limit.
4. **`{**name}` zero segments.**
   - Held up: leading, middle and trailing zero-segment cases resolve as a developer would expect (A11, A12, A13, F10, A22).
   - Held up: `{**name}` and `{dir}` cannot escape, since inputs are normalised.
   - Gaps: the empty leading variable (P2-4), a `related` that resolves to `.` or a directory (P2-2), and variables glued to literals (P2-2).
   - The escape is through `{name}` (P1-3), not `{**name}`.
5. **Three states and `skipIf`.** The §3 table is internally correct (C19, D05, D06). Gaps:
   - D01/D02: conflicting lines for one path.
   - D03: removed and touched together.
   - D07: case-only rename.
   - D08: touched but missing.
   - D09: deleted source in plain mode.
   - Copy status `C` is rejected as the text says (D04).
   - Rename onto an existing path shows as D plus M in git. I reasoned this, but did not run real git.
6. **Validation completeness.** Gaps are listed in P1-2 and P2-2. No two rules contradict each other, but rule 4 interacts with `expect: none`. Rules 5 and 7 are not redundant, but are incomplete.
7. **Resolve-time errors.** The dry-run recipe is not sufficient (P1-5, P2-6). It covers only existing tracked paths, only in plain mode, and it skips rule 10 under `changeType`.

## Cases tried

I ran about 105 simulator scenarios. Counts by file:

| File | Scenarios | Topic |
| --- | --- | --- |
| `mine.yaml` | 45 | A01–A24 binding and `related`; B01–B21 dialect and grammar |
| `mine2.yaml` | 39 | C01–C19 validation; D01–D09 states and modes; E01–E05 names |
| `mine3.yaml` | 12 | F01–F12 |
| `mine4.yaml` | 5 | G01–G05 greedy and alternation |
| `mine5.yaml` | 3 | H01–H03 rule 10 ordering |

Beyond those, I ran about 48 doublestar glob and path probes, and a differential test of about 30 globs against 32 paths between the simulator's compiler and doublestar.

Cases that behaved correctly:
- A01, A03–A08, A10–A13 (greedy and zero-segment bindings).
- A19–A22.
- B05, B06, B07, B09, B10 (alternation with slash, empty member, invalid names).
- B15, B16, B19–B21 (rejections).
- C06, C07, C12–C17, C19, D04–D06.
- E01–E04.
- F02, F04, F05, F08–F11.
- G03, G04, H02, H03.

Plain-mode `{**name}` behaved as specified.

## Strengths and what I verified

- Rules 2, 3, 4, 5, 6 and 8 fire correctly and with clear messages. Adjacent captures, reserved `dir`, a second `{**}`, `x{**p}`, `{**p}x` and unused captures are all rejected as specified.
- Zero-segment `{**name}` in the leading, middle and trailing positions resolves sensibly.
- The §3 `skipIf` table and the `removed` state are internally consistent. D03, D06 and C19 matched `02-semantics.md`, apart from the gaps above.
- Rule 9 ordering and rule 10 self-match detection work, including with `exclude`.
- Instance-name uniqueness holds for normalised inputs and a single slash-capable variable.
- I verified by running the simulator on my own scenarios, a doublestar v4.10.0 module in `ds/` (module cache, offline), and a differential harness in `diff/` that reuses the simulator's `compileGlob`.
- I confirmed existing behaviour by reading `internal/matcher/normalize.go`, `internal/matcher/glob.go`, `internal/family/template.go` and `internal/changes/manual.go`.

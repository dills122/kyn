# Independent review 7 of the v3 matching section

| | |
| --- | --- |
| Source | a fresh-context subagent given `docs/reviews/v3-config-review-7-bootstrap.md` and no prior context, 2026-11-15 |
| Reviewed | `580896e` (scope frozen in `154ffaf`), branch `design/v3-config` |
| Dispositions | `docs/design/v3/00-workplan.md`, section Reviews |

Verbatim as delivered. Scope was the matching section only. Line citations refer to that commit. The reviewer's scratch files were under its own `$TMPDIR` and are not preserved; the cases it describes are reproducible from the text.

---

# Independent review 7 of the v3 matching section

State reviewed: `154ffaf` (design `580896e`), branch `design/v3-config`. Read-only. Scratch is `/var/folders/lg/j56llg991d1gpjvtjh8l9gbh0000gn/T/tmp.nbkRpjuJ5Z` (`diff/` has my matcher, rule 7 checker and harness; `sc/` has 3 scenario files). The repository is unchanged (`git status` is clean). The author's `go run ... scenarios.yaml` is green: 115 scenarios, 0 mismatched.

## Verdict

**Not ready to freeze, but close.** There is no P0. There is one P1 and several P2s. Binding, rule 7 and the dialect hold up under differential testing. The defects are in the new text around plain-path input, in input parsing, and in places where the new validation is too weak or the model is laxer than the text.

- **Genuinely closed from review 6:**
  - Binding totality (P1-1).
  - Code points (P2-3).
  - The stale `skipIf: exists` column.
  - `{dir}.md` and `related: "{dir}"`.
  - The `git ls-files` recipe's vacuous pass, mostly.
  - The compatibility-break disclosure in `06`.
  - Review 6's P3 items.
- **Not closed:**
  - P1-2 and P2-1, the plain-path derivation, which now contradicts its own row 4.
  - P2-5, whitespace, which is closed only at segment edges.
  - P2-4, rule 7, which is closed in one direction only.

## Findings, ranked

### P1-1: Row 4 contradicts the derivation that the text says wins (*wrong*)

- **Evidence.**
  - `02-semantics.md:189-193` says the differences "follow from two facts" and that "when a row and the derivation disagree, the derivation is right". The facts are: the list is the whole change set, every path is `modified`, and a path not on disk is `missing`.
  - `02-semantics.md:200` row 4 says a plain-mode listed source that is not on disk is **dropped**.
  - The derivation says the opposite. A listed, `modified` source that is absent from disk is just a source, so it makes an instance, exactly as git-mode `M x` does.
  - Nothing in the two facts derives a drop. The drop is a third, unstated fact.
- **Model.** The model follows the row, not the derivation (`main.go:902`). So the oracle and the spec's stated precedence disagree.
- **Input.**
  - Config: `match: src/{n}.go`, `exclude: src/*_t.go`, `related: src/{n}_t.go`, `expect: exists`.
  - Disk: empty. Change: `src/a.go`.
  - Plain mode gives PASS with zero instances. Git mode `M src/a.go` gives `FAIL a does not exist [missing]`. The text says the derivation is right, which would mean plain also gives FAIL.
- **Rationale is weak.** The stated reason is "a deleted source and a typo look the same" (`02-semantics.md:207-208`). But the same text refuses to infer deletion for a related path because it would "turn typos into deletions", and then does infer it for sources.
- **Consequence.** `--name-only` after deleting a handler and its test silently evaluates nothing. In git mode a source absent from disk still makes an instance.
- **Fix.** Pick one:
  - (a) Make "plain mode drops sources not on disk" a stated third fact, and delete the "derivation wins" sentence.
  - (b) Keep sources, and let a missing source fail naturally.
  - Option (a) matches the `ls-files` recipe, which depends on the drop. In either case rewrite the "follow from two facts" claim.
- **Related.** `02-semantics.md:218` cites "§6, row 2" for "every source is `modified`". That is row 3 (P3 stale reference).

### P2-1: Whitespace rejection misses the most likely typo, `{ts, html}` (*unspecified*, a wrong rationale)

- **Evidence.** `04-grammar.md:77` rejects a segment "with leading or trailing whitespace" because it is "almost always a typo". The model checks the glob text before alternatives expand (`main.go:218`).
- **Input.**
  - `match: "src/**/{n}.component.{ts, html}"` is accepted. The alternative is `" html"`, so `b.component.html` never matches.
  - Run through the simulator: only the `.ts` source made an instance, and there was no error.
  - `src/{n}.{ts,html }` is accepted too, and matches nothing.
- **Why it matters.** This is the never-matches class that D51 and D63 set out to remove, and a space after a comma is the most likely way to produce it.
- **Fix.** Reject whitespace at the edge of any alternative, or anywhere inside `{}`.
- **Same class, accepted silently.**
  - `related: " docs/{n}.md"` and `"docs/{n}.md "` (the dialect rejects the glob form but not `related`).
  - A glob with a BOM (U+FEFF) or a control character, which `TrimSpace` does not trim. A glob with U+00A0 is rejected, because `TrimSpace` is Unicode-aware while the text never defines "whitespace".

### P2-2: Rule 7 is closed in one direction only; a vacuous always-pass rule is accepted, and `02` recommends it (*unspecified*)

- **Input.** `skipIf: exists` with `expect: missing`. The remaining pairs are `(removed,no)`, `(missing,no)` and `(missing,yes)`, and `missing` is true on all three. The rule can never fail.
- **Simulator.** It accepted the config, and the result was `PASS ... [missing]`.
- **Text.** `02-semantics.md:72-73` says "`skipIf: exists` is for `none`, `untouched` and `missing`".
- **Why it matters.** This is the same defect class as an informational rule with an error severity, which rule 8 and OS14 forbid.
- **Fix.** Either reject a rule that is true on every remaining pair, which would also fix `02:73`, or state that it is allowed and say why.
- **Checked.** Of the 24 non-contradictory (skipIf, expect-subset) combinations, this is the only always-pass one.

### P2-3: The `--files-from` and input grammar is under-specified, and the model is laxer than the text (*unspecified*, partly *wrong* in the model)

- **Mixed files.** A TAB marks a name-status line and no TAB marks a plain path, line by line (`02:153`). A file mixing both (for example `git diff --name-status` plus untracked files, a realistic use) has no stated mode. Do plain lines get `modified` and row 4's drop while `M` lines keep absent sources? The model cannot express it, because `mode` is per scenario (`main.go:707`, `strings.Fields` fallback at `:702`).
- **Status strictness.** The text says "one of A, M, D, T, or R followed by a score".
  - The model reads only the first byte (`main.go:707`). `MODIFIED\tsrc/a.go`, a bare `R\ta\tb` and `M\ta\tb` (extra field ignored) are all accepted.
  - The text never says what a score is (digits, range) or whether extra fields are errors.
- **Blank lines.**
  - A whitespace-only line `"   "` is a hard error in the model, while "Blank lines are ignored" (`02:166`) suggests it is ignored, as v2 did.
  - A UTF-8 BOM on the first line is not whitespace, so the path is silently dropped (a never-match).
- **"Clean path" scope.**
  - It is ambiguous whether "no leading or trailing whitespace" applies to the whole path or each segment. The model uses the whole path, so `src/a /b.go` is accepted.
  - An interior TAB is accepted by `--files`, but `--files-from` cannot carry such a path.
  - The drive regex `^[A-Za-z]:` rejects any `x:name` file, although the text says "`C:`".
  - The resolved-path check (rule 11) shares only part of the predicate. For example, a captured value gives `related: "docs/{n}"` the resolved path `docs/ ` with trailing whitespace, and nothing complains.
- **`U` status.** `02:176` says "Kyn never asks git for copies, so `C`, `U` and `X` are errors". That does not follow for `U`, which is an unmerged path in a conflicted working-tree diff. A pre-commit hook mid-merge fed by `git diff --name-status` would exit 2. Say so or accept `U`.
- **Fix.**
  - Define the predicate once, per segment, including what whitespace means.
  - Define mixed files.
  - Define the status grammar exactly.
  - Make the model strict.

### P2-4: `src/**/{feature}/**` binds on directory-name length, which makes the "realistic glob" claim doubtful (*matter of taste*, leaning *wrong*)

- **Evidence.** D59 (`00-workplan.md:98`) keeps the glob because "`src/**/{feature}/**` is a realistic glob". Binding is longest first (`04:107-117`).
- **Input.** `match: "src/**/{feature}/**"`, `related: "docs/{feature}.md"`.
  - `src/billing/ui/a.ts` binds `billing`.
  - `src/b/ui/a.ts` binds `ui`.
  - `src/auth/login/a.ts` binds `login`, and `src/app/features/login/x` binds `features`.
- **Why it matters.** Renaming an unrelated directory changes which instance a file belongs to. The tie-break makes this deterministic but not meaningful.
- **Fix.** Either reject an unanchored `{n}` between two bare `**`, or bind it by position (shortest or leftmost) rather than by length, or document that such globs must be anchored by a literal.

### P3 findings

- **`02:72` plus `expect: missing`** has a failure message that is not in the §5 message table (the simulator prints "exists").
- **Rule 11's `\` check is unreachable.** Input rejects `\` and `related` rejects `\`, so no captured value can contain one. Removing the check (`main.go:560`) keeps all 115 scenarios green. The only reachable parts are `.`, `..`, empty and the drive prefix.
- **Mutation survivors.** Weakening `passability`'s `untouched` to ignore `removed` leaves 0 scenarios red, so there is no scenario for `[untouched, missing]` with `skipIf: missing`. The mutation test is real (the other mutations I tried turn 1-2 scenarios red), but not complete.
- **Templated drive prefix.** `related: "C:{n}"` is accepted at load and then aborts the run at resolve time. It is statically detectable.
- **`{**,a}`** is rejected, but the message says `",a" is not a valid capture name`, not that `*` is not allowed in an alternative.
- **`--files` is comma-separated today.** A path containing a comma cannot be passed, and `06` does not mention it.
- **Rule 7 and the plain modes.** The five-pair set is mode-dependent. In plain modes `(removed,no)` is unreachable, so `skipIf: missing` with `expect: missing` can never pass there. It passes in git mode, so it is not rejected. Stating this costs one sentence.
- **Capture start for empty `{**p}`** is undefined in the text. It does not matter in practice (see differential test 1).

## Closure check

### D59-D65

| ID | Verdict |
| --- | --- |
| D59 binding total | **Closed.** Binding is total (differential test 1: 0 ties in about 20M pairs). P2-4 is a semantic caveat, not a hole. |
| D60 plain derivation, four rows | **Not closed.** Rows 1-3 and the derivation hold (scenarios 94 to 94e; I re-derived all of them). Row 4 contradicts the derivation (P1-1). |
| D61 code points | **Closed.** Differential test 1 includes non-ASCII and combining marks: 0 disagreements. NFD does not match NFC. `?` is one code point. |
| D62 general rule 7, 90 cells | **Closed for never-pass and disagree-only.** The 90-cell count is verified: 3 combos × 6 rows = 18, so 108 − 18 = 90. Always-pass is left open (P2-2). |
| D63 more rejections | **Partly.** `{..,a}`, `{.,a}`, `{a,b/c}`, `{dir}.md` and `{dir}` are closed. Whitespace is closed only at segment edges (P2-1). |
| D64 input cleaning is a break | **Disclosed and mostly complete in `06:18`** (whitespace, `\`, `./`, drive, quoting, UTF-8). Gaps are in the clean predicate (P2-3). I found no real repository config affected. |
| D65 run-level rule 11, OS limits, exclude wording | **Closed.** `.x.go` under `{n}x.go` aborts the run exactly as stated (scenario 111). |

### Review 6 table

| Review 6 item | Verdict |
| --- | --- |
| P1 binding total | Closed |
| P1 plain-path whole list | Mostly closed; reopened by row 4 (P1-1) |
| P2 source absent from disk | **Not closed.** The decision is now explicit but self-contradictory |
| P2 `ls-files` recipe | Closed (`--fail-on-empty`, quoting, "tracked, on-disk sources" stated) |
| P2 bytes vs code points | Closed |
| P2 rule 7 gaps, stale column | Closed, with the always-pass residue (P2-2) |
| P2 `{..,a}`, whitespace, `{a,b/c}` | Mostly closed (P2-1) |
| P2 `{dir}.md`, `related: "{dir}"` | Closed |
| P2 dotfile aborts run | Closed as a maintainer decision |
| P2 OS dependence | Closed as limits. `\` and drive checks work. The `\` check is dead code (P3) |
| P2 compatibility break | Disclosed in `06` (P2-3 for the remaining ambiguity) |
| P3 exclude wording, constant `related`, ASCII names, `*` vs capture, scenario counts | All closed (115 everywhere) |

## Overclaim check

- **"Follow from two facts, not a list"** (`02:189`) is false for row 4 (P1-1).
- **"Binding is total"** (`04:116`) is true.
- **"No configuration in this repository uses a rejected form"** holds for every glob string I could extract from the four presets (`internal/cli/init.go`), `e2e/projects/*`, `testdata/`, `docs/site/` and the recipes. The only compile rejections were v2 related templates (`{dir}/{base}...`), which are not v3 globs, plus shell tokens.
- **"Almost always a typo"** (`04:77`) is not defended where it is most likely (P2-1).
- **The limits list** ("not claimed complete") is still not complete. I found no new limit beyond the gaps above.

## Cases tried

- **Ties and tie-breaks.** `src/**/{n}/**` on `src/auth/user/a.ts` binds `auth`, on `src/user/auth/a.ts` binds `user`, and on `src/auth/login/a.ts` binds `login`. A rule with two globs that tie differently gives rule 9. `**/{a}/**/{b}/**` is deterministic. All correct per the text.
- **Dialect.**
  - Accepted when they should not be: `{ts, html}`, `{ts,html }`, a leading BOM in a glob, a control character in a glob.
  - Correctly rejected: `src/{}.go`, `{**,a}` (message imprecise), a NBSP-padded segment, `{dir}` as a match capture, `src/{n}**`, `{**p}/{**q}/x`, and a mismatch in kind between `{n}` and `{**n}`.
  - Accepted as the text intends: `{a,a}`, duplicate alternatives, and the bare `**` glob.
- **Exclude.** `match: [src/**, lib/**]` with `exclude: src/**` is rejected as pointless (the text is accurate).
- **Resolution.** Whole-segment `{dir}` and `{**p}`, a repeated `{dir}`, and `{p}` twice are accepted and work. `C:{n}` aborts the run at resolve time. A captured whitespace-only segment is accepted. `{n}x.go` on `.x.go` and `..x.go` aborts the run (run-level, as documented).
- **Input.**
  - Rejected with exit 2, as the text says: `./src/a.go`, `src/`, `a:b.go`, a whitespace-only line, a conflicting `D x` and `R x y` pair.
  - Accepted and silently harmless: a BOM line (never matches), `MODIFIED\t...`, a bare `R`, extra fields.
- **Plain vs git.**
  - A deleted related file that is listed makes `touched` pass and `untouched` fail "was touched" (94, 94c). A deleted related file that is not listed makes `skipIf: missing` skip (94e).
  - A source absent from disk is kept in git mode and dropped in plain mode (B1, B2; P1-1).
  - In plain mode `changeType: added` never matches.
- **Rule 7.** Every subset of expectations × all three `skipIf` values was checked against my own checker. The scenario rules were checked too: 102 rules, 0 disagreements, 0 violations that the scenarios expected no error for.
- **Mutation tests I ran on the model.** The author's claim that mutations turn scenarios red holds for the tie-break, rule 7's disagreement branch, the plain drop, byte vs code point, the drive check, `wholeSegments`, the whitespace check, `dotSegment`, CR stripping and git quoting. It does not hold for the `\` check in resolution or the `untouched` branch of `passability` (P3).

## Differential test results

1. **Matcher.**
   - I wrote a rune-based enumerating matcher from the Binding paragraph alone, without reading `bind`. It was diffed against the simulator's `bind` on about 20M (glob, path) pairs.
   - Setup: random dialect-valid globs of 1 to 4 segments, with `*`, `?`, three capture names, `{a,b}` alternation, bare `**` and `{**p}`. Four runs (200,000 candidate globs each, 40 paths each, about half derived from the glob), with and without non-ASCII text (`é`, `日`, combining mark, 4-byte emoji) and with and without bare `**`. About 10.2M pairs were matched.
   - **Match decision: 0 disagreements. Bindings: 0 disagreements.** Cases where the best parse is decided only by the earliest-start rule were common (about 160-200K globs per run), so the tie-break is exercised.
   - **Ties:** 0 cases where two parses had equal capture lengths and equal starts but different values (as the text claims).
   - **Empty `{**p}` start convention:** I changed my definition of an empty capture's start to an extreme value (+1000). There were still 0 disagreements over 2.5M pairs, so the convention does not matter.
2. **Rule 7.**
   - I wrote an independent checker from `02 §4` and rule 7. It was compared with `passability` on every non-empty subset of {touched, untouched, exists, missing} × {none, missing, exists}, excluding the combinations rule 5 rejects first: 24 combos, 0 disagreements. It was compared again on every `skipIf` and `expect` pair in `scenarios.yaml`: 102 rules, 0 disagreements.
   - The only extra finding is that `skipIf: exists` with `expect: missing` is accepted but can never fail (P2-2).

Files: `docs/design/v3/04-grammar.md`, `docs/design/v3/02-semantics.md`, `docs/design/v3/00-workplan.md`, `docs/design/v3/06-cutover.md`, `docs/design/v3/experiments/sim/main.go`.

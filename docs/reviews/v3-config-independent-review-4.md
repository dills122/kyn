# Independent review 4 of the v3 configuration design

| | |
| --- | --- |
| Source | a fresh-context subagent given `docs/reviews/v3-config-review-4-bootstrap.md` and no prior context, 2026-11-14 |
| Reviewed | `73f33dc`, branch `design/v3-config` |
| Dispositions | `docs/design/v3/00-workplan.md`, section Reviews |

Verbatim as delivered. Line citations refer to that commit; the documents have been revised since, and the dispositions are in `docs/design/v3/00-workplan.md`.

---

# Independent review 4: v3 configuration design (revised shape)

Reviewed read-only at `73f33dc` on `design/v3-config`; the worktree is still clean. Scratch files are under `$TMPDIR/kyn-r4-review/`.

## Verdict

**Not ready to approve as a specification yet. Ready to advance once five P1 findings are closed; there is no P0.**

- **What holds:** the semantics (three states, `skipIf: missing`, instance keys, reports, headline) are sound and reproducible, and the common rule really does read aloud.
- **What does not:** the capture grammar, which is the part no one has reviewed, cannot be implemented without guessing. One brace rule silently changes the meaning of an existing glob idiom, and `skipIf: exists` contradicts its own prose.
- **Size of the fix:** all five P1s are document-level decisions; none needs a redesign.

## Cold-written configs

Written from `04-grammar.md` alone, then run through the simulator.

| # | Policy | Config (abridged) | Outcome |
| --- | --- | --- | --- |
| C1 | Any Go file needs its test touched | `match: "**/{name}.go"`, `exclude: "**/*_test.go"`, `related: "{dir}/{name}_test.go"`, `expect: touched`, `skipIf: missing` | **Right.** I guessed that `**/` matches zero segments and what `{dir}` is at the root; `04` states neither. |
| C2 | Schema change needs some new migration | `match: "db/schema.sql"`, `related: "db/migrations/*.sql"`, `expect: touched` | **Wrong, inexpressible.** `related` is a single path, and the simulator accepts the `*` as a literal filename and fails forever. The same shape covers "any src change needs a `.changeset/*.md`". |
| C3 | Applied migrations are frozen | `match: "db/migrations/{name}.sql"`, `related: "{dir}/{name}.sql"`, `expect: untouched`, `changeType: modified` | **Wrong, inexpressible.** Rule 9 rejects it as a self-match, and `expect: none` cannot fail. |
| C4 | Python mirrored tree | `match: "src/{**pkg}/{name}.py"`, `related: "tests/{pkg}/test_{name}.py"`, `expect: touched` | **Silently partial.** `src/app.py` matches nothing and is never checked. Adding `src/{name}.py` as a second glob is forbidden by rule 2, so it needs a second rule. |
| C5 | New command needs a doc page | `match: "cmd/{name}.go"`, `related: "docs/commands/{name}.md"`, `expect: exists`, `changeType: [added, renamed]` | **Right.** I hesitated on whether `changeType` takes a list, and briefly read it as the related file's change type. |
| C6 | Per-package changelog in a monorepo | `match: "packages/{pkg}/src/**"`, `related: "packages/{pkg}/CHANGELOG.md"`, `expect: touched` | **Right, and a gain** v2 could not express; the documents do not claim it. |

Other things I could not answer from `04`:
- whether `[touched, untouched]` or `[exists, missing]` is rejected;
- whether `exclude` may contain captures;
- how to write a literal `{`;
- which characters a capture name may use;
- what a deleted source does (answered only in `02` §7).

The simple cases (C1, C5, C6) were pleasant to write. The friction was all in captures and in policies that have no single related path.

## Findings

### P1

**P1-1. Capture matching is not specified well enough to implement. (*unspecified*)**

`04-grammar.md:46-57` gives a four-row table and one sentence. It leaves these open:

- **A file matching two globs with different bindings.** With `match: ["internal/**/{name}_{handler,service}.go", "internal/**/{name}.go"]`, `order_handler.go` demands `order_test.go`; with the list reversed it demands `order_handler_test.go`. The simulator takes the first matching glob (`sim/main.go:209-216`); the documents never say list order matters. This is the natural translation of v2's optional `stripSuffixes`.
- **Greedy or lazy.** `src/{name}.{kind}.ts` on `user.profile.component.ts` gives `name=user.profile` only because Go's regexp is greedy. The word "greedy" appears nowhere in the set.
- **Adjacent captures.** `{a}{b}` on `abc` is accepted and gives `ab`/`c`.
- **Capture-name characters.** The simulator's `varRe` is `[A-Za-z]+` (`sim/main.go:113`). `{file_name}` is captured in `match` but not recognised in `related`, so `src/a.ts` and `src/b.ts` merge into one `(all)` instance demanding `docs/a.md`.
- **`{**name}` and segment boundaries.** `src/{**pkg}Impl/{name}.java` matches a partial segment, because the simulator uses `(.+)` (`sim/main.go:95-97`), against "whole path segments" at `04:54`.
- **Everything else:** captures in `exclude`, glob metacharacters in `related`, brace escaping, and which glob dialect "standard" means. Today it is doublestar, which cannot capture, so `06-cutover.md:17` "Extend" is really a new matcher.

Required fix: a short normative matching section covering glob precedence (or reject overlapping bindings), the disambiguation rule (or reject ambiguous globs), the name charset, `{**x}` occupying whole segments, and rejection of captures in `exclude` and metacharacters in `related`.

**P1-2. `{x}` and `{a,b}` distinguished by a comma silently changes an existing glob. (*wrong*)**

- Rule: `04-grammar.md:57`.
- Scenario: an author trims `src/**/{name}.component.{ts,html}` to `{ts}`. In doublestar today `{ts}` is the literal `ts`; in v3 it becomes a capture named `ts`. The simulator then makes `src/b.component.scss` demand a story, with no error.
- Required fix: make an unused capture a validation error, which also catches typos. Alternatively give captures a distinct sigil.

**P1-3. `skipIf: exists` on a removed file contradicts the document's own prose. (*wrong*)**

- Prose: `04:113` says it is for acting "only when the related file is absent"; `02:99` says "flag only when the file is absent".
- Tables: `04:106` and `02:46` skip when the file "was removed by this change", when it is absent.
- Scenario: `needs-story` with the story deleted in the same change. The simulator reports `SKIPPED` and no `needsStory` flag. The next change to that component does raise it, and `--files` mode raises it immediately (`02:132`).
- Overloading: the same words mean different sets in the two fields.

| Word | In `expect` | In `skipIf` |
| --- | --- | --- |
| `exists` | `exists` | `exists` or `removed` |
| `missing` | `removed` or `missing` | `missing` only |

- Required fix: make `removed` apply under both gates, so the rule is "a removal never causes a skip". That matches the prose, makes `skipIf: exists` agree with `expect: exists`, and removes one of the two git-versus-plain divergences. Otherwise rewrite the prose and rename the values.

**P1-4. `{**name}` requiring one or more segments leaves top-level files unchecked, and rule 2 blocks the fix. (*unspecified consequence; the direction is a matter of taste*)**

- Scenario: C4 above. Simulator scenario 18 (`scenarios.yaml:216-226`) only uses a nested file, so it hides this.
- Asymmetry: plain `**/` matches zero segments, `{**pkg}/` does not.
- Required fix: either allow zero segments (clean the path, use a placeholder name component) or keep one-or-more and document the two-rule idiom. Add the top-level case as a scenario and a conformance cell either way.

**P1-5. The conformance suite has no matcher layer. (*unspecified*)**

- Layer 1 (`07-conformance.md:45-54`) is gate × assertion × state. Layer 2 lists only "instance shape".
- Nothing would catch a regression in greedy matching, glob precedence, comma-versus-capture, dotted names, or `{**x}` boundaries. That is D27, the largest new mechanism.
- Required fix: add a table-driven capture layer of glob, path, expected bindings or no-match, plus invalid globs.

### P2

1. **`(root)` breaks instance-name uniqueness. (*wrong*)** `03-instances-and-reports.md:49-55` claims names are unique. With `match: "**/{name}.tsx"` and `related: "{dir}/{name}.test.tsx"`, `page.tsx` and `(root)/page.tsx` both render `(root)/page`. `(root)` is a common Next.js route-group name and is reachable with `--cwd app`. The simulator merged them into one instance with two sources. Fix: use a placeholder that cannot be a path component, or render the root as empty.

2. **`..` in `related` re-creates OS13. (*unspecified*)** `03:8-9` defines an instance as one related path; `03:11-12` keys it on variable values. With `{dir}/../README.md`, `pkg/a/x` and `pkg/a/y` give two failures for the same `pkg/a/README.md`. Fix: forbid `..`, or key on the resolved path.

3. **Stale content survives the "current design only" rewrite. (*wrong*)**
   - `05-compatibility-matrix.md:103-110`, under "What still applies to v3", cites `use` alongside inline `match` and a `patterns:` block. `:52-53` and `:95-96` say nothing changes for v1 and v2.
   - `01-observed-semantics.md:33` labels `changed-if-present` and similar as "(v3 name)". `:174-177`, `:201-204` and `:468` argue against the old proposal and "patterns".
   - `02-semantics.md:35` cites `internal/changes/git.go:45` and `:60` for the discarded `D` line and rename source. Those lines are `Path: normalized`; the discards are at `:63-64` and `:48-52`. `01:130` has the same wrong line.
   - `04:38` gives the `severity` default as `error`; rule 8 at `04:157` says `info` is the default for `expect: none`.
   - `04:146` says all eleven rules "fail with exit 2", but rule 11 and half of rule 6 are normalisation. `04:222-223` omits rule 10 from the resolve-time rules.

4. **Capability accounting is short in both directions. (*wrong*)** `04:194-210`, D33.
   - Loss: a capture cannot be constrained to an alternation, so v2's `{ext}` over `*.test.{ts,tsx}` (a Jest snapshot path) needs one rule per extension. `{name}.test.{ext}` also matches `.bak` and its own `.snap`.
   - Behaviour change: v2 `if.kinMissing` fired on a just-deleted file; `skipIf: exists` does not.
   - Gains not claimed: captured directory segments (C6), prefix stripping (`test_{name}.py`), and one failure for a constant related path.
   - Awkward but not a loss: no "some file under a directory" related, and no frozen-file rule.

5. **`untouched` passes for a deleted file, and the set's own scenario falls into it. (*matter of taste*)** `04:97` documents it. Scenario 13 (`scenarios.yaml:156-166`) passes when `0001_init.sql` is deleted, and nothing fires when it is edited without `schema.sql`. Consider making `untouched` fail on `removed`; at minimum change scenario 13 to `[exists, untouched]`.

6. **The simulator is thinner than the prose implies. (*wrong by omission*)** `04:227-237`.
   - It has no expected outputs, so "21 scenarios run" asserts nothing.
   - It does not implement validation rules 2, 3, 5, 6, 8 or 10, or enum checks. My probes got results for `{pkg}/{dir}/…`, `[none, touched]` with `severity: error`, `expect: none` with no message or flags, `../../x`, and `expect: changed` with `skipIf: absent`.
   - It keys instances on the rendered name (`sim/main.go:252`), which hides item 1.

7. **The named harness cannot run layer 1. (*unspecified*)** `07:22-25` and D34 say the suite extends `e2e/e2e_test.go`. That harness has no git support and uses static project directories with `--files`. The 108 cells need generated repositories with deletions, renames and a checkout missing a head-ref file. "All seven output modes" (`07:86`, `:110`) is never enumerated; I count five `check` formats plus two `explain` formats, with `--dry-run-resolve` making eight.

8. **Smaller unspecified points.**
   - Instances filtered out by `changeType` produce no result (`02:142-147`), which is the invisibility D32 fixed for skips.
   - Contradictory lists such as `[touched, untouched]` and always-failing pairs such as `skipIf: exists` with `expect: exists` are accepted.
   - There is no message wording for a `removed` failure; the simulator says "not touched" for a file the author deleted.
   - `--fail-on-empty` ("no family instances match") has no v3 meaning.
   - `--files-from` has no rule for telling a name-status line from a plain path, and no handling of git-quoted paths.
   - Rule 9 fires only at resolve time, so a config can validate and then exit 2 on someone's unrelated change.

9. **Cutover inventory gaps. (*wrong*)** `06-cutover.md:9` says it was measured by search, but it omits:
   - struct-based tests of the replaced packages (`internal/config/validate_test.go`, `migrate_test.go`, `internal/family/resolver_test.go`, `internal/rules/evaluator_test.go`, `explain_test.go`);
   - CLI help and verbose strings that use family/kin vocabulary (`internal/cli/check.go:21,158,160`, `explain.go:16,94`, `init.go:87`);
   - `docs/mvp-v2-tasks.md`.

10. **Two overclaims about evidence. (*wrong, minor*)**
    - `02:48-49` says the gate table is "measured against real git output" by `e9`. `e9` derives the state from real git but computes every verdict, including the "Today" column, in bash (`e9-gate-prototype.sh:50-54`) and never invokes the binary. `e4` and `e8` do measure today's behaviour for those rows.
    - `00-workplan.md:13` says each of the 15 observations has a script; OS2 and OS10 do not.

## The eight sceptical areas

1. **Captures in `match`: did not hold.** See P1-1, P1-2 and P1-4. Dotted values are fine with a single capture (`a.pb.go` gives `a.pb`), though a leading-dot file gives `name=.eslintrc` and that is unstated.
2. **Removing `patterns`: held up.** I found no v2 capability lost by removing `patterns` itself. Both documented limits come from the single `related`, and the new loss comes from captures. Anchors reproduce (`e16`).
3. **`skipIf`: half held.** "A removed file is never missing" is stated clearly (`04:108`) and I predicted it. `skipIf: exists` with a removed file is not right as written (P1-3). Also unstated: deleting a test in one PR and editing the source in the next still skips, so the protection is same-change only.
4. **Instance names: broken once.** The rule 5 argument is sound for variable values; the break is the `(root)` literal (P2-1). `..` breaks "one related path per instance" but not name uniqueness (P2-2).
5. **Capability accounting: more of both.** One further loss, one behaviour change, three unclaimed gains (P2-4).
6. **Consistency: mostly fixed, not fully.** `02`, `03`, `04`, `06` and `07` agree with each other and with the decision log (I checked D17, D26, D30–D33, D35 and D36). `05` and `01` still carry superseded material, and two line citations are wrong (P2-3).
7. **Conformance suite: the count is right, the coverage is not.** 3 × 6 × 6 = 108, and the six related-path rows match `02` §4. The suite would catch regressions in D5, D18, D23, D24, D30 and D32. It would not catch D27 (P1-5), the "most fundamental cause" message ordering, flags on pass or fail for non-`none` rules, or `skipIf: exists` in plain-path mode (only `missing` is pinned at `07:89-91`).
8. **Wording.**
   - `skipIf`, `flags`, `why`: held up.
   - `exists` / `missing`: overloaded between `expect` and `skipIf` (P1-3).
   - `untouched`: misleading for deleted files (P2-5).
   - `changeType`: sits beside `related` and `expect`, so "expect: touched, changeType: added" reads as "expect the related file to be added" (*matter of taste*; something like `sourceChange`).
   - `expect: none` with a `related`: reads as "expect no such file" (*matter of taste*).
   - `touched`: fine once the reader knows deletion is excluded, and `04:84` says so.

## Strengths

- The three-state model and `skipIf: missing` close a real, measured hole with data Kyn already has.
- Keying instances on the variables `related` uses fixes OS13 and makes OS11 impossible.
- Every claim about current behaviour is backed by a script that reproduces.
- The limits section declines to claim completeness, and the plain-path degradation is stated openly.
- `expect` stays required, and rule 9 turns a silent mistake into a loud one.

## What I verified and how

- **Experiments:** built `/tmp/kyn-r4` and ran all 15 scripts. I compared the output with `01` for OS1 (all 24 rows), OS3, OS8, OS9, OS11, OS12, OS13, OS14, OS15, and with the `05` matrix; all match. OS7 reproduces as non-deterministic with different counts, as expected.
- **Simulator:** ran the 21 shipped scenarios, read `sim/main.go` in full, and ran 30 scenarios of my own from `$TMPDIR/kyn-r4-review/mine.yaml` and `m2.yaml`.
- **Source:** read `internal/config`, `internal/family`, `internal/matcher`, `internal/changes`, `internal/rules/evaluator.go` and `internal/cli/init.go`, and checked cited lines in `engine.go`, `sarif.go` and `docs/spec.md`.
- **Inventory:** recounted the cutover numbers; goldens (8 and 13) and inline configs (16, 2, 3) are correct.
- **Not verified:** the review 2 and review 3 dispositions. Review 2 was a PR comment and there is no review 3 record in `docs/reviews/`, so I could only check the decision log against the documents.

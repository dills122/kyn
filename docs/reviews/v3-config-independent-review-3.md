# Independent review 3 of the v3 configuration design

| | |
| --- | --- |
| Source | a fresh-context subagent given `docs/reviews/v3-config-review-3-bootstrap.md` and no prior context, 2026-09-19 |
| Reviewed | `8efbfb2`, branch `design/v3-config` |
| Dispositions | `docs/design/v3/00-workplan.md`, section Reviews |

Verbatim as delivered. File and line citations, including the absolute worktree paths, refer to that commit and to a worktree that no longer exists.

---

# Independent review: v3 configuration design set

Review instance: 3 of 3

Reviewed target: `docs/design/v3/` at `8efbfb2`, branch `design/v3-config`

Verdict: **Ready to advance into a formal specification and implementation plan, conditional on closing the five P1 findings first.** The measured baseline is real — all thirteen experiment scripts reproduce exactly what the prose claims — and the core model (two-field expectation grid, tri-state gate, shape-plus-footprint instance key) is coherent and well argued. The blockers are not in the model. They are in the two documents that carry the model into execution: `07-conformance.md` was not revised after independent review 2 and now specifies a suite that would miss the regressions it exists to catch, and `06-cutover.md` was not revised after D17 and its removal inventory misses most of the work. Both gates are currently marked closed.

---

## Findings

No P0 findings.

### P1 — The 75-cell reachability argument is wrong, and the cell it excludes is the one that justifies a grammar feature

**Wrong**, not unspecified.

`07-conformance.md:90-94` reduces the layer-1 matrix by asserting:

> - a `vanished` or `absent` path cannot be in the change set, since deleted paths never enter it and a rename records only the destination
> - so `in change set` is free only for `present`

The justification given covers `vanished` only. It is silently extended to `absent`, where it is false. An `absent` path — not in the working tree, not deleted or renamed in this change — is in the change set whenever the selected diff and the working tree disagree. This design set says so itself, twice, and measured it: `01-observed-semantics.md:45` records `changed / exists=N / in-change-set=Y / pass`, and `02-semantics.md:276-281` uses exactly that cell to argue that `exists-and-changed` is not redundant with `changed`. I reproduced it with `e3-truthtable.sh`.

Failure scenario: CI runs `kyn check --base main --head origin/feature` from a checkout that does not contain a newly added test file. A regression that makes `expect: [exists, in-change-set]` behave identically to `expect: in-change-set` passes the whole conformance suite, because the only cell that distinguishes them is excluded from the matrix by construction. The grammar keeps a compound assertion that no test can tell apart from its simple form.

Required fix: correct the reachable-state enumeration to six rows — `present` × {in-change-set Y, N}, `vanished`(deleted), `vanished`(renamed), `absent` × {in-change-set Y, N} — giving 3 × 5 × 6 = 90 cells, and name the `absent`/in-change-set row as a required assertion citing OS1. Also state explicitly how `absent` + in-change-set is generated in a fixture, since it needs the diff and the tree to disagree.

Evidence:
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/07-conformance.md:88-94`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/02-semantics.md:276-281`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/01-observed-semantics.md:45`

---

### P1 — The conformance suite was never updated for review 2's two substantive decisions

**Wrong** (stale contradictions) and **unspecified** (missing coverage).

`git show --stat 8efbfb2` confirms that the commit settling informational rules touched `00`, `02` and `04` and did not touch `07`. `ab29520` touched `07` but left earlier text intact. Three concrete consequences:

1. `07-conformance.md:80` lists the `expect` dimension as five values. `expect: none` is absent. D26 and validation rule 15 are therefore untested: an emit-only rule's `status`, its `flags` array, the `severity: info` coercion and the "at least one of `message` or `emit`" rejection have no cell anywhere in the matrix. This is the capability that review 2 found missing and the design set restored — it is the single most recently added semantic and it has zero conformance coverage.
2. `07-conformance.md:106` still lists the input-mode dimension as "`--files-from` 2-column (D9)". D9 was superseded by D24, which accepts `git diff --name-status -M` verbatim precisely because two columns cannot encode a rename. As specified, the suite tests a format the design no longer has and never tests the three-field rename form — which is the only way `relatedState = vanished` is reachable in explicit mode, and therefore the only way the OS8 half of the D5 fix is exercised outside git mode.
3. `07-conformance.md:125-127` says "Validation rules 7 and 10 — unused pattern, self-match — are new and unproven". Rules 12–15 (single ID namespace, uniform footprint, `stripSuffixes` requires `{base}`, `expect: none` constraints) are newer and equally unproven. Rule 13 and rule 14 are the enforcement arms of D22; without them footprint keying is unsound, and the corpus sentence does not call them out.

Failure scenario: an implementer builds G5 to the document. The suite goes green. `expect: none` silently defaults `severity` to `error`, reproducing the `errors: 1 / failed: 0 / ok: true` pattern that D26 exists to prevent — the OS12 family a fourth time — and nothing fails.

Required fix: re-derive `07` against the current decision table. Add `none` to the layer-1 `expect` dimension with its own assertions (status, flags, counter attribution), replace the D9 row with D24's three input forms including a rename line, and extend the "new and unproven" corpus note to rules 12–15.

Evidence:
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/07-conformance.md:80`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/07-conformance.md:106`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/07-conformance.md:125-127`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/00-workplan.md:248-251`

---

### P1 — Cross-kin applicability is a third capability loss, and the evidence offered for completeness does not cover it

**Wrong** — a claimed-complete list is incomplete for the third time.

`04-grammar.md:154-180` and `06-cutover.md:17-27` both assert that v3 loses exactly one thing: atomic multi-path rules. The evidence offered is "every kin clause across the four presets, `docs/site/recipes/*`, and `docs/site/config.md` names exactly one kin."

That measurement is about clause *arity*. It says nothing about clause *agreement*. V2 permits a rule whose `if` names one kin and whose `assert` names a different one:

```yaml
if:     { kinExists: [figma] }
assert: { kinChanged: [story] }
```

Every clause names exactly one kin, so the measurement passes, and the rule is unexpressible in v3: `when` is defined over the tri-state of *the rule's own* `related` path (`02-semantics.md:56-70`, `04-grammar.md:98-99`). There is no way to gate rule B on the state of rule A's path, and `emit` produces flags that nothing consumes — `Rule.EmitFlags()` (`internal/config/config.go:86-91`) is a v1 alias for `actions.emit`, not an assertion, so flags cannot bridge two rules either.

This was raised by independent review 1 ("Cross-kin policies such as 'if metadata exists, include documentation' cannot fit a single `related` field", `docs/reviews/v3-config-independent-review-1.md:94-96`). IR3 was then marked "largely dissolved by the scope change", and the cross-kin half of it disappeared with it — but the scope change removed the *migration* obligation, not the *capability* question, which `06-cutover.md:17-19` explicitly says survives.

Failure scenario: a user follows `docs/site/config.md`, which teaches `kinExists` as the applicability idiom and shows a family with two kin (`story`, `figma`). They write "if the component has a Figma sidecar, require the story to be updated." This loads today. After the cutover there is no v3 form for it and no documented workaround, and the release notes say the only gap is atomic multi-path rules.

I confirmed no shipped config uses the shape: the four presets, three recipes, `docs/site/config.md`, `docs/site/concepts.md` and `testdata/angular/kyn.config.yaml` all gate and assert on the same kin. So the *conclusion* (accept the loss) is probably right. The *reasoning* is not — the stated measurement does not support it, and the list is again presented as exhaustive after being wrong twice.

Required fix: record cross-kin applicability as a second accepted capability limit in `04-grammar.md#known-capability-limit` and in the workplan register, with its own measurement ("no config in the repository gates on a different kin than it asserts on" — which I verified holds). State whether a future `when.related` / gate-on-another-rule form is reserved or rejected. Separately, stop asserting exhaustiveness: say "the limits found by this review are X and Y" rather than "exactly one place."

Evidence:
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/04-grammar.md:156-176`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/06-cutover.md:21-27`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/spec.md:452-465`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/site/config.md` (kin clause table; `story` + `figma` example)

---

### P1 — `check` has no representation for a skipped result, and v3 adds a gate whose whole purpose is to produce them

**Unspecified**, with one **wrong** consequence in `07`.

`RuleResult.Status` is `pass | fail | info` (`internal/rules/result.go:13-17`). `Summary` has no `Skipped` field (`internal/rules/result.go:31-40`). I verified that a skipped rule vanishes entirely from `check` output:

```
$ kyn check --files src/a.go --format json    # rule gated off by if.kinExists
{ "ok": true, "passed": 0, "failed": 0, ..., "results": [] }
```

Text output is the same: no `Skipped:` line. Only `explain` models skips (`ExplainStatusSkipped`, `ExplainSummary.Skipped`).

The design set never addresses this. `02-semantics.md:99-101` says `explain` can now name which state produced a skip. `03-instance-and-identity.md:365-384` specifies the headline over `OK` and `Failed`. Nothing specifies what `check` does with a skipped result under the new tri-state — and v3 makes skips *more* significant, not less: `when: related-absent` is a gate whose non-firing case is a skip, and the entire OS3/OS5 thesis is that concealed skips are the defect class this redesign exists to fix. D6 makes the skip visible *in the config*; nothing makes it visible *in the report*.

Then `07-conformance.md:45` lists the observables as "result counts | `passed`, `failed`, `infos`, `skipped`, `warnings`, `errors`" — asserting a `skipped` counter on an output that has none. An implementer would either add it (an unspecified report-schema change, which the repository steering guards) or drop the assertion.

Failure scenario: a user writes `when: related-existed` on a rule whose related path never existed. The rule silently does nothing, `check` prints `PASS` with `Rules failed: 0`, and the only way to discover it is to re-run `explain`. This is OS5 with a new spelling — the phantom-instance skip that `e7-selfmatch.sh` reproduces, still invisible in `check`.

Required fix: decide and record whether v3's `check` reports skipped results and a `Skipped` count in text and JSON (and what `--show-passes` does with them), then make `07`'s observable list match. If the answer is "no", say so explicitly and remove `skipped` from the observables.

Evidence:
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/internal/rules/result.go:13-17`, `:31-40`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/internal/rules/explain.go:14-20`, `:49-59`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/07-conformance.md:45`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/03-instance-and-identity.md:365-384`

---

### P1 — The G4 removal inventory misses most of the cutover, including an existing e2e harness that overlaps G5

**Unspecified**, presented as complete. G4 is marked CLOSED.

`06-cutover.md:29-39` lists seven items. Measured against the repository, it omits:

| Omitted | What it is |
| --- | --- |
| `e2e/projects/{angular-storybook,frontend-recipe,go-api,terraform-iac}/kyn.config.yaml` | four v1/v2 configs driving `TestE2EProjects` |
| `e2e/projects/*/scenarios.json` | scenario definitions across text/json/dry-run and `--fail-on` |
| `e2e/projects/*/testdata/*.golden` | eight committed golden outputs |
| `testdata/angular/kyn.config.yaml` | a v1 config consumed by CLI tests |
| `internal/cli/execute_test.go` (16 embedded configs), `internal/cli/output_stream_test.go` (2), `internal/config/load_test.go` (3) | 21 inline v1/v2 fixtures |
| `docs/spec.md` | the canonical specification of families, kin, `when`/`require` |
| `README.md`, `docs/ci.md` | user-facing v2 examples |

The e2e omission is the serious one. `e2e/e2e_test.go:82` already walks every fixture under `e2e/projects`, runs the real binary and asserts rendered output plus exit code across output modes — which is precisely the harness `07-conformance.md:129-136` proposes to build, citing `e2e/workflows_test.go:95` (a different, single-scenario test) as "the right *shape* to copy". The design set does not say whether G5 replaces this harness, extends it, or duplicates it.

Every one of those goldens is invalidated by decisions already taken. `e2e/projects/terraform-iac/testdata/module-changed-readme-not-updated-warns-only.golden` currently reads:

```
kyn check

PASS

Rules failed: 1
Warnings: 1
...
Family: terraform-module
Instance: terraform/network/main
```

That single file changes in three independent ways: `PASS` → `NON-BLOCKING` (D18), `Family:` → `Shape:` (D16), and `Instance: terraform/network/main` → `terraform/network` (D17 footprint keying, since `{dir}/README.md` has footprint `dir`). The OS12 defect is currently enshrined as a committed external-contract fixture, and the cutover document does not mention it.

Failure scenario: an implementation plan built from `06`'s list is scoped at seven items and lands with `go test ./e2e` red, eight goldens to re-record under a steering rule that treats golden diffs as deliberate review events, and no prior decision about whether the existing project harness survives.

Required fix: complete the inventory by grepping for v1/v2 artefacts rather than enumerating from memory, decide the fate of `e2e/projects/*` explicitly (it is the natural home for G5's corpus), and record which goldens change and why. Note the scope boundary: `docs/site/*` may belong to PR #49, but `docs/spec.md`, `README.md`, `docs/ci.md` and all Go fixtures do not.

Evidence:
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/06-cutover.md:29-39`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/e2e/e2e_test.go:82-100`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/e2e/projects/terraform-iac/testdata/module-changed-readme-not-updated-warns-only.golden:1-10`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/07-conformance.md:129-141`

---

### P2 — OS9 is cited nine times and does not exist

**Wrong** (a dangling citation), low impact because the underlying evidence is real.

`01-observed-semantics.md` contains OS1–OS8 and OS10–OS13. There is no OS9 section. OS9 is nevertheless cited as the basis of D11, of D16's ordering claim, of `03`'s §1 contract, of `03:199` and `03:253`, of `04:67`, and of `07:47`. It is the measured cardinality-and-ordering contract that the entire instance model is built to preserve.

The evidence does exist — `e10-shared-instance.sh` is headed "OS9: cardinality contract" and I ran it; it produces exactly the four-result table at `03-instance-and-identity.md:23-28`. The observation was simply never written into the document whose stated purpose is to hold them ("each observation cites the one that produced it", `01:6-8`). Separately, `01:349` cites "OS9 matrix" to mean the compatibility matrix, which is `05`/`e11` — a different thing entirely.

Required fix: write OS9 into `01` from `e10-shared-instance.sh` output, and correct the `01:349` citation to `05`.

Evidence:
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/01-observed-semantics.md:25-372` (heading list), `:349`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/experiments/e10-shared-instance.sh:2`

---

### P2 — `instanceName` rendering is underspecified for empty components, and diverges from v2 where the document says it does not

**Unspecified**, with one **wrong** claim.

D22 defines `instanceName` as "the footprint values joined with `/`" (`03:169-171`) and specifies exactly one rendering edge: an extensionless file under an `{ext}` signature yields a trailing `/`. Two more edges are unhandled, both involving `dir`, which is `""` for a repository-root file (`internal/family/resolver.go:142-145`):

1. Footprint `dir, base`, source `a.go` at the root → `"" + "/" + "a"` = `/a`. V2 renders `a`; I verified this. `03:137` claims this footprint is "unchanged from v2". For root-level sources it is not.
2. Footprint `dir` alone, source `a.go` at the root → `instanceName` is the empty string. Both `text.go:46` and `renderRDJSONMessage` suppress the field when empty, so the instance identity disappears from the report entirely. This is reachable from the exact policy the design highlights: "any change under X requires a README", applied to a repo with root-level sources.

Related: `primaryLocationPath` (`internal/report/rdjson.go:122`) falls back to `FamilyName` as a *file path* for the RDJSON/Checkstyle location. Under D22 that value can be `(all)` — a marker the design says "cannot be mistaken for a path" — or empty. The fallback appears unreachable today (`ChangedFiles` is always non-empty for a real instance), so this is latent rather than live, but `03`'s §6 projection table does not mention it and an implementer renaming fields would carry it forward unexamined.

Required fix: specify the join over possibly-empty components (an explicit placeholder, or omit-and-collapse, or forbid an empty `dir` component), correct the "unchanged from v2" claim to name the root-level exception, and add `primaryLocationPath` to the §6 projection table with a decision.

Evidence:
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/03-instance-and-identity.md:137`, `:164-176`, `:204-206`, `:322-325`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/internal/family/resolver.go:53-57`, `:142-145`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/internal/report/rdjson.go:116-123`

---

### P2 — `e9-gate-prototype.sh` still tests the boolean D23 replaced, and `related-absent` has no experimental backing

**Unspecified** evidence, over-claimed heading.

`02-semantics.md:103` heads its validation "Validated against real git output" and presents a six-row table whose columns include the tri-state `state` and both gates. The script produces neither. Its header still says `Claim under test: "existedAtBase"` and "simulates the proposed gate over **four** scenarios" (there are five), and its output column is `EXISTED-BASE` — the misnamed boolean R2-3 eliminated. I ran it; it emits exactly the pre-D23 table.

The substance survives — `existsNow` and `vanished` are both columns, so the tri-state is derivable — but the `related-absent` gate is simulated nowhere at all. D23 is the newest semantic change in the set and half of it has no evidence behind it.

Required fix: update `e9` to emit the tri-state and both gates, add a `related-absent` column, fix the "four scenarios" comment, and re-paste its output into `02` §4 so the table is script output rather than hand derivation.

Evidence:
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/experiments/e9-gate-prototype.sh:4-7`, `:55-57`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/02-semantics.md:103-121`

---

### P2 — D26's measured claim has no experiment script

**Unspecified** evidence. The claim itself is true.

D26, `02-semantics.md:225-231` and validation rule 15 all rest on: "an emit-only rule with `severity: error` reports `errors: 1` alongside `failed: 0` and `ok: true`." No script is cited. `e13-severity-failon.sh` measures the warn/`--fail-on` headline (OS12), not emit-only rules; it contains no `actions.emit`.

I reproduced the claim by hand and it holds exactly, plus one detail the design did not record: the result carries `"status": "info"` while `"infos": 0`, so an emit-only rule is miscounted in two places, not one.

This matters because the set's headline methodological claim is that every statement about current behavior was produced by running the binary with a cited script (`01:5-8`). This is the one place that is not true.

Required fix: add an `e15` script covering emit-only rules at each severity, record it as an OS-numbered observation in `01` including the `infos: 0` detail, and cite it from D26.

Evidence:
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/00-workplan.md:251`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/02-semantics.md:225-231`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/experiments/e13-severity-failon.sh:1-34`

---

### P2 — Input-mode degradation is specified for one gate and not the other

**Unspecified.**

`02-semantics.md:174-178` says that in `--files` mode, without vanish information, "the gate degrades to `existsNow`, which is exactly today's behavior." That is accurate for `related-existed`. It is not analysed for `related-absent`, which is new and has no "today's behavior" to degrade to: in `--files` mode a deleted related path is indistinguishable from one that never existed, so `related-absent` **fires** on it. A rule like `when: related-absent / expect: none / emit: [needsDocs]` would emit spuriously whenever the related file was deleted — the mirror image of the OS3 hazard, introduced by the fix for OS3.

Related and also unspecified: D24 makes real statuses available through `--files-from` for the first time. `on:` currently cannot match `added`/`deleted` in explicit mode because `manual.go` forces `StatusModified`. Whether D24's statuses now feed `on:` is not stated anywhere, though `07:109` crosses `on` filter with input mode in layer 2.

Required fix: extend `02` §5 with a per-gate degradation table covering all three `when` values in each of the three input modes, and state whether `--files-from` statuses feed `on:`.

Evidence:
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/02-semantics.md:146-182`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/internal/changes/manual.go:24`, `:70`

---

### P2 — The `iac` deduplication question is simultaneously decided, deferred, and untracked

**Wrong** — three documents disagree about the state of a stated G5 blocker.

- `03-instance-and-identity.md:111-143` decides it: footprint keying makes `{dir}/README.md` produce one instance per directory. D17 is recorded as a maintainer decision.
- `06-cutover.md:94-100` says the same question is "Deferred, not decided — it is a report-shaping question, tracked as an open item in the workplan", and warns it "must be settled before the G5 goldens are recorded".
- `07-conformance.md:140-141` makes recording goldens conditional on it being settled.
- `00-workplan.md:277-286` lists two open questions, neither of which is this one.

`06` was last touched by `cde14c3`, before D17 landed, which explains the drift — but the consequence is that G5's exit path has an explicit precondition that the workplan says is closed and that no document tracks.

Required fix: delete the deferral from `06` and state that D17 resolves it, or, if D17 does not fully cover it (it changes instance *cardinality*, which is not identical to deduplicating *results* with the same resolved path), add the residual question to the workplan's open list.

Evidence:
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/06-cutover.md:83-100`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/03-instance-and-identity.md:129-141`
- `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/00-workplan.md:277-286`

---

### P2 — Superseded decisions are still cited as delivered

**Wrong**, cosmetic individually, corrosive in aggregate for a document about to become a specification.

- `00-workplan.md:175-176` lists G2's deliverables as including "Report identity resolved by keeping the wire format unchanged (**D10**)" and "`explain` clause vocabulary follows the config version (**D13**)". D10 is struck through at `:235` as superseded by D16, and `:28` declares D13 moot. The gate summary says the opposite of the decision table.
- `03-instance-and-identity.md:322-325` gives `instanceName`'s value as `{dir}/{base}` — the pre-D17 definition, in the same document that redefines it forty lines earlier.
- `05-compatibility-matrix.md:73` says rows 8 and 13 are dropped by "the G4 migration … under D2"; there is no migration and D2 is moot.
- `06-cutover.md:119-122` lists as unsettled whether `docs/decisions.md` and `docs/migration-v1-to-v2.md` are retired or annotated; D19 decides it.
- `docs/design/v3/README.md:15` advertises "decisions D1–D14" (there are 26) and describes `03` as "report identity preserved" (it is renamed).
- `04-grammar.md:144` has a stray blank line that splits the validation-rules table, so rule 15 renders as a separate headerless table.

Required fix: one consistency pass with the decision table as the authority.

---

### P2 — The D18 headline has no dimension that can exercise it

**Unspecified.**

D18 introduces three headlines as a function of `(OK, Failed)`, which is driven by severity crossed with `--fail-on`. `07`'s layer 2 varies severity (`:110`) but never `--fail-on`, and layer 1 asserts only on `explain --format json` plus the exit code — `explain` has no headline. So `NON-BLOCKING` is reachable in the suite only if a sampled cell happens to combine a `warn`-severity failure with default `--fail-on`. Given that the existing terraform golden encodes the *old* behavior, a silent revert is a live risk.

Required fix: add `--fail-on` to layer 2's dimensions and assert all three headlines explicitly against their `(OK, Failed)` conditions.

Evidence: `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1/docs/design/v3/07-conformance.md:100-111`; `03-instance-and-identity.md:374-380`

---

## Areas where I was asked to be sceptical and found nothing serious

The bootstrap named six. Three of them I could not break, and I want to say so plainly rather than manufacture a finding.

**Footprint-derived instance keys (D17, D22) — sound.** I attacked the injectivity claim at `03:197` directly. The rendering joins footprint values with `/` in the fixed order `dir, base, name, ext, file`; `base`, `name` and `ext` derive from `path.Base` and contain no separator, while `dir` and `file` both can. The document's justification ("inconsistent tuples are unreachable") is terse but it is the right argument, and it holds: because `file` is the full normalized path and `dir = path.Dir(file)`, any rendering containing both reduces to `d + "/" + d + "/" + basename`, and equality forces `|d1| = |d2|` hence `d1 = d2`. For every other signature only the leading `dir` component can contain `/`, and the component count is fixed by the signature, so right-to-left parsing is unambiguous — including when components are empty. The uniform-footprint restriction does the work claimed of it, and I verified the "no existing config is broken" claim independently across all four presets, all three recipes, `docs/site/config.md`, `docs/site/concepts.md` and `testdata/angular/kyn.config.yaml`: every multi-rule shape uses `dir, base`. The only defect I found in this area is the empty-component rendering edge above, which is a specification gap, not a soundness one. The claim that footprint keying makes the template-agreement guard unreachable (`03:228-233`) also checks out — every variable a template can use is in the key by construction.

**Report identity (D16, D21) — essentially complete.** I enumerated every reference to family identity in `internal/report` and `internal/rules` and checked each against D16's projection table. `json`, `text`, `sarif` (both rule-level `properties.familyId` and result-level), `rdjson` message prose, `checkstyle` via the shared renderer, `--dry-run-resolve` (`familyId`, `name`, and the `[fam] name` text prefix), `explain` — all accounted for. The `kin` map projection to rule IDs (`03:345-363`) is a clean answer to review 1's sharpest point. Sort order really is structurally unchanged: `evaluator.go:83-91` and `text.go:94-99` both sort on the two renamed fields plus `ruleId`, so the rename cannot perturb ordering. The only gap is `primaryLocationPath`'s latent fallback, noted above as P2, and the `properties.familyName` rename is implied rather than stated.

**Self-assessment risk — I did not find findings accepted too readily.** Review 2's six points are all settled with reasoning that goes beyond agreement: R2-3 produced an admission that the predicate was misnamed rather than merely vague, R2-4 produced an explicit falsification of the set's own completeness claim, and D22 records "a deliberate re-check" and rejects two alternatives on stated grounds. D26's choice to keep `expect` required and spell informational rules `expect: none` is the harder option, chosen against the lower-typing alternative for a principled reason. If anything the risk runs the other way: the *documents* were updated where the author was working (`00`, `02`, `03`, `04`) and not where they were not (`06`, `07`, the experiment scripts), which is how three of my five P1s arose. That is a process failure, not a judgement failure.

The other three sceptical areas produced the findings above: the tri-state gate is incomplete for `related-absent` in explicit mode (P2), capability loss number three exists (P1), and the conformance suite would not catch several of the regressions it exists to catch (two P1s).

---

## Notable strengths

- **The measurement discipline is real and it survives verification.** I ran all thirteen scripts. Every table in `01-observed-semantics.md` reproduced exactly: the 24-row truth table, the deletion and rename blindness traces, the `stripSuffixes` cardinality split, the byte-identical inert-groups comparison, the self-match phantom instance, the `api` preset's exit 2, the `PASS`-above-`Rules failed: 1` headline, the three-template footprint table, the four-result shared-instance contract, and all twenty compatibility-matrix rows with matching error strings. The determinism defect reproduced with a different distribution each time, which is the point. `go test ./...` is green. This is a materially higher evidentiary standard than the proposal review 1 saw, and it is the main reason the verdict is conditional-yes rather than no.
- **The experiment harness itself is well built.** `lib.sh`'s two safety rules, and the explicit written justification for *not* using `set -e`, are the kind of thing that normally gets discovered twice. D25 records why.
- **Several findings in this set are ones nobody asked for.** OS11 (a shipped preset that exits 2 on an ordinary change set), OS12, OS13 and OS7 were found by probing, not by reviewing. OS3 being raised to P0 on the grounds that "a policy tool whose default rule is disabled by deleting the file it protects has a correctness problem, not a vocabulary problem" is the single best judgement call in the set.
- **D6's two-field grammar is a genuine simplification**, and the argument that it converts multiplicative enum growth into additive growth is correct. Making `when` default to `always` rather than the lenient gate inverts the current failure mode in the right direction.
- **D17 earns its keep three times over**: it fixes OS13's over-partition, makes the kin-agreement error class unreachable, and dissolves OS11 structurally. Deriving the partition from the template rather than the source is the correct insight, and `03:223-243` argues it honestly, including conceding that the preset fix becomes "belt and braces rather than load-bearing".
- **Decisions are reversible and their reversals are recorded.** D10 struck and superseded, D9 superseded by D24, D11 refined by D17 with an explicit note that "instances belong to a shape" is no longer literally true. Very few design sets admit that in writing.
- **The rejected alternatives are stated with their costs.** `03:178-191` names what the uniform-footprint restriction costs (two patterns with duplicated globs), why the uglier-but-unrestricted rendering was rejected, and why accepting the collision was rejected. That is what makes the decision reviewable.

---

## Verification performed

- Confirmed worktree at `/Users/dsteele/go/src/kyn/.claude/worktrees/e2e-testing-test-projects-139bd1`, branch `design/v3-config`, clean. HEAD is `8efbfb2`, one commit ahead of the bootstrap's stated freeze at `ab29520`; the extra commit is the D26 settlement plus the bootstrap itself. No files modified by this review.
- `go build -o /tmp/kyn-r3 ./cmd/kyn`; `go test ./...` — all nine packages pass.
- Ran all thirteen experiment scripts with `KYN=/tmp/kyn-r3` and a scratch `WORKROOT`. All reproduce their documented tables. No script touched the repository.
- Hand-built four additional fixtures to test claims with no script behind them: emit-only rule counters (D26 — reproduces, plus an unrecorded `infos: 0`), duplicate YAML mapping-key rejection (D3 — reproduces with the documented message form), root-level `instanceName` rendering, and skipped-result representation in `check --format json`.
- Read `AGENTS.md`, both `.codex/steering/` documents, `docs/spec.md`, `docs/decisions.md`, and inspected `internal/config`, `internal/family`, `internal/rules`, `internal/report`, `internal/changes`, `internal/cli/init.go`, `e2e/`, and `testdata/`.
- Used `git log` and `git show --stat` only to establish which design documents each decision commit touched — the basis for the staleness findings, not for reconstructing intent.

---

## Open questions and residual risks

1. Does `check` report skipped results in v3? Nothing in the set answers this, and it determines whether the OS5 class of silent skip is actually closed.
2. Is cross-kin applicability rejected permanently or reserved? `expect` is already list-shaped for a future list-shaped `related`; a future gate-on-another-path has no reserved space.
3. Does G5 replace, extend, or duplicate `e2e/projects`? The answer materially changes the implementation plan's size.
4. `03:405-408` leaves the `explain` JSON key rename (`if`/`assert` → `when`/`expect`) as a conditional sentence with no decision ID, in a document that otherwise freezes wire contracts.
5. OS7 remains the one G0 item the workplan calls "still urgent" and it is not yet written. `07:150-153` correctly makes the determinism guard the first thing built. Open question 6 in the workplan — whether G0 ships before approval — is still unanswered and is the only sequencing question that actually blocks G5.
6. `expect: not-in-change-set` still passes for a file deleted in the diff. The set discloses this at `02:306-309` and answers it with "compose `expect: [exists, not-in-change-set]`", which is correct but appears only in a deferral note. It should be the documented form for "this file must not change", or users will write the single-value version and inherit OS3.

---

## Recommended next actions

1. Re-derive `07-conformance.md` against the current decision table: fix the reachable-cell count to 90, add `expect: none`, replace D9 with D24's three input forms including a rename, add `--fail-on` to layer 2, and reconcile the `skipped` observable with whatever `check` is specified to do.
2. Complete `06-cutover.md`'s inventory from a repository-wide grep, and decide the fate of `e2e/projects/*` and its eight goldens.
3. Record cross-kin applicability as capability limit number two, with its own measurement, and stop claiming the list is exhaustive.
4. Specify `check`'s treatment of skipped results.
5. Run one consistency pass over `00`, `03` §6, `05`, `06` and the README against the decision table, and close the `iac` deduplication question one way or the other.
6. Add the missing `e15` emit-only script and write OS9 into `01`; update `e9` for the tri-state.

Items 1–4 are the conditions on the verdict. Items 5–6 are hygiene that should land in the same pass, because the specification that follows will be read as authoritative and three of my five P1s exist only because a decision did not propagate to a downstream document.

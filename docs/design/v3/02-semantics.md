# G1 — Frozen semantics

Status: proposed. Closes gate G1 in [`00-workplan.md`](00-workplan.md).

Depends on the measured baseline in
[`01-observed-semantics.md`](01-observed-semantics.md). Every claim about
current behavior here is cited to an observation there, not to source reading.

Maintainer decision recorded 2026-09-09: **close the deletion capability gap**
(OS3). This document works out what that costs and what it buys.

## 1. The vanishing-path problem

A related path can stop being where the policy expects it in three ways. Today
all three produce the identical outcome — `skipped`, exit `0`:

| # | How the path vanishes | Git reports | Kyn keeps | Should the rule fire? |
| --- | --- | --- | --- | --- |
| 1 | Deleted in this change | `D src/a_test.go` | nothing (OS3) | **yes** — coverage removed |
| 2 | Renamed away in this change | `R100 src/a_test.go src/renamed_test.go` | destination only (OS8) | **yes** — coverage moved, relationship broken |
| 3 | Never existed | nothing | nothing | no — genuinely not applicable |

Cases 1 and 2 are the policy violations the tool exists to catch. Case 3 is
real non-applicability. Conflating all three into one silent skip is the defect,
and it is reachable from the shape every `kyn init` preset ships.

The vocabulary problem (IR4) sits on top of this: `unchanged` reports
`pass (rel unchanged (src/a_test.go))` for a file deleted in the very diff under
evaluation (OS3). Renaming the words does not make the tool see the deletion.

## 2. What Kyn already has and throws away

Both signals are already on the wire. No new `git` invocation is required.

- `internal/changes/git.go:60` matches `D` and discards it, with the comment
  "Deleted files are intentionally excluded for MVP evaluation".
- `internal/changes/git.go:45` handles `R` by recording `fields[2]`, the
  destination. `fields[1]`, the source path, is discarded.
- `changes.StatusDeleted` already exists in the type system
  (`internal/changes/changes.go:16`) and `mapStatus` already maps `"D"` to it.
  Nothing ever produces it.

So closing the gap is a matter of retaining data already parsed, which
`docs/decisions.md` explicitly anticipated: "Deleted paths may be retained as
metadata for future enhancements".

## 3. Proposed model — a tri-state, not a boolean

An earlier draft derived a boolean `existedAtBase = existsNow OR
vanishedInChange`. Independent review 2 was right that this is not literal
existence at base: a file **added** in this change has `existsNow = Y` and never
existed at base. The predicate was misnamed, and only one of the two gates was
ever tabulated against it.

Two facts per related path, both locally computable:

| Fact | Source | Available in `--files` mode |
| --- | --- | --- |
| `existsNow` | `os.Stat` against `--cwd` | yes |
| `vanishedInChange` | appeared as `D`, or as the **source** of an `R` | git mode only (see §5) |

They are mutually exclusive in practice — a path present in the working tree was
not removed by this change — so they collapse into one tri-state:

| `relatedState` | Condition |
| --- | --- |
| `present` | exists in the working tree |
| `vanished` | not in the tree, and appeared as `D` or a rename source in this change |
| `absent` | neither |

Change-set membership stays a separate, orthogonal fact. Deleted paths are
**not** added to the change set — doing so would make `expect: in-change-set`
pass for a file the change deleted, the opposite of the intent.

## 4. The gate, defined over all three states

| `when` | fires on |
| --- | --- |
| `always` | `present`, `vanished`, `absent` |
| `related-existed` | `present`, `vanished` |
| `related-absent` | `absent` |

Every case review 2 asked for, with the previous behavior for comparison:

| Scenario | `relatedState` | `related-existed` | `related-absent` | old gate |
| --- | --- | --- | --- | --- |
| Related present, untouched | `present` | fires | skips | fires |
| Related present, updated | `present` | fires | skips | fires |
| Related **added** in this change | `present` | fires | skips | fires |
| Related **deleted** (OS3) | `vanished` | **fires** | skips | skipped |
| Related **renamed away** (OS8) | `vanished` | **fires** | skips | skipped |
| Related never existed | `absent` | skips | **fires** | skipped |

The tri-state fixes OS3 and OS8 together, with no new expectation vocabulary, no
new git calls, and no change to what change-set membership means. `absent` keeps
skipping under `related-existed`, which is correct, and is the only state that
activates `related-absent`.

It also makes the OS5 self-match hazard legible: the phantom instance's related
path is `absent`, so it still skips — but `explain` can now name which state
produced the skip rather than reporting a bare `skipped`.

### Validated against real git output

[`e9-gate-prototype.sh`](experiments/e9-gate-prototype.sh) simulates the
`related-existed` gate over five scenarios, deriving `vanishedInChange` from
`git diff --name-status -M <base>...<head>` — the exact command Kyn already runs
(`internal/changes/git.go:10`). No additional git invocation, no revision
walking, no object reads.

| Scenario | `existsNow` | `vanished` | `inChangeSet` | state | old gate | new gate |
| --- | --- | --- | --- | --- | --- | --- |
| related untouched | Y | N | N | `present` | FAIL | FAIL |
| related updated | Y | N | Y | `present` | pass | pass |
| related **deleted** | N | Y | N | `vanished` | skipped | **FAIL** |
| related **renamed away** | N | Y | N | `vanished` | skipped | **FAIL** |
| related never existed | N | N | N | `absent` | skipped | skipped |

Two of five rows change, and they are exactly the two the fix targets. The
already-correct rows — including the genuinely-not-applicable case — are
untouched, so the change adds no false positives.

The "2 of 5" blast radius is also the honest framing for the release note: a
policy behaves as before unless the related file is deleted or renamed away in
the change under evaluation.

### Cost: a behavior change, no longer a divergence

A repository that deletes a test today exits `0`. After this change it exits
`1`. That is the point.

This was originally version-gated — v1/v2 keeping the blind gate, v3 getting the
fix — to protect a shipped contract, which made it the first and only
intentional divergence between a v2 policy and its v3 equivalent. The
[scope change](00-workplan.md#scope-change-2026-09-10) removed that constraint:
there are no users, v1 and v2 are being retired rather than carried, so the
corrected gate is simply **the** gate.

Consequences of dropping the version gate:

- No divergence to special-case, so the G5 harness loses its S1 exception.
- No dry-run divergence warning to write.
- `relatedState` is computed one way, not two, which removes a branch from the
  normalized model and from every test that would have covered both sides.

## 5. Input-mode asymmetry

`--files` and `--files-from` record every supplied path as
`StatusModified` (`internal/changes/manual.go:24`,
`internal/changes/manual.go:70`). Explicit mode has no way to say "this path was
deleted", so `vanishedInChange` is unavailable there.

Recommendation:

1. **`--files` stays a plain path list.** It is the quick-iteration flag; adding
   status syntax to a comma-separated list would be unreadable.
2. **`--files-from` accepts `git diff --name-status -M` output verbatim.**
   Review 2 was right that `status<TAB>path` alone cannot encode a rename: git
   emits three fields for one, `R100<TAB>old<TAB>new`, and both paths matter —
   the destination joins the change set, the source is what makes the related
   path `vanished`.

   So the accepted grammar is git's own, not a two-column approximation of it:

   ```text
   M	src/a.go
   D	src/a_test.go
   R100	src/old.go	src/new.go
   ```

   Single-field lines keep meaning `modified`, so a plain path list still works
   and nothing that reads today stops reading. Users can pipe git straight in,
   and there is no second format to learn or to keep in sync.
3. **`--files` documents the fallback**: without vanish information the gate
   degrades to `existsNow`, which is exactly today's behavior. A policy designed
   in `--files` mode and enforced in git mode can therefore fail in CI having
   passed locally. This must be called out in the tutorial, not buried in
   reference material.

An alternative — inferring deletion in explicit mode from "supplied path that
does not exist on disk" — is rejected. It would silently reclassify typos as
deletions, and typo-tolerance is already weak (OS5).

## 6. Frozen expectation grid

OS2 established that the vocabulary is a closed 2×2 (existence × change-set
membership) with existence optionally promoted to a gate. Rather than fusing
those into compound names (`changed-if-present`, `exists-and-changed`), the two
axes stay two fields:

```yaml
when:   always | related-existed | related-absent   # applicability
expect: in-change-set | not-in-change-set | exists | missing
```

`expect` accepts a list, AND-combined, which is how the one genuinely compound
v2 form is expressed without inventing a fused name.

### Mapping

| `when` | `expect` | V2 clauses | Proposal's fused name |
| --- | --- | --- | --- |
| `always` | `in-change-set` | `assert.kinChanged` | `changed` |
| `always` | `not-in-change-set` | `assert.kinUnchanged` | `unchanged` |
| `always` | `exists` | `assert.kinExists` | `exists` |
| `always` | `missing` | `assert.kinMissing` | `missing` |
| `related-existed` | `in-change-set` | `if.kinExists` + `assert.kinChanged` | `changed-if-present` |
| `always` | `[exists, in-change-set]` | both asserts | `exists-and-changed` |

### Why two fields instead of an enum

- The author's stated risk was combinatorial enum growth. Two orthogonal fields
  give 3×4 combinations from seven words, and adding a gate or an assertion is
  additive rather than multiplicative.
- IR4's strongest sub-finding was that `changed-if-present` conceals the skip.
  With `when` as its own field, the skip is written down in the config.
- `exists-and-changed` stops being a special case that has to be memorized.

### Why `in-change-set` and not `changed`

Measured: `changed` passes for a path that does not exist but was supplied to
`--files` (OS1, row `changed / N / Y`), and `unchanged` passes for a path
deleted in the diff (OS3). Kyn observes set membership, never content. The names
must say so.

### `when` has no lenient default

`when` defaults to `always`. It is **not** defaulted to `related-existed`.

A rule with no `when` therefore demands the related path unconditionally and can
never silently skip. Users who want leniency write `when: related-existed` and
can see that they did. This inverts today's situation, where the lenient,
silenceable behavior is what the starter configs hand you and what the proposal
nominated as the default.

### `exists-and-changed` is not redundant

It differs from `changed` in exactly one cell — `existsNow=N, inChangeSet=Y`
(OS1) — reachable when the selected diff and the working tree disagree, e.g.
`--base main --head origin/feature` from a checkout lacking the added file.
Documentation must state this, or users will read the two as synonyms and pick
arbitrarily.

## 7. Generated messages (CR4)

`message` is required today (`internal/config/validate.go:117`). The proposal
makes it optional with a generated default.

Recommendation: allow the default, and constrain it.

- The generated text is a **versioned contract**, fixture-tested like the other
  report golden files.
- It contains only the rule ID and the resolved related path, both already
  present in structured fields.
- It never contains the config file path, an absolute path, or a pattern name.
  Absolute paths would leak the CI checkout root into reviewdog comments.
- The `(family instance: ...)` suffix that `renderRDJSONMessage` appends is out
  of scope here and belongs to gate G2 (IR1).

## 8. What G1 does not settle

- Instance-key derivation and `stripSuffixes` placement (IR2, CR1) — gate G2.
- The identity that replaces `familyId` / `familyName` in all seven output
  modes (IR1) — gate G2.
- Clause naming in `explain` traces under the two-field form (CR5) — gate G2.
  Note the two-field grammar makes this easier: `when` and `expect` map onto the
  existing `If` / `Assert` trace sections without inventing a vocabulary.
- Whether a `deleted` expectation is worth adding on top of the gate fix.
  Deferred: `expect: exists` already covers "must not be deleted", and D4 says
  keep the grid closed until a real case forces it open.

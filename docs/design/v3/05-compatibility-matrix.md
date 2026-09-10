# G3 — v1/v2 compatibility matrix

Status: measured. Answers IR5.

> **Superseded as a constraint.** The
> [scope change](00-workplan.md#scope-change-2026-09-10) retires v1 and v2
> instead of carrying them, so "do not break shapes that load today" no longer
> binds. This document is kept as measured evidence and for the three findings
> below, which still apply to v3.

Source: [`e11-compat-matrix.sh`](experiments/e11-compat-matrix.sh)

IR5's objection was that "continue loading v1 and v2" does not define what the
current loader actually accepts, so clean version-specific structs could reject
configurations that load today. This is the inventory. Twenty shapes were
probed against the shipped binary; eleven load, nine are rejected.

`LOADS` means exit 0 or 1 — the config parsed, validated and evaluated.
`REJECT` means exit 2.

## Matrix

| # | Shape | Verdict | Contractual? |
| --- | --- | --- | --- |
| 1 | v1, top-level `include`, `when` / `require` | LOADS | **yes** |
| 2 | v1 with `groups.source` and no `include` | REJECT | yes — v1 has no `groups` |
| 3 | v1 with `baseName.stripSuffixes` | LOADS | **yes** |
| 4 | v2 native `groups.source` | LOADS | **yes** |
| 5 | v2 top-level `include`, no `groups` | LOADS | **yes** — migration compatibility |
| 6 | v2 top-level `include` **and** `groups.source.include` | REJECT | yes — deliberate ambiguity guard |
| 7 | v2 `groups` without a `source` group | REJECT | yes |
| 8 | v2 with inert non-`source` groups | LOADS | **incidental** — see below |
| 9 | v2 legacy `when` / `require` aliases | LOADS | **yes** |
| 10 | v2 mixing `if` and `when` | REJECT | yes — deliberate guard |
| 11 | v2 legacy `require.emitFlag` | LOADS | **yes** |
| 12 | v2 `actions.emit` | LOADS | **yes** |
| 13 | v2 `rule.description` | LOADS | **incidental** — inert, OS10 |
| 14 | v2 with `message` omitted | REJECT | changes under D7 |
| 15 | v2 with `severity` omitted | REJECT | changes in v3 |
| 16 | `version: 3` | REJECT | becomes valid in v3 |
| 17 | `version` omitted | REJECT | yes — but see error quality |
| 18 | unknown top-level field | REJECT | **yes — load-bearing** |
| 19 | v2 `assert.changedAny` | REJECT | yes — routed to `if` |
| 20 | v2 `if.changedAny` naming a non-`source` group | REJECT | yes |

## Findings

### Row 18 is what makes v3 safe to add

`KnownFields(true)` (`internal/config/load.go:76`) rejects any unrecognized
top-level key, so a v1/v2 file containing `patterns:` fails today with a parse
error rather than silently ignoring it. Introducing `patterns` and a mapping
`rules` in v3 therefore cannot change the meaning of any existing file. There is
no ambiguous overlap to design around.

### Row 5 and row 9 are the real compatibility surface

These are the shapes a clean rewrite is most likely to drop, because they are
undocumented-looking leftovers rather than the canonical form:

- **Row 5** — a v2 family may still use top-level `include`/`exclude`, but only
  when it has no `groups` map at all. Row 6 shows the ambiguous combination is
  rejected on purpose, with an explanatory message.
- **Rows 9 and 11** — v1's `when`, `require` and `require.emitFlag` still load
  under `version: 2`. Row 10 shows mixing old and new spellings is rejected.

Both must appear as loader fixtures before the shared decoder is replaced.

### Rows 8 and 13 are incidental, not contractual

Both load, and both do nothing (OS6, OS10). They are accidents of a permissive
struct plus `KnownFields`, not promises. v3 need not carry them, and the G4
migration drops them under D2.

Row 8 has an odd consequence worth naming: validation *requires* every declared
group to carry a non-empty `include` (row 7's sibling check), so v2 enforces
correctness on data it never reads.

### Row 17's error message is wrong

An omitted `version` reports:

```text
invalid config: unsupported config version 0; expected version 1 or 2
```

No user wrote `0`. The zero value leaks into the message. Fix belongs in G0
alongside the other error-quality work: report a missing `version` as missing.

### Rows 14, 15 and 16 are the v3 deltas

- Row 14 — `message` becomes optional under D7, with a generated default.
- Row 15 — `severity` becomes optional in v3, defaulting to `error`. In v1/v2 it
  stays required; the empty-severity message (`has invalid severity ""`) has the
  same zero-value leak as row 17.
- Row 16 — `version: 3` becomes valid. Nothing else changes about how 1 and 2
  are handled.

## What still applies to v3

The contractual/incidental column is now historical. Three findings survive the
scope change:

1. **Row 18 makes v3 safe to add.** `KnownFields(true)` already rejects unknown
   top-level keys, so no v1/v2 file ever silently ignored a `patterns:` block.
   There is no ambiguous overlap to design around, whether or not v2 is kept.
2. **Rows 6 and 10 are design lessons worth repeating.** Both reject an
   *ambiguous* combination — top-level `include` alongside `groups.source`, and
   `if` alongside `when` — with an explanatory message rather than silently
   picking one. v3 should keep that habit: `use` alongside inline `match` is the
   direct analogue, and validation rule 1 already covers it.
3. **Rows 15 and 17 are real error-quality bugs** that carry straight into v3 if
   the code is reused: a missing `version` reports `unsupported config version 0`
   and a missing `severity` reports `has invalid severity ""`. Both leak a Go
   zero value into a user-facing message.

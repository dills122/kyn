# Instances and reports

Status: proposed. Current state only; history is in
[`notes.md`](notes.md).

## 1. What an instance is

A rule is evaluated once per **instance**: the group of changed source files
that agree on every variable the rule's `related` template uses.

> A rule's instance key is the values of **exactly the variables its `related`
> template uses**.

| `related` | Variables used | Instances |
| --- | --- | --- |
| `{dir}/{name}.stories.ts` | `dir`, `name` | one per component |
| `{dir}/README.md` | `dir` | one per directory |
| `locales/fr/{name}.json` | `name` | one per file name |
| `CHANGELOG.md` | none | one for the whole rule |
| *(no `related`)* | none | one for the whole rule |

Today the key is always `{dir}/{base}`, taken from the source file, so a
template that ignores the file name demands the same path once per source file
(OS13: three changed files under `CHANGELOG.md` produce three identical
failures). Deriving the key from the template makes that one failure listing
three sources.

Consequences:

- Source files that agree on every variable `related` uses resolve to the same
  path. The "template resolves differently within one instance" error
  (`internal/family/resolver.go:117`) cannot occur, and neither can the
  contradiction in the shipped `api` preset (OS11).
- Instances belong to a rule. Two rules with the same `match` each produce
  their own, identically named, instances. Result counts are the same as v2's
  rules × instances (OS9).
- The converse is not guaranteed. Two instances can demand the same path when
  `related` joins variables with text that can also occur inside their values
  (`docs/{a}-{b}.md`). That is reported as two failures, not hidden; see the
  limits in [`04-grammar.md`](04-grammar.md).

`related` may not contain `..`, so a template cannot reach the same file from
different directories by walking back up. The resolved path is checked too: a
captured value can itself be `..`, and a result with a `.` or `..` segment is an
error ([`04-grammar.md`](04-grammar.md), rule 11).

## 2. Instance names

`instanceName` is the key rendered for people: the variable values joined with
`/`, in the order the variables first appear in `related`, with empty values
left out. If nothing is left, the name is `.`. A rule that uses no variables has
a single instance named `(all)`.

| Case | Name |
| --- | --- |
| `{dir}/{name}.stories.ts`, `src/ui/button.component.ts` | `src/ui/button` |
| `{dir}/README.md`, `terraform/vpc/main.tf` | `terraform/vpc` |
| `{dir}/package-lock.json`, `package.json` at the root | `.` |
| `{dir}/{name}.test.tsx`, `page.tsx` at the root | `page` |
| `tests/{pkg}/test_{name}.py`, `src/app.py` | `app` |
| `src/test/java/{pkg}/{name}Test.java` | `com/acme/billing/Invoice` |
| no variables | `(all)` |

The name is unique within a rule. At most one variable in a template can contain
`/` (`{dir}` or a `{**name}` capture; validation enforces this), so every other
value is a single segment. The number of single-segment values is fixed by the
template, which makes the join unambiguous: whatever remains after taking them
is the one variable that can span segments, and an empty value contributes no
component. `.` can only be produced when every value is empty, and a real
directory is never named `.`. An earlier draft rendered the root as `(root)`,
which collides with a real directory of that name (`(root)` is a common
Next.js route group).

Without the one-variable rule, two multi-segment values could render the same
string (`p/q` + `r` and `p` + `q/r`).

Internally the key is the NUL-joined values; it is never reported. Names are
derived from source-side variables only, never from the resolved related path.

## 3. Ordering

Results sort by `ruleId`, then `instanceName`, byte-wise. Source files within an
instance are sorted. Rule IDs are processed in sorted order, so the first error
reported is also stable (OS7 is the record of what happens otherwise).

## 4. Report fields

v3 removes the family vocabulary. There are no consumers to break, so fields are
named for what they are.

| Output | v2 | v3 |
| --- | --- | --- |
| `json`, `explain` JSON | `ruleId`, `familyId`, `familyName` | `ruleId`, `instanceName` |
| `text` | `Family:`, `Instance:` | `Instance:` |
| `sarif` | `properties.familyId`, `familyName` on results; `familyId` on rules; rule description = first failing message | `properties.instanceName`, `reason`, `relatedState` and `expectedFiles` on results; `relatedLocations` kept; rule `fullDescription` = `why`, never an instance's message |
| `rdjson`, `checkstyle` | message suffix `(family instance: src/a)` | `(instance: src/a, removed by this change)`: the instance, then the reason text |
| `--dry-run-resolve` | instances with a `kin` map | per rule: instances, each with `sources` and `related` |
| RDJSON/Checkstyle location | first changed file, falling back to the family name as a path | first changed file; no fallback, since an instance name is not a path |
| `check` JSON | `ok`, `passed`, `failed`, `infos`, `warnings`, `errors`, `flags`; per result `ruleId`, `familyId`, `familyName`, `changedFiles`, `expectedFiles` | `ok`, `headline`, `inputMode`, `removalInfo`, `summary` (`evaluated`, `passed`, `failed`, `info`, `skipped`, `errors`, `warnings`), `flags`, `rulesWithoutInstances`; per result `ruleId`, `instanceName`, `sources` (was `changedFiles`), `related` (was `expectedFiles`), `reason`, `relatedState`. Every removed or renamed key is listed in [`06-cutover.md`](06-cutover.md) |
| `--fail-on-empty` | "no family instances match" | fails when no rule produced any instance, skipped or not |

`explain` traces name the fields the author wrote. Text sections are `SkipIf:`
and `Expect:`; the JSON keys are `skipIf` and `expect`; clause strings are
`skipIf.missing`, `expect.touched`, and so on. A `skipIf` trace states which of
the three states it saw.

`why` is carried into SARIF `fullDescription` and `explain`. Today
`description` is accepted and read by nothing, while SARIF fills both
description slots with the message (OS10).

## 5. Skipped results are visible

Today `check` has no representation for a skipped rule: no status, no counter,
and the result is absent from JSON. Only `explain` shows it. A rule that never
applies is indistinguishable from a rule that passed.

In v3:

- `check` counts skips in text (`Skipped: N`) and JSON (`skipped`), always.
- JSON `results` include skipped entries with `status: "skipped"` and a short
  `reason`.
- Text lists skipped results with `--show-passes`, alongside passes.
- Machine formats that carry only failures (SARIF, RDJSON, Checkstyle) are
  unchanged.

This covers instances excluded by `changeType` as well as by `skipIf`, so a rule
that is configured but never applies shows up as skips rather than as silence.

## 6. Headline

| Condition | Headline | Exit |
| --- | --- | --- |
| a failure at or above `--fail-on` | `FAIL` | 1 |
| failures exist, none at or above `--fail-on` | `NON-BLOCKING` | 0 |
| no failures | `PASS` | 0 |
| `--fail-on-empty` and no rule produced an instance | `FAIL (nothing matched: --fail-on-empty)` | 1 |

Today a failing `severity: warn` rule prints `PASS` directly above
`Rules failed: 1` (OS12), and the shipped `web-ui` preset triggers it. `PASS`
becomes a claim that nothing failed.

## 7. Normalized model

Ordered slices, never maps.

```go
type Policy struct {
    Rules []Rule // sorted by ID
}

type Rule struct {
    ID         string
    Match      []Glob      // compiled; each carries its capture names
    Exclude    []Glob
    Related    Template    // empty for a related-less informational rule
    Expect     []Assertion // sorted; or the single value none
    SkipIf     Gate        // never | missing | exists
    ChangeType []changes.Status
    Severity   string
    Message    string // generated when empty
    Why        string
    Flags      []string
    Origin     Position // file, line, column, for validation errors
}
```

## 8. Report contracts

The examples below are the contract; golden fixtures are recorded from them. **The field names are provisional** until the maintainer has reviewed them.
Keys are listed in the order they are written. Ordering of arrays is §3, and
`flags` and `rulesWithoutInstances` are sorted and deduplicated.

`check --format json`, one failure from a deletion, one skip:

```json
{
  "ok": false,
  "headline": "FAIL",
  "inputMode": "git",
  "removalInfo": true,
  "summary": {"evaluated": 1, "passed": 0, "failed": 1, "info": 0, "skipped": 1, "errors": 1, "warnings": 0},
  "flags": [],
  "rulesWithoutInstances": ["changelog"],
  "results": [
    {
      "ruleId": "handler-test",
      "instanceName": "internal/orders/order",
      "status": "fail",
      "severity": "error",
      "message": "Handler changed but its test did not.",
      "reason": "removed",
      "sources": ["internal/orders/order_handler.go"],
      "related": "internal/orders/order_handler_test.go",
      "relatedState": "removed",
      "touched": false,
      "flags": []
    },
    {
      "ruleId": "story-sync",
      "instanceName": "src/ui/button",
      "status": "skipped",
      "severity": "error",
      "reason": "related_missing",
      "sources": ["src/ui/button.component.ts"],
      "related": "src/ui/button.stories.ts",
      "relatedState": "missing",
      "touched": false,
      "flags": []
    }
  ]
}
```

Definitions. `ok` is true when no failure is at or above `--fail-on`. `inputMode`
is `git`, `name-status` or `paths`; `removalInfo` is false for `paths`. In
`summary`, `evaluated` is `passed + failed + info` (skipped results are not
evaluated); `errors` and `warnings` count **failed** results by severity (v2 counted
every result by severity, which inflated `errors` for informational rules, OS14).
Top-level `flags` is the union of every applied rule's flags, and is present in
`--summary-only` output. `reason` is absent on `pass` and `info`. A rule with no
instance appears in `rulesWithoutInstances`, never as a result. `--summary-only`
is the same object without `results`.

An **empty run** (no instance at all):

```json
{"ok": true, "headline": "PASS", "inputMode": "git", "removalInfo": true, "summary": {"evaluated": 0, "passed": 0, "failed": 0, "info": 0, "skipped": 0, "errors": 0, "warnings": 0}, "flags": [], "rulesWithoutInstances": ["changelog", "handler-test"], "results": []}
```

It exits 0. With `--fail-on-empty` it exits 1, `ok` is false, the headline is
`FAIL (nothing matched: --fail-on-empty)`, and `summary` gains `"empty": true`. SARIF,
RDJSON and Checkstyle carry one synthetic diagnostic, rule `kyn/fail-on-empty`,
message `no rule produced an instance`, as v2 does, so a failing exit is never
without a diagnostic. The text headline separates the three situations a reader
can confuse: `PASS` (evaluated and passed), `PASS (nothing evaluated: N instances,
all skipped)`, and `PASS (nothing matched)`.

`check --format text`, the same failure (golden-grade):

```text
FAIL
Instance: handler-test internal/orders/order
  Source:   internal/orders/order_handler.go
  Related:  internal/orders/order_handler_test.go (removed)
  Expected: touched
  Reason:   removed by this change
  Message:  Handler changed but its test did not.
Rules: 1 evaluated, 1 failed, 0 info; Skipped: 1
Input: git (removal information: yes)
```

`Expected:` prints the `expect` value, which is what would satisfy the rule.
`Message:` is the authored or generated message (§5 of
[`02-semantics.md`](02-semantics.md)). When `inputMode` is `paths` the `Input:`
line says `paths (no removal information)`.

`explain --format json` is the `check` result for each instance plus a `trace`,
and nothing else:

```json
{
  "ruleId": "handler-test",
  "instanceName": "internal/orders/order",
  "status": "fail",
  "reason": "removed",
  "sources": ["internal/orders/order_handler.go"],
  "related": "internal/orders/order_handler_test.go",
  "relatedState": "removed",
  "touched": false,
  "trace": {
    "changeType": {"value": ["added"], "matched": true},
    "skipIf": {"value": "missing", "relatedState": "removed", "skipped": false, "clause": "skipIf.missing"},
    "expect": [{"value": "touched", "passed": false, "clause": "expect.touched"}]
  }
}
```

`changeType` is omitted when the rule has none; `expect` is always a list, one
entry per listed value. `explain` never exits 1 for a result.

`--dry-run-resolve --format json` evaluates nothing and lists every rule,
**zeros included**:

```json
{
  "inputMode": "git",
  "removalInfo": true,
  "changedCount": 2,
  "changedFiles": ["internal/orders/order_handler.go", "internal/orders/order_service.go"],
  "matchedInstances": 1,
  "rules": [
    {"ruleId": "handler-test", "matchedInstances": 1, "instances": [
      {"instanceName": "internal/orders/order", "sources": ["internal/orders/order_handler.go"], "related": "internal/orders/order_handler_test.go", "relatedState": "exists"}
    ]},
    {"ruleId": "handler-tset", "matchedInstances": 0, "instances": []}
  ]
}
```

`changedFiles` is how a user debugs a `--stdin` stream: it lists what Kyn read.

## 9. Resolve preview

`--dry-run-resolve` runs matching, binding and resolution (rules 9 to 11
included) and stops before evaluation. Exit codes: 0 when it resolved, 2 for a
config or resolve-time error, 1 with `--fail-on-empty` when no rule produced an
instance. Text output prints one line per rule with its instance count, so a
misspelled rule shows as `handler-tset: 0` even when another rule matched. This
is what the recipe in [`04-grammar.md`](04-grammar.md) relies on.

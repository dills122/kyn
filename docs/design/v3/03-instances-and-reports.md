# Instances and reports

Status: proposed. Current state only; history is in
[`00-workplan.md`](00-workplan.md).

## 1. What an instance is

A rule is evaluated once per **instance**: one related path, together with
every changed source file that resolves to it.

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

Two consequences:

- Source files that agree on every variable `related` uses always resolve to
  the same path. The "template resolves differently within one instance" error
  (`internal/family/resolver.go:117`) cannot occur, and neither can the
  contradiction in the shipped `api` preset (OS11).
- Instances belong to a rule. There is no shared grouping across rules; two
  rules with the same `match` each produce their own, identically named,
  instances. Result counts are the same as v2's rules × instances (OS9).

## 2. Instance names

`instanceName` is the key rendered for people: the variable values joined with
`/`, in the order the variables first appear in `related`.

| Case | Name |
| --- | --- |
| `{dir}/{name}.stories.ts`, `src/ui/button.component.ts` | `src/ui/button` |
| `{dir}/README.md`, `terraform/vpc/main.tf` | `terraform/vpc` |
| `src/test/java/{pkg}/{name}Test.java` | `com/acme/billing/Invoice` |
| no variables | `(all)` |
| `{dir}` resolving to the repository root | `(root)` in place of the empty directory |

The name is unique within a rule. Single-segment captures cannot contain `/`,
and validation allows at most one variable that can (`{dir}` or a `{**name}`
capture) in a `related` template, so the number of components is fixed and the
join can be split back unambiguously. Without that rule two multi-segment
values could render the same string (`p/q` + `r` and `p` + `q/r`).

Internally the key is the rule ID and the NUL-joined values; it is never
reported.

Names are derived from source-side variables only, never from the resolved
related path.

## 3. Ordering

Results sort by `ruleId`, then `instanceName`, byte-wise. Source files within
an instance are sorted. Rule IDs are processed in sorted order, so the first
error reported is also stable (OS7 is the record of what happens otherwise).

## 4. Report fields

v3 removes the family vocabulary. There are no consumers to break, so fields
are named for what they are.

| Output | v2 | v3 |
| --- | --- | --- |
| `json`, `explain` JSON | `ruleId`, `familyId`, `familyName` | `ruleId`, `instanceName` |
| `text` | `Family:`, `Instance:` | `Instance:` |
| `sarif` | `properties.familyId`, `familyName` on results; `familyId` on rules | `properties.instanceName` on results; nothing on rules |
| `rdjson`, `checkstyle` | message suffix `(family instance: src/a)` | `(instance: src/a)` |
| `--dry-run-resolve` | instances with a `kin` map | per rule: instances, each with `sources` and `related` |
| RDJSON/Checkstyle location | first changed file, falling back to the family name as a path | first changed file; no fallback to the instance name, which is not a path |

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
- JSON `results` include skipped entries with `status: "skipped"`.
- Text lists skipped results with `--show-passes`, alongside passes.
- Machine formats that carry only failures (SARIF, RDJSON, Checkstyle) are
  unchanged.

## 6. Headline

| Condition | Headline | Exit |
| --- | --- | --- |
| a failure at or above `--fail-on` | `FAIL` | 1 |
| failures exist, none at or above `--fail-on` | `NON-BLOCKING` | 0 |
| no failures | `PASS` | 0 |

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
    Match      []Glob   // compiled; each carries its capture names
    Exclude    []Glob
    Related    Template // empty for a related-less informational rule
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

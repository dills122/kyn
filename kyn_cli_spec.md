# Kyn CLI Specification

## Working Name

**Kyn**

## One-line Description

Kyn is a lightweight CLI for detecting related file changes and enforcing file relationship rules in CI.

## Primary Goal

Build a stateless command-line tool that evaluates a set of changed files against configurable file relationship rules. It should work in normal repositories, package-based repositories, and monorepos such as Nx or Rush.js.

Kyn should answer questions like:

- If a component file changed, did its related story file also change?
- If a component has a Figma config sidecar, should CI mark it for Figma publishing?
- If a public model changed, did the related spec, docs, or schema file also change?
- If a specific family of files changed, should a specific CI rule fail?

The MVP should be a basic CLI only. No daemon, no server, no watcher, no long-running process.

---

# Language Decision

## Preferred Language: Go

Use **Go** for the first version.

### Why Go Fits Kyn

- Easy to ship as a single binary.
- Fast startup time, which is ideal for CI.
- Simple cross-platform builds.
- Strong standard library for filesystem and process execution.
- Lower implementation complexity than Rust for this type of CLI.
- Easier for most teams to maintain.

### Why Not Rust for V1

Rust would also be a good fit, especially if maximum safety, performance, and richer parser internals become important later. However, for this tool, the hard parts are mostly:

- path matching
- config parsing
- git diff collection
- rule evaluation
- reporting

These do not require Rust-level performance or memory control. Go is the better default for a portable CI utility.

---

# Product Scope

## In Scope for MVP

- CLI command: `kyn check`
- Load config from explicit path or default locations.
- Accept changed files directly via CLI.
- Accept changed files from a file.
- Detect changed files using `git diff`.
- Evaluate sibling/related-file rules.
- Support glob-based matching.
- Support file existence checks.
- Support basic pass/fail output.
- Support text and JSON output.
- Exit with deterministic exit codes.

## Out of Scope for MVP

- Daemon/watch mode.
- PR comments.
- GitHub/GitLab API integration.
- Nx project graph parsing.
- Rush project graph parsing.
- Automatic Figma publishing.
- Automatic Storybook publishing.
- Autofixes or file creation.
- SARIF/checkstyle/reviewdog output.
- Plugin system.

These can be added after the core CLI works.

---

# Core Mental Model

Kyn evaluates changed files by grouping related files into **families**.

A family represents one logical unit of code or configuration.

Example Angular component family:

```txt
libs/ui/button/button.component.ts
libs/ui/button/button.component.html
libs/ui/button/button.component.scss
libs/ui/button/button.spec.ts
libs/ui/button/button.stories.ts
libs/ui/button/figma.button.json
```

Kyn should understand that these files are related because they share a directory, base name, naming convention, or configured relationship pattern.

The evaluation flow is:

```txt
Collect changed files
  -> Load config
  -> Match changed files to family definitions
  -> Resolve expected kin/sibling files
  -> Evaluate rules
  -> Print report
  -> Exit 0 or nonzero
```

---

# Terminology

## Family

A group of related files that represent one logical unit.

Example:

```txt
button.component.ts
button.component.html
button.component.scss
button.stories.ts
button.spec.ts
```

## Kin

A related file within a family.

Examples:

- `story`
- `spec`
- `template`
- `style`
- `figma`
- `docs`

## Rule

A configured policy that evaluates a file family and returns pass/fail/info/warn/error.

## Change Set

The list of files that were added, modified, deleted, or renamed.

For MVP, Kyn can treat all changed files as a single flat list. Added/modified/deleted categories can be added as optional metadata later.

---

# CLI Contract

## Main Command

```bash
kyn check
```

## Supported Flags

```txt
--config <path>         Path to Kyn config file
--files <csv>           Comma-separated changed files
--files-from <path>     Path to file containing changed files, one per line; use '-' for stdin
--base <ref>            Git base ref/SHA for diff detection
--head <ref>            Git head ref/SHA for diff detection
--cwd <path>            Working directory; defaults to current directory
--format <format>       Output format: text | json; default text
--fail-on <level>       Minimum severity that fails the command: error | warn; default error
--fail-on-empty         Fail when no family instances match; default false
--show-passes           Include passing rule results in text output; default false
--verbose               Print diagnostic information
```

## Usage Examples

### Explicit files

```bash
kyn check --config kyn.config.yaml --files libs/ui/button/button.component.ts,libs/ui/button/button.component.html
```

### Files from file

```bash
kyn check --config kyn.config.yaml --files-from changed-files.txt
```

### Files from stdin

```bash
git diff --name-only origin/main...HEAD | kyn check --config kyn.config.yaml --files-from -
```

### Git diff mode

```bash
kyn check --config kyn.config.yaml --base origin/main --head HEAD
```

### JSON output

```bash
kyn check --config kyn.config.yaml --base origin/main --head HEAD --format json
```

### Fail on warnings

```bash
kyn check --config kyn.config.yaml --base origin/main --head HEAD --fail-on warn
```

---

# Input Mode Rules

Exactly one change input mode must be selected:

1. `--files`
2. `--files-from`
3. `--base` and `--head` using git diff

Invalid combinations are CLI usage errors (exit code `2`).

If git mode is used, both `--base` and `--head` are required.

For MVP, do not silently infer default git refs unless explicitly configured later.

---

# Exit Codes

```txt
0 = all required rules passed
1 = one or more rules failed
2 = invalid CLI usage or invalid config
3 = runtime/provider error, such as git failure
```

---

# Config Format

## MVP Config Format

Use YAML for MVP.

Reasoning:

- Easier to write than JSON.
- Easier to validate than executable JS/TS.
- Language-neutral for a Go CLI.
- Better fit for CI tooling.

Default config lookup order:

```txt
kyn.config.yaml
kyn.config.yml
.kyn.yaml
.kyn.yml
```

If `--config` is provided, use only that path.

---

# Example Config

```yaml
version: 1

families:
  - id: angular-component
    include:
      - "libs/**/*.component.ts"
      - "libs/**/*.component.html"
      - "libs/**/*.component.scss"
    baseName:
      stripSuffixes:
        - ".component"
    kin:
      story: "{dir}/{base}.stories.ts"
      spec: "{dir}/{base}.spec.ts"
      figma: "{dir}/figma.{base}.json"

rules:
  - id: storybook-sync
    description: "Component changes should be reviewed against Storybook stories."
    family: angular-component
    severity: error
    when:
      changedAny:
        - source
      kinExists:
        - story
    require:
      kinChanged:
        - story
    message: "Component changed but its Storybook file did not change."

  - id: figma-publish-required
    description: "Components with Figma configs require a Figma publish check."
    family: angular-component
    severity: warn
    when:
      changedAny:
        - source
      kinExists:
        - figma
    require:
      emitFlag: figmaPublishRequired
    message: "Component changed and has a Figma config. Figma publish may be required."
```

---

# Config Schema Concepts

## Top Level

```yaml
version: 1
families: []
rules: []
```

## Family Definition

```yaml
families:
  - id: string
    include: string[]
    exclude?: string[]
    baseName?:
      stripSuffixes?: string[]
    kin:
      [kinName: string]: string
```

### Family Fields

#### `id`

Unique family identifier.

Example:

```yaml
id: angular-component
```

#### `include`

Glob patterns used to detect files that belong to this family.

Example:

```yaml
include:
  - "libs/**/*.component.ts"
  - "libs/**/*.component.html"
```

#### `exclude`

Optional glob patterns to ignore.

Example:

```yaml
exclude:
  - "**/*.spec.ts"
```

#### `baseName.stripSuffixes`

Suffixes to remove from the filename stem before resolving kin patterns.

Example:

Input:

```txt
button.component.ts
```

Filename stem:

```txt
button.component
```

After stripping `.component`:

```txt
button
```

#### `kin`

A map of related file names to path templates.

Supported template variables:

```txt
{dir}       Directory of changed file
{file}      Full changed file path
{name}      Filename without extension
{base}      Normalized base name after suffix stripping
{ext}       File extension
```

Example:

```yaml
kin:
  story: "{dir}/{base}.stories.ts"
  spec: "{dir}/{base}.spec.ts"
```

---

# Rule Definition

```yaml
rules:
  - id: string
    description?: string
    family: string
    severity: info | warn | error
    when?:
      changedAny?: string[]
      kinExists?: string[]
      kinMissing?: string[]
    require?:
      kinChanged?: string[]
      kinUnchanged?: string[]
      kinExists?: string[]
      kinMissing?: string[]
      emitFlag?: string
    message: string
```

## Rule Behavior

A rule applies to each resolved family instance matching the configured `family` id.

A rule has two phases:

1. `when`: determines if the rule should run.
2. `require`: determines if the rule passes or fails.

If `when` is omitted, the rule runs for every family instance.

If `require` is omitted, the rule should emit an informational result if the `when` block matches.

Predicate semantics for MVP:

- `when` keys are evaluated with `AND` semantics.
- `require` keys are evaluated with `AND` semantics.
- list predicates require all listed kin names unless documented otherwise.
- `changedAny` is `ANY` across listed groups.

---

# Built-in Rule Semantics for MVP

## `changedAny`

Checks whether any files in a named group changed.

For MVP, the default group is `source`.

A file matched by the family `include` patterns is considered part of `source`.

Example:

```yaml
when:
  changedAny:
    - source
```

## `kinExists`

Checks that specific kin files exist in the repository.

Example:

```yaml
when:
  kinExists:
    - story
```

## `kinMissing`

Checks that specific kin files do not exist.

Example:

```yaml
when:
  kinMissing:
    - spec
```

## `kinChanged`

Checks that specific kin files are present in the changed file list.

Example:

```yaml
require:
  kinChanged:
    - story
```

## `kinUnchanged`

Checks that specific kin files are not present in the changed file list.

Example:

```yaml
require:
  kinUnchanged:
    - generated
```

## `emitFlag`

Emits a named output flag instead of validating a file condition.

Example:

```yaml
require:
  emitFlag: figmaPublishRequired
```

This is informational and does not fail by itself.

---

# Result Model

Internally, use a result model similar to:

```go
type Severity string

const (
    SeverityInfo  Severity = "info"
    SeverityWarn  Severity = "warn"
    SeverityError Severity = "error"
)

type ResultStatus string

const (
    StatusPass ResultStatus = "pass"
    StatusFail ResultStatus = "fail"
    StatusInfo ResultStatus = "info"
)

type RuleResult struct {
    RuleID        string            `json:"ruleId"`
    FamilyID      string            `json:"familyId"`
    FamilyName    string            `json:"familyName"`
    Severity      Severity          `json:"severity"`
    Status        ResultStatus      `json:"status"`
    Message       string            `json:"message"`
    ChangedFiles  []string          `json:"changedFiles,omitempty"`
    ExpectedFiles []string          `json:"expectedFiles,omitempty"`
    Metadata      map[string]string `json:"metadata,omitempty"`
}
```

Summary:

```go
type Summary struct {
    OK       bool         `json:"ok"`
    Passed   int          `json:"passed"`
    Failed   int          `json:"failed"`
    Infos    int          `json:"infos"`
    Warnings int          `json:"warnings"`
    Errors   int          `json:"errors"`
    Results  []RuleResult `json:"results"`
    Flags    []string     `json:"flags,omitempty"`
}
```

---

# Text Output Example

```txt
kyn check

FAIL

Rules failed: 1
Warnings: 1
Infos: 0

[ERROR] storybook-sync
Family: angular-component
Instance: libs/ui/button/button
Message: Component changed but its Storybook file did not change.
Changed files:
  - libs/ui/button/button.component.ts
  - libs/ui/button/button.component.html
Expected files:
  - libs/ui/button/button.stories.ts

[WARN] figma-publish-required
Family: angular-component
Instance: libs/ui/button/button
Message: Component changed and has a Figma config. Figma publish may be required.
Expected files:
  - libs/ui/button/figma.button.json
```

---

# JSON Output Example

```json
{
  "ok": false,
  "passed": 1,
  "failed": 1,
  "infos": 0,
  "warnings": 1,
  "errors": 1,
  "flags": ["figmaPublishRequired"],
  "results": [
    {
      "ruleId": "storybook-sync",
      "familyId": "angular-component",
      "familyName": "libs/ui/button/button",
      "severity": "error",
      "status": "fail",
      "message": "Component changed but its Storybook file did not change.",
      "changedFiles": [
        "libs/ui/button/button.component.ts",
        "libs/ui/button/button.component.html"
      ],
      "expectedFiles": [
        "libs/ui/button/button.stories.ts"
      ]
    }
  ]
}
```

---

# Git Diff Provider

For MVP, implement a git provider using:

```bash
git diff --name-status -M <base>...<head>
```

If `--base` and `--head` are provided, run:

```bash
git -C <cwd> diff --name-status -M <base>...<head>
```

Normalize output by:

- trimming whitespace
- dropping empty lines
- converting Windows path separators to `/`
- removing leading `./`

MVP flattening behavior:

- include `A` and `M` paths in changed-set evaluation
- include rename destination path for `R*` statuses
- exclude `D` paths from changed-set evaluation

If git fails, exit with code `3` and show the git error.

---

# File Matching

Use a Go glob library that supports common doublestar patterns such as:

```txt
**/*.component.ts
libs/**/*.stories.ts
```

Recommended package:

```txt
github.com/bmatcuk/doublestar/v4
```

Use slash-normalized relative paths internally.

---

# Filesystem Rules

All path checks should be relative to `--cwd`.

`kinExists` should use the real filesystem, not just changed files.

A kin file is considered changed if its normalized path appears in the changed file set.

Family instances should be deduplicated by default.

---

# Suggested Go Project Structure

```txt
kyn/
  go.mod
  cmd/
    kyn/
      main.go
  internal/
    cli/
      check.go
      flags.go
    config/
      config.go
      load.go
      validate.go
    changes/
      changes.go
      git.go
      manual.go
    matcher/
      glob.go
      normalize.go
    family/
      family.go
      resolver.go
      template.go
    rules/
      engine.go
      evaluator.go
      result.go
    report/
      text.go
      json.go
  testdata/
    angular/
      kyn.config.yaml
      libs/
        ui/
          button/
            button.component.ts
            button.component.html
            button.stories.ts
            figma.button.json
```

---

# Suggested Dependencies

```txt
github.com/spf13/cobra
```

For CLI command parsing.

```txt
gopkg.in/yaml.v3
```

For YAML config parsing.

```txt
github.com/bmatcuk/doublestar/v4
```

For glob matching with `**` support.

Optional:

```txt
github.com/stretchr/testify
```

For test assertions.

---

# MVP Acceptance Criteria

## CLI Behavior

- `kyn check --help` prints usage.
- `kyn check --config ./kyn.config.yaml --files a.ts,b.ts` runs rules.
- `kyn check --config ./kyn.config.yaml --files-from changed.txt` runs rules.
- `kyn check --config ./kyn.config.yaml --base origin/main --head HEAD` runs git diff and rules.
- Missing config returns exit code `2`.
- Invalid config returns exit code `2`.
- Git failure returns exit code `3`.
- Rule failure returns exit code `1`.
- Successful run returns exit code `0`.

## Rule Behavior

- Can detect changed files matching family includes.
- Can normalize base names with configured suffix stripping.
- Can resolve kin paths from templates.
- Can check whether kin files exist.
- Can check whether kin files changed.
- Can fail when required kin files did not change.
- Can emit flags for informational/CI follow-up rules.

## Output Behavior

- Text output is human-readable.
- Text output lists failures before informational/passing results.
- JSON output is valid JSON.
- JSON output includes `ok`, counts, results, and flags.
- `--fail-on error` fails only on error results.
- `--fail-on warn` fails on warn or error results.
- `--fail-on-empty` fails when no family instances matched.
- `--show-passes` includes pass results in text output.
- Output ordering for results/files/flags is deterministic.

---

# Example MVP Test Case

## Files

```txt
libs/ui/button/button.component.ts
libs/ui/button/button.component.html
libs/ui/button/button.component.scss
libs/ui/button/button.stories.ts
libs/ui/button/figma.button.json
```

## Changed Files

```txt
libs/ui/button/button.component.ts
libs/ui/button/button.component.html
```

## Expected Result

- `storybook-sync` fails because `button.stories.ts` exists but did not change.
- `figma-publish-required` emits a warning/flag because `figma.button.json` exists.
- Exit code is `1` when `storybook-sync` severity is `error`.

---

# Future Enhancements

## Monorepo Adapters

Later versions can add providers/adapters for:

- Nx affected files/projects
- Rush.js changed projects
- Turborepo package scopes
- pnpm/yarn/npm workspaces

These should be optional integrations and should not be required for the core CLI.

## Additional Output Formats

Possible future reporters:

- SARIF
- Checkstyle XML
- reviewdog rdjson
- GitLab code quality report
- Markdown summary

## Plugins / Rule Packs

Possible future packages:

```txt
@kyn/angular
@kyn/react
@kyn/storybook
@kyn/figma
```

In Go, this may be better implemented as built-in rule presets rather than runtime plugins.

## Autofix

Future commands could include:

```bash
kyn init
kyn explain
kyn check --suggest
kyn create-missing
```

Do not implement these for MVP.

---

# Implementation Notes for Codex

Please implement this as a Go CLI.

Start with the MVP only. Prioritize clean architecture and tests over extra features.

Use `cobra` for the CLI, `yaml.v3` for config parsing, and `doublestar` for glob matching.

Keep all paths normalized to slash-separated relative paths internally.

Do not implement Nx, Rush, GitHub, GitLab, PR comments, daemon mode, watch mode, or plugins in the first version.

The main deliverable is a working `kyn check` command with config-driven family/rule evaluation.

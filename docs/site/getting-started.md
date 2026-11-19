# Your First Check

This exercise proves one related-file rule before you adapt Kyn to a real
repository. You will resolve one source-to-test relationship, deliberately fail
it, inspect the decision, and then pass it.

Kyn works with paths, so the example files do not need valid Go code and nothing
in this exercise is compiled.

## Before you start

[Install Kyn](install.md), then confirm that the binary is available:

```bash
kyn --version
```

## 1. Create a disposable example

Create an empty directory with one handler and its existing test:

```bash
mkdir -p kyn-first-check/internal/order
cd kyn-first-check
touch internal/order/order_handler.go
touch internal/order/order_handler_test.go
```

The exercise now has this layout:

```text
kyn-first-check/
├── internal/
│   └── order/
│       ├── order_handler.go
│       └── order_handler_test.go
└── kyn.config.yaml  # create this next
```

## 2. Define the relationship

Create `kyn.config.yaml` with one family and one rule:

```yaml
version: 2

families:
  - id: go-handler
    groups:
      source:
        include:
          - "internal/**/*_handler.go"
    kin:
      test: "{dir}/{name}_test.go"

rules:
  - id: test-sync
    family: go-handler
    severity: error
    if:
      changedAny: [source]
      kinExists: [test]
    assert:
      kinChanged: [test]
    message: "Handler changed but its existing test did not."
```

The important choices are visible in the example:

| Config choice | Meaning in this exercise |
| --- | --- |
| `internal/**/*_handler.go` | Handler paths create family instances. |
| `{dir}/{name}_test.go` | The handler name and directory resolve its test path. |
| `if.kinExists: [test]` | The rule protects tests that already exist. |
| `assert.kinChanged: [test]` | An applicable rule requires the test in the change set. |
| `severity: error` | A failed assertion makes `check` exit 1 by default. |

## 3. Confirm that the path resolves

Supply the handler as a simulated change set and preview the relationship:

```bash
kyn check --files internal/order/order_handler.go --dry-run-resolve
```

The preview should include exactly one family instance and the expected test:

```text
Mode: files
Changed files: 1
Matched instances: 1

[go-handler] internal/order/order_handler
Kin:
  - test: internal/order/order_handler_test.go
```

If `Matched instances` is `0`, stop here and check the source pattern. A green
policy result with zero instances would not prove that this rule ran.

## 4. Deliberately fail the rule

Tell Kyn that only the handler changed:

```bash
kyn check --files internal/order/order_handler.go
```

This failure is intentional. The important lines are:

```text
FAIL

Rules failed: 1

[ERROR] test-sync
Status: fail
Expected files:
  - internal/order/order_handler_test.go
```

`check` exits 1. The handler matched, its test exists, and the test was absent
from the supplied change set. That result proves that the rule is active.

!!! note "Explicit paths simulate a change set"
    `--files` does not edit files or ask Git what changed. It records each path
    you supply as modified, which makes it useful while designing a policy.

## 5. Explain the failure

Run the same input through the diagnostic command:

```bash
kyn explain --files internal/order/order_handler.go
```

Read `If` first to learn why the rule applied, then read `Assert` to find the
failed requirement:

```text
If:
  - if.changedAny: pass (1 source files changed)
  - if.kinExists: pass (test exists (internal/order/order_handler_test.go))
Assert:
  - assert.kinChanged: fail (test was not changed (internal/order/order_handler_test.go))
```

`explain` exits 0 when evaluation and rendering complete, even when it displays
`FAIL`. Use `check` when a failed policy must produce exit 1.

## 6. Pass the rule

Include both paths in the simulated change set:

```bash
kyn check \
  --files internal/order/order_handler.go,internal/order/order_handler_test.go \
  --show-passes
```

The same rule now passes and `check` exits 0:

```text
PASS

Rules failed: 0

[ERROR] test-sync
Status: pass
```

!!! note "Why a passing result still says ERROR"
    `[ERROR]` is the rule's configured severity. `Status` is the outcome of this
    evaluation, and `Message` is static policy text. Read those fields together.

This result proves the path relationship and change membership. It does not
prove that the test contains meaningful assertions. Kyn does not edit files,
inspect their syntax, or run the test suite; tests and reviewers still judge the
quality of the accompanying change.

## 7. Choose what an absent test means

The example protects existing coverage. If
`internal/order/order_handler_test.go` were absent, `if.kinExists` would make
the rule skip rather than fail. That supports gradual adoption in a repository
where not every handler already has a test.

Choose the policy that matches the guarantee you want:

| Policy | Kin absent | Kin exists but is not in the change set |
| --- | --- | --- |
| Protect existing tests with `if.kinExists` | Rule skips | Rule fails |
| Require a test with `assert.kinExists` and `assert.kinChanged` | Rule fails | Rule fails |

For the stricter policy, remove `kinExists` from `if` and put both requirements
under `assert`:

```yaml
if:
  changedAny: [source]
assert:
  kinExists: [test]
  kinChanged: [test]
```

`--fail-on-empty` checks whether zero family instances matched. It does not
turn a skipped rule into a failure.

## 8. Apply the lesson to your repository

You now have a known failure and a known pass. Carry that proof into a real
repository:

1. Generate the [preset](presets.md) closest to your filenames.
2. Adapt one source pattern and kin template.
3. Preview one real source path with `--dry-run-resolve`.
4. Run deliberate failing and passing path lists before relying on the policy.
5. Commit the config, then follow the [CI adoption recipe](ci.md).

!!! warning "Automatic Git mode checks committed changes"
    Inside a Git repository, `kyn check` defaults to `origin/main...HEAD`: the
    commits on your branch since its merge base with `origin/main`. It does not
    include staged, unstaged, or untracked changes. Continue using an explicit
    input such as `--files` when testing a working-tree path deliberately.

Kin existence is checked in the current checkout; selecting Git refs does not
check out their files. See [How Kyn works](concepts.md) for input semantics and
[Troubleshooting](troubleshooting.md) when a rule does not appear.

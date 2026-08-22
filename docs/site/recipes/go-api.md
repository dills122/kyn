# Recipe: Require Tests with Go Handler Changes

## The review problem

An API handler changes without its focused test file. The full suite may still
pass, but reviewers must notice that the behavioral contract was not updated.

Kyn can require the existing handler test to be present in the same change set.

## Example layout

```text
internal/order/
├── order_handler.go
└── order_handler_test.go
```

## Policy

```yaml
version: 2

families:
  - id: go-handler
    groups:
      source:
        include:
          - "internal/**/*_handler.go"
      tests:
        include:
          - "internal/**/*_handler_test.go"
    kin:
      test: "{dir}/{name}_test.go"

rules:
  - id: handler-tests-sync
    family: go-handler
    severity: error
    if:
      changedAny: [source]
      kinExists: [test]
    assert:
      kinChanged: [test]
    message: "Handler changed but its existing test file did not."
```

For `internal/order/order_handler.go`, `{name}` is `order_handler`, so Kyn
resolves `internal/order/order_handler_test.go`.

## Reproduce the missing-test failure

```bash
kyn check -c kyn.config.yaml \
  --files internal/order/order_handler.go \
  --show-passes
```

Expected result:

```text
[ERROR] handler-tests-sync
Status: fail
Expected files:
  - internal/order/order_handler_test.go
```

The command exits 1.

## Pass with the test change

```bash
kyn check -c kyn.config.yaml \
  --files internal/order/order_handler.go,internal/order/order_handler_test.go \
  --show-passes
```

The assertion passes and the command exits 0.

## Surface the failure in code scanning

Use SARIF when your provider accepts it:

```bash
kyn check -c kyn.config.yaml \
  --base origin/main \
  --head HEAD \
  --format sarif > kyn.sarif
```

The report identifies the changed handler as the primary location and the
expected test as a related location.

## Adapt the recipe

- Add separate families for handlers and services when their test naming differs.
- Change `kinExists` from an applicability condition to
  `assert.kinExists` when every new handler must create a test.
- Start with `severity: warn` in a legacy service, then tighten after measuring
  the existing exceptions.

The repository's
[executable Go API fixture](https://github.com/dills122/kyn/tree/main/e2e/projects/go-api)
verifies both the failed SARIF report and passing text report.

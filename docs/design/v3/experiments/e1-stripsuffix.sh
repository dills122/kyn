#!/usr/bin/env bash
# OS4: does `stripSuffixes` change instance cardinality?
#
# `stripSuffixes` feeds the instance key (familyID + {dir}/{base}), so it decides
# how many instances exist -- not just how a template renders. A v3 inline rule
# has no `stripSuffixes`, so this measures what the inline form cannot express.
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
W=$(fixture e1)

mkdir -p "$W/internal/order"
touch "$W/internal/order/order_handler.go" \
      "$W/internal/order/order_service.go" \
      "$W/internal/order/order_test.go" \
      "$W/internal/order/order_handler_test.go" \
      "$W/internal/order/order_service_test.go"

# A: the shipped `api` preset shape.
cat > "$W/with-strip.yaml" <<'YAML'
version: 2
families:
  - id: go-api
    groups:
      source:
        include: ["internal/**/*_handler.go", "internal/**/*_service.go"]
    baseName:
      stripSuffixes: ["_handler", "_service"]
    kin:
      test: "{dir}/{base}_test.go"
rules:
  - id: api-tests-sync
    family: go-api
    severity: error
    if: { kinExists: [test] }
    assert: { kinChanged: [test] }
    message: "Include the test."
YAML

# B: what a v3 inline rule can express -- no stripSuffixes.
cat > "$W/no-strip.yaml" <<'YAML'
version: 2
families:
  - id: go-api
    groups:
      source:
        include: ["internal/**/*_handler.go", "internal/**/*_service.go"]
    kin:
      test: "{dir}/{name}_test.go"
rules:
  - id: api-tests-sync
    family: go-api
    severity: error
    if: { kinExists: [test] }
    assert: { kinChanged: [test] }
    message: "Include the test."
YAML

FILES="internal/order/order_handler.go,internal/order/order_service.go"
for v in with-strip no-strip; do
  echo "##### $v"
  "$KYN" check --cwd "$W" -c "$W/$v.yaml" --files "$FILES" --dry-run-resolve
  "$KYN" check --cwd "$W" -c "$W/$v.yaml" --files "$FILES" --format json
  echo "exit=$?"
  echo
done

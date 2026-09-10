#!/usr/bin/env bash
# OS12: a failing rule below the --fail-on threshold is counted as failed but
# the report headline still says PASS. The shipped web-ui preset's tests-sync
# rule is severity: warn, so this is the common case, not a corner.
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
W=$(fixture e13)
mkdir -p "$W/src"; touch "$W/src/a.go" "$W/src/a_test.go"

cat > "$W/c.yaml" <<'YAML'
version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.go"]
        exclude: ["src/**/*_test.go"]
    kin:
      rel: "{dir}/{name}_test.go"
rules:
  - id: r
    family: fam
    severity: warn
    if: { kinExists: [rel] }
    assert: { kinChanged: [rel] }
    message: "Test not updated."
YAML

for fo in error warn; do
  echo "##### severity: warn, --fail-on $fo"
  "$KYN" check --cwd "$W" -c "$W/c.yaml" --files src/a.go --fail-on "$fo" | head -6
  "$KYN" check --cwd "$W" -c "$W/c.yaml" --files src/a.go --fail-on "$fo" >/dev/null 2>&1
  echo "EXIT=$?"
  echo
done

#!/usr/bin/env bash
# OS5: the v3 common form self-matches. `match: src/**/*.go` also matches the
# related path `{dir}/{name}_test.go`, producing a phantom second instance --
# which the changed-if-present gate then silently skips.
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
W=$(fixture e7)

mkdir -p "$W/src"; touch "$W/src/a.go" "$W/src/a_test.go"
cat > "$W/c.yaml" <<'YAML'
version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.go"]
    kin:
      rel: "{dir}/{name}_test.go"
rules:
  - id: r
    family: fam
    severity: error
    if: { kinExists: [rel] }
    assert: { kinChanged: [rel] }
    message: "m"
YAML

echo "### no exclude: how many instances?"
"$KYN" check --cwd "$W" -c "$W/c.yaml" --files src/a.go,src/a_test.go --dry-run-resolve
echo "### check result"
"$KYN" check --cwd "$W" -c "$W/c.yaml" --files src/a.go,src/a_test.go
echo "exit=$?"

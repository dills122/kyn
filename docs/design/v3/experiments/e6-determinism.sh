#!/usr/bin/env bash
# OS7: is error SELECTION map-iteration-order dependent?
#
# Two paths range over a Go map and return the first hit:
#   internal/family/resolver.go:117  checkKinAgreement, ranges fam.Kin
#   internal/config/validate.go:89   Validate, ranges fam.Kin
# Identical runs should report an identical error. They do not.
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

RUNS="${RUNS:-30}"

echo "### resolve-time: checkKinAgreement (3 conflicting kin templates)"
W=$(fixture e6a)
mkdir -p "$W/src"; touch "$W/src/a.ts" "$W/src/a.html"
cat > "$W/c.yaml" <<'YAML'
version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.ts", "src/**/*.html"]
    kin:
      alpha: "{dir}/{name}{ext}.alpha"
      bravo: "{dir}/{name}{ext}.bravo"
      charlie: "{dir}/{name}{ext}.charlie"
rules:
  - id: r
    family: fam
    severity: error
    assert: { kinChanged: [alpha] }
    message: "m"
YAML
for _ in $(seq 1 "$RUNS"); do
  "$KYN" check --cwd "$W" -c "$W/c.yaml" --files "src/a.ts,src/a.html" 2>&1 \
    | grep -o 'kin "[a-z]*"' | head -1
done | sort | uniq -c | sort -rn

echo
echo "### validation-time: Validate (3 invalid template variables)"
W=$(fixture e6b)
mkdir -p "$W/src"; touch "$W/src/a.ts"
cat > "$W/c.yaml" <<'YAML'
version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.ts"]
    kin:
      alpha: "{dir}/{bogusA}.ts"
      bravo: "{dir}/{bogusB}.ts"
      charlie: "{dir}/{bogusC}.ts"
rules:
  - id: r
    family: fam
    severity: error
    assert: { kinChanged: [alpha] }
    message: "m"
YAML
for _ in $(seq 1 "$RUNS"); do
  "$KYN" check --cwd "$W" -c "$W/c.yaml" --files src/a.ts 2>&1 | grep -o '{bogus[A-C]}'
done | sort | uniq -c | sort -rn

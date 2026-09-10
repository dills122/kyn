#!/usr/bin/env bash
# IR5: inventory every v1/v2 config shape and record whether it LOADS today.
# The v3 loader must reproduce this matrix for any shape marked contractual.
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
W=$(fixture e11)
mkdir -p "$W/src"; touch "$W/src/a.ts" "$W/src/a.spec.ts"

probe() { # $1=label  $2=yaml on stdin
  printf '%s' "$2" > "$W/probe.yaml"
  local out; out=$("$KYN" check --cwd "$W" -c "$W/probe.yaml" --files src/a.ts 2>&1)
  local code=$?
  local verdict
  case $code in
    0|1) verdict="LOADS" ;;
    2)   verdict="REJECT" ;;
    *)   verdict="ERR($code)" ;;
  esac
  local detail=""
  [ "$verdict" = "REJECT" ] && detail=$(printf '%s' "$out" | head -1 | cut -c1-72)
  printf '%-38s %-8s %s\n' "$1" "$verdict" "$detail"
}

RULE_V2='rules:
  - id: r
    family: fam
    severity: error
    if: { kinExists: [spec] }
    assert: { kinChanged: [spec] }
    message: "m"'

printf '%-38s %-8s %s\n' SHAPE VERDICT DETAIL
printf '%s\n' "---------------------------------------------------------------------------------------"

probe "v1 top-level include + when/require" 'version: 1
families:
  - id: fam
    include: ["src/**/*.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
rules:
  - id: r
    family: fam
    severity: error
    when: { kinExists: [spec] }
    require: { kinChanged: [spec] }
    message: "m"'

probe "v1 with groups.source (no include)" 'version: 1
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
'"$RULE_V2"

probe "v1 + baseName.stripSuffixes" 'version: 1
families:
  - id: fam
    include: ["src/**/*.ts"]
    baseName:
      stripSuffixes: [".component"]
    kin:
      spec: "{dir}/{base}.spec.ts"
'"$RULE_V2"

probe "v2 groups.source (native)" 'version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
'"$RULE_V2"

probe "v2 top-level include, no groups" 'version: 2
families:
  - id: fam
    include: ["src/**/*.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
'"$RULE_V2"

probe "v2 top-level include + groups.source" 'version: 2
families:
  - id: fam
    include: ["src/**/*.ts"]
    groups:
      source:
        include: ["src/**/*.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
'"$RULE_V2"

probe "v2 groups without source" 'version: 2
families:
  - id: fam
    groups:
      tests:
        include: ["src/**/*.spec.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
'"$RULE_V2"

probe "v2 inert non-source groups" 'version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.ts"]
      tests:
        include: ["src/**/*.spec.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
'"$RULE_V2"

probe "v2 legacy when/require aliases" 'version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
rules:
  - id: r
    family: fam
    severity: error
    when: { kinExists: [spec] }
    require: { kinChanged: [spec] }
    message: "m"'

probe "v2 mixing if + when" 'version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
rules:
  - id: r
    family: fam
    severity: error
    if: { kinExists: [spec] }
    when: { kinExists: [spec] }
    assert: { kinChanged: [spec] }
    message: "m"'

probe "v2 legacy require.emitFlag" 'version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
rules:
  - id: r
    family: fam
    severity: error
    require: { emitFlag: "needsReview" }
    message: "m"'

probe "v2 actions.emit" 'version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
rules:
  - id: r
    family: fam
    severity: error
    actions: { emit: ["needsReview"] }
    message: "m"'

probe "v2 rule.description" 'version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
rules:
  - id: r
    family: fam
    description: "why this exists"
    severity: error
    if: { kinExists: [spec] }
    assert: { kinChanged: [spec] }
    message: "m"'

probe "v2 omitted message" 'version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
rules:
  - id: r
    family: fam
    severity: error
    if: { kinExists: [spec] }
    assert: { kinChanged: [spec] }'

probe "v2 omitted severity" 'version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
rules:
  - id: r
    family: fam
    if: { kinExists: [spec] }
    assert: { kinChanged: [spec] }
    message: "m"'

probe "version: 3" 'version: 3
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
'"$RULE_V2"

probe "version omitted" 'families:
  - id: fam
    include: ["src/**/*.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
'"$RULE_V2"

probe "unknown top-level field" 'version: 2
patterns:
  p:
    match: ["src/**/*.ts"]
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
'"$RULE_V2"

probe "v2 assert.changedAny" 'version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
rules:
  - id: r
    family: fam
    severity: error
    assert: { changedAny: [source], kinChanged: [spec] }
    message: "m"'

probe "v2 if.changedAny non-source group" 'version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.ts"]
      tests:
        include: ["src/**/*.spec.ts"]
    kin:
      spec: "{dir}/{name}.spec.ts"
rules:
  - id: r
    family: fam
    severity: error
    if: { changedAny: [tests] }
    assert: { kinChanged: [spec] }
    message: "m"'

#!/usr/bin/env bash
# OS9: cardinality contract. Two rules over one family share instances.
# Results = rules x instances, and every result carries the same familyId /
# familyName. Any v3 model must reproduce these counts and this grouping.
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh" || { echo "FATAL: cannot source lib.sh next to this script" >&2; exit 1; }
W=$(fixture e10)

mkdir -p "$W/src"
touch "$W/src/a.component.ts" "$W/src/a.component.html" \
      "$W/src/b.component.ts" \
      "$W/src/a.stories.ts" "$W/src/a.spec.ts" \
      "$W/src/b.stories.ts" "$W/src/b.spec.ts"

cat > "$W/c.yaml" <<'YAML'
version: 2
families:
  - id: web-component
    groups:
      source:
        include: ["src/**/*.component.ts", "src/**/*.component.html"]
    baseName:
      stripSuffixes: [".component"]
    kin:
      story: "{dir}/{base}.stories.ts"
      spec: "{dir}/{base}.spec.ts"
rules:
  - id: story-sync
    family: web-component
    severity: error
    if: { kinExists: [story] }
    assert: { kinChanged: [story] }
    message: "Story out of sync."
  - id: spec-sync
    family: web-component
    severity: warn
    if: { kinExists: [spec] }
    assert: { kinChanged: [spec] }
    message: "Spec out of sync."
YAML

# two source files collapse into instance src/a; src/b.component.ts is its own
FILES="src/a.component.ts,src/a.component.html,src/b.component.ts"

echo "### instances"
"$KYN" check --cwd "$W" -c "$W/c.yaml" --files "$FILES" --dry-run-resolve | sed -n '/^\[/,$p'
echo "### results (rule x instance)"
"$KYN" check --cwd "$W" -c "$W/c.yaml" --files "$FILES" --format json --show-passes \
  | grep -E '"ruleId"|"familyId"|"familyName"|"status"'

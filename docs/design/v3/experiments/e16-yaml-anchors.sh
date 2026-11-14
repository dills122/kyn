#!/usr/bin/env bash
# OS15: YAML anchors and aliases pass through Kyn's strict loader today, so a
# glob list can be shared between declarations with no Kyn feature at all.
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh" || { echo "FATAL: cannot source lib.sh next to this script" >&2; exit 1; }
W=$(fixture e16)
mkdir -p "$W/src"; touch "$W/src/a.component.ts" "$W/src/a.stories.ts"
cat > "$W/c.yaml" <<'YAML'
version: 2
families:
  - id: fam
    groups:
      source:
        include: &component_globs
          - "src/**/*.component.ts"
          - "src/**/*.component.html"
    baseName: { stripSuffixes: [".component"] }
    kin:
      story: "{dir}/{base}.stories.ts"
  - id: fam2
    groups:
      source:
        include: *component_globs
    baseName: { stripSuffixes: [".component"] }
    kin:
      spec: "{dir}/{base}.spec.ts"
rules:
  - id: r
    family: fam
    severity: error
    if: { kinExists: [story] }
    assert: { kinChanged: [story] }
    message: "m"
YAML
"$KYN" check --cwd "$W" -c "$W/c.yaml" --files src/a.component.ts --dry-run-resolve | sed -n '/^\[/,$p'

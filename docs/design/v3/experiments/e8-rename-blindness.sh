#!/usr/bin/env bash
# OS8: git.go records only the DESTINATION of a rename (StatusRenamed) and
# discards the source path. Renaming the related file away is therefore
# indistinguishable from deleting it -- the same silent skip as OS3.
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh" || { echo "FATAL: cannot source lib.sh next to this script" >&2; exit 1; }
W=$(fixture e8)

mkdir -p "$W/src"
git_init
echo a > "$W/src/a.go"; echo t > "$W/src/a_test.go"
g add -A && g commit -qm base

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
    severity: error
    if: { kinExists: [rel] }
    assert: { kinChanged: [rel] }
    message: "Include the test."
YAML

# modify the source, rename its test away
echo a2 >> "$W/src/a.go"
g mv src/a_test.go src/renamed_test.go
g add -A && g commit -qm work

echo "--- diff under evaluation ---"
git -C "$W" diff --name-status -M HEAD~1...HEAD
echo "--- what kyn collected ---"
"$KYN" check --cwd "$W" -c "$W/c.yaml" --base HEAD~1 --head HEAD --dry-run-resolve | sed -n '/Changed file list/,/^$/p'
echo "--- result ---"
"$KYN" explain --cwd "$W" -c "$W/c.yaml" --base HEAD~1 --head HEAD | sed -n '/^\[/,$p' | head -10
"$KYN" check --cwd "$W" -c "$W/c.yaml" --base HEAD~1 --head HEAD >/dev/null 2>&1
echo "check exit=$?"

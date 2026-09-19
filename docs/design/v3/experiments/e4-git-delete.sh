#!/usr/bin/env bash
# OS3: git mode. `docs/decisions.md` excludes D paths from the change set.
# What do the expectations do when the related file is DELETED in the diff?
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh" || { echo "FATAL: cannot source lib.sh next to this script" >&2; exit 1; }
W=$(fixture e4)

mkdir -p "$W/src"
git_init
echo a > "$W/src/a.go"
echo t > "$W/src/a_test.go"
g add -A && g commit -qm base

emit() { cat > "$W/$1.yaml" <<YAML
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
$2
    message: "m"
YAML
}
emit unchanged "    assert:
      kinUnchanged: [rel]"
emit changed-if-present "    if:
      kinExists: [rel]
    assert:
      kinChanged: [rel]"
emit changed "    assert:
      kinChanged: [rel]"

# modify the source, delete its test, in one commit
echo a2 >> "$W/src/a.go"
g rm -q src/a_test.go
g add -A && g commit -qm work

echo "--- diff under evaluation ---"
git -C "$W" diff --name-status -M HEAD~1...HEAD
echo "--- working tree ---"; ls "$W/src/"
echo

for s in unchanged changed-if-present changed; do
  echo "##### $s"
  "$KYN" explain --cwd "$W" -c "$W/$s.yaml" --base HEAD~1 --head HEAD 2>&1 | sed -n '/^\[/,$p' | head -12
  "$KYN" check --cwd "$W" -c "$W/$s.yaml" --base HEAD~1 --head HEAD >/dev/null 2>&1
  echo "check exit=$?"
  echo
done

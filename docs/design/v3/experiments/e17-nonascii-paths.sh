#!/usr/bin/env bash
# OS16: git quotes unusual filenames in `--name-status` output. Kyn runs
# `git diff --name-status -M` without disabling quoting and does not unquote, so
# a non-ASCII path reaches the matcher as the literal text "caf\303\251.go".
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh" || { echo "FATAL: cannot source lib.sh next to this script" >&2; exit 1; }
W=$(fixture e17)

mkdir -p "$W/src"
git_init
echo a > "$W/src/ok.go"; echo a > "$W/src/ok_test.go"
g add -A; g commit -qm base
echo b >> "$W/src/ok.go"
echo b > "$W/src/café.go"; echo b > "$W/src/café_test.go"
g add -A; g commit -qm work

cat > "$W/c.yaml" <<'YAML'
version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.go"]
        exclude: ["src/**/*_test.go"]
    kin:
      t: "{dir}/{name}_test.go"
rules:
  - id: r
    family: fam
    severity: error
    assert: { kinChanged: [t] }
    message: "m"
YAML

echo "### what git prints (default quoting)"
git -C "$W" diff --name-status -M HEAD~1...HEAD
echo
echo "### what git prints with quoting disabled"
git -C "$W" -c core.quotePath=false diff --name-status -M HEAD~1...HEAD
echo
echo "### kyn: changed files and resolved instances (git mode)"
"$KYN" check --cwd "$W" -c "$W/c.yaml" --base HEAD~1 --head HEAD --dry-run-resolve | sed -n '/Changed file list/,$p'

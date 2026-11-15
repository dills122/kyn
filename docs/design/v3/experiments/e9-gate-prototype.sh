#!/usr/bin/env bash
# Validates the gate design (02-semantics.md) WITHOUT changing Kyn.
#
# Claim under test: the related path's state -- exists / removed-by-this-change /
# missing -- is derivable from the git output Kyn already runs,
# `git diff --name-status -M <base>...<head>`, with no extra git call.
#
# What is measured and what is not: the related path's state and whether it is in
# the change set are derived from REAL git output in a real repository. The
# verdict columns (what each gate would do) are computed by this script from the
# rules in 02-semantics.md; the kyn binary is never invoked. e4 and e8 are the
# scripts that measure what the shipped binary does today for the deleted and
# renamed rows.
#
# A removal never causes a skip: `skipIf: missing` skips only when the related
# path is missing, `skipIf: exists` skips only when it exists.
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh" || { echo "FATAL: cannot source lib.sh next to this script" >&2; exit 1; }

REL="src/a_test.go"

scenario() { # $1=label $2=setup-commands
  W=$(fixture "e9-$1")
  mkdir -p "$W/src"
  git_init
  echo a > "$W/src/a.go"; echo t > "$W/$REL"
  g add -A && g commit -qm base
  eval "$2"
  g add -A
  g commit -qm work

  local diff; diff=$(git -C "$W" diff --name-status -M HEAD~1...HEAD)

  # --- what Kyn collects today: A / M / R-destination ---
  local in_change_set=N
  while IFS=$'\t' read -r st p1 p2; do
    case "$st" in
      A*|M*) [ "$p1" = "$REL" ] && in_change_set=Y ;;
      R*)    [ "$p2" = "$REL" ] && in_change_set=Y ;;
    esac
  done <<< "$diff"

  # --- the signal Kyn parses and discards: D, and R-source ---
  local vanished=N
  while IFS=$'\t' read -r st p1 p2; do
    case "$st" in
      D*) [ "$p1" = "$REL" ] && vanished=Y ;;
      R*) [ "$p1" = "$REL" ] && vanished=Y ;;
    esac
  done <<< "$diff"

  local exists_now=N; [ -f "$W/$REL" ] && exists_now=Y
  local state=missing
  if [ "$exists_now" = Y ]; then state=exists
  elif [ "$vanished" = Y ]; then state=removed; fi

  # outcomes for `expect: touched`
  local verdict; verdict=$([ "$in_change_set" = Y ] && echo pass || echo FAIL)
  local old skip_missing skip_exists
  [ "$exists_now" = Y ] && old=$verdict || old=skipped
  [ "$state" = missing ] && skip_missing=skipped || skip_missing=$verdict
  [ "$state" = exists ] && skip_exists=skipped || skip_exists=$verdict

  printf '%-20s %-8s %-6s %-9s %-15s %s\n' \
    "$1" "$state" "$in_change_set" "$old" "$skip_missing" "$skip_exists"
}

printf '%-20s %-8s %-6s %-9s %-15s %s\n' \
  SCENARIO STATE IN-CS OLD-GATE skipIf:missing skipIf:exists
printf '%s\n' "--------------------------------------------------------------------------------"

scenario "related-untouched"  'echo a2 >> "$W/src/a.go"'
scenario "related-updated"    'echo a2 >> "$W/src/a.go"; echo t2 >> "$W/src/a_test.go"'
scenario "related-deleted"    'echo a2 >> "$W/src/a.go"; g rm -q src/a_test.go'
scenario "related-renamed"    'echo a2 >> "$W/src/a.go"; g mv src/a_test.go src/renamed_test.go'
scenario "related-never-made" 'g rm -q src/a_test.go; g commit -qm drop; echo a2 >> "$W/src/a.go"'

#!/usr/bin/env bash
# Validates the gate design (02-semantics.md) WITHOUT changing Kyn.
#
# Claim under test: the related path's state -- exists / removed-by-this-change /
# missing -- is derivable from the git output Kyn already runs,
# `git diff --name-status -M <base>...<head>`, with no extra git call.
#
# Simulates both gates (`skipIf: missing`, `skipIf: exists`) over five scenarios
# and compares `skipIf: missing` with the gate Kyn ships today.
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
  [ "$state" = missing ] && skip_exists=$verdict || skip_exists=skipped

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

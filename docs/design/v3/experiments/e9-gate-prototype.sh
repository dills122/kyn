#!/usr/bin/env bash
# Validates the G1 gate design (02-semantics.md section 4) WITHOUT changing Kyn.
#
# Claim under test: "existedAtBase" is derivable from the git output Kyn already
# runs -- `git diff --name-status -M <base>...<head>` -- with no extra git call.
#
# This simulates the proposed gate over four scenarios and compares it to the
# gate Kyn ships today.
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
  local existed_at_base=N
  { [ "$exists_now" = Y ] || [ "$vanished" = Y ]; } && existed_at_base=Y

  # gate -> outcome, for `when: related-existed` + `expect: in-change-set`
  local old new
  [ "$exists_now"      = Y ] && old=$([ "$in_change_set" = Y ] && echo pass || echo FAIL) || old="skipped"
  [ "$existed_at_base" = Y ] && new=$([ "$in_change_set" = Y ] && echo pass || echo FAIL) || new="skipped"

  printf '%-22s %-9s %-9s %-9s %-11s %-9s %s\n' \
    "$1" "$exists_now" "$vanished" "$in_change_set" "$existed_at_base" "$old" "$new"
}

printf '%-22s %-9s %-9s %-9s %-11s %-9s %s\n' \
  SCENARIO EXISTS-NOW VANISHED IN-CS EXISTED-BASE "OLD-GATE" "NEW-GATE"
printf '%s\n' "-------------------------------------------------------------------------------------------"

scenario "related-untouched"  'echo a2 >> "$W/src/a.go"'
scenario "related-updated"    'echo a2 >> "$W/src/a.go"; echo t2 >> "$W/src/a_test.go"'
scenario "related-deleted"    'echo a2 >> "$W/src/a.go"; g rm -q src/a_test.go'
scenario "related-renamed"    'echo a2 >> "$W/src/a.go"; g mv src/a_test.go src/renamed_test.go'
scenario "related-never-made" 'g rm -q src/a_test.go; g commit -qm drop; echo a2 >> "$W/src/a.go"'

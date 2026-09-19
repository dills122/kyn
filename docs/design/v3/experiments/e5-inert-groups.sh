#!/usr/bin/env bash
# OS6: are non-`source` groups inert? Compare every output mode with and
# without the `groups.story` / `groups.tests` blocks that `kyn init` emits.
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh" || { echo "FATAL: cannot source lib.sh next to this script" >&2; exit 1; }
W=$(fixture e5)

mkdir -p "$W/src"
touch "$W/src/a.component.ts" "$W/src/a.component.html" \
      "$W/src/a.stories.ts" "$W/src/a.spec.ts"

"$KYN" init --cwd "$W" --preset web-ui -c with-groups.yaml >/dev/null

# strip the inert non-source groups
awk '
  /^      (story|tests):$/ { skip=1; next }
  skip && /^        include:$/ { next }
  skip && /^          - / { next }
  { skip=0; print }
' "$W/with-groups.yaml" > "$W/no-groups.yaml"

echo "--- config difference ---"
diff "$W/with-groups.yaml" "$W/no-groups.yaml"
echo

F=src/a.component.ts
for fmt in text json sarif rdjson checkstyle; do
  "$KYN" check --cwd "$W" -c "$W/with-groups.yaml" --files "$F" --format "$fmt" --show-passes > "$W/out-with.$fmt" 2>&1
  "$KYN" check --cwd "$W" -c "$W/no-groups.yaml"   --files "$F" --format "$fmt" --show-passes > "$W/out-no.$fmt"   2>&1
  if diff -q "$W/out-with.$fmt" "$W/out-no.$fmt" >/dev/null; then
    echo "$fmt: IDENTICAL"
  else
    echo "$fmt: DIFFERS"; diff "$W/out-with.$fmt" "$W/out-no.$fmt" | head -5
  fi
done

"$KYN" check --cwd "$W" -c "$W/with-groups.yaml" --files "$F" --dry-run-resolve > "$W/dr-with" 2>&1
"$KYN" check --cwd "$W" -c "$W/no-groups.yaml"   --files "$F" --dry-run-resolve > "$W/dr-no"   2>&1
diff -q "$W/dr-with" "$W/dr-no" >/dev/null \
  && echo "dry-run-resolve: IDENTICAL" \
  || echo "dry-run-resolve: DIFFERS"

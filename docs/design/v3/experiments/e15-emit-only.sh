#!/usr/bin/env bash
# OS14: v2 informational (emit-only) rules. A rule with actions.emit and no
# assertion yields status "info" and adds a flag. Its `severity` still feeds the
# severity counters, so a rule that cannot fail can report `errors: 1`.
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh" || { echo "FATAL: cannot source lib.sh next to this script" >&2; exit 1; }
W=$(fixture e15)
mkdir -p "$W/src"; touch "$W/src/a.go"

printf '%-9s %s\n' SEVERITY "ok failed infos warnings errors status flags"
for sev in info warn error; do
  cat > "$W/c.yaml" <<YAML
version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.go"]
    kin:
      rel: "{dir}/{name}_test.go"
rules:
  - id: r
    family: fam
    severity: $sev
    actions: { emit: ["reviewRequired"] }
    message: "Touched payment code."
YAML
  out=$("$KYN" check --cwd "$W" -c "$W/c.yaml" --files src/a.go --format json --show-passes 2>&1)
  field() { printf '%s' "$out" | sed -n "s/.*\"$1\": \([a-z0-9]*\).*/\1/p" | head -1; }
  status=$(printf '%s' "$out" | sed -n 's/.*"status": "\([a-z]*\)".*/\1/p' | head -1)
  flags=$(printf '%s' "$out" | grep -c '"reviewRequired"')
  printf '%-9s %-5s %-6s %-5s %-8s %-6s %-6s %s\n' "$sev" "$(field ok)" "$(field failed)" "$(field infos)" "$(field warnings)" "$(field errors)" "$status" "$flags"
done

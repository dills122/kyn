#!/usr/bin/env bash
# OS13: the instance key is {dir}/{base}, derived from the SOURCE file, while
# the demanded path comes from the template. When the template does not use the
# per-file variables, the key over-partitions and one file is demanded many
# times.
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh" || { echo "FATAL: cannot source lib.sh next to this script" >&2; exit 1; }

probe() { # $1=label $2=template
  W=$(fixture "e14-$1")
  mkdir -p "$W/src"; touch "$W/src/a.go" "$W/src/b.go" "$W/src/c.go" "$W/CHANGELOG.md"
  touch "$W/src/a_test.go" "$W/src/b_test.go" "$W/src/c_test.go"
  cat > "$W/c.yaml" <<YAML
version: 2
families:
  - id: fam
    groups:
      source:
        include: ["src/**/*.go"]
        exclude: ["src/**/*_test.go"]
    kin:
      rel: "$2"
rules:
  - id: r
    family: fam
    severity: error
    assert: { kinChanged: [rel] }
    message: "m"
YAML
  local inst; inst=$("$KYN" check --cwd "$W" -c "$W/c.yaml" \
    --files src/a.go,src/b.go,src/c.go --dry-run-resolve 2>&1 \
    | sed -n 's/Matched instances: //p')
  local failed; failed=$("$KYN" check --cwd "$W" -c "$W/c.yaml" \
    --files src/a.go,src/b.go,src/c.go --format json 2>&1 \
    | sed -n 's/.*"failed": \([0-9]*\).*/\1/p')
  local demanded; demanded=$("$KYN" check --cwd "$W" -c "$W/c.yaml" \
    --files src/a.go,src/b.go,src/c.go --format json 2>&1 \
    | grep -o '"[^"]*\.\(md\|go\)"' | grep -v 'src/[abc]\.go' | sort -u | tr -d '"' | paste -sd, -)
  printf '%-26s %-12s %-9s %s\n' "$2" "${inst:-?}" "${failed:-0}" "$demanded"
}

printf '%-26s %-12s %-9s %s\n' TEMPLATE INSTANCES FAILURES "DISTINCT PATHS DEMANDED"
printf '%s\n' "-------------------------------------------------------------------------"
probe percomponent '{dir}/{name}_test.go'
probe perdir       '{dir}/README.md'
probe constant     'CHANGELOG.md'

#!/usr/bin/env bash
# OS1/OS2: ground-truth semantics for every candidate v3 expectation, across
# (related exists on disk) x (related in change set), in --files mode.
#
# The source glob excludes the related path on purpose; see e7-selfmatch.sh for
# what happens without that exclude.
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

emit_cfg() { # $1=path $2=if-block $3=assert-block
cat > "$1" <<YAML
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
$3
    message: "m"
YAML
}

printf '%-22s %-8s %-8s %-9s %-4s %s\n' EXPECTATION EXISTS IN-CS STATUS EXIT EXPECTED-FILES
printf '%s\n' "--------------------------------------------------------------------------------"

for shape in changed-if-present exists-and-changed changed unchanged exists missing; do
  for exists in Y N; do
    for incs in Y N; do
      W=$(fixture e3)
      mkdir -p "$W/src"
      touch "$W/src/a.go"
      [ "$exists" = Y ] && touch "$W/src/a_test.go"

      case "$shape" in
        changed-if-present) emit_cfg "$W/c.yaml" "    if:
      kinExists: [rel]" "    assert:
      kinChanged: [rel]" ;;
        exists-and-changed) emit_cfg "$W/c.yaml" "" "    assert:
      kinExists: [rel]
      kinChanged: [rel]" ;;
        changed)   emit_cfg "$W/c.yaml" "" "    assert:
      kinChanged: [rel]" ;;
        unchanged) emit_cfg "$W/c.yaml" "" "    assert:
      kinUnchanged: [rel]" ;;
        exists)    emit_cfg "$W/c.yaml" "" "    assert:
      kinExists: [rel]" ;;
        missing)   emit_cfg "$W/c.yaml" "" "    assert:
      kinMissing: [rel]" ;;
      esac

      files="src/a.go"
      [ "$incs" = Y ] && files="src/a.go,src/a_test.go"

      out=$("$KYN" explain --cwd "$W" -c "$W/c.yaml" --files "$files" --format json 2>&1)
      status=$(printf '%s' "$out" | grep -o '"status": "[a-z]*"' | tail -1 | sed 's/.*: "//;s/"//')
      expf=$(printf '%s' "$out" | sed -n '/"expectedFiles"/,/]/p' | grep -o '"src/[^"]*"' | tr -d '"' | paste -sd, -)
      "$KYN" check --cwd "$W" -c "$W/c.yaml" --files "$files" >/dev/null 2>&1
      code=$?
      printf '%-22s %-8s %-8s %-9s %-4s %s\n' "$shape" "$exists" "$incs" "${status:-ERR}" "$code" "${expf:--}"
    done
  done
done

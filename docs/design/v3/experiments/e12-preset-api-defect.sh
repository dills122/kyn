#!/usr/bin/env bash
# OS11: the shipped `api` preset combines stripSuffixes ["_handler","_service"]
# (which collapses handler+service into ONE instance) with a "{name}" template
# (which differs per source file). The two disagree, so an ordinary change set
# hits the kin-agreement guard and the whole run fails with exit 2.
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
W=$(fixture e12)

"$KYN" init --cwd "$W" --preset api >/dev/null
mkdir -p "$W/internal/order"
touch "$W/internal/order/order_handler.go" "$W/internal/order/order_service.go" \
      "$W/internal/order/order_handler_test.go" "$W/internal/order/order_service_test.go"

echo "### the shipped preset's family"
sed -n '/baseName:/,/kin:/p;/    kin:/,+1p' "$W/kyn.config.yaml"

echo
echo "### handler alone -- fine"
"$KYN" check --cwd "$W" -c "$W/kyn.config.yaml" --files internal/order/order_handler.go --dry-run-resolve 2>&1 | tail -5
"$KYN" check --cwd "$W" -c "$W/kyn.config.yaml" --files internal/order/order_handler.go >/dev/null 2>&1
echo "exit=$?"

echo
echo "### handler AND service together -- both match the preset's own globs"
"$KYN" check --cwd "$W" -c "$W/kyn.config.yaml" \
  --files internal/order/order_handler.go,internal/order/order_service.go 2>&1 | head -3
"$KYN" check --cwd "$W" -c "$W/kyn.config.yaml" \
  --files internal/order/order_handler.go,internal/order/order_service.go >/dev/null 2>&1
echo "exit=$?"

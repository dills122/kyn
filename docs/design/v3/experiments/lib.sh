#!/usr/bin/env bash
# Shared setup for the v3 design experiments.
#
# Each experiment builds a disposable fixture under $WORKDIR and runs the kyn
# binary against it. Nothing in the repository is modified.
#
#   go build -o /tmp/kyn ./cmd/kyn
#   KYN=/tmp/kyn bash docs/design/v3/experiments/e1-stripsuffix.sh
set -u

KYN="${KYN:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)/kyn}"
if [ ! -x "$KYN" ]; then
  echo "kyn binary not found at $KYN" >&2
  echo "build it first:  go build -o /tmp/kyn ./cmd/kyn   then re-run with KYN=/tmp/kyn" >&2
  exit 1
fi

WORKROOT="${WORKROOT:-${TMPDIR:-/tmp}/kyn-v3-experiments}"

# fixture <name> -> echoes a clean directory path
fixture() {
  local d="$WORKROOT/$1"
  rm -rf "$d"
  mkdir -p "$d"
  printf '%s' "$d"
}

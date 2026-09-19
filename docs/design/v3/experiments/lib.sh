#!/usr/bin/env bash
# Shared setup for the v3 design experiments.
#
# Each experiment builds a disposable fixture under $WORKROOT and runs the kyn
# binary against it. Nothing in the repository is modified.
#
#   go build -o /tmp/kyn ./cmd/kyn
#   KYN=/tmp/kyn bash docs/design/v3/experiments/e1-stripsuffix.sh
#
# Two safety rules, both learned the hard way:
#
# 1. Every git call against a fixture goes through `g` or `git -C "$W"`. Never
#    `cd`. An empty or failed fixture turns `cd "$W"` into a no-op, which left
#    an earlier version of these scripts running `git add -A && git commit`
#    against whatever repository the user invoked them from.
#
# 2. `set -e` is deliberately NOT used. These scripts measure kyn's exit codes,
#    and a failing `kyn check` is usually the observation, not an error. Setup
#    steps fail loudly via `die`/`g` instead, so a broken fixture aborts rather
#    than printing a plausible but meaningless result table.
set -u

die() { echo "FATAL: $*" >&2; exit 1; }

KYN="${KYN:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)/kyn}"
if [ ! -x "$KYN" ]; then
  echo "kyn binary not found at $KYN" >&2
  echo "build it first:  go build -o /tmp/kyn ./cmd/kyn   then re-run with KYN=/tmp/kyn" >&2
  exit 1
fi

WORKROOT="${WORKROOT:-${TMPDIR:-/tmp}/kyn-v3-experiments}"

# fixture <name> -> echoes a clean directory path, or aborts.
fixture() {
  [ "$#" -eq 1 ] && [ -n "${1:-}" ] || die "fixture: needs a name"
  local d="$WORKROOT/$1"
  rm -rf "$d" || die "fixture: cannot clear $d"
  mkdir -p "$d" || die "fixture: cannot create $d"
  [ -d "$d" ] || die "fixture: $d missing after mkdir"
  printf '%s' "$d"
}

# g <git args...> -- git inside "$W", aborting on failure.
# Setup only. Never wrap a measurement in this.
g() {
  [ -n "${W:-}" ] || die "g: \$W is unset"
  [ -d "$W" ] || die "g: \$W ($W) is not a directory"
  git -C "$W" "$@" || die "git -C $W $* failed"
}

# git_init -- a committable repo in "$W", isolated from the caller's git config.
git_init() {
  g init -q .
  g config user.email design@example.invalid
  g config user.name design
  g config commit.gpgsign false
}

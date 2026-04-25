#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

chmod +x .githooks/pre-commit .githooks/pre-push
git config core.hooksPath .githooks

echo "Git hooks installed via core.hooksPath=.githooks"
echo "Configured hooks:"
echo "  - pre-commit (gofmt staged files + go vet)"
echo "  - pre-push (go test ./...)"


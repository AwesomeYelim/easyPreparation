#!/usr/bin/env bash
# Commit docs/context.md alone to the default branch, whatever branch this session is on.
# Run it when a session ends (local or cloud) after updating docs/context.md.
set -euo pipefail
root="$(git rev-parse --show-toplevel)"
cd "$root"
base="$(git symbolic-ref --short refs/remotes/origin/HEAD 2>/dev/null | sed 's#^origin/##' || true)"
base="${base:-master}"
git fetch -q origin "$base"
tmp="$(mktemp -d)"
cleanup() { git -C "$root" worktree remove --force "$tmp" >/dev/null 2>&1 || rm -rf "$tmp"; }
trap cleanup EXIT
git worktree add -q --detach "$tmp" "origin/$base"
cp docs/context.md "$tmp/docs/context.md"
cd "$tmp"
git add docs/context.md
if git diff --cached --quiet; then
  echo "context: no change"
  exit 0
fi
git commit -q -m "docs: update context [skip ci]"
git push -q origin "HEAD:$base"
echo "context: pushed to $base"

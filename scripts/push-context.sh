#!/usr/bin/env bash
# Put docs/context.md on master on its own, whatever branch this session is on.
# Run it when a session ends (local or cloud) after updating docs/context.md.
# If master cannot be pushed, it pushes the file on the current branch instead; the life repo's
# sync-context workflow then copies the newest docs/context.md from any branch to master within the hour.
set -euo pipefail
root="$(git rev-parse --show-toplevel)"
cd "$root"
base="master"
git fetch -q origin "$base"
tmp="$(mktemp -d)"
cleanup() { git -C "$root" worktree remove --force "$tmp" >/dev/null 2>&1 || rm -rf "$tmp"; }
trap cleanup EXIT
git worktree add -q --detach "$tmp" "origin/$base"
cp docs/context.md "$tmp/docs/context.md"
git -C "$tmp" add docs/context.md
if git -C "$tmp" diff --cached --quiet; then
  echo "context: no change on $base"
  exit 0
fi
git -C "$tmp" commit -q -m "docs: update context [skip ci]"
if git -C "$tmp" push -q origin "HEAD:$base"; then
  echo "context: pushed to $base"
  exit 0
fi
branch="$(git branch --show-current)"
[ -n "$branch" ] && [ "$branch" != "$base" ] || { echo "context: push to $base failed" >&2; exit 1; }
git add docs/context.md
git diff --cached --quiet -- docs/context.md || git commit -q -m "docs: update context [skip ci]" -- docs/context.md
git push -q origin "HEAD:$branch"
echo "context: $base push failed, pushed $branch instead (life copies it to $base within the hour)"

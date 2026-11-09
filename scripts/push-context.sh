#!/usr/bin/env bash
# Commit docs/context.md alone on the current branch and push that branch.
# Run it when a session ends (local or cloud) after updating docs/context.md.
# Any branch is fine: the life repo's sync-context workflow takes the newest docs/context.md
# from all branches every hour and copies it to master, so the next session starts from it.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
branch="$(git branch --show-current)"
[ -n "$branch" ] || { echo "context: detached HEAD, check out a branch first" >&2; exit 1; }
git add docs/context.md
if git diff --cached --quiet -- docs/context.md; then
  echo "context: no change"
else
  git commit -q -m "docs: update context [skip ci]" -- docs/context.md
fi
git push -q origin "HEAD:$branch"
echo "context: pushed $branch (master gets it within the hour)"

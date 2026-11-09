#!/bin/bash
# update-baseline.sh — visual regression baseline 갱신
# 현재 캡처(.claude/screenshots/current/)를 baseline/으로 복사.
# 사용자가 의도된 UI 변경 후 수동 실행.

set -euo pipefail

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
SCREENSHOT_BASE="$PROJECT_DIR/.claude/screenshots"
CURRENT_DIR="$SCREENSHOT_BASE/current"
BASELINE_DIR="$SCREENSHOT_BASE/baseline"

if [ ! -d "$CURRENT_DIR" ] || [ -z "$(ls -A "$CURRENT_DIR" 2>/dev/null || true)" ]; then
  echo "현재 캡처가 없습니다. 먼저 (cd ui && npm run screenshot)" >&2
  exit 1
fi

mkdir -p "$BASELINE_DIR"
cp -f "$CURRENT_DIR"/*.png "$BASELINE_DIR"/
echo "baseline 갱신: $BASELINE_DIR"
ls -1 "$BASELINE_DIR"/*.png 2>/dev/null | xargs -n1 basename

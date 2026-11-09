#!/bin/bash
# diff-self-review.sh — Stop hook
# 세션 종료 시 git diff --stat 출력 + 무관 변경 자기점검 메시지.

set -uo pipefail

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
cd "$PROJECT_DIR" || exit 0

STAT=$(git diff --stat HEAD 2>/dev/null)
if [ -z "$STAT" ]; then
  exit 0
fi

CHANGED=$(git diff --name-only HEAD 2>/dev/null | wc -l | tr -d ' ')

echo ""
echo "=== DIFF SELF-REVIEW (${CHANGED} files) ==="
printf '%s\n' "$STAT"
echo ""
echo "위 변경 중 작업 범위 외 파일이 있으면 무관 변경입니다."
echo "  - 의도된 변경만 남기고: git checkout -- <file>"
echo "  - 부분 되돌리기: git checkout -p <file>"
echo ""

exit 0

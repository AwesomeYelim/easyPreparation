#!/bin/bash
# ts-check.sh — PostToolUse(Edit|Write) hook
# TS/TSX 변경 시 ui 디렉토리에서 tsc --noEmit 실행. 에러 시 차단(exit 2).

set -uo pipefail

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
UI_DIR="$PROJECT_DIR/ui"

INPUT=$(cat)
FILE=$(printf '%s' "$INPUT" | python3 -c "import sys,json
try:
    d = json.load(sys.stdin)
    print(d.get('tool_input', {}).get('file_path', ''))
except Exception:
    print('')" 2>/dev/null || echo "")

[ -z "$FILE" ] && exit 0
FILE="${FILE//\\//}"

case "$FILE" in
  *ui/*.ts|*ui/*.tsx) ;;
  *) exit 0 ;;
esac

if [ ! -d "$UI_DIR/node_modules" ]; then
  exit 0
fi

OUTPUT=$(cd "$UI_DIR" && npx --no-install tsc --noEmit --pretty false 2>&1)
STATUS=$?

if [ "$STATUS" -ne 0 ]; then
  {
    echo ""
    echo "=== TS CHECK FAILED ==="
    printf '%s\n' "$OUTPUT" | head -20
    echo ""
    echo "ui/ 타입 에러. 위 파일/라인 수정 후 재시도."
    echo ""
  } >&2
  exit 2
fi

exit 0

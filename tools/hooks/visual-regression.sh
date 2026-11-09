#!/bin/bash
# visual-regression.sh — PostToolUse(Edit|Write) hook
# UI 파일 변경 시 스크린샷 캡처 후 baseline과 sha256 비교, 변경 감지된 뷰 보고.
# 블로킹 안 함 (알림만). 디바운스: VISUAL_DEBOUNCE_SEC (기본 20초).

set -uo pipefail

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
SCREENSHOT_BASE="$PROJECT_DIR/.claude/screenshots"
CURRENT_DIR="$SCREENSHOT_BASE/current"
BASELINE_DIR="$SCREENSHOT_BASE/baseline"
STATE_DIR="$PROJECT_DIR/.claude/state"
TS_FILE="$STATE_DIR/last-screenshot"
DEBOUNCE_SEC="${VISUAL_DEBOUNCE_SEC:-20}"

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
  *ui/*) ;;
  *) exit 0 ;;
esac

mkdir -p "$STATE_DIR"

NOW=$(date +%s)
if [ -f "$TS_FILE" ]; then
  LAST=$(cat "$TS_FILE" 2>/dev/null || echo 0)
  DIFF=$((NOW - LAST))
  if [ "$DIFF" -lt "$DEBOUNCE_SEC" ]; then
    exit 0
  fi
fi

if ! curl -sf -o /dev/null --max-time 2 http://localhost:3000 2>/dev/null; then
  echo "[visual-regression] dev 서버(:3000) 응답 없음 — skip" >&2
  exit 0
fi

mkdir -p "$CURRENT_DIR" "$BASELINE_DIR"
echo "$NOW" > "$TS_FILE"

if ! (cd "$PROJECT_DIR/ui" && SCREENSHOT_DIR="$CURRENT_DIR" node ../tools/ui-tests/ui-screenshot.js) >/dev/null 2>&1; then
  echo "[visual-regression] 캡처 실패 — skip" >&2
  exit 0
fi

DIFFS=()
NEW_VIEWS=()
for cur in "$CURRENT_DIR"/*.png; do
  [ -f "$cur" ] || continue
  name=$(basename "$cur")
  base="$BASELINE_DIR/$name"
  if [ ! -f "$base" ]; then
    NEW_VIEWS+=("$name")
    continue
  fi
  cur_hash=$(sha256sum "$cur" | awk '{print $1}')
  base_hash=$(sha256sum "$base" | awk '{print $1}')
  if [ "$cur_hash" != "$base_hash" ]; then
    DIFFS+=("$name")
  fi
done

if [ ${#DIFFS[@]} -gt 0 ] || [ ${#NEW_VIEWS[@]} -gt 0 ]; then
  {
    echo ""
    echo "=== VISUAL REGRESSION ==="
    if [ ${#DIFFS[@]} -gt 0 ]; then
      echo "변경 감지 (${#DIFFS[@]}개):"
      for d in "${DIFFS[@]}"; do
        echo "  - $d"
      done
    fi
    if [ ${#NEW_VIEWS[@]} -gt 0 ]; then
      echo "baseline 없음 (${#NEW_VIEWS[@]}개):"
      for n in "${NEW_VIEWS[@]}"; do
        echo "  - $n"
      done
    fi
    echo ""
    echo "의도된 변경이면: bash tools/hooks/update-baseline.sh"
    echo "수정 대상 외 뷰가 변경됐다면 무관 변경 — 되돌리세요."
    echo ""
  } >&2
fi

exit 0

#!/bin/bash
# easyPreparation 환경 초기 세팅
# SessionStart hook에서 자동 실행됨

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# 필수 디렉토리 생성
mkdir -p "$SCRIPT_DIR/output"
mkdir -p "$ROOT_DIR/data"

# 환경 체크
ERRORS=0

check() {
  local label="$1"
  local cmd="$2"
  local result
  result=$(eval "$cmd" 2>/dev/null)
  if [ -n "$result" ]; then
    echo "  OK  $label: $result"
  else
    echo "  --  $label: 없음"
    ERRORS=$((ERRORS + 1))
  fi
}

echo "=== easyPreparation 환경 체크 ==="
check "Go    " "go version 2>/dev/null | awk '{print \$3}'"
check "Node  " "node --version 2>/dev/null"
check "Make  " "mingw32-make --version 2>/dev/null | head -1 | awk '{print \$1,\$2,\$3}' || make --version 2>/dev/null | head -1"
check "Python" "python3 --version 2>/dev/null || python --version 2>/dev/null"

# Python venv 상태
if [ -d "$SCRIPT_DIR/.venv" ]; then
  echo "  OK  venv : $SCRIPT_DIR/.venv"
else
  echo "  --  venv : 없음 (tools 파이썬 스크립트 필요 시 생성)"
fi

if [ "$ERRORS" -eq 0 ]; then
  echo "=== 준비 완료 ==="
else
  echo "=== 경고: ${ERRORS}개 미설치 항목 ==="
fi

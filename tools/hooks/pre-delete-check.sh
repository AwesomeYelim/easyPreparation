#!/bin/bash
# pre-delete-check.sh — 파일/디렉토리 삭제 전 잔해 탐지
#
# 사용법: bash tools/hooks/pre-delete-check.sh <삭제할 파일 또는 키워드>
# 예시:   bash tools/hooks/pre-delete-check.sh "forPrint"
#         bash tools/hooks/pre-delete-check.sh "bible_info.json"
#         bash tools/hooks/pre-delete-check.sh "Frame.png"

set -euo pipefail

TARGET="${1:-}"
if [ -z "$TARGET" ]; then
  echo "사용법: $0 <파일명 또는 키워드>"
  exit 1
fi

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$ROOT"

echo ""
echo "=== 삭제 전 잔해 탐지: \"$TARGET\" ==="
echo ""

FOUND=0

# 1. 코드 참조 (Go, TS, JS)
echo "── 1. 코드 참조 ──"
REFS=$(grep -rn "$TARGET" --include="*.go" --include="*.ts" --include="*.tsx" --include="*.js" --include="*.jsx" . 2>/dev/null | grep -v node_modules | grep -v ".next/" | grep -v ".claude/worktrees" || true)
if [ -n "$REFS" ]; then
  echo "$REFS"
  FOUND=$((FOUND + 1))
else
  echo "  없음"
fi
echo ""

# 2. 파일/디렉토리 복사본
echo "── 2. 파일 시스템 검색 ──"
FILES=$(find . -name "*$(basename "$TARGET")*" -not -path "*/node_modules/*" -not -path "*/.next/*" -not -path "*/.git/*" -not -path "*/.claude/worktrees/*" 2>/dev/null || true)
if [ -n "$FILES" ]; then
  echo "$FILES"
  FOUND=$((FOUND + 1))
else
  echo "  없음"
fi
echo ""

# 3. Makefile 참조
echo "── 3. Makefile 참조 ──"
MK=$(grep -n "$TARGET" Makefile 2>/dev/null || true)
if [ -n "$MK" ]; then
  echo "$MK"
  FOUND=$((FOUND + 1))
else
  echo "  없음"
fi
echo ""

# 4. CI 워크플로우 참조
echo "── 4. CI 워크플로우 참조 ──"
CI=$(grep -rn "$TARGET" .github/workflows/ 2>/dev/null || true)
if [ -n "$CI" ]; then
  echo "$CI"
  FOUND=$((FOUND + 1))
else
  echo "  없음"
fi
echo ""

# 5. CLAUDE.md / 문서 참조
echo "── 5. 문서 참조 ──"
DOCS=$(grep -rn "$TARGET" --include="*.md" . 2>/dev/null | grep -v node_modules | grep -v ".claude/worktrees" || true)
if [ -n "$DOCS" ]; then
  echo "$DOCS"
  FOUND=$((FOUND + 1))
else
  echo "  없음"
fi
echo ""

# 6. .gitignore 참조
echo "── 6. .gitignore 참조 ──"
GI=$(grep -n "$TARGET" .gitignore 2>/dev/null || true)
if [ -n "$GI" ]; then
  echo "$GI"
  FOUND=$((FOUND + 1))
else
  echo "  없음"
fi
echo ""

# 7. 빈 디렉토리 탐지 (관련 경로)
echo "── 7. 빈 디렉토리 ──"
EMPTY=$(find . -type d -empty -not -path "*/.git/*" -not -path "*/node_modules/*" -not -path "*/.next/*" -not -path "*/.claude/worktrees/*" 2>/dev/null || true)
if [ -n "$EMPTY" ]; then
  echo "$EMPTY"
else
  echo "  없음"
fi
echo ""

# 결과
echo "=== 결과: ${FOUND}개 영역에서 발견 ==="
if [ $FOUND -eq 0 ]; then
  echo "깨끗합니다. 안전하게 삭제 가능."
else
  echo "위 항목을 확인한 뒤 삭제하세요."
fi

#!/bin/bash
# project-clean-check.sh — 프로젝트 전체 클린 상태 점검
#
# 사용법: bash tools/hooks/project-clean-check.sh

set -euo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$ROOT"

echo ""
echo "=== easyPreparation 클린 체크 ==="
echo ""

ISSUES=0

# 1. 빈 디렉토리
echo "── 1. 빈 디렉토리 ──"
EMPTY=$(find . -type d -empty \
  -not -path "*/.git/*" \
  -not -path "*/node_modules/*" \
  -not -path "*/.next/*" \
  -not -path "*/.claude/*" \
  -not -path "*/tools/.venv/*" \
  -not -path "./output*" \
  2>/dev/null || true)
if [ -n "$EMPTY" ]; then
  echo "$EMPTY"
  ISSUES=$((ISSUES + 1))
else
  echo "  없음"
fi
echo ""

# 2. .DS_Store
echo "── 2. .DS_Store 파일 ──"
DS=$(find . -name ".DS_Store" -not -path "*/.git/*" 2>/dev/null || true)
if [ -n "$DS" ]; then
  echo "$DS"
  ISSUES=$((ISSUES + 1))
else
  echo "  없음"
fi
echo ""

# 3. 동일 파일명 중복 (같은 이름이 3곳 이상)
echo "── 3. 중복 의심 파일 (동일 이름 3개+) ──"
find . -type f \
  -not -path "*/.git/*" \
  -not -path "*/node_modules/*" \
  -not -path "*/.next/*" \
  -not -path "*/.claude/worktrees/*" \
  2>/dev/null | xargs -I{} basename {} | sort | uniq -c | sort -rn | awk '$1 >= 3 && $2 !~ /\.(js|ts|tsx|css|map|json|svg|woff2?)$/ {print "  " $1 "x " $2}' | head -10
echo ""

# 4. 대용량 파일 (50MB+)
echo "── 4. 대용량 파일 (50MB+) ──"
find . -type f -size +50M \
  -not -path "*/.git/*" \
  -not -path "*/node_modules/*" \
  -not -path "*/.next/*" \
  2>/dev/null | while read f; do
  SIZE=$(du -sh "$f" | cut -f1)
  echo "  $SIZE  $f"
done
echo ""

# 5. .bak / .old / .tmp 잔해
echo "── 5. 임시/백업 파일 잔해 ──"
TEMPS=$(find . -type f \( -name "*.bak" -o -name "*.old" -o -name "*.tmp" -o -name "*~" \) \
  -not -path "*/.git/*" \
  -not -path "*/node_modules/*" \
  -not -path "*/.next/*" \
  2>/dev/null || true)
if [ -n "$TEMPS" ]; then
  echo "$TEMPS"
  ISSUES=$((ISSUES + 1))
else
  echo "  없음"
fi
echo ""

# 6. Go 미사용 import (go vet이 잡지만 한번 더)
echo "── 6. Go vet ──"
export PATH="/usr/local/go/bin:$PATH"
VET=$(go vet ./cmd/server/ ./internal/... 2>&1 || true)
if [ -n "$VET" ]; then
  echo "$VET"
  ISSUES=$((ISSUES + 1))
else
  echo "  통과"
fi
echo ""

# 7. gitignore에 있어야 하는데 추적 중인 파일
echo "── 7. git 추적 중인 output/bin 파일 ──"
TRACKED=$(git ls-files -- "output/" "bin/" "*.exe" 2>/dev/null || true)
if [ -n "$TRACKED" ]; then
  echo "$TRACKED"
  ISSUES=$((ISSUES + 1))
else
  echo "  없음"
fi
echo ""

# 결과
echo "=== 결과: ${ISSUES}개 이슈 발견 ==="
if [ $ISSUES -eq 0 ]; then
  echo "프로젝트가 깨끗합니다."
else
  echo "위 항목을 정리하세요."
fi

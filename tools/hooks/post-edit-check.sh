#!/bin/bash
# post-edit-check.sh — 편집 직후 알려진 안티패턴 자동 검출
# PostToolUse 훅에서 호출됨. 종료코드 1 = 안티패턴 발견 (블로킹).
#
# 사용법: bash tools/hooks/post-edit-check.sh <파일경로>
# 환경변수 TOOL_INPUT (JSON) 에서 file_path를 추출하거나 인자로 받음.

set -euo pipefail

# 파일 경로 결정 — 인자 또는 stdin JSON
if [ -n "${1:-}" ]; then
  FILE="$1"
else
  FILE=$(cat | python3 -c "import sys,json; print(json.load(sys.stdin).get('file_path',''))" 2>/dev/null || echo "")
fi

[ -z "$FILE" ] && exit 0
[ ! -f "$FILE" ] && exit 0

ERRORS=0
WARNINGS=""

add_error() {
  ERRORS=$((ERRORS + 1))
  WARNINGS="${WARNINGS}\n  $1"
}

# ── Go 파일 ──────────────────────────────────────────
if [[ "$FILE" == *.go ]]; then

  # log.Fatalf → 서버 크래시
  if grep -n 'log\.Fatalf' "$FILE" 2>/dev/null | head -3 | grep -q .; then
    add_error "$(grep -n 'log\.Fatalf' "$FILE" | head -1) — log.Fatalf 사용 금지 (return fmt.Errorf 사용)"
  fi

  # 하드코딩 절대경로 (Ghostscript 예외)
  if grep -n '"/usr/local/' "$FILE" 2>/dev/null | grep -v 'go/bin' | head -1 | grep -q .; then
    add_error "$(grep -n '"/usr/local/' "$FILE" | grep -v 'go/bin' | head -1) — 하드코딩 경로 금지"
  fi

fi

# ── TypeScript/React 파일 ────────────────────────────
if [[ "$FILE" == *.ts || "$FILE" == *.tsx ]]; then

  # useState(useRecoilValue(...)) → 탭 전환 시 데이터 유실
  if grep -n 'useState.*useRecoilValue' "$FILE" 2>/dev/null | head -1 | grep -q .; then
    add_error "$(grep -n 'useState.*useRecoilValue' "$FILE" | head -1) — useState로 Recoil 값 복사 금지 (useRecoilState 직접 사용)"
  fi

  # useState(recoilValue) 패턴 (변수명으로 전달하는 경우)
  if grep -n 'useState(.*[Aa]tom\|useState(.*[Ss]tate)' "$FILE" 2>/dev/null | grep -v 'RecoilState\|recoilState\.ts' | head -1 | grep -q .; then
    MATCH=$(grep -n 'useState(.*[Aa]tom\|useState(.*[Ss]tate)' "$FILE" | grep -v 'RecoilState\|recoilState\.ts' | head -1)
    add_error "$MATCH — useState에 Recoil atom/state를 초기값으로 넣지 마세요 (의심 패턴)"
  fi

  # fetch('/api/...') — Go API인데 상대경로로 호출 (Next.js로 감)
  # BASE_URL 없이 /api/worship-order, /api/display-config 등 Go 전용 경로를 호출하는지 확인
  GO_ROUTES="worship-order\|display-config\|display/navigate\|display/jump\|display/order\|display/status\|assets/\|schedule\|license\|update\|video-bg\|logo\|obs\|submit"
  if grep -n "fetch.*['\"][^B]*\/api\/\(${GO_ROUTES}\)" "$FILE" 2>/dev/null | grep -v 'BASE_URL\|baseUrl\|baseURL' | head -1 | grep -q .; then
    MATCH=$(grep -n "fetch.*['\"][^B]*\/api\/\(${GO_ROUTES}\)" "$FILE" | grep -v 'BASE_URL\|baseUrl\|baseURL' | head -1)
    add_error "$MATCH — Go API는 \${BASE_URL}/api/... 사용 필수 (상대경로는 Next.js로 감)"
  fi

fi

# ── 결과 출력 ────────────────────────────────────────
if [ $ERRORS -gt 0 ]; then
  echo ""
  echo "=== ANTI-PATTERN DETECTED ($ERRORS) in $(basename "$FILE") ==="
  echo -e "$WARNINGS"
  echo ""
  echo "위 패턴은 알려진 반복 버그입니다. 수정 후 다시 저장하세요."
  echo "================================================================"
  exit 1
fi

exit 0

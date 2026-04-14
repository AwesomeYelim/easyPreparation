#!/usr/bin/env bash
# tools/sync-ai-supporter.sh
# ai_supporter 레포에서 공유 툴링 파일을 easyPreparation에 동기화합니다.
#
# ai_supporter 위치 탐색 순서:
#   1. 환경변수: AI_SUPPORTER_PATH
#   2. 로컬 설정: .ai-supporter-path 파일
#   3. 형제 폴더: ../ai_supporter

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
AI_PATH_FILE="$ROOT_DIR/.ai-supporter-path"

# ── ai_supporter 경로 탐색 ─────────────────────────────────────────────────────
find_supporter() {
  # 1. 환경변수
  if [ -n "${AI_SUPPORTER_PATH:-}" ] && [ -d "$AI_SUPPORTER_PATH/.git" ]; then
    echo "$AI_SUPPORTER_PATH"
    return
  fi
  # 2. 로컬 설정 파일
  if [ -f "$AI_PATH_FILE" ]; then
    local p
    p=$(cat "$AI_PATH_FILE" | tr -d '[:space:]')
    if [ -d "$p/.git" ]; then
      echo "$p"
      return
    fi
  fi
  # 3. 형제 폴더
  local sibling
  sibling="$(dirname "$ROOT_DIR")/ai_supporter"
  if [ -d "$sibling/.git" ]; then
    echo "$sibling"
    return
  fi
  echo ""
}

SUPPORTER=$(find_supporter)

if [ -z "$SUPPORTER" ]; then
  echo "ai_supporter: 로컬 경로 없음 — 동기화 스킵"
  echo "  설정하려면: echo '/path/to/ai_supporter' > .ai-supporter-path"
  echo "  또는: export AI_SUPPORTER_PATH=/path/to/ai_supporter"
  exit 0
fi

echo "ai_supporter: $SUPPORTER"

# ── ai_supporter 최신화 ────────────────────────────────────────────────────────
echo "  git pull 중..."
git -C "$SUPPORTER" pull --ff-only --quiet 2>&1 || {
  echo "  WARN: git pull 실패 — 로컬 변경사항 있거나 네트워크 오류. 기존 버전으로 동기화 진행."
}

# ── 동기화 대상 파일 목록 ─────────────────────────────────────────────────────
# key: ai_supporter 내 경로 (relative)
# value: easyPreparation 내 경로 (relative)
declare -A SYNC_MAP
SYNC_MAP=(
  # indexer 바이너리 (OS별)
  [".claude/bin/indexer-darwin-amd64"]=".claude/bin/indexer-darwin-amd64"
  [".claude/bin/indexer-darwin-arm64"]=".claude/bin/indexer-darwin-arm64"
  [".claude/bin/indexer-linux-amd64"]=".claude/bin/indexer-linux-amd64"
  [".claude/bin/indexer.exe"]=".claude/bin/indexer.exe"
  # Python MCP 툴링
  ["tools/requirements.txt"]="tools/requirements.txt"
  ["tools/run.sh"]="tools/run.sh"
  ["tools/start_mcp.sh"]="tools/start_mcp.sh"
  ["tools/auth.py"]="tools/auth.py"
  ["tools/gdrive_mcp.py"]="tools/gdrive_mcp.py"
)

UPDATED=0
SKIPPED=0

for src_rel in "${!SYNC_MAP[@]}"; do
  dst_rel="${SYNC_MAP[$src_rel]}"
  src="$SUPPORTER/$src_rel"
  dst="$ROOT_DIR/$dst_rel"

  [ -f "$src" ] || { echo "  SKIP $src_rel (소스 없음)"; SKIPPED=$((SKIPPED+1)); continue; }

  # 변경 여부 확인 (바이너리 포함 안전하게 비교)
  if [ -f "$dst" ] && cmp -s "$src" "$dst"; then
    continue  # 동일 — 스킵
  fi

  mkdir -p "$(dirname "$dst")"
  cp "$src" "$dst"
  echo "  SYNC $src_rel"
  UPDATED=$((UPDATED+1))
done

# ── scripts/ 디렉토리 동기화 ──────────────────────────────────────────────────
SCRIPTS_SRC="$SUPPORTER/tools/scripts"
SCRIPTS_DST="$ROOT_DIR/tools/scripts"

if [ -d "$SCRIPTS_SRC" ]; then
  mkdir -p "$SCRIPTS_DST"
  while IFS= read -r -d '' f; do
    rel="${f#$SCRIPTS_SRC/}"
    dst_f="$SCRIPTS_DST/$rel"
    if ! [ -f "$dst_f" ] || ! cmp -s "$f" "$dst_f"; then
      cp "$f" "$dst_f"
      echo "  SYNC tools/scripts/$rel"
      UPDATED=$((UPDATED+1))
    fi
  done < <(find "$SCRIPTS_SRC" -type f -print0)
fi

# ── .gitkeep 생성 ─────────────────────────────────────────────────────────────
for d in "tools/output" "tools/templates"; do
  [ -f "$ROOT_DIR/$d/.gitkeep" ] || touch "$ROOT_DIR/$d/.gitkeep"
done

# ── 결과 보고 ─────────────────────────────────────────────────────────────────
echo ""
if [ "$UPDATED" -gt 0 ]; then
  echo "동기화 완료: ${UPDATED}개 업데이트, ${SKIPPED}개 스킵"
else
  echo "이미 최신 상태 (업데이트 없음)"
fi

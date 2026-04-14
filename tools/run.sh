#!/usr/bin/env bash
# tools/run.sh — ai-tools 통합 진입점
#
# 사용법:
#   tools/run.sh upload <파일> [--folder ID] [--name 이름] [--key 키]
#   tools/run.sh analyze <모듈> [함수]
#   tools/run.sh token status|refresh
#   tools/run.sh server status
#   tools/run.sh pack [--dry-run] [--host user@host:/path]
#   tools/run.sh install [--local FILE] [--host HOST] [--path PATH] [--no-setup]

set -euo pipefail

TOOLS_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(cd "$TOOLS_DIR/.." && pwd)"
INVOKE_DIR="$(pwd)"   # install 서브커맨드가 설치 대상으로 사용

DEFAULT_HOST="root@192.168.1.10"
DEFAULT_PATH="/home/clouddev/tool"
TARBALL="ai-tools.tar.gz"

# ── Python 감지 (pack/install 은 불필요하므로 non-fatal) ───────────────────
PYTHON=""
if [ -x "$TOOLS_DIR/.venv/Scripts/python.exe" ]; then
    PYTHON="$TOOLS_DIR/.venv/Scripts/python.exe"
elif [ -x "$TOOLS_DIR/.venv/bin/python" ]; then
    PYTHON="$TOOLS_DIR/.venv/bin/python"
fi
export PYTHONUTF8=1

# Windows MINGW: Python용 Windows 경로 별도 보관 (bash/tar는 POSIX 경로 유지)
REPO_DIR_PY="$REPO_DIR"
TOOLS_DIR_PY="$TOOLS_DIR"
if command -v cygpath &>/dev/null; then
    REPO_DIR_PY=$(cygpath -w "$REPO_DIR")
    TOOLS_DIR_PY=$(cygpath -w "$TOOLS_DIR")
    cd "$REPO_DIR_PY"
fi

# ── indexer 경로 감지 ─────────────────────────────────────────────────────
_indexer() {
    local BIN
    case "$(uname -s)-$(uname -m)" in
        Linux-x86_64)         BIN="$REPO_DIR/.claude/bin/indexer-linux-amd64" ;;
        Darwin-x86_64)        BIN="$REPO_DIR/.claude/bin/indexer-darwin-amd64" ;;
        Darwin-arm64)         BIN="$REPO_DIR/.claude/bin/indexer-darwin-arm64" ;;
        MINGW*|MSYS*|CYGWIN*) BIN="$REPO_DIR/.claude/bin/indexer.exe" ;;
        *)                    BIN="$REPO_DIR/.claude/bin/indexer.exe" ;;
    esac
    [ -x "$BIN" ] || { echo "❌ indexer 없음: $BIN" >&2; exit 1; }
    "$BIN" "$@"
}

# ── 서브커맨드 ────────────────────────────────────────────────────────────
SUBCMD="${1:-}"
shift || true

case "$SUBCMD" in

  # ── report ──────────────────────────────────────────────────────────────
  report)
    [ -n "$PYTHON" ] || { echo "❌ venv 없음. tools/setup.sh 를 먼저 실행하세요." >&2; exit 1; }
    REPORT_DIR="$INVOKE_DIR"
    REPORT_OUT="$TOOLS_DIR/output/report.pptx"
    REPORT_NAME=""
    REPORT_FOLDER=""
    REPORT_KEY=""
    REPORT_TYPE="code"
    REPORT_DATA=""
    while [[ $# -gt 0 ]]; do
      case "$1" in
        --out)    REPORT_OUT="$2";    shift ;;
        --folder) REPORT_FOLDER="$2"; shift ;;
        --key)    REPORT_KEY="$2";    shift ;;
        --name)   REPORT_NAME="$2";   shift ;;
        --dir)    REPORT_DIR="$2";    shift ;;
        --type)   REPORT_TYPE="$2";   shift ;;
        --data)   REPORT_DATA="$2";   shift ;;
        --list-types)
          PYTHONUTF8=1 "$PYTHON" "$TOOLS_DIR_PY/scripts/generate_report.py" --list-types
          exit 0 ;;
        *) echo "❌ 알 수 없는 옵션: $1" >&2; exit 1 ;;
      esac; shift
    done

    if command -v cygpath &>/dev/null; then
      REPORT_DIR_PY=$(cygpath -w "$REPORT_DIR")
      REPORT_OUT_PY=$(cygpath -w "$REPORT_OUT")
    else
      REPORT_DIR_PY="$REPORT_DIR"
      REPORT_OUT_PY="$REPORT_OUT"
    fi

    if command -v cygpath &>/dev/null && [ -n "$REPORT_DATA" ]; then
      REPORT_DATA_PY=$(cygpath -w "$REPORT_DATA")
    else
      REPORT_DATA_PY="$REPORT_DATA"
    fi

    echo "📊 [$REPORT_TYPE] 리포트 생성 중: $REPORT_DIR"
    PYTHONUTF8=1 "$PYTHON" "$TOOLS_DIR_PY/scripts/generate_report.py" \
      --project-dir "$REPORT_DIR_PY" \
      --out "$REPORT_OUT_PY" \
      --type "$REPORT_TYPE" \
      ${REPORT_NAME:+--name "$REPORT_NAME"} \
      ${REPORT_DATA_PY:+--data "$REPORT_DATA_PY"}

    if [ -n "$REPORT_FOLDER" ] || [ -n "$REPORT_KEY" ]; then
      echo "📤 Drive 업로드..."
      "$PYTHON" "$TOOLS_DIR_PY/scripts/upload_sheet.py" "$REPORT_OUT_PY" \
        ${REPORT_FOLDER:+--folder "$REPORT_FOLDER"} \
        ${REPORT_KEY:+--key "$REPORT_KEY"} \
        ${REPORT_NAME:+--name "$REPORT_NAME"}
    fi
    ;;

  # ── upload ──────────────────────────────────────────────────────────────
  upload)
    [ -n "$PYTHON" ] || { echo "❌ venv 없음. tools/setup.sh 를 먼저 실행하세요." >&2; exit 1; }
    [ $# -gt 0 ]    || { echo "Usage: tools/run.sh upload <파일> [--folder ID] [--name 이름] [--key 키]" >&2; exit 1; }
    "$PYTHON" "$TOOLS_DIR_PY/scripts/upload_sheet.py" "$@"
    ;;

  # ── analyze ─────────────────────────────────────────────────────────────
  analyze)
    MODULE="${1:-}"
    FUNC="${2:-}"
    [ -n "$MODULE" ] || { echo "Usage: tools/run.sh analyze <모듈> [함수]" >&2; exit 1; }
    shift; shift 2>/dev/null || true

    if [ -z "$FUNC" ]; then
        echo "=== [$MODULE] 통계 ===" && _indexer -m "$MODULE" -stats 2>/dev/null
        echo && echo "=== [$MODULE] 함수 목록 ===" && _indexer -m "$MODULE" -list 2>/dev/null
    else
        echo "=== [$MODULE::$FUNC] 위치 ===" && _indexer -m "$MODULE" -f "$FUNC" -l 2>/dev/null
        echo && echo "=== [$MODULE::$FUNC] 호출 그래프 ===" && _indexer -m "$MODULE" -f "$FUNC" -calls 2>/dev/null
        echo && echo "=== [$MODULE::$FUNC] 역방향 호출 ===" && _indexer -m "$MODULE" -f "$FUNC" -callers 2>/dev/null
    fi
    ;;

  # ── token ────────────────────────────────────────────────────────────────
  token)
    [ -n "$PYTHON" ] || { echo "❌ venv 없음. tools/setup.sh 를 먼저 실행하세요." >&2; exit 1; }
    ACTION="${1:-status}"
    case "$ACTION" in
      status)
        "$PYTHON" - <<'PYEOF'
import json, time, sys
try:
    with open("tools/output/.gtoken", "r") as f:
        d = json.load(f)
except FileNotFoundError:
    print("❌ .gtoken 없음 — tools/run.sh token refresh 로 인증하세요"); sys.exit(1)
remaining = int((d.get("expiry_date", 0) - time.time() * 1000) / 1000)
if remaining > 300:
    print(f"✅ 토큰 유효 (남은 시간: {remaining//60}분)")
else:
    print("⚠️  토큰 만료됨 (tools/run.sh token refresh 로 갱신)")
PYEOF
        ;;
      refresh)
        "$PYTHON" - <<'PYEOF'
import sys
sys.path.insert(0, "tools/scripts")
from upload_sheet import get_access_token
try:
    get_access_token(force=True); print("✅ 토큰 갱신 완료")
except Exception as e:
    print(f"❌ {e}", file=sys.stderr)
    print("→ 재인증 필요. 아래 절차를 따르세요:", file=sys.stderr)
    sys.exit(1)
PYEOF
        if [ $? -ne 0 ]; then
            echo ""
            echo "  1) 브라우저 인증 URL 출력:"
            echo "       $PYTHON $TOOLS_DIR_PY/auth.py --url"
            echo ""
            echo "  2) 브라우저에서 인증 후 리다이렉트 URL 붙여넣기:"
            echo "       $PYTHON $TOOLS_DIR_PY/auth.py --paste"
        fi
        ;;
      *) echo "Usage: tools/run.sh token <status|refresh>" >&2; exit 1 ;;
    esac
    ;;

  # ── server ───────────────────────────────────────────────────────────────
  server)
    ACTION="${1:-status}"
    case "$ACTION" in
      status)
        OUT=$(ssh -o BatchMode=yes -o ConnectTimeout=5 "$DEFAULT_HOST" \
            "stat -c '%s %Y' $DEFAULT_PATH/$TARBALL" 2>&1) || {
            echo "❌ 서버 연결 실패: $OUT" >&2; exit 1
        }
        SIZE_KB=$(( $(echo "$OUT" | awk '{print $1}') / 1024 ))
        MTIME=$(echo "$OUT" | awk '{print $2}')
        MTIME_FMT=$(date -d "@$MTIME" "+%Y-%m-%d %H:%M:%S" 2>/dev/null \
                    || date -r "$MTIME" "+%Y-%m-%d %H:%M:%S" 2>/dev/null \
                    || echo "$MTIME")
        echo "✅ 서버 파일 확인됨"
        echo "   크기: ${SIZE_KB} KB"
        echo "   수정: $MTIME_FMT"
        ;;
      *) echo "Usage: tools/run.sh server status" >&2; exit 1 ;;
    esac
    ;;

  # ── pack ─────────────────────────────────────────────────────────────────
  pack)
    DRY_RUN=false
    REMOTE_HOST="$DEFAULT_HOST:$DEFAULT_PATH/"
    while [[ $# -gt 0 ]]; do
      case "$1" in
        --dry-run) DRY_RUN=true ;;
        --host) REMOTE_HOST="$2"; shift ;;
        *) echo "❌ 알 수 없는 옵션: $1" >&2; exit 1 ;;
      esac; shift
    done

    OUT="$TOOLS_DIR/output/$TARBALL"
    mkdir -p "$TOOLS_DIR/output"

    echo "📦 패키징 중..."
    TMP_DIR=$(mktemp -d); trap "rm -rf $TMP_DIR" EXIT
    PKG="$TMP_DIR/ai-tools"
    mkdir -p "$PKG/.claude/bin" "$PKG/tools/scripts" "$PKG/tools/output" "$PKG/tools/templates"

    for f in indexer-linux-amd64 indexer-darwin-amd64 indexer-darwin-arm64 indexer.exe config.json; do
        src="$REPO_DIR/.claude/bin/$f"; [ -f "$src" ] && cp "$src" "$PKG/.claude/bin/"
    done
    for f in generate_slide.py generate_docx.py generate_sheet.py upload_sheet.py registry.py generate_report.py; do
        src="$TOOLS_DIR/scripts/$f"; [ -f "$src" ] && cp "$src" "$PKG/tools/scripts/"
    done
    for f in gdrive_mcp.py auth.py run.sh setup.sh start_mcp.sh requirements.txt README.md CLAUDE.md; do
        src="$TOOLS_DIR/$f"; [ -f "$src" ] && cp "$src" "$PKG/tools/"
    done
    [ -d "$TOOLS_DIR/templates" ] && cp "$TOOLS_DIR/templates/"* "$PKG/tools/templates/" 2>/dev/null || true

    cp "$REPO_DIR/.mcp.json" "$PKG/" 2>/dev/null || true
    for f in AGENTS.md GEMINI.md; do
        src="$REPO_DIR/$f"; [ -f "$src" ] && cp "$src" "$PKG/tools/"
    done

    # CLAUDE.md 템플릿 (프로젝트 고유 내용 제거)
    cat > "$PKG/CLAUDE.md" << 'TMPL'
# CLAUDE.md

## 코드 분석 시 indexer 우선 사용

코드베이스 탐색/분석 시 **반드시** `.claude/bin/indexer`를 먼저 사용할 것.

### 바이너리 선택
- Linux: `.claude/bin/indexer-linux-amd64`  / macOS: `indexer-darwin-arm64` / Windows: `indexer.exe`

### 주요 플래그
| 플래그 | 설명 | 예시 |
|--------|------|------|
| `-m` | 모듈 | `-m myModule -stats` |
| `-list` | 함수 목록 | `-m myModule -list` |
| `-f -l` | 함수 위치 | `-m myModule -f Run -l` |
| `-calls` / `-callers` | 호출 그래프 | `-m myModule -f Run -calls` |
| `-impact` | 변경 영향 | `-m myModule -f Run -impact` |

### 사용 원칙
1. `-stats` → `-list` → 필요한 함수 `-calls`/`-callers` 순서
2. indexer로 구조 파악 후 상세 코드만 Read 사용

---

## 문서 생성 / Drive 업로드

`tools/run.sh` 를 사용할 것. 자세한 내용: `tools/AGENTS.md`

```bash
tools/run.sh upload tools/output/result.pptx --folder <폴더ID> --key my_key
tools/run.sh analyze <모듈>
tools/run.sh token status
```
TMPL

    # .gitignore 추가 항목
    cat > "$PKG/.gitignore-tools-append" << 'EOF'
# ── ai-tools — setup.sh / run.sh 이 생성·다운로드하는 모든 파일 ──
tools/.venv/
tools/output/*
!tools/output/.gitkeep
tools/templates/
tools/**/__pycache__/
.ai-tools-backup-*/
# ─────────────────────────────────────────────────────────
EOF

    chmod +x "$PKG/.claude/bin/"* 2>/dev/null || true
    chmod +x "$PKG/tools/run.sh" "$PKG/tools/setup.sh" "$PKG/tools/start_mcp.sh" 2>/dev/null || true

    (cd "$TMP_DIR" && tar czf "$OUT" ai-tools/)
    echo "✅ 패키지 생성: $OUT ($(du -sh "$OUT" | cut -f1))"
    echo "포함 파일:"; tar tzf "$OUT" | grep -v '/$' | sort | sed 's/^/  /'

    if [ "$DRY_RUN" = true ]; then
        echo; echo "ℹ️  --dry-run: 업로드 생략 / scp $OUT $REMOTE_HOST"
    else
        echo; echo "📤 업로드: $REMOTE_HOST"
        scp "$OUT" "$REMOTE_HOST"
        echo "✅ 업로드 완료"
    fi
    ;;

  # ── install ──────────────────────────────────────────────────────────────
  install)
    REMOTE_HOST="$DEFAULT_HOST"
    REMOTE_PATH="$DEFAULT_PATH"
    LOCAL_FILE=""
    RUN_SETUP=true
    while [[ $# -gt 0 ]]; do
      case "$1" in
        --host)     REMOTE_HOST="$2"; shift ;;
        --path)     REMOTE_PATH="$2"; shift ;;
        --local)    LOCAL_FILE="$2"; shift ;;
        --no-setup) RUN_SETUP=false ;;
        *) echo "❌ 알 수 없는 옵션: $1" >&2; exit 1 ;;
      esac; shift
    done

    PROJECT_DIR="$INVOKE_DIR"
    TMP_TAR="$(mktemp /tmp/ai-tools-XXXXXX.tar.gz)"
    trap "rm -f $TMP_TAR" EXIT

    if [ -n "$LOCAL_FILE" ]; then
        echo "📦 로컬 파일 사용: $LOCAL_FILE"
        cp "$LOCAL_FILE" "$TMP_TAR"
    else
        HTTP_URL="http://${REMOTE_HOST##*@}/$TARBALL"
        echo "⬇️  HTTP 시도: $HTTP_URL"
        if command -v curl &>/dev/null && curl -fsSL --connect-timeout 5 -o "$TMP_TAR" "$HTTP_URL" 2>/dev/null; then
            echo "✅ HTTP 다운로드 완료"
        else
            echo "⚠️  HTTP 실패 → SCP..."
            scp "$REMOTE_HOST:$REMOTE_PATH/$TARBALL" "$TMP_TAR"
            echo "✅ SCP 완료"
        fi
    fi

    # CLAUDE.md 기존 내용 저장 (병합용 — 덮어쓰기 대신 추가)
    ORIG_CLAUDE=""
    [ -f "$PROJECT_DIR/CLAUDE.md" ] && ORIG_CLAUDE=$(cat "$PROJECT_DIR/CLAUDE.md")

    # 충돌 파일 백업 (CLAUDE.md 제외 — 별도 병합 처리)
    CONFLICTS=()
    for f in .mcp.json tools/AGENTS.md tools/GEMINI.md; do
        [ -f "$PROJECT_DIR/$f" ] && CONFLICTS+=("$f")
    done
    if [ ${#CONFLICTS[@]} -gt 0 ]; then
        echo "⚠️  기존 파일: ${CONFLICTS[*]}"
        read -r -p "   덮어쓰기? [y/N] " REPLY
        if [[ ! "$REPLY" =~ ^[Yy]$ ]]; then
            BDIR="$PROJECT_DIR/.ai-tools-backup-$(date +%Y%m%d%H%M%S)"
            mkdir -p "$BDIR"
            for f in "${CONFLICTS[@]}"; do cp "$PROJECT_DIR/$f" "$BDIR/"; done
            echo "   백업: $BDIR"
        fi
    fi

    echo "📦 압축 해제..."
    tar xzf "$TMP_TAR" --strip-components=1 -C "$PROJECT_DIR"

    # CLAUDE.md 병합: 기존 내용 유지 + indexer 섹션 추가
    if [ -n "$ORIG_CLAUDE" ]; then
        NEW_SECTION=$(cat "$PROJECT_DIR/CLAUDE.md")
        if echo "$ORIG_CLAUDE" | grep -q "indexer 우선 사용"; then
            # 이미 섹션 존재 → 기존 내용 복원
            echo "$ORIG_CLAUDE" > "$PROJECT_DIR/CLAUDE.md"
            echo "ℹ️  CLAUDE.md: indexer 섹션 이미 존재 → 기존 내용 유지"
        else
            # 기존 내용 + 구분선 + 새 섹션 병합
            printf '%s\n\n---\n\n%s\n' "$ORIG_CLAUDE" "$NEW_SECTION" > "$PROJECT_DIR/CLAUDE.md"
            echo "✅ CLAUDE.md: 기존 내용 보존 + indexer 섹션 추가"
        fi
    fi
    chmod +x "$PROJECT_DIR/.claude/bin/"* 2>/dev/null || true
    chmod +x "$PROJECT_DIR/tools/run.sh" "$PROJECT_DIR/tools/setup.sh" \
             "$PROJECT_DIR/tools/start_mcp.sh" 2>/dev/null || true

    # .gitignore 병합 — 항목별로 없는 것만 추가 (중복 방지)
    APPEND="$PROJECT_DIR/.gitignore-tools-append"
    if [ -f "$APPEND" ]; then
        GI="$PROJECT_DIR/.gitignore"
        [ -f "$GI" ] || touch "$GI"
        ADDED=0
        while IFS= read -r line; do
            # 빈 줄·주석은 기존 gsac-tools 섹션 없을 때만 추가
            [[ "$line" =~ ^#.*ai-tools ]] && grep -q "ai-tools" "$GI" && continue
            [[ -z "$line" || "$line" =~ ^# ]] && continue
            grep -qF "$line" "$GI" || { echo "$line" >> "$GI"; ADDED=$((ADDED+1)); }
        done < "$APPEND"
        if [ "$ADDED" -gt 0 ]; then
            echo "✅ .gitignore: $ADDED개 항목 추가"
        else
            echo "ℹ️  .gitignore: 이미 최신"
        fi
        rm "$APPEND"
    fi

    [ "$RUN_SETUP" = true ] && bash "$PROJECT_DIR/tools/setup.sh"

    echo; echo "✅ 설치 완료!"
    echo "  다음: tools/run.sh token refresh  →  tools/run.sh token status"
    ;;

  # ── clean ────────────────────────────────────────────────────────────────
  clean)
    echo "🧹 설치된 파일 삭제 중..."
    # setup.sh 이 생성하는 파일
    rm -rf "$TOOLS_DIR/.venv"                        && echo "  ✓ tools/.venv/"
    rm -rf "$TOOLS_DIR/templates"                    && echo "  ✓ tools/templates/"
    PYCACHE_COUNT=$(find "$TOOLS_DIR" -type d -name "__pycache__" | wc -l)
    find "$TOOLS_DIR" -type d -name "__pycache__" -prune -exec rm -rf {} + 2>/dev/null
    [ "$PYCACHE_COUNT" -gt 0 ] && echo "  ✓ __pycache__/ (${PYCACHE_COUNT}개)" || echo "  - __pycache__/ 없음"
    find "$TOOLS_DIR/output" -mindepth 1 ! -name ".gitkeep" -delete 2>/dev/null; echo "  ✓ tools/output/ (내용)"
    # tarball 이 설치하는 파일
    rm -rf "$REPO_DIR/.claude/bin"                   && echo "  ✓ .claude/bin/"
    rm -f  "$REPO_DIR/.mcp.json"                     && echo "  ✓ .mcp.json"
    rm -f  "$TOOLS_DIR/AGENTS.md" "$TOOLS_DIR/GEMINI.md" && echo "  ✓ tools/AGENTS.md / GEMINI.md"
    # tools/ 자체 (마지막에 삭제 — 실행 중인 스크립트 포함)
    rm -rf "$TOOLS_DIR"                              && echo "  ✓ tools/"
    echo "✅ 완료 — 재설치: tools/run.sh install  (venv·템플릿 포함 전체 재설치)"
    ;;

  # ── help ─────────────────────────────────────────────────────────────────
  help|--help|-h|"")
    cat <<'EOF'
tools/run.sh — ai-tools 통합 진입점

  report [--type 타입] [--data JSON파일] [--dir 경로] [--out 파일] [--folder ID] [--key 키]
      리포트 자동 생성 → 슬라이드 [→ Drive 업로드]
      타입: code(기본) | release | test | vuln | interface | design | wbs
      --list-types  : 타입별 설명 + JSON 스키마 출력
      예) tools/run.sh report --type release --folder ID --key v4_rel
          tools/run.sh report --type vuln --data findings.json --folder ID

  upload <파일> [--folder ID] [--name 이름] [--key 키]
      Drive 업로드 (토큰 자동 갱신)

  analyze <모듈> [함수]
      모듈만: stats + 함수목록 / 모듈+함수: 위치·호출그래프·역방향호출

  token status|refresh
      토큰 유효기간 확인 / 강제 갱신

  server status
      1.10 서버의 ai-tools.tar.gz 업로드 상태 확인

  pack [--dry-run] [--host user@host:/path]
      tarball 빌드 후 서버 업로드 (기본: root@192.168.1.10:/home/clouddev/tool/)

  install [--local FILE] [--host HOST] [--path PATH] [--no-setup]
      서버에서 tarball 받아 현재 디렉터리에 설치

  clean
      설치된 파일 일괄 삭제 (.venv / templates/ / output/ / __pycache__)

예시:
  tools/run.sh upload tools/output/report.pptx --folder 1F-VUZ... --key q3
  tools/run.sh analyze AD Ping
  tools/run.sh token refresh
  tools/run.sh server status
  tools/run.sh pack
  tools/run.sh pack --dry-run
  tools/run.sh install
  tools/run.sh install --local ai-tools.tar.gz
EOF
    ;;

  *)
    echo "❌ 알 수 없는 서브커맨드: $SUBCMD  (tools/run.sh help)" >&2; exit 1
    ;;
esac

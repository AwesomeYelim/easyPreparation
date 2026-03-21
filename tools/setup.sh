#!/usr/bin/env bash
# tools/setup.sh — 초기 환경 세팅 (Linux/macOS/Windows)
#
# 역할:
#   1. Python venv 생성 및 패키지 설치
#   2. Google OAuth 토큰(.gtoken) 상태 확인
#   3. 문서 템플릿 다운로드
#
# 실행: bash tools/setup.sh
# SessionStart 훅에서도 자동 실행됨

set -euo pipefail

TOOLS_DIR="$(cd "$(dirname "$0")" && pwd)"
VENV_DIR="$TOOLS_DIR/.venv"
GTOKEN="$TOOLS_DIR/output/.gtoken"

ok()   { echo "  ✓ $*"; }
warn() { echo "  ! $*"; }
step() { echo ""; echo "[$1] $2"; }

# ── 1. Python venv + 패키지 ─────────────────────────────────────
step 1 "Python venv (tools/.venv)"
PYTHON=$(command -v python3 2>/dev/null || command -v python 2>/dev/null || echo "")
if [ -z "$PYTHON" ]; then
    echo "  ✗ Python 3를 찾을 수 없습니다."
    exit 1
fi

# 최소 Python 3.10 필요 (mcp 패키지 요구사항)
MIN_VER="3.10"
SYS_VER=$("$PYTHON" -c "import sys; print('%d.%d' % sys.version_info[:2])" 2>/dev/null || echo "")
if [ -n "$SYS_VER" ] && ! "$PYTHON" -c "import sys; exit(0 if sys.version_info >= (3,10) else 1)" 2>/dev/null; then
    echo "  ✗ Python $SYS_VER — 최소 $MIN_VER 이상 필요 (mcp 패키지)"
    exit 1
fi

# Windows: venv은 Scripts/python.exe, Linux/macOS: bin/python
if [ -f "$VENV_DIR/Scripts/python.exe" ]; then
    VENV_PY_CHECK="$VENV_DIR/Scripts/python.exe"
elif [ -f "$VENV_DIR/bin/python" ]; then
    VENV_PY_CHECK="$VENV_DIR/bin/python"
else
    VENV_PY_CHECK=""
fi

NEED_VENV=false
if [ -z "$VENV_PY_CHECK" ]; then
    NEED_VENV=true
else
    # 시스템 Python과 venv Python의 minor 버전이 다르면 재생성
    ENV_VER=$("$VENV_PY_CHECK" -c "import sys; print('%d.%d' % sys.version_info[:2])" 2>/dev/null || echo "broken")
    if [ "$ENV_VER" = "broken" ] || [ "$SYS_VER" != "$ENV_VER" ]; then
        warn "venv Python 불일치 (시스템: $SYS_VER, venv: $ENV_VER) → 재생성"
        rm -rf "$VENV_DIR"
        NEED_VENV=true
    else
        ok "venv 이미 존재 (Python $ENV_VER)"
    fi
fi

if [ "$NEED_VENV" = true ]; then
    "$PYTHON" -m venv "$VENV_DIR" >/dev/null 2>&1
    ok "venv 생성 완료 (Python $SYS_VER)"
fi

# pip 실행 경로 (크로스플랫폼)
if [ -f "$VENV_DIR/Scripts/pip.exe" ]; then
    PIP="$VENV_DIR/Scripts/pip.exe"
    VENV_PY="$VENV_DIR/Scripts/python.exe"
else
    PIP="$VENV_DIR/bin/pip"
    VENV_PY="$VENV_DIR/bin/python"
fi

REQS="$TOOLS_DIR/requirements.txt"
if [ -f "$REQS" ]; then
    if ! "$PIP" install -q -r "$REQS"; then
        echo "  ✗ 패키지 설치 실패 (위 에러 확인)"
        exit 1
    fi
    ok "패키지 설치 완료 (requirements.txt)"
fi

# ── 2. Google OAuth 토큰 확인 ───────────────────────────────────
step 2 "Google OAuth 토큰 확인"
mkdir -p "$TOOLS_DIR/output"
if [ -f "$GTOKEN" ]; then
    ok ".gtoken 존재 (토큰 갱신은 자동)"
else
    warn ".gtoken 없음 — Google 인증이 필요합니다."
    warn "아래 명령으로 인증하세요:"
    warn ""
    warn "  $VENV_PY tools/auth.py --url"
    warn "  $VENV_PY tools/auth.py --paste"
fi

# ── 3. 문서 템플릿 다운로드 ──────────────────────────────────────
step 3 "문서 템플릿 확인 (slide_template.pptx / doc_template.docx)"
TMPL_DIR="$TOOLS_DIR/templates"
mkdir -p "$TMPL_DIR"

SLIDE_TMPL="$TMPL_DIR/slide_template.pptx"
DOC_TMPL="$TMPL_DIR/doc_template.docx"
LOGO="$TMPL_DIR/ssrinc_logo.png"
SLIDE_ID="1V7E6vVK5WqnNwhgTdHoi0TSsdtMryRdv"
DOC_ID="1h_-rj1f_xeP_LRCdVd06evMpmWI8DHTs"

# Windows Git Bash: /c/... 형식 경로를 Python이 인식 못 함 → cygpath로 변환
if command -v cygpath &>/dev/null; then
    GTOKEN_PY=$(cygpath -m "$GTOKEN")
    SLIDE_TMPL_PY=$(cygpath -m "$SLIDE_TMPL")
    DOC_TMPL_PY=$(cygpath -m "$DOC_TMPL")
    LOGO_PY=$(cygpath -m "$LOGO")
else
    GTOKEN_PY="$GTOKEN"
    SLIDE_TMPL_PY="$SLIDE_TMPL"
    DOC_TMPL_PY="$DOC_TMPL"
    LOGO_PY="$LOGO"
fi

if [ -f "$SLIDE_TMPL" ] && [ -f "$DOC_TMPL" ] && [ -f "$LOGO" ]; then
    ok "템플릿 이미 존재"
elif [ -f "$GTOKEN" ]; then
    warn "템플릿 다운로드 중..."
    PYTHONUTF8=1 "$VENV_PY" - <<PYEOF
import urllib.request, json, os, sys, time
TOKEN_PATH = "$GTOKEN_PY"
PROXY_URL  = "https://google-workspace-extension.geminicli.com/refreshToken"
# 백업 토큰 경로 (google-workspace MCP 공유 토큰)
import platform, pathlib
_home = pathlib.Path.home()
_tmp  = pathlib.Path(os.environ.get("TEMP", "/tmp"))
BACKUP_PATHS = [
    str(_tmp / "gtoken_backup"),
    str(pathlib.Path(os.environ.get("APPDATA", "")) / "google-workspace-mcp" / "gtoken_backup"),
    str(_home / ".config" / "google-workspace-mcp" / "gtoken_backup"),
]

def _proxy_refresh(token):
    body = json.dumps({"refresh_token": token["refresh_token"]}).encode()
    req  = urllib.request.Request(PROXY_URL, data=body, headers={"Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(req, timeout=10) as r:
            data = json.loads(r.read())
        token["access_token"] = data["access_token"]
        token["expiry_date"]  = int(time.time()*1000) + data.get("expires_in",3600)*1000
        open(TOKEN_PATH, "w").write(json.dumps(token))
        return token["access_token"]
    except Exception:
        return None

try:
    token = json.loads(open(TOKEN_PATH).read())
    is_expired = token.get("expiry_date", 0) < (time.time() * 1000 + 300_000)
    access_token = None
    if not is_expired:
        access_token = token["access_token"]
    else:
        # 전략 1: 현재 .gtoken refresh_token으로 proxy 시도
        access_token = _proxy_refresh(token)
        # 전략 2: 백업 토큰으로 proxy 시도
        if not access_token:
            for bp in BACKUP_PATHS:
                if not os.path.exists(bp):
                    continue
                try:
                    bt = json.loads(open(bp).read())
                    access_token = _proxy_refresh(bt)
                    if access_token:
                        break
                except Exception:
                    continue
    if not access_token:
        raise RuntimeError("토큰 갱신 실패 — tools/.venv/bin/python tools/auth.py 로 재인증 필요")
    for fid, fpath in [("$SLIDE_ID", "$SLIDE_TMPL_PY"), ("$DOC_ID", "$DOC_TMPL_PY")]:
        if os.path.exists(fpath):
            continue
        url = f"https://www.googleapis.com/drive/v3/files/{fid}?alt=media&supportsAllDrives=true"
        r = urllib.request.Request(url, headers={"Authorization": f"Bearer {access_token}"})
        with urllib.request.urlopen(r) as resp:
            open(fpath, "wb").write(resp.read())
        print(f"  ✓ {os.path.basename(fpath)} 다운로드 완료")
    # slide_template.pptx 에서 로고 추출 (slide[0] 우하단 이미지)
    logo_path = "$LOGO_PY"
    if not os.path.exists(logo_path):
        from pptx import Presentation
        prs = Presentation("$SLIDE_TMPL_PY")
        if prs.slides:
            for shape in prs.slides[0].shapes:
                if shape.shape_type == 13:  # PICTURE
                    open(logo_path, "wb").write(shape.image.blob)
                    print(f"  ✓ ssrinc_logo.png 추출 완료")
                    break
except Exception as e:
    print(f"  ! 템플릿 다운로드 실패: {e}")
PYEOF
    [ -f "$SLIDE_TMPL" ] && [ -f "$DOC_TMPL" ] && ok "다운로드 완료" || warn "다운로드 실패 — 인증 후 재시도"
else
    warn ".gtoken 없음 — 인증 후 자동 다운로드됩니다"
fi

# ── 완료 ────────────────────────────────────────────────────────
echo ""
echo "✓ 세팅 완료"

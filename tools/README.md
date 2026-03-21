# GSAC Tools

Claude Code에서 Google Drive 연동 및 문서·시트 생성 기능을 제공하는 MCP 서버와 유틸리티 모음.

---

## 최소 설치 요구사항

| 도구 | 버전 | 비고 |
|------|------|------|
| **Git** | 무관 | OS 패키지 관리자 |
| **Python** | 3.8+ | OS 패키지 관리자 |
| **Claude Code** | 최신 | `npm install -g @anthropic-ai/claude-code` |

> **자동 처리**: Python venv, pip 패키지, 문서 템플릿, 로고 — `setup.sh`이 모두 설치합니다.

> **서버 업로드 (pack/server status)**: `root@192.168.1.10` SSH 공개키 인증 필요.
> 최초 1회 설정: `ssh-copy-id -i ~/.ssh/id_ed25519.pub root@192.168.1.10`

---

## 새 환경 온보딩

```bash
git clone <repo-url>
cd <project>
claude
```

`claude` 실행 시 SessionStart 훅이 자동으로 처리하는 것들:

```
tools/setup.sh 자동 실행
  ├─ python venv 생성/검증       →  tools/.venv/
  │   └─ 시스템 Python 버전과 venv 버전 불일치 시 자동 재생성
  ├─ pip install                 →  requirements.txt
  ├─ 템플릿 다운로드              →  tools/templates/slide_template.pptx
  │                                               doc_template.docx
  └─ 로고 추출                   →  tools/templates/ssrinc_logo.png

MCP 서버 기동 (.mcp.json)
  └─ gdrive  →  tools/start_mcp.sh  →  tools/gdrive_mcp.py
```

---

## Google 인증

### 사전 준비 — 환경변수 설정 (최초 1회)

`auth.py`, `gdrive_mcp.py`, `upload_sheet.py` 모두 아래 두 환경변수를 사용합니다.

```bash
export GOOGLE_CLIENT_ID="920267494314-xxx.apps.googleusercontent.com"
export GOOGLE_CLIENT_SECRET="GOCSPX-xxx"
```

> Google Cloud Console → APIs & Services → Credentials 에서 확인.

**영구 설정** (`~/.bashrc` 또는 `~/.profile`):

```bash
echo 'export GOOGLE_CLIENT_ID="..."'     >> ~/.bashrc
echo 'export GOOGLE_CLIENT_SECRET="..."' >> ~/.bashrc
source ~/.bashrc
```

**MCP 서버 환경 (`.mcp.json` 또는 `start_mcp.sh`)** 에도 동일 변수가 필요합니다.
Claude Code가 MCP 서버를 별도 프로세스로 기동하므로 shell 환경변수가 상속됩니다.

### 초기 인증 (최초 1회)

환경변수 설정 후 아래 명령 실행:

```bash
python tools/auth.py
```

| 모드 | 명령 | 환경 |
|------|------|------|
| 브라우저 자동 | `python tools/auth.py` | Windows / macOS |
| URL만 출력 | `python tools/auth.py --url` | Linux 서버 (headless) |
| 붙여넣기 | `python tools/auth.py --paste` | Linux 서버 (대화형) |

**Linux 서버 흐름 (`--paste` 권장)**:

```bash
# Step 1. 인증 URL 출력
python tools/auth.py --url
# → https://accounts.google.com/o/oauth2/v2/auth?...

# Step 2. URL을 로컬 브라우저에서 열고 개인 Google 계정으로 로그인
#         Google이 http://localhost/?code=4/0AX... 로 리디렉션

# Step 3. 리디렉션된 URL 전체를 붙여넣기
python tools/auth.py --paste
# URL: http://localhost/?code=4/0AX...
# → tools/output/.gtoken 저장 완료
```

인증 결과로 `tools/output/.gtoken`이 생성되며, 이후 모든 API 호출은 로그인한 개인 계정으로 실행됩니다.

### 인증 구조

```
CLIENT_ID / CLIENT_SECRET  →  OAuth 앱 식별 (환경변수)
개인 Google 계정 로그인     →  auth.py 실행 시 브라우저에서
.gtoken                    →  로그인 결과 (access_token + refresh_token)
토큰 갱신                  →  만료 5분 전 자동 갱신 (Google 직접 호출)
```

> `gcloud auth login`과는 **완전히 별개**입니다.

---

## 파일 구조

```
tools/
├── setup.sh                ★ 진입점 — 전체 환경 세팅 (SessionStart 훅으로 자동 실행)
├── auth.py                 ★ Google OAuth 초기 인증 (최초 1회)
├── gdrive_mcp.py           MCP 서버 — Drive/Sheets/Slides/Docs/Calendar/Gmail API
├── start_mcp.sh            크로스플랫폼 Python 경로 래퍼 (.mcp.json에서 호출)
├── requirements.txt        Python 의존성
├── AGENTS.md               AI 어시스턴트 가이드 (Codex / Gemini 등)
├── GEMINI.md               Gemini 전용 가이드
├── scripts/
│   ├── generate_report.py  ★ 리포트 자동 생성 (7가지 타입, run.sh report 로 호출)
│   ├── generate_slide.py   pptx 생성 헬퍼 함수 (add_slide, txt, tbl, logo 등)
│   ├── generate_docx.py    docx 생성 헬퍼 함수
│   ├── generate_sheet.py   xlsx 생성 헬퍼 함수
│   ├── registry.py         doc_registry.json 읽기/쓰기 헬퍼
│   └── upload_sheet.py     로컬 파일 → Google Drive 업로드
├── tests/
│   └── test_e2e.py         e2e 테스트 (DRY / --live 모드)
├── output/                 생성된 파일 저장소 (gitignore)
├── templates/              문서 템플릿 (setup.sh로 자동 다운로드, gitignore)
└── .venv/                  Python 가상환경 (gitignore, setup.sh로 자동 생성)
```

---

## MCP 서버 구성

두 개의 MCP 서버가 역할을 나눠 동작합니다.

### 1. `gdrive` — 프로젝트 MCP (`.mcp.json` → `tools/gdrive_mcp.py`)

로컬 파일 업로드 / Sheets·Slides·Docs 생성·쓰기 전용.

| 툴 | 설명 |
|----|------|
| `drive_uploadFile` | 로컬 파일 → Drive 업로드 (pptx/docx/xlsx/pdf) |
| `drive_createFolder` | Drive 폴더 생성 |
| `drive_listFiles` | 폴더 파일 목록 조회 |
| `drive_downloadFile` | Drive 파일 다운로드 |
| `sheets_create` | Google Sheets 생성 |
| `sheets_updateRange` | 셀 범위 쓰기 |
| `sheets_appendRows` | 행 추가 |
| `sheets_addSheet` | 시트(탭) 추가 |
| `slides_create` | Google Slides 생성 |
| `slides_updateText` | Slides 텍스트 교체 |
| `docs_create` | Google Docs 생성 |
| `docs_appendText` | Docs 끝에 텍스트 추가 |
| `docs_replaceText` | Docs 텍스트 교체 |

### 2. `google-workspace` — 글로벌 MCP (읽기 전용 위주)

Drive 검색, Sheets/Slides/Docs 읽기, Gmail, Calendar, Chat 등.
프로젝트 `.mcp.json`과 별개로 전역 설정에서 로드됨.

| 분류 | 주요 툴 |
|------|---------|
| Drive | `drive_search`, `drive_findFolder` |
| Sheets | `sheets_getRange`, `sheets_getText`, `sheets_getMetadata`, `sheets_find` |
| Slides | `slides_getText`, `slides_getMetadata`, `slides_find` |
| Docs | `docs_getText`, `docs_find`, `docs_move`, `docs_extractIdFromUrl` |
| Gmail | `gmail_search`, `gmail_get`, `gmail_send` 등 |
| Calendar | `calendar_listEvents`, `calendar_createEvent` 등 |
| Chat | `chat_sendMessage`, `chat_sendDm` 등 |
| Auth | `auth_refreshToken`, `auth_clear` |

---

## e2e 테스트

```bash
# DRY 모드 (Google API 호출 없음 — 임포트·구조·env 오류 검사)
tools/.venv/Scripts/python.exe tools/tests/test_e2e.py        # Windows
tools/.venv/bin/python         tools/tests/test_e2e.py        # Linux/macOS

# LIVE 모드 (실제 Google API 호출 — .gtoken + 환경변수 필요)
tools/.venv/Scripts/python.exe tools/tests/test_e2e.py --live
```

> ⚠️ **반드시 venv Python으로 실행** — 시스템 `python`/`python3` 사용 시 패키지를 못 찾아 import 에러 발생

| 테스트 | DRY | LIVE |
|--------|:---:|:----:|
| auth.py 함수 구조 확인 | ✅ | ✅ |
| 3rd-party URL 제거 확인 | ✅ | ✅ |
| generate_slide/docx/sheet 임포트 | ✅ | ✅ |
| registry CRUD | ✅ | ✅ |
| 환경변수 없을 때 에러 메시지 | ✅ | ✅ |
| .gtoken 유효성 | — | ✅ |
| People API 계정 조회 | — | ✅ |
| Drive 파일 목록 조회 | — | ✅ |
| upload_sheet.py 토큰 갱신 | — | ✅ |

---

## 파일 경로 규칙

| 종류 | 저장 위치 | git |
|------|----------|-----|
| 일회성 생성 스크립트 | 파일 저장 없음 — 인라인 실행 후 삭제 | — |
| 출력 파일 (.pptx/.docx/.xlsx) | `tools/output/xxx.pptx` | ❌ |
| 문서 레지스트리 | `tools/output/doc_registry.json` | ❌ |
| OAuth 토큰 | `tools/output/.gtoken` | ❌ |
| 헬퍼 스크립트 | `tools/scripts/generate_slide.py` 등 | ✅ |

> **절대 하지 말 것**: 일회성 스크립트를 `tools/scripts/`에 `.py` 파일로 저장하지 마세요.
> 반드시 임시 파일(`tools/output/_tmp.py`) 패턴으로 실행 후 삭제합니다.

---

## 사용 방법

### 1. 리포트 자동 생성 (`run.sh report`)

Claude 없이 터미널에서 직접 실행 가능한 자동화 리포트입니다.

```bash
# 프로젝트 코드 분석 슬라이드
run.sh report --type code

# 릴리즈 히스토리 슬라이드 (git log 자동 수집)
run.sh report --type release

# 테스트 결과 슬라이드 (JSON 입력)
run.sh report --type test --data tools/output/test_results.json

# 취약점 보고서 슬라이드 (JSON 입력)
run.sh report --type vuln --data tools/output/findings.json

# 생성 + Drive 업로드 한 번에
run.sh report --type release --folder <폴더ID> --key v4_rel

# 타입 목록 + JSON 스키마 확인
run.sh report --list-types
```

**test_results.json 최소 스키마:**
```json
{
  "version": "4.0.34",
  "date": "2026-03-16",
  "suites": [
    {"name": "AD 모듈", "total": 45, "pass": 43, "fail": 2, "skip": 0}
  ]
}
```

**findings.json 최소 스키마:**
```json
{
  "target": "gsac v4.0.34",
  "date": "2026-03-16",
  "findings": [
    {"id": "V-001", "severity": "critical", "title": "...", "location": "...", "description": "...", "recommendation": "..."}
  ]
}
```

---

### 2. 문서 생성 → Drive 업로드

**임시 파일 패턴** — heredoc 금지, 반드시 파일로 저장 후 실행:

```python
# tools/output/_tmp.py (Write 도구로 생성 → 실행 → 삭제)
import sys; sys.path.insert(0, "tools/scripts")
from generate_slide import *
from pptx.util import Inches
from pptx.enum.text import PP_ALIGN

prs, LAYOUT_TITLE, LAYOUT_CONTENT = new_prs()

s1 = add_slide(prs, LAYOUT_TITLE)
txt(s1, "제목", Inches(0.5), Inches(2.9), SLIDE_W - Inches(1), Inches(1),
    size=36, bold=True, color=C_WHITE, align=PP_ALIGN.CENTER)
logo(s1)

s2 = add_slide(prs, LAYOUT_CONTENT)
slide_title(s2, "섹션 제목")
tbl(s2, ["컬럼1", "컬럼2"], [("행1", "값1")])
logo(s2)

prs.save("tools/output/result.pptx")
```

```bash
# 실행 (Windows)
PYTHONUTF8=1 tools/.venv/Scripts/python.exe tools/output/_tmp.py
rm tools/output/_tmp.py

# 실행 (Linux/macOS)
PYTHONUTF8=1 tools/.venv/bin/python tools/output/_tmp.py
rm tools/output/_tmp.py
```

```bash
# Drive 업로드 (--key 지정 시 재실행해도 URL 유지)
run.sh upload tools/output/result.pptx --folder <폴더ID> --name "파일명" --key my_doc_key
```

### 2. Sheets 데이터 업데이트 — MCP 직접 호출 (권장)

데이터만 바뀌는 경우 xlsx 재생성 없이 MCP 툴로 직접 업데이트합니다.

| 상황 | 사용할 툴 |
|------|----------|
| 셀 범위 수정 | `sheets_updateRange(spreadsheetId, range, values)` |
| 행 추가 | `sheets_appendRows(spreadsheetId, sheetName, rows)` |
| 내용 확인 | `sheets_getRange(spreadsheetId, range)` |
| 시트 구조 확인 | `sheets_getMetadata(spreadsheetId)` |

> 한국어 계정은 기본 시트명이 `"시트1"`. `sheets_getMetadata`로 실제 이름 확인 필요.

---

## 문제 해결

**`OAuth 토큰 파일 없음`**
```bash
python tools/auth.py
```

**`환경변수 GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET 가 설정되지 않았습니다`**
```bash
export GOOGLE_CLIENT_ID="..."
export GOOGLE_CLIENT_SECRET="..."
```

**`gdrive` MCP 연결 실패 / Python 패키지 오류**
```bash
bash tools/setup.sh
```

**완전 초기화 후 재설치**
```bash
run.sh clean    # 설치된 파일 일괄 삭제 (.venv / templates/ / output/ / __pycache__ / .claude/bin/ 등)
run.sh install  # 재설치
```

**토큰 만료 (`refresh_token expired`)**
```bash
rm tools/output/.gtoken
python tools/auth.py
```

**Windows에서 Python 경로**
```
tools\.venv\Scripts\python.exe   (Windows)
tools/.venv/bin/python           (Linux/macOS)
```

**`PYTHONUTF8=1` 없이 실행 시 한글 깨짐**

항상 `PYTHONUTF8=1` 환경변수와 함께 실행하세요.

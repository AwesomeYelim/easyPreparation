# AGENTS.md — gsac AI Agent Guide

> **Claude 사용자**: `CLAUDE.md` 를 참조하세요. 이 파일은 Codex / Gemini 등 타 AI용입니다.

이 프로젝트는 Go 기반 보안 감사 에이전트(gsac) 코드베이스입니다.
AI 어시스턴트가 코드 분석, 문서 생성, Drive 업로드를 수행할 때 아래 도구를 사용합니다.

---

## 1. 코드 분석 — `tools/run.sh analyze`

```bash
# 모듈 전체 분석 (통계 + 함수목록)
tools/run.sh analyze AD

# 특정 함수 분석 (위치 + 호출그래프 + 역방향호출)
tools/run.sh analyze AD Ping
```

**모듈 목록:**
| 모듈 | 역할 |
|------|------|
| AA | 리눅스/유닉스 시스템 수집 |
| AD | DB 보안 감사 (MySQL, Oracle, Cassandra 등 25+ DBMS) |
| AW | 웹/WAS 감사 (IIS, JBoss, WebLogic 등) |
| AF | 파일 취약점 분석 |
| AH | 하이퍼바이저 감사 (Hyper-V, Xen) |
| AM | 에이전트 관리 (설치/업데이트/CVE) |
| AS | 에이전트 스크립트 수집 |
| CC | 컨테이너 감사 (Kubernetes, EKS, OpenShift) |
| CH | 수집 헬퍼 |
| GS | 서비스 탐지 (Elasticsearch, Podman, Redis) |
| SA | 서버 에이전트 오케스트레이션 |
| COMMON | 공통 유틸리티 (503개 함수) |

직접 바이너리를 사용할 경우:
```bash
# 플랫폼별 바이너리
.claude/bin/indexer-linux-amd64   # Linux
.claude/bin/indexer-darwin-arm64  # macOS ARM
.claude/bin/indexer.exe           # Windows

# 주요 플래그
-m AD -stats           # 모듈 통계
-m AD -list            # 함수 목록
-m AD -f Ping -l       # 함수 위치
-m AD -f Ping -calls   # 호출 그래프
-m AD -f Ping -callers # 역방향 호출
-m AD -p linux -list   # 플랫폼 필터
-m AD -compare linux:windows  # 플랫폼 비교
```

---

## 2. 리포트 자동 생성 — `tools/run.sh report`

프로젝트를 자동 분석해 슬라이드를 생성합니다. Claude 없이 터미널에서 직접 실행 가능합니다.

```bash
# 코드 구조 분석 (기본)
tools/run.sh report --type code

# 릴리즈 히스토리 (git log 자동 수집)
tools/run.sh report --type release

# 테스트 결과 보고서 (JSON 입력)
tools/run.sh report --type test --data tools/output/test_results.json

# 취약점 보고서 (JSON 입력)
tools/run.sh report --type vuln --data tools/output/findings.json

# 인터페이스 구조 분석
tools/run.sh report --type interface

# 생성 + Drive 업로드 한 번에
tools/run.sh report --type release --folder <폴더ID> --key v4_rel

# 지원 타입 목록 + JSON 스키마 확인
tools/run.sh report --list-types
```

**지원 타입:**
| 타입 | 데이터 수집 | 설명 |
|------|------------|------|
| `code` | 자동 (indexer / TS walk) | 모듈 구조, 함수 수, 파일 통계 |
| `release` | 자동 (git log) | 커밋 히스토리, 기여자, 릴리즈 요약 |
| `interface` | 자동 (indexer) | API/인터페이스 구조 분석 |
| `test` | JSON 입력 | 테스트 스위트 결과 (pass/fail/skip) |
| `vuln` | JSON 입력 | 취약점 목록 (severity 분류) |
| `design` | JSON 입력 | 설계 다이어그램/아키텍처 |
| `wbs` | JSON 입력 | WBS/일정 시각화 |

---

## 3. 문서/슬라이드 직접 생성

Python venv를 사용합니다. 설치가 안 되어 있으면:
```bash
bash tools/setup.sh
```

### 슬라이드 생성 (python-pptx)

> **중요**: heredoc(`cat > file << 'PYEOF'`) 절대 금지.
> 반드시 **Write 도구 → Bash 실행 → Bash 삭제** 패턴 사용.

`tools/output/_tmp.py` 작성 후:

```bash
# Windows
PYTHONUTF8=1 tools/.venv/Scripts/python.exe tools/output/_tmp.py && rm tools/output/_tmp.py

# Linux/macOS
PYTHONUTF8=1 tools/.venv/bin/python tools/output/_tmp.py && rm tools/output/_tmp.py
```

`_tmp.py` 작성 예시:
```python
import sys; sys.path.insert(0, "tools/scripts")
from generate_slide import *
from pptx.util import Inches
from pptx.enum.text import PP_ALIGN

prs, LAYOUT_TITLE, LAYOUT_CONTENT = new_prs()  # 반드시 3개 언패킹

s1 = add_slide(prs, LAYOUT_TITLE)              # add_slide() 필수
txt(s1, "제목", Inches(0.5), Inches(2.9), SLIDE_W - Inches(1), Inches(1),
    size=36, bold=True, color=C_WHITE, align=PP_ALIGN.CENTER)
logo(s1)

s2 = add_slide(prs, LAYOUT_CONTENT)
slide_title(s2, "섹션 제목")
tbl(s2, ["컬럼1", "컬럼2"], [("행1", "값1"), ("행2", "값2")])
logo(s2)

prs.save("tools/output/result.pptx")
```

**generate_slide.py 제공 함수:**
| 함수 | 설명 |
|------|------|
| `new_prs()` | `prs, LAYOUT_TITLE, LAYOUT_CONTENT` 반환 |
| `add_slide(prs, layout)` | 슬라이드 추가 (placeholder 자동 제거) |
| `txt(slide, text, x, y, w, h, ...)` | 텍스트박스, color는 hex 문자열만 |
| `slide_title(slide, text)` | 슬라이드 상단 제목 |
| `tbl(slide, headers, rows, ...)` | 표 삽입 |
| `logo(slide)` | ssrinc 로고 삽입 |
| `steps_list(slide, steps, ...)` | 단계별 목록 |

---

## 4. Drive 업로드

```bash
# 신규 업로드 (토큰 자동 갱신)
tools/run.sh upload tools/output/result.pptx --folder <폴더ID> --key my_doc

# 기존 파일 업데이트 (URL 유지)
tools/run.sh upload tools/output/result.pptx --key my_doc --name "파일명"

# 토큰 상태 확인
tools/run.sh token status

# 토큰 갱신
tools/run.sh token refresh
```

토큰이 없으면: `python tools/auth.py` 로 초기 인증.

---

## 5. 파일 경로 규칙

| 종류 | 경로 |
|------|------|
| 출력 파일 (.pptx/.docx/.xlsx) | `tools/output/xxx.pptx` |
| 임시 스크립트 | `tools/output/_tmp.py` (실행 후 삭제) |
| 템플릿 | `tools/templates/` |
| 업로드 레지스트리 | `tools/output/doc_registry.json` |

---

## 6. 환경 확인

```bash
tools/run.sh help          # 전체 사용법
tools/run.sh token status  # 토큰 유효 여부
tools/run.sh analyze COMMON  # 코드베이스 탐색
```

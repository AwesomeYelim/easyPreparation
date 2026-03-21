# CLAUDE.md — easyPreparation

## 프로젝트 개요

Go 기반 예배 준비 자동화 서버. 찬양/주보 PDF를 생성하고, Google Drive에서 파일을 내려받아 관리하는 도구.

### 진입점
- 서버: `cmd/server/main.go`
- 악보 추출 독립 실행: `cmd/extractMusic/main.go`

### 주요 패키지
| 패키지 | 역할 |
|--------|------|
| `internal/bulletin` | 주보 PDF 생성 |
| `internal/lyrics` | 찬양 PDF 생성 |
| `internal/presentation` | gofpdf 래퍼 (NFC 정규화 포함) |
| `internal/googleCloud` | Google Drive 파일 다운로드 |
| `internal/handlers` | HTTP + WebSocket 핸들러 |
| `internal/types` | 공유 데이터 타입 |
| `internal/utils` | 유틸리티 함수 |
| `internal/middleware` | CORS 미들웨어 |

### 설정 파일
- `config/auth.json` — Google Drive 서비스 계정 키
- `config/db.json` — PostgreSQL DSN
- `config/main_worship.json` — 주예배 순서 데이터

---

## 문서 생성 / Drive 업로드

`tools/run.sh` 사용. 초기 세팅: `bash tools/setup.sh`

```bash
# 슬라이드 생성 (PPTX)
tools/run.sh report --type code

# Drive 업로드
tools/run.sh upload tools/output/result.pptx --folder <폴더ID> --key my_doc

# 토큰 상태 확인
tools/run.sh token status
```

상세 사용법: `AGENTS.md`

### Python 스크립트 실행 패턴

> heredoc 금지. 반드시 **Write 도구 → Bash 실행 → Bash 삭제** 패턴 사용.

`tools/output/_tmp.py` 작성 후:

```bash
PYTHONUTF8=1 tools/.venv/bin/python tools/output/_tmp.py && rm tools/output/_tmp.py
```

### 파일 경로 규칙

| 종류 | 경로 |
|------|------|
| 출력 파일 (.pptx/.docx/.xlsx) | `tools/output/` |
| 임시 스크립트 | `tools/output/_tmp.py` (실행 후 삭제) |
| 템플릿 | `tools/templates/` |
| 업로드 레지스트리 | `tools/output/doc_registry.json` |

---

## 주의사항

- `internal/presentation/presentation.go` — PDF 텍스트 메서드는 모두 NFC 정규화 래퍼 사용 (macOS NFD 문제)
- Google Drive 파일 검색 시 NFC 변환 금지 (Drive에 저장된 파일명이 NFD일 수 있음)
- `config/*` 는 `.gitignore`에 포함됨

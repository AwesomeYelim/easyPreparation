# CLAUDE.md — easyPreparation

## 프로젝트 개요

Go 기반 예배 준비 자동화 서버. 찬양/주보 PDF 생성, Google Drive 연동, OBS 방송 송출을 지원하는 도구.

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
| `internal/handlers` | HTTP + WebSocket + Display 핸들러 |
| `internal/obs` | OBS WebSocket 매니저 (goobs) |
| `internal/types` | 공유 데이터 타입 |
| `internal/utils` | 유틸리티 함수 |
| `internal/middleware` | CORS 미들웨어 |

### 설정 파일
- `config/auth.json` — Google Drive 서비스 계정 키
- `config/db.json` — PostgreSQL DSN
- `config/main_worship.json` — 주예배 순서 데이터
- `config/obs.json` — OBS WebSocket 씬 매핑 (없으면 OBS 비활성)

---

## Display / OBS 시스템

### 예배 화면 (Display)
`/display` — OBS Browser Source 또는 별도 창으로 예배 슬라이드 표시.

- 항목별 렌더링: 성경본문, 찬송(이미지), 교독(이미지), 대표기도, 신앙고백(사도신경), 참회의기도, 말씀, 교회소식
- 찬송/교독: Google Drive PDF → Ghostscript PNG 변환 → `data/hymn/`, `data/responsive_reading/` 캐시
- 성경: DB 자동 조회, 5절 단위 페이징
- 배경: Figma 이미지 (`output/lyrics/tmp/Frame 1.png`) + 어두운 오버레이

### OBS WebSocket
`internal/obs/obs.go` — 싱글턴 매니저, 자동 재연결(5초), `config/obs.json` 없으면 no-op.

### 제어 패널 UI
`ui/app/bulletin/components/DisplayControlPanel.tsx` — 예배 순서 목록 + 클릭 점프 + OBS 상태.

### Display API
| 메서드 | 경로 | 설명 |
|--------|------|------|
| GET | /display | 슬라이드 HTML |
| POST | /display/order | 예배 순서 전송 (성경/찬송 자동 전처리) |
| POST | /display/navigate | next/prev 이동 |
| POST | /display/jump | 특정 항목으로 점프 |
| GET | /display/status | 현재 상태 (idx, OBS) |

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
| 찬송 PDF 캐시 | `data/hymn/` |
| 교독 PDF 캐시 | `data/responsive_reading/` |
| Figma 이미지 캐시 | `output/bulletin/presentation/tmp/` |
| Display PNG 캐시 | `output/display/tmp/` |

---

## 주의사항

- `internal/presentation/presentation.go` — PDF 텍스트 메서드는 모두 NFC 정규화 래퍼 사용 (macOS NFD 문제)
- Google Drive 파일 검색 시 NFC 변환 금지 (Drive에 저장된 파일명이 NFD일 수 있음)
- Ghostscript 경로: `/opt/homebrew/bin/gs` 직접 지정 (fallback: `gs`). `bash -c "gs ..."` 사용 금지
- `config/*`, `output/display/` 는 `.gitignore`에 포함됨
- Figma PNG 캐시가 있으면 API 호출 스킵 — 새 이미지 필요 시 tmp 폴더 비우기

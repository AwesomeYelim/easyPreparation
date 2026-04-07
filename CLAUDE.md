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
| `internal/handlers` | HTTP + WebSocket + Display + 모바일 리모컨 핸들러 |
| `internal/obs` | OBS WebSocket 매니저 (goobs) |
| `internal/types` | 공유 데이터 타입 |
| `internal/utils` | 유틸리티 함수 |
| `internal/middleware` | CORS + 라이선스 기능 게이팅 미들웨어 |
| `internal/license` | 라이선스 관리 (플랜/기능 게이팅/오프라인 캐시) |
| `internal/selfupdate` | GitHub Releases API 기반 업데이트 체커 |
| `internal/version` | 빌드 시 ldflags로 주입되는 버전 정보 |
| `internal/path` | 실행 파일 기준 경로 해석 유틸리티 |

### 설정 파일
- `config/auth.json` — Google Drive 서비스 계정 키
- `config/db.json` — PostgreSQL DSN
- `config/main_worship.json` — 주예배 순서 데이터
- `config/obs.json` — OBS WebSocket 씬 매핑 (없으면 OBS 비활성)
- `config/license.json` — 라이선스 서버 URL + HMAC 시크릿 (없으면 오프라인 모드)

---

## Display / OBS 시스템

### 예배 화면 (Display)
`/display` — OBS Browser Source 또는 별도 창으로 예배 슬라이드 표시.

- 항목별 렌더링: 성경본문, 찬송(이미지), 교독(이미지), 대표기도, 신앙고백(사도신경), 참회의기도, 말씀, 교회소식
- 찬송/헌금봉헌: Google Drive PDF → Ghostscript PNG 변환 → `data/hymn/` 캐시, 표지+이미지 페이지
- 성시교독: Google Drive PDF → PNG, 표지 없이 이미지만 바로 표시
- 성경: DB 자동 조회, 3절 단위 페이징
- 주기도문: 7줄 축약, 한 화면 표시 (페이지 분할 없음)
- **배경 이미지**: 기본 Figma 배경 (`output/lyrics/tmp/Frame 1.png`) + 어두운 오버레이
- **항목별 커스텀 배경**: `output/bulletin/presentation/tmp/{title}.png` — 전주, 찬양, 참회의 기도 3개 자동 매핑
- **교회명 박스**: 커스텀 배경 항목에 영문 교회명 표시 (JacquesFrancois 폰트, 우하단 opacity 박스)
- **상태 영속화**: `data/display_state.json`에 order+idx+churchName 자동 저장, 서버 재시작 시 복원

### 가사 오버레이 (Display Overlay)
`/display/overlay` — OBS Browser Source로 방송 화면에 가사/성경 텍스트 오버레이.

- **반투명 배경 박스**: `rgba(0,0,0,0.75)` 배경, 1500px 고정 너비 (1920×1080 해상도 기준)
- **가사↔페이지 자동 매핑**: 찬송 전처리 시 `hymns` 테이블에서 가사 조회 → 2줄 단위 청크로 분할 → PDF 이미지 페이지 수에 균등 배분 → `item["lyricsMap"]`
- **WS 동기화**: `/display`와 동일한 WS를 통해 navigate/jump 동기화
- 찬송: `lyricsMap[pageIdx-1]` (표지=곡번호, 이미지=가사 텍스트, 가운데 정렬)
- 성경/가사곡/신앙고백 등: 기존과 동일한 텍스트 표시
- 제어판 sections: 찬송 이미지 페이지별 가사 미리보기 (60자 truncate)
- CSS 변수로 폰트 크기/색상/간격 커스터마이즈 가능 (OBS Custom CSS)

**OBS 설정 예시:**
- 프로젝터 출력: Browser Source → `http://localhost:8080/display`
- 방송 출력: Browser Source → `http://localhost:8080/display/overlay` (반투명 배경, 하단 자막)

### OBS WebSocket
`internal/obs/obs.go` — 싱글턴 매니저, 자동 재연결(5초), `config/obs.json` 없으면 no-op.
- 씬 매핑: `config/obs.json`의 `scenes` 맵으로 항목 title → OBS 씬 전환
- 매핑 없는 항목은 씬 전환 스킵 (default fallback 없음)
- `resolveScene` → `(string, bool)` 반환
- **스트리밍 제어**: `StartStreaming()`, `StopStreaming()`, `GetStreamStatus()` — 스케줄러 자동 시작 + 수동 제어

### 자동 스케줄러
`internal/handlers/scheduler.go` — 예배 시간 자동 감지 + OBS 스트리밍 자동 시작.

- **스케줄 설정**: `data/schedule.json`에 영속화 (서버 시작 시 자동 로드)
- **기본 스케줄**: 주일 11:00 / 오후 14:00 / 수요 19:30 / 금요 20:30
- **카운트다운**: 예배 시작 N분 전부터 매초 WS `schedule_countdown` 브로드캐스트 → Display/Overlay에 카운트다운 오버레이
- **자동 실행**: T-0초에 `config/{worshipType}.json` 로드 → 전처리 → Display 순서 교체 + OBS 스트리밍 시작(설정 시)
- **테스트 API**: `/api/schedule/test` — 카운트다운 10초 테스트 또는 즉시 실행 테스트
- **중복 방지**: `lastExecuted` 맵으로 동일 스케줄 하루 1회만 실행
- **UI**: `SchedulePanel.tsx` (설정 모달) + `DisplayControlPanel.tsx` (카운트다운 배너, LIVE 뱃지, 스트리밍 제어)

### 제어 패널 UI
`ui/app/bulletin/components/DisplayControlPanel.tsx` — 예배 순서 목록 + 클릭 점프 + 드래그 앤 드롭 순서 변경 + OBS 상태.

### 성경 탭 UI
`ui/app/bible/page.tsx` — 성경 조회 + Shift-click 범위 선택 (토글 해제 지원).

### Display API
| 메서드 | 경로 | 설명 |
|--------|------|------|
| GET | /display | 슬라이드 HTML (프로젝터용 — 악보 이미지 포함) |
| GET | /display/overlay | 텍스트 오버레이 HTML (방송용 — 반투명 배경) |
| GET | /display/font/{name} | 폰트 파일 서빙 (public/font/) |
| POST | /display/order | 예배 순서 전송 — `{items, churchName}` 또는 배열 (성경/찬송 자동 전처리) |
| POST | /display/append | 항목 추가 — 기존 순서 뒤에 추가 (가사/성경 탭 사용) |
| POST | /display/remove | 항목 삭제 — 인덱스 기반 제거 |
| POST | /display/reorder | 항목 순서 변경 — 드래그 앤 드롭 ({from, to}) |
| POST | /display/navigate | next/prev 이동 |
| POST | /display/jump | 특정 항목으로 점프 (subPageIdx 지원) |
| POST | /display/timer | 자동 넘김 타이머 제어 (enable/disable/speed) |
| GET | /display/status | 현재 상태 (items, idx, OBS, stream) |

### 스케줄러 API
| 메서드 | 경로 | 설명 |
|--------|------|------|
| GET | /api/schedule | 스케줄 설정 조회 |
| POST | /api/schedule | 스케줄 설정 저장 |
| POST | /api/schedule/test | 스케줄 테스트 — `{action: "countdown"\|"trigger", worshipType}` |
| POST | /api/schedule/stream | OBS 스트리밍 수동 제어 — `{action: "start"\|"stop"\|"status"}` |

### Display 통합 구조
- **주보 탭**: `/display/order` — 전체 교체 (예배 순서 일괄 전송)
- **가사/성경 탭**: `/display/append` — 기존 순서 뒤에 추가
- **제어판**: `/display/remove` — 개별 항목 삭제, `/display/reorder` — 드래그 순서 변경
- `openDisplayWindow()` 유틸리티 — 이미 열린 Display 창 reload 방지
- `GlobalDisplayPanel` — 페이지 새로고침 시 서버 상태 자동 복원
- **상태 영속화**: 서버 재시작해도 마지막 순서/위치 유지 (`data/display_state.json`)

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
| 항목별 배경 이미지 | `output/bulletin/presentation/tmp/{title}.png` (전주, 찬양, 참회의 기도) |
| Display PNG 캐시 | `output/display/tmp/` |
| Display 상태 파일 | `data/display_state.json` (order + idx + churchName) |
| 스케줄 설정 파일 | `data/schedule.json` (entries + autoStream + countdownMinutes) |
| 라이선스 캐시 파일 | `data/license.json` (LicenseInfo JSON, 권한 0600) |
| 업데이트 다운로드 | `data/update/` (다운로드 바이너리 임시 저장) |
| 라이선스 서버 설정 | `config/license.json` (server_url + hmac_secret, 없으면 오프라인) |
| CF Workers 라이선스 서버 | `workers/license-api/` (Hono + 토스페이먼츠 + KV) |
| Desktop 앱 빌드 출력 | `build/bin/easyPreparation.app` (macOS) |
| Desktop Windows 빌드 설정 | `cmd/desktop/build/windows/` (manifest, info.json, icon.ico) |
| Desktop Linux 빌드 설정 | `cmd/desktop/build/linux/easyPreparation.desktop` |
| Desktop 프론트엔드 embed | `cmd/desktop/frontend/` (build 시 ui/out 복사) |
| Server 프론트엔드 embed | `cmd/server/frontend/` (build 시 ui/out 복사) |

---

## 에이전트 시스템 (7-Agent Orchestration)

사용자가 계획표를 주면 7개 sub-agent가 자동으로 분업 실행합니다.

| 에이전트 | 역할 | sub-agent 타입 | 프롬프트 |
|----------|------|----------------|----------|
| 시행자 (Planner) | 코드 탐색 → 상세 태스크 JSON 생성 | `Plan` | `.claude/agents/planner.md` |
| 수행자 (Executor) | 태스크별 코드 수정 | `general-purpose` (sonnet) | `.claude/agents/executor.md` |
| 리뷰어 (Reviewer) | 완성도/일관성/누락 감지 | `general-purpose` | `.claude/agents/reviewer.md` |
| 코드 검증자 (Code Inspector) | 빌드/타입/API 정합성/서버 시작 | `Bash` | `.claude/agents/inspector.md` |
| UX 검증자 (UX Inspector) | z-index/반응형/상태흐름/테마 | `general-purpose` | `.claude/agents/ux-inspector.md` |
| 문서 에이전트 (Documenter) | 개발문서/사용자가이드/테스트체크리스트/Git | `general-purpose` | `.claude/agents/documenter.md` |
| 감시자 (Monitor) | 포트/프로세스 관리, 환경 정리 | `Bash` (haiku) | `.claude/agents/monitor.md` |

**실행 흐름**: 감시자(정리) → 시행자(분석) → 수행자(구현, 병렬) → **리뷰어(완성도 체크)** → 감시자(정리) → 코드검증+UX검증(병렬) → 문서에이전트(문서+가이드+Git)

**트리거**: 사용자가 "이 계획을 실행해줘" 또는 계획표를 전달하면:
1. 감시자(haiku)로 포트/프로세스 정리
2. 시행자로 코드베이스 탐색 → 상세 태스크 JSON
3. 수행자(sonnet)를 parallel_group별 병렬 실행
4. **리뷰어가 완성도/일관성/누락 감지** → fix_tasks 있으면 수행자 재실행
5. 감시자(haiku)로 포트 정리
6. 코드 검증자 + UX 검증자 **병렬** 실행
7. 둘 다 pass → 문서 에이전트 (개발문서 + 사용자 가이드 + 테스트 체크리스트 + Git commit & push)
8. 하나라도 fail → fix_tasks로 수행자 재실행 (최대 2회)

**리뷰어가 잡는 것 (검증자가 못 잡는 것)**:
- 데이터 파일 누락 (Display 지원 항목 vs fix_data.json 교차검증)
- UI 중복 메뉴 (같은 모달을 다른 필터로 여는 패턴 → 합치기 제안)
- API 경로 불일치 (Next.js route를 Go BASE_URL로 호출하는 실수)
- 코드 안티패턴 (`if (string)` falsy 트랩, nested Recoil setter)
- 계획 대비 구현 누락

**모델 배정**: 감시자 = haiku, 수행자 = sonnet, 시행자/리뷰어/검증자/문서 = 기본(opus)

**단독 호출**:
- 리뷰어: "코드 리뷰해줘", "누락 확인해줘", "일관성 체크해줘"
- 감시자: "서버 상태 확인", "포트 정리", 세션 터짐 시
- 문서 에이전트: "가이드 업데이트", "테스트 체크리스트 업데이트", "문서 정리"

상세 프로토콜: `.claude/agents/protocol.md`

---

## 예배 순서 데이터 규칙 (`ui/app/data/*.json`)

### `info` 필드 — 편집 가능 여부 결정
| info 값 | 의미 | Detail 컴포넌트 동작 |
|---------|------|---------------------|
| `"c_edit"` | 텍스트 편집 (찬송번호, 제목 등) | textarea + lead input |
| `"b_edit"` | 성경 구절 편집 | BibleSelect + lead input |
| `"edit"` | 일반 편집 (obj + lead) | textarea + lead input |
| `"r_edit"` | lead만 편집 (obj 자동) | lead input only |
| `"notice"` | 교회소식 (ChurchNews 컴포넌트) | ChurchNews UI |
| `"-"` | **자동 처리 (편집 불가)** | "이 항목은 자동으로 처리됩니다" |

### 편집 가능 여부 판단 기준
- `info.includes("edit")` → 편집 UI 표시
- `info === "-"` → 편집 UI 숨김
- **찬송/찬양/성경 관련 항목**은 반드시 `c_edit` 또는 `b_edit` — `"-"`이면 사용자가 번호/구절 입력 불가
- **기도자 입력 필요 항목** (대표기도, 합심기도 등) → `edit` 또는 `r_edit`
- **고정 항목** (전주, 축도, 주기도문, 신앙고백) → `"-"` 정상

### 상태 관리 규칙
- 예배 순서 편집 데이터는 `worshipOrderState` (Recoil atom)에 저장
- `useState`로 복사본을 만들어 편집하면 **드롭다운 전환 시 데이터 유실** — 반드시 Recoil 직접 업데이트
- 새 예배 타입 추가 시: `recoilState.ts`의 `WorshipType` + import + default 모두 추가

---

## 주의사항

- `internal/presentation/presentation.go` — PDF 텍스트 메서드는 모두 NFC 정규화 래퍼 사용 (macOS NFD 문제)
- Google Drive 파일 검색 시 NFC 변환 금지 (Drive에 저장된 파일명이 NFD일 수 있음)
- Ghostscript 경로: `/opt/homebrew/bin/gs` 직접 지정 (fallback: `gs`). `bash -c "gs ..."` 사용 금지
- `config/*`, `output/display/` 는 `.gitignore`에 포함됨
- Figma PNG 캐시가 있으면 API 호출 스킵 — 새 이미지 필요 시 tmp 폴더 비우기

---

## Desktop 앱 (Wails v2)

### 빌드 명령

```bash
# Desktop 앱 빌드 (macOS — Next.js 빌드 + Wails 패키징)
make build-desktop
# 결과물: build/bin/easyPreparation.app

# Desktop 개발 모드 (Wails dev — 핫리로드)
make dev-desktop
```

### 아키텍처

```
easyPreparation.app
└── Wails WebView
    └── http://localhost:8080  ←→  Go HTTP 서버 (내장)
                                   ├── cmd/desktop/main.go  (진입점)
                                   └── internal/api/server.go (동일 라우터)
```

- **진입점**: `cmd/desktop/main.go`
- Wails WebView는 내장 asset server를 사용하지 않고 `http://localhost:8080`을 직접 로드
- `startup()` — DB/OBS/YouTube/스케줄러 초기화 → goroutine으로 HTTP 서버 시작 → 서버 준비 완료 후 `WindowShow`
- `shutdown()` — HTTP graceful shutdown (5초 타임아웃) → 스케줄러 중지 → OBS/DB 닫기
- `StopServer()` — `internal/api/server.go`의 `http.Server.Shutdown()` 래퍼

### CGO_LDFLAGS (macOS 필수)

```bash
CGO_LDFLAGS="-framework UniformTypeIdentifiers" wails build -o easyPreparation
CGO_LDFLAGS="-framework UniformTypeIdentifiers" wails dev
```

Wails가 macOS UniformTypeIdentifiers 프레임워크를 필요로 하므로 반드시 설정해야 함.

### embed_dev.go / embed_prod.go 빌드 태그 분리

| 파일 | 빌드 태그 | 동작 |
|------|-----------|------|
| `cmd/server/embed_dev.go` | `//go:build dev` | `getFrontendFS()` → `nil` (Next.js dev server 사용) |
| `cmd/server/embed_prod.go` | `//go:build !dev` | `//go:embed all:frontend` → `embed.FS` 서빙 |

- 개발 모드(`go run -tags dev` / `make dev`): embed 없음, Next.js dev server(:3000) 프록시
- 프로덕션(`go build` / `make build`): `cmd/server/frontend/`에 `ui/out` 복사 후 embed
- Desktop 앱도 동일 패턴: `cmd/desktop/` 아래에 별도 `embed_dev.go` / `embed_prod.go` 존재

---

## 버전 관리 + 자동 업데이트

### ldflags 주입 (Makefile)

```makefile
VERSION   ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
COMMIT    ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    = -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)
```

- `cmd/server/main.go`, `cmd/desktop/main.go` 모두 `var Version/Commit/BuildTime` 선언 후 `version.Set()` 호출
- `internal/version/version.go` — 싱글턴으로 버전 정보 보관 (thread-safe)
- 태그 없이 빌드하면 `Version="dev"`, 자동 업데이트 알림 항상 표시됨

### 버전 + 업데이트 API

| 메서드 | 경로 | 설명 |
|--------|------|------|
| GET | `/api/version` | `{version, commit, buildTime}` JSON 반환 |
| GET | `/api/update/check` | GitHub Releases 최신 버전과 비교 → `{hasUpdate, latest, current, url}` |

- `internal/selfupdate/checker.go` — `CheckLatest()` (GitHub API), `IsNewer()` (semver 비교)
- GitHub 저장소: `AwesomeYelim/easyPreparation`
- dev 빌드는 `IsNewer()` 항상 true → 개발 중 업데이트 알림 항상 노출

### GitHub Release 워크플로우

`.github/workflows/release.yml` — `v*` 태그 push 시 자동 트리거:

1. 4개 플랫폼 병렬 빌드: `darwin/arm64`, `darwin/amd64`, `linux/amd64`, `windows/amd64`
2. Next.js static export → `cmd/server/frontend/` 복사 → Go binary 빌드 (CGO_ENABLED=0)
3. `softprops/action-gh-release@v2`로 GitHub Release 자동 생성 (릴리즈 노트 자동 생성)
4. 아티팩트: `easyPreparation_darwin_arm64`, `easyPreparation_linux_amd64`, `easyPreparation_windows_amd64.exe` 등

`.github/workflows/test.yml` — PR 시 자동 트리거:

1. `go vet -tags dev ./cmd/server/`
2. `go build -tags dev -o /dev/null ./cmd/server/`

### UpdateChecker.tsx 동작

- 앱 로드 시 `/api/update/check` 폴링 → `hasUpdate: true`이면 헤더에 업데이트 배너 표시
- 배너 클릭 → GitHub Release 페이지(`url`)로 이동
- 업데이트 확인 실패 시 무시 (네트워크 오류 등)

---

## 라이선스 시스템

### 패키지: `internal/license/`

| 파일 | 역할 |
|------|------|
| `types.go` | Plan/Feature 상수 + LicenseInfo 구조체 + PlanFeatures 맵 |
| `manager.go` | 싱글턴 Manager (DB + 파일 캐시 이중 저장, 만료/grace period 관리) |
| `offline.go` | 파일 캐시 (`data/license.json`), 디바이스 ID 생성 (MAC 주소 기반 SHA256) |
| `keygen.go` | 라이선스 키 서명 검증 로직 |

### 플랜별 기능 맵

| 기능 상수 | Free | Pro | Enterprise |
|-----------|------|-----|------------|
| `FeatureOBSControl` | | V | V |
| `FeatureAutoScheduler` | | V | V |
| `FeatureYouTube` | | V | V |
| `FeatureThumbnail` | | V | V |
| `FeatureMultiWorship` | | V | V |
| `FeatureCloudBackup` | | | V |

### middleware.FeatureGate() 사용법

```go
// CORS + Pro 기능 게이팅 조합 — server.go에서 라우트 등록 시 사용
mux.Handle("/api/schedule", middleware.FeatureGate(license.FeatureAutoScheduler, handlers.ScheduleHandler))
```

- 기능 없으면 HTTP 403 + `{"error":"feature_locked","feature":"...","plan":"free","message":"이 기능은 Pro 플랜에서 사용할 수 있습니다."}` 반환
- OPTIONS 요청은 게이팅 없이 통과 (CORS preflight)

### 라이선스 API

| 메서드 | 경로 | 설명 |
|--------|------|------|
| GET | `/api/license` | 현재 라이선스 상태 조회 (plan, deviceId, daysUntilExpiry 등) |
| POST | `/api/license/activate` | 라이선스 키 활성화 — 서버 검증 우선 + 오프라인 fallback |
| POST | `/api/license/deactivate` | 라이선스 비활성화 (무료 플랜으로 복귀) |
| POST | `/api/license/verify` | 서버 측 라이선스 재검증 |
| POST | `/api/license/checkout` | 토스페이먼츠 결제 세션 생성 → `{checkoutUrl, sessionId}` |
| POST | `/api/license/callback` | 결제 완료 폴링 (orderId) → `{status, plan, licenseKey}` |
| POST | `/api/license/portal` | 결제 정보 조회 + 구독 관리 |

### 업데이트 API

| 메서드 | 경로 | 설명 |
|--------|------|------|
| GET | `/api/update/check` | 최신 버전 확인 |
| GET | `/api/update/status` | 다운로드/적용 상태 조회 |
| POST | `/api/update/download` | GitHub Release 바이너리 다운로드 시작 |
| POST | `/api/update/apply` | 바이너리 교체 + 재시작 안내 |
| POST | `/api/update/cancel` | 다운로드 취소 |

### CF Workers 라이선스 서버

`workers/license-api/` — Hono 기반 Cloudflare Workers 라이선스 서버

| 파일 | 역할 |
|------|------|
| `src/index.ts` | API 라우터 (checkout/confirm/activate/verify/portal/cancel/webhook) |
| `src/toss.ts` | 토스페이먼츠 REST API 래퍼 (결제 승인/취소/조회/빌링키) |
| `src/webhook.ts` | 토스페이먼츠 Webhook 이벤트 처리 |
| `src/kv.ts` | KV 스토리지 CRUD (lic:/dev:/sess: 키 네임스페이스) |
| `src/crypto.ts` | HMAC-SHA256 서명 + 라이선스 키 + 주문번호 생성 |

**결제 흐름**: Go 앱 → CF Worker `/api/checkout` → 토스페이먼츠 SDK 결제 페이지 → `/api/confirm` 승인 → 키 생성 → Go 앱 폴링 `/api/activate` → 로컬 활성화

### 라이선스 패키지 구조

| 파일 | 역할 |
|------|------|
| `internal/license/types.go` | Plan/Feature 상수, LicenseInfo 구조체 |
| `internal/license/manager.go` | 싱글턴 Manager (DB + 파일 캐시 이중 저장) |
| `internal/license/offline.go` | 파일 캐시, 디바이스 ID 생성 (MAC SHA256) |
| `internal/license/keygen.go` | HMAC-SHA256 서명 검증 |
| `internal/license/config.go` | `config/license.json` 서버 설정 로드 |
| `internal/license/verifier.go` | 24시간 주기 백그라운드 서버 검증 |

### 자동 업데이트 패키지 구조

| 파일 | 역할 |
|------|------|
| `internal/selfupdate/checker.go` | GitHub Release 최신 버전 확인 + Asset 매칭 |
| `internal/selfupdate/updater.go` | 다운로드 + WS 진행률 브로드캐스트 |
| `internal/selfupdate/updater_unix.go` | macOS/Linux 바이너리 교체 (rename) |
| `internal/selfupdate/updater_windows.go` | Windows batch script 방식 교체 |
| `internal/selfupdate/signature.go` | SHA256 checksums.txt 검증 |

### 프론트엔드 컴포넌트

| 컴포넌트 | 경로 | 역할 |
|----------|------|------|
| `LicensePanel.tsx` | `ui/app/components/LicensePanel.tsx` | 결제 UI + 키 입력 + 구독 관리 모달 |
| `UpdateChecker.tsx` | `ui/app/components/UpdateChecker.tsx` | 다운로드 프로그레스 + 적용/재시작 배너 |
| `FeatureGate.tsx` | `ui/app/components/FeatureGate.tsx` | Pro 전용 UI 영역 래퍼 (잠금 오버레이) |
| `LicenseContext.tsx` | `ui/app/lib/LicenseContext.tsx` | 라이선스 상태 전역 Context (React) |

### 캐시 파일

- `data/license.json` — `LicenseInfo` JSON 직렬화 (권한 0600)
- DB (`licenses` 테이블)와 이중 저장 — DB 없는 환경(Desktop 앱)에서도 라이선스 유지
- 만료 후 30일 grace period (`GracePeriodDays = 30`) — 오프라인 환경 대응

### 동작 방식

- **결제 활성화**: 토스페이먼츠 결제 → CF Worker 키 생성 → Go 앱 폴링으로 수신 → 로컬 저장
- **키 직접 입력**: CF Worker 검증 우선, 실패 시 오프라인 fallback (형식 유효하면 Pro 1년)
- **24시간 검증**: 백그라운드 고루틴이 CF Worker에 라이선스 유효성 확인, 실패 시 Free 전환
- **오프라인 지원**: `config/license.json` 없으면 서버 검증 스킵, grace period 30일

---

## 모바일 PWA 리모컨

### 접근 경로

- 같은 WiFi에서: `http://{로컬IP}:8080/mobile`
- QR 코드: `http://{로컬IP}:8080/mobile/qr.png` (서버가 로컬 IP 자동 감지)
- 제어판 UI에서 QR 아이콘 클릭 → QR 이미지 표시

### 구현 구조

- **`/mobile`** — `MobileRemoteHandler` (인라인 HTML Go 문자열 리터럴)
  - 예배 슬라이드 next/prev 버튼, 현재 항목 표시
  - WebSocket(`/ws`)으로 서버와 실시간 동기화
- **`/mobile/manifest.json`** — `MobileManifestHandler` (PWA 설치 메타데이터)
- **`/mobile/sw.js`** — `MobileServiceWorkerHandler` (오프라인 캐시 Service Worker)
- **`/mobile/icon-192.svg`** — `MobileIconHandler` (SVG 앱 아이콘, EP 텍스트)
- **`/mobile/qr.png`** — `MobileQRHandler` (`github.com/skip2/go-qrcode` PNG 생성)

### 로컬 IP 감지 (`getLocalIP()`)

사설 IP 대역(`192.168.x.x`, `10.x.x.x`, `172.16-31.x.x`) 중 첫 번째 활성 인터페이스 IP 반환. 감지 실패 시 `localhost` fallback.

### PWA 설치

iOS/Android Chrome에서 `/mobile` 접속 → 홈 화면에 추가 → 전체화면 앱으로 실행.

---

## CI/CD

### `.github/workflows/release.yml`

- **트리거**: `v*` 패턴 태그 push (`git tag v1.0.0 && git push origin v1.0.0`)
- **4-job 파이프라인**: `build-frontend`(1회) → `build-server`(4플랫폼) + `build-desktop`(3플랫폼) → `release`
- **Server 빌드**: ubuntu-latest에서 CGO_ENABLED=0 크로스컴파일 (darwin arm64/amd64, linux amd64, windows amd64)
- **Desktop 빌드**: macOS arm64 (.app→.zip), Windows amd64 (NSIS 인스톨러), Linux amd64 (raw binary)
- **8개 아티팩트 + checksums.txt** → GitHub Release 자동 생성
- **릴리즈 노트**: `generate_release_notes: true` (커밋 메시지 자동 수집)

### `.github/workflows/test.yml`

- **트리거**: Pull Request
- **단계**: `go vet -tags dev ./cmd/server/` → `go build -tags dev -o /dev/null ./cmd/server/`
- Desktop 앱은 CI 빌드 대상 아님 (Wails + CGO 환경 구성 복잡성으로 제외)

### 릴리즈 절차

```bash
# 1. 변경사항 커밋 + push
git add -p && git commit -m "feat: ..." && git push

# 2. 태그 생성 + push → GitHub Actions 자동 빌드 + Release 생성
git tag v1.2.0 && git push origin v1.2.0
```

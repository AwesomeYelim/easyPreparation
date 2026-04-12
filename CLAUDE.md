# CLAUDE.md — easyPreparation

Go 기반 예배 준비 자동화 서버. 찬양/주보 PDF, Cloudflare R2 에셋, OBS 방송 송출.

## 구조

| 디렉토리 | 역할 | 자체 CLAUDE.md |
|----------|------|:-:|
| `cmd/server/` | Go 서버 진입점 (:8080) | O |
| `cmd/desktop/` | Wails v2 Desktop 앱 | O |
| `internal/api/` | HTTP 라우터 + SPA 핸들러 | O |
| `internal/handlers/` | Display, WebSocket, 스케줄러, 모바일 | O |
| `internal/obs/` | OBS WebSocket 매니저 | O |
| `internal/license/` | 라이선스 + Feature Gating | O |
| `internal/selfupdate/` | 자동 업데이트 | O |
| `internal/bulletin/` | 주보 PDF 생성 | O |
| `internal/lyrics/` | 찬양 PDF 생성 | O |
| `internal/assets/` | R2 에셋 다운로더 | O |
| `ui/` | Next.js 프론트엔드 | O |
| `landing/` | 홍보 랜딩 페이지 | O |
| `workers/license-api/` | CF Workers 라이선스 서버 | O |
| `tools/` | Python AI 툴킷 | O |
| `.github/workflows/` | CI/CD | O |

## 설정 파일 (gitignore)

- `config/db.json` — DB DSN
- `config/main_worship.json` — 주예배 순서 데이터
- `config/obs.json` — OBS 씬 매핑 (없으면 비활성)
- `config/license.json` — 라이선스 서버 설정 (없으면 오프라인)

## 실행

```bash
make dev          # Go(:8080) + Next.js(:3000)
make build        # 프로덕션 서버 빌드
make build-desktop # Wails Desktop 앱
```

## 핵심 규칙

- PDF 텍스트 → NFC 정규화 래퍼 필수 (`internal/presentation/`)
- Ghostscript: `/opt/homebrew/bin/gs` 직접 지정
- 예배 순서 편집 → Recoil atom 직접 저장 (`useState` 복사 금지)
- `info` 필드: 찬송 `c_edit`, 성경 `b_edit`, 기도자 `edit`/`r_edit`, 고정 `"-"`

## 모델 전략 (Advisor Pattern)

기본 모델: **Sonnet** (매 턴 코드 실행/수정)
어드바이저: **Opus** (복잡한 판단 시 자동 호출)

| 구분 | 모델 | 호출 조건 |
|------|------|-----------|
| 실행 | Sonnet | 단일 파일 수정, 빌드/테스트, grep, git, 간단한 기능 추가 |
| 어드바이저 | Opus | 아키텍처 설계, 복잡한 버그 분석, 대규모 변경 계획, 트레이드오프 분석 |

설정: `.claude/settings.json` → `"model": "sonnet"`, `.claude/agents/advisor.md` → `model: opus`

## ⚠️ 워크플로우 규칙 (MANDATORY — 모든 작업에 적용)

> **이 섹션의 규칙은 선택이 아닌 필수다. 매 작업 시작 전에 이 규칙을 확인하고, 기준표대로 기계적으로 실행하라.**

### Step 1: 영향 범위 분석 (작업 시작 전 반드시 수행)

코드 수정 전에 **먼저** Grep/Glob으로 영향받는 파일 수를 확인한다. 감으로 판단하지 않는다.

### Step 2: 실행 경로 결정 (기계적 — 예외 금지)

| 조건 | 경로 | 실행 방식 |
|------|------|-----------|
| 파일 1개, 단순 변경 | **Direct** | 직접 수정 |
| 파일 2~3개 | **TaskSplit** | TaskCreate로 태스크 나눈 뒤 순차 수행 |
| 파일 4개+ 또는 Go+UI 동시 | **FullOrchestration** | `.claude/agents/protocol.md` 8-agent 풀 실행 |
| 사용자가 계획표/목록 전달 | **FullOrchestration** | 무조건 풀 실행 |
| 단독 에이전트 트리거 (아래) | **SingleAgent** | 해당 에이전트만 호출 |

**절대 금지**: "간단해 보이니까 그냥 한다" — 파일 수로 기계적 결정. 애매하면 상위 경로 선택.

### Step 3: 완료 체크리스트 (작업 끝에 반드시 — 스킵 금지)

```
□ 빌드: Go 수정 시 `go vet ./cmd/server/ ./cmd/desktop/` → make dev 재시작
□ 문서: API 변경 → 패키지 CLAUDE.md | 새 파일 → 루트 CLAUDE.md | 패턴 → MEMORY.md
□ 리포트: 변경 파일 목록 + 문서 업데이트 여부 + 주의사항을 사용자에게 보고
```

### 에이전트 단독 호출 트리거

| 사용자 발화 | 에이전트 |
|------------|---------|
| "코드 리뷰", "누락 확인" | 리뷰어 |
| "서버 상태", "포트 정리" | 감시자 |
| "문서 업데이트", "가이드" | 문서 에이전트 |
| "릴리즈", "태그", "배포" | 배포자 |
| "빌드 확인", "타입 체크" | 코드 검증자 |

## 에이전트 시스템

8-agent 오케스트레이션 + advisor. 상세: `.claude/agents/protocol.md`

감시자 → 시행자 → 수행자(병렬) → 리뷰어 → 검증자(병렬) → 문서에이전트

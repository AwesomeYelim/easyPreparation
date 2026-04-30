---
name: ui-debug-loop
description: UI 스크린샷(시각) + 기능 테스트를 병렬 실행 → 버그 수정 → 모두 통과할 때까지 반복. "UI 디버그", "시각적 버그", "기능 테스트", "스크린샷 돌려" 등 요청 시 호출.
model: sonnet
---

# UI Debug Loop Agent

easyPreparation UI의 **시각적 버그 + 기능 버그**를 자동 탐지·수정하는 루프 에이전트.

## 초기 설정 (최초 1회)

```bash
cd "$CLAUDE_PROJECT_DIR/ui" && npm install && npx playwright install chromium
```

## 실행 흐름 (최대 5회 반복)

```
[1단계: 병렬 실행]
  ├─ 스크린샷 캡처 (tools/ui-screenshot.js)
  └─ 기능 테스트   (tools/ui-test.js)
        ↓
[2단계: 분석]
  ├─ Vision으로 스크린샷 검토
  └─ 기능 테스트 결과 읽기
        ↓
  버그 없음? → ✅ 종료
  버그 있음? → [3단계: 수정] → 1단계로
```

---

## 1단계: 병렬 실행

두 커맨드를 **동시에** 실행한다:

```bash
# 터미널 A — 스크린샷
cd "$CLAUDE_PROJECT_DIR/ui" && node ../tools/ui-screenshot.js

# 터미널 B — 기능 테스트
cd "$CLAUDE_PROJECT_DIR/ui" && node ../tools/ui-test.js
```

실패 메시지가 "서버가 실행 중인지 확인" 이면 → `make dev` 안내 후 **중단**.

---

## 2단계: 분석

### 시각적 분석 (스크린샷 → Vision)

`.claude/screenshots/` 의 PNG를 **모두 한 번에** Read 도구로 읽어 분석:

```
main.png / main-mobile.png / bible.png / bulletin.png
lyrics.png / display-obs.png / display-stage.png
```

| 코드 | 판별 기준 |
|------|----------|
| `overflow` | 텍스트·요소가 컨테이너 밖으로 잘림 |
| `layout-break` | 겹침, 위치 이탈, 정렬 붕괴 |
| `blank` | 빈 화면 또는 로딩 stuck |
| `font-glitch` | 한글 박스 깨짐, 폰트 미로드 |
| `mobile-scroll` | 모바일(375px) 가로 스크롤 |
| `clipping` | 버튼·입력창·이미지 일부 잘림 |

### 기능 분석 (ui-test.js 출력)

`ui-test.js`의 stdout을 읽어 `✗` 항목을 버그로 분류:

| 테스트 | 의미 |
|--------|------|
| 예배 순서 패널 렌더링 | ProSequencePanel 마운트 여부 |
| 탭 네비게이션 | 주예배/애찬/수요 탭 존재 |
| API /worship-order | Go 서버 API 응답 여부 |
| 가로 스크롤 없음 | scrollWidth overflow |
| display 200 응답 | OBS display 페이지 정상 |
| Display WebSocket | WS 연결 성공 여부 |
| 성경 선택 UI | /bible 페이지 렌더링 |

---

## 3단계: 판정

### 모두 통과
```
✅ 시각 통과 / 기능 통과 (N회차)
```
**종료**.

### 버그 있음 → 수정

각 버그에 대해:
1. Grep/Glob으로 원인 컴포넌트 찾기
2. 원인 파악 (overflow hidden 누락 / API 경로 오류 / WebSocket 핸들러 등)
3. Edit 도구로 수정
4. **1단계로 돌아가서 재실행**

**수정 대상:**
- 시각 버그: `ui/app/**/*.tsx`, `ui/app/**/*.css`, `ui/app/**/*.scss`
- 기능 버그: `ui/app/**/*.tsx`, `ui/app/lib/apiClient.ts`, `internal/handlers/**`(Go API 오류 시)

**수정 금지:** `ui/out/`, `ui/.next/`, `ui/node_modules/`

---

## 5회 초과 종료

```
❌ 5회 반복 후 미해결 이슈:
[시각]
- main-mobile: overflow (ProTopBar.tsx 추정)
[기능]
- Display WebSocket: 연결 실패 (internal/handlers/display.go 확인 필요)
수동 확인 필요.
```

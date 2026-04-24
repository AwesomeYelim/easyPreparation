# ui — Next.js 프론트엔드

Next.js 14 + Tailwind CSS v3 + Recoil. `output: "export"` 정적 빌드.

## 디자인 시스템 (Stitch)

| 토큰 | 값 |
|------|-----|
| 폰트 | Inter (Google Fonts) |
| 아이콘 | Material Symbols Outlined |
| 액센트 | electric-blue `#3B82F6` |
| 카드 | `bg-white rounded-2xl border-slate-100 shadow-sm` |
| Pro 배경 | `bg-pro-bg` (`#111`) |
| Pro 텍스트 | `text-pro-text` (`#ccc`) |

## 레이아웃 (Pro Shell — v1.3.0~)

```
ProShell (CSS Grid: 48px | seqPanel | 1fr | inspPanel  ×  44px | 1fr | 90px)
├── ProTopBar       — 탭 네비게이션 + 교회명 + 인스펙터 토글 (row 1, col 1-4)
├── ProIconBar      — 아이콘 사이드바 (48px, row 2-3, col 1)
├── ProSequencePanel — 예배 순서 패널 (280px, row 2, col 2) — seqOpen 토글
├── ProMainArea     — 페이지 콘텐츠 영역 (row 2, col 3)
├── ProInspectorPanel — 세부 속성 패널 (280px, row 2, col 4) — inspOpen 토글
└── ProTimeline     — 씬 타임라인 + 자동 진행 타이머 (90px, row 3, col 2-4)
```

> `GlobalDisplayPanel` — ProShell 외부, 조건부 렌더링 (OBS 전용)

## 주요 파일

| 파일 | 역할 |
|------|------|
| `tailwind.config.ts` | Stitch + Pro 디자인 토큰 |
| `postcss.config.mjs` | Tailwind PostCSS |
| `app/globals.css` | `@import` (폰트) → `@tailwind` directives → CSS 변수 |
| `app/layout.tsx` | Provider 계층 + ProShell 래핑 |
| `app/recoilState.ts` | Recoil atoms (worshipOrder, displayItems, itemTimers 등) |
| `app/lib/apiClient.ts` | Go 서버 API 호출 중앙화 |
| `app/components/ProTimeline.tsx` | 씬 타임라인 — 타이머(초 단위), 드래그 리사이즈, 자동 진행 |
| `app/components/ProSequencePanel.tsx` | 예배 순서 목록 + OBS 방송 제어 |
| `app/components/ProInspectorPanel.tsx` | 현재 씬 세부 정보 |
| `app/components/pro/` | 재사용 서브컴포넌트 (SequenceItem, SequenceStatusBadge) |

## 예배 순서 규칙

- 편집 상태 → `worshipOrderState` Recoil atom 직접 저장 (`useState` 복사 금지)
- `info` 필드: `c_edit`(찬송), `b_edit`(성경), `edit`(기도자), `"-"`(고정)
- 새 예배 타입 → `recoilState.ts` WorshipType + config JSON + Go validWorshipTypes

## ProTimeline 타이머 규칙

- 타이머 값은 **초(seconds)** 단위로 저장 (`itemTimersState`)
- 더블클릭 시 초 값 입력 (예: `120` = 2분), 최소 30초
- 표시 형식: `1:30` (1분 이상), `45s` (1분 미만)
- 드래그로 타일 너비 조정 → 비율 기반 초 재계산

## CSS 주의

- `globals.css`에서 `@import`는 반드시 `@tailwind` 보다 위에 위치 (Turbo dev 서버 스펙 준수)

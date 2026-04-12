# ui — Next.js 프론트엔드

Next.js 14 + Tailwind CSS v3 + Recoil. `output: "export"` 정적 빌드.

## 디자인 시스템 (Stitch)

| 토큰 | 값 |
|------|-----|
| 폰트 | Inter (Google Fonts) |
| 아이콘 | Material Symbols Outlined |
| 액센트 | electric-blue `#3B82F6` |
| 카드 | `bg-white rounded-2xl border-slate-100 shadow-sm` |

## 레이아웃

```
AppShell (CSS Grid)
├── LeftSidebar (w-64) — 주보/찬양/성경 + Display 토글 + 설정
├── TopHeader — 교회명 + 업데이트 배너
└── main — { children } + GlobalDisplayPanel(조건부)
```

## 주요 파일

| 파일 | 역할 |
|------|------|
| `tailwind.config.ts` | Stitch 디자인 토큰 |
| `postcss.config.mjs` | Tailwind PostCSS |
| `app/globals.css` | `@import` (폰트) → `@tailwind` directives → CSS 변수 |
| `app/layout.tsx` | Provider 계층 + AppShell 래핑 |
| `app/recoilState.ts` | Recoil atoms (worshipOrder, displayItems 등) |
| `app/lib/apiClient.ts` | Go 서버 API 호출 중앙화 |

## 예배 순서 규칙

- 편집 상태 → `worshipOrderState` Recoil atom 직접 저장 (`useState` 복사 금지)
- `info` 필드: `c_edit`(찬송), `b_edit`(성경), `edit`(기도자), `"-"`(고정)
- 새 예배 타입 → `recoilState.ts` WorshipType + config JSON + Go validWorshipTypes

## CSS 주의

- `globals.css`에서 `@import`는 반드시 `@tailwind` 보다 위에 위치 (Turbo dev 서버 스펙 준수)

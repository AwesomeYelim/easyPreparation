<div align="center">

<img src="./public/images/image.svg" width="300" alt="EP-UI Logo" />

# EP-UI

**교회 주보 · 가사 PPT 자동 생성 플랫폼**

[![Next.js](https://img.shields.io/badge/Next.js-14-black?logo=next.js)](https://nextjs.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?logo=typescript&logoColor=white)](https://www.typescriptlang.org)
[![Recoil](https://img.shields.io/badge/Recoil-0.7-3578E5?logo=react)](https://recoiljs.org)
[![NextAuth](https://img.shields.io/badge/NextAuth-4-purple)](https://next-auth.js.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-8-336791?logo=postgresql&logoColor=white)](https://www.postgresql.org)
[![SCSS](https://img.shields.io/badge/SCSS-CC6699?logo=sass&logoColor=white)](https://sass-lang.com)

🌐 **[easypreparation.site](https://easypreparation.site)**

</div>

---

## 주요 기능

| 기능 | 설명 |
|------|------|
| 📋 **주보 편집** | 예배 순서를 드래그 없이 선택·편집하고 즉시 미리보기 |
| 📖 **성경 구절 선택** | 책 / 장 / 절을 드롭다운으로 선택, 범위 지정 가능 |
| 📰 **교회 소식 관리** | 트리 구조 소식 항목 추가 / 삭제 / 편집 |
| 🎵 **가사 PPT 생성** | 곡명 입력 → 가사 자동 검색 → ZIP 다운로드 |
| 🔐 **소셜 로그인** | NextAuth 기반 인증 + 교회 프로필 등록 |
| ⚡ **실시간 상태** | WebSocket으로 파일 생성 진행 상황 실시간 수신 |

---

## 스크린샷

> 주보 편집 화면 — 예배 순서 선택 → 항목 편집 → 미리보기

```
┌─────────────────────────────────────────────────────────┐
│  NavBar                                        [프로필]  │
├──────────────┬──────────────────┬───────────────────────┤
│  예배 순서   │   선택된 순서    │    생성된 예배 내용   │
│  선택하기    │                  │                       │
│  ──────────  │  [전주] [예배의] │  전주                 │
│  [전주]      │  [부름] [찬송]   │  예배의 부름  시편..  │
│  [예배의부름]│  [성시교독] ...  │  찬송        27장     │
│  [찬송]      │                  │  성시교독    31편     │
│  [성경봉독]  │   ──────────     │  ...                  │
│  ...         │   [상세 편집]    │                       │
└──────────────┴──────────────────┴───────────────────────┘
```

---

## 프로젝트 구조

```
ep-ui/
├── app/
│   │
│   ├── types/
│   │   └── index.ts              # 공유 타입 (WorshipOrderItem, UserChurchInfo)
│   │
│   ├── lib/
│   │   ├── apiClient.ts          # 외부 API 호출 중앙화
│   │   ├── bibleUtils.ts         # 성경 구절 포맷 유틸
│   │   ├── treeUtils.ts          # 트리 조작 유틸 (delete / insert / find)
│   │   ├── db.ts                 # PostgreSQL 커넥션 풀
│   │   └── next-auth/
│   │       ├── index.tsx         # SessionProvider 래퍼
│   │       └── page.tsx          # 프로필 등록 모달
│   │
│   ├── api/
│   │   ├── auth/[...nextauth]/route.ts   # NextAuth 핸들러
│   │   ├── saveBulletin/route.ts         # 주보 저장
│   │   └── user/route.ts                 # 유저 정보 CRUD
│   │
│   ├── components/               # 전역 공용 컴포넌트
│   │   ├── NavBar.tsx
│   │   ├── NavLink.tsx
│   │   ├── ProfileButton.tsx
│   │   ├── RecoilProvider.tsx
│   │   ├── SideBar.tsx           # 유저 프로필 사이드바
│   │   └── WebSocketProvider.tsx # WebSocket 컨텍스트 + useWS 훅
│   │
│   ├── bulletin/                 # 📋 주보 편집 페이지
│   │   ├── page.tsx
│   │   └── components/
│   │       ├── WorshipOrder.tsx  # 예배 순서 선택
│   │       ├── SelectedOrder.tsx # 선택된 항목 칩 표시
│   │       ├── Detail.tsx        # 항목 상세 편집
│   │       ├── BibleSelect.tsx   # 성경 구절 선택기
│   │       ├── ChurchNews.tsx    # 교회 소식 트리 편집
│   │       ├── EditChildNews.tsx # 소식 하위 항목 편집
│   │       └── ResultPage.tsx    # 미리보기
│   │
│   ├── lyrics/                   # 🎵 가사 PPT 생성 페이지
│   │   ├── page.tsx
│   │   └── components/
│   │       └── LyricsManager.tsx
│   │
│   ├── data/                     # 정적 JSON 데이터
│   │   ├── main_worship.json     # 주일예배 기본 순서
│   │   ├── after_worship.json    # 오후예배 기본 순서
│   │   ├── wed_worship.json      # 수요예배 기본 순서
│   │   ├── fix_data.json         # 고정 선택 항목
│   │   └── bible_info.json       # 성경 책 / 장 / 절 데이터
│   │
│   ├── recoilState.ts            # 전역 Recoil atoms
│   ├── layout.tsx                # 루트 레이아웃
│   └── globals.css / styles.scss
│
└── .claude/
    └── bin/
        ├── indexer.js            # 세션 시작 시 자동 인덱싱
        └── output/index.json     # 인덱스 캐시
```

---

## 아키텍처

### Provider 계층

```
RecoilProvider
  └─ AuthProvider
       └─ WebSocketProvider
            ├─ NavBar
            └─ { children }
```

### 주보 편집 데이터 흐름

```
[WorshipOrder]          선택
    │         ──────────────────▶  selectedInfo (useState)
    │                                     │
[SelectedOrder]         표시 / 삭제       │
    │         ◀────────────────────────────┤
    │                                     │
[Detail]                편집              │
    ├─ BibleSelect                        │
    └─ ChurchNews ◀────────────────────────┤
                                          │
[ResultPage]            미리보기  ◀────────┘
    │
    ▼
apiClient.saveBulletin()
apiClient.submitBulletin()  ──▶  Go 서버  ──▶  [WebSocket] 진행 상황
    │                                               │
    ▼                                               ▼
ZIP 다운로드  ◀─────────────────────────────  done 메시지
```

### 가사 PPT 흐름

```
곡명 입력
    │
    ▼
apiClient.searchLyrics()  ──▶  Go 서버 (가사 자동 검색)
    │
    ▼
가사 확인 / 수정
    │
    ▼
apiClient.submitLyrics()  ──▶  Go 서버  ──▶  ZIP 다운로드
```

---

## 전역 상태

```ts
// app/recoilState.ts

worshipOrderState    // Record<WorshipType, WorshipOrderItem[]>  예배 순서 전체
selectedDetailState  // WorshipOrderItem                         현재 편집 항목
userInfoState        // UserChurchInfo                           로그인 유저 정보
```

---

## 공유 유틸

### `app/lib/treeUtils.ts`

```ts
deleteNode(items, key)          // 트리에서 key 노드 삭제
insertSiblingNode(items, item)  // 형제 노드로 삽입
findNode(items, key)            // 트리에서 key 노드 탐색
```

### `app/lib/bibleUtils.ts`

```ts
formatBibleRanges(multiSelection)  // Selection[][] → "신_5/4:5-4:6, 수_6/5:6"
formatBibleReference(obj)          // "신_5/4:5"    → "신명기 4:5"
```

### `app/lib/apiClient.ts`

```ts
apiClient.saveBulletin(target, targetInfo)    // POST /api/saveBulletin
apiClient.submitBulletin(payload)             // POST {GO}/submit
apiClient.searchLyrics(songs)                 // POST {GO}/searchLyrics
apiClient.submitLyrics(payload)               // POST {GO}/submitLyrics
apiClient.downloadFile(fileName)              // GET  {GO}/download
```

---

## 환경변수

```env
NEXTAUTH_URL=
NEXTAUTH_SECRET=
NEXT_PUBLIC_API_BASE_URL=   # Go 서버 주소
DATABASE_URL=               # PostgreSQL 연결 문자열
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
```

---

## 로컬 실행

```bash
pnpm install
pnpm dev
```

---

<div align="center">
  <sub>Built with ❤️ for church bulletin automation</sub>
</div>

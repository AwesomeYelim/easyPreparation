# UX 개선 계획: 4가지 항목

## 개요
어르신 사용성 개선을 위해 4가지 작업을 진행합니다.
외부 라이브러리 추가 없이, HTML5 Drag & Drop API + 기존 React 상태 관리로 구현합니다.

---

## 1. 순서 드래그 앤 드롭 재배치 (SelectedOrder.tsx)

**현재 문제**: 순서를 잘못 넣으면 삭제 후 다시 추가해야 함

**구현 방식**: HTML5 Drag and Drop API (외부 라이브러리 불필요)
- 각 태그에 `draggable` 속성 추가
- `onDragStart` / `onDragOver` / `onDrop` 핸들러로 순서 변경
- 드래그 중인 항목에 시각적 피드백 (반투명 + 드롭 위치 표시선)
- 드래그 핸들 아이콘(⠿) 추가하여 "이거 잡고 끌어라" 직관적 표시

**수정 파일**: `ui/app/bulletin/components/SelectedOrder.tsx`, `ui/app/styles.scss`

**변경 내용**:
- `SelectedOrder` 컴포넌트: dragIndex/overIndex 상태 추가, drag 이벤트 핸들러, drop 시 배열 재배치
- SCSS: `.tag.dragging` (opacity: 0.4), `.tag.drag-over` (border-top 표시선)
- key를 drop 후 재할당 (index 기반)

---

## 2. 교회소식 트리 UI 개선 (ChurchNews.tsx + styles.scss)

**현재 문제**: 깊이 구분 어려움, + 버튼 찾기 어려움

**구현 방식**:
- 깊이별 왼쪽 들여쓰기 강화 (10px → 24px)
- 깊이별 색상 차이 강화 (파란색 계열 유지하되 채도/밝기 차이 크게)
- 깊이 레이블 표시: depth 0은 "▪ 카테고리", depth 1+은 "└ 세부항목"
- "+" 버튼을 마지막 항목뿐만 아니라 카테고리 영역 하단에도 배치
- 전체적으로 태그/버튼 크기 키우기 (padding 늘리기)

**수정 파일**: `ui/app/bulletin/components/ChurchNews.tsx`, `ui/app/styles.scss`

---

## 3. 삭제 확인 팝업 (SelectedOrder.tsx + ChurchNews.tsx)

**현재 문제**: 실수로 삭제하면 되돌릴 수 없음

**구현 방식**: 간단한 `window.confirm()` 사용 (가벼운 구현, 어르신에게 익숙한 브라우저 팝업)
- SelectedOrder의 삭제 버튼: `confirm("'{title}' 항목을 삭제하시겠습니까?")`
- ChurchNews의 삭제 버튼: `confirm("'{title}' 소식을 삭제하시겠습니까?")`
- 확인 시에만 실제 삭제 수행

**수정 파일**: `ui/app/bulletin/components/SelectedOrder.tsx`, `ui/app/bulletin/components/ChurchNews.tsx`

---

## 4. 안내 메시지 한국어화 + 빈 상태 가이드 (Detail.tsx + SelectedOrder.tsx)

**현재 문제**:
- Detail: 편집 불가 항목에 영어 "is not editable" 표시
- SelectedOrder: 비어있을 때 아무 안내 없음
- Detail: 아무것도 선택 안 했을 때 안내 없음

**구현 방식**:
- Detail.tsx: `"is not editable"` → `"이 항목은 자동으로 처리됩니다"` (편집 불필요라는 뜻)
- SelectedOrder.tsx: 빈 배열일 때 → 안내 메시지: `"위에서 예배 순서를 클릭하여 추가하세요"`
- Detail.tsx: selectedDetail이 초기값(실제 선택 전)일 때 → `"왼쪽에서 항목을 클릭하면 여기서 편집할 수 있습니다"`

**수정 파일**: `ui/app/bulletin/components/Detail.tsx`, `ui/app/bulletin/components/SelectedOrder.tsx`

---

## 수정 파일 요약

| 파일 | 변경 |
|------|------|
| `SelectedOrder.tsx` | 드래그 앤 드롭 + 삭제 확인 + 빈 상태 안내 |
| `ChurchNews.tsx` | 트리 UI 개선 + 삭제 확인 |
| `Detail.tsx` | 한국어 안내 메시지 |
| `styles.scss` | 드래그 스타일 + 교회소식 트리 스타일 |

---

## 완료: 교회소식 수정 UI 버그 수정

### Bug 1 — DELETE가 ADD도 실행 (Critical) [완료]
`ChurchNews.tsx`: `case "DELETE"` 뒤에 `break` 누락 → fall-through로 삭제 시 새 항목 추가됨
- 수정: `break;` 추가

### Bug 2 — `c-edit` info 타입 편집 무시 (Critical) [완료]
`Detail.tsx`: `handleValueChange`에서 `c-edit` (하이픈) 미매칭 → 자식 뉴스 편집 무시
- 수정: `["b_edit", "c_edit", "c-edit", "edit"]`로 배열 확장

### Bug 3 — EditChildNews 초기 상태 (Minor) [완료]
`ChurchNews.tsx`: `selectedChild` 초기값이 부모 항목(info: "notice") → 빈 편집기
- 수정: 초기값 `null`, null guard 추가

---

## 검증
- `next build` 통과 확인
- 외부 라이브러리 추가 없음 (npm install 불필요)

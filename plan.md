# UX 개선 계획

## 완료된 작업

### 1. 순서 드래그 앤 드롭 재배치 [완료]
- `SelectedOrder.tsx`: HTML5 DnD, 드래그 핸들(⠿), 반투명+드롭 위치 표시선
- 삭제를 인덱스 기반으로 변경 (드래그 후 참조 꼬임 방지)
- `styles.scss`: `.tag.dragging`, `.tag.drag-over`, `.drag-handle`

### 2. 교회소식 트리 UI 개선 [완료]
- `ChurchNews.tsx`: 깊이별 들여쓰기 24px, 색상 차이 강화, 깊이 레이블(▪/└)
- `styles.scss`: `.sub-news` 패딩 24px

### 3. 삭제 확인 팝업 [완료]
- `SelectedOrder.tsx` + `ChurchNews.tsx`: `window.confirm()` 추가

### 4. 안내 메시지 한국어화 [완료]
- `Detail.tsx`: "is not editable" → "이 항목은 자동으로 처리됩니다"
- `SelectedOrder.tsx`: 빈 상태 → "위에서 예배 순서를 클릭하여 추가하세요"

### 5. 교회소식 버그 수정 3건 [완료]
- DELETE fall-through → `break;` 추가
- `c-edit` 미매칭 → 배열에 `"c-edit"` 추가
- `selectedChild` 초기값 → `null` + null guard

### 6. 항목 추가 시 key 충돌 수정 [완료]
- `WorshipOrder.tsx`: `key: String(length)` → `key: add_${Date.now()}_${length}`
- `obj` 기본값 보장 (`obj: item.obj || ""`)

### 7. WS 로그 시스템 개선 [완료]
- `wsClient.ts`: `setMessage` → `subscribe` 콜백 패턴 (메시지 누락 방지)
- `WebSocketProvider.tsx`: 타입 적용
- `page.tsx`: 메시지 큐 + 400ms 간격 순차 표시 + 중복 제거
- `bulletin.go`: 단계별 progress ("인쇄용 주보 생성 중...", "프레젠테이션 주보 생성 중...", "주보 생성 완료!")
- `presentation.go`: Google Drive 다운로드/캐시 사용 시 WS broadcast
- `display.go`: Display 전처리 시 Drive 다운로드/캐시 로그 broadcast

### 8. 기타 수정
- `bibleUtils.ts`: `formatBibleReference` obj undefined 가드
- `main_worship.json`: 성찬예식 `obj` 필드 누락 수정

---

## 검증
- `next build` 통과 확인
- 외부 라이브러리 추가 없음

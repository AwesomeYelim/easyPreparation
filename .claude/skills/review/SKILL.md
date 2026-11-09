---
name: review
description: 리뷰어 에이전트 단독 실행. "코드 리뷰", "누락 확인", "일관성 체크", 수행자 작업 후 "이거 맞아?" 등 요청 시 자동 호출.
---

## 변경 파일 현황
- 스테이징됨: !`git diff --name-only --cached 2>/dev/null || echo "없음"`
- 미스테이징: !`git diff --name-only 2>/dev/null || echo "없음"`
- 최근 커밋: !`git log -3 --oneline 2>/dev/null`

---

위 변경 파일 목록으로 general-purpose 서브에이전트(model: opus)를 Agent 도구로 띄워 리뷰를 맡긴다. 프롬프트에 넣을 기준:
- CLAUDE.md 의 금지 패턴(useState(recoilValue), BASE_URL/상대경로 혼동, NFC 누락, log.Fatalf, WS mutex)
- 바꾼 함수의 호출처 누락(indexer `-callers` 로 확인)
- 작업 범위 밖 변경, 런타임 로그가 영어인지
- 결과 형식: 문제마다 file:line + 근거 + 수정안을 `fix_tasks` 목록으로. 확인 못 한 건 "미확인". 추측 금지

리뷰어가 반환한 `fix_tasks`가 있으면 사용자에게 보고하세요. 직접 수정하지 않습니다.

추가 요청: $ARGUMENTS

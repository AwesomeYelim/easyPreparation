---
name: review
description: 리뷰어 에이전트 단독 실행. "코드 리뷰", "누락 확인", "일관성 체크", 수행자 작업 후 "이거 맞아?" 등 요청 시 자동 호출.
---

## 변경 파일 현황
- 스테이징됨: !`git diff --name-only --cached 2>/dev/null || echo "없음"`
- 미스테이징: !`git diff --name-only 2>/dev/null || echo "없음"`
- 최근 커밋: !`git log -3 --oneline 2>/dev/null`

---

`.claude/agents/reviewer.md` 파일을 읽은 뒤, 해당 내용 + 위 변경 파일 목록을 프롬프트에 포함하여 리뷰어 서브에이전트(general-purpose 타입)를 Task 도구로 실행하세요.

리뷰어가 반환한 `fix_tasks`가 있으면 사용자에게 보고하세요. 직접 수정하지 않습니다.

추가 요청: $ARGUMENTS

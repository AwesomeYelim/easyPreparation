---
name: deploy
description: 배포자 에이전트 실행 (GitHub Release, 태그, 릴리즈 노트, CI 모니터링). 사용자가 명시적으로 "배포", "릴리즈", "태그" 요청 시만 사용.
disable-model-invocation: true
---

## 현재 Git 상태
- 브랜치: !`git branch --show-current 2>/dev/null`
- 최근 태그: !`git describe --tags --abbrev=0 2>/dev/null || echo "태그 없음"`
- 마지막 커밋: !`git log -1 --oneline 2>/dev/null`
- 미커밋 변경: !`git status --short 2>/dev/null | head -10`

---

`.claude/agents/deployer.md` 파일을 읽은 뒤, 해당 내용 + 위 Git 상태를 포함하여 배포자 서브에이전트(Bash 타입)를 Task 도구로 실행하세요.

배포 대상 버전/요청: $ARGUMENTS

---
name: health
description: 시스템 상태 확인 (포트, 서버, go vet). "서버 상태", "포트 확인", "세션 터짐" 등 요청 시 사용.
disable-model-invocation: true
---

## 현재 포트 상태
!`lsof -i:8080,3000 -sTCP:LISTEN 2>/dev/null | awk 'NR>1 {print $1, $2, $9}' || echo "리스닝 프로세스 없음"`

## 최근 Go vet 결과
!`cd /Users/hongyelim/easyPreparation && /usr/local/go/bin/go vet ./cmd/... ./internal/... 2>&1 | head -15 || echo "go vet 실패"`

---

`.claude/agents/monitor.md` 파일을 읽은 뒤, 해당 내용 + 위 포트/vet 현황을 포함하여 감시자 서브에이전트(Bash 타입, model: haiku)를 Task 도구로 실행하세요.

추가 확인: 포트 8080/3000 헬스체크, 좀비 프로세스 여부

추가 요청: $ARGUMENTS

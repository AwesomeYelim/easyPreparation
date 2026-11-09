---
name: health
description: 시스템 상태 확인 (포트, 서버, go vet). "서버 상태", "포트 확인", "세션 터짐" 등 요청 시 사용.
disable-model-invocation: true
---

## 현재 포트 상태
!`(command -v lsof >/dev/null && lsof -i:8080,3000 -sTCP:LISTEN 2>/dev/null | awk 'NR>1 {print $1, $2, $9}') || (netstat -ano 2>/dev/null | grep -E ':(8080|3000) .*LISTEN') || echo "리스닝 프로세스 없음"`

## 헬스체크
- Go: !`curl -s -m 3 http://localhost:8080/api/version || echo "응답 없음"`
- Next: !`curl -s -o /dev/null -m 3 -w '%{http_code}' http://localhost:3000 || echo "응답 없음"`

## go vet
!`go vet ./cmd/... ./internal/... 2>&1 | head -15 || echo "go vet 실패"`

---

위 결과를 보고 직접 판단해 보고한다(서브에이전트 불필요).
- 포트를 잡고 있지만 헬스체크가 응답 없는 프로세스는 좀비로 보고, 종료 명령을 제안한다(직접 kill 하기 전에 알린다).
- go vet 경고는 file:line 과 고칠 방법을 적는다.

추가 요청: $ARGUMENTS

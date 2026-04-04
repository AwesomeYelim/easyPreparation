# 에이전트 오케스트레이션 프로토콜

사용자가 계획표를 제공하면 아래 프로토콜로 에이전트를 순차/병렬 실행합니다.

## 에이전트 목록

| 에이전트 | 타입 | 프롬프트 | 역할 |
|----------|------|----------|------|
| 시행자 (Planner) | `Plan` | `.claude/agents/planner.md` | 계획 분석 → 상세 태스크 JSON |
| 수행자 (Executor) | `general-purpose` | `.claude/agents/executor.md` | 태스크별 코드 수정 |
| 코드 검증자 (Code Inspector) | `Bash` | `.claude/agents/inspector.md` | 빌드/타입/API 정합성/문서/Git |
| UX 검증자 (UX Inspector) | `general-purpose` | `.claude/agents/ux-inspector.md` | z-index/반응형/상태흐름/테마 |
| 감시자 (Monitor) | `Bash` | `.claude/agents/monitor.md` | 포트/프로세스 관리 |

## 실행 흐름

```
사용자 계획표
     │
     ▼
┌──────────┐
│  감시자   │  (Bash sub-agent) — 선행 정리
│ monitor  │  포트 충돌 해소, 좀비 프로세스 정리
└────┬─────┘
     │ clean
     ▼
┌──────────┐
│  시행자   │  (Plan sub-agent)
│ planner  │  코드베이스 탐색 → 상세 태스크 JSON 생성
└────┬─────┘
     │ tasks JSON
     ▼
┌──────────┐  parallel_group별 병렬 실행
│  수행자   │  (general-purpose sub-agent × N)
│ executor │  태스크별 코드 수정
└────┬─────┘
     │ 변경 완료
     ▼
┌──────────┐
│  감시자   │  포트 정리 (검증 전)
└────┬─────┘
     ▼
┌──────────────┐  병렬 실행
│ 코드 검증자   │  (Bash) 빌드/타입/API/문서
│ UX 검증자    │  (general-purpose) z-index/반응형/상태흐름
└──────┬───────┘
       │
       ├─ 둘 다 pass → Git commit & push → 서버 시작 → 완료
       │
       └─ 하나라도 fail → fix_tasks 합산
                │
                ▼
           수행자 재실행 (fix_tasks만)
                │
                ▼
           감시자 → 검증자 재실행
```

## 메인 에이전트 실행 코드 (의사코드)

### Phase 0: 감시자 — 환경 정리

```
감시자 = Task(
  subagent_type: "Bash",
  prompt: [.claude/agents/monitor.md 내용] + "\n\n포트 상태 확인 및 충돌 해소. 서버 시작은 하지 않음."
)
→ 포트 상태 JSON 반환
```

- 감시자는 **검사자 전**, **에러 발생 시**, **수동 요청 시** 호출
- 서버 시작은 검사자가 담당 (감시자는 정리만)

### Phase 1: 시행자 실행

```
시행자 = Task(
  subagent_type: "Plan",
  prompt: [.claude/agents/planner.md 내용] + "\n\n## 사용자 계획:\n" + [계획표]
)
→ tasks_json 반환
```

### Phase 2: 수행자 실행 (그룹별)

```
tasks = parse(tasks_json)
groups = group_by(tasks, "parallel_group")

for group in sorted(groups):
  // 같은 그룹은 병렬 실행
  parallel:
    for task in groups[group]:
      수행자 = Task(
        subagent_type: "general-purpose",
        prompt: [.claude/agents/executor.md 내용] + "\n\n## 태스크:\n" + json(task)
      )
```

### Phase 3: 감시자 → 검증자 실행 (코드 + UX 병렬)

```
// 검증 전 포트 정리
감시자 = Task(subagent_type: "Bash", model: "haiku", prompt: monitor.md + "포트 정리")

// 코드 검증자 + UX 검증자 병렬 실행
parallel:
  코드검증 = Task(
    subagent_type: "Bash",
    prompt: [.claude/agents/inspector.md 내용] + "\n\n## 원래 계획:\n" + [계획표 요약]
  )
  UX검증 = Task(
    subagent_type: "general-purpose",
    prompt: [.claude/agents/ux-inspector.md 내용] + "\n\n## 변경된 UI 파일:\n" + [git diff 중 ui/ 파일 목록]
  )

// fix_tasks 합산
all_fixes = 코드검증.fix_tasks + UX검증.fix_tasks

if all_fixes:
  for fix in all_fixes:
    Task(subagent_type: "general-purpose", prompt: executor.md + fix)
  // 감시자 → 검증자 재실행
  Task(subagent_type: "Bash", model: "haiku", prompt: monitor.md + "포트 정리")
  // 재검증 (코드/UX 다시 병렬)
```

## 사용법

사용자가 아래 형태로 계획을 전달:

```
이 계획을 실행해줘:

## Task 1: ...
### 수정
| 파일 | 변경 |
...

## Task 2: ...
...
```

메인 에이전트는:
1. `.claude/agents/monitor.md` 읽고 감시자로 환경 정리
2. `.claude/agents/planner.md` 읽고 시행자 sub-agent 실행
3. `.claude/agents/executor.md` 읽고 수행자 sub-agent 실행 (병렬)
4. `.claude/agents/monitor.md` 읽고 감시자로 포트 정리
5. `.claude/agents/inspector.md` 읽고 검사자 sub-agent 실행

## 감시자 단독 호출

사용자가 아래 요청 시 감시자만 단독 실행:
- "서버 상태 확인해줘"
- "포트 정리해줘"
- "프로세스 정리해줘"
- 세션 터짐/서버 안 뜰 때

```
감시자 = Task(
  subagent_type: "Bash",
  prompt: [.claude/agents/monitor.md 내용] + "\n\n전체 진단 + 정리 + 서버 재시작"
)
```

## 재시도 정책

- 검사자 fail → 최대 2회 재시도
- 2회 재시도 후에도 fail → 사용자에게 보고하고 수동 개입 요청
- 수행자가 old 문자열 매칭 실패 → 시행자에게 해당 태스크 재분석 요청
- 포트 충돌 해소 실패 → 감시자 재실행 1회 후 사용자에게 보고

## 주의사항

- 커밋/푸시는 전체 프로세스 완료 후 사용자가 직접 요청할 때만 수행
- Go 서버 코드 수정 시 `make dev` 재시작은 검사자 단계에서 수행
- 감시자는 kill 전 항상 프로세스 확인 — 무차별 kill 금지
- 각 에이전트의 응답은 JSON 형식이어야 파싱 가능

# 에이전트 오케스트레이션 프로토콜

사용자가 계획표를 제공하면 아래 프로토콜로 에이전트를 순차/병렬 실행합니다.

## 에이전트 목록

| 에이전트 | 타입 | 프롬프트 | 역할 |
|----------|------|----------|------|
| 시행자 (Planner) | `Plan` | `.claude/agents/planner.md` | 계획 분석 → 상세 태스크 JSON |
| 수행자 (Executor) | `general-purpose` | `.claude/agents/executor.md` | 태스크별 코드 수정 |
| 리뷰어 (Reviewer) | `general-purpose` | `.claude/agents/reviewer.md` | 완성도/일관성/누락 감지 |
| 코드 검증자 (Code Inspector) | `Bash` | `.claude/agents/inspector.md` | 빌드/타입/API 정합성/서버 시작 |
| UX 검증자 (UX Inspector) | `general-purpose` | `.claude/agents/ux-inspector.md` | z-index/반응형/상태흐름/테마 |
| 문서 에이전트 (Documenter) | `general-purpose` | `.claude/agents/documenter.md` | 개발문서/사용자가이드/테스트체크리스트/Git |
| 감시자 (Monitor) | `Bash` | `.claude/agents/monitor.md` | 포트/프로세스 관리 |
| 배포자 (Deployer) | `Bash` | `.claude/agents/deployer.md` | GitHub Release/태그/릴리즈노트/CI 모니터링 |

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
│  리뷰어   │  (general-purpose) — 완성도/일관성/누락 감지
│ reviewer │  데이터↔코드 교차검증, API 경로 정합, 안티패턴
└────┬─────┘
     │
     ├─ fix_tasks 있음 → 수행자에게 전달 → 리뷰어 재실행 (1회)
     │
     ├─ fix_tasks 없음 (pass)
     ▼
┌──────────┐
│  감시자   │  포트 정리 (검증 전)
└────┬─────┘
     ▼
┌──────────────┐  병렬 실행
│ 코드 검증자   │  (Bash) 빌드/타입/API/서버 시작
│ UX 검증자    │  (general-purpose) z-index/반응형/상태흐름
└──────┬───────┘
       │
       ├─ 둘 다 pass ──▼
       │         ┌────────────┐
       │         │ 문서 에이전트 │  개발문서 + 가이드 + 체크리스트 + Git
       │         │ documenter  │
       │         └─────────────┘
       │                → 완료
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

### Phase 2.5: 리뷰어 실행

```
리뷰어 = Task(
  subagent_type: "general-purpose",
  prompt: [.claude/agents/reviewer.md 내용] +
          "\n\n## 원래 계획:\n" + [계획표 요약] +
          "\n\n## 변경된 파일:\n" + [git diff --name-only]
)

if 리뷰어.fix_tasks (priority == "high"):
  for fix in 리뷰어.fix_tasks:
    Task(subagent_type: "general-purpose", prompt: executor.md + fix)
  // 리뷰어 재실행 (1회만)
```

**리뷰어가 잡는 것들 (검증자가 못 잡는 것들):**
- 데이터 파일 누락 (Display 지원 항목 vs fix_data.json)
- UI 중복 메뉴 (같은 모달을 다른 필터로 여는 패턴)
- API 경로 불일치 (Next.js route를 Go BASE_URL로 호출)
- 코드 안티패턴 (`if (string)`, nested Recoil setter)
- 계획 대비 누락 기능

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

### Phase 4: 문서 에이전트 실행

```
// 코드검증 + UX검증 모두 pass 후에만 실행
if 코드검증.status == "pass" and UX검증.status == "pass":
  문서 = Task(
    subagent_type: "general-purpose",
    prompt: [.claude/agents/documenter.md 내용] +
            "\n\n## 원래 계획:\n" + [계획표 요약] +
            "\n\n## 변경된 파일:\n" + [git diff --name-only]
  )
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
4. `.claude/agents/reviewer.md` 읽고 리뷰어 sub-agent 실행 → fix_tasks 있으면 수행자 재실행
5. `.claude/agents/monitor.md` 읽고 감시자로 포트 정리
6. `.claude/agents/inspector.md` + `.claude/agents/ux-inspector.md` 읽고 검증자 병렬 실행
7. `.claude/agents/documenter.md` 읽고 문서 에이전트 실행 (검증 pass 후)

## 리뷰어 단독 호출

사용자가 아래 요청 시 리뷰어만 단독 실행:
- "코드 리뷰해줘"
- "누락 확인해줘"
- "일관성 체크해줘"
- 수행자 작업 후 "이거 맞아?"

```
리뷰어 = Task(
  subagent_type: "general-purpose",
  prompt: [.claude/agents/reviewer.md 내용] +
          "\n\n## 변경된 파일:\n" + [git diff --name-only] +
          "\n\n## 확인 요청:\n" + [사용자 요청]
)
```

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

## 문서 에이전트 단독 호출

사용자가 아래 요청 시 문서 에이전트만 단독 실행:
- "가이드 업데이트해줘"
- "사용자 가이드 만들어줘"
- "테스트 체크리스트 업데이트해줘"
- "문서 정리해줘"

```
문서 = Task(
  subagent_type: "general-purpose",
  prompt: [.claude/agents/documenter.md 내용] + "\n\n## 요청:\n" + [사용자 요청]
)
```

## 배포 에이전트 단독 호출

사용자가 아래 요청 시 배포 에이전트만 단독 실행:
- "릴리즈 배포해줘", "v1.0.0 배포해줘"
- "릴리즈 노트 작성해줘"
- "이전 릴리즈 삭제해줘"
- "CI 상태 확인해줘", "빌드 확인해줘"
- "태그 정리해줘"

```
배포자 = Task(
  subagent_type: "Bash",
  prompt: [.claude/agents/deployer.md 내용] + "\n\n## 요청:\n" + [사용자 요청]
)
```

## 재시도 정책

- 리뷰어 fix_tasks → 수행자 1회 재실행 후 리뷰어 재확인 (최대 1회)
- 검사자 fail → 최대 2회 재시도
- 2회 재시도 후에도 fail → 사용자에게 보고하고 수동 개입 요청
- 수행자가 old 문자열 매칭 실패 → 시행자에게 해당 태스크 재분석 요청
- 포트 충돌 해소 실패 → 감시자 재실행 1회 후 사용자에게 보고

## 주의사항

- 커밋/푸시는 문서 에이전트가 최종 단계에서 수행
- Go 서버 코드 수정 시 `make dev` 재시작은 검사자 단계에서 수행
- 감시자는 kill 전 항상 프로세스 확인 — 무차별 kill 금지
- 각 에이전트의 응답은 JSON 형식이어야 파싱 가능
- 리뷰어는 코드를 직접 수정하지 않음 — fix_tasks로 수행자에게 전달

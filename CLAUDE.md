# easyPreparation

Go(:8080) + Next.js(:3000) 예배 준비 자동화. 주보 PDF, OBS 송출, R2 에셋. 데스크톱은 Wails v2, 찬송가는 SQLite, 라이선스는 CF Workers.

## Context — docs/context.md 가 맥락 원본
운영 현황·결정·진행 중 브랜치·다음 할 일:
@docs/context.md

- 로컬·클라우드(폰) 어느 세션이든 끝낼 때 docs/context.md 의 "현재 상태"·"진행 중 브랜치"·"다음 할 일" 을 고치고 `bash scripts/push-context.sh` 를 실행한다. 작업 브랜치와 상관없이 이 파일만 master 에 바로 커밋·push 한다(코드는 그대로 브랜치에). master push 가 막히면 지금 브랜치에 올리고, life 워크플로가 1시간 안에 master 로 옮긴다.
- 클라우드 세션이 못 하는 일(교회 PC·배포 서버·로컬 `.env`·Windows/OBS 실측)은 "다음 할 일" 에 `(PC)` 를 붙여 남긴다.
- 공개 저장소다. 호스트명·키·개인 정보는 적지 않는다.
- life 저장소의 같은 노트는 매시간 받아 가는 복사본이다.

## 세션 규칙 (로컬·클라우드 공통)
클라우드(폰) 세션은 PC 의 `~/.claude` 전역 규칙을 못 읽는다. 이 절이 그 대신이다.
- 첫 응답 첫 줄에 `세션 이름: <요약 10자 이내>` 를 적는다.
- 되돌릴 수 있는 일은 스스로 결정하고 한 뒤 응답에 `결정: ...` 한 줄로 남긴다. 두 방법 중 고를 때는 되돌리기 쉬운 쪽·기존 코드 패턴과 같은 쪽.
- 묻는 경우는 셋뿐: 돈이 드는 일, 외부로 나가는 것(릴리스·태그·배포·외부 게시), 되돌리기 어려운 변경(운영 데이터·삭제). 물을 때는 응답 **맨 끝 한 줄**을 `?` 로 끝내고, 답 없이 할 수 있는 다른 일을 먼저 한다. 세션을 끝낼 때 열린 질문은 docs/context.md "다음 할 일" 에도 남긴다.
- 코드는 브랜치(`feature/...`, 클라우드는 `claude/...`)에서 고치고, 확인이 끝나면 master 에 병합한다(PR 필수 아님). 운영 중인 앱이므로 동작이 바뀌는 수정은 먼저 보여 주고 승인받는다.
- 요청 밖 개선은 응답 끝 제안 한 줄로. 승인 후 범위에 넣는다.
- 런타임 로그 메시지는 영어. 주석·문서는 한국어 가능.
- 답은 한국어, 불릿 위주. 긴 표는 쓰지 않는다.

## 코드 분석 — indexer 먼저
Go 코드를 볼 때는 `.claude/bin/indexer` 를 먼저 쓴다(세션 시작 훅이 인덱스를 만든다). Windows `indexer.exe`, Linux `indexer-linux-amd64`, macOS `indexer-darwin-*`.
- `-stats` → `-m <모듈> -list` → `-m <모듈> -f <함수> -l` / `-calls` / `-callers` / `-impact` 순서. 모듈 이름은 대문자(API, HANDLERS, OBS, PRESENTATION, SERVER …).
- indexer 로 위치를 잡은 뒤 필요한 함수 블록만 offset+limit 로 Read 한다. 파일 전체 Read 는 피한다.
- 프론트(ui/)는 indexer 대상이 아니다. Grep/Glob 으로 찾는다.

## 위임 — 검색과 분석 분리
- 위치 찾기·목록·호출처 → Explore(`model: haiku` 또는 `sonnet`). 산출물은 file:line, 확인 안 된 건 "미확인".
- "무엇을 하는가" 해석·영향 분석·코드 수정 → general-purpose(opus). Explore 에 해석을 시키지 않는다.
- 파일 3개 이상 읽는 조사는 위임한다. 서로 독립인 작업은 한 응답에 Agent 여러 개를 동시에 띄운다.
- 위임 결과가 의심스러우면 처음부터 다시 스캔하지 말고 해당 file:line 만 직접 확인하거나, 같은 에이전트에 SendMessage 로 부족분만 요청한다.
- 설계 판단·트레이드오프는 `advisor` 도구로 묻는다.

## 구현 전 검증
4개 이상 파일 변경이 예상되면 코드 수정 전에 `/grill-me` 로 설계를 검증하고, 사용자가 "시작해" 하면 들어간다. 사용자가 "바로 해" 라고 하면 생략.

## 금지 패턴 (훅이 자동 검출하지만 기억해둘 것)

- `useState(recoilValue)` → `useRecoilState(atom)` 직접 사용. 탭 전환 시 데이터 유실.
- Go API는 `${BASE_URL}/api/...`, Next.js API는 `/api/...`. 헷갈리면 404.
- PDF 텍스트 → `presentation.NFC()` 래퍼 필수. 안 쓰면 한글 깨짐.
- `log.Fatalf` 금지 → `return fmt.Errorf`. 서버 크래시됨.
- WS 쓰기 → mutex 필수. race condition.

## 수정 후 반드시
1. `git diff` 재확인
2. 변경한 함수의 호출처 1개 이상 확인(indexer `-callers` 또는 Grep)
3. Go 수정 → 훅이 해당 패키지 `go vet` 을 돌린다. 서버 동작은 `make dev` 로 확인
4. UI 수정 → 훅이 `tsc --noEmit` 를 돌린다. 화면은 브라우저(클라우드면 헤드리스 스크린샷)로 확인

## 라우팅
- `${BASE_URL}` = Go 서버 `http://localhost:8080`
- 상대경로 `/api/...` = Next.js :3000
- `apiClient.ts` 에 모든 Go API 함수 정의됨
- 운영 빌드는 Go 가 Next 정적 결과를 임베드해 서빙한다. 개발(:3000)과 상대경로 동작이 다르다.

## 상태

- 예배 순서: `config/{type}.json` (Go API 경유)
- 편집 상태: Recoil atom (`recoilState.ts`)
- Display: `data/display_config.json`, `data/display_state.json`
- info 필드: `c_edit`(찬송) `b_edit`(성경) `edit`(기도자) `r_edit`(인도자만) `"-"`(고정)

## 실행
```bash
make dev            # Go + Next.js 동시
make build          # 프로덕션
make build-desktop  # Wails Desktop
```

## 에이전트·스킬
- `deployer` 에이전트 / `/deploy` — 릴리스·태그·릴리스 노트·CI. 사용자가 "배포/릴리스/태그" 라고 할 때만. 정식 릴리스는 배포 서버가 다음 날 03:00 자동 반영한다(일요일 제외).
- `ui-debug-loop`, `ui-inspector` — UI 스크린샷·기능 테스트, 레이아웃 교정
- `/review`, `/health`, `/worship-plan`, `/grill-me`, `/handoff`
- CI/CD 상세: `.github/workflows/CLAUDE.md`

## 코드 규칙 (세션 실수 기록)
세션이 지적받아 정정된 규칙을 모은다. 새 실수가 생기면 여기에 한 줄 추가한다.

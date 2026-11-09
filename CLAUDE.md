# easyPreparation

Go(:8080) + Next.js(:3000) 예배 준비 자동화. 주보 PDF, OBS 송출, R2 에셋.

## 금지 패턴 (훅이 자동 검출하지만 기억해둘 것)

- `useState(recoilValue)` → `useRecoilState(atom)` 직접 사용. 탭 전환 시 데이터 유실.
- Go API는 `${BASE_URL}/api/...`, Next.js API는 `/api/...`. 헷갈리면 404.
- PDF 텍스트 → `presentation.NFC()` 래퍼 필수. 안 쓰면 한글 깨짐.
- `log.Fatalf` 금지 → `return fmt.Errorf`. 서버 크래시됨.
- WS 쓰기 → mutex 필수. race condition.

## 수정 후 반드시

1. `git diff` 재확인
2. 변경한 함수의 호출처 1개 이상 Grep
3. UI 수정 → 브라우저에서 확인
4. Go 수정 → `go vet` + `make dev`

## 라우팅

- `${BASE_URL}` = Go 서버 `http://localhost:8080`
- 상대경로 `/api/...` = Next.js :3000
- `apiClient.ts`에 모든 Go API 함수 정의됨

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

## 구현 전 검증

4개 이상 파일 변경이 예상되는 작업이면, 코드 수정 전에 `/grill-me` 프로세스를 먼저 실행하라.
설계 트리의 각 가지를 질문으로 검증한 뒤, 사용자가 "시작해"/"구현해" 신호를 보내면 구현에 들어간다.
사용자가 명시적으로 건너뛰면 ("바로 해", "질문 필요없어") 검증 없이 진행.

## 에이전트: `.claude/agents/protocol.md`

대규모 변경(4개+ 파일) 또는 사용자가 계획표 전달 시 사용.
"코드 리뷰" → 리뷰어 / "포트 정리" → 감시자 / "배포" → 배포자

## Context — docs/context.md 가 맥락 원본
운영 현황·결정·진행 중 브랜치·다음 할 일:
@docs/context.md

- 로컬·클라우드(폰) 어느 세션이든 끝낼 때 docs/context.md 의 "현재 상태"·"진행 중 브랜치"·"다음 할 일" 을 고치고 `bash scripts/push-context.sh` 를 실행한다. 작업 브랜치와 상관없이 이 파일만 master 에 push 한다(코드는 그대로 브랜치에).
- 클라우드 세션이 못 하는 일(교회 PC·배포 서버·로컬 `.env`·Windows/OBS 실측)은 "다음 할 일" 에 `(PC)` 를 붙여 남긴다.
- 공개 저장소다. 호스트명·키·개인 정보는 적지 않는다.
- life 저장소의 같은 노트는 매시간 받아 가는 복사본이다.

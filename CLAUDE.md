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

## 에이전트: `.claude/agents/protocol.md`

대규모 변경(4개+ 파일) 또는 사용자가 계획표 전달 시 사용.
"코드 리뷰" → 리뷰어 / "포트 정리" → 감시자 / "배포" → 배포자

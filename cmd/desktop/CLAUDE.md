# cmd/desktop — Wails v2 Desktop 앱

WebView로 `http://localhost:8080` 로드. 서버와 동일한 라우터 사용.

## 구조

```
easyPreparation.app
└── Wails WebView → AssetServer(로딩 화면) → localhost:8080 자동 전환
    └── Go HTTP 서버 (내장, goroutine)
```

## 주요 동작

- `startup()` — DB/OBS/YouTube/스케줄러 초기화 → `go api.StartServer()` → 서버 준비 후 `WindowShow`
- 포트 8080 충돌 → `api.ServerError` 채널 → Wails 에러 다이얼로그 (panic 아님)
- `shutdown()` — HTTP graceful shutdown(5초) → 스케줄러/OBS/DB 정리

## 빌드

```bash
CGO_LDFLAGS="-framework UniformTypeIdentifiers" wails build -o easyPreparation
make build-desktop  # 위 명령 래핑
```

macOS `UniformTypeIdentifiers` 프레임워크 필수.

## embed / 초기화

- embed 자산은 `internal/embedded` 공용 패키지(`FrontendFS()`, `DataFS()`). 이 디렉터리에 embed 코드 없음.
- `uibase_dev.go` / `uibase_prod.go` — WebView 가 리디렉션할 UI URL 만 분기 (dev: `:3000` Next.js, prod: `:8080` 내장 서버).
- 서버/DB/라이선스/OBS/스케줄러 초기화는 `internal/app.Initialize()` 를 그대로 사용한다(`main.go` startup). 데스크톱 고유 동작은 `SetDesktopMode`(다운로드 경로), `OnServerError` 다이얼로그, 헬스체크 실패 시 롤백 제안, `WindowShow` 만.
- 초기화 단계를 추가할 때는 `internal/app/init.go` 에 넣을 것 — 여기 따로 넣으면 서버와 다시 어긋난다(과거에 라이선스 백그라운드 검증이 데스크톱에서 빠졌던 원인).

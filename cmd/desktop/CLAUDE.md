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

## embed 분리

`embed_dev.go` / `embed_prod.go` — `cmd/server/`와 동일 패턴.

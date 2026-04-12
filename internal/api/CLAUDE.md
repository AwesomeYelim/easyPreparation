# internal/api — HTTP 서버 + SPA 핸들러

## server.go

- `StartServer(dataChan, readyCh...)` — 모든 라우트 등록 + `net.Listen` + `srv.Serve`
- 포트 충돌 시 `ServerError` 채널로 에러 전달 (panic 아님)
- `StopServer(ctx)` — graceful shutdown
- `spaHandler(frontFS)` — 파일 있으면 서빙, 없으면 `index.html` fallback (SPA 라우팅)
- `FrontendFS` — `nil`이면 정적 파일 비활성 (dev 모드)

## 라우트 등록 순서

1. CORS 미들웨어 적용
2. `/ws` WebSocket
3. `/display/*` Display API
4. `/mobile/*` 모바일 리모컨
5. `/api/*` REST API (일부 `middleware.FeatureGate` 적용)
6. `/` SPA fallback (프로덕션 모드만)

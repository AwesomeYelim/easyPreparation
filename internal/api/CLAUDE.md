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

## 주요 Display 라우트

| 경로 | 핸들러 | 설명 |
|------|--------|------|
| `/display` | `DisplayHandler` | 프로젝터 슬라이드 HTML |
| `/display/overlay` | `DisplayOverlayHandler` | 방송 자막 오버레이 HTML |
| `/display/stage` | `DisplayStageHandler` | 무대 모니터 HTML |
| `/display/preview` | `DisplayPreviewHandler` | 씬 패널 미리보기 HTML (WS 없음, index 쿼리) |
| `/display/print` | `HandleDisplayPrint` | 인쇄용 슬라이드 HTML |
| `/display/video-bg/` | `VideoBgServeHandler` | 비디오 배경 파일 서빙 |
| `/api/display-config` | `HandleDisplayConfigGet/Set` | Display 설정 GET/PUT |
| `/api/video-bg/upload` | `VideoBgUploadHandler` | 비디오 파일 업로드 |
| `/api/video-bg/list` | `VideoBgListHandler` | 비디오 목록 조회 |
| `/api/video-bg/delete` | `VideoBgDeleteHandler` | 비디오 파일 삭제 |

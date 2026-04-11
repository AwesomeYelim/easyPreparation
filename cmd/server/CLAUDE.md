# cmd/server — Go 서버 진입점

포트 `:8080`. `make dev`로 실행 시 `-tags dev` 빌드.

## 파일

| 파일 | 역할 |
|------|------|
| `main.go` | 진입점 — DB/OBS/YouTube/스케줄러 초기화 → HTTP 서버 시작 |
| `embed_dev.go` | `//go:build dev` — 프론트엔드 embed 없음 (Next.js dev server 사용) |
| `embed_prod.go` | `//go:build !dev` — `frontend/` 디렉토리 embed → 정적 파일 서빙 |

## ldflags

```makefile
-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)
```

태그 없이 빌드하면 `Version="dev"`.

## 프론트엔드 embed

프로덕션: `ui/out` → `cmd/server/frontend/` 복사 후 `//go:embed all:frontend`.

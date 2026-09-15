# cmd/server — Go 서버 진입점

포트 `:8080`. `make dev`로 실행 시 `-tags dev` 빌드.

## 파일

| 파일 | 역할 |
|------|------|
| `main.go` | 진입점 — `internal/app.Initialize()` 호출(DB/라이선스/OBS/YouTube/스케줄러/HTTP 서버) → 헬스체크 통과 시 `.bak` 정리 → 시그널 대기 |

embed 코드는 이 디렉터리에 없다. `internal/embedded` 가 `FrontendFS()`/`DataFS()` 를 제공하며 `-tags dev` 빌드에서는 nil 을 반환한다(Next.js dev server + 저장소 로컬 파일 사용). 초기화 로직은 `cmd/desktop` 과 공유하므로 여기서 직접 추가하지 말고 `internal/app/init.go` 에 넣을 것.

## ldflags

```makefile
-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)
```

태그 없이 빌드하면 `Version="dev"`.

## 프론트엔드 embed

프로덕션: `ui/out` → `internal/embedded/frontend/` 복사(`make embed-assets`) 후 `internal/embedded`가 `//go:embed all:frontend`. 서버/데스크톱이 같은 패키지를 공유.

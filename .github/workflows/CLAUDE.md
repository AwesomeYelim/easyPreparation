# .github/workflows — CI/CD

## release.yml

`v*` 태그 push → 자동 빌드 + GitHub Release.

- 4-job: `build-frontend` → `build-server`(4플랫폼) + `build-desktop`(3플랫폼) → `release`
- Server: CGO_ENABLED=0 크로스컴파일 (darwin arm64/amd64, linux, windows)
- Desktop: macOS arm64 (.dmg), Windows (NSIS setup .exe), Linux (binary)
- 8개 아티팩트 + checksums.txt

```bash
git tag v1.2.0 && git push origin v1.2.0
```

## test.yml

PR → `go vet` + `go build` (server만, desktop 제외).

## landing.yml

PR(`landing/**` 변경) → `npm ci` + `npm run build` 빌드 검사만. 배포는 하지 않는다.
실제 배포는 `master` push 시 Vercel Git 연동이 처리한다 (`landing/CLAUDE.md` 참고).

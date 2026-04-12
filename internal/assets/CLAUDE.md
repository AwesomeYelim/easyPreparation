# internal/assets — 에셋 다운로더

Oracle Cloud nginx(`138.2.119.220`)에서 찬송가/교독 PDF·PNG 다운로드 → 로컬 캐시.

## 캐시 경로

- 찬송 PDF: `data/pdf/hymn/`
- 교독 PDF: `data/pdf/responsive_reading/`
- PNG 캐시: `data/cache/hymn_pages/` (파일명: `{category}_{NNN}_{page}.png`)

## PNG 우선 다운로드 전략 (v1.2.x~)

찬송가·교독 PNG는 Oracle Cloud에 사전 변환 서빙 중.
**변환 불필요** → Windows 포함 전 플랫폼에서 동작.

```
서버 URL: /assets/hymn_pages/032/1.png, 2.png, ...
          /assets/responsive_reading_pages/001/1.png, ...
```

1. `DownloadPNGPages(category, filename, cacheDir)` — PNG 순서대로 다운로드, 404 시 중단
2. PNG 없으면 `DownloadPDF(category, filename, cacheDir)` → 로컬 변환 (Mac/Linux fallback)

## 호출처

- `internal/handlers/display.go` — `fetchDisplayImages()`
- `internal/presentation/presentation.go` — `setOutDirFiles()`

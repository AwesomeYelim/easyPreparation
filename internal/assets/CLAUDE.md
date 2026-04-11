# internal/assets — PDF 에셋 다운로더

Cloudflare R2에서 찬송가/교독 PDF 다운로드 → 로컬 캐시.

## 캐시 경로

- 찬송 PDF: `data/hymn/`
- 교독 PDF: `data/responsive_reading/`
- PNG 캐시: `data/cache/hymn_pages/`

## 동작

1. 로컬 캐시 확인
2. 없으면 R2 URL로 다운로드
3. Ghostscript로 PDF → PNG 변환 (`/opt/homebrew/bin/gs`)
4. 파일명 NFC 정규화 적용

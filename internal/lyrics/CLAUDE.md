# internal/lyrics — 찬양 PDF 생성

`/submitLyrics` 요청 → 가사 슬라이드 PDF 생성.

## 파일

- `lyricsPDF.go` — PDF 생성 메인 로직

## 주의

- `internal/presentation/` 래퍼 사용 (NFC 정규화)
- 가사 크롤링: bugs.co.kr
- 출력: `output/lyrics/*.pdf`

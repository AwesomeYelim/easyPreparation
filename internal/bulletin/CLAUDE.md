# internal/bulletin — 주보 PDF 생성

`/submit` 요청 → 프레젠테이션용 PDF 생성.

## 파일

| 파일 | 역할 |
|------|------|
| `bulletin.go` | 메인 오케스트레이터 |
| `define/define.go` | 주보 데이터 구조체 |
| `forPresentation/presentation.go` | 16:9 프레젠테이션 PDF |

## 주의

- `internal/presentation/` 래퍼 사용 (NFC 정규화)
- 찬송/교독 PDF → `internal/assets/` 경유 R2 다운로드
- 배경 이미지 (프레젠테이션): `data/templates/display/{title}.png`
- A4 인쇄용 PDF (forPrint) 패키지는 삭제됨

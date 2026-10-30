# landing — 홍보 랜딩 페이지

Next.js 14 정적 사이트. Vercel 배포.

## 페이지

- `/` — 홈 (기능 소개)
- `/pricing` — 요금제
- `/download` — 다운로드 + OS별 설치 안내

## 배포

- `master` push → Vercel Git 연동이 자동 빌드·배포 (저장소에 Vercel 설정 파일 없음, Vercel 대시보드에서 관리).
- `.github/workflows/landing.yml` 은 배포가 아니라 **PR 빌드 검사**만 한다 (`pull_request` + `landing/**` 경로). master 직접 push 에는 돌지 않는다.
- 배포 확인: `curl -s https://easy-preparation.vercel.app | grep <새 문구>` (응답 헤더 `server: Vercel`).

## URL

https://easy-preparation.vercel.app

# workers/license-api — CF Workers 라이선스 서버

Hono 기반. 토스페이먼츠 결제 + KV 스토리지.

## 파일

| 파일 | 역할 |
|------|------|
| `src/index.ts` | API 라우터 |
| `src/toss.ts` | 토스페이먼츠 REST API (결제 승인/취소/조회) |
| `src/webhook.ts` | 토스페이먼츠 Webhook 처리 |
| `src/kv.ts` | KV CRUD (`lic:` / `dev:` / `sess:` 키 네임스페이스) |
| `src/crypto.ts` | HMAC-SHA256 서명 + 라이선스 키 생성 |

## 결제 흐름

Go 앱 → `/api/checkout` → 토스페이먼츠 결제 → `/api/confirm` 승인 → 키 생성 → Go 앱 폴링


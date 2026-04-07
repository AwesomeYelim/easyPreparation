# 토스페이먼츠 가맹점 등록 + 키 설정 가이드

## 1. 가맹점 등록 절차

### 1.1 개발자 계정 생성

1. https://developers.tosspayments.com 접속
2. 회원가입 (이메일 또는 소셜 로그인)
3. 회사/개인 기본 정보 입력

### 1.2 사업자 정보 등록

1. 대시보드 > 사업자 정보 > 신규 등록
2. 다음 정보 준비:
   - 사업자등록증 (전자) 또는 이미지
   - 대표자명
   - 사업장 주소
   - 계약금 입금 계좌
3. 정보 입력 후 서류 업로드
4. 토스페이먼츠 자동 검증 (1-3영업일)

### 1.3 심사 및 승인

- 토스페이먼츠 내부 검증 진행 (1-3영업일)
- 결과는 등록 이메일로 발송
- 승인 후 Live 키 자동 발급 (테스트 키는 즉시 제공)

## 2. 테스트 키 vs Live 키

### 2.1 테스트 키

- **시크릿 키**: `test_gsk_XXXXXXXXXXX`
- **클라이언트 키**: `test_gck_XXXXXXXXXXX`
- 특징:
  - 실제 결제 처리 안 됨
  - 테스트 카드 사용 (4111111111111111 등)
  - 환경 설정 및 기능 검증 용도
  - 무제한 사용 가능

### 2.2 Live 키

- **시크릿 키**: `live_gsk_XXXXXXXXXXX`
- **클라이언트 키**: `live_gck_XXXXXXXXXXX`
- 특징:
  - 실제 고객 결제 처리
  - 선불금(보증금) 필요 (계약 체결 시 협의)
  - 사용 수수료 발생 (PG 수수료 ~3.0% + 부가세)
  - 승인 후 발급

## 3. CF Workers에 Live 키 설정

### 3.1 사전 준비

- 토스페이먼츠에서 Live 키 발급 완료 (승인 대기 중이면 진행 불가)
- Cloudflare 계정 + Workers 환경 준비 완료
- `wrangler` CLI 설치 (npm install -g wrangler)

### 3.2 시크릿 등록 절차

```bash
# 프로젝트 디렉토리 이동
cd /path/to/easyPreparation/workers/license-api

# 토스페이먼츠 시크릿 키 등록 (대화형)
wrangler secret put TOSS_SECRET_KEY
# 입력: live_gsk_XXXXXXXXXXX (또는 테스트 키)

# 토스페이먼츠 클라이언트 키 등록
wrangler secret put TOSS_CLIENT_KEY
# 입력: live_gck_XXXXXXXXXXX (또는 테스트 키)

# 라이선스 서버 HMAC 시크릿 등록 (기존 설정 유지)
wrangler secret put LICENSE_HMAC_SECRET
# 입력: your_hmac_secret_key
```

### 3.3 배포 및 확인

```bash
# Workers 배포
wrangler deploy

# 배포 완료 후 시크릿 확인 (실제 값은 보이지 않음)
wrangler secret list
```

**출력 예시:**
```
TOSS_SECRET_KEY   (set)
TOSS_CLIENT_KEY   (set)
LICENSE_HMAC_SECRET (set)
```

## 4. 정산 흐름 및 주기

### 4.1 결제에서 정산까지의 과정

```
사용자 결제 → 토스페이먼츠 승인
         ↓
    수수료 공제 (PG 수수료 ~3.0% + 부가세)
         ↓
가맹점 계좌 자동 입금 (D+2, 영업일 기준)
```

### 4.2 정산 관리

1. **정산 주기**: 매일 거래분을 D+2(영업일)에 입금
2. **정산 대시보드 접근**:
   - https://admin.tosspayments.com 로그인
   - 좌측 메뉴 > 정산
   - 기간별 거래액, 수수료, 실제 입금액 확인
3. **세금계산서**: 정산액에 대해 자동 발행 (월별)

### 4.3 정산 관련 주의사항

- 선불금(보증금)이 정산액보다 많으면 월말에 정산
- 환불/취소 거래는 정산액에서 차감
- 정산 계좌 변경은 대시보드에서 수행 (1회만 가능, 이후 문의)

## 5. 웹훅(Webhook) 설정

### 5.1 웹훅이란

- 결제 이벤트 발생 시 토스페이먼츠가 가맹점 서버로 즉시 알림
- `workers/license-api/src/webhook.ts` → 결제 성공/취소/환불 자동 처리

### 5.2 웹훅 URL 등록 절차

1. **Workers URL 확인**:
   ```bash
   wrangler deployments list
   # 또는 wrangler.toml에서 routes 확인
   # https://{account-name}.{workers-domain}/api/webhook
   ```

2. **토스페이먼츠 대시보드에서 등록**:
   - https://admin.tosspayments.com 로그인
   - 좌측 메뉴 > 개발자 > 웹훅
   - URL 추가: `https://{your-worker-url}/api/webhook`
   - 이벤트 선택:
     - `payment.completed` — 결제 성공
     - `payment.failed` — 결제 실패
     - `payment.cancelled` — 결제 취소
     - `refund.completed` — 환불 완료

3. **테스트 웹훅 발송** (대시보드):
   - 웹훅 목록에서 해당 URL 선택 > "테스트 발송"
   - 실제 이벤트가 아닌 더미 데이터로 연결성 검증

### 5.3 웹훅 시크릿 검증

- `src/webhook.ts`에서 X-Toss-Webhook-Signature 헤더 검증
- HMAC-SHA256 방식으로 서명 확인 (무결성 보장)
- 미인증 요청은 거부 (401 에러)

## 6. 배포 체크리스트

### 6.1 사전 준비 (테스트 단계)

- [ ] 토스페이먼츠 개발자 계정 생성
- [ ] 테스트 키 획득 (즉시)
- [ ] CF Workers에 테스트 키 등록
- [ ] 웹훅 URL 등록 + 테스트 발송 확인
- [ ] 테스트 환경에서 전체 결제 흐름 검증
  - 테스트 카드 결제
  - 라이선스 활성화 확인
  - 환불 테스트
  - 웹훅 이벤트 로그 확인

### 6.2 Live 배포 (실제 서비스)

- [ ] 사업자 정보 등록 + 서류 제출
- [ ] 토스페이먼츠 가맹점 심사 대기 (1-3영업일)
- [ ] Live 키 발급 받기
- [ ] CF Workers 시크릿 교체 (live_gsk_, live_gck_)
- [ ] 웹훅 URL이 프로덕션 Worker를 가리키는지 확인
- [ ] 소량 테스트 결제 (선불금으로 실제 정산 확인)
- [ ] 대시보드 정산 화면에서 입금액 확인
- [ ] 운영 담당자에게 알림 설정 (이메일/슬랙)

### 6.3 운영 중 확인 사항

- [ ] 일일 정산액 모니터링 (이상 징후 감지)
- [ ] 웹훅 실패 로그 확인 (재전송 필요 시)
- [ ] 환불/취소 거래 확인
- [ ] 월간 세금계산서 발행 자동화 (필요 시)

## 7. 문제 해결

### 7.1 결제 실패 원인 (Live 키 기준)

| 에러 코드 | 의미 | 해결 방법 |
|----------|------|---------|
| `CARD_DECLINED` | 카드사 거부 | 다른 카드 사용 |
| `INVALID_AMOUNT` | 결제 금액 오류 | 최소 100원, 최대 1천만원 확인 |
| `INVALID_MERCHANT` | 가맹점 미등록 | 가맹점 심사 완료 대기 |
| `FRAUD_SUSPECTED` | 의심 거래 | 토스페이먼츠 고객센터 문의 |

### 7.2 웹훅이 도착하지 않음

1. **Workers 로그 확인**:
   ```bash
   wrangler tail
   # 요청이 들어오는지 확인
   ```

2. **웹훅 URL 재확인**:
   - 토스페이먼츠 대시보드에 등록된 URL이 정확한지
   - HTTPS 필수 (HTTP 불가)

3. **신크레 환경**:
   - 토스페이먼츠: 테스트 키 → 테스트 서버로 웹훅 발송
   - 토스페이먼츠: Live 키 → 프로덕션 서버로 웹훅 발송
   - 키와 URL 환경이 일치하는지 확인

### 7.3 정산이 안 됨

1. **선불금(보증금) 부족**:
   - 토스페이먼츠 대시보드 > 자금 관리 > 선불금 확인
   - 부족하면 충전 (계약 체결 필요)

2. **정산 계좌 오류**:
   - 대시보드에서 계좌 정보 확인
   - 등록된 계좌로 소액 테스트 송금 (환입 후 재시도)

3. **보류된 거래**:
   - 대시보드 > 정산 > 보류 현황 확인
   - 토스페이먼츠 고객센터 문의

## 8. 참고 자료

- **공식 문서**: https://docs.tosspayments.com
- **개발자 대시보드**: https://developers.tosspayments.com
- **결제 관리**: https://admin.tosspayments.com
- **REST API 레퍼런스**: https://docs.tosspayments.com/reference
- **샘플 코드**: https://github.com/toss/payment-sdk
- **고객센터**: https://toss.im/support (평일 10:00-18:00, 유료 기술 지원)

## 9. 추가 자료

### easyPreparation 구현체

| 파일 | 역할 |
|------|------|
| `workers/license-api/src/index.ts` | 결제 API 라우터 (checkout, confirm, callback 등) |
| `workers/license-api/src/toss.ts` | 토스페이먼츠 REST API 래퍼 (승인/취소/조회) |
| `workers/license-api/src/webhook.ts` | 웹훅 이벤트 핸들러 |
| `workers/license-api/src/kv.ts` | KV 스토리지 CRUD (거래 기록) |
| `internal/license/manager.go` | Go 라이선스 매니저 (활성화/검증) |
| `ui/app/components/LicensePanel.tsx` | 결제 UI (모달/로딩/success) |

### API 엔드포인트 (CF Workers)

| 메서드 | 경로 | 설명 |
|--------|------|------|
| POST | `/api/checkout` | 결제 세션 생성 → `checkoutUrl` |
| POST | `/api/confirm` | 결제 승인 (클라이언트 → 서버) |
| POST | `/api/activate` | 라이선스 활성화 (Go 앱 폴링) |
| POST | `/api/verify` | 라이선스 유효성 검증 |
| POST | `/api/portal` | 구독 정보 조회 (정산액, 다음 결제 등) |
| POST | `/api/cancel` | 구독 취소 (환불 처리) |
| POST | `/api/webhook` | 웹훅 이벤트 수신 (토스페이먼츠 → CF) |

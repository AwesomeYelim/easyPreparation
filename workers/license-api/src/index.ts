/**
 * easyPreparation 라이선스 서버
 * Cloudflare Workers + Hono
 *
 * 엔드포인트:
 *   POST /api/checkout         - 주문 생성 → 결제 페이지 URL 반환
 *   GET  /checkout/:orderId    - 토스페이먼츠 SDK 결제 페이지 (HTML)
 *   GET  /api/confirm          - 결제 성공 콜백 → 토스 승인 → 라이선스 발급
 *   GET  /checkout/success     - 결제 완료 안내 페이지 (HTML)
 *   GET  /checkout/fail        - 결제 실패 안내 페이지 (HTML)
 *   POST /api/activate         - 폴링: orderId로 라이선스 키 조회
 *   POST /api/verify           - 24시간 주기 라이선스 검증
 *   POST /api/portal           - 결제 정보 조회 + 취소 기능
 *   POST /api/webhook          - 토스페이먼츠 웹훅 수신
 */

import { Hono } from 'hono'
import { cors } from 'hono/cors'
import { logger } from 'hono/logger'

import type {
  Env,
  CheckoutRequest,
  ActivateRequest,
  VerifyRequest,
  PortalRequest,
  ErrorResponse,
} from './types'
import { confirmPayment, getPayment, cancelPayment } from './toss'
import { handleTossWebhook } from './webhook'
import {
  getSession,
  setSession,
  getLicense,
  setLicense,
  getDeviceLicense,
  setDeviceLicense,
  touchLicense,
} from './kv'
import {
  generateLicenseKey,
  generateOrderId,
  signLicense,
  verifySignature,
  isValidLicenseKeyFormat,
  calculateExpiresAt,
} from './crypto'

const app = new Hono<{ Bindings: Env }>()

// ─── 미들웨어 ─────────────────────────────────────────────────────────────────

app.use('*', logger())

app.use(
  '/api/*',
  cors({
    origin: '*', // easyPreparation 앱 → 모든 로컬 origin 허용
    allowMethods: ['GET', 'POST', 'OPTIONS'],
    allowHeaders: ['Content-Type'],
  }),
)

// ─── 헬스체크 ─────────────────────────────────────────────────────────────────

app.get('/', (c) => c.json({ status: 'ok', service: 'easyprep-license-api' }))

// ─── 공통 HTML 레이아웃 ───────────────────────────────────────────────────────

function htmlLayout(title: string, bodyContent: string): string {
  return `<!DOCTYPE html>
<html lang="ko">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>${title}</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      background: #f8f9fa;
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 1rem;
    }
    .card {
      background: #fff;
      border-radius: 12px;
      box-shadow: 0 2px 16px rgba(0,0,0,0.08);
      padding: 2rem;
      width: 100%;
      max-width: 480px;
    }
    h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 0.5rem; color: #1a1a1a; }
    p { color: #555; line-height: 1.6; margin-top: 0.5rem; }
    .badge {
      display: inline-block;
      padding: 0.25rem 0.75rem;
      border-radius: 999px;
      font-size: 0.85rem;
      font-weight: 600;
      margin-bottom: 1rem;
    }
    .badge-success { background: #e6f9f0; color: #1a7f4b; }
    .badge-error { background: #fde8e8; color: #c0392b; }
    .close-hint {
      margin-top: 1.5rem;
      font-size: 0.875rem;
      color: #888;
      text-align: center;
    }
    #payment-widget { margin-top: 1rem; }
    .plan-info {
      background: #f0f4ff;
      border-radius: 8px;
      padding: 1rem;
      margin-bottom: 1rem;
    }
    .plan-info strong { display: block; font-size: 1.1rem; color: #2c3e50; }
    .plan-info span { font-size: 0.95rem; color: #555; }
    .error-msg { color: #c0392b; margin-top: 0.5rem; font-size: 0.9rem; }
  </style>
</head>
<body>
  <div class="card">
    ${bodyContent}
  </div>
</body>
</html>`
}

// ─── POST /api/checkout ───────────────────────────────────────────────────────

app.post('/api/checkout', async (c) => {
  let body: CheckoutRequest

  try {
    body = await c.req.json<CheckoutRequest>()
  } catch {
    return c.json<ErrorResponse>({ error: '잘못된 요청 형식입니다.' }, 400)
  }

  const { deviceId, plan } = body

  if (!deviceId || !plan) {
    return c.json<ErrorResponse>(
      { error: 'deviceId와 plan은 필수입니다.' },
      400,
    )
  }

  if (plan !== 'pro_monthly' && plan !== 'pro_annual') {
    return c.json<ErrorResponse>(
      {
        error:
          '유효하지 않은 플랜입니다. pro_monthly 또는 pro_annual을 사용하세요.',
      },
      400,
    )
  }

  // 이미 활성 Pro 라이선스가 있는 디바이스 확인
  const existingLicenseKey = await getDeviceLicense(c.env.LICENSE_KV, deviceId)
  if (existingLicenseKey) {
    const existingLicense = await getLicense(c.env.LICENSE_KV, existingLicenseKey)
    if (
      existingLicense &&
      existingLicense.plan === 'pro' &&
      existingLicense.subscriptionStatus === 'active'
    ) {
      return c.json<ErrorResponse>(
        {
          error: '이미 활성 Pro 라이선스가 있습니다.',
          code: 'already_subscribed',
        },
        409,
      )
    }
  }

  // 금액 계산
  const amount = plan === 'pro_annual' ? 99000 : 9900

  // orderId 생성 (EP-yyyyMMddHHmmss-XXXX)
  const orderId = generateOrderId()

  // Worker URL 추출 (결제 페이지 URL 구성용)
  const workerUrl = new URL(c.req.url).origin

  // KV에 pending 세션 저장 (TTL: 1시간)
  await setSession(
    c.env.LICENSE_KV,
    orderId,
    { deviceId, plan, status: 'pending', amount, orderId },
    3600,
  )

  const checkoutUrl = `${workerUrl}/checkout/${orderId}`

  return c.json({ checkoutUrl, sessionId: orderId })
})

// ─── GET /checkout/success ────────────────────────────────────────────────────
// 주의: /checkout/:orderId 보다 먼저 등록해야 "success"가 파라미터로 잡히지 않음

app.get('/checkout/success', (c) => {
  const html = htmlLayout(
    '결제 완료 — easyPreparation',
    `<span class="badge badge-success">결제 완료</span>
    <h1>Pro 플랜이 활성화되었습니다!</h1>
    <p>easyPreparation Pro 기능이 활성화되었습니다.<br />앱으로 돌아가면 자동으로 Pro 기능을 사용할 수 있습니다.</p>
    <p class="close-hint">이 창을 닫아주세요.</p>`,
  )
  return c.html(html)
})

// ─── GET /checkout/fail ───────────────────────────────────────────────────────

app.get('/checkout/fail', (c) => {
  const errorMsg = c.req.query('message') || '결제가 완료되지 않았습니다.'
  const errorCode = c.req.query('code') || ''

  const html = htmlLayout(
    '결제 실패 — easyPreparation',
    `<span class="badge badge-error">결제 실패</span>
    <h1>결제에 실패했습니다</h1>
    <p class="error-msg">${errorMsg}${errorCode ? ` (${errorCode})` : ''}</p>
    <p>앱으로 돌아가서 다시 시도해주세요.</p>
    <p class="close-hint">이 창을 닫아주세요.</p>`,
  )
  return c.html(html)
})

// ─── GET /checkout/:orderId ───────────────────────────────────────────────────

app.get('/checkout/:orderId', async (c) => {
  const orderId = c.req.param('orderId')
  const workerUrl = new URL(c.req.url).origin

  // KV에서 세션 조회
  const session = await getSession(c.env.LICENSE_KV, orderId)

  if (!session) {
    const html = htmlLayout(
      '주문 오류 — easyPreparation',
      `<span class="badge badge-error">오류</span>
      <h1>주문을 찾을 수 없습니다</h1>
      <p>세션이 만료되었거나 유효하지 않은 주문입니다.<br />앱에서 다시 시도해주세요.</p>`,
    )
    return c.html(html, 404)
  }

  if (session.status === 'completed') {
    return c.redirect(`${workerUrl}/checkout/success`)
  }

  const planLabel =
    session.plan === 'pro_annual'
      ? 'easyPreparation Pro (연간)'
      : 'easyPreparation Pro (월간)'

  const amountFormatted = session.amount.toLocaleString('ko-KR')

  const html = htmlLayout(
    '결제 — easyPreparation Pro',
    `<div class="plan-info">
      <strong>${planLabel}</strong>
      <span>₩${amountFormatted}</span>
    </div>
    <div id="payment-widget"></div>
    <script src="https://js.tosspayments.com/v2/standard"></script>
    <script>
      (async () => {
        try {
          const tossPayments = TossPayments('${c.env.TOSS_CLIENT_KEY}');
          const widgets = tossPayments.widgets({
            customerKey: '${session.deviceId}',
          });

          await widgets.setAmount({
            currency: 'KRW',
            value: ${session.amount},
          });

          await widgets.renderPaymentMethods({
            selector: '#payment-widget',
            variantKey: 'DEFAULT',
          });

          // 결제 요청 버튼 동적 생성
          const btn = document.createElement('button');
          btn.textContent = '₩${amountFormatted} 결제하기';
          btn.style.cssText = [
            'display:block', 'width:100%', 'margin-top:1rem',
            'padding:0.875rem', 'background:#0064FF', 'color:#fff',
            'border:none', 'border-radius:8px', 'font-size:1rem',
            'font-weight:600', 'cursor:pointer',
          ].join(';');
          btn.addEventListener('click', async () => {
            btn.disabled = true;
            btn.textContent = '처리 중...';
            try {
              await widgets.requestPayment({
                orderId: '${orderId}',
                orderName: '${planLabel}',
                successUrl: '${workerUrl}/api/confirm',
                failUrl: '${workerUrl}/checkout/fail',
              });
            } catch (e) {
              btn.disabled = false;
              btn.textContent = '₩${amountFormatted} 결제하기';
              console.error('결제 오류:', e);
            }
          });

          document.querySelector('.card').appendChild(btn);
        } catch (e) {
          document.getElementById('payment-widget').innerHTML =
            '<p style="color:#c0392b">결제 위젯 로드에 실패했습니다. 잠시 후 다시 시도해주세요.</p>';
          console.error('위젯 초기화 오류:', e);
        }
      })();
    </script>`,
  )

  return c.html(html)
})

// ─── GET /api/confirm ─────────────────────────────────────────────────────────
// 토스페이먼츠 successUrl 리다이렉트 처리

app.get('/api/confirm', async (c) => {
  const paymentKey = c.req.query('paymentKey')
  const orderId = c.req.query('orderId')
  const amountStr = c.req.query('amount')
  const workerUrl = new URL(c.req.url).origin

  if (!paymentKey || !orderId || !amountStr) {
    return c.redirect(
      `${workerUrl}/checkout/fail?message=${encodeURIComponent('필수 파라미터가 누락되었습니다.')}`,
    )
  }

  const amount = parseInt(amountStr, 10)
  if (isNaN(amount)) {
    return c.redirect(
      `${workerUrl}/checkout/fail?message=${encodeURIComponent('유효하지 않은 금액입니다.')}`,
    )
  }

  // KV에서 세션 조회
  const session = await getSession(c.env.LICENSE_KV, orderId)

  if (!session) {
    return c.redirect(
      `${workerUrl}/checkout/fail?message=${encodeURIComponent('세션을 찾을 수 없거나 만료되었습니다.')}`,
    )
  }

  // 이중 결제 방지: 이미 completed 상태면 success 페이지로
  if (session.status === 'completed') {
    return c.redirect(`${workerUrl}/checkout/success`)
  }

  // 금액 일치 확인
  if (session.amount !== amount) {
    console.error(
      `[/api/confirm] 금액 불일치: expected=${session.amount}, actual=${amount}`,
    )
    return c.redirect(
      `${workerUrl}/checkout/fail?message=${encodeURIComponent('결제 금액이 일치하지 않습니다.')}`,
    )
  }

  // 토스페이먼츠 결제 승인 API 호출
  try {
    await confirmPayment(c.env, paymentKey, orderId, amount)
  } catch (err) {
    const errMsg = err instanceof Error ? err.message : '결제 승인 실패'
    console.error(`[/api/confirm] 결제 승인 실패 (orderId: ${orderId}):`, err)
    return c.redirect(
      `${workerUrl}/checkout/fail?message=${encodeURIComponent(errMsg)}`,
    )
  }

  // 라이선스 발급
  const expiresAt = calculateExpiresAt(session.plan)
  const licenseKey = generateLicenseKey()
  const signature = await signLicense(
    licenseKey,
    session.deviceId,
    'pro',
    expiresAt,
    c.env.HMAC_SECRET,
  )

  // KV에 라이선스 저장
  await Promise.all([
    setLicense(c.env.LICENSE_KV, licenseKey, {
      plan: 'pro',
      deviceId: session.deviceId,
      paymentKey,
      customerKey: session.deviceId,
      subscriptionStatus: 'active',
      expiresAt,
      issuedAt: new Date().toISOString(),
      lastVerified: new Date().toISOString(),
      amount,
    }),
    setDeviceLicense(c.env.LICENSE_KV, session.deviceId, licenseKey),
  ])

  // 세션 상태를 "completed"로 갱신 (TTL: 1시간 유지)
  await setSession(
    c.env.LICENSE_KV,
    orderId,
    {
      ...session,
      status: 'completed',
      licenseKey,
      expiresAt,
      signature,
    },
    3600,
  )

  console.log(
    `[/api/confirm] 라이선스 발급 완료: ${licenseKey} (device: ${session.deviceId})`,
  )

  return c.redirect(`${workerUrl}/checkout/success`)
})

// ─── POST /api/activate ───────────────────────────────────────────────────────

app.post('/api/activate', async (c) => {
  let body: ActivateRequest

  try {
    body = await c.req.json<ActivateRequest>()
  } catch {
    return c.json<ErrorResponse>({ error: '잘못된 요청 형식입니다.' }, 400)
  }

  const { sessionId, deviceId } = body

  if (!sessionId || !deviceId) {
    return c.json<ErrorResponse>(
      { error: 'sessionId와 deviceId는 필수입니다.' },
      400,
    )
  }

  const session = await getSession(c.env.LICENSE_KV, sessionId)

  if (!session) {
    return c.json<ErrorResponse>(
      {
        error:
          '세션을 찾을 수 없습니다. 만료되었거나 유효하지 않은 세션입니다.',
        code: 'session_not_found',
      },
      404,
    )
  }

  // deviceId 일치 확인 (다른 디바이스의 세션 도용 방지)
  if (session.deviceId !== deviceId) {
    return c.json<ErrorResponse>(
      { error: '디바이스 ID가 일치하지 않습니다.', code: 'device_mismatch' },
      403,
    )
  }

  if (session.status === 'pending') {
    return c.json({ status: 'pending' })
  }

  // completed 상태
  return c.json({
    status: 'completed',
    plan: 'pro',
    licenseKey: session.licenseKey,
    expiresAt: session.expiresAt,
    signature: session.signature,
  })
})

// ─── POST /api/verify ─────────────────────────────────────────────────────────

app.post('/api/verify', async (c) => {
  let body: VerifyRequest

  try {
    body = await c.req.json<VerifyRequest>()
  } catch {
    return c.json<ErrorResponse>({ error: '잘못된 요청 형식입니다.' }, 400)
  }

  const { licenseKey, deviceId, signature } = body

  if (!licenseKey || !deviceId || !signature) {
    return c.json<ErrorResponse>(
      { error: 'licenseKey, deviceId, signature는 필수입니다.' },
      400,
    )
  }

  // 라이선스 키 형식 검사
  if (!isValidLicenseKeyFormat(licenseKey)) {
    return c.json<ErrorResponse>(
      { error: '유효하지 않은 라이선스 키 형식입니다.' },
      400,
    )
  }

  // KV에서 라이선스 조회
  const licenseData = await getLicense(c.env.LICENSE_KV, licenseKey)

  if (!licenseData) {
    return c.json({ valid: false, message: '라이선스를 찾을 수 없습니다.' }, 404)
  }

  // deviceId 일치 확인
  if (licenseData.deviceId !== deviceId) {
    return c.json(
      {
        valid: false,
        message: '다른 디바이스에 등록된 라이선스입니다.',
        code: 'device_mismatch',
      },
      403,
    )
  }

  // HMAC 서명 검증
  const isSignatureValid = await verifySignature(
    licenseKey,
    deviceId,
    licenseData.plan,
    licenseData.expiresAt,
    signature,
    c.env.HMAC_SECRET,
  )

  if (!isSignatureValid) {
    return c.json({ valid: false, message: '서명이 유효하지 않습니다.' }, 403)
  }

  // 구독 상태 확인
  const { subscriptionStatus, expiresAt, plan } = licenseData
  const now = new Date()
  const expiry = new Date(expiresAt)

  // canceled 상태이거나 만료된 경우
  if (subscriptionStatus === 'canceled' && expiry < now) {
    await touchLicense(c.env.LICENSE_KV, licenseKey, {})
    return c.json({
      valid: false,
      plan: 'free',
      expiresAt,
      subscriptionStatus,
      message: '구독이 취소되었습니다.',
    })
  }

  // past_due 상태: grace period 허용 (30일)
  if (subscriptionStatus === 'past_due') {
    const gracePeriodEnd = new Date(expiry.getTime() + 30 * 24 * 60 * 60 * 1000)
    if (now > gracePeriodEnd) {
      await touchLicense(c.env.LICENSE_KV, licenseKey, {})
      return c.json({
        valid: false,
        plan: 'free',
        expiresAt,
        subscriptionStatus,
        message: '결제가 실패했습니다. 결제 수단을 확인해주세요.',
      })
    }
    // grace period 내에는 valid로 처리
  }

  // lastVerified 업데이트
  await touchLicense(c.env.LICENSE_KV, licenseKey, {})

  return c.json({
    valid: true,
    plan,
    expiresAt,
    subscriptionStatus,
  })
})

// ─── POST /api/portal ─────────────────────────────────────────────────────────
// 결제 정보 조회 + 취소 기능 (토스페이먼츠에는 Stripe Customer Portal 없음)

app.post('/api/portal', async (c) => {
  let body: PortalRequest

  try {
    body = await c.req.json<PortalRequest>()
  } catch {
    return c.json<ErrorResponse>({ error: '잘못된 요청 형식입니다.' }, 400)
  }

  const { licenseKey, deviceId } = body

  if (!licenseKey || !deviceId) {
    return c.json<ErrorResponse>(
      { error: 'licenseKey와 deviceId는 필수입니다.' },
      400,
    )
  }

  const licenseData = await getLicense(c.env.LICENSE_KV, licenseKey)

  if (!licenseData) {
    return c.json<ErrorResponse>({ error: '라이선스를 찾을 수 없습니다.' }, 404)
  }

  if (licenseData.deviceId !== deviceId) {
    return c.json<ErrorResponse>(
      { error: '디바이스 ID가 일치하지 않습니다.' },
      403,
    )
  }

  // paymentKey가 없으면 결제 정보 없음
  if (!licenseData.paymentKey) {
    return c.json({
      paymentInfo: null,
      message: '결제 정보가 없습니다.',
    })
  }

  // 토스페이먼츠 결제 정보 조회
  try {
    const paymentInfo = await getPayment(c.env, licenseData.paymentKey)
    return c.json({
      paymentInfo,
      cancelUrl: null, // 취소는 앱 내에서 직접 처리
    })
  } catch (err) {
    console.error('[/api/portal] 결제 정보 조회 실패:', err)
    return c.json<ErrorResponse>(
      { error: '결제 정보 조회에 실패했습니다.' },
      500,
    )
  }
})

// ─── POST /api/cancel ─────────────────────────────────────────────────────────
// 구독 취소 (결제 취소 = 자동갱신 중지)

app.post('/api/cancel', async (c) => {
  let body: PortalRequest

  try {
    body = await c.req.json<PortalRequest>()
  } catch {
    return c.json<ErrorResponse>({ error: '잘못된 요청 형식입니다.' }, 400)
  }

  const { licenseKey, deviceId } = body

  if (!licenseKey || !deviceId) {
    return c.json<ErrorResponse>(
      { error: 'licenseKey와 deviceId는 필수입니다.' },
      400,
    )
  }

  const licenseData = await getLicense(c.env.LICENSE_KV, licenseKey)

  if (!licenseData) {
    return c.json<ErrorResponse>({ error: '라이선스를 찾을 수 없습니다.' }, 404)
  }

  if (licenseData.deviceId !== deviceId) {
    return c.json<ErrorResponse>(
      { error: '디바이스 ID가 일치하지 않습니다.' },
      403,
    )
  }

  if (!licenseData.paymentKey) {
    return c.json<ErrorResponse>(
      { error: '취소할 결제 정보가 없습니다.' },
      400,
    )
  }

  try {
    await cancelPayment(c.env, licenseData.paymentKey, '사용자 구독 취소')
    await touchLicense(c.env.LICENSE_KV, licenseKey, {
      subscriptionStatus: 'canceled',
    })

    return c.json({ success: true, message: '구독이 취소되었습니다.' })
  } catch (err) {
    console.error('[/api/cancel] 결제 취소 실패:', err)
    return c.json<ErrorResponse>({ error: '구독 취소에 실패했습니다.' }, 500)
  }
})

// ─── POST /api/webhook ────────────────────────────────────────────────────────

app.post('/api/webhook', async (c) => {
  return handleTossWebhook(c.env, c.req.raw)
})

// ─── 404 핸들러 ───────────────────────────────────────────────────────────────

app.notFound((c) =>
  c.json<ErrorResponse>({ error: '요청한 경로를 찾을 수 없습니다.' }, 404),
)

// ─── 전역 에러 핸들러 ─────────────────────────────────────────────────────────

app.onError((err, c) => {
  console.error('[Global Error]', err)
  return c.json<ErrorResponse>(
    { error: '서버 내부 오류가 발생했습니다.' },
    500,
  )
})

export default app

/**
 * 토스페이먼츠 웹훅 처리
 *
 * 토스페이먼츠 웹훅 이벤트:
 *   - DONE          : 결제 완료
 *   - CANCELED      : 결제 취소
 *   - PARTIAL_CANCELED: 부분 취소
 *
 * MVP에서는 결제 승인을 /api/confirm에서 직접 처리하므로
 * 웹훅은 취소/환불 알림 수신용 및 향후 정기결제 실패 알림에 활용.
 *
 * 토스페이먼츠 웹훅 서명 검증:
 *   - 헤더: Toss-Payments-Signature (HMAC-SHA256)
 *   - 검증 키: 토스페이먼츠 웹훅 시크릿 (대시보드에서 발급)
 */

import type { Env } from './types'
import { touchLicense } from './kv'

/**
 * 토스페이먼츠 웹훅 이벤트 처리
 */
export async function handleTossWebhook(
  env: Env,
  request: Request,
): Promise<Response> {
  let body: Record<string, unknown>

  try {
    body = (await request.json()) as Record<string, unknown>
  } catch {
    console.error('[Webhook] 요청 본문 파싱 실패')
    return new Response('Bad Request', { status: 400 })
  }

  const eventType = body.eventType as string | undefined
  console.log(`[Webhook] 토스페이먼츠 이벤트 수신: ${eventType}`)

  switch (eventType) {
    case 'DONE':
      // 결제 완료 — /api/confirm에서 이미 처리됨, 중복 방지
      console.log('[Webhook] DONE: /api/confirm에서 처리 완료, 웹훅 스킵')
      break

    case 'CANCELED':
    case 'PARTIAL_CANCELED':
      await handlePaymentCanceled(env, body)
      break

    default:
      console.log(`[Webhook] 미처리 이벤트 무시: ${eventType}`)
  }

  return new Response('OK', { status: 200 })
}

/**
 * 결제 취소 이벤트 처리
 * paymentKey로 라이선스 조회 후 구독 상태를 'canceled'로 변경
 */
async function handlePaymentCanceled(
  env: Env,
  body: Record<string, unknown>,
): Promise<void> {
  const data = body.data as Record<string, unknown> | undefined
  if (!data) {
    console.error('[Webhook] CANCELED: data 필드 없음')
    return
  }

  // 토스페이먼츠 웹훅은 orderId 포함
  const orderId = data.orderId as string | undefined
  if (!orderId) {
    console.error('[Webhook] CANCELED: orderId 없음')
    return
  }

  // orderId로 KV에서 세션 조회 후 라이선스 취소 처리
  try {
    const { getSession } = await import('./kv')
    const session = await getSession(env.LICENSE_KV, orderId)

    if (!session || !session.licenseKey) {
      console.log(`[Webhook] CANCELED: orderId(${orderId})에 해당하는 라이선스 없음`)
      return
    }

    await touchLicense(env.LICENSE_KV, session.licenseKey, {
      subscriptionStatus: 'canceled',
      expiresAt: new Date().toISOString(),
    })

    console.log(`[Webhook] 결제 취소 처리: ${session.licenseKey} → canceled`)
  } catch (err) {
    console.error('[Webhook] CANCELED 처리 실패:', err)
  }
}

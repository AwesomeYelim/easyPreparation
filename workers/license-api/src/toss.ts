/**
 * 토스페이먼츠 API 래퍼
 *
 * Cloudflare Workers 환경에서 토스페이먼츠 API를 호출하기 위한 헬퍼.
 * fetch 기반으로 직접 구현 (별도 SDK 없음).
 */

import type { Env } from './types'

const TOSS_API_URL = 'https://api.tosspayments.com/v1'

/**
 * 토스페이먼츠 Basic 인증 헤더 생성
 * Base64(secretKey + ':')
 */
function authHeader(secretKey: string): string {
  return 'Basic ' + btoa(secretKey + ':')
}

/**
 * 결제 승인
 * POST /v1/payments/confirm
 */
export async function confirmPayment(
  env: Env,
  paymentKey: string,
  orderId: string,
  amount: number,
): Promise<Record<string, unknown>> {
  const res = await fetch(`${TOSS_API_URL}/payments/confirm`, {
    method: 'POST',
    headers: {
      Authorization: authHeader(env.TOSS_SECRET_KEY),
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ paymentKey, orderId, amount }),
  })

  if (!res.ok) {
    const err = (await res.json()) as { message?: string; code?: string }
    throw new Error(err.message || '결제 승인 실패')
  }

  return res.json() as Promise<Record<string, unknown>>
}

/**
 * 결제 취소
 * POST /v1/payments/{paymentKey}/cancel
 */
export async function cancelPayment(
  env: Env,
  paymentKey: string,
  cancelReason: string,
): Promise<Record<string, unknown>> {
  const res = await fetch(`${TOSS_API_URL}/payments/${paymentKey}/cancel`, {
    method: 'POST',
    headers: {
      Authorization: authHeader(env.TOSS_SECRET_KEY),
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ cancelReason }),
  })

  return res.json() as Promise<Record<string, unknown>>
}

/**
 * 결제 조회
 * GET /v1/payments/{paymentKey}
 */
export async function getPayment(
  env: Env,
  paymentKey: string,
): Promise<Record<string, unknown> | null> {
  try {
    const res = await fetch(`${TOSS_API_URL}/payments/${paymentKey}`, {
      method: 'GET',
      headers: {
        Authorization: authHeader(env.TOSS_SECRET_KEY),
      },
    })

    if (!res.ok) return null
    return res.json() as Promise<Record<string, unknown>>
  } catch (err) {
    console.error(`[Toss] 결제 조회 실패 (${paymentKey}):`, err)
    return null
  }
}

/**
 * 빌링키 발급 (자동결제 등록용)
 * POST /v1/billing/authorizations/issue
 */
export async function issueBillingKey(
  env: Env,
  authKey: string,
  customerKey: string,
): Promise<Record<string, unknown>> {
  const res = await fetch(`${TOSS_API_URL}/billing/authorizations/issue`, {
    method: 'POST',
    headers: {
      Authorization: authHeader(env.TOSS_SECRET_KEY),
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ authKey, customerKey }),
  })

  if (!res.ok) {
    const err = (await res.json()) as { message?: string; code?: string }
    throw new Error(err.message || '빌링키 발급 실패')
  }

  return res.json() as Promise<Record<string, unknown>>
}

/**
 * 빌링키로 자동결제 실행
 * POST /v1/billing/{billingKey}
 */
export async function executeBilling(
  env: Env,
  billingKey: string,
  customerKey: string,
  amount: number,
  orderId: string,
  orderName: string,
): Promise<Record<string, unknown>> {
  const res = await fetch(`${TOSS_API_URL}/billing/${billingKey}`, {
    method: 'POST',
    headers: {
      Authorization: authHeader(env.TOSS_SECRET_KEY),
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ customerKey, amount, orderId, orderName }),
  })

  return res.json() as Promise<Record<string, unknown>>
}

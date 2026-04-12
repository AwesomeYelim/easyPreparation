/**
 * 라이선스 키 생성 및 HMAC-SHA256 서명/검증
 *
 * Go 앱의 internal/license/keygen.go 와 동일한 알고리즘 사용
 * 서명 메시지 형식: {licenseKey}:{deviceId}:{plan}:{expiresAt}
 */

const LICENSE_CHARSET = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
const SEGMENT_LENGTH = 4
const SEGMENT_COUNT = 4

/**
 * EP-XXXX-XXXX-XXXX-XXXX 형식의 라이선스 키 생성
 */
export function generateLicenseKey(): string {
  const segments: string[] = []

  for (let s = 0; s < SEGMENT_COUNT; s++) {
    let segment = ''
    const randomBytes = new Uint8Array(SEGMENT_LENGTH)
    crypto.getRandomValues(randomBytes)

    for (let i = 0; i < SEGMENT_LENGTH; i++) {
      segment += LICENSE_CHARSET[randomBytes[i] % LICENSE_CHARSET.length]
    }
    segments.push(segment)
  }

  return `EP-${segments.join('-')}`
}

/**
 * 라이선스 키 형식 유효성 검사
 */
export function isValidLicenseKeyFormat(key: string): boolean {
  const pattern = /^EP-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$/
  return pattern.test(key)
}

/**
 * 서명 메시지 구성
 */
function buildSignatureMessage(
  licenseKey: string,
  deviceId: string,
  plan: string,
  expiresAt: string,
): string {
  return `${licenseKey}:${deviceId}:${plan}:${expiresAt}`
}

/**
 * HMAC-SHA256 서명 생성
 */
export async function signLicense(
  licenseKey: string,
  deviceId: string,
  plan: string,
  expiresAt: string,
  secret: string,
): Promise<string> {
  const message = buildSignatureMessage(licenseKey, deviceId, plan, expiresAt)

  const keyMaterial = await crypto.subtle.importKey(
    'raw',
    new TextEncoder().encode(secret),
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign'],
  )

  const signature = await crypto.subtle.sign(
    'HMAC',
    keyMaterial,
    new TextEncoder().encode(message),
  )

  // hex string으로 변환
  return Array.from(new Uint8Array(signature))
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('')
}

/**
 * HMAC-SHA256 서명 검증
 */
export async function verifySignature(
  licenseKey: string,
  deviceId: string,
  plan: string,
  expiresAt: string,
  signature: string,
  secret: string,
): Promise<boolean> {
  try {
    const expectedSignature = await signLicense(
      licenseKey,
      deviceId,
      plan,
      expiresAt,
      secret,
    )

    // timing-safe 비교 (길이가 다르면 false)
    if (expectedSignature.length !== signature.length) {
      return false
    }

    // 문자별 XOR 비교 (timing-safe)
    let diff = 0
    for (let i = 0; i < expectedSignature.length; i++) {
      diff |=
        expectedSignature.charCodeAt(i) ^ signature.toLowerCase().charCodeAt(i)
    }

    return diff === 0
  } catch {
    return false
  }
}

/**
 * 구독 주기에 따른 만료일 계산
 */
export function calculateExpiresAt(plan: 'pro_monthly' | 'pro_annual'): string {
  const now = new Date()

  if (plan === 'pro_annual') {
    now.setFullYear(now.getFullYear() + 1)
  } else {
    now.setMonth(now.getMonth() + 1)
  }

  return now.toISOString()
}

/**
 * orderId 생성
 * 형식: EP-{yyyyMMddHHmmss}-{4자리랜덤}
 */
export function generateOrderId(): string {
  const now = new Date()
  const pad = (n: number, len = 2) => String(n).padStart(len, '0')
  const datePart =
    now.getFullYear().toString() +
    pad(now.getMonth() + 1) +
    pad(now.getDate()) +
    pad(now.getHours()) +
    pad(now.getMinutes()) +
    pad(now.getSeconds())

  const randomBytes = new Uint8Array(2)
  crypto.getRandomValues(randomBytes)
  const random = Array.from(randomBytes)
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('')
    .toUpperCase()
    .slice(0, 4)

  return `EP-${datePart}-${random}`
}

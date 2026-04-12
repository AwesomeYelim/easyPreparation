/**
 * Cloudflare KV 스토리지 CRUD 헬퍼
 *
 * 키 네임스페이스:
 *   lic:{licenseKey}     → LicenseData (영구 저장)
 *   dev:{deviceId}       → licenseKey  (역참조, 영구)
 *   sess:{sessionId}     → SessionData (TTL: 3600초)
 */

import type { LicenseData, SessionData } from './types'

// ─── 라이선스 CRUD ──────────────────────────────────────────────────────────

export async function getLicense(
  kv: KVNamespace,
  licenseKey: string,
): Promise<LicenseData | null> {
  try {
    const raw = await kv.get(`lic:${licenseKey}`)
    if (!raw) return null
    return JSON.parse(raw) as LicenseData
  } catch (err) {
    console.error(`[KV] getLicense error (${licenseKey}):`, err)
    return null
  }
}

export async function setLicense(
  kv: KVNamespace,
  licenseKey: string,
  data: LicenseData,
): Promise<void> {
  try {
    await kv.put(`lic:${licenseKey}`, JSON.stringify(data))
  } catch (err) {
    console.error(`[KV] setLicense error (${licenseKey}):`, err)
    throw new Error('라이선스 저장 실패')
  }
}

export async function deleteLicense(
  kv: KVNamespace,
  licenseKey: string,
): Promise<void> {
  try {
    const data = await getLicense(kv, licenseKey)
    if (data) {
      // 역참조도 함께 삭제 (토스페이먼츠: subscriptionId 없음, deviceId만 정리)
      await Promise.all([
        kv.delete(`lic:${licenseKey}`),
        data.deviceId ? kv.delete(`dev:${data.deviceId}`) : Promise.resolve(),
      ])
    }
  } catch (err) {
    console.error(`[KV] deleteLicense error (${licenseKey}):`, err)
    throw new Error('라이선스 삭제 실패')
  }
}

// ─── 세션 관리 ───────────────────────────────────────────────────────────────

export async function getSession(
  kv: KVNamespace,
  sessionId: string,
): Promise<SessionData | null> {
  try {
    const raw = await kv.get(`sess:${sessionId}`)
    if (!raw) return null
    return JSON.parse(raw) as SessionData
  } catch (err) {
    console.error(`[KV] getSession error (${sessionId}):`, err)
    return null
  }
}

export async function setSession(
  kv: KVNamespace,
  sessionId: string,
  data: SessionData,
  ttl = 3600,
): Promise<void> {
  try {
    await kv.put(`sess:${sessionId}`, JSON.stringify(data), {
      expirationTtl: ttl,
    })
  } catch (err) {
    console.error(`[KV] setSession error (${sessionId}):`, err)
    throw new Error('세션 저장 실패')
  }
}

export async function deleteSession(
  kv: KVNamespace,
  sessionId: string,
): Promise<void> {
  try {
    await kv.delete(`sess:${sessionId}`)
  } catch (err) {
    console.error(`[KV] deleteSession error (${sessionId}):`, err)
  }
}

// ─── 역참조 ──────────────────────────────────────────────────────────────────

/**
 * deviceId → licenseKey 역참조 조회
 */
export async function getDeviceLicense(
  kv: KVNamespace,
  deviceId: string,
): Promise<string | null> {
  try {
    return await kv.get(`dev:${deviceId}`)
  } catch (err) {
    console.error(`[KV] getDeviceLicense error (${deviceId}):`, err)
    return null
  }
}

/**
 * deviceId → licenseKey 역참조 저장
 */
export async function setDeviceLicense(
  kv: KVNamespace,
  deviceId: string,
  licenseKey: string,
): Promise<void> {
  try {
    await kv.put(`dev:${deviceId}`, licenseKey)
  } catch (err) {
    console.error(`[KV] setDeviceLicense error (${deviceId}):`, err)
    throw new Error('디바이스 역참조 저장 실패')
  }
}

// ─── 헬퍼 ────────────────────────────────────────────────────────────────────

/**
 * licenseKey로 전체 라이선스 갱신 (lastVerified 자동 업데이트)
 */
export async function touchLicense(
  kv: KVNamespace,
  licenseKey: string,
  updates: Partial<LicenseData>,
): Promise<LicenseData | null> {
  const existing = await getLicense(kv, licenseKey)
  if (!existing) return null

  const updated: LicenseData = {
    ...existing,
    ...updates,
    lastVerified: new Date().toISOString(),
  }

  await setLicense(kv, licenseKey, updated)
  return updated
}

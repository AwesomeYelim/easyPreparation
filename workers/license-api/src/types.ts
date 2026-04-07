export interface LicenseData {
  plan: 'free' | 'pro' | 'enterprise'
  deviceId: string
  paymentKey?: string       // 토스페이먼츠 paymentKey (결제 취소/조회용)
  billingKey?: string       // 정기결제용 빌링키
  customerKey?: string      // 토스페이먼츠 고객 키
  subscriptionStatus: 'active' | 'past_due' | 'canceled' | 'none'
  expiresAt: string         // ISO 8601
  issuedAt: string
  lastVerified: string
  amount: number            // 결제 금액 (KRW)
}

export interface SessionData {
  deviceId: string
  plan: 'pro_monthly' | 'pro_annual'
  status: 'pending' | 'completed'
  amount: number
  orderId: string           // 주문 ID (EP-{timestamp}-{4자리랜덤})
  licenseKey?: string
  expiresAt?: string
  signature?: string
}

export interface Env {
  LICENSE_KV: KVNamespace
  ASSETS_BUCKET: R2Bucket
  TOSS_SECRET_KEY: string   // 토스페이먼츠 시크릿 키 (test_gsk_... 또는 live_gsk_...)
  TOSS_CLIENT_KEY: string   // 토스페이먼츠 클라이언트 키 (test_gck_... 또는 live_gck_...)
  HMAC_SECRET: string
}

// API Request/Response 타입

export interface CheckoutRequest {
  deviceId: string
  plan: 'pro_monthly' | 'pro_annual'
}

export interface CheckoutResponse {
  checkoutUrl: string
  sessionId: string
}

export interface ActivateRequest {
  sessionId: string
  deviceId: string
}

export interface ActivateResponse {
  status: 'pending' | 'completed'
  plan?: string
  licenseKey?: string
  expiresAt?: string
  signature?: string
}

export interface VerifyRequest {
  licenseKey: string
  deviceId: string
  signature: string
}

export interface VerifyResponse {
  valid: boolean
  plan?: string
  expiresAt?: string
  subscriptionStatus?: string
  message?: string
}

export interface PortalRequest {
  licenseKey: string
  deviceId: string
}

export interface PortalResponse {
  paymentInfo?: Record<string, unknown>
  cancelUrl?: string
}

export interface ErrorResponse {
  error: string
  code?: string
}

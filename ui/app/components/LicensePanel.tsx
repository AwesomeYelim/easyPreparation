"use client";

import { useState, useEffect, useCallback } from "react";
import { useLicense } from "@/lib/LicenseContext";
import { apiClient } from "@/lib/apiClient";
import { LicenseFeature, LicensePlan, LicenseStatus } from "@/types";

interface LicensePanelProps {
  open: boolean;
  onClose: () => void;
}

const PLAN_LABELS: Record<LicensePlan, string> = {
  free: "무료",
  pro: "Pro",
  enterprise: "Enterprise",
};

const FEATURE_LABELS: Record<LicenseFeature, string> = {
  obs_control: "OBS 연동",
  auto_scheduler: "자동 스케줄러",
  youtube_integration: "YouTube 연동",
  thumbnail: "썸네일 자동 생성",
  multi_worship: "다중 예배 순서",
  cloud_backup: "클라우드 백업",
};

const FEATURE_PLAN_MAP: Record<LicenseFeature, LicensePlan> = {
  obs_control: "pro",
  auto_scheduler: "pro",
  youtube_integration: "pro",
  thumbnail: "pro",
  multi_worship: "pro",
  cloud_backup: "enterprise",
};

const ALL_FEATURES: LicenseFeature[] = [
  "obs_control",
  "auto_scheduler",
  "youtube_integration",
  "thumbnail",
  "multi_worship",
  "cloud_backup",
];

function formatKey(raw: string): string {
  const stripped = raw.replace(/[^a-zA-Z0-9]/g, "").toUpperCase();
  const parts: string[] = [];
  let body = stripped;
  if (body.startsWith("EP")) body = body.slice(2);
  const chunks = body.slice(0, 16).match(/.{1,4}/g) || [];
  const prefix = stripped.startsWith("EP") ? "EP" : "";
  if (prefix) parts.push("EP");
  for (const c of chunks) parts.push(c);
  return parts.join("-");
}

export default function LicensePanel({ open, onClose }: LicensePanelProps) {
  const { license, refresh } = useLicense();

  const [keyInput, setKeyInput] = useState("");
  const [activating, setActivating] = useState(false);
  const [deactivating, setDeactivating] = useState(false);
  const [confirmDeactivate, setConfirmDeactivate] = useState(false);
  const [error, setError] = useState("");
  const [successMsg, setSuccessMsg] = useState("");

  const [selectedPlan, setSelectedPlan] = useState<"pro_monthly" | "pro_annual">("pro_annual");
  const [checkoutLoading, setCheckoutLoading] = useState(false);
  const [pollingSessionId, setPollingSessionId] = useState<string | null>(null);
  const [portalLoading, setPortalLoading] = useState(false);
  const [showKeyInput, setShowKeyInput] = useState(false);

  useEffect(() => {
    if (open) {
      setKeyInput("");
      setError("");
      setSuccessMsg("");
      setConfirmDeactivate(false);
      setCheckoutLoading(false);
      setPollingSessionId(null);
      setShowKeyInput(false);
      refresh();
    }
  }, [open, refresh]);

  useEffect(() => {
    if (!pollingSessionId) return;
    let cancelled = false;
    const startTime = Date.now();
    const poll = async () => {
      while (!cancelled && Date.now() - startTime < 60000) {
        try {
          const result = await apiClient.pollActivation(pollingSessionId);
          if (result.status === "completed") {
            setPollingSessionId(null);
            setCheckoutLoading(false);
            setSuccessMsg("Pro 라이선스가 성공적으로 활성화되었습니다!");
            await refresh();
            return;
          }
        } catch {}
        await new Promise((r) => setTimeout(r, 3000));
      }
      if (!cancelled) {
        setCheckoutLoading(false);
        setPollingSessionId(null);
        setError("결제 확인에 시간이 걸리고 있습니다. 페이지를 새로고침 해주세요.");
      }
    };
    poll();
    return () => { cancelled = true; };
  }, [pollingSessionId, refresh]);

  const handleKeyChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setError(""); setSuccessMsg("");
    setKeyInput(formatKey(e.target.value));
  };

  const handleActivate = async () => {
    const raw = keyInput.replace(/-/g, "");
    if (raw.length < 10) { setError("라이선스 키를 입력해주세요."); return; }
    setActivating(true); setError(""); setSuccessMsg("");
    try {
      const res = await apiClient.activateLicense(keyInput);
      if (res.error) { setError(res.error); }
      else {
        setSuccessMsg("라이선스가 성공적으로 활성화되었습니다.");
        setKeyInput(""); setShowKeyInput(false);
        await refresh();
      }
    } catch (e: any) {
      setError(`활성화에 실패했습니다: ${e?.message || e}`);
    } finally {
      setActivating(false);
    }
  };

  const handleDeactivate = async () => {
    if (!confirmDeactivate) { setConfirmDeactivate(true); return; }
    setDeactivating(true); setError(""); setSuccessMsg("");
    try {
      const res = await apiClient.deactivateLicense();
      if (res.error) { setError(res.error); }
      else {
        setSuccessMsg("라이선스가 해제되었습니다.");
        setConfirmDeactivate(false);
        await refresh();
      }
    } catch (e: any) {
      setError(`해제에 실패했습니다: ${e?.message || e}`);
    } finally {
      setDeactivating(false);
    }
  };

  const handleCheckout = async () => {
    setCheckoutLoading(true); setError(""); setSuccessMsg("");
    try {
      const { checkoutUrl, sessionId } = await apiClient.createCheckoutSession(selectedPlan);
      window.open(checkoutUrl, "_blank");
      setPollingSessionId(sessionId);
    } catch (e: any) {
      setCheckoutLoading(false);
      setError(`결제 세션 생성에 실패했습니다: ${e?.message || e}`);
    }
  };

  const handlePortal = async () => {
    setPortalLoading(true); setError("");
    try {
      const { portalUrl } = await apiClient.getPortalUrl();
      window.open(portalUrl, "_blank");
    } catch (e: any) {
      setError(`구독 관리 페이지를 열 수 없습니다: ${e?.message || e}`);
    } finally {
      setPortalLoading(false);
    }
  };

  const truncateDeviceId = (id: string) => {
    if (!id) return "-";
    if (id.length <= 16) return id;
    return `EP-${id.slice(0, 4)}...`;
  };

  const formatExpiry = (expiresAt: string | null): string => {
    if (!expiresAt) return "-";
    const d = new Date(expiresAt);
    return `${d.getFullYear()}.${String(d.getMonth() + 1).padStart(2, "0")}.${String(d.getDate()).padStart(2, "0")}`;
  };

  const isPro = license.plan === "pro" || license.plan === "enterprise";

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-[11000] flex items-center justify-center bg-black/50"
      onClick={onClose}
    >
      <div
        className="bg-[var(--surface-elevated)] rounded-2xl w-[480px] max-w-[92vw] max-h-[90vh] overflow-y-auto shadow-2xl flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        {/* 헤더 (sticky) */}
        <div className="flex justify-between items-center px-6 pt-5 pb-4 border-b border-[var(--border)] sticky top-0 bg-[var(--surface-elevated)] z-10 rounded-t-2xl">
          <h3 className="m-0 text-lg font-bold text-[var(--text-primary)]">라이선스 정보</h3>
          <button
            className="bg-transparent border-none text-2xl cursor-pointer text-[var(--text-secondary)] leading-none p-0 flex items-center justify-center hover:text-[var(--text-primary)] transition-colors"
            onClick={onClose}
          >
            &times;
          </button>
        </div>

        {/* body */}
        <div className="px-6 py-5 flex flex-col gap-0">

          {/* 플랜 + 기기 ID 요약 행 */}
          <div className="flex gap-4 pb-4">
            <div className="flex-1">
              <span className="text-[11px] font-bold uppercase tracking-[0.8px] text-[var(--text-muted)]">
                현재 플랜
              </span>
              <div className="flex items-center gap-2 mt-1">
                <span className={`inline-block px-3.5 py-1 rounded-[20px] text-xs font-bold tracking-wide ${
                  license.plan === "free"
                    ? "bg-[#f3f4f6] text-[#6b7280] border border-[#d1d5db]"
                    : license.plan === "pro"
                    ? "bg-[#eff6ff] text-[#1d4ed8] border border-[#93c5fd]"
                    : "bg-[#f5f3ff] text-[#6d28d9] border border-[#c4b5fd]"
                }`}>
                  {PLAN_LABELS[license.plan]}
                </span>
                {license.is_active && (
                  <span className="text-[11px] font-semibold px-2 py-0.5 rounded-[10px] bg-[#d1fae5] text-[#065f46]">
                    활성
                  </span>
                )}
                {license.grace_period && (
                  <span className="text-[11px] font-semibold px-2 py-0.5 rounded-[10px] bg-[#fef3c7] text-[#92400e]">
                    유예 기간
                  </span>
                )}
              </div>
            </div>
            <div className="flex-none text-right">
              <span className="text-[11px] font-bold uppercase tracking-[0.8px] text-[var(--text-muted)]">
                기기 ID
              </span>
              <div
                className="mt-1 text-xs font-mono text-[var(--text-secondary)] bg-[var(--surface-input)] border border-[var(--border)] rounded-md px-2.5 py-1 tracking-[0.5px] cursor-text select-all inline-block"
                title={license.device_id}
              >
                {truncateDeviceId(license.device_id)}
              </div>
            </div>
          </div>

          <div className="h-px bg-[var(--border)] mb-4" />

          {/* ── Free 플랜: 업그레이드 카드 ── */}
          {!isPro && (
            <div className="pb-4">
              <div className="bg-gradient-to-br from-[#eff6ff] to-[#f5f3ff] border border-[#c7d2fe] rounded-xl p-4 pb-3.5">
                <div className="flex flex-col gap-0.5 mb-4">
                  <span className="text-[15px] font-bold text-[#1e40af]">Pro 업그레이드</span>
                  <span className="text-xs text-[#4b5563] leading-snug">
                    OBS 연동, 스케줄러, YouTube, 썸네일 자동 생성
                  </span>
                </div>

                {/* 가격 카드 */}
                <div className="flex gap-2.5 mb-3.5">
                  {/* 월간 */}
                  <button
                    type="button"
                    onClick={() => setSelectedPlan("pro_monthly")}
                    className={`flex-1 relative px-3 py-3.5 rounded-xl border-2 text-center flex flex-col items-center gap-0.5 cursor-pointer transition-all ${
                      selectedPlan === "pro_monthly"
                        ? "border-[#4f46e5] shadow-[0_0_0_3px_rgba(79,70,229,0.15)] bg-[#f0f0ff]"
                        : "border-[#e0e7ff] bg-[var(--surface-elevated)] hover:border-[#818cf8]"
                    }`}
                  >
                    <div className="text-xl font-extrabold text-[#1e1b4b] leading-tight">₩9,900</div>
                    <div className="text-xs text-[#6b7280] font-medium">/ 월</div>
                  </button>

                  {/* 연간 */}
                  <button
                    type="button"
                    onClick={() => setSelectedPlan("pro_annual")}
                    className={`flex-1 relative px-3 py-3.5 rounded-xl border-2 text-center flex flex-col items-center gap-0.5 cursor-pointer transition-all ${
                      selectedPlan === "pro_annual"
                        ? "border-[#4f46e5] shadow-[0_0_0_3px_rgba(79,70,229,0.15)] bg-[#f0f0ff]"
                        : "border-[#e0e7ff] bg-[var(--surface-elevated)] hover:border-[#818cf8]"
                    }`}
                  >
                    <div className="absolute -top-2.5 left-1/2 -translate-x-1/2 bg-[#4f46e5] text-white text-[10px] font-bold px-2.5 py-0.5 rounded-[10px] whitespace-nowrap">
                      추천
                    </div>
                    <div className="text-xl font-extrabold text-[#1e1b4b] leading-tight">₩99,000</div>
                    <div className="text-xs text-[#6b7280] font-medium">/ 년</div>
                    <div className="mt-1 text-[11px] font-semibold text-[#059669] bg-[#d1fae5] px-1.5 py-0.5 rounded-md border border-[#a7f3d0]">
                      ~17% 할인
                    </div>
                  </button>
                </div>

                <button
                  type="button"
                  onClick={handleCheckout}
                  disabled={checkoutLoading}
                  className="w-full py-3 text-sm font-bold bg-[#4f46e5] text-white border-none rounded-xl cursor-pointer flex items-center justify-center transition-colors hover:bg-[#4338ca] disabled:bg-[#a5b4fc] disabled:cursor-default"
                >
                  {checkoutLoading ? (
                    <span className="flex items-center gap-2">
                      <span className="w-3.5 h-3.5 rounded-full border-2 border-white/30 border-t-white animate-spin flex-shrink-0" />
                      결제 창 열기 대기 중...
                    </span>
                  ) : (
                    "Pro 업그레이드"
                  )}
                </button>

                {checkoutLoading && (
                  <div className="mt-2.5 text-xs text-[#4b5563] text-center px-3 py-2 bg-[rgba(79,70,229,0.06)] rounded-lg">
                    결제 창에서 완료하면 자동으로 활성화됩니다.
                  </div>
                )}
              </div>
            </div>
          )}

          {/* ── Pro 플랜: 구독 정보 + 관리 ── */}
          {isPro && license.is_active && (
            <div className="pb-4">
              <div className="text-[11px] font-bold uppercase tracking-[0.8px] text-[var(--text-muted)] mb-2">
                구독 정보
              </div>
              <div className="flex flex-col gap-1.5">
                {[
                  { key: "플랜", val: `${PLAN_LABELS[license.plan]}${license.plan === "pro" ? " (월간)" : ""}`, warn: false },
                  { key: "만료일", val: formatExpiry(license.expires_at), warn: false },
                  { key: "남은 일수", val: license.days_remaining > 0 ? `${license.days_remaining}일` : "만료됨", warn: license.days_remaining <= 7 },
                ].map(({ key, val, warn }) => (
                  <div key={key} className="flex justify-between items-center px-3 py-1.5 bg-[var(--surface-input)] border border-[var(--border)] rounded-lg">
                    <span className="text-xs text-[var(--text-secondary)] font-medium">{key}</span>
                    <span className={`text-xs font-semibold ${warn ? "text-[#d97706]" : "text-[var(--text-primary)]"}`}>
                      {val}
                    </span>
                  </div>
                ))}
              </div>

              {license.grace_period && (
                <div className="mt-2 px-3 py-2 bg-[#fffbeb] text-[#92400e] border border-[#fcd34d] rounded-lg text-xs font-medium">
                  유예 기간 중입니다. 라이선스를 갱신해주세요.
                </div>
              )}

              <div className="mt-3 flex flex-col gap-2">
                <button
                  type="button"
                  onClick={handlePortal}
                  disabled={portalLoading}
                  className="self-start px-4 py-2 text-xs font-semibold bg-[var(--accent)] text-[var(--surface-elevated)] border-none rounded-lg cursor-pointer hover:bg-[var(--accent-hover)] disabled:bg-[var(--text-muted)] disabled:cursor-default transition-colors"
                >
                  {portalLoading ? "로딩 중..." : "구독 관리"}
                </button>

                {confirmDeactivate ? (
                  <div className="bg-[#fef2f2] border border-[#fecaca] rounded-lg px-3.5 py-3">
                    <p className="m-0 mb-3 text-xs text-[#7f1d1d] leading-relaxed">
                      정말 라이선스를 해제하시겠습니까? 해제 후 Pro 기능을 사용할 수 없습니다.
                    </p>
                    <div className="flex gap-2 justify-end">
                      <button
                        type="button"
                        onClick={() => setConfirmDeactivate(false)}
                        disabled={deactivating}
                        className="px-3.5 py-1.5 text-xs bg-[var(--surface-hover)] border border-[var(--border-input)] rounded-lg cursor-pointer text-[var(--text-primary)]"
                      >
                        취소
                      </button>
                      <button
                        type="button"
                        onClick={handleDeactivate}
                        disabled={deactivating}
                        className="px-3.5 py-1.5 text-xs font-semibold bg-[var(--error)] text-white border-none rounded-lg cursor-pointer disabled:opacity-50 disabled:cursor-default"
                      >
                        {deactivating ? "처리 중..." : "해제 확인"}
                      </button>
                    </div>
                  </div>
                ) : (
                  <button
                    type="button"
                    onClick={handleDeactivate}
                    disabled={deactivating}
                    className="self-start px-4 py-1.5 text-xs font-medium bg-[var(--surface-hover)] border border-[var(--border-input)] rounded-lg cursor-pointer text-[var(--error)] hover:bg-[#fef2f2] hover:border-[#fecaca] disabled:opacity-50 disabled:cursor-default transition-colors"
                  >
                    라이선스 해제
                  </button>
                )}
              </div>
            </div>
          )}

          <div className="h-px bg-[var(--border)] mb-4" />

          {/* 기능 목록 */}
          <div className="pb-4">
            <div className="text-[11px] font-bold uppercase tracking-[0.8px] text-[var(--text-muted)] mb-2">
              기능 목록
            </div>
            <div className="flex flex-col gap-1">
              {ALL_FEATURES.map((feat) => {
                const available = license.features.includes(feat);
                const requiredPlan = FEATURE_PLAN_MAP[feat];
                return (
                  <div
                    key={feat}
                    className={`flex items-center gap-2.5 px-2.5 py-2 rounded-lg border ${
                      available
                        ? "border-[#a7f3d0] bg-[#f0fdf4]"
                        : "border-[var(--border)] bg-[var(--surface-input)] opacity-70"
                    }`}
                  >
                    <span className="flex items-center flex-shrink-0">
                      {available ? (
                        <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                          <circle cx="8" cy="8" r="8" fill="#059669" fillOpacity="0.15" />
                          <path d="M4.5 8L7 10.5L11.5 6" stroke="#059669" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                        </svg>
                      ) : (
                        <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                          <rect x="4" y="7" width="8" height="6" rx="1.5" stroke="#9ca3af" strokeWidth="1.5" />
                          <path d="M5.5 7V5.5a2.5 2.5 0 015 0V7" stroke="#9ca3af" strokeWidth="1.5" strokeLinecap="round" />
                        </svg>
                      )}
                    </span>
                    <span className={`text-xs font-medium flex-1 ${available ? "text-[var(--text-primary)]" : "text-[var(--text-secondary)]"}`}>
                      {FEATURE_LABELS[feat]}
                    </span>
                    {!available && (
                      <span className={`text-[10px] font-bold px-1.5 py-0.5 rounded-lg tracking-wide ${
                        requiredPlan === "pro"
                          ? "bg-[#eff6ff] text-[#1d4ed8] border border-[#bfdbfe]"
                          : "bg-[#f5f3ff] text-[#6d28d9] border border-[#c4b5fd]"
                      }`}>
                        {PLAN_LABELS[requiredPlan]}
                      </span>
                    )}
                  </div>
                );
              })}
            </div>
          </div>

          <div className="h-px bg-[var(--border)] mb-0" />

          {/* 라이선스 키 직접 입력 */}
          <div className="pt-0">
            <button
              type="button"
              onClick={() => setShowKeyInput((v) => !v)}
              className="flex items-center justify-between w-full py-2 bg-transparent border-none border-t border-dashed border-[var(--border)] cursor-pointer text-[var(--text-secondary)] text-xs font-semibold tracking-wide hover:text-[var(--text-primary)] transition-colors"
            >
              <span>또는 라이선스 키로 직접 활성화</span>
              <span className={`flex items-center transition-transform duration-200 ${showKeyInput ? "rotate-180" : ""}`}>
                <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
                  <path d="M3 5l4 4 4-4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
              </span>
            </button>

            {showKeyInput && (
              <div className="mt-2.5">
                <div className="flex gap-2">
                  <input
                    type="text"
                    placeholder="EP-XXXX-XXXX-XXXX-XXXX"
                    value={keyInput}
                    onChange={handleKeyChange}
                    maxLength={24}
                    spellCheck={false}
                    autoComplete="off"
                    className="flex-1 px-3 py-2 text-xs font-mono tracking-[1px] border border-[var(--border-input)] rounded-lg bg-[var(--surface-input)] text-[var(--text-primary)] outline-none uppercase focus:border-[var(--accent)]"
                  />
                  <button
                    type="button"
                    onClick={handleActivate}
                    disabled={activating || keyInput.replace(/-/g, "").length < 10}
                    className="px-4 py-2 text-xs font-semibold bg-[var(--accent)] text-[var(--surface-elevated)] border-none rounded-lg cursor-pointer flex-shrink-0 hover:bg-[var(--accent-hover)] disabled:bg-[var(--text-muted)] disabled:cursor-default transition-colors"
                  >
                    {activating ? "확인 중..." : "활성화"}
                  </button>
                </div>
              </div>
            )}

            {error && (
              <div className="mt-2 px-3 py-2 bg-[var(--error-bg)] text-[var(--error)] border border-[var(--error-border)] rounded-lg text-xs">
                {error}
              </div>
            )}
            {successMsg && (
              <div className="mt-2 px-3 py-2 bg-[#f0fdf4] text-[#065f46] border border-[#a7f3d0] rounded-lg text-xs font-medium">
                {successMsg}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

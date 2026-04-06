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

const ALL_FEATURES: LicenseFeature[] = [
  "obs_control",
  "auto_scheduler",
  "youtube_integration",
  "thumbnail",
  "multi_worship",
  "cloud_backup",
];

function formatKey(raw: string): string {
  // 영숫자만 추출 후 대문자 변환
  const stripped = raw.replace(/[^a-zA-Z0-9]/g, "").toUpperCase();
  // EP-XXXX-XXXX-XXXX-XXXX 형태로 삽입
  const parts: string[] = [];
  // 앞 2자리가 EP면 prefix 처리
  let body = stripped;
  if (body.startsWith("EP")) {
    body = body.slice(2);
  }
  // 최대 16자리를 4글자씩 분할
  const chunks = body.slice(0, 16).match(/.{1,4}/g) || [];
  const prefix = stripped.startsWith("EP") ? "EP" : "";
  if (prefix) {
    parts.push("EP");
  }
  for (const c of chunks) {
    parts.push(c);
  }
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

  // 열릴 때마다 상태 초기화
  useEffect(() => {
    if (open) {
      setKeyInput("");
      setError("");
      setSuccessMsg("");
      setConfirmDeactivate(false);
      refresh();
    }
  }, [open, refresh]);

  const handleKeyChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setError("");
    setSuccessMsg("");
    const formatted = formatKey(e.target.value);
    setKeyInput(formatted);
  };

  const handleActivate = async () => {
    const raw = keyInput.replace(/-/g, "");
    if (raw.length < 10) {
      setError("라이선스 키를 입력해주세요.");
      return;
    }
    setActivating(true);
    setError("");
    setSuccessMsg("");
    try {
      const res = await apiClient.activateLicense(keyInput);
      if (res.error) {
        setError(res.error);
      } else {
        setSuccessMsg("라이선스가 성공적으로 활성화되었습니다.");
        setKeyInput("");
        await refresh();
      }
    } catch (e: any) {
      setError(`활성화에 실패했습니다: ${e?.message || e}`);
    } finally {
      setActivating(false);
    }
  };

  const handleDeactivate = async () => {
    if (!confirmDeactivate) {
      setConfirmDeactivate(true);
      return;
    }
    setDeactivating(true);
    setError("");
    setSuccessMsg("");
    try {
      const res = await apiClient.deactivateLicense();
      if (res.error) {
        setError(res.error);
      } else {
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

  const truncateDeviceId = (id: string) => {
    if (!id) return "-";
    if (id.length <= 16) return id;
    return `${id.slice(0, 8)}...${id.slice(-8)}`;
  };

  const formatExpiry = (expiresAt: string | null): string => {
    if (!expiresAt) return "-";
    const d = new Date(expiresAt);
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");
    return `${y}.${m}.${day}`;
  };

  const isPro = license.plan === "pro" || license.plan === "enterprise";

  if (!open) return null;

  return (
    <div className="lp_overlay" onClick={onClose}>
      <div className="lp_panel" onClick={(e) => e.stopPropagation()}>
        {/* 헤더 */}
        <div className="lp_header">
          <h3>라이선스 정보</h3>
          <button className="lp_close" onClick={onClose}>&times;</button>
        </div>

        <div className="lp_body">
          {/* 플랜 배지 */}
          <div className="lp_section">
            <div className="lp_label">현재 플랜</div>
            <div style={{ display: "flex", alignItems: "center", gap: 10, marginTop: 6 }}>
              <span className={`lp_plan_badge lp_plan_${license.plan}`}>
                {PLAN_LABELS[license.plan]}
              </span>
              {license.is_active && (
                <span className="lp_active_badge">활성</span>
              )}
              {license.grace_period && (
                <span className="lp_grace_badge">유예 기간</span>
              )}
            </div>
          </div>

          {/* 기기 ID */}
          <div className="lp_section">
            <div className="lp_label">기기 ID</div>
            <div className="lp_device_id" title={license.device_id}>
              {truncateDeviceId(license.device_id)}
            </div>
          </div>

          <div className="lp_divider" />

          {/* 기능 목록 */}
          <div className="lp_section">
            <div className="lp_label">기능 목록</div>
            <div className="lp_feature_list">
              {ALL_FEATURES.map((feat) => {
                const available = license.features.includes(feat);
                return (
                  <div key={feat} className={`lp_feature_item ${available ? "available" : "locked"}`}>
                    <span className="lp_feature_icon">
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
                    <span className="lp_feature_name">{FEATURE_LABELS[feat]}</span>
                    {!available && <span className="lp_pro_tag">Pro</span>}
                  </div>
                );
              })}
            </div>
          </div>

          <div className="lp_divider" />

          {/* Pro 활성 시 만료일 정보 */}
          {isPro && license.is_active && (
            <>
              <div className="lp_section">
                <div className="lp_label">라이선스 상세</div>
                <div className="lp_info_grid">
                  <div className="lp_info_row">
                    <span className="lp_info_key">만료일</span>
                    <span className="lp_info_val">{formatExpiry(license.expires_at)}</span>
                  </div>
                  <div className="lp_info_row">
                    <span className="lp_info_key">남은 일수</span>
                    <span className={`lp_info_val ${license.days_remaining <= 7 ? "warn" : ""}`}>
                      {license.days_remaining > 0 ? `${license.days_remaining}일` : "만료됨"}
                    </span>
                  </div>
                </div>
                {license.grace_period && (
                  <div className="lp_grace_notice">
                    유예 기간 중입니다. 라이선스를 갱신해주세요.
                  </div>
                )}
              </div>

              <div className="lp_divider" />

              {/* 라이선스 해제 */}
              <div className="lp_section">
                <div className="lp_label">라이선스 관리</div>
                <div style={{ marginTop: 8 }}>
                  {confirmDeactivate ? (
                    <div className="lp_confirm_box">
                      <p className="lp_confirm_text">정말 라이선스를 해제하시겠습니까? 해제 후 Pro 기능을 사용할 수 없습니다.</p>
                      <div style={{ display: "flex", gap: 8, justifyContent: "flex-end" }}>
                        <button
                          className="lp_cancel_btn"
                          onClick={() => setConfirmDeactivate(false)}
                          disabled={deactivating}
                        >
                          취소
                        </button>
                        <button
                          className="lp_deactivate_confirm_btn"
                          onClick={handleDeactivate}
                          disabled={deactivating}
                        >
                          {deactivating ? "처리 중..." : "해제 확인"}
                        </button>
                      </div>
                    </div>
                  ) : (
                    <button
                      className="lp_deactivate_btn"
                      onClick={handleDeactivate}
                      disabled={deactivating}
                    >
                      라이선스 해제
                    </button>
                  )}
                </div>
              </div>

              <div className="lp_divider" />
            </>
          )}

          {/* 라이선스 키 입력 */}
          <div className="lp_section">
            <div className="lp_label">{isPro ? "라이선스 키 변경" : "라이선스 활성화"}</div>
            <div style={{ marginTop: 8, display: "flex", gap: 8 }}>
              <input
                className="lp_key_input"
                type="text"
                placeholder="EP-XXXX-XXXX-XXXX-XXXX"
                value={keyInput}
                onChange={handleKeyChange}
                maxLength={24}
                spellCheck={false}
                autoComplete="off"
              />
              <button
                className="lp_activate_btn"
                onClick={handleActivate}
                disabled={activating || keyInput.replace(/-/g, "").length < 10}
              >
                {activating ? "확인 중..." : "활성화"}
              </button>
            </div>
            {error && <div className="lp_error">{error}</div>}
            {successMsg && <div className="lp_success">{successMsg}</div>}
          </div>
        </div>
      </div>

      <style jsx>{`
        .lp_overlay {
          position: fixed; top: 0; left: 0; width: 100%; height: 100%;
          background: rgba(0,0,0,0.5);
          display: flex; align-items: center; justify-content: center;
          z-index: 11000;
        }
        .lp_panel {
          background: var(--surface-elevated); border-radius: 16px;
          width: 460px; max-width: 90vw; max-height: 88vh;
          overflow-y: auto; box-shadow: 0 8px 32px rgba(0,0,0,0.2);
          display: flex; flex-direction: column;
        }
        .lp_header {
          display: flex; justify-content: space-between; align-items: center;
          padding: 20px 24px 16px; border-bottom: 1px solid var(--border);
          position: sticky; top: 0; background: var(--surface-elevated); z-index: 1;
          border-radius: 16px 16px 0 0;
        }
        .lp_header h3 {
          margin: 0; font-size: 18px; font-weight: 700; color: var(--text-primary);
        }
        .lp_close {
          background: none; border: none; font-size: 24px;
          cursor: pointer; color: var(--text-secondary); line-height: 1;
          padding: 0; display: flex; align-items: center; justify-content: center;
        }
        .lp_close:hover { color: var(--text-primary); }
        .lp_body {
          padding: 20px 24px; display: flex; flex-direction: column; gap: 0;
        }
        .lp_section { padding: 4px 0 16px; }
        .lp_label {
          font-size: 11px; font-weight: 700; letter-spacing: 0.8px;
          text-transform: uppercase; color: var(--text-muted); margin-bottom: 2px;
        }
        .lp_divider {
          height: 1px; background: var(--border); margin: 4px 0 16px;
        }

        /* 플랜 배지 */
        .lp_plan_badge {
          display: inline-block; padding: 4px 14px; border-radius: 20px;
          font-size: 13px; font-weight: 700; letter-spacing: 0.3px;
        }
        .lp_plan_free {
          background: #f3f4f6; color: #6b7280;
          border: 1.5px solid #d1d5db;
        }
        .lp_plan_pro {
          background: #eff6ff; color: #1d4ed8;
          border: 1.5px solid #93c5fd;
        }
        .lp_plan_enterprise {
          background: #f5f3ff; color: #6d28d9;
          border: 1.5px solid #c4b5fd;
        }
        .lp_active_badge {
          font-size: 11px; font-weight: 600; padding: 3px 8px; border-radius: 10px;
          background: #d1fae5; color: #065f46;
        }
        .lp_grace_badge {
          font-size: 11px; font-weight: 600; padding: 3px 8px; border-radius: 10px;
          background: #fef3c7; color: #92400e;
        }

        /* 기기 ID */
        .lp_device_id {
          margin-top: 6px; font-size: 12px; font-family: monospace;
          color: var(--text-secondary); background: var(--surface-input);
          border: 1px solid var(--border); border-radius: 6px;
          padding: 7px 12px; letter-spacing: 0.5px;
          user-select: all; cursor: text;
        }

        /* 기능 목록 */
        .lp_feature_list {
          margin-top: 8px; display: flex; flex-direction: column; gap: 4px;
        }
        .lp_feature_item {
          display: flex; align-items: center; gap: 10px;
          padding: 8px 10px; border-radius: 8px;
          border: 1px solid var(--border);
          background: var(--surface-input);
        }
        .lp_feature_item.available {
          border-color: #a7f3d0; background: #f0fdf4;
        }
        .lp_feature_item.locked {
          opacity: 0.7;
        }
        .lp_feature_icon { display: flex; align-items: center; flex-shrink: 0; }
        .lp_feature_name {
          font-size: 13px; font-weight: 500;
          color: var(--text-primary); flex: 1;
        }
        .lp_feature_item.locked .lp_feature_name { color: var(--text-secondary); }
        .lp_pro_tag {
          font-size: 10px; font-weight: 700; padding: 2px 7px;
          border-radius: 8px; background: #eff6ff; color: #1d4ed8;
          border: 1px solid #bfdbfe; letter-spacing: 0.3px;
        }

        /* 만료일 그리드 */
        .lp_info_grid {
          margin-top: 8px; display: flex; flex-direction: column; gap: 6px;
        }
        .lp_info_row {
          display: flex; justify-content: space-between; align-items: center;
          padding: 7px 12px; background: var(--surface-input);
          border: 1px solid var(--border); border-radius: 8px;
        }
        .lp_info_key { font-size: 13px; color: var(--text-secondary); font-weight: 500; }
        .lp_info_val { font-size: 13px; color: var(--text-primary); font-weight: 600; }
        .lp_info_val.warn { color: #d97706; }

        /* 유예 기간 알림 */
        .lp_grace_notice {
          margin-top: 8px; padding: 8px 12px;
          background: #fffbeb; color: #92400e;
          border: 1px solid #fcd34d; border-radius: 8px;
          font-size: 12px; font-weight: 500;
        }

        /* 라이선스 해제 */
        .lp_deactivate_btn {
          padding: 7px 16px; font-size: 13px; font-weight: 500;
          background: var(--surface-hover); border: 1px solid var(--border-input);
          border-radius: 8px; cursor: pointer; color: var(--error);
        }
        .lp_deactivate_btn:hover { background: #fef2f2; border-color: #fecaca; }
        .lp_deactivate_btn:disabled { opacity: 0.5; cursor: default; }
        .lp_confirm_box {
          background: #fef2f2; border: 1px solid #fecaca; border-radius: 8px;
          padding: 12px 14px;
        }
        .lp_confirm_text {
          margin: 0 0 12px; font-size: 13px; color: #7f1d1d; line-height: 1.5;
        }
        .lp_cancel_btn {
          padding: 6px 14px; font-size: 13px; background: var(--surface-hover);
          border: 1px solid var(--border-input); border-radius: 7px;
          cursor: pointer; color: var(--text-primary);
        }
        .lp_deactivate_confirm_btn {
          padding: 6px 14px; font-size: 13px; font-weight: 600;
          background: var(--error); color: #fff; border: none;
          border-radius: 7px; cursor: pointer;
        }
        .lp_deactivate_confirm_btn:disabled { opacity: 0.5; cursor: default; }

        /* 키 입력 영역 */
        .lp_key_input {
          flex: 1; padding: 8px 12px; font-size: 13px; font-family: monospace;
          letter-spacing: 1px; border: 1.5px solid var(--border-input);
          border-radius: 8px; background: var(--surface-input);
          color: var(--text-primary); outline: none;
          text-transform: uppercase;
        }
        .lp_key_input:focus { border-color: var(--accent); }
        .lp_activate_btn {
          padding: 8px 18px; font-size: 13px; font-weight: 600;
          background: var(--accent); color: var(--surface-elevated);
          border: none; border-radius: 8px; cursor: pointer; white-space: nowrap;
          flex-shrink: 0;
        }
        .lp_activate_btn:hover { background: var(--accent-hover); }
        .lp_activate_btn:disabled { background: var(--text-muted); cursor: default; }
        .lp_error {
          margin-top: 8px; padding: 8px 12px;
          background: var(--error-bg); color: var(--error);
          border: 1px solid var(--error-border); border-radius: 8px;
          font-size: 13px;
        }
        .lp_success {
          margin-top: 8px; padding: 8px 12px;
          background: #f0fdf4; color: #065f46;
          border: 1px solid #a7f3d0; border-radius: 8px;
          font-size: 13px; font-weight: 500;
        }

        /* 다크 테마 오버라이드 */
        :global([data-theme="dark"]) .lp_plan_free {
          background: #334155; color: #94a3b8; border-color: #475569;
        }
        :global([data-theme="dark"]) .lp_plan_pro {
          background: #1e3a5f; color: #93c5fd; border-color: #2563eb;
        }
        :global([data-theme="dark"]) .lp_plan_enterprise {
          background: #2e1065; color: #c4b5fd; border-color: #7c3aed;
        }
        :global([data-theme="dark"]) .lp_active_badge {
          background: #065f46; color: #a7f3d0;
        }
        :global([data-theme="dark"]) .lp_grace_badge {
          background: #78350f; color: #fcd34d;
        }
        :global([data-theme="dark"]) .lp_feature_item.available {
          border-color: #065f46; background: #022c22;
        }
        :global([data-theme="dark"]) .lp_pro_tag {
          background: #1e3a5f; color: #93c5fd; border-color: #2563eb;
        }
        :global([data-theme="dark"]) .lp_grace_notice {
          background: #451a03; color: #fcd34d; border-color: #92400e;
        }
        :global([data-theme="dark"]) .lp_confirm_box {
          background: #451a1a; border-color: #7f1d1d;
        }
        :global([data-theme="dark"]) .lp_confirm_text { color: #fca5a5; }
        :global([data-theme="dark"]) .lp_success {
          background: #022c22; color: #a7f3d0; border-color: #065f46;
        }
      `}</style>
    </div>
  );
}

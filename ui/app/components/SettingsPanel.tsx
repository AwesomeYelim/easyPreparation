"use client";

import { useState, useEffect } from "react";
import { useRecoilState, useRecoilValue } from "recoil";
import { userSettingsState, userInfoState } from "@/recoilState";
import { apiClient } from "@/lib/apiClient";
import { UserSettings } from "@/types";
import { useFeature } from "./FeatureGate";

function useCurrentVersion() {
  const [version, setVersion] = useState<string | null>(null);
  useEffect(() => {
    apiClient.getVersion().then((v) => setVersion(v.version)).catch(() => {});
  }, []);
  return version;
}

interface SettingsPanelProps {
  open: boolean;
  onClose: () => void;
}

export default function SettingsPanel({ open, onClose }: SettingsPanelProps) {
  const userInfo = useRecoilValue(userInfoState);
  const [settings, setSettings] = useRecoilState(userSettingsState);
  const [local, setLocal] = useState<UserSettings>(settings);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const currentVersion = useCurrentVersion();

  useEffect(() => {
    setLocal(settings);
    setError("");
  }, [settings, open]);

  const handleSave = async () => {
    if (!userInfo.email) {
      setError("로그인 정보가 없습니다. 다시 로그인해주세요.");
      return;
    }
    setSaving(true);
    setError("");
    try {
      const res = await apiClient.saveSettings(userInfo.email, local);
      if (res.error) {
        setError(`설정 저장 실패: ${res.error}`);
        return;
      }
      setSettings(local);
      document.documentElement.setAttribute("data-theme", local.theme || "light");
      if (local.font_size) {
        document.documentElement.style.setProperty("--user-font-size", `${local.font_size}px`);
      }
      onClose();
    } catch (e: any) {
      console.error("설정 저장 에러:", e);
      setError(`설정 저장에 실패했습니다: ${e?.message || e}`);
    } finally {
      setSaving(false);
    }
  };

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-[11000] flex items-center justify-center bg-black/50"
      onClick={onClose}
    >
      <div
        className="bg-[var(--surface-elevated)] rounded-2xl w-[480px] max-w-[90vw] max-h-[85vh] overflow-hidden flex flex-col shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        {/* header */}
        <div className="flex justify-between items-center px-6 pt-5 pb-3">
          <div className="flex flex-col gap-0.5">
            <h3 className="m-0 text-lg font-bold text-[var(--text-primary)]">설정</h3>
            {currentVersion && (
              <span className="text-[11px] text-[var(--text-muted)] font-mono tracking-[0.3px]">
                v{currentVersion}
              </span>
            )}
          </div>
          <button
            className="bg-transparent border-none text-2xl cursor-pointer text-[var(--text-secondary)] leading-none"
            onClick={onClose}
          >
            &times;
          </button>
        </div>

        {/* body */}
        <div className="px-6 py-5 flex flex-col gap-4 min-h-[200px] overflow-y-auto flex-1">
          <GeneralTab
            local={local}
            setLocal={setLocal}
          />
        </div>

        {/* error */}
        {error && (
          <div className="mx-6 px-3 py-2 bg-[var(--error-bg)] text-[var(--error)] border border-[var(--error-border)] rounded-lg text-xs">
            {error}
          </div>
        )}

        {/* footer */}
        <div className="flex justify-end gap-2 px-6 pb-5 pt-4 border-t border-[var(--border)]">
          <button
            className="px-5 py-2 text-xs bg-[var(--surface-hover)] border border-[var(--border-input)] rounded-xl cursor-pointer"
            onClick={onClose}
          >
            취소
          </button>
          <button
            className="px-5 py-2 text-xs font-semibold bg-[var(--accent)] text-[var(--surface-elevated)] border-none rounded-xl cursor-pointer hover:bg-[var(--accent-hover)] disabled:bg-[var(--text-muted)]"
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? "저장 중..." : "저장"}
          </button>
        </div>
      </div>

    </div>
  );
}

/* ── 일반 탭 ── */
function GeneralTab({
  local,
  setLocal,
}: {
  local: UserSettings;
  setLocal: (s: UserSettings) => void;
}) {
  useFeature("auto_scheduler"); // 미사용 제거 방지용

  const rowClass = "flex justify-between items-center gap-3";
  const labelClass = "text-sm font-medium text-[var(--text-primary)]";
  const inputClass =
    "px-3 py-1.5 border border-[var(--border-input)] rounded-lg text-xs bg-[var(--surface-input)] min-w-[140px]";
  const numberInputClass =
    "px-3 py-1.5 border border-[var(--border-input)] rounded-lg text-xs bg-[var(--surface-input)] w-20";

  return (
    <>
      <label className={rowClass}>
        <span className={labelClass}>선호 성경 번역</span>
        <select
          className={inputClass}
          value={local.preferred_bible_version}
          onChange={(e) => setLocal({ ...local, preferred_bible_version: Number(e.target.value) })}
        >
          <option value={1}>개역개정</option>
          <option value={2}>개역한글</option>
          <option value={3}>공동번역</option>
          <option value={4}>표준새번역</option>
          <option value={5}>NIV</option>
          <option value={6}>KJV</option>
          <option value={7}>우리말성경</option>
        </select>
      </label>
      <label className={rowClass}>
        <span className={labelClass}>테마</span>
        <select
          className={inputClass}
          value={local.theme}
          onChange={(e) => setLocal({ ...local, theme: e.target.value })}
        >
          <option value="light">라이트</option>
          <option value="dark">다크</option>
        </select>
      </label>
      <label className={rowClass}>
        <span className={labelClass}>본문 폰트 크기</span>
        <input
          type="number"
          min={12}
          max={24}
          className={numberInputClass}
          value={local.font_size}
          onChange={(e) => setLocal({ ...local, font_size: Number(e.target.value) })}
        />
      </label>
      <label className={rowClass}>
        <span className={labelClass}>기본 BPM</span>
        <input
          type="number"
          min={40}
          max={200}
          className={numberInputClass}
          value={local.default_bpm}
          onChange={(e) => setLocal({ ...local, default_bpm: Number(e.target.value) })}
        />
      </label>
      <label className={rowClass}>
        <span className={labelClass}>Display 레이아웃</span>
        <select
          className={inputClass}
          value={local.display_layout}
          onChange={(e) => setLocal({ ...local, display_layout: e.target.value })}
        >
          <option value="default">기본</option>
          <option value="wide">와이드</option>
          <option value="compact">컴팩트</option>
        </select>
      </label>

    </>
  );
}



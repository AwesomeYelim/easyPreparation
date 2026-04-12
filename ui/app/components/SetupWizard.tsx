"use client";

import { useState } from "react";
import { useAuth } from "@/lib/LocalAuthContext";

export default function SetupWizard() {
  const { needsSetup, isLoading, completeSetup, setupError } = useAuth();
  const [name, setName] = useState("");
  const [englishName, setEnglishName] = useState("");
  const [saving, setSaving] = useState(false);

  if (isLoading || !needsSetup) return null;

  const handleSubmit = async () => {
    if (!name.trim() || !englishName.trim()) return;
    setSaving(true);
    try {
      await completeSetup(name.trim(), englishName.trim());
    } catch (e) {
      console.error("Setup failed:", e);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[20000] flex items-center justify-center bg-black/40">
      <div className="w-[90%] max-w-[420px] bg-white rounded-xl shadow-2xl flex flex-col gap-4 p-7 border border-[#ccc]">
        <h2 className="text-center text-xl font-bold mb-1">교회 정보 설정</h2>
        <p className="text-center text-sm text-[#666] m-0">
          처음 사용하시나요? 교회 정보를 입력해주세요.
        </p>
        <input
          placeholder="교회 이름 (한글)"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="w-full px-3 py-2.5 border border-outline rounded-lg bg-surface-low text-on-surface text-sm focus:ring-2 focus:ring-electric-blue focus:outline-none"
        />
        <input
          placeholder="영문 이름 (예: Sarang Church)"
          value={englishName}
          onChange={(e) => setEnglishName(e.target.value)}
          className="w-full px-3 py-2.5 border border-outline rounded-lg bg-surface-low text-on-surface text-sm focus:ring-2 focus:ring-electric-blue focus:outline-none"
        />
        <button
          onClick={handleSubmit}
          disabled={saving || !name.trim() || !englishName.trim()}
          className={`w-full py-3 font-bold text-white rounded-lg text-sm transition-colors ${
            saving
              ? "bg-[#999] cursor-default"
              : "bg-electric-blue hover:bg-secondary cursor-pointer"
          } disabled:opacity-60`}
        >
          {saving ? "저장 중..." : "시작하기"}
        </button>
        {setupError && (
          <p className="text-[#e00] text-xs text-center m-0">
            오류: {setupError}
          </p>
        )}
      </div>
    </div>
  );
}

"use client";

import { useState } from "react";
import { useAuth } from "@/lib/LocalAuthContext";

// Step indicator
function StepDot({ index, current }: { index: number; current: number }) {
  const done = index < current;
  const active = index === current;
  return (
    <div
      className={`w-2.5 h-2.5 rounded-full transition-colors ${
        done
          ? "bg-electric-blue"
          : active
          ? "bg-[#1E293B]"
          : "bg-[#CBD5E1]"
      }`}
    />
  );
}

const inputCls =
  "w-full px-3 py-2.5 border border-outline rounded-lg bg-surface-low text-on-surface text-sm focus:ring-2 focus:ring-electric-blue focus:outline-none";

const primaryBtnCls =
  "flex-1 py-2.5 font-bold text-white rounded-lg text-sm transition-colors bg-electric-blue hover:bg-secondary cursor-pointer disabled:opacity-60 disabled:cursor-default";

const secondaryBtnCls =
  "flex-1 py-2.5 font-semibold text-[#475569] rounded-lg text-sm transition-colors bg-[#F1F5F9] hover:bg-[#E2E8F0] cursor-pointer";

export default function SetupWizard() {
  const { needsSetup, isLoading, completeSetup, setupError } = useAuth();

  // Step 1 — 교회 정보
  const [name, setName] = useState("");
  const [englishName, setEnglishName] = useState("");

  // Step 2 — OBS 연결
  const [obsHost, setObsHost] = useState("localhost");
  const [obsPort, setObsPort] = useState("4455");
  const [obsPassword, setObsPassword] = useState("");
  const [obsConnecting, setObsConnecting] = useState(false);
  const [obsError, setObsError] = useState<string | null>(null);

  // Step 3 — 완료
  const [step, setStep] = useState(0); // 0 | 1 | 2
  const [saving, setSaving] = useState(false);

  if (isLoading || !needsSetup) return null;

  // ── handlers ──────────────────────────────────────────────────────────────

  const handleStep1Next = () => {
    if (!name.trim() || !englishName.trim()) return;
    setStep(1);
  };

  const handleObsConnect = async () => {
    setObsError(null);
    setObsConnecting(true);
    try {
      const res = await fetch("/api/obs/auto-configure", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          host: obsHost,
          port: Number(obsPort),
          password: obsPassword,
        }),
      });
      if (!res.ok) {
        const text = await res.text().catch(() => `오류 ${res.status}`);
        let msg = text;
        try {
          const j = JSON.parse(text);
          if (j.error) msg = j.error;
        } catch {
          // ignore
        }
        setObsError(msg);
        return;
      }
      setStep(2);
    } catch (e) {
      setObsError("연결에 실패했습니다. 주소·포트·비밀번호를 확인해주세요.");
    } finally {
      setObsConnecting(false);
    }
  };

  const handleFinish = async () => {
    setSaving(true);
    try {
      await completeSetup(name.trim(), englishName.trim());
    } catch (e) {
      console.error("Setup failed:", e);
    } finally {
      setSaving(false);
    }
  };

  // ── render ─────────────────────────────────────────────────────────────────

  return (
    <div className="fixed inset-0 z-[20000] flex items-center justify-center bg-black/40">
      <div className="w-[90%] max-w-[460px] bg-white rounded-xl shadow-2xl flex flex-col gap-5 p-7 border border-[#ccc]">
        {/* Step indicator */}
        <div className="flex items-center justify-center gap-2">
          {[0, 1, 2].map((i) => (
            <StepDot key={i} index={i} current={step} />
          ))}
        </div>

        {/* ── STEP 1 ── */}
        {step === 0 && (
          <>
            <div>
              <h2 className="text-center text-xl font-bold mb-1">교회 정보 설정</h2>
              <p className="text-center text-sm text-[#666]">
                처음 사용하시나요? 교회 정보를 입력해주세요.
              </p>
            </div>
            <input
              placeholder="교회 이름 (한글)"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className={inputCls}
            />
            <input
              placeholder="영문 이름 (예: Sarang Church)"
              value={englishName}
              onChange={(e) => setEnglishName(e.target.value)}
              className={inputCls}
              onKeyDown={(e) => e.key === "Enter" && handleStep1Next()}
            />
            <button
              onClick={handleStep1Next}
              disabled={!name.trim() || !englishName.trim()}
              className={`${primaryBtnCls} w-full`}
            >
              다음 →
            </button>
          </>
        )}

        {/* ── STEP 2 ── */}
        {step === 1 && (
          <>
            <div>
              <h2 className="text-center text-xl font-bold mb-1">OBS 연결 설정</h2>
              <p className="text-center text-sm text-[#666]">
                OBS Studio와 연결하면 방송 송출을 자동으로 제어할 수 있어요.
              </p>
            </div>
            <div className="flex gap-2">
              <input
                placeholder="OBS 주소"
                value={obsHost}
                onChange={(e) => setObsHost(e.target.value)}
                className={`${inputCls} flex-1`}
              />
              <input
                placeholder="포트"
                value={obsPort}
                onChange={(e) => setObsPort(e.target.value)}
                className={`${inputCls} w-24`}
                maxLength={5}
              />
            </div>
            <input
              type="password"
              placeholder="비밀번호 (설정 안 했으면 비워두세요)"
              value={obsPassword}
              onChange={(e) => setObsPassword(e.target.value)}
              className={inputCls}
            />
            {obsError && (
              <p className="text-[#e00] text-xs text-center -mt-2">
                {obsError}
                {" "}
                <button
                  onClick={() => { setObsError(null); setStep(2); }}
                  className="underline text-[#666] ml-1"
                >
                  건너뛰기
                </button>
              </p>
            )}
            <div className="flex gap-2">
              <button onClick={() => setStep(0)} className={secondaryBtnCls}>
                ← 이전
              </button>
              <button
                onClick={() => { setObsError(null); setStep(2); }}
                className={secondaryBtnCls}
              >
                건너뛰기
              </button>
              <button
                onClick={handleObsConnect}
                disabled={obsConnecting}
                className={primaryBtnCls}
              >
                {obsConnecting ? "연결 중..." : "연결 테스트 →"}
              </button>
            </div>
          </>
        )}

        {/* ── STEP 3 ── */}
        {step === 2 && (
          <>
            <div>
              <h2 className="text-center text-xl font-bold mb-1">
                모든 준비가 완료됐어요! 🎉
              </h2>
            </div>
            <div className="flex flex-col gap-3">
              {[
                {
                  icon: "newspaper",
                  text: "주보 탭에서 예배 순서를 구성하세요",
                },
                {
                  icon: "cast",
                  text: "Space 키 또는 → 키로 슬라이드를 넘기세요",
                },
                {
                  icon: "settings",
                  text: "설정 > Display에서 배경과 폰트를 꾸미세요",
                },
              ].map(({ icon, text }) => (
                <div
                  key={icon}
                  className="flex items-center gap-3 p-3 bg-[#F8FAFC] rounded-lg border border-[#E2E8F0]"
                >
                  <span className="material-symbols-outlined text-electric-blue text-xl leading-none shrink-0">
                    {icon}
                  </span>
                  <span className="text-sm text-[#1E293B]">{text}</span>
                </div>
              ))}
            </div>
            {setupError && (
              <p className="text-[#e00] text-xs text-center -mt-2">
                오류: {setupError}
              </p>
            )}
            <div className="flex gap-2">
              <button onClick={() => setStep(1)} className={secondaryBtnCls}>
                ← 이전
              </button>
              <button
                onClick={handleFinish}
                disabled={saving}
                className={primaryBtnCls}
              >
                {saving ? "저장 중..." : "시작하기"}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

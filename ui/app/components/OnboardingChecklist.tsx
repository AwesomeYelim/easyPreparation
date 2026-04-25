"use client";

import { useState, useEffect } from "react";
import { useRecoilValue } from "recoil";
import { displayItemsState, worshipOrderState, userInfoState } from "@/recoilState";

const STORAGE_KEY = "ep_onboarding_dismissed";

export default function OnboardingChecklist() {
  const displayItems = useRecoilValue(displayItemsState);
  const worshipOrder = useRecoilValue(worshipOrderState);
  const userInfo = useRecoilValue(userInfoState);
  const [dismissed, setDismissed] = useState(true); // 초기값 true로 flicker 방지

  useEffect(() => {
    const val = localStorage.getItem(STORAGE_KEY);
    setDismissed(val === "true");
  }, []);

  const displaySent = displayItems.length > 0;
  const hasAnyOrder = Object.values(worshipOrder).some((arr) => arr.length > 0);
  const hasChurchName = !!userInfo.name;

  const steps = [
    {
      id: "church",
      label: "교회 이름 설정",
      desc: "설정 → 교회 정보",
      done: hasChurchName,
    },
    {
      id: "order",
      label: "예배 순서 구성",
      desc: "상단 탭 → 예배 편집",
      done: hasAnyOrder,
    },
    {
      id: "display",
      label: "프로젝터로 전송",
      desc: "Display 전송 버튼",
      done: displaySent,
    },
  ];

  const allDone = steps.every((s) => s.done);

  // 모두 완료되면 자동 숨김
  useEffect(() => {
    if (allDone) {
      localStorage.setItem(STORAGE_KEY, "true");
      setDismissed(true);
    }
  }, [allDone]);

  // 이미 dismiss됐거나 display에 항목이 있으면 표시 안 함
  if (dismissed || displaySent) return null;

  const handleDismiss = () => {
    localStorage.setItem(STORAGE_KEY, "true");
    setDismissed(true);
  };

  const completedCount = steps.filter((s) => s.done).length;

  return (
    <div className="mb-6 bg-pro-surface border border-pro-border rounded-xl overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-pro-border">
        <div className="flex items-center gap-2">
          <span
            className="material-symbols-outlined text-[#4a9eff]"
            style={{ fontSize: "16px" }}
          >
            rocket_launch
          </span>
          <span className="text-[12px] font-semibold text-pro-text">
            시작 가이드
          </span>
          <span className="text-[10px] text-pro-text-dim">
            {completedCount}/{steps.length} 완료
          </span>
        </div>
        <button
          onClick={handleDismiss}
          className="text-pro-text-dim hover:text-pro-text transition-colors text-[11px] px-1.5 py-0.5 rounded hover:bg-pro-hover cursor-pointer"
          title="닫기"
        >
          ✕
        </button>
      </div>

      {/* Progress bar */}
      <div className="h-0.5 bg-pro-elevated">
        <div
          className="h-full bg-[#4a9eff] transition-all duration-500"
          style={{ width: `${(completedCount / steps.length) * 100}%` }}
        />
      </div>

      {/* Steps */}
      <div className="flex gap-0 divide-x divide-pro-border">
        {steps.map((step) => (
          <div
            key={step.id}
            className={`flex-1 flex items-start gap-2.5 px-4 py-3 transition-colors ${
              step.done ? "opacity-50" : ""
            }`}
          >
            {/* Check icon */}
            <div
              className={`w-4 h-4 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5 border ${
                step.done
                  ? "bg-[#4a9eff] border-[#4a9eff]"
                  : "border-pro-border bg-transparent"
              }`}
            >
              {step.done && (
                <span
                  className="material-symbols-outlined text-white"
                  style={{ fontSize: "10px", fontVariationSettings: "'wght' 700" }}
                >
                  check
                </span>
              )}
            </div>

            {/* Text */}
            <div className="min-w-0">
              <div
                className={`text-[11px] font-semibold leading-tight ${
                  step.done ? "text-pro-text-dim line-through" : "text-pro-text"
                }`}
              >
                {step.label}
              </div>
              <div className="text-[10px] text-pro-text-dim mt-0.5 leading-tight">
                {step.desc}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

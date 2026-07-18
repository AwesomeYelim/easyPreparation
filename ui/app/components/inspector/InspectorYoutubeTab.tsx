"use client";

import { useState, useEffect, useMemo } from "react";
import { apiClient } from "@/lib/apiClient";
import { ThumbnailConfig } from "@/types";
import { useAutoSave } from "./useAutoSave";

const WORSHIP_TYPES: { key: string; label: string }[] = [
  { key: "main_worship", label: "주일예배" },
  { key: "after_worship", label: "오후예배" },
  { key: "wed_worship", label: "수요예배" },
  { key: "fri_worship", label: "금요예배" },
];

// Go의 thumbnail.FormatTitle과 동일한 치환 로직 (미리보기용 재현)
function weekOrdinal(date: Date): string {
  const day = date.getDate();
  const week = Math.floor((day - 1) / 7) + 1;
  const ordinals = ["첫째주", "둘째주", "셋째주", "넷째주", "다섯째주"];
  return ordinals[week - 1] || `${week}째주`;
}

function formatTemplate(format: string, date: Date): string {
  return format
    .replaceAll("{year}", String(date.getFullYear()))
    .replaceAll("{month}", String(date.getMonth() + 1))
    .replaceAll("{day}", String(date.getDate()))
    .replaceAll("{weekOrd}", weekOrdinal(date));
}

export default function InspectorYoutubeTab() {
  const [thumbConfig, setThumbConfig] = useState<ThumbnailConfig | null>(null);
  const today = useMemo(() => new Date(), []);

  useEffect(() => {
    apiClient.getThumbnailConfig().then(setThumbConfig).catch(console.error);
  }, []);

  useAutoSave(thumbConfig, (c) => apiClient.saveThumbnailConfig(c).catch(console.error));

  if (!thumbConfig) {
    return <div className="text-pro-text-dim text-center py-5 text-xs">로딩 중...</div>;
  }

  const updateField = (type: string, field: "titleFormat" | "descriptionFormat", value: string) => {
    setThumbConfig((prev) => {
      if (!prev) return prev;
      const current = prev.defaults[type] ?? { background: "", titleFormat: "" };
      return {
        ...prev,
        defaults: {
          ...prev.defaults,
          [type]: { ...current, [field]: value },
        },
      };
    });
  };

  const inputClass =
    "w-full px-2 py-1.5 border border-white/20 rounded-md text-[11px] bg-white/10 text-white outline-none placeholder:text-[#666] focus:border-electric-blue";

  return (
    <div className="flex flex-col gap-4">
      <div className="text-[10px] text-[#888]">
        방송 시작 시 유튜브에 자동 반영되는 제목·설명 템플릿입니다. 매번 직접 입력할 필요 없이 아래 변수가 자동으로 채워집니다.
        <div className="mt-1 flex flex-wrap gap-1">
          {["{year}", "{month}", "{day}", "{weekOrd}"].map((v) => (
            <code key={v} className="px-1.5 py-0.5 bg-white/10 rounded text-[9px] text-[#9ac2ff]">{v}</code>
          ))}
        </div>
      </div>

      {WORSHIP_TYPES.map(({ key, label }) => {
        const theme = thumbConfig.defaults[key];
        const titleFormat = theme?.titleFormat ?? "";
        const descriptionFormat = theme?.descriptionFormat ?? "";
        return (
          <div key={key} className="flex flex-col gap-1.5 px-2 py-2.5 bg-white/5 rounded-lg border border-white/10">
            <div className="text-[11px] font-semibold text-pro-text">{label}</div>

            <div className="flex flex-col gap-1">
              <span className="text-[9px] text-[#888]">제목 형식</span>
              <input
                type="text"
                value={titleFormat}
                placeholder="예: {month}월 {weekOrd} 주일예배"
                onChange={(e) => updateField(key, "titleFormat", e.target.value)}
                className={inputClass}
              />
              {titleFormat && (
                <div className="text-[9px] text-[#60a5fa] truncate">→ {formatTemplate(titleFormat, today)}</div>
              )}
            </div>

            <div className="flex flex-col gap-1">
              <span className="text-[9px] text-[#888]">설명 (선택)</span>
              <textarea
                value={descriptionFormat}
                placeholder="예: {year}년 {month}월 {weekOrd} 주일예배입니다."
                onChange={(e) => updateField(key, "descriptionFormat", e.target.value)}
                rows={4}
                className={`${inputClass} resize-y max-h-40 overflow-y-auto`}
              />
              {descriptionFormat && (
                <div className="text-[9px] text-[#60a5fa] whitespace-pre-wrap max-h-24 overflow-y-auto break-words border-t border-white/10 pt-1">
                  → {formatTemplate(descriptionFormat, today)}
                </div>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}

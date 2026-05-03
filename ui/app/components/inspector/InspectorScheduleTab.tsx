"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/apiClient";
import { ScheduleConfig, ScheduleEntry } from "@/types";
import FeatureGate from "@/components/FeatureGate";
import { useAutoSave } from "./useAutoSave";

const WEEKDAY_NAMES = ["일", "월", "화", "수", "목", "금", "토"];

export default function InspectorScheduleTab() {
  const [scheduleConfig, setScheduleConfig] = useState<ScheduleConfig | null>(null);

  useEffect(() => {
    apiClient.getSchedule().then(setScheduleConfig).catch(console.error);
  }, []);

  useAutoSave(
    scheduleConfig,
    (c) => apiClient.saveSchedule(c).catch(console.error)
  );

  if (!scheduleConfig) {
    return <div className="text-[#888] text-center py-5 text-xs">로딩 중...</div>;
  }

  const updateEntry = (idx: number, patch: Partial<ScheduleEntry>) => {
    setScheduleConfig({
      ...scheduleConfig,
      entries: scheduleConfig.entries.map((e, i) => (i === idx ? { ...e, ...patch } : e)),
    });
  };

  return (
    <FeatureGate feature="auto_scheduler">
      <div className="flex flex-col gap-3">
        {scheduleConfig.entries.map((entry, i) => (
          <div key={entry.worshipType} className="flex items-center gap-2">
            <label className="flex items-center gap-1.5 flex-1 cursor-pointer">
              <input
                type="checkbox"
                className="w-3.5 h-3.5 accent-[#4a9eff]"
                checked={entry.enabled}
                onChange={(e) => updateEntry(i, { enabled: e.target.checked })}
              />
              <span className="text-[11px] font-medium text-pro-text">{entry.label}</span>
            </label>
            <span className="text-[10px] text-pro-text-dim min-w-[32px] text-center">
              {WEEKDAY_NAMES[entry.weekday]}요일
            </span>
            <input
              type="time"
              className="px-2 py-1 border border-pro-border rounded-md text-[10px] bg-pro-elevated text-pro-text outline-none"
              value={`${String(entry.hour).padStart(2, "0")}:${String(entry.minute).padStart(2, "0")}`}
              onChange={(e) => {
                const [h, m] = e.target.value.split(":").map(Number);
                updateEntry(i, { hour: h, minute: m });
              }}
            />
          </div>
        ))}

        <div className="h-px bg-pro-border my-1" />

        {/* 카운트다운 */}
        <div className="flex justify-between items-center">
          <span className="text-[11px] font-medium text-pro-text">사전 카운트다운</span>
          <div className="flex items-center gap-1.5">
            <input
              type="number"
              min={1}
              max={30}
              value={scheduleConfig.countdownMinutes}
              onChange={(e) =>
                setScheduleConfig({ ...scheduleConfig, countdownMinutes: Number(e.target.value) })
              }
              className="w-14 px-2 py-1 border border-pro-border rounded-md text-[10px] text-center bg-pro-elevated text-pro-text outline-none"
            />
            <span className="text-[10px] text-pro-text-dim">분</span>
          </div>
        </div>

        {/* OBS 자동 스트리밍 */}
        <div className="flex justify-between items-center">
          <span className="text-[11px] font-medium text-pro-text">OBS 자동 스트리밍</span>
          <button
            className={`px-3.5 py-1 rounded-md text-[10px] font-semibold cursor-pointer border-none transition-colors ${
              scheduleConfig.autoStream
                ? "bg-[#4a9eff] text-white"
                : "bg-pro-hover text-pro-text-dim"
            }`}
            onClick={() =>
              setScheduleConfig({ ...scheduleConfig, autoStream: !scheduleConfig.autoStream })
            }
          >
            {scheduleConfig.autoStream ? "ON" : "OFF"}
          </button>
        </div>

        <div className="h-px bg-pro-border my-1" />
      </div>
    </FeatureGate>
  );
}

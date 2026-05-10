"use client";

import { useState, useEffect } from "react";
import { useSetRecoilState } from "recoil";
import { scheduleActiveState } from "@/recoilState";
import { apiClient } from "@/lib/apiClient";
import { ScheduleConfig, ScheduleEntry } from "@/types";
import FeatureGate from "@/components/FeatureGate";
import { useAutoSave } from "./useAutoSave";

const WEEKDAY_NAMES = ["일", "월", "화", "수", "목", "금", "토"];

export default function InspectorScheduleTab() {
  const [scheduleConfig, setScheduleConfig] = useState<ScheduleConfig | null>(null);
  const setScheduleActive = useSetRecoilState(scheduleActiveState);

  useEffect(() => {
    apiClient.getSchedule().then((conf) => {
      setScheduleConfig(conf);
      setScheduleActive((conf.entries || []).some((e: ScheduleEntry) => e.enabled));
    }).catch(console.error);
  }, []);

  useAutoSave(
    scheduleConfig,
    (c) => apiClient.saveSchedule(c).catch(console.error)
  );

  if (!scheduleConfig || !scheduleConfig.entries) {
    return <div className="text-[#888] text-center py-5 text-xs">로딩 중...</div>;
  }

  const updateEntry = (idx: number, patch: Partial<ScheduleEntry>) => {
    const updated = {
      ...scheduleConfig,
      entries: scheduleConfig.entries.map((e, i) => (i === idx ? { ...e, ...patch } : e)),
    };
    setScheduleConfig(updated);
    setScheduleActive(updated.entries.some((e) => e.enabled));
  };

  return (
    <FeatureGate feature="auto_scheduler">
      <div className="flex flex-col gap-3">
        {scheduleConfig.entries.map((entry, i) => (
          <div key={entry.worshipType} className="flex items-center gap-2">
            <button
              className={`relative w-9 h-5 rounded-full transition-all flex-shrink-0 ${
                entry.enabled ? "bg-[#4a9eff]" : "bg-white/20"
              }`}
              onClick={() => updateEntry(i, { enabled: !entry.enabled })}
            >
              <div className={`absolute top-0.5 w-4 h-4 rounded-full bg-white shadow transition-all ${
                entry.enabled ? "left-[18px]" : "left-0.5"
              }`} />
            </button>
            <span className={`text-[11px] font-medium flex-1 ${entry.enabled ? "text-pro-text" : "text-pro-text-dim"}`}>
              {entry.label}
            </span>
            <span className="text-[10px] text-pro-text-dim min-w-[32px] text-center">
              {WEEKDAY_NAMES[entry.weekday]}
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
            className={`relative w-11 h-6 rounded-full transition-all ${
              scheduleConfig.autoStream ? "bg-[#4a9eff]" : "bg-white/20"
            }`}
            onClick={() =>
              setScheduleConfig({ ...scheduleConfig, autoStream: !scheduleConfig.autoStream })
            }
          >
            <div className={`absolute top-0.5 w-5 h-5 rounded-full bg-white shadow transition-all ${
              scheduleConfig.autoStream ? "left-[22px]" : "left-0.5"
            }`} />
          </button>
        </div>

        <div className="h-px bg-pro-border my-1" />
      </div>
    </FeatureGate>
  );
}

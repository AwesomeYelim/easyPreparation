"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/apiClient";
import { ScheduleConfig, ScheduleEntry } from "@/types";
import FeatureGate from "./FeatureGate";

const WEEKDAY_NAMES = ["일", "월", "화", "수", "목", "금", "토"];

interface SchedulePanelProps {
  open: boolean;
  onClose: () => void;
}

export default function SchedulePanel({ open, onClose }: SchedulePanelProps) {
  const [config, setConfig] = useState<ScheduleConfig | null>(null);
  const [saving, setSaving] = useState(false);
  useEffect(() => {
    if (!open) return;
    apiClient.getSchedule().then(setConfig).catch(console.error);
  }, [open]);

  if (!open || !config) return null;

  const updateEntry = (idx: number, patch: Partial<ScheduleEntry>) => {
    setConfig({
      ...config,
      entries: config.entries.map((e, i) =>
        i === idx ? { ...e, ...patch } : e
      ),
    });
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await apiClient.saveSchedule(config);
      onClose();
    } catch (e) {
      console.error("저장 에러:", e);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div
      className="fixed inset-0 z-[11000] flex items-center justify-center bg-black/50"
      onClick={onClose}
    >
      <div
        className="bg-[var(--surface-elevated)] rounded-2xl w-[500px] max-w-[90vw] max-h-[85vh] overflow-y-auto shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        {/* header */}
        <div className="flex justify-between items-center px-6 pt-5 pb-3 border-b border-[var(--border)]">
          <h3 className="m-0 text-lg font-bold text-[var(--text-primary)]">정기 스케줄</h3>
          <button
            className="bg-transparent border-none text-2xl cursor-pointer text-[var(--text-secondary)] leading-none"
            onClick={onClose}
          >
            &times;
          </button>
        </div>

        <FeatureGate feature="auto_scheduler">
          {/* body */}
          <div className="px-6 py-5 flex flex-col gap-3.5">
            {config.entries.map((entry, i) => (
              <div key={entry.worshipType} className="flex items-center gap-3">
                <label className="flex items-center gap-2 flex-1 cursor-pointer">
                  <input
                    type="checkbox"
                    className="w-4 h-4 accent-[var(--accent)]"
                    checked={entry.enabled}
                    onChange={(e) => updateEntry(i, { enabled: e.target.checked })}
                  />
                  <span className="text-sm font-medium text-[var(--text-primary)]">
                    {entry.label}
                  </span>
                </label>
                <span className="text-xs text-[var(--text-secondary)] min-w-[44px] text-center">
                  {WEEKDAY_NAMES[entry.weekday]}요일
                </span>
                <input
                  type="time"
                  className="px-2 py-1 border border-[var(--border-input)] rounded-md text-xs bg-[var(--surface-input)]"
                  value={`${String(entry.hour).padStart(2, "0")}:${String(entry.minute).padStart(2, "0")}`}
                  onChange={(e) => {
                    const [h, m] = e.target.value.split(":").map(Number);
                    updateEntry(i, { hour: h, minute: m });
                  }}
                />
              </div>
            ))}

            <div className="h-px bg-[var(--border)] my-1" />

            {/* 카운트다운 */}
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium text-[var(--text-primary)]">사전 카운트다운</span>
              <div className="flex items-center gap-1.5">
                <input
                  type="number"
                  min={1}
                  max={30}
                  value={config.countdownMinutes}
                  onChange={(e) =>
                    setConfig({ ...config, countdownMinutes: Number(e.target.value) })
                  }
                  className="w-16 px-2 py-1 border border-[var(--border-input)] rounded-md text-xs text-center bg-[var(--surface-input)]"
                />
                <span className="text-xs text-[var(--text-secondary)]">분</span>
              </div>
            </div>

            {/* OBS 자동 스트리밍 */}
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium text-[var(--text-primary)]">OBS 자동 스트리밍</span>
              <button
                className={`px-4 py-1.5 rounded-md text-xs font-semibold cursor-pointer border-none transition-colors ${
                  config.autoStream
                    ? "bg-[var(--accent)] text-[var(--surface-elevated)]"
                    : "bg-[var(--border)] text-[var(--text-secondary)]"
                }`}
                onClick={() =>
                  setConfig({ ...config, autoStream: !config.autoStream })
                }
              >
                {config.autoStream ? "ON" : "OFF"}
              </button>
            </div>
          </div>

          {/* footer */}
          <div className="flex justify-end gap-2 px-6 pb-5 pt-4 border-t border-[var(--border)]">
            <button
              className="px-5 py-2 text-sm bg-[var(--surface-hover)] border border-[var(--border-input)] rounded-xl cursor-pointer"
              onClick={onClose}
            >
              취소
            </button>
            <button
              className="px-5 py-2 text-sm font-semibold bg-[var(--accent)] text-[var(--surface-elevated)] border-none rounded-xl cursor-pointer hover:bg-[var(--accent-hover)] disabled:bg-[var(--text-muted)]"
              onClick={handleSave}
              disabled={saving}
            >
              {saving ? "저장 중..." : "저장"}
            </button>
          </div>
        </FeatureGate>
      </div>
    </div>
  );
}

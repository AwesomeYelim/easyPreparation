"use client";

import { useState, useEffect, useCallback } from "react";
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
    <div className="sched_overlay" onClick={onClose}>
      <div className="sched_panel" onClick={(e) => e.stopPropagation()}>
        <div className="sched_header">
          <h3>정기 스케줄</h3>
          <button className="sched_close" onClick={onClose}>
            &times;
          </button>
        </div>

        <FeatureGate feature="auto_scheduler">
          <div className="sched_body">
            {config.entries.map((entry, i) => (
              <div key={entry.worshipType} className="sched_entry">
                <label className="sched_check">
                  <input
                    type="checkbox"
                    checked={entry.enabled}
                    onChange={(e) => updateEntry(i, { enabled: e.target.checked })}
                  />
                  <span className="sched_label">{entry.label}</span>
                </label>
                <span className="sched_day">{WEEKDAY_NAMES[entry.weekday]}요일</span>
                <input
                  type="time"
                  className="sched_time"
                  value={`${String(entry.hour).padStart(2, "0")}:${String(entry.minute).padStart(2, "0")}`}
                  onChange={(e) => {
                    const [h, m] = e.target.value.split(":").map(Number);
                    updateEntry(i, { hour: h, minute: m });
                  }}
                />
              </div>
            ))}

            <div style={{ height: 1, background: "var(--border)", margin: "4px 0" }} />

            <div className="sched_option">
              <span>사전 카운트다운</span>
              <div className="sched_option_input">
                <input
                  type="number"
                  min={1}
                  max={30}
                  value={config.countdownMinutes}
                  onChange={(e) =>
                    setConfig({ ...config, countdownMinutes: Number(e.target.value) })
                  }
                />
                <span>분</span>
              </div>
            </div>

            <div className="sched_option">
              <span>OBS 자동 스트리밍</span>
              <button
                className={`sched_auto_btn ${config.autoStream ? "on" : ""}`}
                onClick={() =>
                  setConfig({ ...config, autoStream: !config.autoStream })
                }
              >
                {config.autoStream ? "ON" : "OFF"}
              </button>
            </div>
          </div>

          <div className="sched_footer">
            <button className="sched_cancel_btn" onClick={onClose}>
              취소
            </button>
            <button
              className="sched_save_btn"
              onClick={handleSave}
              disabled={saving}
            >
              {saving ? "저장 중..." : "저장"}
            </button>
          </div>
        </FeatureGate>
      </div>

      <style jsx>{`
        .sched_overlay {
          position: fixed;
          top: 0; left: 0; width: 100%; height: 100%;
          background: rgba(0,0,0,0.5);
          display: flex; align-items: center; justify-content: center;
          z-index: 11000;
        }
        .sched_panel {
          background: var(--surface-elevated);
          border-radius: 16px;
          width: 500px;
          max-width: 90vw;
          max-height: 85vh;
          overflow-y: auto;
          box-shadow: 0 8px 32px rgba(0,0,0,0.2);
        }
        .sched_header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 20px 24px 12px;
          border-bottom: 1px solid var(--border);
        }
        .sched_header h3 { margin: 0; font-size: 18px; font-weight: 700; color: var(--text-primary); }
        .sched_close {
          background: none; border: none; font-size: 24px;
          cursor: pointer; color: var(--text-secondary); line-height: 1;
        }
        .sched_body {
          padding: 20px 24px;
          display: flex; flex-direction: column; gap: 14px;
        }
        .sched_entry {
          display: flex; align-items: center; gap: 12px;
        }
        .sched_check {
          display: flex; align-items: center; gap: 8px; flex: 1;
        }
        .sched_check input[type="checkbox"] {
          width: 18px; height: 18px; accent-color: var(--accent);
        }
        .sched_label { font-size: 14px; font-weight: 500; color: var(--text-primary); }
        .sched_day { font-size: 13px; color: var(--text-secondary); min-width: 44px; text-align: center; }
        .sched_time {
          padding: 4px 8px; border: 1px solid var(--border-input);
          border-radius: 6px; font-size: 13px; background: var(--surface-input);
        }
        .sched_option {
          display: flex; justify-content: space-between; align-items: center;
        }
        .sched_option > span { font-size: 14px; color: var(--text-primary); font-weight: 500; }
        .sched_option_input { display: flex; align-items: center; gap: 6px; }
        .sched_option_input input {
          width: 60px; padding: 4px 8px; border: 1px solid var(--border-input);
          border-radius: 6px; font-size: 13px; text-align: center; background: var(--surface-input);
        }
        .sched_option_input span { font-size: 13px; color: var(--text-secondary); }
        .sched_auto_btn {
          padding: 6px 16px; border-radius: 6px; font-size: 13px;
          font-weight: 600; cursor: pointer; border: none;
          transition: background 0.15s;
          background: var(--border); color: var(--text-secondary);
        }
        .sched_auto_btn.on { background: var(--accent); color: var(--surface-elevated); }
        .sched_footer {
          display: flex; justify-content: flex-end; gap: 8px;
          padding: 16px 24px 20px; border-top: 1px solid var(--border);
        }
        .sched_cancel_btn {
          padding: 8px 20px; font-size: 13px; background: var(--surface-hover);
          border: 1px solid var(--border-input); border-radius: 8px; cursor: pointer;
        }
        .sched_save_btn {
          padding: 8px 20px; font-size: 13px; font-weight: 600;
          background: var(--accent); color: var(--surface-elevated); border: none;
          border-radius: 8px; cursor: pointer;
        }
        .sched_save_btn:hover { background: var(--accent-hover); }
        .sched_save_btn:disabled { background: var(--text-muted); }
      `}</style>
    </div>
  );
}

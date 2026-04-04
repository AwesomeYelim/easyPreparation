"use client";

import { useState, useEffect, useCallback } from "react";
import { apiClient } from "@/lib/apiClient";
import { ScheduleConfig, ScheduleEntry } from "@/types";

const WEEKDAY_NAMES = ["일", "월", "화", "수", "목", "금", "토"];

interface SchedulePanelProps {
  open: boolean;
  onClose: () => void;
}

export default function SchedulePanel({ open, onClose }: SchedulePanelProps) {
  const [config, setConfig] = useState<ScheduleConfig | null>(null);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState<string | null>(null);

  const handleTest = useCallback(async (action: "countdown" | "trigger", worshipType: string) => {
    setTesting(`${action}_${worshipType}`);
    try {
      const res = await apiClient.scheduleTest(action, worshipType);
      if (res.ok) {
        alert(res.message);
      } else {
        alert("테스트 실패");
      }
    } catch (e) {
      console.error("스케줄 테스트 에러:", e);
      alert("테스트 요청 실패");
    } finally {
      setTesting(null);
    }
  }, []);

  useEffect(() => {
    if (open) {
      apiClient.getSchedule().then(setConfig).catch(console.error);
    }
  }, [open]);

  if (!open || !config) return null;

  const updateEntry = (idx: number, patch: Partial<ScheduleEntry>) => {
    setConfig({
      ...config,
      entries: config.entries.map((e, i) => (i === idx ? { ...e, ...patch } : e)),
    });
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await apiClient.saveSchedule(config);
      onClose();
    } catch (e) {
      console.error("스케줄 저장 에러:", e);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="sched_overlay" onClick={onClose}>
      <div className="sched_panel" onClick={(e) => e.stopPropagation()}>
        <div className="sched_header">
          <h3>스케줄 설정</h3>
          <button className="sched_close" onClick={onClose}>&times;</button>
        </div>

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
              <div className="sched_test_btns">
                <button
                  className="sched_test_btn"
                  disabled={testing !== null}
                  onClick={() => handleTest("countdown", entry.worshipType)}
                  title="10초 카운트다운 테스트"
                >
                  {testing === `countdown_${entry.worshipType}` ? "..." : "⏱"}
                </button>
                <button
                  className="sched_test_btn sched_test_trigger"
                  disabled={testing !== null}
                  onClick={() => handleTest("trigger", entry.worshipType)}
                  title="즉시 실행 테스트"
                >
                  {testing === `trigger_${entry.worshipType}` ? "..." : "▶"}
                </button>
              </div>
            </div>
          ))}

          <div className="sched_divider" />

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
              className={`sched_toggle ${config.autoStream ? "on" : "off"}`}
              onClick={() => setConfig({ ...config, autoStream: !config.autoStream })}
            >
              {config.autoStream ? "ON" : "OFF"}
            </button>
          </div>
        </div>

        <div className="sched_footer">
          <button className="sched_cancel_btn" onClick={onClose}>취소</button>
          <button className="sched_save_btn" onClick={handleSave} disabled={saving}>
            {saving ? "저장 중..." : "저장"}
          </button>
        </div>
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
          background: #fff;
          border-radius: 16px;
          width: 440px;
          max-width: 90vw;
          max-height: 80vh;
          overflow-y: auto;
          box-shadow: 0 8px 32px rgba(0,0,0,0.2);
        }
        .sched_header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 20px 24px 16px;
          border-bottom: 1px solid #e5e7eb;
        }
        .sched_header h3 {
          margin: 0; font-size: 18px; font-weight: 700; color: #1f2937;
        }
        .sched_close {
          background: none; border: none; font-size: 24px;
          cursor: pointer; color: #6b7280; line-height: 1;
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
          width: 18px; height: 18px; accent-color: #1f3f62;
        }
        .sched_label {
          font-size: 14px; font-weight: 500; color: #374151;
        }
        .sched_day {
          font-size: 13px; color: #6b7280; min-width: 44px; text-align: center;
        }
        .sched_time {
          padding: 4px 8px;
          border: 1px solid #d1d5db;
          border-radius: 6px;
          font-size: 13px;
          background: #f9fafb;
        }
        .sched_test_btns {
          display: flex; gap: 4px; margin-left: 4px;
        }
        .sched_test_btn {
          width: 28px; height: 28px;
          border: 1px solid #d1d5db; border-radius: 6px;
          background: #f9fafb; cursor: pointer;
          font-size: 12px; display: flex;
          align-items: center; justify-content: center;
          transition: background 0.15s;
        }
        .sched_test_btn:hover { background: #e5e7eb; }
        .sched_test_btn:disabled { opacity: 0.5; cursor: default; }
        .sched_test_trigger { color: #059669; }
        .sched_divider {
          height: 1px; background: #e5e7eb; margin: 4px 0;
        }
        .sched_option {
          display: flex; justify-content: space-between; align-items: center;
        }
        .sched_option > span {
          font-size: 14px; color: #374151; font-weight: 500;
        }
        .sched_option_input {
          display: flex; align-items: center; gap: 6px;
        }
        .sched_option_input input {
          width: 60px; padding: 4px 8px;
          border: 1px solid #d1d5db; border-radius: 6px;
          font-size: 13px; text-align: center; background: #f9fafb;
        }
        .sched_option_input span {
          font-size: 13px; color: #6b7280;
        }
        .sched_toggle {
          padding: 6px 16px; border-radius: 6px; font-size: 13px;
          font-weight: 600; cursor: pointer; border: none;
          transition: background 0.15s;
        }
        .sched_toggle.on {
          background: #1f3f62; color: #fff;
        }
        .sched_toggle.off {
          background: #e5e7eb; color: #6b7280;
        }
        .sched_footer {
          display: flex; justify-content: flex-end; gap: 8px;
          padding: 16px 24px 20px;
          border-top: 1px solid #e5e7eb;
        }
        .sched_cancel_btn {
          padding: 8px 20px; font-size: 13px;
          background: #f3f4f6; border: 1px solid #d1d5db;
          border-radius: 8px; cursor: pointer;
        }
        .sched_save_btn {
          padding: 8px 20px; font-size: 13px; font-weight: 600;
          background: #1f3f62; color: #fff; border: none;
          border-radius: 8px; cursor: pointer;
        }
        .sched_save_btn:hover { background: #2d5a8a; }
        .sched_save_btn:disabled { background: #9ca3af; }
      `}</style>
    </div>
  );
}

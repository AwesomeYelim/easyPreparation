"use client";

import { useState, useEffect } from "react";
import { useRecoilState, useRecoilValue } from "recoil";
import { userSettingsState, userInfoState } from "@/recoilState";
import { apiClient } from "@/lib/apiClient";
import { UserSettings } from "@/types";

interface SettingsPanelProps {
  open: boolean;
  onClose: () => void;
}

export default function SettingsPanel({ open, onClose }: SettingsPanelProps) {
  const userInfo = useRecoilValue(userInfoState);
  const [settings, setSettings] = useRecoilState(userSettingsState);
  const [local, setLocal] = useState<UserSettings>(settings);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    setLocal(settings);
  }, [settings, open]);

  const handleSave = async () => {
    if (!userInfo.email) return;
    setSaving(true);
    try {
      await apiClient.saveSettings(userInfo.email, local);
      setSettings(local);
      onClose();
    } catch (e) {
      console.error("설정 저장 에러:", e);
    } finally {
      setSaving(false);
    }
  };

  if (!open) return null;

  return (
    <div className="settings_overlay" onClick={onClose}>
      <div className="settings_panel" onClick={(e) => e.stopPropagation()}>
        <div className="settings_header">
          <h3>설정</h3>
          <button className="settings_close" onClick={onClose}>
            &times;
          </button>
        </div>

        <div className="settings_body">
          <label className="settings_row">
            <span>선호 성경 번역</span>
            <select
              value={local.preferred_bible_version}
              onChange={(e) =>
                setLocal({ ...local, preferred_bible_version: Number(e.target.value) })
              }
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

          <label className="settings_row">
            <span>테마</span>
            <select
              value={local.theme}
              onChange={(e) => setLocal({ ...local, theme: e.target.value })}
            >
              <option value="light">라이트</option>
              <option value="dark">다크</option>
            </select>
          </label>

          <label className="settings_row">
            <span>본문 폰트 크기</span>
            <input
              type="number"
              min={12}
              max={24}
              value={local.font_size}
              onChange={(e) =>
                setLocal({ ...local, font_size: Number(e.target.value) })
              }
            />
          </label>

          <label className="settings_row">
            <span>기본 BPM</span>
            <input
              type="number"
              min={40}
              max={200}
              value={local.default_bpm}
              onChange={(e) =>
                setLocal({ ...local, default_bpm: Number(e.target.value) })
              }
            />
          </label>

          <label className="settings_row">
            <span>Display 레이아웃</span>
            <select
              value={local.display_layout}
              onChange={(e) =>
                setLocal({ ...local, display_layout: e.target.value })
              }
            >
              <option value="default">기본</option>
              <option value="wide">와이드</option>
              <option value="compact">컴팩트</option>
            </select>
          </label>
        </div>

        <div className="settings_footer">
          <button className="settings_cancel_btn" onClick={onClose}>
            취소
          </button>
          <button
            className="settings_save_btn"
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? "저장 중..." : "저장"}
          </button>
        </div>
      </div>

      <style jsx>{`
        .settings_overlay {
          position: fixed;
          top: 0;
          left: 0;
          width: 100%;
          height: 100%;
          background: rgba(0, 0, 0, 0.5);
          display: flex;
          align-items: center;
          justify-content: center;
          z-index: 1000;
        }
        .settings_panel {
          background: #fff;
          border-radius: 16px;
          width: 420px;
          max-width: 90vw;
          max-height: 80vh;
          overflow-y: auto;
          box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
        }
        .settings_header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 20px 24px 16px;
          border-bottom: 1px solid #e5e7eb;
        }
        .settings_header h3 {
          margin: 0;
          font-size: 18px;
          font-weight: 700;
          color: #1f2937;
        }
        .settings_close {
          background: none;
          border: none;
          font-size: 24px;
          cursor: pointer;
          color: #6b7280;
          line-height: 1;
        }
        .settings_body {
          padding: 20px 24px;
          display: flex;
          flex-direction: column;
          gap: 16px;
        }
        .settings_row {
          display: flex;
          justify-content: space-between;
          align-items: center;
          gap: 12px;
        }
        .settings_row span {
          font-size: 14px;
          color: #374151;
          font-weight: 500;
        }
        .settings_row select,
        .settings_row input {
          padding: 6px 12px;
          border: 1px solid #d1d5db;
          border-radius: 8px;
          font-size: 13px;
          background: #f9fafb;
          min-width: 140px;
        }
        .settings_row input[type="number"] {
          width: 80px;
          min-width: 80px;
        }
        .settings_footer {
          display: flex;
          justify-content: flex-end;
          gap: 8px;
          padding: 16px 24px 20px;
          border-top: 1px solid #e5e7eb;
        }
        .settings_cancel_btn {
          padding: 8px 20px;
          font-size: 13px;
          background: #f3f4f6;
          border: 1px solid #d1d5db;
          border-radius: 8px;
          cursor: pointer;
        }
        .settings_save_btn {
          padding: 8px 20px;
          font-size: 13px;
          font-weight: 600;
          background: #1f3f62;
          color: #fff;
          border: none;
          border-radius: 8px;
          cursor: pointer;
        }
        .settings_save_btn:hover {
          background: #2d5a8a;
        }
        .settings_save_btn:disabled {
          background: #9ca3af;
        }
      `}</style>
    </div>
  );
}

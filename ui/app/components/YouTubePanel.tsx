"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/apiClient";
import { YouTubeStatus } from "@/types";
import FeatureGate from "./FeatureGate";

interface YouTubePanelProps {
  open: boolean;
  onClose: () => void;
}

export default function YouTubePanel({ open, onClose }: YouTubePanelProps) {
  const [ytStatus, setYtStatus] = useState<YouTubeStatus | null>(null);
  const [settingUp, setSettingUp] = useState(false);
  const [uploading, setUploading] = useState(false);

  useEffect(() => {
    if (!open) return;
    apiClient.getYoutubeStatus().then(setYtStatus).catch(console.error);
  }, [open]);

  if (!open) return null;

  const connected = ytStatus?.connected ?? false;

  const handleConnect = () => {
    window.open(apiClient.getYoutubeAuthUrl(), "yt_auth", "width=600,height=700");
    const check = setInterval(async () => {
      try {
        const s = await apiClient.getYoutubeStatus();
        if (s.connected) {
          setYtStatus(s);
          clearInterval(check);
        }
      } catch { /* ignore */ }
    }, 2000);
    setTimeout(() => clearInterval(check), 120000);
  };

  const handleSetupOBS = async () => {
    setSettingUp(true);
    try {
      const res = await apiClient.setupOBSStream("main_worship");
      if (res.ok) alert(res.message || "OBS 스트림 설정 완료!");
      else alert(`실패: ${res.error || "알 수 없는 오류"}`);
    } catch {
      alert("요청 실패");
    } finally {
      setSettingUp(false);
    }
  };

  const handleManualUpload = async () => {
    setUploading(true);
    try {
      const res = await apiClient.generateThumbnail("main_worship", undefined, true);
      if (res.ok) alert("썸네일 생성 + YouTube 업로드 요청 완료");
      else alert(`실패: ${res.error || "알 수 없는 오류"}`);
    } catch {
      alert("요청 실패");
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="yt_overlay" onClick={onClose}>
      <div className="yt_panel" onClick={(e) => e.stopPropagation()}>
        <div className="yt_header">
          <h3>YouTube</h3>
          <button className="yt_close" onClick={onClose}>&times;</button>
        </div>

        <FeatureGate feature="youtube_integration">
          <div className="yt_body">
            <div className="yt_status_row">
              <span className="yt_status_label">연결 상태</span>
              <span className={`yt_badge ${connected ? "on" : "off"}`}>
                {connected ? "연결됨" : "미연결"}
              </span>
            </div>

            {!connected && (
              <div style={{ textAlign: "center", padding: "20px 0" }}>
                <p style={{ fontSize: 13, color: "#6b7280", marginBottom: 16 }}>
                  YouTube 계정을 연결하면 예배 시작 시 썸네일이 자동으로 업로드됩니다.
                </p>
                <button className="yt_connect_btn" onClick={handleConnect}>
                  YouTube 계정 연결
                </button>
              </div>
            )}

            {connected && (
              <>
                <div className="yt_divider" />
                <div className="yt_section_title">OBS 스트리밍 설정</div>
                <p style={{ fontSize: 13, color: "#6b7280", margin: 0 }}>
                  YouTube 스트림 키를 자동으로 가져와 OBS에 세팅합니다.
                </p>
                <button
                  className="yt_action_btn green"
                  onClick={handleSetupOBS}
                  disabled={settingUp}
                >
                  {settingUp ? "설정 중..." : "OBS 스트림 자동 세팅"}
                </button>

                <div className="yt_divider" />
                <div className="yt_section_title">수동 업로드</div>
                <p style={{ fontSize: 13, color: "#6b7280", margin: 0 }}>
                  현재 활성/예정 라이브 방송에 썸네일을 즉시 생성하고 업로드합니다.
                </p>
                <button
                  className="yt_action_btn blue"
                  onClick={handleManualUpload}
                  disabled={uploading}
                >
                  {uploading ? "업로드 중..." : "썸네일 생성 + 업로드"}
                </button>
              </>
            )}
          </div>
        </FeatureGate>
      </div>

      <style jsx>{`
        .yt_overlay {
          position: fixed; top: 0; left: 0; width: 100%; height: 100%;
          background: rgba(0,0,0,0.5);
          display: flex; align-items: center; justify-content: center;
          z-index: 11000;
        }
        .yt_panel {
          background: #fff; border-radius: 16px;
          width: 420px; max-width: 90vw; max-height: 80vh;
          overflow-y: auto; box-shadow: 0 8px 32px rgba(0,0,0,0.2);
        }
        .yt_header {
          display: flex; justify-content: space-between; align-items: center;
          padding: 20px 24px 16px; border-bottom: 1px solid #e5e7eb;
        }
        .yt_header h3 { margin: 0; font-size: 18px; font-weight: 700; color: #1f2937; }
        .yt_close {
          background: none; border: none; font-size: 24px;
          cursor: pointer; color: #6b7280; line-height: 1;
        }
        .yt_body {
          padding: 20px 24px; display: flex; flex-direction: column; gap: 14px;
        }
        .yt_status_row {
          display: flex; justify-content: space-between; align-items: center;
        }
        .yt_status_label { font-size: 14px; color: #374151; font-weight: 500; }
        .yt_badge {
          font-size: 12px; font-weight: 600; padding: 4px 12px; border-radius: 12px;
        }
        .yt_badge.on { background: #d1fae5; color: #059669; }
        .yt_badge.off { background: #fee2e2; color: #dc2626; }
        .yt_connect_btn {
          padding: 10px 24px; font-size: 14px; font-weight: 600;
          background: #dc2626; color: #fff; border: none;
          border-radius: 8px; cursor: pointer;
        }
        .yt_connect_btn:hover { background: #b91c1c; }
        .yt_divider { height: 1px; background: #e5e7eb; }
        .yt_section_title { font-size: 13px; font-weight: 600; color: #374151; }
        .yt_action_btn {
          padding: 8px 20px; font-size: 13px; font-weight: 600;
          border: none; border-radius: 8px; cursor: pointer; align-self: flex-start;
        }
        .yt_action_btn:disabled { opacity: 0.6; cursor: default; }
        .yt_action_btn.green { background: #059669; color: #fff; }
        .yt_action_btn.green:hover:not(:disabled) { background: #047857; }
        .yt_action_btn.blue { background: #1f3f62; color: #fff; }
        .yt_action_btn.blue:hover:not(:disabled) { background: #2d5a8a; }
      `}</style>
    </div>
  );
}

"use client";

import { useState, useEffect } from "react";
import toast from "react-hot-toast";
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

  const handleConnect = async () => {
    // Desktop 모드: 시스템 브라우저로 열기 (Wails WebView window.open 우회)
    const res = await fetch("/api/youtube/open-auth").catch(() => null);
    if (!res || !res.ok) {
      window.open(apiClient.getYoutubeAuthUrl(), "yt_auth", "width=600,height=700");
    }
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
      if (res.ok) toast.success(res.message || "OBS 스트림 설정 완료!");
      else toast.error(`실패: ${res.error || "알 수 없는 오류"}`);
    } catch {
      toast.error("요청 실패");
    } finally {
      setSettingUp(false);
    }
  };

  const handleManualUpload = async () => {
    setUploading(true);
    try {
      const res = await apiClient.generateThumbnail("main_worship", undefined, true);
      if (res.ok) toast.success("썸네일 생성 + YouTube 업로드 요청 완료");
      else toast.error(`실패: ${res.error || "알 수 없는 오류"}`);
    } catch {
      toast.error("요청 실패");
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[11000] flex items-center justify-center bg-black/50" onClick={onClose}>
      <div
        className="bg-white rounded-2xl w-[420px] max-w-[90vw] max-h-[80vh] overflow-y-auto shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        {/* header */}
        <div className="flex justify-between items-center px-6 pt-5 pb-4 border-b border-[#e5e7eb]">
          <h3 className="m-0 text-lg font-bold text-[#1f2937]">YouTube</h3>
          <button
            className="bg-transparent border-none text-2xl cursor-pointer text-[#6b7280] leading-none"
            onClick={onClose}
          >
            &times;
          </button>
        </div>

        <FeatureGate feature="youtube_integration">
          <div className="px-6 py-5 flex flex-col gap-3.5">
            {/* 연결 상태 */}
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium text-[#374151]">연결 상태</span>
              <span
                className={`text-xs font-semibold px-3 py-1 rounded-xl ${
                  connected
                    ? "bg-[#d1fae5] text-[#059669]"
                    : "bg-[#fee2e2] text-[#dc2626]"
                }`}
              >
                {connected ? "연결됨" : "미연결"}
              </span>
            </div>

            {!connected && (
              <div className="text-center py-5">
                <p className="text-xs text-[#6b7280] mb-4">
                  YouTube 계정을 연결하면 예배 시작 시 썸네일이 자동으로 업로드됩니다.
                </p>
                <button
                  className="px-6 py-2.5 text-sm font-semibold bg-[#dc2626] text-white border-none rounded-lg cursor-pointer hover:bg-[#b91c1c] transition-colors"
                  onClick={handleConnect}
                >
                  YouTube 계정 연결
                </button>
              </div>
            )}

            {connected && (
              <>
                <div className="h-px bg-[#e5e7eb]" />
                <div className="text-xs font-semibold text-[#374151]">OBS 스트리밍 설정</div>
                <p className="text-xs text-[#6b7280] m-0">
                  YouTube 스트림 키를 자동으로 가져와 OBS에 세팅합니다.
                </p>
                <button
                  className="self-start px-5 py-2 text-xs font-semibold bg-[#059669] text-white border-none rounded-lg cursor-pointer hover:bg-[#047857] disabled:opacity-60 disabled:cursor-default transition-colors"
                  onClick={handleSetupOBS}
                  disabled={settingUp}
                >
                  {settingUp ? "설정 중..." : "OBS 스트림 자동 세팅"}
                </button>

                <div className="h-px bg-[#e5e7eb]" />
                <div className="text-xs font-semibold text-[#374151]">수동 업로드</div>
                <p className="text-xs text-[#6b7280] m-0">
                  현재 활성/예정 라이브 방송에 썸네일을 즉시 생성하고 업로드합니다.
                </p>
                <button
                  className="self-start px-5 py-2 text-xs font-semibold bg-[#1f3f62] text-white border-none rounded-lg cursor-pointer hover:bg-[#2d5a8a] disabled:opacity-60 disabled:cursor-default transition-colors"
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
    </div>
  );
}

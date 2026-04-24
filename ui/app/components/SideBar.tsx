"use client";

import { useState, useEffect } from "react";
import { userInfoState } from "@/recoilState";
import { useAuth } from "@/lib/LocalAuthContext";
import { useRecoilState } from "recoil";
import SettingsPanel from "./SettingsPanel";
import HistoryList from "./HistoryList";
import YouTubePanel from "./YouTubePanel";
import LicensePanel from "./LicensePanel";

const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL;

interface SidebarProps {
  open: boolean;
  onClose: () => void;
}

export default function Sidebar({ open, onClose }: SidebarProps) {
  const { church, updateChurch } = useAuth();
  const [userInfo, setUserInfo] = useRecoilState(userInfoState);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [historyType, setHistoryType] = useState<string | undefined>();
  const [youtubeOpen, setYoutubeOpen] = useState(false);
  const [licenseOpen, setLicenseOpen] = useState(false);
  // 교회 정보 편집
  const [editingChurch, setEditingChurch] = useState(false);
  const [churchName, setChurchName] = useState("");
  const [churchEngName, setChurchEngName] = useState("");
  const [saving, setSaving] = useState(false);

  // 사이드바 열릴 때 현재 값 동기화
  useEffect(() => {
    if (open) {
      setChurchName(userInfo?.name || "");
      setChurchEngName(userInfo?.english_name || "");
      setEditingChurch(false);
    }
  }, [open, userInfo]);

  const openHistory = (type?: string) => {
    setHistoryType(type);
    setHistoryOpen(true);
  };

  const handleSaveChurch = async () => {
    const email = church?.email || userInfo?.email || "local@localhost";
    setSaving(true);
    try {
      const res = await fetch(`${BASE_URL}/api/user`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          email,
          name: churchName.trim(),
          english_name: churchEngName.trim(),
        }),
      });
      if (res.ok) {
        const newName = churchName.trim();
        const newEngName = churchEngName.trim();
        setUserInfo((prev) => ({
          ...prev,
          name: newName,
          english_name: newEngName,
        }));
        updateChurch({ name: newName, englishName: newEngName });

        // Display 교회명도 즉시 업데이트
        fetch(`${BASE_URL}/display/church-name`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ churchName: newEngName }),
        }).catch(() => {});

        setEditingChurch(false);
      }
    } catch (e) {
      console.error("교회 정보 저장 에러:", e);
    } finally {
      setSaving(false);
    }
  };

  const menuActions: { title: string; action?: () => void }[] = [
    { title: "설정", action: () => setSettingsOpen(true) },
    { title: "YouTube", action: () => setYoutubeOpen(true) },
    { title: "라이선스 정보", action: () => setLicenseOpen(true) },
    { title: "생성 내역", action: () => openHistory() },
  ];

  return (
    <>
      {/* Background Overlay */}
      <div
        className={`fixed inset-0 bg-black/60 z-[10499] transition-all duration-300 ${
          open ? "opacity-100 visible" : "opacity-0 invisible"
        }`}
        onClick={onClose}
      />

      {/* Sidebar */}
      <div
        className={`fixed top-0 right-0 w-80 h-screen bg-[#1a1a1a] border-l border-[#3a3a3a] flex flex-col items-center p-5 shadow-[-4px_0_16px_rgba(0,0,0,0.5)] transition-all duration-300 ease-in-out z-[10500] ${
          open ? "translate-x-0" : "translate-x-full"
        }`}
      >
        {/* 닫기 버튼 */}
        <button
          className="self-end bg-transparent border-none text-xl text-[#e8e8e8] cursor-pointer leading-none"
          onClick={onClose}
        >
          ✕
        </button>

        {/* 아바타 */}
        <div className="w-20 h-20 rounded-full mt-2.5 bg-[#1a3a5f] flex items-center justify-center text-3xl text-[#e8e8e8] font-bold">
          {(church?.name || userInfo?.name || "EP").charAt(0)}
        </div>

        {/* 교회명 */}
        <div className="font-bold text-base text-[#4a9eff] mt-2.5">
          {church?.name || userInfo?.name || "교회명 미설정"}
        </div>
        {/* 메뉴 카드 */}
        <div className="bg-[#222222] border border-[#3a3a3a] p-4 rounded-xl w-full">
          {/* 교회 정보 편집 */}
          {editingChurch ? (
            <div className="py-2 pb-3 border-b border-[#3a3a3a]">
              <div className="mb-2.5">
                <div className="font-bold text-xs text-[#e8e8e8]">소속교회</div>
                <input
                  className="w-full mt-1 px-2.5 py-1.5 border border-[#4a4a4a] rounded-md bg-[#2a2a2a] text-[#e8e8e8] text-xs outline-none"
                  value={churchName}
                  onChange={(e) => setChurchName(e.target.value)}
                  placeholder="예: 사랑의교회"
                />
              </div>
              <div className="mb-2.5">
                <div className="font-bold text-xs text-[#e8e8e8]">교회표기 (영문)</div>
                <input
                  className="w-full mt-1 px-2.5 py-1.5 border border-[#4a4a4a] rounded-md bg-[#2a2a2a] text-[#e8e8e8] text-xs outline-none"
                  value={churchEngName}
                  onChange={(e) => setChurchEngName(e.target.value)}
                  placeholder="예: Sarang Church"
                />
              </div>
              <div className="flex gap-2 justify-end">
                <button
                  className="px-3.5 py-1 text-xs bg-[#333333] border-none rounded-md text-[#e8e8e8] cursor-pointer"
                  onClick={() => setEditingChurch(false)}
                >
                  취소
                </button>
                <button
                  className={`px-3.5 py-1 text-xs font-semibold bg-[#4a9eff] border-none rounded-md text-[#e8e8e8] ${
                    saving ? "opacity-60 cursor-default" : "cursor-pointer"
                  }`}
                  onClick={handleSaveChurch}
                  disabled={saving}
                >
                  {saving ? "저장 중..." : "저장"}
                </button>
              </div>
            </div>
          ) : (
            <div
              className="flex justify-between items-center gap-2 py-2.5 border-b border-[#3a3a3a] text-[#e8e8e8] cursor-pointer"
              onClick={() => setEditingChurch(true)}
            >
              <div className="min-w-0">
                <div className="font-bold text-[15px]">교회 정보</div>
                <div className="text-xs text-[#a0a0a0] mt-0.5 truncate">
                  {userInfo?.name || "미등록"}
                  {userInfo?.english_name ? ` (${userInfo.english_name})` : ""}
                </div>
              </div>
              <div className="text-xs opacity-60 text-[#ccc] flex-shrink-0 whitespace-nowrap">수정</div>
            </div>
          )}

          {/* 나머지 메뉴 항목들 */}
          {menuActions.map(({ title, action }) => (
            <div
              key={title}
              className={`flex justify-between items-center py-2.5 border-b border-[#3a3a3a] text-[#e8e8e8] ${
                action ? "cursor-pointer" : "cursor-default"
              }`}
              onClick={action}
            >
              <div className="font-bold text-[15px]">{title}</div>
              {action && (
                <div className="text-lg opacity-70">›</div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Settings Modal */}
      <SettingsPanel open={settingsOpen} onClose={() => setSettingsOpen(false)} />

      {/* History Modal */}
      <HistoryList
        open={historyOpen}
        onClose={() => setHistoryOpen(false)}
        filterType={historyType}
      />

      {/* YouTube Modal */}
      <YouTubePanel open={youtubeOpen} onClose={() => setYoutubeOpen(false)} />

      {/* License Modal */}
      <LicensePanel open={licenseOpen} onClose={() => setLicenseOpen(false)} />

    </>
  );
}

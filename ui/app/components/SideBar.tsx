"use client";

import { useState, useEffect } from "react";
import { userInfoState } from "@/recoilState";
import { signOut } from "next-auth/react";
import { useRecoilState } from "recoil";
import SettingsPanel from "./SettingsPanel";
import HistoryList from "./HistoryList";

const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL;

interface SidebarProps {
  open: boolean;
  onClose: () => void;
  user: {
    name?: string | null;
    email?: string | null;
    image?: string | null;
  };
}

export default function Sidebar({ open, onClose, user }: SidebarProps) {
  const [userInfo, setUserInfo] = useRecoilState(userInfoState);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [historyType, setHistoryType] = useState<string | undefined>();

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
    if (!user.email) return;
    setSaving(true);
    try {
      const res = await fetch(`${BASE_URL}/api/user`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          email: user.email,
          name: churchName.trim(),
          english_name: churchEngName.trim(),
        }),
      });
      if (res.ok) {
        setUserInfo((prev) => ({
          ...prev,
          name: churchName.trim(),
          english_name: churchEngName.trim(),
        }));
        setEditingChurch(false);
      }
    } catch (e) {
      console.error("교회 정보 저장 에러:", e);
    } finally {
      setSaving(false);
    }
  };

  const inputStyle: React.CSSProperties = {
    width: "100%",
    padding: "6px 10px",
    border: "1px solid rgba(255,255,255,0.4)",
    borderRadius: "6px",
    background: "rgba(255,255,255,0.15)",
    color: "#fff",
    fontSize: "13px",
    outline: "none",
    marginTop: "4px",
  };

  const menuActions: { title: string; action?: () => void }[] = [
    { title: "설정", action: () => setSettingsOpen(true) },
    { title: "라이선스 정보" },
    { title: "주보 생성 내역", action: () => openHistory("bulletin") },
    { title: "PPT 생성 내역", action: () => openHistory("ppt") },
    { title: "가사 PPT 생성 내역", action: () => openHistory("lyrics_ppt") },
  ];

  return (
    <>
      {/* Background Overlay */}
      <div
        style={{
          position: "fixed",
          top: 0,
          left: 0,
          width: "100%",
          height: "100vh",
          backgroundColor: "rgba(0, 0, 0, 0.6)",
          opacity: open ? 1 : 0,
          visibility: open ? "visible" : "hidden",
          transition: "opacity 0.3s ease, visibility 0.3s ease",
          zIndex: 998,
        }}
        onClick={onClose}
      />

      {/* Sidebar */}
      <div
        style={{
          position: "fixed",
          top: 0,
          right: open ? 0 : "-320px",
          width: "320px",
          height: "100vh",
          backgroundColor: "#8a8a8a",
          padding: "20px",
          boxShadow: "-2px 0 6px rgba(0,0,0,0.3)",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          transition: "right 0.3s ease-in-out",
          zIndex: 999,
        }}
      >
        <button
          style={{
            alignSelf: "flex-end",
            background: "none",
            border: "none",
            fontSize: "20px",
            color: "#fff",
            cursor: "pointer",
          }}
          onClick={onClose}
        >
          ✕
        </button>

        <img
          src={user.image || "/default-profile.png"}
          alt="profile"
          style={{
            width: "80px",
            height: "80px",
            borderRadius: "50%",
            marginTop: "10px",
          }}
        />
        <div
          style={{
            fontWeight: "bold",
            fontSize: "16px",
            color: "#204d87",
            marginTop: "10px",
          }}
        >
          {user.name}
        </div>
        <div style={{ fontSize: "12px", color: "#ddd", marginBottom: "20px" }}>
          {user.email}
        </div>

        <div
          style={{
            backgroundColor: "rgba(255, 255, 255, 0.1)",
            padding: "20px",
            borderRadius: "12px",
            width: "100%",
          }}
        >
          {/* 소속교회 / 교회표기 — 편집 가능 */}
          {editingChurch ? (
            <div style={{ padding: "8px 0 12px", borderBottom: "1px solid rgba(255,255,255,0.2)" }}>
              <div style={{ marginBottom: "10px" }}>
                <div style={{ fontWeight: "bold", fontSize: "13px", color: "#fff" }}>소속교회</div>
                <input
                  style={inputStyle}
                  value={churchName}
                  onChange={(e) => setChurchName(e.target.value)}
                  placeholder="예: 사랑의교회"
                />
              </div>
              <div style={{ marginBottom: "10px" }}>
                <div style={{ fontWeight: "bold", fontSize: "13px", color: "#fff" }}>교회표기 (영문)</div>
                <input
                  style={inputStyle}
                  value={churchEngName}
                  onChange={(e) => setChurchEngName(e.target.value)}
                  placeholder="예: Sarang Church"
                />
              </div>
              <div style={{ display: "flex", gap: "8px", justifyContent: "flex-end" }}>
                <button
                  style={{
                    padding: "5px 14px",
                    fontSize: "12px",
                    background: "rgba(255,255,255,0.2)",
                    border: "none",
                    borderRadius: "6px",
                    color: "#fff",
                    cursor: "pointer",
                  }}
                  onClick={() => setEditingChurch(false)}
                >
                  취소
                </button>
                <button
                  style={{
                    padding: "5px 14px",
                    fontSize: "12px",
                    fontWeight: 600,
                    background: "#204d87",
                    border: "none",
                    borderRadius: "6px",
                    color: "#fff",
                    cursor: saving ? "default" : "pointer",
                    opacity: saving ? 0.6 : 1,
                  }}
                  onClick={handleSaveChurch}
                  disabled={saving}
                >
                  {saving ? "저장 중..." : "저장"}
                </button>
              </div>
            </div>
          ) : (
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                padding: "10px 0",
                borderBottom: "1px solid rgba(255,255,255,0.2)",
                color: "#fff",
                cursor: "pointer",
              }}
              onClick={() => setEditingChurch(true)}
            >
              <div>
                <div style={{ fontWeight: "bold", fontSize: "15px" }}>교회 정보</div>
                <div style={{ fontSize: "13px", opacity: 0.8, marginTop: "2px" }}>
                  {userInfo?.name || "미등록"}
                  {userInfo?.english_name ? ` (${userInfo.english_name})` : ""}
                </div>
              </div>
              <div style={{ fontSize: "12px", opacity: 0.6, color: "#ccc" }}>수정</div>
            </div>
          )}

          {/* 나머지 메뉴 항목들 */}
          {menuActions.map(({ title, action }) => (
            <div
              key={title}
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                padding: "10px 0",
                borderBottom: "1px solid rgba(255,255,255,0.2)",
                color: "#fff",
                cursor: action ? "pointer" : "default",
              }}
              onClick={action}
            >
              <div style={{ fontWeight: "bold", fontSize: "15px" }}>{title}</div>
              {action && (
                <div style={{ fontSize: "18px", opacity: 0.7 }}>›</div>
              )}
            </div>
          ))}
        </div>

        <div
          style={{
            marginTop: "auto",
            color: "#b65050",
            fontWeight: "bold",
            cursor: "pointer",
          }}
          onClick={() => {
            onClose();
            signOut();
          }}
        >
          logout
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
    </>
  );
}

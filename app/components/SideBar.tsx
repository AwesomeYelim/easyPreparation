"use client";

import { signOut } from "next-auth/react";

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
          {[
            ["소속교회", "동남생명의 빛 교회"],
            ["교회표기", "Light of Light Church"],
            ["라이선스 정보"],
            ["주보 생성 내역"],
            ["PPT 생성 내역"],
            ["가사 PPT 생성 내역"],
          ].map(([title, value]) => (
            <div
              key={title}
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                padding: "10px 0",
                borderBottom: "1px solid rgba(255,255,255,0.2)",
                color: "#fff",
              }}
            >
              <div>
                <div style={{ fontWeight: "bold", fontSize: "15px" }}>
                  {title}
                </div>
                {value && (
                  <div style={{ fontSize: "13px", opacity: 0.8 }}>{value}</div>
                )}
              </div>
              <div style={{ fontSize: "18px", opacity: 0.7 }}>›</div>
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
    </>
  );
}

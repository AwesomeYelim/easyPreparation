"use client";

import { useState } from "react";
import { useAuth } from "@/lib/LocalAuthContext";

export default function SetupWizard() {
  const { needsSetup, isLoading, completeSetup, setupError } = useAuth();
  const [name, setName] = useState("");
  const [englishName, setEnglishName] = useState("");
  const [saving, setSaving] = useState(false);

  if (isLoading || !needsSetup) return null;

  const handleSubmit = async () => {
    if (!name.trim() || !englishName.trim()) return;
    setSaving(true);
    try {
      await completeSetup(name.trim(), englishName.trim());
    } catch (e) {
      console.error("Setup failed:", e);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div
      style={{
        position: "fixed",
        top: 0,
        left: 0,
        width: "100vw",
        height: "100vh",
        backgroundColor: "rgba(0, 0, 0, 0.4)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        zIndex: 20000,
      }}
    >
      <div
        style={{
          maxWidth: "420px",
          width: "90%",
          padding: "28px",
          border: "1px solid #ccc",
          borderRadius: "12px",
          backgroundColor: "#fff",
          display: "flex",
          flexDirection: "column",
          gap: "14px",
          boxShadow: "0 8px 24px rgba(0,0,0,0.2)",
        }}
      >
        <h2 style={{ textAlign: "center", marginBottom: "4px", fontSize: "20px" }}>
          교회 정보 설정
        </h2>
        <p style={{ textAlign: "center", fontSize: "14px", color: "#666", margin: 0 }}>
          처음 사용하시나요? 교회 정보를 입력해주세요.
        </p>
        <input
          placeholder="교회 이름 (한글)"
          value={name}
          onChange={(e) => setName(e.target.value)}
          style={{
            padding: "10px 12px",
            borderRadius: "6px",
            border: "1px solid #ccc",
            fontSize: "14px",
          }}
        />
        <input
          placeholder="영문 이름 (예: Sarang Church)"
          value={englishName}
          onChange={(e) => setEnglishName(e.target.value)}
          style={{
            padding: "10px 12px",
            borderRadius: "6px",
            border: "1px solid #ccc",
            fontSize: "14px",
          }}
        />
        <button
          onClick={handleSubmit}
          disabled={saving || !name.trim() || !englishName.trim()}
          style={{
            padding: "12px",
            backgroundColor: saving ? "#999" : "#0070f3",
            color: "#fff",
            border: "none",
            borderRadius: "6px",
            fontWeight: "bold",
            cursor: saving ? "default" : "pointer",
            fontSize: "14px",
          }}
        >
          {saving ? "저장 중..." : "시작하기"}
        </button>
        {setupError && (
          <p style={{ color: "#e00", fontSize: "13px", margin: 0, textAlign: "center" }}>
            오류: {setupError}
          </p>
        )}
      </div>
    </div>
  );
}

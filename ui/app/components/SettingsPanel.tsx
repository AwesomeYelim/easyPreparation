"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { useRecoilState, useRecoilValue } from "recoil";
import { userSettingsState, userInfoState } from "@/recoilState";
import { apiClient } from "@/lib/apiClient";
import { UserSettings, ThumbnailConfig } from "@/types";
import SchedulePanel from "./SchedulePanel";
import { useFeature } from "./FeatureGate";

function useCurrentVersion() {
  const [version, setVersion] = useState<string | null>(null);
  useEffect(() => {
    apiClient.getVersion().then((v) => setVersion(v.version)).catch(() => {});
  }, []);
  return version;
}

interface SettingsPanelProps {
  open: boolean;
  onClose: () => void;
}

type Tab = "general" | "special";

const WORSHIP_LABELS: Record<string, string> = {
  main_worship: "주일예배",
  after_worship: "오후예배",
  wed_worship: "수요예배",
  fri_worship: "금요예배",
};

export default function SettingsPanel({ open, onClose }: SettingsPanelProps) {
  const userInfo = useRecoilValue(userInfoState);
  const [settings, setSettings] = useRecoilState(userSettingsState);
  const [local, setLocal] = useState<UserSettings>(settings);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [scheduleOpen, setScheduleOpen] = useState(false);
  const [tab, setTab] = useState<Tab>("general");
  const currentVersion = useCurrentVersion();

  // 기념주일 탭 데이터
  const [thumbConfig, setThumbConfig] = useState<ThumbnailConfig | null>(null);

  useEffect(() => {
    setLocal(settings);
    setError("");
    if (open) {
      apiClient.getThumbnailConfig().then(setThumbConfig).catch(console.error);
    }
  }, [settings, open]);

  const handleSave = async () => {
    if (!userInfo.email) {
      setError("로그인 정보가 없습니다. 다시 로그인해주세요.");
      return;
    }
    setSaving(true);
    setError("");
    try {
      const res = await apiClient.saveSettings(userInfo.email, local);
      if (res.error) {
        setError(`설정 저장 실패: ${res.error}`);
        return;
      }
      setSettings(local);
      document.documentElement.setAttribute("data-theme", local.theme || "light");
      if (local.font_size) {
        document.documentElement.style.setProperty("--user-font-size", `${local.font_size}px`);
      }
      if (thumbConfig) {
        try {
          await apiClient.saveThumbnailConfig(thumbConfig);
        } catch (te) {
          console.error("썸네일 설정 저장 에러:", te);
        }
      }
      onClose();
    } catch (e: any) {
      console.error("설정 저장 에러:", e);
      setError(`설정 저장에 실패했습니다: ${e?.message || e}`);
    } finally {
      setSaving(false);
    }
  };

  if (!open) return null;

  return (
    <div className="settings_overlay" onClick={onClose}>
      <div className="settings_panel" onClick={(e) => e.stopPropagation()}>
        <div className="settings_header">
          <div style={{ display: "flex", flexDirection: "column", gap: 2 }}>
            <h3>설정</h3>
            {currentVersion && (
              <span style={{
                fontSize: 11,
                color: "var(--text-muted)",
                fontFamily: "monospace",
                letterSpacing: "0.3px",
              }}>
                v{currentVersion}
              </span>
            )}
          </div>
          <button className="settings_close" onClick={onClose}>&times;</button>
        </div>

        <div className="settings_tabs">
          <button
            className={`settings_tab ${tab === "general" ? "active" : ""}`}
            onClick={() => setTab("general")}
          >
            일반
          </button>
          <button
            className={`settings_tab ${tab === "special" ? "active" : ""}`}
            onClick={() => setTab("special")}
          >
            기념주일
          </button>
        </div>

        <div className="settings_body">
          {tab === "general" && (
            <GeneralTab
              local={local}
              setLocal={setLocal}
              onScheduleOpen={() => setScheduleOpen(true)}
            />
          )}
          {tab === "special" && thumbConfig && (
            <SpecialTab thumbConfig={thumbConfig} setThumbConfig={setThumbConfig} />
          )}
        </div>

        {error && <div className="settings_error">{error}</div>}

        <div className="settings_footer">
          <button className="settings_cancel_btn" onClick={onClose}>취소</button>
          <button className="settings_save_btn" onClick={handleSave} disabled={saving}>
            {saving ? "저장 중..." : "저장"}
          </button>
        </div>
      </div>

      <SchedulePanel open={scheduleOpen} onClose={() => setScheduleOpen(false)} />

      <style jsx>{`
        .settings_overlay {
          position: fixed; top: 0; left: 0; width: 100%; height: 100%;
          background: rgba(0,0,0,0.5);
          display: flex; align-items: center; justify-content: center;
          z-index: 11000;
        }
        .settings_panel {
          background: var(--surface-elevated); border-radius: 16px;
          width: 480px; max-width: 90vw; max-height: 85vh;
          overflow-y: auto; box-shadow: 0 8px 32px rgba(0,0,0,0.2);
        }
        .settings_header {
          display: flex; justify-content: space-between; align-items: center;
          padding: 20px 24px 12px;
        }
        .settings_header h3 { margin: 0; font-size: 18px; font-weight: 700; color: var(--text-primary); }
        .settings_close {
          background: none; border: none; font-size: 24px;
          cursor: pointer; color: var(--text-secondary); line-height: 1;
        }
        .settings_tabs {
          display: flex; gap: 0; padding: 0 24px; border-bottom: 1px solid var(--border);
        }
        .settings_tab {
          padding: 10px 16px; font-size: 13px; font-weight: 500;
          color: var(--text-secondary); background: none; border: none;
          border-bottom: 2px solid transparent; cursor: pointer;
          transition: all 0.15s;
        }
        .settings_tab.active { color: var(--accent); border-bottom-color: var(--accent); font-weight: 600; }
        .settings_tab:hover { color: var(--text-primary); }
        .settings_body {
          padding: 20px 24px; display: flex; flex-direction: column; gap: 16px;
          min-height: 200px;
        }
        .settings_footer {
          display: flex; justify-content: flex-end; gap: 8px;
          padding: 16px 24px 20px; border-top: 1px solid var(--border);
        }
        .settings_cancel_btn {
          padding: 8px 20px; font-size: 13px; background: var(--surface-hover);
          border: 1px solid var(--border-input); border-radius: 8px; cursor: pointer;
        }
        .settings_save_btn {
          padding: 8px 20px; font-size: 13px; font-weight: 600;
          background: var(--accent); color: var(--surface-elevated); border: none;
          border-radius: 8px; cursor: pointer;
        }
        .settings_save_btn:hover { background: var(--accent-hover); }
        .settings_save_btn:disabled { background: var(--text-muted); }
        .settings_error {
          margin: 0 24px; padding: 8px 12px;
          background: var(--error-bg); color: var(--error); border-radius: 8px;
          font-size: 13px; border: 1px solid var(--error-border);
        }
      `}</style>
    </div>
  );
}

/* ── 일반 탭 ── */
function GeneralTab({
  local,
  setLocal,
  onScheduleOpen,
}: {
  local: UserSettings;
  setLocal: (s: UserSettings) => void;
  onScheduleOpen: () => void;
}) {
  const hasScheduler = useFeature("auto_scheduler");

  return (
    <>
      <label className="s_row">
        <span>선호 성경 번역</span>
        <select
          value={local.preferred_bible_version}
          onChange={(e) => setLocal({ ...local, preferred_bible_version: Number(e.target.value) })}
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
      <label className="s_row">
        <span>테마</span>
        <select value={local.theme} onChange={(e) => setLocal({ ...local, theme: e.target.value })}>
          <option value="light">라이트</option>
          <option value="dark">다크</option>
        </select>
      </label>
      <label className="s_row">
        <span>본문 폰트 크기</span>
        <input type="number" min={12} max={24} value={local.font_size}
          onChange={(e) => setLocal({ ...local, font_size: Number(e.target.value) })} />
      </label>
      <label className="s_row">
        <span>기본 BPM</span>
        <input type="number" min={40} max={200} value={local.default_bpm}
          onChange={(e) => setLocal({ ...local, default_bpm: Number(e.target.value) })} />
      </label>
      <label className="s_row">
        <span>Display 레이아웃</span>
        <select value={local.display_layout}
          onChange={(e) => setLocal({ ...local, display_layout: e.target.value })}>
          <option value="default">기본</option>
          <option value="wide">와이드</option>
          <option value="compact">컴팩트</option>
        </select>
      </label>
      <div style={{ height: 1, background: "var(--border)", margin: "4px 0" }} />
      <div className="s_row">
        <span style={{ display: "flex", alignItems: "center", gap: 6 }}>
          스트리밍 스케줄
          {!hasScheduler && (
            <span style={{
              fontSize: 10, fontWeight: 700, padding: "1px 6px",
              borderRadius: 8, background: "#7c3aed", color: "#fff",
              letterSpacing: "0.3px",
            }}>Pro</span>
          )}
        </span>
        <button
          className="s_action_btn"
          onClick={onScheduleOpen}
          disabled={!hasScheduler}
          style={!hasScheduler ? { opacity: 0.45, cursor: "not-allowed" } : undefined}
        >
          설정
        </button>
      </div>
      <style jsx>{`
        .s_row {
          display: flex; justify-content: space-between; align-items: center; gap: 12px;
        }
        .s_row > span { font-size: 14px; color: var(--text-primary); font-weight: 500; }
        .s_row select, .s_row input {
          padding: 6px 12px; border: 1px solid var(--border-input); border-radius: 8px;
          font-size: 13px; background: var(--surface-input); min-width: 140px;
        }
        .s_row input[type="number"] { width: 80px; min-width: 80px; }
        .s_action_btn {
          padding: 6px 16px; font-size: 13px; font-weight: 500;
          background: var(--surface-hover); border: 1px solid var(--border-input); border-radius: 8px;
          cursor: pointer; color: var(--text-primary);
        }
        .s_action_btn:hover:not(:disabled) { background: var(--border); }
      `}</style>
    </>
  );
}

/* ── 이미지 드롭존 ── */
function ImageDropZone({
  imageUrl, loading, onFile, onClear, height = 100, placeholder,
}: {
  imageUrl?: string; loading?: boolean; onFile: (f: File) => void;
  onClear?: () => void; height?: number; placeholder?: string;
}) {
  const [dragging, setDragging] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  return (
    <div
      className={`dz ${dragging ? "drag" : ""} ${imageUrl ? "has" : ""}`}
      onDragOver={(e) => { e.preventDefault(); e.stopPropagation(); if (!dragging) setDragging(true); }}
      onDragLeave={(e) => { e.stopPropagation(); setDragging(false); }}
      onDrop={(e) => {
        e.preventDefault(); e.stopPropagation(); setDragging(false);
        const file = e.dataTransfer.files[0];
        if (file && file.type.startsWith("image/")) onFile(file);
      }}
      onClick={(e) => { e.stopPropagation(); inputRef.current?.click(); }}
      style={{ height }}
    >
      {loading ? (
        <span className="dz_text">업로드 중...</span>
      ) : imageUrl ? (
        <>
          <img src={imageUrl} alt="" className="dz_img" />
          {onClear && (
            <button className="dz_clear"
              onClick={(e) => { e.stopPropagation(); onClear(); }}
              title="기본 배경으로"
            >&times;</button>
          )}
        </>
      ) : (
        <span className="dz_text">{placeholder || "이미지를 드래그하거나 클릭"}</span>
      )}
      <input ref={inputRef} type="file" accept="image/png,image/jpeg"
        style={{ display: "none" }}
        onChange={(e) => { const f = e.target.files?.[0]; if (f) onFile(f); e.target.value = ""; }}
      />
      <style jsx>{`
        .dz {
          position: relative; border: 2px dashed var(--border-input); border-radius: 8px;
          cursor: pointer; display: flex; align-items: center; justify-content: center;
          overflow: hidden; transition: border-color 0.15s, background 0.15s; background: var(--surface-input);
        }
        .dz:hover { border-color: var(--text-muted); }
        .dz.drag { border-color: #3b82f6; background: #eff6ff; }
        .dz.has { border-style: solid; border-color: var(--border); }
        .dz_text { font-size: 12px; color: var(--text-muted); pointer-events: none; text-align: center; padding: 0 12px; }
        .dz_img { width: 100%; height: 100%; object-fit: cover; }
        .dz_clear {
          position: absolute; top: 4px; right: 4px; width: 22px; height: 22px;
          border-radius: 50%; background: rgba(0,0,0,0.55); color: #fff;
          border: none; font-size: 14px; line-height: 1; cursor: pointer;
          display: flex; align-items: center; justify-content: center;
        }
        .dz_clear:hover { background: rgba(239,68,68,0.8); }
      `}</style>
    </div>
  );
}

/* ── 기념주일 탭 ── */
function SpecialTab({
  thumbConfig,
  setThumbConfig,
}: {
  thumbConfig: ThumbnailConfig;
  setThumbConfig: React.Dispatch<React.SetStateAction<ThumbnailConfig | null>>;
}) {
  const [newDate, setNewDate] = useState("");
  const [newLabel, setNewLabel] = useState("");
  const [newBgFile, setNewBgFile] = useState<File | null>(null);
  const [newBgPreview, setNewBgPreview] = useState<string | null>(null);
  const [uploading, setUploading] = useState<string | null>(null);
  const [previewType, setPreviewType] = useState("main_worship");
  const [previewDate, setPreviewDate] = useState(new Date().toISOString().slice(0, 10));
  const [previewKey, setPreviewKey] = useState(0);
  const [generating, setGenerating] = useState(false);

  const previewSpecial = thumbConfig.specials.find((s) => s.date === previewDate);

  const handleNewBgFile = useCallback((file: File) => {
    setNewBgFile(file);
    setNewBgPreview(URL.createObjectURL(file));
  }, []);

  const addSpecial = async () => {
    if (!newDate || !newLabel) return;
    let bgPath = "";
    if (newBgFile) {
      setUploading("new");
      try {
        const slug = newLabel.replace(/\s+/g, "_").toLowerCase();
        const res = await apiClient.uploadThumbnailBg(newBgFile, slug);
        if (res.ok) bgPath = res.path;
      } catch (e) { console.error("배경 업로드 실패:", e); }
      finally { setUploading(null); }
    }
    const finalBgPath = bgPath, finalDate = newDate, finalLabel = newLabel;
    setThumbConfig(prev => prev ? {
      ...prev,
      specials: [...prev.specials, { date: finalDate, label: finalLabel, background: finalBgPath, titleOverride: finalLabel }],
    } : prev);
    setNewDate(""); setNewLabel(""); setNewBgFile(null);
    if (newBgPreview) { URL.revokeObjectURL(newBgPreview); setNewBgPreview(null); }
  };

  const handleSpecialBgUpload = async (idx: number, file: File) => {
    setUploading(`special_${idx}`);
    try {
      const label = thumbConfig.specials[idx]?.label || "special";
      const res = await apiClient.uploadThumbnailBg(file, label.replace(/\s+/g, "_").toLowerCase());
      if (res.ok) {
        setThumbConfig(prev => {
          if (!prev) return prev;
          const updated = [...prev.specials];
          updated[idx] = { ...updated[idx], background: res.path };
          return { ...prev, specials: updated };
        });
      }
    } catch (e) { console.error("배경 업로드 실패:", e); }
    finally { setUploading(null); }
  };

  const handleDefaultBgUpload = async (worshipType: string, file: File) => {
    setUploading(`default_${worshipType}`);
    try {
      const res = await apiClient.uploadThumbnailBg(file, `default_${worshipType}`);
      if (res.ok) {
        setThumbConfig(prev => prev ? {
          ...prev,
          defaults: { ...prev.defaults, [worshipType]: { ...prev.defaults[worshipType], background: res.path } },
        } : prev);
      }
    } catch (e) { console.error("기본 배경 업로드 실패:", e); }
    finally { setUploading(null); }
  };

  const removeSpecial = (idx: number) => {
    setThumbConfig(prev => prev ? { ...prev, specials: prev.specials.filter((_, i) => i !== idx) } : prev);
  };

  const handlePreview = async () => {
    setGenerating(true);
    try { await apiClient.generateThumbnail(previewType, previewDate); setPreviewKey(k => k + 1); }
    catch (e) { console.error(e); }
    finally { setGenerating(false); }
  };

  return (
    <>
      <div style={{ fontSize: 13, fontWeight: 600, color: "var(--text-primary)" }}>기본 배경 이미지</div>
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 8 }}>
        {Object.entries(WORSHIP_LABELS).map(([type, label]) => {
          const theme = thumbConfig.defaults[type];
          const hasBg = !!theme?.background;
          return (
            <div key={type} style={{ display: "flex", flexDirection: "column", gap: 4 }}>
              <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
                <span style={{ fontSize: 12, color: "var(--text-primary)", fontWeight: 500 }}>{label}</span>
                <span style={{
                  fontSize: 10, fontWeight: 600, padding: "1px 6px", borderRadius: 8,
                  background: hasBg ? "#d1fae5" : "#fef3c7", color: hasBg ? "#059669" : "#d97706",
                }}>{hasBg ? "설정됨" : "미설정"}</span>
              </div>
              <ImageDropZone
                imageUrl={hasBg ? apiClient.getThumbnailImageUrl(theme.background) : undefined}
                loading={uploading === `default_${type}`}
                onFile={(f) => handleDefaultBgUpload(type, f)}
                height={60} placeholder="드래그 또는 클릭"
              />
            </div>
          );
        })}
      </div>

      <div style={{ height: 1, background: "var(--border)", margin: "4px 0" }} />

      <div style={{ fontSize: 13, fontWeight: 600, color: "var(--text-primary)" }}>기념 주일</div>
      {thumbConfig.specials.length === 0 && (
        <div style={{ fontSize: 13, color: "var(--text-muted)", padding: "4px 0" }}>등록된 기념 주일이 없습니다</div>
      )}
      {thumbConfig.specials.map((s, i) => (
        <div key={i} style={{
          display: "flex", gap: 10, padding: "10px 12px",
          background: "var(--surface-input)", borderRadius: 8, alignItems: "stretch",
        }}>
          <div style={{ width: 140, flexShrink: 0 }}>
            <ImageDropZone
              imageUrl={s.background ? apiClient.getThumbnailImageUrl(s.background) : undefined}
              loading={uploading === `special_${i}`}
              onFile={(f) => handleSpecialBgUpload(i, f)}
              onClear={s.background ? () => {
                setThumbConfig(prev => {
                  if (!prev) return prev;
                  const updated = [...prev.specials];
                  updated[i] = { ...updated[i], background: "" };
                  return { ...prev, specials: updated };
                });
              } : undefined}
              height={80} placeholder="배경 드래그"
            />
          </div>
          <div style={{ flex: 1, display: "flex", flexDirection: "column", justifyContent: "center", gap: 4 }}>
            <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
              <span style={{ fontSize: 14, fontWeight: 500, color: "var(--text-primary)" }}>{s.label}</span>
              <span style={{
                fontSize: 10, fontWeight: 600, padding: "1px 7px", borderRadius: 8,
                background: s.background ? "#dbeafe" : "var(--surface-hover)", color: s.background ? "#2563eb" : "var(--text-secondary)",
              }}>{s.background ? "커스텀" : "기본 배경"}</span>
            </div>
            <span style={{ fontSize: 12, color: "var(--text-secondary)" }}>{s.date}</span>
          </div>
          <button onClick={() => removeSpecial(i)} style={{
            background: "none", border: "none", fontSize: 18, color: "var(--text-muted)",
            cursor: "pointer", alignSelf: "flex-start", lineHeight: 1, padding: 0,
          }}
            onMouseEnter={(e) => (e.currentTarget.style.color = "#ef4444")}
            onMouseLeave={(e) => (e.currentTarget.style.color = "var(--text-muted)")}
          >&times;</button>
        </div>
      ))}

      <div style={{ height: 1, background: "var(--border)", margin: "2px 0" }} />
      <div style={{ fontSize: 13, fontWeight: 600, color: "var(--text-primary)" }}>추가</div>
      <div style={{ display: "flex", gap: 10, alignItems: "stretch" }}>
        <div style={{ width: 140, flexShrink: 0 }}>
          <ImageDropZone
            imageUrl={newBgPreview || undefined}
            loading={uploading === "new"}
            onFile={handleNewBgFile}
            onClear={newBgPreview ? () => {
              setNewBgFile(null); URL.revokeObjectURL(newBgPreview); setNewBgPreview(null);
            } : undefined}
            height={80} placeholder="배경 드래그 (선택)"
          />
        </div>
        <div style={{ flex: 1, display: "flex", flexDirection: "column", gap: 6, justifyContent: "center" }}>
          <input type="date" value={newDate} onChange={(e) => setNewDate(e.target.value)}
            style={{ padding: "5px 8px", border: "1px solid var(--border-input)", borderRadius: 6, fontSize: 13, background: "var(--surface-elevated)" }}
          />
          <div style={{ display: "flex", gap: 6 }}>
            <input type="text" placeholder="이름 (예: 부활절 예배)" value={newLabel}
              onChange={(e) => setNewLabel(e.target.value)}
              style={{ flex: 1, padding: "5px 8px", border: "1px solid var(--border-input)", borderRadius: 6, fontSize: 13, background: "var(--surface-elevated)" }}
            />
            <button onClick={addSpecial} disabled={!newDate || !newLabel || uploading === "new"}
              style={{
                padding: "5px 14px", fontSize: 13, fontWeight: 600,
                background: "var(--accent)", color: "var(--surface-elevated)", border: "none",
                borderRadius: 6, cursor: "pointer", whiteSpace: "nowrap",
                opacity: !newDate || !newLabel ? 0.5 : 1,
              }}
            >{uploading === "new" ? "..." : "추가"}</button>
          </div>
        </div>
      </div>

      <div style={{ height: 1, background: "var(--border)", margin: "4px 0" }} />
      <div style={{ fontSize: 13, fontWeight: 600, color: "var(--text-primary)" }}>썸네일 미리보기</div>
      <div style={{ display: "flex", gap: 8, alignItems: "center", flexWrap: "wrap" }}>
        <select value={previewType} onChange={(e) => setPreviewType(e.target.value)}
          style={{ padding: "6px 10px", border: "1px solid var(--border-input)", borderRadius: 6, fontSize: 13, background: "var(--surface-input)" }}>
          {Object.entries(WORSHIP_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
        </select>
        <input type="date" value={previewDate} onChange={(e) => setPreviewDate(e.target.value)}
          style={{ padding: "5px 8px", border: "1px solid var(--border-input)", borderRadius: 6, fontSize: 13, background: "var(--surface-input)" }} />
        <button onClick={handlePreview} disabled={generating} style={{
          padding: "6px 14px", fontSize: 13, fontWeight: 600,
          background: "var(--success)", color: "#fff", border: "none", borderRadius: 6, cursor: "pointer",
        }}>{generating ? "생성 중..." : "생성"}</button>
      </div>
      {previewSpecial && (
        <div style={{ fontSize: 12, color: "#2563eb", background: "#eff6ff", padding: "6px 10px", borderRadius: 6, fontWeight: 500 }}>
          기념주일 적용: {previewSpecial.label}{previewSpecial.background ? " (커스텀 배경)" : " (기본 배경)"}
        </div>
      )}
      {previewKey > 0 && (
        <img key={previewKey} src={apiClient.getThumbnailPreviewUrl(previewType, previewDate)}
          alt="썸네일 미리보기" style={{ width: "100%", borderRadius: 8, border: "1px solid var(--border)" }} />
      )}
    </>
  );
}

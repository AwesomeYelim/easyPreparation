"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { apiClient } from "@/lib/apiClient";
import {
  ScheduleConfig,
  ScheduleEntry,
  ThumbnailConfig,
  SpecialDate,
  YouTubeStatus,
} from "@/types";

const WEEKDAY_NAMES = ["일", "월", "화", "수", "목", "금", "토"];
type Tab = "schedule" | "special" | "youtube";

interface SchedulePanelProps {
  open: boolean;
  onClose: () => void;
}

export default function SchedulePanel({ open, onClose }: SchedulePanelProps) {
  const [tab, setTab] = useState<Tab>("schedule");
  const [config, setConfig] = useState<ScheduleConfig | null>(null);
  const [thumbConfig, setThumbConfig] = useState<ThumbnailConfig | null>(null);
  const [ytStatus, setYtStatus] = useState<YouTubeStatus | null>(null);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState<string | null>(null);

  // 스케줄 테스트
  const handleTest = useCallback(
    async (action: "countdown" | "trigger", worshipType: string) => {
      setTesting(`${action}_${worshipType}`);
      try {
        const res = await apiClient.scheduleTest(action, worshipType);
        if (res.ok) alert(res.message);
        else alert("테스트 실패");
      } catch {
        alert("테스트 요청 실패");
      } finally {
        setTesting(null);
      }
    },
    []
  );

  // 데이터 로드
  useEffect(() => {
    if (!open) return;
    apiClient.getSchedule().then(setConfig).catch(console.error);
    apiClient.getThumbnailConfig().then(setThumbConfig).catch(console.error);
    apiClient.getYoutubeStatus().then(setYtStatus).catch(console.error);
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
      if (thumbConfig) await apiClient.saveThumbnailConfig(thumbConfig);
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
        {/* 헤더 */}
        <div className="sched_header">
          <h3>스케줄 설정</h3>
          <button className="sched_close" onClick={onClose}>
            &times;
          </button>
        </div>

        {/* 탭 */}
        <div className="sched_tabs">
          <button
            className={`sched_tab ${tab === "schedule" ? "active" : ""}`}
            onClick={() => setTab("schedule")}
          >
            정기 스케줄
          </button>
          <button
            className={`sched_tab ${tab === "special" ? "active" : ""}`}
            onClick={() => setTab("special")}
          >
            기념주일
          </button>
          <button
            className={`sched_tab ${tab === "youtube" ? "active" : ""}`}
            onClick={() => setTab("youtube")}
          >
            YouTube
          </button>
        </div>

        {/* 탭 내용 */}
        <div className="sched_body">
          {tab === "schedule" && (
            <ScheduleTab
              config={config}
              setConfig={setConfig}
              updateEntry={updateEntry}
              testing={testing}
              handleTest={handleTest}
            />
          )}
          {tab === "special" && thumbConfig && (
            <SpecialTab
              thumbConfig={thumbConfig}
              setThumbConfig={setThumbConfig}
            />
          )}
          {tab === "youtube" && (
            <YouTubeTab ytStatus={ytStatus} setYtStatus={setYtStatus} />
          )}
        </div>

        {/* 푸터 */}
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
      </div>

      <style jsx>{`
        .sched_overlay {
          position: fixed;
          top: 0;
          left: 0;
          width: 100%;
          height: 100%;
          background: rgba(0, 0, 0, 0.5);
          display: flex;
          align-items: center;
          justify-content: center;
          z-index: 11000;
        }
        .sched_panel {
          background: #fff;
          border-radius: 16px;
          width: 500px;
          max-width: 90vw;
          max-height: 85vh;
          overflow-y: auto;
          box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
        }
        .sched_header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 20px 24px 12px;
        }
        .sched_header h3 {
          margin: 0;
          font-size: 18px;
          font-weight: 700;
          color: #1f2937;
        }
        .sched_close {
          background: none;
          border: none;
          font-size: 24px;
          cursor: pointer;
          color: #6b7280;
          line-height: 1;
        }
        .sched_tabs {
          display: flex;
          gap: 0;
          padding: 0 24px;
          border-bottom: 1px solid #e5e7eb;
        }
        .sched_tab {
          padding: 10px 16px;
          font-size: 13px;
          font-weight: 500;
          color: #6b7280;
          background: none;
          border: none;
          border-bottom: 2px solid transparent;
          cursor: pointer;
          transition: all 0.15s;
        }
        .sched_tab.active {
          color: #1f3f62;
          border-bottom-color: #1f3f62;
          font-weight: 600;
        }
        .sched_tab:hover {
          color: #374151;
        }
        .sched_body {
          padding: 20px 24px;
          display: flex;
          flex-direction: column;
          gap: 14px;
          min-height: 200px;
        }
        .sched_footer {
          display: flex;
          justify-content: flex-end;
          gap: 8px;
          padding: 16px 24px 20px;
          border-top: 1px solid #e5e7eb;
        }
        .sched_cancel_btn {
          padding: 8px 20px;
          font-size: 13px;
          background: #f3f4f6;
          border: 1px solid #d1d5db;
          border-radius: 8px;
          cursor: pointer;
        }
        .sched_save_btn {
          padding: 8px 20px;
          font-size: 13px;
          font-weight: 600;
          background: #1f3f62;
          color: #fff;
          border: none;
          border-radius: 8px;
          cursor: pointer;
        }
        .sched_save_btn:hover {
          background: #2d5a8a;
        }
        .sched_save_btn:disabled {
          background: #9ca3af;
        }
      `}</style>
    </div>
  );
}

/* ── 정기 스케줄 탭 ── */
function ScheduleTab({
  config,
  setConfig,
  updateEntry,
  testing,
  handleTest,
}: {
  config: ScheduleConfig;
  setConfig: (c: ScheduleConfig) => void;
  updateEntry: (idx: number, patch: Partial<ScheduleEntry>) => void;
  testing: string | null;
  handleTest: (action: "countdown" | "trigger", wt: string) => void;
}) {
  return (
    <>
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
              {testing === `countdown_${entry.worshipType}` ? "..." : "\u23F1"}
            </button>
            <button
              className="sched_test_btn sched_test_trigger"
              disabled={testing !== null}
              onClick={() => handleTest("trigger", entry.worshipType)}
              title="즉시 실행 테스트"
            >
              {testing === `trigger_${entry.worshipType}` ? "..." : "\u25B6"}
            </button>
          </div>
          <style jsx>{`
            .sched_entry {
              display: flex;
              align-items: center;
              gap: 12px;
            }
            .sched_check {
              display: flex;
              align-items: center;
              gap: 8px;
              flex: 1;
            }
            .sched_check input[type="checkbox"] {
              width: 18px;
              height: 18px;
              accent-color: #1f3f62;
            }
            .sched_label {
              font-size: 14px;
              font-weight: 500;
              color: #374151;
            }
            .sched_day {
              font-size: 13px;
              color: #6b7280;
              min-width: 44px;
              text-align: center;
            }
            .sched_time {
              padding: 4px 8px;
              border: 1px solid #d1d5db;
              border-radius: 6px;
              font-size: 13px;
              background: #f9fafb;
            }
            .sched_test_btns {
              display: flex;
              gap: 4px;
              margin-left: 4px;
            }
            .sched_test_btn {
              width: 28px;
              height: 28px;
              border: 1px solid #d1d5db;
              border-radius: 6px;
              background: #f9fafb;
              cursor: pointer;
              font-size: 12px;
              display: flex;
              align-items: center;
              justify-content: center;
              transition: background 0.15s;
            }
            .sched_test_btn:hover {
              background: #e5e7eb;
            }
            .sched_test_btn:disabled {
              opacity: 0.5;
              cursor: default;
            }
            .sched_test_trigger {
              color: #059669;
            }
          `}</style>
        </div>
      ))}

      <div style={{ height: 1, background: "#e5e7eb", margin: "4px 0" }} />

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
        <style jsx>{`
          .sched_option {
            display: flex;
            justify-content: space-between;
            align-items: center;
          }
          .sched_option > span {
            font-size: 14px;
            color: #374151;
            font-weight: 500;
          }
          .sched_option_input {
            display: flex;
            align-items: center;
            gap: 6px;
          }
          .sched_option_input input {
            width: 60px;
            padding: 4px 8px;
            border: 1px solid #d1d5db;
            border-radius: 6px;
            font-size: 13px;
            text-align: center;
            background: #f9fafb;
          }
          .sched_option_input span {
            font-size: 13px;
            color: #6b7280;
          }
        `}</style>
      </div>

      <div className="sched_option">
        <span style={{ fontSize: 14, color: "#374151", fontWeight: 500 }}>
          OBS 자동 스트리밍
        </span>
        <button
          style={{
            padding: "6px 16px",
            borderRadius: 6,
            fontSize: 13,
            fontWeight: 600,
            cursor: "pointer",
            border: "none",
            transition: "background 0.15s",
            background: config.autoStream ? "#1f3f62" : "#e5e7eb",
            color: config.autoStream ? "#fff" : "#6b7280",
          }}
          onClick={() =>
            setConfig({ ...config, autoStream: !config.autoStream })
          }
        >
          {config.autoStream ? "ON" : "OFF"}
        </button>
      </div>
    </>
  );
}

/* ── 이미지 드롭존 ── */
function ImageDropZone({
  imageUrl,
  loading,
  onFile,
  onClear,
  height = 100,
  placeholder,
}: {
  imageUrl?: string;
  loading?: boolean;
  onFile: (f: File) => void;
  onClear?: () => void;
  height?: number;
  placeholder?: string;
}) {
  const [dragging, setDragging] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  return (
    <div
      className={`dz ${dragging ? "drag" : ""} ${imageUrl ? "has" : ""}`}
      onDragOver={(e) => { e.preventDefault(); e.stopPropagation(); if (!dragging) setDragging(true); }}
      onDragLeave={(e) => { e.stopPropagation(); setDragging(false); }}
      onDrop={(e) => {
        e.preventDefault();
        e.stopPropagation();
        setDragging(false);
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
            <button
              className="dz_clear"
              onClick={(e) => { e.stopPropagation(); onClear(); }}
              title="기본 배경으로"
            >
              &times;
            </button>
          )}
        </>
      ) : (
        <span className="dz_text">{placeholder || "이미지를 드래그하거나 클릭"}</span>
      )}
      <input
        ref={inputRef}
        type="file"
        accept="image/png,image/jpeg"
        style={{ display: "none" }}
        onChange={(e) => {
          const f = e.target.files?.[0];
          if (f) onFile(f);
          e.target.value = "";
        }}
      />
      <style jsx>{`
        .dz {
          position: relative;
          border: 2px dashed #d1d5db;
          border-radius: 8px;
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
          overflow: hidden;
          transition: border-color 0.15s, background 0.15s;
          background: #fafafa;
        }
        .dz:hover { border-color: #9ca3af; }
        .dz.drag { border-color: #3b82f6; background: #eff6ff; }
        .dz.has { border-style: solid; border-color: #e5e7eb; }
        .dz_text {
          font-size: 12px;
          color: #9ca3af;
          pointer-events: none;
          text-align: center;
          padding: 0 12px;
        }
        .dz_img {
          width: 100%;
          height: 100%;
          object-fit: cover;
        }
        .dz_clear {
          position: absolute;
          top: 4px;
          right: 4px;
          width: 22px;
          height: 22px;
          border-radius: 50%;
          background: rgba(0,0,0,0.55);
          color: #fff;
          border: none;
          font-size: 14px;
          line-height: 1;
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
        }
        .dz_clear:hover { background: rgba(239,68,68,0.8); }
      `}</style>
    </div>
  );
}

/* ── 기념 주일 탭 ── */

const WORSHIP_LABELS: Record<string, string> = {
  main_worship: "주일예배",
  after_worship: "오후예배",
  wed_worship: "수요예배",
  fri_worship: "금요예배",
};

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

  // 미리보기 날짜가 기념주일인지 확인
  const previewSpecial = thumbConfig.specials.find((s) => s.date === previewDate);

  // 새 항목 파일 선택 시 로컬 미리보기
  const handleNewBgFile = useCallback((file: File) => {
    setNewBgFile(file);
    const url = URL.createObjectURL(file);
    setNewBgPreview(url);
  }, []);

  // 기념주일 추가 — functional update로 stale closure 방지
  const addSpecial = async () => {
    if (!newDate || !newLabel) return;
    let bgPath = "";

    if (newBgFile) {
      setUploading("new");
      try {
        const slug = newLabel.replace(/\s+/g, "_").toLowerCase();
        const res = await apiClient.uploadThumbnailBg(newBgFile, slug);
        if (res.ok) bgPath = res.path;
      } catch (e) {
        console.error("배경 업로드 실패:", e);
      } finally {
        setUploading(null);
      }
    }

    const finalBgPath = bgPath;
    const finalDate = newDate;
    const finalLabel = newLabel;
    setThumbConfig(prev => prev ? {
      ...prev,
      specials: [
        ...prev.specials,
        { date: finalDate, label: finalLabel, background: finalBgPath, titleOverride: finalLabel },
      ],
    } : prev);
    setNewDate("");
    setNewLabel("");
    setNewBgFile(null);
    if (newBgPreview) { URL.revokeObjectURL(newBgPreview); setNewBgPreview(null); }
  };

  // 기념주일 배경 교체 — functional update
  const handleSpecialBgUpload = async (idx: number, file: File) => {
    setUploading(`special_${idx}`);
    try {
      const label = thumbConfig.specials[idx]?.label || "special";
      const slug = label.replace(/\s+/g, "_").toLowerCase();
      const res = await apiClient.uploadThumbnailBg(file, slug);
      if (res.ok) {
        setThumbConfig(prev => {
          if (!prev) return prev;
          const updated = [...prev.specials];
          updated[idx] = { ...updated[idx], background: res.path };
          return { ...prev, specials: updated };
        });
      }
    } catch (e) {
      console.error("배경 업로드 실패:", e);
    } finally {
      setUploading(null);
    }
  };

  // 기본 배경 교체 — functional update
  const handleDefaultBgUpload = async (worshipType: string, file: File) => {
    setUploading(`default_${worshipType}`);
    try {
      const res = await apiClient.uploadThumbnailBg(file, `default_${worshipType}`);
      if (res.ok) {
        setThumbConfig(prev => prev ? {
          ...prev,
          defaults: {
            ...prev.defaults,
            [worshipType]: { ...prev.defaults[worshipType], background: res.path },
          },
        } : prev);
      }
    } catch (e) {
      console.error("기본 배경 업로드 실패:", e);
    } finally {
      setUploading(null);
    }
  };

  const removeSpecial = (idx: number) => {
    setThumbConfig(prev => prev ? {
      ...prev,
      specials: prev.specials.filter((_, i) => i !== idx),
    } : prev);
  };

  const handlePreview = async () => {
    setGenerating(true);
    try {
      await apiClient.generateThumbnail(previewType, previewDate);
      setPreviewKey((k) => k + 1);
    } catch (e) {
      console.error(e);
    } finally {
      setGenerating(false);
    }
  };

  return (
    <>
      {/* ── 기본 배경 이미지 ── */}
      <div style={{ fontSize: 13, fontWeight: 600, color: "#374151" }}>
        기본 배경 이미지
      </div>
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 8 }}>
        {Object.entries(WORSHIP_LABELS).map(([type, label]) => {
          const theme = thumbConfig.defaults[type];
          const hasBg = !!theme?.background;
          return (
            <div key={type} style={{ display: "flex", flexDirection: "column", gap: 4 }}>
              <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
                <span style={{ fontSize: 12, color: "#374151", fontWeight: 500 }}>{label}</span>
                <span style={{
                  fontSize: 10, fontWeight: 600, padding: "1px 6px", borderRadius: 8,
                  background: hasBg ? "#d1fae5" : "#fef3c7",
                  color: hasBg ? "#059669" : "#d97706",
                }}>
                  {hasBg ? "설정됨" : "미설정"}
                </span>
              </div>
              <ImageDropZone
                imageUrl={hasBg ? apiClient.getThumbnailImageUrl(theme.background) : undefined}
                loading={uploading === `default_${type}`}
                onFile={(f) => handleDefaultBgUpload(type, f)}
                height={60}
                placeholder="드래그 또는 클릭"
              />
            </div>
          );
        })}
      </div>

      <div style={{ height: 1, background: "#e5e7eb", margin: "4px 0" }} />

      {/* ── 기념 주일 목록 ── */}
      <div style={{ fontSize: 13, fontWeight: 600, color: "#374151" }}>
        기념 주일
      </div>
      {thumbConfig.specials.length === 0 && (
        <div style={{ fontSize: 13, color: "#9ca3af", padding: "4px 0" }}>
          등록된 기념 주일이 없습니다
        </div>
      )}
      {thumbConfig.specials.map((s, i) => (
        <div key={i} style={{
          display: "flex", gap: 10, padding: "10px 12px",
          background: "#f9fafb", borderRadius: 8, alignItems: "stretch",
        }}>
          {/* 왼쪽: 드롭존 */}
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
              height={80}
              placeholder="배경 드래그"
            />
          </div>
          {/* 오른쪽: 정보 */}
          <div style={{ flex: 1, display: "flex", flexDirection: "column", justifyContent: "center", gap: 4 }}>
            <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
              <span style={{ fontSize: 14, fontWeight: 500, color: "#374151" }}>{s.label}</span>
              <span style={{
                fontSize: 10, fontWeight: 600, padding: "1px 7px", borderRadius: 8,
                background: s.background ? "#dbeafe" : "#f3f4f6",
                color: s.background ? "#2563eb" : "#6b7280",
              }}>
                {s.background ? "커스텀" : "기본 배경"}
              </span>
            </div>
            <span style={{ fontSize: 12, color: "#6b7280" }}>{s.date}</span>
          </div>
          {/* 삭제 */}
          <button
            onClick={() => removeSpecial(i)}
            style={{
              background: "none", border: "none", fontSize: 18,
              color: "#9ca3af", cursor: "pointer", alignSelf: "flex-start",
              lineHeight: 1, padding: 0,
            }}
            onMouseEnter={(e) => (e.currentTarget.style.color = "#ef4444")}
            onMouseLeave={(e) => (e.currentTarget.style.color = "#9ca3af")}
          >
            &times;
          </button>
        </div>
      ))}

      {/* ── 기념주일 추가 폼 ── */}
      <div style={{ height: 1, background: "#e5e7eb", margin: "2px 0" }} />
      <div style={{ fontSize: 13, fontWeight: 600, color: "#374151" }}>추가</div>
      <div style={{ display: "flex", gap: 10, alignItems: "stretch" }}>
        {/* 드롭존 */}
        <div style={{ width: 140, flexShrink: 0 }}>
          <ImageDropZone
            imageUrl={newBgPreview || undefined}
            loading={uploading === "new"}
            onFile={handleNewBgFile}
            onClear={newBgPreview ? () => {
              setNewBgFile(null);
              URL.revokeObjectURL(newBgPreview);
              setNewBgPreview(null);
            } : undefined}
            height={80}
            placeholder="배경 드래그 (선택)"
          />
        </div>
        {/* 입력 */}
        <div style={{ flex: 1, display: "flex", flexDirection: "column", gap: 6, justifyContent: "center" }}>
          <input
            type="date"
            value={newDate}
            onChange={(e) => setNewDate(e.target.value)}
            style={{
              padding: "5px 8px", border: "1px solid #d1d5db",
              borderRadius: 6, fontSize: 13, background: "#fff",
            }}
          />
          <div style={{ display: "flex", gap: 6 }}>
            <input
              type="text"
              placeholder="이름 (예: 부활절 예배)"
              value={newLabel}
              onChange={(e) => setNewLabel(e.target.value)}
              style={{
                flex: 1, padding: "5px 8px", border: "1px solid #d1d5db",
                borderRadius: 6, fontSize: 13, background: "#fff",
              }}
            />
            <button
              onClick={addSpecial}
              disabled={!newDate || !newLabel || uploading === "new"}
              style={{
                padding: "5px 14px", fontSize: 13, fontWeight: 600,
                background: "#1f3f62", color: "#fff", border: "none",
                borderRadius: 6, cursor: "pointer", whiteSpace: "nowrap",
                opacity: !newDate || !newLabel ? 0.5 : 1,
              }}
            >
              {uploading === "new" ? "..." : "추가"}
            </button>
          </div>
        </div>
      </div>

      {/* ── 썸네일 미리보기 ── */}
      <div style={{ height: 1, background: "#e5e7eb", margin: "4px 0" }} />
      <div style={{ fontSize: 13, fontWeight: 600, color: "#374151" }}>
        썸네일 미리보기
      </div>
      <div style={{ display: "flex", gap: 8, alignItems: "center", flexWrap: "wrap" }}>
        <select
          value={previewType}
          onChange={(e) => setPreviewType(e.target.value)}
          style={{
            padding: "6px 10px", border: "1px solid #d1d5db",
            borderRadius: 6, fontSize: 13, background: "#f9fafb",
          }}
        >
          {Object.entries(WORSHIP_LABELS).map(([k, v]) => (
            <option key={k} value={k}>{v}</option>
          ))}
        </select>
        <input
          type="date"
          value={previewDate}
          onChange={(e) => setPreviewDate(e.target.value)}
          style={{
            padding: "5px 8px", border: "1px solid #d1d5db",
            borderRadius: 6, fontSize: 13, background: "#f9fafb",
          }}
        />
        <button
          onClick={handlePreview}
          disabled={generating}
          style={{
            padding: "6px 14px", fontSize: 13, fontWeight: 600,
            background: "#059669", color: "#fff", border: "none",
            borderRadius: 6, cursor: "pointer",
          }}
        >
          {generating ? "생성 중..." : "생성"}
        </button>
      </div>
      {previewSpecial && (
        <div style={{
          fontSize: 12, color: "#2563eb", background: "#eff6ff",
          padding: "6px 10px", borderRadius: 6, fontWeight: 500,
        }}>
          기념주일 적용: {previewSpecial.label}
          {previewSpecial.background ? " (커스텀 배경)" : " (기본 배경)"}
        </div>
      )}
      {previewKey > 0 && (
        <img
          key={previewKey}
          src={apiClient.getThumbnailPreviewUrl(previewType, previewDate)}
          alt="썸네일 미리보기"
          style={{ width: "100%", borderRadius: 8, border: "1px solid #e5e7eb" }}
        />
      )}
    </>
  );
}

/* ── YouTube 탭 ── */
function YouTubeTab({
  ytStatus,
  setYtStatus,
}: {
  ytStatus: YouTubeStatus | null;
  setYtStatus: (s: YouTubeStatus) => void;
}) {
  const [uploading, setUploading] = useState(false);
  const [settingUp, setSettingUp] = useState(false);

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

  const connected = ytStatus?.connected ?? false;

  return (
    <>
      <div className="yt_status">
        <span className="yt_status_label">YouTube 연결 상태</span>
        <span className={`yt_badge ${connected ? "on" : "off"}`}>
          {connected ? "연결됨" : "미연결"}
        </span>
        <style jsx>{`
          .yt_status {
            display: flex;
            justify-content: space-between;
            align-items: center;
          }
          .yt_status_label {
            font-size: 14px;
            color: #374151;
            font-weight: 500;
          }
          .yt_badge {
            font-size: 12px;
            font-weight: 600;
            padding: 4px 12px;
            border-radius: 12px;
          }
          .yt_badge.on {
            background: #d1fae5;
            color: #059669;
          }
          .yt_badge.off {
            background: #fee2e2;
            color: #dc2626;
          }
        `}</style>
      </div>

      {!connected && (
        <div style={{ textAlign: "center", padding: "20px 0" }}>
          <p style={{ fontSize: 13, color: "#6b7280", marginBottom: 16 }}>
            YouTube 계정을 연결하면 예배 시작 시 썸네일이 자동으로 업로드됩니다.
          </p>
          <button
            onClick={handleConnect}
            style={{
              padding: "10px 24px",
              fontSize: 14,
              fontWeight: 600,
              background: "#dc2626",
              color: "#fff",
              border: "none",
              borderRadius: 8,
              cursor: "pointer",
            }}
          >
            YouTube 계정 연결
          </button>
        </div>
      )}

      {connected && (
        <>
          <div style={{ height: 1, background: "#e5e7eb", margin: "4px 0" }} />
          <div style={{ fontSize: 13, fontWeight: 600, color: "#374151" }}>
            OBS 스트리밍 설정
          </div>
          <p style={{ fontSize: 13, color: "#6b7280", margin: 0 }}>
            YouTube 스트림 키를 자동으로 가져와 OBS에 세팅합니다. (커스텀 RTMP 방식)
          </p>
          <button
            onClick={handleSetupOBS}
            disabled={settingUp}
            style={{
              padding: "8px 20px",
              fontSize: 13,
              fontWeight: 600,
              background: "#059669",
              color: "#fff",
              border: "none",
              borderRadius: 8,
              cursor: "pointer",
              alignSelf: "flex-start",
            }}
          >
            {settingUp ? "설정 중..." : "OBS 스트림 자동 세팅"}
          </button>

          <div style={{ height: 1, background: "#e5e7eb", margin: "4px 0" }} />
          <div style={{ fontSize: 13, fontWeight: 600, color: "#374151" }}>
            수동 업로드
          </div>
          <p style={{ fontSize: 13, color: "#6b7280", margin: 0 }}>
            현재 활성/예정 라이브 방송에 썸네일을 즉시 생성하고 업로드합니다.
          </p>
          <button
            onClick={handleManualUpload}
            disabled={uploading}
            style={{
              padding: "8px 20px",
              fontSize: 13,
              fontWeight: 600,
              background: "#1f3f62",
              color: "#fff",
              border: "none",
              borderRadius: 8,
              cursor: "pointer",
              alignSelf: "flex-start",
            }}
          >
            {uploading ? "업로드 중..." : "썸네일 생성 + 업로드"}
          </button>
        </>
      )}
    </>
  );
}

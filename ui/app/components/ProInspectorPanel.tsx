"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { useRecoilState } from "recoil";
import { inspectorOpenState } from "@/recoilState";
import { apiClient } from "@/lib/apiClient";
import { ThumbnailConfig, ScheduleConfig, ScheduleEntry } from "@/types";
import ConfirmModal from "./ConfirmModal";
import FeatureGate from "./FeatureGate";
import OBSSourcePanel from "./OBSSourcePanel";

type Category = "display" | "backgrounds" | "special" | "schedule" | "obs" | "config";
type FileItem = { name: string; url: string; size: number };

const TABS: { key: Category; label: string; desc: string }[] = [
  { key: "display",     label: "화면",   desc: "항목별 배경 이미지" },
  { key: "backgrounds", label: "배경",   desc: "기본 배경 · 가사 배경 · 비디오 배경" },
  { key: "special",     label: "썸네일", desc: "기념주일 썸네일 배경" },
  { key: "schedule",    label: "스케줄", desc: "정기 스트리밍 스케줄" },
  { key: "obs",         label: "OBS",    desc: "OBS 소스 및 씬 관리" },
  { key: "config",      label: "스타일", desc: "폰트 · 로고 · 오버레이" },
];

const WORSHIP_LABELS: Record<string, string> = {
  main_worship: "주일예배",
  after_worship: "오후예배",
  wed_worship: "수요예배",
  fri_worship: "금요예배",
};

const BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL
  || (typeof window !== "undefined" ? window.location.origin : "http://localhost:8080");

export default function ProInspectorPanel() {
  const [inspOpen, setInspOpen] = useRecoilState(inspectorOpenState);

  const [tab, setTab] = useState<Category>("display");
  const [files, setFiles] = useState<FileItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [dragOver, setDragOver] = useState(false);
  const [uploadName, setUploadName] = useState("");
  const [confirmTarget, setConfirmTarget] = useState<string | null>(null);
  const [toastMsg, setToastMsg] = useState<{ msg: string; type: "error" | "info" } | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // 특별일 탭 상태
  const [thumbConfig, setThumbConfig] = useState<ThumbnailConfig | null>(null);
  const [thumbSaving, setThumbSaving] = useState(false);

  // 스케줄 탭 상태
  const [scheduleConfig, setScheduleConfig] = useState<ScheduleConfig | null>(null);
  const [scheduleSaving, setScheduleSaving] = useState(false);

  // 이미지 캐시 버스팅 — 렌더 중 Date.now() 사용 금지 (hydration mismatch 방지)
  const [imgRevision, setImgRevision] = useState(0);

  // 배경 탭 (기본+가사 통합)
  const [defaultFiles, setDefaultFiles] = useState<FileItem[]>([]);
  const [lyricsFiles, setLyricsFiles] = useState<FileItem[]>([]);
  const fileInputDefaultRef = useRef<HTMLInputElement>(null);
  const fileInputLyricsRef = useRef<HTMLInputElement>(null);

  const showToast = (msg: string, type: "error" | "info" = "error") => {
    setToastMsg({ msg, type });
    setTimeout(() => setToastMsg(null), 3000);
  };

  const fetchFiles = useCallback(async () => {
    setLoading(true);
    try {
      if (tab === "backgrounds") {
        const [defRes, lyrRes] = await Promise.all([
          apiClient.getTemplates("display-default"),
          apiClient.getTemplates("lyrics"),
        ]);
        setDefaultFiles(defRes.files || []);
        setLyricsFiles(lyrRes.files || []);
      } else if (tab === "display") {
        const res = await apiClient.getTemplates("display");
        setFiles(res.files || []);
      }
      setImgRevision((r) => r + 1);
    } catch {
      setFiles([]);
      setDefaultFiles([]);
      setLyricsFiles([]);
    }
    setLoading(false);
  }, [tab]);

  useEffect(() => {
    if (inspOpen && (tab === "display" || tab === "backgrounds")) fetchFiles();
  }, [inspOpen, fetchFiles, tab]);

  useEffect(() => {
    if (inspOpen && tab === "special") {
      apiClient.getThumbnailConfig().then(setThumbConfig).catch(console.error);
    }
  }, [inspOpen, tab]);

  useEffect(() => {
    if (inspOpen && tab === "schedule" && !scheduleConfig) {
      apiClient.getSchedule().then(setScheduleConfig).catch(console.error);
    }
  }, [inspOpen, tab]); // eslint-disable-line react-hooks/exhaustive-deps

  const handleUpload = async (fileList: FileList | null, category?: "display" | "display-default" | "lyrics") => {
    if (!fileList || fileList.length === 0) return;
    const file = fileList[0];
    const ext = file.name.toLowerCase().split(".").pop();
    if (!["png", "jpg", "jpeg"].includes(ext || "")) {
      showToast("PNG/JPG 파일만 업로드 가능합니다.");
      return;
    }
    const cat = category || (tab as "display" | "display-default" | "lyrics");
    try {
      const name = cat === "display" ? uploadName.trim() || undefined : undefined;
      await apiClient.uploadTemplate(file, cat, name);
      if (cat === "display") setUploadName("");
      fetchFiles();
    } catch {
      showToast("업로드 실패");
    }
  };

  const handleDelete = async (name: string) => {
    try {
      await apiClient.deleteTemplate("display", name);
      fetchFiles();
    } catch {
      showToast("삭제 실패");
    }
  };

  const handleThumbSave = async () => {
    if (!thumbConfig) return;
    setThumbSaving(true);
    try {
      await apiClient.saveThumbnailConfig(thumbConfig);
      showToast("특별일 설정이 저장되었습니다.", "info");
    } catch (e) {
      console.error("썸네일 설정 저장 에러:", e);
      showToast("저장 실패");
    } finally {
      setThumbSaving(false);
    }
  };

  const onDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(true);
  };
  const onDragLeave = () => setDragOver(false);
  const onDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    handleUpload(e.dataTransfer.files);
  };

  if (!inspOpen) return null;

  return (
    <div
      className="flex flex-col bg-pro-surface border-l border-pro-border overflow-hidden"
      style={{ gridColumn: "4", gridRow: "2" }}
    >
      {/* header */}
      <div className="flex justify-between items-center px-3 py-2.5 border-b border-pro-border flex-shrink-0">
        <span className="text-[12px] font-semibold text-pro-text">Studio</span>
        <button
          onClick={() => setInspOpen(false)}
          className="bg-transparent border-none text-pro-text-dim text-base cursor-pointer leading-none hover:text-pro-text transition-colors p-0"
        >
          ✕
        </button>
      </div>

      {/* tabs */}
      <div className="flex flex-wrap border-b border-pro-border px-1 flex-shrink-0 overflow-x-hidden">
        {TABS.map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={`px-2 py-1.5 bg-transparent border-none border-b-2 text-[10px] cursor-pointer transition-colors whitespace-nowrap ${
              tab === t.key
                ? "border-[#4a9eff] text-[#4a9eff] font-semibold"
                : "border-transparent text-pro-text-dim font-normal hover:text-pro-text"
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* body */}
      <div className={`flex-1 overflow-y-auto overflow-x-hidden ${tab === "obs" ? "" : "px-3 py-3"}`}>
        {tab !== "obs" && (
          <div className="text-[10px] text-[#888] mb-2">
            {TABS.find((t) => t.key === tab)?.desc}
          </div>
        )}

        {tab === "config" ? (
          <DisplayConfigTab showToast={showToast} />
        ) : tab === "obs" ? (
          <OBSSourcePanel inline open onClose={() => setTab("display")} />
        ) : tab === "backgrounds" ? (
          <BackgroundsTab
            defaultFiles={defaultFiles}
            lyricsFiles={lyricsFiles}
            loading={loading}
            imgRevision={imgRevision}
            fileInputDefaultRef={fileInputDefaultRef}
            fileInputLyricsRef={fileInputLyricsRef}
            onUpload={handleUpload}
            baseUrl={BASE_URL}
            showToast={showToast}
          />
        ) : tab === "schedule" ? (
          <ScheduleTabInline
            config={scheduleConfig}
            saving={scheduleSaving}
            setSaving={setScheduleSaving}
            setConfig={setScheduleConfig}
          />
        ) : tab === "special" ? (
          thumbConfig ? (
            <SpecialSection
              thumbConfig={thumbConfig}
              setThumbConfig={setThumbConfig}
              onSave={handleThumbSave}
              saving={thumbSaving}
            />
          ) : (
            <div className="text-[#888] text-center py-5 text-xs">로딩 중...</div>
          )
        ) : (
          <>
            {/* upload zone */}
            <div
              onDragOver={onDragOver}
              onDragLeave={onDragLeave}
              onDrop={onDrop}
              onClick={() => fileInputRef.current?.click()}
              className={`border-2 border-dashed rounded-lg p-4 text-center cursor-pointer mb-3 transition-all ${
                dragOver
                  ? "border-[#4a9eff] bg-[rgba(74,158,255,0.1)]"
                  : "border-white/20 bg-transparent hover:border-white/40"
              }`}
            >
              <div className="text-[#aaa] text-[11px]">이미지를 드래그하거나 클릭하여 업로드</div>
              <div className="text-[#666] text-[10px] mt-1">PNG, JPG (최대 10MB)</div>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/png,image/jpeg"
                className="hidden"
                onChange={(e) => handleUpload(e.target.files)}
              />
            </div>

            {/* display 카테고리: 항목명 입력 */}
            {tab === "display" && (
              <div className="mb-2 flex gap-2 items-center">
                <input
                  value={uploadName}
                  onChange={(e) => setUploadName(e.target.value)}
                  placeholder="항목명 (예: 전주, 찬양) — 비워두면 파일명 사용"
                  className="flex-1 px-2 py-1.5 bg-white/10 border border-white/20 rounded-md text-white text-[10px] outline-none placeholder:text-[#666]"
                />
              </div>
            )}

            {/* file grid */}
            {loading ? (
              <div className="text-[#888] text-center py-5 text-xs">로딩 중...</div>
            ) : files.length === 0 ? (
              <div className="text-[#666] text-center py-5 text-xs">배경 이미지가 없습니다.</div>
            ) : (
              <div
                className={`grid gap-2 ${
                  tab === "display" ? "grid-cols-2" : "grid-cols-1"
                }`}
              >
                {files.map((f) => (
                  <div
                    key={f.name}
                    className="relative rounded-lg overflow-hidden bg-[#1a1a1a]"
                  >
                    <img
                      src={`${BASE_URL}${f.url}?v=${imgRevision}`}
                      alt={f.name}
                      className={`w-full object-cover block ${
                        tab === "display" ? "aspect-video" : "max-h-[160px] h-auto"
                      }`}
                    />
                    <div className="px-2 py-1 flex justify-between items-center">
                      <span className="text-[#ccc] text-[10px] overflow-hidden text-ellipsis whitespace-nowrap flex-1">
                        {f.name}
                      </span>
                      {tab === "display" && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            setConfirmTarget(f.name);
                          }}
                          className="bg-[rgba(255,60,60,0.2)] border-none text-[#ff6b6b] text-[10px] px-1.5 py-0.5 rounded cursor-pointer ml-1 flex-shrink-0 hover:bg-[rgba(255,60,60,0.35)] transition-colors"
                        >
                          삭제
                        </button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </>
        )}
      </div>

      {/* toast */}
      {toastMsg && (
        <div
          className={`absolute bottom-4 left-1/2 -translate-x-1/2 px-4 py-2 rounded-lg text-white text-[10px] shadow-lg z-10 ${
            toastMsg.type === "error" ? "bg-[#dc3545]" : "bg-[#204d87]"
          }`}
        >
          {toastMsg.msg}
        </div>
      )}

      <ConfirmModal
        open={confirmTarget !== null}
        message={`"${confirmTarget}" 배경을 삭제하시겠습니까?`}
        confirmLabel="삭제"
        danger
        onConfirm={() => {
          if (confirmTarget) handleDelete(confirmTarget);
          setConfirmTarget(null);
        }}
        onCancel={() => setConfirmTarget(null)}
      />
    </div>
  );
}

/* ── 스케줄 탭 인라인 컴포넌트 ── */
const WEEKDAY_NAMES = ["일", "월", "화", "수", "목", "금", "토"];

function ScheduleTabInline({
  config,
  saving,
  setSaving,
  setConfig,
}: {
  config: ScheduleConfig | null;
  saving: boolean;
  setSaving: (v: boolean) => void;
  setConfig: React.Dispatch<React.SetStateAction<ScheduleConfig | null>>;
}) {
  if (!config) {
    return <div className="text-[#888] text-center py-5 text-xs">로딩 중...</div>;
  }

  const updateEntry = (idx: number, patch: Partial<ScheduleEntry>) => {
    setConfig({
      ...config,
      entries: config.entries.map((e, i) => i === idx ? { ...e, ...patch } : e),
    });
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await apiClient.saveSchedule(config);
    } catch (e) {
      console.error("저장 에러:", e);
    } finally {
      setSaving(false);
    }
  };

  return (
    <FeatureGate feature="auto_scheduler">
      <div className="flex flex-col gap-3">
        {config.entries.map((entry, i) => (
          <div key={entry.worshipType} className="flex items-center gap-2">
            <label className="flex items-center gap-1.5 flex-1 cursor-pointer">
              <input
                type="checkbox"
                className="w-3.5 h-3.5 accent-[#4a9eff]"
                checked={entry.enabled}
                onChange={(e) => updateEntry(i, { enabled: e.target.checked })}
              />
              <span className="text-[11px] font-medium text-pro-text">{entry.label}</span>
            </label>
            <span className="text-[10px] text-pro-text-dim min-w-[32px] text-center">
              {WEEKDAY_NAMES[entry.weekday]}요일
            </span>
            <input
              type="time"
              className="px-2 py-1 border border-pro-border rounded-md text-[10px] bg-pro-elevated text-pro-text outline-none"
              value={`${String(entry.hour).padStart(2, "0")}:${String(entry.minute).padStart(2, "0")}`}
              onChange={(e) => {
                const [h, m] = e.target.value.split(":").map(Number);
                updateEntry(i, { hour: h, minute: m });
              }}
            />
          </div>
        ))}

        <div className="h-px bg-pro-border my-1" />

        {/* 카운트다운 */}
        <div className="flex justify-between items-center">
          <span className="text-[11px] font-medium text-pro-text">사전 카운트다운</span>
          <div className="flex items-center gap-1.5">
            <input
              type="number"
              min={1}
              max={30}
              value={config.countdownMinutes}
              onChange={(e) => setConfig({ ...config, countdownMinutes: Number(e.target.value) })}
              className="w-14 px-2 py-1 border border-pro-border rounded-md text-[10px] text-center bg-pro-elevated text-pro-text outline-none"
            />
            <span className="text-[10px] text-pro-text-dim">분</span>
          </div>
        </div>

        {/* OBS 자동 스트리밍 */}
        <div className="flex justify-between items-center">
          <span className="text-[11px] font-medium text-pro-text">OBS 자동 스트리밍</span>
          <button
            className={`px-3.5 py-1 rounded-md text-[10px] font-semibold cursor-pointer border-none transition-colors ${
              config.autoStream ? "bg-[#4a9eff] text-white" : "bg-pro-hover text-pro-text-dim"
            }`}
            onClick={() => setConfig({ ...config, autoStream: !config.autoStream })}
          >
            {config.autoStream ? "ON" : "OFF"}
          </button>
        </div>

        <div className="h-px bg-pro-border my-1" />

        <div className="flex justify-end">
          <button
            onClick={handleSave}
            disabled={saving}
            className="px-4 py-1.5 text-[10px] font-semibold bg-[#4a9eff] text-white border-none rounded-lg cursor-pointer hover:bg-[#3b8fe8] disabled:opacity-50 transition-colors"
          >
            {saving ? "저장 중..." : "저장"}
          </button>
        </div>
      </div>
    </FeatureGate>
  );
}

/* ── 이미지 드롭존 (다크 테마) ── */
function DarkImageDropZone({
  imageUrl, loading, onFile, onClear, height = 100, placeholder,
}: {
  imageUrl?: string; loading?: boolean; onFile: (f: File) => void;
  onClear?: () => void; height?: number; placeholder?: string;
}) {
  const [dragging, setDragging] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  return (
    <div
      className={`relative border-2 rounded-lg cursor-pointer flex items-center justify-center overflow-hidden transition-all ${
        dragging
          ? "border-[#4a9eff] bg-[rgba(74,158,255,0.1)] border-solid"
          : imageUrl
          ? "border-white/20 border-solid"
          : "border-white/15 border-dashed bg-white/5 hover:border-white/30"
      }`}
      style={{ height }}
      onDragOver={(e) => { e.preventDefault(); e.stopPropagation(); if (!dragging) setDragging(true); }}
      onDragLeave={(e) => { e.stopPropagation(); setDragging(false); }}
      onDrop={(e) => {
        e.preventDefault(); e.stopPropagation(); setDragging(false);
        const file = e.dataTransfer.files[0];
        if (file && file.type.startsWith("image/")) onFile(file);
      }}
      onClick={(e) => { e.stopPropagation(); inputRef.current?.click(); }}
    >
      {loading ? (
        <span className="text-[10px] text-[#888] pointer-events-none text-center px-3">
          업로드 중...
        </span>
      ) : imageUrl ? (
        <>
          <img src={imageUrl} alt="" className="w-full h-full object-cover" />
          {onClear && (
            <button
              className="absolute top-1 right-1 w-[20px] h-[20px] rounded-full bg-black/70 text-white border-none text-sm leading-none cursor-pointer flex items-center justify-center hover:bg-[rgba(239,68,68,0.8)] transition-colors"
              onClick={(e) => { e.stopPropagation(); onClear(); }}
              title="기본 배경으로"
            >
              &times;
            </button>
          )}
        </>
      ) : (
        <span className="text-[10px] text-[#666] pointer-events-none text-center px-3">
          {placeholder || "이미지를 드래그하거나 클릭"}
        </span>
      )}
      <input
        ref={inputRef}
        type="file"
        accept="image/png,image/jpeg"
        className="hidden"
        onChange={(e) => { const f = e.target.files?.[0]; if (f) onFile(f); e.target.value = ""; }}
      />
    </div>
  );
}

/* ── 특별일 설정 섹션 ── */
function SpecialSection({
  thumbConfig,
  setThumbConfig,
  onSave,
  saving,
}: {
  thumbConfig: ThumbnailConfig;
  setThumbConfig: React.Dispatch<React.SetStateAction<ThumbnailConfig | null>>;
  onSave: () => void;
  saving: boolean;
}) {
  const [newDate, setNewDate] = useState("");
  const [newLabel, setNewLabel] = useState("");
  const [newBgFile, setNewBgFile] = useState<File | null>(null);
  const [newBgPreview, setNewBgPreview] = useState<string | null>(null);
  const [uploading, setUploading] = useState<string | null>(null);
  const [previewType, setPreviewType] = useState("main_worship");
  const [previewDate, setPreviewDate] = useState("");
  useEffect(() => { setPreviewDate(new Date().toISOString().slice(0, 10)); }, []);
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

  const inputClass = "px-2 py-1 border border-white/20 rounded-md text-[10px] bg-white/10 text-white outline-none placeholder:text-[#666]";

  return (
    <div className="flex flex-col gap-3">
      {/* 기본 배경 이미지 */}
      <div className="text-[10px] font-semibold text-[#ccc]">기본 배경 이미지</div>
      <div className="grid grid-cols-2 gap-2">
        {Object.entries(WORSHIP_LABELS).map(([type, label]) => {
          const theme = thumbConfig.defaults[type];
          const hasBg = !!theme?.background;
          return (
            <div key={type} className="flex flex-col gap-1">
              <div className="flex items-center gap-1">
                <span className="text-[10px] font-medium text-[#ccc]">{label}</span>
                <span className={`text-[9px] font-semibold px-1 py-0.5 rounded-lg ${
                  hasBg ? "bg-[#052e16] text-[#4ade80]" : "bg-[#451a03] text-[#fb923c]"
                }`}>
                  {hasBg ? "설정됨" : "미설정"}
                </span>
              </div>
              <DarkImageDropZone
                imageUrl={hasBg ? apiClient.getThumbnailImageUrl(theme.background) : undefined}
                loading={uploading === `default_${type}`}
                onFile={(f) => handleDefaultBgUpload(type, f)}
                height={55}
                placeholder="드래그 또는 클릭"
              />
            </div>
          );
        })}
      </div>

      <div className="h-px bg-white/10" />

      {/* 기념 주일 목록 */}
      <div className="text-[10px] font-semibold text-[#ccc]">기념 주일</div>
      {thumbConfig.specials.length === 0 && (
        <div className="text-[10px] text-[#666] py-1">등록된 기념 주일이 없습니다</div>
      )}
      {thumbConfig.specials.map((s, i) => (
        <div key={i} className="flex gap-2 px-2 py-2 bg-white/5 rounded-lg items-stretch border border-white/10">
          <div className="w-[100px] flex-shrink-0">
            <DarkImageDropZone
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
              height={70}
              placeholder="배경 드래그"
            />
          </div>
          <div className="flex-1 flex flex-col justify-center gap-1">
            <div className="flex items-center gap-1">
              <span className="text-xs font-medium text-white">{s.label}</span>
              <span className={`text-[9px] font-semibold px-1 py-0.5 rounded-lg ${
                s.background ? "bg-[#1e3a5f] text-[#60a5fa]" : "bg-white/10 text-[#aaa]"
              }`}>
                {s.background ? "커스텀" : "기본 배경"}
              </span>
            </div>
            <span className="text-[10px] text-[#888]">{s.date}</span>
          </div>
          <button
            onClick={() => removeSpecial(i)}
            className="bg-transparent border-none text-base text-[#666] cursor-pointer self-start leading-none p-0 hover:text-[#ff6b6b] transition-colors"
          >
            &times;
          </button>
        </div>
      ))}

      <div className="h-px bg-white/10" />

      {/* 추가 */}
      <div className="text-[10px] font-semibold text-[#ccc]">추가</div>
      <div className="flex flex-col gap-1.5">
        <input
          type="date"
          value={newDate}
          onChange={(e) => setNewDate(e.target.value)}
          className={`${inputClass} w-full`}
        />
        <input
          type="text"
          placeholder="이름 (예: 부활절 예배)"
          value={newLabel}
          onChange={(e) => setNewLabel(e.target.value)}
          className={`${inputClass} w-full`}
        />
        <div className="flex gap-1.5 items-end">
          <div className="flex-1">
            <DarkImageDropZone
              imageUrl={newBgPreview || undefined}
              loading={uploading === "new"}
              onFile={handleNewBgFile}
              onClear={newBgPreview ? () => {
                setNewBgFile(null); URL.revokeObjectURL(newBgPreview); setNewBgPreview(null);
              } : undefined}
              height={50}
              placeholder="배경 이미지 (선택)"
            />
          </div>
          <button
            onClick={addSpecial}
            disabled={!newDate || !newLabel || uploading === "new"}
            className={`px-3 py-1.5 text-[10px] font-semibold bg-[#4a9eff] text-white border-none rounded-md cursor-pointer whitespace-nowrap transition-opacity flex-shrink-0 ${
              !newDate || !newLabel ? "opacity-40" : "hover:bg-[#3b8fe8]"
            }`}
          >
            {uploading === "new" ? "..." : "추가"}
          </button>
        </div>
      </div>

      <div className="h-px bg-white/10" />

      {/* 썸네일 미리보기 */}
      <div className="text-[10px] font-semibold text-[#ccc]">썸네일 미리보기</div>
      <div className="flex gap-1.5 items-center flex-wrap">
        <select
          value={previewType}
          onChange={(e) => setPreviewType(e.target.value)}
          className="px-2 py-1 border border-white/20 rounded-md text-[10px] bg-white/10 text-white outline-none"
        >
          {Object.entries(WORSHIP_LABELS).map(([k, v]) => (
            <option key={k} value={k} className="bg-[#2c2c2c]">{v}</option>
          ))}
        </select>
        <input
          type="date"
          value={previewDate}
          onChange={(e) => setPreviewDate(e.target.value)}
          className="px-2 py-1 border border-white/20 rounded-md text-[10px] bg-white/10 text-white outline-none"
        />
        <button
          onClick={handlePreview}
          disabled={generating}
          className="px-2.5 py-1 text-[10px] font-semibold bg-[#22c55e] text-white border-none rounded-md cursor-pointer hover:bg-[#16a34a] disabled:opacity-50 transition-colors"
        >
          {generating ? "생성 중..." : "생성"}
        </button>
      </div>
      {previewSpecial && (
        <div className="text-[10px] text-[#60a5fa] bg-[#1e3a5f] px-2 py-1.5 rounded-md font-medium">
          기념주일 적용: {previewSpecial.label}{previewSpecial.background ? " (커스텀 배경)" : " (기본 배경)"}
        </div>
      )}
      {previewKey > 0 && (
        <img
          key={previewKey}
          src={apiClient.getThumbnailPreviewUrl(previewType, previewDate)}
          alt="썸네일 미리보기"
          className="w-full rounded-lg border border-white/20"
        />
      )}

      {/* 저장 버튼 */}
      <div className="flex justify-end pt-1">
        <button
          onClick={onSave}
          disabled={saving}
          className="px-4 py-1.5 text-[10px] font-semibold bg-[#4a9eff] text-white border-none rounded-lg cursor-pointer hover:bg-[#3b8fe8] disabled:opacity-50 transition-colors"
        >
          {saving ? "저장 중..." : "저장"}
        </button>
      </div>
    </div>
  );
}

/* ── Display 스타일 설정 탭 (자동 저장) ── */
const DISPLAY_FONTS = [
  { value: "default",        label: "기본 (맑은 고딕)" },
  { value: "noto-sans-kr",   label: "Noto Sans KR" },
  { value: "gowun-dodum",    label: "Gowun Dodum" },
  { value: "nanum-myeongjo", label: "나눔명조" },
  { value: "black-han-sans", label: "Black Han Sans" },
];
const LOGO_POSITIONS = [
  { key: "top-left",     label: "↖ 좌상" },
  { key: "top-right",    label: "↗ 우상" },
  { key: "bottom-left",  label: "↙ 좌하" },
  { key: "bottom-right", label: "↘ 우하" },
];

function DisplayConfigTab({ showToast }: { showToast: (msg: string, type?: "error" | "info") => void }) {
  const [config, setConfig] = useState({
    font: "default",
    overlayBgOpacity: 0.75,
    overlayTextColor: "#ffffff",
    overlayPosition: "flex-end",
    overlayFontScale: 1.0,
    globalVideoBg: "",
    logoPosition: "bottom-right",
    logoSizePercent: 18,
  });
  const [logoExists, setLogoExists] = useState(false);
  const [logoTs, setLogoTs] = useState(0);
  const [savedAt, setSavedAt] = useState(0); // 자동 저장 표시용
  const isLoadedUpdate = useRef(false);
  const loaded = useRef(false);
  const fileRef = useRef<HTMLInputElement>(null);

  // 초기 로드
  useEffect(() => {
    apiClient.getDisplayConfig().then((c) => {
      isLoadedUpdate.current = true;
      setConfig({
        font: c.font || "default",
        overlayBgOpacity: c.overlayBgOpacity ?? 0.75,
        overlayTextColor: c.overlayTextColor || "#ffffff",
        overlayPosition: c.overlayPosition || "flex-end",
        overlayFontScale: c.overlayFontScale ?? 1.0,
        globalVideoBg: c.globalVideoBg || "",
        logoPosition: c.logoPosition || "bottom-right",
        logoSizePercent: c.logoSizePercent ?? 18,
      });
      loaded.current = true;
    }).catch(() => { loaded.current = true; });
    apiClient.hasLogo().then(setLogoExists).catch(() => {});
  }, []);

  // 자동 저장 (600ms 디바운스) — 초기 로드는 skip
  useEffect(() => {
    if (isLoadedUpdate.current) { isLoadedUpdate.current = false; return; }
    if (!loaded.current) return;
    const timer = setTimeout(() => {
      apiClient.saveDisplayConfig(config)
        .then(() => setSavedAt(Date.now()))
        .catch(() => {});
    }, 600);
    return () => clearTimeout(timer);
  }, [config]);

  const updateConfig = (patch: Partial<typeof config>) =>
    setConfig((prev) => ({ ...prev, ...patch }));

  const handleLogoUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    try {
      await apiClient.uploadLogo(file);
      setLogoExists(true);
      setLogoTs(Date.now());
      showToast("로고가 업로드되었습니다.", "info");
    } catch {
      showToast("업로드 실패");
    }
    e.target.value = "";
  };

  const handleLogoDelete = async () => {
    try {
      await apiClient.deleteLogo();
      setLogoExists(false);
      showToast("로고가 삭제되었습니다.", "info");
    } catch {
      showToast("삭제 실패");
    }
  };

  const inputCls = "px-2 py-1 border border-pro-border rounded-md text-[10px] bg-pro-elevated text-pro-text outline-none";
  const rowCls = "flex justify-between items-center gap-2";
  const labelCls = "text-[11px] font-medium text-pro-text";
  const divider = <div className="h-px bg-pro-border my-1" />;

  return (
    <div className="flex flex-col gap-3">
      {/* 자동 저장 표시 */}
      {savedAt > 0 && (
        <div className="flex items-center gap-1 text-[9px] text-[#4a9eff] opacity-70">
          <span>✓</span><span>자동 저장됨</span>
        </div>
      )}

      {/* 프로젝터 로고 */}
      <div className="text-[10px] font-semibold text-[#ccc]">프로젝터 로고</div>
      <p className="text-[10px] text-[#555] -mt-2">Display 화면 표시용 · OBS 방송 로고와 별개</p>
      <div className="flex items-center gap-2 flex-wrap">
        {logoExists ? (
          <>
            <img
              src={`${apiClient.getLogoUrl()}?t=${logoTs}`}
              alt="로고"
              className="h-8 object-contain rounded border border-pro-border bg-pro-elevated px-2 py-1"
            />
            <button className="px-2 py-1 text-[10px] bg-pro-hover border border-pro-border rounded cursor-pointer hover:opacity-80 text-pro-text"
              onClick={() => fileRef.current?.click()}>교체</button>
            <button className="px-2 py-1 text-[10px] text-[#ff6b6b] border border-[rgba(255,107,107,0.3)] rounded cursor-pointer hover:opacity-80 bg-transparent"
              onClick={handleLogoDelete}>삭제</button>
          </>
        ) : (
          <button className="px-3 py-1.5 text-[10px] bg-[#4a9eff] text-white border-none rounded-md cursor-pointer hover:opacity-90"
            onClick={() => fileRef.current?.click()}>로고 업로드</button>
        )}
        <input ref={fileRef} type="file" accept="image/png,image/jpeg" className="hidden" onChange={handleLogoUpload} />
      </div>

      {/* 로고 위치 (OBS 로고 탭과 동일한 방식) */}
      {logoExists && (
        <>
          <div className={rowCls}>
            <span className={labelCls}>위치</span>
            {/* 미니 프리뷰 */}
            <div className="relative w-16 h-10 bg-[#1a1a1a] rounded border border-pro-border flex-shrink-0">
              <div
                className="absolute w-[30%] h-[30%] bg-[rgba(74,158,255,0.7)] rounded-sm"
                style={{
                  ...(config.logoPosition.startsWith("top") ? { top: 3 } : { bottom: 3 }),
                  ...(config.logoPosition.endsWith("right") ? { right: 3 } : { left: 3 }),
                }}
              />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-1">
            {LOGO_POSITIONS.map((p) => (
              <button key={p.key}
                onClick={() => updateConfig({ logoPosition: p.key })}
                className={`py-1 rounded text-[10px] cursor-pointer border transition-all ${
                  config.logoPosition === p.key
                    ? "bg-[rgba(74,158,255,0.2)] border-[#4a9eff] text-[#4a9eff]"
                    : "bg-white/[0.06] border-white/15 text-[#aaa] hover:border-white/30"
                }`}
              >{p.label}</button>
            ))}
          </div>
          <div className={rowCls}>
            <span className={labelCls}>크기: {Math.round(config.logoSizePercent)}%</span>
            <input type="range" min={5} max={30} step={1} value={config.logoSizePercent}
              onChange={(e) => updateConfig({ logoSizePercent: Number(e.target.value) })}
              className="w-24 accent-[#4a9eff]" />
          </div>
        </>
      )}

      {divider}

      {/* Display 폰트 */}
      <div className={rowCls}>
        <span className={labelCls}>Display 폰트</span>
        <select className={inputCls} value={config.font} onChange={(e) => updateConfig({ font: e.target.value })}>
          {DISPLAY_FONTS.map((f) => (
            <option key={f.value} value={f.value} className="bg-[#1a1a1a]">{f.label}</option>
          ))}
        </select>
      </div>

      {divider}
      <div className="text-[10px] font-semibold text-[#ccc]">오버레이 커스터마이징</div>

      {/* 배경 투명도 */}
      <div className={rowCls}>
        <span className={labelCls}>배경 투명도</span>
        <div className="flex items-center gap-2">
          <input type="range" min={0} max={1} step={0.05} value={config.overlayBgOpacity}
            onChange={(e) => updateConfig({ overlayBgOpacity: Number(e.target.value) })}
            className="w-24 accent-[#4a9eff]" />
          <span className="text-[10px] text-[#888] w-8 text-right">{Math.round(config.overlayBgOpacity * 100)}%</span>
        </div>
      </div>

      {/* 텍스트 색상 */}
      <div className={rowCls}>
        <span className={labelCls}>텍스트 색상</span>
        <div className="flex items-center gap-2">
          <input type="color" value={config.overlayTextColor}
            onChange={(e) => updateConfig({ overlayTextColor: e.target.value })}
            className="w-7 h-7 rounded cursor-pointer border border-pro-border bg-transparent p-0.5" />
          <span className="text-[10px] text-[#888] font-mono">{config.overlayTextColor}</span>
        </div>
      </div>

      {/* 텍스트 위치 */}
      <div className={rowCls}>
        <span className={labelCls}>텍스트 위치</span>
        <select className={inputCls} value={config.overlayPosition}
          onChange={(e) => updateConfig({ overlayPosition: e.target.value })}>
          <option value="flex-end" className="bg-[#1a1a1a]">하단</option>
          <option value="center" className="bg-[#1a1a1a]">중앙</option>
          <option value="flex-start" className="bg-[#1a1a1a]">상단</option>
        </select>
      </div>

      {/* 폰트 배율 */}
      <div className={rowCls}>
        <span className={labelCls}>폰트 배율</span>
        <div className="flex items-center gap-2">
          <input type="range" min={0.5} max={2.0} step={0.1} value={config.overlayFontScale}
            onChange={(e) => updateConfig({ overlayFontScale: Number(e.target.value) })}
            className="w-24 accent-[#4a9eff]" />
          <span className="text-[10px] text-[#888] w-8 text-right">{config.overlayFontScale.toFixed(1)}×</span>
        </div>
      </div>

    </div>
  );
}

/* ── 배경 탭 (기본 + 가사 + 비디오 통합) ── */
function BackgroundsTab({
  defaultFiles,
  lyricsFiles,
  loading,
  imgRevision,
  fileInputDefaultRef,
  fileInputLyricsRef,
  onUpload,
  baseUrl,
  showToast,
}: {
  defaultFiles: FileItem[];
  lyricsFiles: FileItem[];
  loading: boolean;
  imgRevision: number;
  fileInputDefaultRef: React.RefObject<HTMLInputElement>;
  fileInputLyricsRef: React.RefObject<HTMLInputElement>;
  onUpload: (files: FileList | null, category?: "display" | "display-default" | "lyrics") => void;
  baseUrl: string;
  showToast: (msg: string, type?: "error" | "info") => void;
}) {
  const [videoBgList, setVideoBgList] = useState<{ filename: string; url: string }[]>([]);
  const [globalVideoBg, setGlobalVideoBg] = useState("");
  const videoBgRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    apiClient.listVideoBg().then(setVideoBgList).catch(() => {});
    apiClient.getDisplayConfig().then((c) => setGlobalVideoBg(c.globalVideoBg || "")).catch(() => {});
  }, []);

  const saveVideoBg = async (filename: string) => {
    try {
      const cfg = await apiClient.getDisplayConfig();
      await apiClient.saveDisplayConfig({ ...cfg, globalVideoBg: filename });
      setGlobalVideoBg(filename);
    } catch {
      showToast("저장 실패");
    }
  };

  const handleVideoBgUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    try {
      const res = await apiClient.uploadVideoBg(file);
      const newList = await apiClient.listVideoBg();
      setVideoBgList(newList);
      await saveVideoBg(res.filename);
      showToast("비디오 배경이 업로드되었습니다.", "info");
    } catch {
      showToast("업로드 실패");
    }
    e.target.value = "";
  };

  const handleVideoBgDelete = async (filename: string) => {
    try {
      await apiClient.deleteVideoBg(filename);
      const newList = await apiClient.listVideoBg();
      setVideoBgList(newList);
      if (globalVideoBg === filename) await saveVideoBg("");
      showToast("삭제되었습니다.", "info");
    } catch {
      showToast("삭제 실패");
    }
  };

  return (
    <div className="flex flex-col gap-4">
      {/* 기본 배경 (Frame 2) */}
      <div>
        <div className="flex justify-between items-center mb-2">
          <span className="text-[10px] font-semibold text-[#ccc]">기본 배경 (Frame 2)</span>
          <button
            className="px-2 py-1 text-[10px] bg-[#4a9eff] text-white border-none rounded cursor-pointer hover:opacity-90"
            onClick={() => fileInputDefaultRef.current?.click()}
          >업로드</button>
          <input
            ref={fileInputDefaultRef}
            type="file"
            accept="image/png,image/jpeg"
            className="hidden"
            onChange={(e) => onUpload(e.target.files, "display-default")}
          />
        </div>
        {loading ? (
          <div className="text-[#888] text-center py-3 text-xs">로딩 중...</div>
        ) : defaultFiles.length === 0 ? (
          <div className="text-[#555] text-center py-3 text-xs">기본 배경 없음</div>
        ) : (
          <div className="flex flex-col gap-1.5">
            {defaultFiles.map((f) => (
              <div key={f.name} className="relative rounded overflow-hidden bg-[#1a1a1a]">
                <img
                  src={`${baseUrl}${f.url}?v=${imgRevision}`}
                  alt={f.name}
                  className="w-full object-cover max-h-[120px]"
                />
                <div className="px-2 py-1">
                  <span className="text-[#aaa] text-[10px] truncate block">{f.name}</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="h-px bg-pro-border" />

      {/* 가사 배경 (Frame 1) */}
      <div>
        <div className="flex justify-between items-center mb-2">
          <span className="text-[10px] font-semibold text-[#ccc]">가사 배경 (Frame 1)</span>
          <button
            className="px-2 py-1 text-[10px] bg-[#4a9eff] text-white border-none rounded cursor-pointer hover:opacity-90"
            onClick={() => fileInputLyricsRef.current?.click()}
          >업로드</button>
          <input
            ref={fileInputLyricsRef}
            type="file"
            accept="image/png,image/jpeg"
            className="hidden"
            onChange={(e) => onUpload(e.target.files, "lyrics")}
          />
        </div>
        {loading ? (
          <div className="text-[#888] text-center py-3 text-xs">로딩 중...</div>
        ) : lyricsFiles.length === 0 ? (
          <div className="text-[#555] text-center py-3 text-xs">가사 배경 없음</div>
        ) : (
          <div className="flex flex-col gap-1.5">
            {lyricsFiles.map((f) => (
              <div key={f.name} className="relative rounded overflow-hidden bg-[#1a1a1a]">
                <img
                  src={`${baseUrl}${f.url}?v=${imgRevision}`}
                  alt={f.name}
                  className="w-full object-cover max-h-[120px]"
                />
                <div className="px-2 py-1">
                  <span className="text-[#aaa] text-[10px] truncate block">{f.name}</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="h-px bg-pro-border" />

      {/* 비디오 배경 */}
      <div>
        <div className="flex justify-between items-center mb-1">
          <span className="text-[10px] font-semibold text-[#ccc]">비디오 배경</span>
          <button
            className="px-2 py-1 text-[10px] bg-[#4a9eff] text-white border-none rounded cursor-pointer hover:opacity-90"
            onClick={() => videoBgRef.current?.click()}
          >업로드</button>
          <input ref={videoBgRef} type="file" accept="video/mp4,video/webm,video/quicktime,.mov"
            className="hidden" onChange={handleVideoBgUpload} />
        </div>
        <p className="text-[10px] text-[#555] mb-2">MP4 · WebM · 최대 300MB · 클릭으로 적용/해제</p>

        <div className="grid grid-cols-2 gap-2">
          {/* 없음 (검정 배경) 타일 */}
          <div
            onClick={() => saveVideoBg("")}
            className={`relative rounded-lg overflow-hidden cursor-pointer border-2 transition-all flex flex-col items-center justify-center bg-[#0a0a0a] aspect-video ${
              globalVideoBg === "" ? "border-[#4a9eff]" : "border-transparent hover:border-white/30"
            }`}
          >
            <span className="text-[16px] mb-0.5">⬛</span>
            <span className="text-[10px] text-[#666]">없음</span>
            {globalVideoBg === "" && (
              <div className="absolute top-1 right-1 bg-[#4a9eff] text-white text-[8px] font-bold px-1.5 py-0.5 rounded">
                적용 중
              </div>
            )}
          </div>

          {videoBgList.map((v) => {
            const isActive = globalVideoBg === v.filename;
            return (
              <div
                key={v.filename}
                onClick={() => saveVideoBg(isActive ? "" : v.filename)}
                className={`relative rounded-lg overflow-hidden cursor-pointer border-2 transition-all ${
                  isActive ? "border-[#4a9eff]" : "border-transparent hover:border-white/30"
                }`}
              >
                <video
                  src={`${baseUrl}${v.url}`}
                  className="w-full aspect-video object-cover block bg-[#1a1a1a]"
                  muted
                  preload="metadata"
                  onLoadedMetadata={(e) => {
                    (e.target as HTMLVideoElement).currentTime = 1;
                  }}
                />
                <div className="flex items-center justify-between px-1.5 py-1 bg-black/60 absolute bottom-0 left-0 right-0">
                  <span className="text-[9px] text-white/80 truncate flex-1">{v.filename}</span>
                  <button
                    className="text-[9px] text-[#ff6b6b] hover:opacity-80 cursor-pointer ml-1 shrink-0 bg-transparent border-none"
                    onClick={(e) => { e.stopPropagation(); handleVideoBgDelete(v.filename); }}
                  >삭제</button>
                </div>
                {isActive && (
                  <div className="absolute top-1 right-1 bg-[#4a9eff] text-white text-[8px] font-bold px-1.5 py-0.5 rounded">
                    적용 중
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

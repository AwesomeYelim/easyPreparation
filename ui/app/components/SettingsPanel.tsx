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
    <div
      className="fixed inset-0 z-[11000] flex items-center justify-center bg-black/50"
      onClick={onClose}
    >
      <div
        className="bg-[var(--surface-elevated)] rounded-2xl w-[480px] max-w-[90vw] max-h-[85vh] overflow-hidden flex flex-col shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        {/* header */}
        <div className="flex justify-between items-center px-6 pt-5 pb-3">
          <div className="flex flex-col gap-0.5">
            <h3 className="m-0 text-lg font-bold text-[var(--text-primary)]">설정</h3>
            {currentVersion && (
              <span className="text-[11px] text-[var(--text-muted)] font-mono tracking-[0.3px]">
                v{currentVersion}
              </span>
            )}
          </div>
          <button
            className="bg-transparent border-none text-2xl cursor-pointer text-[var(--text-secondary)] leading-none"
            onClick={onClose}
          >
            &times;
          </button>
        </div>

        {/* tabs */}
        <div className="flex gap-0 px-6 border-b border-[var(--border)]">
          {(["general", "special"] as Tab[]).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`px-4 py-2.5 text-xs font-medium bg-transparent border-none border-b-2 cursor-pointer transition-all ${
                tab === t
                  ? "text-[var(--accent)] border-b-[var(--accent)] font-semibold"
                  : "text-[var(--text-secondary)] border-b-transparent hover:text-[var(--text-primary)]"
              }`}
            >
              {t === "general" ? "일반" : "기념주일"}
            </button>
          ))}
        </div>

        {/* body */}
        <div className="px-6 py-5 flex flex-col gap-4 min-h-[200px] overflow-y-auto flex-1">
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

        {/* error */}
        {error && (
          <div className="mx-6 px-3 py-2 bg-[var(--error-bg)] text-[var(--error)] border border-[var(--error-border)] rounded-lg text-xs">
            {error}
          </div>
        )}

        {/* footer */}
        <div className="flex justify-end gap-2 px-6 pb-5 pt-4 border-t border-[var(--border)]">
          <button
            className="px-5 py-2 text-xs bg-[var(--surface-hover)] border border-[var(--border-input)] rounded-xl cursor-pointer"
            onClick={onClose}
          >
            취소
          </button>
          <button
            className="px-5 py-2 text-xs font-semibold bg-[var(--accent)] text-[var(--surface-elevated)] border-none rounded-xl cursor-pointer hover:bg-[var(--accent-hover)] disabled:bg-[var(--text-muted)]"
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? "저장 중..." : "저장"}
          </button>
        </div>
      </div>

      <SchedulePanel open={scheduleOpen} onClose={() => setScheduleOpen(false)} />
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

  const rowClass = "flex justify-between items-center gap-3";
  const labelClass = "text-sm font-medium text-[var(--text-primary)]";
  const inputClass =
    "px-3 py-1.5 border border-[var(--border-input)] rounded-lg text-xs bg-[var(--surface-input)] min-w-[140px]";
  const numberInputClass =
    "px-3 py-1.5 border border-[var(--border-input)] rounded-lg text-xs bg-[var(--surface-input)] w-20";

  return (
    <>
      <label className={rowClass}>
        <span className={labelClass}>선호 성경 번역</span>
        <select
          className={inputClass}
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
      <label className={rowClass}>
        <span className={labelClass}>테마</span>
        <select
          className={inputClass}
          value={local.theme}
          onChange={(e) => setLocal({ ...local, theme: e.target.value })}
        >
          <option value="light">라이트</option>
          <option value="dark">다크</option>
        </select>
      </label>
      <label className={rowClass}>
        <span className={labelClass}>본문 폰트 크기</span>
        <input
          type="number"
          min={12}
          max={24}
          className={numberInputClass}
          value={local.font_size}
          onChange={(e) => setLocal({ ...local, font_size: Number(e.target.value) })}
        />
      </label>
      <label className={rowClass}>
        <span className={labelClass}>기본 BPM</span>
        <input
          type="number"
          min={40}
          max={200}
          className={numberInputClass}
          value={local.default_bpm}
          onChange={(e) => setLocal({ ...local, default_bpm: Number(e.target.value) })}
        />
      </label>
      <label className={rowClass}>
        <span className={labelClass}>Display 레이아웃</span>
        <select
          className={inputClass}
          value={local.display_layout}
          onChange={(e) => setLocal({ ...local, display_layout: e.target.value })}
        >
          <option value="default">기본</option>
          <option value="wide">와이드</option>
          <option value="compact">컴팩트</option>
        </select>
      </label>

      <div className="h-px bg-[var(--border)] my-1" />

      <div className={rowClass}>
        <span className="flex items-center gap-1.5 text-sm font-medium text-[var(--text-primary)]">
          스트리밍 스케줄
          {!hasScheduler && (
            <span className="text-[10px] font-bold px-1.5 py-0.5 rounded-lg bg-[#7c3aed] text-white tracking-[0.3px]">
              Pro
            </span>
          )}
        </span>
        <button
          className={`px-4 py-1.5 text-xs font-medium bg-[var(--surface-hover)] border border-[var(--border-input)] rounded-lg cursor-pointer text-[var(--text-primary)] hover:bg-[var(--border)] ${
            !hasScheduler ? "opacity-45 cursor-not-allowed" : ""
          }`}
          onClick={onScheduleOpen}
          disabled={!hasScheduler}
        >
          설정
        </button>
      </div>
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
      className={`relative border-2 rounded-lg cursor-pointer flex items-center justify-center overflow-hidden transition-all ${
        dragging
          ? "border-[#3b82f6] bg-[#eff6ff] border-solid"
          : imageUrl
          ? "border-[var(--border)] border-solid"
          : "border-[var(--border-input)] border-dashed bg-[var(--surface-input)] hover:border-[var(--text-muted)]"
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
        <span className="text-xs text-[var(--text-muted)] pointer-events-none text-center px-3">
          업로드 중...
        </span>
      ) : imageUrl ? (
        <>
          <img src={imageUrl} alt="" className="w-full h-full object-cover" />
          {onClear && (
            <button
              className="absolute top-1 right-1 w-[22px] h-[22px] rounded-full bg-black/55 text-white border-none text-sm leading-none cursor-pointer flex items-center justify-center hover:bg-[rgba(239,68,68,0.8)] transition-colors"
              onClick={(e) => { e.stopPropagation(); onClear(); }}
              title="기본 배경으로"
            >
              &times;
            </button>
          )}
        </>
      ) : (
        <span className="text-xs text-[var(--text-muted)] pointer-events-none text-center px-3">
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

  const inputClass = "px-2 py-1 border border-[var(--border-input)] rounded-md text-xs bg-[var(--surface-elevated)]";

  return (
    <>
      <div className="text-xs font-semibold text-[var(--text-primary)]">기본 배경 이미지</div>
      <div className="grid grid-cols-2 gap-2">
        {Object.entries(WORSHIP_LABELS).map(([type, label]) => {
          const theme = thumbConfig.defaults[type];
          const hasBg = !!theme?.background;
          return (
            <div key={type} className="flex flex-col gap-1">
              <div className="flex items-center gap-1.5">
                <span className="text-xs font-medium text-[var(--text-primary)]">{label}</span>
                <span className={`text-[10px] font-semibold px-1.5 py-0.5 rounded-lg ${
                  hasBg ? "bg-[#d1fae5] text-[#059669]" : "bg-[#fef3c7] text-[#d97706]"
                }`}>
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

      <div className="h-px bg-[var(--border)] my-1" />

      <div className="text-xs font-semibold text-[var(--text-primary)]">기념 주일</div>
      {thumbConfig.specials.length === 0 && (
        <div className="text-xs text-[var(--text-muted)] py-1">등록된 기념 주일이 없습니다</div>
      )}
      {thumbConfig.specials.map((s, i) => (
        <div key={i} className="flex gap-2.5 px-3 py-2.5 bg-[var(--surface-input)] rounded-lg items-stretch">
          <div className="w-[140px] flex-shrink-0">
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
          <div className="flex-1 flex flex-col justify-center gap-1">
            <div className="flex items-center gap-1.5">
              <span className="text-sm font-medium text-[var(--text-primary)]">{s.label}</span>
              <span className={`text-[10px] font-semibold px-1.5 py-0.5 rounded-lg ${
                s.background ? "bg-[#dbeafe] text-[#2563eb]" : "bg-[var(--surface-hover)] text-[var(--text-secondary)]"
              }`}>
                {s.background ? "커스텀" : "기본 배경"}
              </span>
            </div>
            <span className="text-xs text-[var(--text-secondary)]">{s.date}</span>
          </div>
          <button
            onClick={() => removeSpecial(i)}
            className="bg-transparent border-none text-lg text-[var(--text-muted)] cursor-pointer self-start leading-none p-0 hover:text-[#ef4444] transition-colors"
          >
            &times;
          </button>
        </div>
      ))}

      <div className="h-px bg-[var(--border)] my-0.5" />
      <div className="text-xs font-semibold text-[var(--text-primary)]">추가</div>
      <div className="flex gap-2.5 items-stretch">
        <div className="w-[140px] flex-shrink-0">
          <ImageDropZone
            imageUrl={newBgPreview || undefined}
            loading={uploading === "new"}
            onFile={handleNewBgFile}
            onClear={newBgPreview ? () => {
              setNewBgFile(null); URL.revokeObjectURL(newBgPreview); setNewBgPreview(null);
            } : undefined}
            height={80}
            placeholder="배경 드래그 (선택)"
          />
        </div>
        <div className="flex-1 flex flex-col gap-1.5 justify-center">
          <input
            type="date"
            value={newDate}
            onChange={(e) => setNewDate(e.target.value)}
            className={inputClass}
          />
          <div className="flex gap-1.5">
            <input
              type="text"
              placeholder="이름 (예: 부활절 예배)"
              value={newLabel}
              onChange={(e) => setNewLabel(e.target.value)}
              className={`${inputClass} flex-1`}
            />
            <button
              onClick={addSpecial}
              disabled={!newDate || !newLabel || uploading === "new"}
              className={`px-3.5 py-1 text-xs font-semibold bg-[var(--accent)] text-[var(--surface-elevated)] border-none rounded-md cursor-pointer whitespace-nowrap transition-opacity ${
                !newDate || !newLabel ? "opacity-50" : ""
              }`}
            >
              {uploading === "new" ? "..." : "추가"}
            </button>
          </div>
        </div>
      </div>

      <div className="h-px bg-[var(--border)] my-1" />
      <div className="text-xs font-semibold text-[var(--text-primary)]">썸네일 미리보기</div>
      <div className="flex gap-2 items-center flex-wrap">
        <select
          value={previewType}
          onChange={(e) => setPreviewType(e.target.value)}
          className="px-2.5 py-1.5 border border-[var(--border-input)] rounded-md text-xs bg-[var(--surface-input)]"
        >
          {Object.entries(WORSHIP_LABELS).map(([k, v]) => (
            <option key={k} value={k}>{v}</option>
          ))}
        </select>
        <input
          type="date"
          value={previewDate}
          onChange={(e) => setPreviewDate(e.target.value)}
          className="px-2 py-1 border border-[var(--border-input)] rounded-md text-xs bg-[var(--surface-input)]"
        />
        <button
          onClick={handlePreview}
          disabled={generating}
          className="px-3.5 py-1.5 text-xs font-semibold bg-[var(--success)] text-white border-none rounded-md cursor-pointer"
        >
          {generating ? "생성 중..." : "생성"}
        </button>
      </div>
      {previewSpecial && (
        <div className="text-xs text-[#2563eb] bg-[#eff6ff] px-2.5 py-1.5 rounded-md font-medium">
          기념주일 적용: {previewSpecial.label}{previewSpecial.background ? " (커스텀 배경)" : " (기본 배경)"}
        </div>
      )}
      {previewKey > 0 && (
        <img
          key={previewKey}
          src={apiClient.getThumbnailPreviewUrl(previewType, previewDate)}
          alt="썸네일 미리보기"
          className="w-full rounded-lg border border-[var(--border)]"
        />
      )}
    </>
  );
}

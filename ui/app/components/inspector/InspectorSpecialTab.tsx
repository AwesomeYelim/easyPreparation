"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { apiClient } from "@/lib/apiClient";
import { ThumbnailConfig } from "@/types";
import { useAutoSave } from "./useAutoSave";
import DarkImageDropZone from "./DarkImageDropZone";

const WORSHIP_LABELS: Record<string, string> = {
  main_worship: "주일예배",
  after_worship: "오후예배",
  wed_worship: "수요예배",
  fri_worship: "금요예배",
};

const BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ||
  (typeof window !== "undefined" ? window.location.origin : "http://localhost:8080");

/* ── 생성된 썸네일 섹션 ── */
function GeneratedThumbnailSection({
  onReuse,
}: {
  onReuse?: (worshipType: string, date: string) => void;
}) {
  const [list, setList] = useState<
    Array<{
      filename: string;
      label: string;
      date: string;
      worshipType: string;
      url: string;
    }>
  >([]);
  const [loading, setLoading] = useState(false);

  const refresh = () => {
    setLoading(true);
    fetch(`${BASE_URL}/api/thumbnail/generated`)
      .then((r) => r.json())
      .then(setList)
      .catch(console.error)
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    refresh();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const handleDelete = async (filename: string) => {
    await fetch(
      `${BASE_URL}/api/thumbnail/generated?filename=${encodeURIComponent(filename)}`,
      { method: "DELETE" }
    );
    refresh();
  };

  if (list.length === 0 && !loading) return null;

  return (
    <>
      <div className="h-px bg-white/10" />
      <div className="flex items-center justify-between">
        <div className="text-[10px] font-semibold text-[#ccc]">생성된 썸네일</div>
        <button onClick={refresh} className="text-[9px] text-[#666] hover:text-[#aaa]">
          새로고침
        </button>
      </div>
      {loading && <div className="text-[10px] text-[#666]">로딩 중...</div>}
      {list.map((item) => (
        <div
          key={item.filename}
          className="flex items-center gap-2 px-2 py-1.5 bg-white/5 rounded border border-white/10 hover:border-white/20 transition-colors"
        >
          {/* 썸네일 미리보기 이미지 */}
          <a href={item.url} target="_blank" rel="noreferrer" className="flex-shrink-0" title="원본 보기">
            <img
              src={item.url}
              alt={item.label}
              className="w-16 h-9 object-cover rounded border border-white/10 hover:border-electric-blue/50 transition-colors"
            />
          </a>
          <div className="flex-1 min-w-0">
            <div className="text-[10px] font-medium text-white truncate">{item.label}</div>
          </div>
          {onReuse && (
            <button
              onClick={() => onReuse(item.worshipType, item.date)}
              className="text-[#666] hover:text-electric-blue transition-colors text-[10px] flex-shrink-0 border-none bg-transparent cursor-pointer px-1"
              title="이 설정으로 재생성"
            >
              ↺
            </button>
          )}
          <button
            onClick={() => handleDelete(item.filename)}
            className="text-[#666] hover:text-[#ff6b6b] transition-colors text-[10px] flex-shrink-0 border-none bg-transparent cursor-pointer px-1"
            title="삭제"
          >
            ×
          </button>
        </div>
      ))}
    </>
  );
}

/* ── 특별일 설정 섹션 ── */
export default function InspectorSpecialTab() {
  const [thumbConfig, setThumbConfig] = useState<ThumbnailConfig | null>(null);

  // 썸네일 로고 설정 (thumbConfig에서 직접 읽음)
  const thumbLogoPos = thumbConfig?.logoPosition ?? "bottom-right";
  const thumbLogoSize = thumbConfig?.logoSizePercent ?? 0;

  useEffect(() => {
    apiClient.getThumbnailConfig().then(setThumbConfig).catch(console.error);
  }, []);

  useAutoSave(
    thumbConfig,
    (c) => apiClient.saveThumbnailConfig(c).catch(console.error)
  );

  const [newDate, setNewDate] = useState("");
  const [newLabel, setNewLabel] = useState("");
  const [newBgFile, setNewBgFile] = useState<File | null>(null);
  const [newBgPreview, setNewBgPreview] = useState<string | null>(null);
  const [uploading, setUploading] = useState<string | null>(null);
  const [previewSelectVal, setPreviewSelectVal] = useState("main_worship");
  const [previewType, setPreviewType] = useState("main_worship");
  const [previewDate, setPreviewDate] = useState(() => new Date().toISOString().slice(0, 10));
  const [previewKey, setPreviewKey] = useState(0);
  const [generating, setGenerating] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const handleSelectChange = (val: string) => {
    setPreviewSelectVal(val);
    if (val.startsWith("special:")) {
      setPreviewDate(val.slice(8));
      setPreviewType("main_worship");
    } else {
      setPreviewType(val);
    }
  };

  // select/date 변경 시 자동 생성 (debounce 400ms)
  useEffect(() => {
    if (!previewDate) return;
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(async () => {
      setGenerating(true);
      try {
        await apiClient.generateThumbnail(previewType, previewDate);
        setPreviewKey((k) => k + 1);
      } catch (e) {
        console.error(e);
      } finally {
        setGenerating(false);
      }
    }, 400);
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, [previewType, previewDate]); // eslint-disable-line react-hooks/exhaustive-deps
  const [bgVersions, setBgVersions] = useState<Record<string, number>>({});

  const previewSpecial = thumbConfig?.specials.find((s) => s.date === previewDate);

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
      } catch (e) {
        console.error("배경 업로드 실패:", e);
      } finally {
        setUploading(null);
      }
    }
    const finalBgPath = bgPath,
      finalDate = newDate,
      finalLabel = newLabel;
    setThumbConfig((prev) =>
      prev
        ? {
            ...prev,
            specials: [
              ...prev.specials,
              {
                date: finalDate,
                label: finalLabel,
                background: finalBgPath,
                titleOverride: finalLabel,
              },
            ],
          }
        : prev
    );
    setNewDate("");
    setNewLabel("");
    setNewBgFile(null);
    if (newBgPreview) {
      URL.revokeObjectURL(newBgPreview);
      setNewBgPreview(null);
    }
  };

  const handleSpecialBgUpload = async (idx: number, file: File) => {
    setUploading(`special_${idx}`);
    try {
      const label = thumbConfig?.specials[idx]?.label || "special";
      const res = await apiClient.uploadThumbnailBg(
        file,
        label.replace(/\s+/g, "_").toLowerCase()
      );
      if (res.ok) {
        setThumbConfig((prev) => {
          if (!prev) return prev;
          const updated = [...prev.specials];
          updated[idx] = { ...updated[idx], background: res.path };
          return { ...prev, specials: updated };
        });
        setBgVersions((prev) => ({
          ...prev,
          [`special_${idx}`]: (prev[`special_${idx}`] || 0) + 1,
        }));
      }
    } catch (e) {
      console.error("배경 업로드 실패:", e);
    } finally {
      setUploading(null);
    }
  };

  const handleDefaultBgUpload = async (worshipType: string, file: File) => {
    setUploading(`default_${worshipType}`);
    try {
      const res = await apiClient.uploadThumbnailBg(file, `default_${worshipType}`);
      if (res.ok) {
        setThumbConfig((prev) =>
          prev
            ? {
                ...prev,
                defaults: {
                  ...prev.defaults,
                  [worshipType]: { ...prev.defaults[worshipType], background: res.path },
                },
              }
            : prev
        );
        setBgVersions((prev) => ({
          ...prev,
          [worshipType]: (prev[worshipType] || 0) + 1,
        }));
      }
    } catch (e) {
      console.error("기본 배경 업로드 실패:", e);
    } finally {
      setUploading(null);
    }
  };

  const removeSpecial = (idx: number) => {
    setThumbConfig((prev) =>
      prev ? { ...prev, specials: prev.specials.filter((_, i) => i !== idx) } : prev
    );
  };

  if (!thumbConfig) {
    return <div className="text-[#888] text-center py-5 text-xs">로딩 중...</div>;
  }

  const inputClass =
    "px-2 py-1 border border-white/20 rounded-md text-[10px] bg-white/10 text-white outline-none placeholder:text-[#666]";

  return (
    <div className="flex flex-col gap-3">
      {/* 폰트 선택 */}
      <div className="text-[10px] font-semibold text-[#ccc]">썸네일 폰트</div>
      <select
        value={(thumbConfig as any).fontName || "NanumBrush"}
        onChange={(e) =>
          setThumbConfig((prev) => (prev ? { ...(prev as any), fontName: e.target.value } : prev))
        }
        className="w-full px-2 py-1 border border-white/20 rounded-md text-[10px] bg-white/10 text-white outline-none"
      >
        <option value="NanumBrush">나눔손글씨 붓 (기본)</option>
        <option value="NanumGothic">나눔고딕</option>
        <option value="NanumGothicBold">나눔고딕 Bold</option>
        <option value="JacquesFrancois">Jacques François</option>
      </select>

      {/* 썸네일 로고 */}
      <div className="mt-3">
        <div className="text-[10px] font-semibold text-pro-text mb-1">썸네일 로고</div>
        <div className="flex items-center gap-2 mb-1">
          <span className="text-[10px] text-pro-text-dim w-10">위치</span>
          <div className="grid grid-cols-2 gap-1 flex-1">
            {[
              { key: "top-left", label: "↖ 좌상" },
              { key: "top-right", label: "↗ 우상" },
              { key: "bottom-left", label: "↙ 좌하" },
              { key: "bottom-right", label: "↘ 우하" },
            ].map((p) => (
              <button
                key={p.key}
                onClick={() =>
                  setThumbConfig((prev) =>
                    prev ? { ...prev, logoPosition: p.key } : prev
                  )
                }
                className={`py-1 rounded text-[10px] cursor-pointer border transition-all ${
                  thumbLogoPos === p.key
                    ? "bg-[rgba(74,158,255,0.2)] border-[#4a9eff] text-[#4a9eff]"
                    : "bg-white/[0.06] border-white/15 text-[#aaa] hover:border-white/30"
                }`}
              >
                {p.label}
              </button>
            ))}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-[10px] text-pro-text-dim w-10">
            크기: {thumbLogoSize === 0 ? "없음" : `${Math.round(thumbLogoSize)}%`}
          </span>
          <input
            type="range"
            min={0}
            max={30}
            step={1}
            value={thumbLogoSize}
            onChange={(e) =>
              setThumbConfig((prev) =>
                prev ? { ...prev, logoSizePercent: Number(e.target.value) } : prev
              )
            }
            className="flex-1 accent-[#4a9eff]"
          />
        </div>
      </div>

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
                <span
                  className={`text-[9px] font-semibold px-1 py-0.5 rounded-lg ${
                    hasBg ? "bg-[#052e16] text-[#4ade80]" : "bg-[#451a03] text-[#fb923c]"
                  }`}
                >
                  {hasBg ? "설정됨" : "미설정"}
                </span>
              </div>
              <DarkImageDropZone
                imageUrl={
                  hasBg
                    ? apiClient.getThumbnailImageUrl(theme.background) +
                      `&v=${bgVersions[type] || 0}`
                    : undefined
                }
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
        <div
          key={i}
          className="flex gap-2 px-2 py-2 bg-white/5 rounded-lg items-stretch border border-white/10"
        >
          <div className="w-[100px] flex-shrink-0">
            <DarkImageDropZone
              imageUrl={
                s.background
                  ? apiClient.getThumbnailImageUrl(s.background) +
                    `&v=${bgVersions["special_" + i] || 0}`
                  : undefined
              }
              loading={uploading === `special_${i}`}
              onFile={(f) => handleSpecialBgUpload(i, f)}
              onClear={
                s.background
                  ? () => {
                      setThumbConfig((prev) => {
                        if (!prev) return prev;
                        const updated = [...prev.specials];
                        updated[i] = { ...updated[i], background: "" };
                        return { ...prev, specials: updated };
                      });
                    }
                  : undefined
              }
              height={70}
              placeholder="배경 드래그"
            />
          </div>
          <div className="flex-1 flex flex-col justify-center gap-1">
            <div className="flex items-center gap-1">
              <span className="text-xs font-medium text-white">{s.label}</span>
              <span
                className={`text-[9px] font-semibold px-1 py-0.5 rounded-lg ${
                  s.background ? "bg-[#1e3a5f] text-[#60a5fa]" : "bg-white/10 text-[#aaa]"
                }`}
              >
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
              onClear={
                newBgPreview
                  ? () => {
                      setNewBgFile(null);
                      URL.revokeObjectURL(newBgPreview);
                      setNewBgPreview(null);
                    }
                  : undefined
              }
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
      <div className="flex gap-1.5 items-center">
        <select
          value={previewSelectVal}
          onChange={(e) => handleSelectChange(e.target.value)}
          className="flex-1 h-7 px-2 border border-white/20 rounded-md text-[10px] bg-white/10 text-white outline-none"
        >
          {Object.entries(WORSHIP_LABELS).map(([k, v]) => (
            <option key={k} value={k} className="bg-[#2c2c2c]">{v}</option>
          ))}
          {thumbConfig.specials.length > 0 && (
            <optgroup label="── 기념 주일 ──" className="bg-[#2c2c2c]">
              {thumbConfig.specials.map((s) => (
                <option key={s.date} value={`special:${s.date}`} className="bg-[#2c2c2c]">
                  {s.label} ({s.date.slice(2).replace(/-/g, ".")})
                </option>
              ))}
            </optgroup>
          )}
        </select>
        <input
          type="date"
          value={previewDate}
          onChange={(e) => setPreviewDate(e.target.value)}
          className="h-7 px-2 border border-white/20 rounded-md text-[10px] bg-white/10 text-white outline-none"
        />
      </div>
      {previewSpecial && (
        <div className="text-[10px] text-[#60a5fa] bg-[#1e3a5f] px-2 py-1.5 rounded-md font-medium">
          기념주일 적용: {previewSpecial.label}
          {previewSpecial.background ? " (커스텀 배경)" : " (기본 배경)"}
        </div>
      )}
      <div className="relative w-full">
        {previewKey > 0 && (
          <img
            key={previewKey}
            src={apiClient.getThumbnailPreviewUrl(previewType, previewDate)}
            alt="썸네일 미리보기"
            className="w-full rounded-lg border border-white/20"
          />
        )}
        {generating && (
          <div className={`${previewKey > 0 ? "absolute inset-0" : "w-full aspect-video"} flex items-center justify-center rounded-lg bg-black/50`}>
            <div className="w-5 h-5 border-2 border-white/20 border-t-white rounded-full animate-spin" />
          </div>
        )}
      </div>

      <GeneratedThumbnailSection
        onReuse={(type, date) => {
          setPreviewType(type);
          setPreviewDate(date);
        }}
      />
    </div>
  );
}

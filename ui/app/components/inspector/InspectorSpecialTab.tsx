"use client";

import { useState, useEffect, useCallback, useRef, useMemo } from "react";
import { useRecoilValue } from "recoil";
import { apiClient } from "@/lib/apiClient";
import { ThumbnailConfig, TextStyles, TextStyle, WorshipOrderItem } from "@/types";
import { worshipOrderState, WorshipType } from "@/recoilState";
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

/** Go 폰트명 -> CSS font-family */
const FONT_CSS_MAP: Record<string, { family: string; weight?: number }> = {
  NanumBrush: { family: "'NanumBrush', cursive" },
  NanumGothic: { family: "'NanumGothic', sans-serif" },
  NanumGothicBold: { family: "'NanumGothicBold', sans-serif", weight: 800 },
  JacquesFrancois: { family: "'JacquesFrancois', serif" },
};

const DEFAULT_TEXT_STYLES: TextStyles = {
  header: { fontName: "NanumGothicBold", size: 50, color: "#ffffff" },
  main: { fontName: "NanumGothicBold", size: 100, color: "#ffffff" },
  footer: { fontName: "NanumGothic", size: 45, color: "#ffffff" },
};

type ActiveArea = "header" | "main" | "footer";

/* ── 생성된 썸네일 섹션 ── */
function GeneratedThumbnailSection({
  onReuse,
  refreshKey,
}: {
  onReuse?: (worshipType: string, date: string) => void;
  refreshKey?: number;
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
  }, [refreshKey]); // eslint-disable-line react-hooks/exhaustive-deps

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

/* ── 캔버스 프리뷰 컴포넌트 ── */
function ThumbnailCanvasPreview({
  bgUrl,
  dateLabel,
  sermonTitle,
  scripture,
  textStyles,
  activeArea,
  onAreaClick,
  onTextEdit,
  logoUrl,
  logoPosition,
  logoSizePercent,
}: {
  bgUrl: string;
  dateLabel: string;
  sermonTitle: string;
  scripture: string;
  textStyles: TextStyles;
  activeArea: ActiveArea | null;
  onAreaClick: (area: ActiveArea) => void;
  onTextEdit: (area: ActiveArea, text: string) => void;
  logoUrl?: string;
  logoPosition?: string;
  logoSizePercent?: number;
}) {
  const [editingArea, setEditingArea] = useState<ActiveArea | null>(null);
  const headerRef = useRef<HTMLSpanElement>(null);
  const mainRef = useRef<HTMLSpanElement>(null);
  const footerRef = useRef<HTMLSpanElement>(null);
  const areaRefs = { header: headerRef, main: mainRef, footer: footerRef };

  const pxToScale = (px: number) => `${(px / 1280) * 100}cqw`;

  const cssForArea = (area: keyof TextStyles) => {
    const style = textStyles[area] ?? DEFAULT_TEXT_STYLES[area];
    const fontMap = FONT_CSS_MAP[style.fontName || "NanumGothicBold"] ?? FONT_CSS_MAP.NanumGothicBold;
    return {
      fontFamily: fontMap.family,
      fontWeight: fontMap.weight ?? "normal",
      fontSize: pxToScale(style.size || DEFAULT_TEXT_STYLES[area].size!),
      color: style.color || "#ffffff",
      textShadow: "1px 1px 3px rgba(0,0,0,0.7), 0 0 8px rgba(0,0,0,0.4)",
    } as React.CSSProperties;
  };

  const borderClass = (area: ActiveArea) =>
    activeArea === area
      ? "outline outline-2 outline-[#4a9eff] outline-offset-2 rounded"
      : "hover:outline hover:outline-1 hover:outline-white/30 hover:outline-offset-2 hover:rounded";

  const mainText = sermonTitle || dateLabel;

  const handleDoubleClick = (area: ActiveArea) => {
    setEditingArea(area);
    onAreaClick(area);
    requestAnimationFrame(() => {
      const el = areaRefs[area].current;
      if (el) {
        el.focus();
        // 전체 텍스트 선택
        const range = document.createRange();
        range.selectNodeContents(el);
        const sel = window.getSelection();
        sel?.removeAllRanges();
        sel?.addRange(range);
      }
    });
  };

  const handleBlur = (area: ActiveArea, e: React.FocusEvent<HTMLSpanElement>) => {
    const text = e.currentTarget.textContent || "";
    onTextEdit(area, text);
    setEditingArea(null);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLSpanElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      (e.target as HTMLElement).blur();
    }
  };

  const editableProps = (area: ActiveArea) =>
    editingArea === area
      ? {
          contentEditable: true as const,
          suppressContentEditableWarning: true as const,
          onBlur: (e: React.FocusEvent<HTMLSpanElement>) => handleBlur(area, e),
          onKeyDown: handleKeyDown,
          style: {
            ...cssForArea(area),
            outline: "none",
            cursor: "text",
            minWidth: "3cqw",
            borderBottom: "1px dashed rgba(74,158,255,0.6)",
          },
        }
      : { style: cssForArea(area) };

  return (
    <div
      className="relative w-full aspect-video bg-cover bg-center rounded-lg overflow-hidden border border-white/20"
      style={{
        backgroundImage: `url(${bgUrl})`,
        containerType: "inline-size",
      }}
    >
      <div className="absolute inset-0 bg-black/30 pointer-events-none" />

      {/* 헤더 (상단 12%) — 구분선 + 텍스트 */}
      <div
        className={`absolute left-0 right-0 cursor-pointer transition-all ${borderClass("header")}`}
        style={{ top: "12%" }}
        onClick={() => onAreaClick("header")}
        onDoubleClick={() => handleDoubleClick("header")}
      >
        <div className="flex items-center gap-[1.5cqw] px-[5%]">
          <div className="flex-1 h-[1px] opacity-40" style={{ background: textStyles.header?.color || "#ffffff" }} />
          <span ref={headerRef} {...editableProps("header")} className="whitespace-nowrap relative z-10 px-[1.5cqw]">
            {dateLabel || "헤더 텍스트"}
          </span>
          <div className="flex-1 h-[1px] opacity-40" style={{ background: textStyles.header?.color || "#ffffff" }} />
        </div>
      </div>

      {/* 메인 (중앙) */}
      <div
        className={`absolute left-0 right-0 flex items-center justify-center cursor-pointer px-4 transition-all ${borderClass("main")}`}
        style={{ top: "50%", transform: "translateY(-50%)" }}
        onClick={() => onAreaClick("main")}
        onDoubleClick={() => handleDoubleClick("main")}
      >
        <span ref={mainRef} {...editableProps("main")} className="text-center relative z-10">
          {mainText || "제목 텍스트"}
        </span>
      </div>

      {/* 푸터 (하단 12%) */}
      <div
        className={`absolute left-0 right-0 flex items-center justify-center cursor-pointer px-4 transition-all ${borderClass("footer")}`}
        style={{ bottom: "12%" }}
        onClick={() => onAreaClick("footer")}
        onDoubleClick={() => handleDoubleClick("footer")}
      >
        <span ref={footerRef} {...editableProps("footer")} className="text-center relative z-10">
          {scripture || "푸터 텍스트"}
        </span>
      </div>

      {/* 로고 */}
      {logoUrl && (() => {
        const isTop = logoPosition?.startsWith("top");
        const isRight = logoPosition?.endsWith("right");
        return (
          <img
            src={logoUrl}
            alt="logo"
            className={`absolute pointer-events-none object-contain opacity-90 ${
              isTop ? "top-[3%]" : "bottom-[3%]"
            } ${
              isRight ? "right-[3%]" : "left-[3%]"
            }`}
            style={{ width: `${logoSizePercent || 12}%` }}
          />
        );
      })()}

      {/* 편집 힌트 */}
      {!editingArea && (
        <div className="absolute bottom-1 left-1/2 -translate-x-1/2 text-[0.7cqw] text-white/40 pointer-events-none">
          클릭: 스타일 편집 · 더블클릭: 텍스트 편집
        </div>
      )}
    </div>
  );
}

/* ── 영역별 설정 패널 ── */
function AreaStyleEditor({
  area,
  label,
  textStyles,
  onUpdateStyle,
}: {
  area: keyof TextStyles;
  label: string;
  textStyles: TextStyles;
  onUpdateStyle: (area: keyof TextStyles, field: keyof TextStyle, value: string | number) => void;
}) {
  const sizeRange = {
    header: { min: 20, max: 80 },
    main: { min: 40, max: 200 },
    footer: { min: 20, max: 80 },
  }[area];

  const fontOptions = [
    { value: "NanumGothicBold", label: "나눔고딕 Bold" },
    { value: "NanumGothic", label: "나눔고딕" },
  ];

  const ts = textStyles[area] ?? DEFAULT_TEXT_STYLES[area];

  return (
    <div className="flex flex-col gap-1.5 px-2 py-2 bg-[rgba(74,158,255,0.08)] rounded-lg border border-[#4a9eff]/30">
      <div className="text-[10px] font-semibold text-[#4a9eff]">{label} 스타일</div>
      <div className="flex items-center gap-1.5">
        <select
          value={ts.fontName || "NanumGothicBold"}
          onChange={(e) => onUpdateStyle(area, "fontName", e.target.value)}
          className="flex-1 px-1.5 py-0.5 border border-white/20 rounded text-[10px] bg-white/10 text-white outline-none"
        >
          {fontOptions.map((f) => (
            <option key={f.value} value={f.value} className="bg-[#2c2c2c]">{f.label}</option>
          ))}
        </select>
        <input
          type="color"
          value={ts.color || "#ffffff"}
          onChange={(e) => onUpdateStyle(area, "color", e.target.value)}
          className="w-6 h-6 p-0 border border-white/20 rounded cursor-pointer bg-transparent"
          title="텍스트 색상"
        />
      </div>
      <div className="flex items-center gap-1.5">
        <span className="text-[9px] text-[#888] w-8 text-right">{Math.round(ts.size || DEFAULT_TEXT_STYLES[area].size!)}px</span>
        <input
          type="range"
          min={sizeRange.min}
          max={sizeRange.max}
          step={1}
          value={ts.size || DEFAULT_TEXT_STYLES[area].size}
          onChange={(e) => onUpdateStyle(area, "size", Number(e.target.value))}
          className="flex-1 accent-[#4a9eff]"
        />
      </div>
    </div>
  );
}

/* ── 특별일 설정 섹션 ── */
export default function InspectorSpecialTab() {
  const [thumbConfig, setThumbConfig] = useState<ThumbnailConfig | null>(null);
  const worshipOrder = useRecoilValue(worshipOrderState);

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
  const [generating, setGenerating] = useState(false);
  const [activeArea, setActiveArea] = useState<ActiveArea | null>(null);
  const [textOverridesByType, setTextOverridesByType] = useState<Record<string, Record<string, string>>>({});
  const textOverrides = textOverridesByType[previewType] || {};
  const setTextOverrides = (updater: (prev: Record<string, string>) => Record<string, string>) => {
    setTextOverridesByType((prev) => ({
      ...prev,
      [previewType]: updater(prev[previewType] || {}),
    }));
  };
  const [genCounter, setGenCounter] = useState(0);
  const [bgVersions, setBgVersions] = useState<Record<string, number>>({});

  const handleSelectChange = (val: string) => {
    setPreviewSelectVal(val);
    setActiveArea(null);
    if (val.startsWith("special:")) {
      setPreviewDate(val.slice(8));
      setPreviewType("main_worship");
    } else {
      setPreviewType(val);
    }
  };

  // 예배 순서에서 말씀 제목 + 성경봉독 추출 (Go 로직 재현)
  const { sermonTitle, scripture } = useMemo(() => {
    const items: WorshipOrderItem[] = worshipOrder[previewType as WorshipType] ?? [];
    let sermon = "";
    let scrip = "";
    let bEditFallback = "";
    for (const item of items) {
      if ((item.title === "말씀" || item.title === "설교") && item.obj && item.obj !== "-") {
        sermon = item.obj;
      }
      if (item.title === "성경봉독" && item.obj && item.obj !== "-") {
        scrip = item.obj;
      }
      if (item.info?.startsWith("b_") && item.obj && item.obj !== "-" && !bEditFallback) {
        bEditFallback = item.obj;
      }
    }
    if (!scrip) scrip = bEditFallback;
    return { sermonTitle: sermon, scripture: scrip };
  }, [worshipOrder, previewType]);

  // dateLabel 빌드 (Go 로직 재현): "26.04.05 주일예배"
  const dateLabel = useMemo(() => {
    if (!previewDate) return "";
    const d = new Date(previewDate + "T00:00:00");
    const yy = String(d.getFullYear()).slice(2);
    const mm = String(d.getMonth() + 1).padStart(2, "0");
    const dd = String(d.getDate()).padStart(2, "0");

    let typeLabel = WORSHIP_LABELS[previewType] || "예배";
    // 기념 주일 오버라이드
    const special = thumbConfig?.specials.find((s) => s.date === previewDate);
    if (special) {
      if (special.titleOverride) typeLabel = special.titleOverride;
      else if (special.label) typeLabel = special.label;
    }
    return `${yy}.${mm}.${dd} ${typeLabel}`;
  }, [previewDate, previewType, thumbConfig?.specials]);

  // 배경 이미지 URL
  const bgUrl = useMemo(() => {
    // 기념 주일 커스텀 배경 확인
    const special = thumbConfig?.specials.find((s) => s.date === previewDate);
    if (special?.background) {
      return apiClient.getThumbnailImageUrl(special.background);
    }
    // 예배 타입별 기본 배경
    const defaultTheme = thumbConfig?.defaults[previewType];
    if (defaultTheme?.background) {
      return apiClient.getThumbnailImageUrl(defaultTheme.background) +
        `&v=${bgVersions[previewType] || 0}`;
    }
    // fallback
    return `${BASE_URL}/display/bg`;
  }, [thumbConfig, previewType, previewDate, bgVersions]);

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

  // 텍스트 스타일 업데이트
  const updateStyle = useCallback(
    (area: keyof TextStyles, field: keyof TextStyle, value: string | number) => {
      setThumbConfig((prev) => {
        if (!prev) return prev;
        const current = prev.textStyles ?? DEFAULT_TEXT_STYLES;
        return {
          ...prev,
          textStyles: {
            ...current,
            [area]: { ...current[area], [field]: value },
          },
        };
      });
    },
    []
  );

  // PNG 생성 (서버 호출) — 캔버스 미리보기에서 편집한 텍스트 오버라이드 전달
  const handleGeneratePNG = async () => {
    if (!previewDate) return;
    setGenerating(true);
    try {
      const overrides = textOverridesByType[previewType];
      await apiClient.generateThumbnail(previewType, previewDate, false, overrides
        ? { header: overrides.header, main: overrides.main, footer: overrides.footer }
        : undefined
      );
      setGenCounter((c) => c + 1);
    } catch (e) {
      console.error(e);
    } finally {
      setGenerating(false);
    }
  };

  if (!thumbConfig) {
    return <div className="text-[#888] text-center py-5 text-xs">로딩 중...</div>;
  }

  const ts: TextStyles = thumbConfig.textStyles ?? DEFAULT_TEXT_STYLES;

  const inputClass =
    "px-2 py-1 border border-white/20 rounded-md text-[10px] bg-white/10 text-white outline-none placeholder:text-[#666]";

  return (
    <div className="flex flex-col gap-3">
      {/* ── 예배 타입 + 날짜 선택 ── */}
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

      {/* ── 캔버스 프리뷰 (최상단) ── */}
      <ThumbnailCanvasPreview
        bgUrl={bgUrl}
        dateLabel={textOverrides.header ?? dateLabel}
        sermonTitle={textOverrides.main ?? sermonTitle}
        scripture={textOverrides.footer ?? scripture}
        textStyles={ts}
        activeArea={activeArea}
        onAreaClick={setActiveArea}
        onTextEdit={(area, text) => setTextOverrides((prev) => ({ ...prev, [area]: text }))}
        logoUrl={`${BASE_URL}/api/logo`}
        logoPosition={thumbLogoPos}
        logoSizePercent={thumbConfig.logoSizePercent ?? 12}
      />

      {/* 선택된 영역 설정 패널 */}
      {activeArea && (
        <AreaStyleEditor
          area={activeArea}
          label={{ header: "헤더", main: "제목", footer: "푸터" }[activeArea]}
          textStyles={ts}
          onUpdateStyle={updateStyle}
        />
      )}

      {/* PNG 생성 버튼 */}
      <div className="flex gap-1.5 items-center">
        <button
          onClick={handleGeneratePNG}
          disabled={generating || !previewDate}
          className={`flex-1 h-7 text-[10px] font-semibold rounded-md border-none cursor-pointer transition-all ${
            generating
              ? "bg-white/10 text-[#666]"
              : "bg-[#4a9eff] text-white hover:bg-[#3b8fe8]"
          }`}
        >
          {generating ? (
            <span className="flex items-center justify-center gap-1.5">
              <span className="w-3 h-3 border border-white/20 border-t-white rounded-full animate-spin" />
              생성 중...
            </span>
          ) : (
            "썸네일 생성"
          )}
        </button>
        {activeArea && (
          <button
            onClick={() => setActiveArea(null)}
            className="h-7 px-3 text-[10px] text-[#888] bg-white/5 border border-white/10 rounded-md cursor-pointer hover:text-white hover:border-white/20 transition-all"
          >
            선택 해제
          </button>
        )}
      </div>

      <div className="h-px bg-white/10" />

      {/* 텍스트 스타일 — 선택 안 했을 때 안내만 */}
      {!activeArea && (
        <div className="text-[10px] text-[#666] text-center py-2">
          캔버스에서 텍스트를 클릭하여 스타일 편집
        </div>
      )}

      {/* 썸네일 로고 */}
      <div className="mt-1">
        <div className="text-[10px] font-semibold text-pro-text mb-1">썸네일 로고</div>
        <div className="flex items-center gap-2 mb-1">
          <span className="text-[10px] text-pro-text-dim w-10">위치</span>
          <div className="grid grid-cols-2 gap-1 flex-1">
            {[
              { key: "top-left", label: "좌상" },
              { key: "top-right", label: "우상" },
              { key: "bottom-left", label: "좌하" },
              { key: "bottom-right", label: "우하" },
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

      <div className="h-px bg-white/10" />

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

      <GeneratedThumbnailSection
        refreshKey={genCounter}
        onReuse={(type, date) => {
          setPreviewType(type);
          setPreviewDate(date);
          setPreviewSelectVal(type);
        }}
      />
    </div>
  );
}

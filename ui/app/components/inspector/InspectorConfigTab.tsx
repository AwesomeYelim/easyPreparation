"use client";

import { useState, useEffect, useRef } from "react";
import { apiClient } from "@/lib/apiClient";

const DISPLAY_FONTS = [
  { value: "default", label: "기본 (맑은 고딕)" },
  { value: "noto-sans-kr", label: "Noto Sans KR" },
  { value: "gowun-dodum", label: "Gowun Dodum" },
  { value: "nanum-myeongjo", label: "나눔명조" },
  { value: "black-han-sans", label: "Black Han Sans" },
];

const LOGO_POSITIONS = [
  { key: "top-left", label: "↖ 좌상" },
  { key: "top-right", label: "↗ 우상" },
  { key: "bottom-left", label: "↙ 좌하" },
  { key: "bottom-right", label: "↘ 우하" },
];

interface Props {
  showToast: (msg: string, type?: "error" | "info") => void;
}

export default function InspectorConfigTab({ showToast }: Props) {
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
    apiClient
      .getDisplayConfig()
      .then((c) => {
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
      })
      .catch(() => {
        loaded.current = true;
      });
    apiClient.hasLogo().then(setLogoExists).catch(() => {});
  }, []);

  // 자동 저장 (600ms 디바운스) — 초기 로드는 skip
  useEffect(() => {
    if (isLoadedUpdate.current) {
      isLoadedUpdate.current = false;
      return;
    }
    if (!loaded.current) return;
    const timer = setTimeout(() => {
      apiClient
        .saveDisplayConfig(config)
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

  const inputCls =
    "px-2 py-1 border border-pro-border rounded-md text-[10px] bg-pro-elevated text-pro-text outline-none";
  const rowCls = "flex justify-between items-center gap-2";
  const labelCls = "text-[11px] font-medium text-pro-text";
  const divider = <div className="h-px bg-pro-border my-1" />;

  return (
    <div className="flex flex-col gap-3">
      {/* 자동 저장 표시 */}
      {savedAt > 0 && (
        <div className="flex items-center gap-1 text-[9px] text-[#4a9eff] opacity-70">
          <span>✓</span>
          <span>자동 저장됨</span>
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
            <button
              className="px-2 py-1 text-[10px] bg-pro-hover border border-pro-border rounded cursor-pointer hover:opacity-80 text-pro-text"
              onClick={() => fileRef.current?.click()}
            >
              교체
            </button>
            <button
              className="px-2 py-1 text-[10px] text-[#ff6b6b] border border-[rgba(255,107,107,0.3)] rounded cursor-pointer hover:opacity-80 bg-transparent"
              onClick={handleLogoDelete}
            >
              삭제
            </button>
          </>
        ) : (
          <button
            className="px-3 py-1.5 text-[10px] bg-[#4a9eff] text-white border-none rounded-md cursor-pointer hover:opacity-90"
            onClick={() => fileRef.current?.click()}
          >
            로고 업로드
          </button>
        )}
        <input
          ref={fileRef}
          type="file"
          accept="image/png,image/jpeg"
          className="hidden"
          onChange={handleLogoUpload}
        />
      </div>

      {/* 로고 위치 */}
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
              <button
                key={p.key}
                onClick={() => updateConfig({ logoPosition: p.key })}
                className={`py-1 rounded text-[10px] cursor-pointer border transition-all ${
                  config.logoPosition === p.key
                    ? "bg-[rgba(74,158,255,0.2)] border-[#4a9eff] text-[#4a9eff]"
                    : "bg-white/[0.06] border-white/15 text-[#aaa] hover:border-white/30"
                }`}
              >
                {p.label}
              </button>
            ))}
          </div>
          <div className={rowCls}>
            <span className={labelCls}>크기: {Math.round(config.logoSizePercent)}%</span>
            <input
              type="range"
              min={5}
              max={30}
              step={1}
              value={config.logoSizePercent}
              onChange={(e) => updateConfig({ logoSizePercent: Number(e.target.value) })}
              className="w-24 accent-[#4a9eff]"
            />
          </div>
        </>
      )}

      {divider}

      {/* Display 폰트 */}
      <div className={rowCls}>
        <span className={labelCls}>Display 폰트</span>
        <select
          className={inputCls}
          value={config.font}
          onChange={(e) => updateConfig({ font: e.target.value })}
        >
          {DISPLAY_FONTS.map((f) => (
            <option key={f.value} value={f.value} className="bg-[#1a1a1a]">
              {f.label}
            </option>
          ))}
        </select>
      </div>

      {divider}
      <div className="text-[10px] font-semibold text-[#ccc]">오버레이 커스터마이징</div>

      {/* 배경 투명도 */}
      <div className={rowCls}>
        <span className={labelCls}>배경 투명도</span>
        <div className="flex items-center gap-2">
          <input
            type="range"
            min={0}
            max={1}
            step={0.05}
            value={config.overlayBgOpacity}
            onChange={(e) => updateConfig({ overlayBgOpacity: Number(e.target.value) })}
            className="w-24 accent-[#4a9eff]"
          />
          <span className="text-[10px] text-[#888] w-8 text-right">
            {Math.round(config.overlayBgOpacity * 100)}%
          </span>
        </div>
      </div>

      {/* 텍스트 색상 */}
      <div className={rowCls}>
        <span className={labelCls}>텍스트 색상</span>
        <div className="flex items-center gap-2">
          <input
            type="color"
            value={config.overlayTextColor}
            onChange={(e) => updateConfig({ overlayTextColor: e.target.value })}
            className="w-7 h-7 rounded cursor-pointer border border-pro-border bg-transparent p-0.5"
          />
          <span className="text-[10px] text-[#888] font-mono">{config.overlayTextColor}</span>
        </div>
      </div>

      {/* 텍스트 위치 */}
      <div className={rowCls}>
        <span className={labelCls}>텍스트 위치</span>
        <select
          className={inputCls}
          value={config.overlayPosition}
          onChange={(e) => updateConfig({ overlayPosition: e.target.value })}
        >
          <option value="flex-end" className="bg-[#1a1a1a]">
            하단
          </option>
          <option value="center" className="bg-[#1a1a1a]">
            중앙
          </option>
          <option value="flex-start" className="bg-[#1a1a1a]">
            상단
          </option>
        </select>
      </div>

      {/* 폰트 배율 */}
      <div className={rowCls}>
        <span className={labelCls}>폰트 배율</span>
        <div className="flex items-center gap-2">
          <input
            type="range"
            min={0.5}
            max={2.0}
            step={0.1}
            value={config.overlayFontScale}
            onChange={(e) => updateConfig({ overlayFontScale: Number(e.target.value) })}
            className="w-24 accent-[#4a9eff]"
          />
          <span className="text-[10px] text-[#888] w-8 text-right">
            {config.overlayFontScale.toFixed(1)}×
          </span>
        </div>
      </div>

    </div>
  );
}

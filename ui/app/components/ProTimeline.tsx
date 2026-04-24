"use client";
import { useEffect, useState, useRef, useCallback } from "react";
import { useRecoilState, useRecoilValue } from "recoil";
import {
  displayItemsState,
  serviceStartTimeState,
  itemTimersState,
  autoAdvanceState,
} from "@/recoilState";
import { useWS } from "@/components/WebSocketProvider";
import { apiClient } from "@/lib/apiClient";

const DEFAULT_SECS = 300; // 5분 기본 (타이머 미설정 시)

const ACTIVE_COLOR = "#4a9eff";

type DragState = {
  startX: number;
  leftKey: string;
  rightKey: string;
  leftInitial: number;
  rightInitial: number;
  pairTotal: number;
  secsPerPx: number;
};

function formatClock(ms: number) {
  return new Date(ms).toLocaleTimeString("ko-KR", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  });
}

function formatDuration(secs: number) {
  const h = Math.floor(secs / 3600);
  const m = Math.floor((secs % 3600) / 60);
  if (h > 0 && m > 0) return `${h}h ${m}m`;
  if (h > 0) return `${h}h`;
  return `${m}m`;
}

// 타임라인 타일 내 짧은 표시: 1m30s → "1:30", 45s → "45s"
function formatDurationShort(secs: number) {
  const m = Math.floor(secs / 60);
  const s = secs % 60;
  if (m === 0) return `${s}s`;
  return `${m}:${String(s).padStart(2, "0")}`;
}

export default function ProTimeline() {
  const items = useRecoilValue(displayItemsState);
  const [serviceStart, setServiceStart] = useRecoilState(serviceStartTimeState);
  const [itemTimers, setItemTimers] = useRecoilState(itemTimersState);
  const [autoEnabled, setAutoEnabled] = useRecoilState(autoAdvanceState);
  const [currentIdx, setCurrentIdx] = useState(0);
  const [elapsed, setElapsed] = useState("00:00");
  const [autoCountdown, setAutoCountdown] = useState(0);
  const [hoverDivider, setHoverDivider] = useState<number | null>(null);
  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [editValue, setEditValue] = useState("");
  const [editingTotal, setEditingTotal] = useState(false);
  const [totalEditValue, setTotalEditValue] = useState("");

  const autoIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const dragRef = useRef<DragState | null>(null);
  const { subscribe } = useWS();

  // ── WS: 현재 위치 추적 ──
  useEffect(() => {
    return subscribe((msg: any) => {
      if (msg.type === "position" && typeof msg.idx === "number") {
        setCurrentIdx(msg.idx);
      }
    });
  }, [subscribe]);

  // ── 경과 시간 ──
  useEffect(() => {
    const update = () => {
      if (!serviceStart) { setElapsed("00:00"); return; }
      const diff = Math.floor((Date.now() - serviceStart) / 1000);
      const m = Math.floor(diff / 60).toString().padStart(2, "0");
      const s = (diff % 60).toString().padStart(2, "0");
      setElapsed(`${m}:${s}`);
    };
    update();
    const id = setInterval(update, 1000);
    return () => clearInterval(id);
  }, [serviceStart]);

  // ── 유틸 ──
  const getEffectiveSecs = useCallback(
    (key: string) => itemTimers[key] || DEFAULT_SECS,
    [itemTimers]
  );

  // ── 자동 진행 ──
  useEffect(() => {
    if (autoIntervalRef.current) { clearInterval(autoIntervalRef.current); autoIntervalRef.current = null; }
    if (!autoEnabled) { setAutoCountdown(0); return; }
    // 마지막 항목이면 자동 진행 종료
    if (currentIdx >= items.length - 1 && items.length > 0) { setAutoCountdown(0); return; }
    const item = items[currentIdx];
    if (!item) return;
    const secs = getEffectiveSecs(item.key);
    if (secs <= 0) { setAutoCountdown(0); return; }
    setAutoCountdown(secs);
    autoIntervalRef.current = setInterval(() => {
      setAutoCountdown((prev) => {
        if (prev <= 1) {
          if (autoIntervalRef.current) { clearInterval(autoIntervalRef.current); autoIntervalRef.current = null; }
          apiClient.navigateDisplay("next");
          // display 페이지가 열려있지 않아도 루프 유지: 낙관적 idx 증가
          // (display가 열려있으면 WS "position" 에코가 같은 값으로 덮어씀 — 무해)
          setCurrentIdx((ci) => Math.min(ci + 1, items.length - 1));
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
    return () => { if (autoIntervalRef.current) { clearInterval(autoIntervalRef.current); autoIntervalRef.current = null; } };
  }, [autoEnabled, currentIdx, items, itemTimers, getEffectiveSecs]);

  const totalSecs = items.reduce((sum, item) => sum + getEffectiveSecs(item.key), 0);
  // serviceStart가 없을 때 Date.now() 사용 금지 → 하이드레이션 불일치 방지
  const startMs = serviceStart ?? null;
  const endMs = startMs !== null ? startMs + totalSecs * 1000 : null;

  // ── 드래그 핸들러 ──
  const handleDividerMouseDown = useCallback(
    (e: React.MouseEvent, dividerIdx: number) => {
      e.preventDefault();
      e.stopPropagation();
      const leftItem = items[dividerIdx];
      const rightItem = items[dividerIdx + 1];
      if (!leftItem || !rightItem) return;
      const leftSecs = getEffectiveSecs(leftItem.key);
      const rightSecs = getEffectiveSecs(rightItem.key);
      const snapTotal = items.reduce((s, it) => s + getEffectiveSecs(it.key), 0);
      const containerWidth = containerRef.current?.getBoundingClientRect().width ?? 800;

      dragRef.current = {
        startX: e.clientX,
        leftKey: leftItem.key,
        rightKey: rightItem.key,
        leftInitial: leftSecs,
        rightInitial: rightSecs,
        pairTotal: leftSecs + rightSecs,
        secsPerPx: snapTotal / containerWidth,
      };

      const handleMouseMove = (ev: MouseEvent) => {
        if (!dragRef.current) return;
        const dx = ev.clientX - dragRef.current.startX;
        const deltaSecs = Math.round(dx * dragRef.current.secsPerPx);
        const newLeft = Math.max(
          30,
          Math.min(dragRef.current.pairTotal - 30, dragRef.current.leftInitial + deltaSecs)
        );
        const newRight = dragRef.current.pairTotal - newLeft;
        setItemTimers((prev) => ({
          ...prev,
          [dragRef.current!.leftKey]: newLeft,
          [dragRef.current!.rightKey]: newRight,
        }));
      };

      const handleMouseUp = () => {
        dragRef.current = null;
        document.removeEventListener("mousemove", handleMouseMove);
        document.removeEventListener("mouseup", handleMouseUp);
      };

      document.addEventListener("mousemove", handleMouseMove);
      document.addEventListener("mouseup", handleMouseUp);
    },
    [items, getEffectiveSecs, setItemTimers]
  );

  const commitEdit = useCallback((key: string, raw: string) => {
    setEditingKey(null);
    const s = parseInt(raw.replace(/[^0-9]/g, ""), 10);
    if (!isNaN(s) && s >= 30) {
      setItemTimers((prev) => ({ ...prev, [key]: s }));
    }
  }, [setItemTimers]);

  const commitTotalEdit = useCallback((raw: string) => {
    setEditingTotal(false);
    const mins = parseFloat(raw.replace(/[^0-9.]/g, ""));
    if (!isNaN(mins) && mins > 0 && items.length > 0 && totalSecs > 0) {
      const newTotalSecs = Math.round(mins * 60);
      const newTimers: Record<string, number> = {};
      for (const item of items) {
        const cur = getEffectiveSecs(item.key);
        newTimers[item.key] = Math.max(30, Math.round((cur / totalSecs) * newTotalSecs));
      }
      setItemTimers((prev) => ({ ...prev, ...newTimers }));
    }
  }, [items, totalSecs, getEffectiveSecs, setItemTimers]);

  return (
    <div
      className="flex flex-col bg-[#0e0e0e] border-t border-pro-border select-none overflow-hidden"
      style={{ gridColumn: "1 / -1", gridRow: "3" }}
    >
      {/* ── 헤더 ── */}
      <div className="flex items-center justify-between px-2 h-[22px] flex-shrink-0 border-b border-white/[0.06]">
        <div className="flex items-center gap-1.5">
          {/* 재생 버튼 */}
          <button
            className="w-5 h-5 flex items-center justify-center rounded hover:bg-white/10 text-[#666] hover:text-white transition-colors"
            onClick={() => apiClient.navigateDisplay("prev")}
            title="이전"
          >
            <span className="material-symbols-outlined" style={{ fontSize: "13px" }}>skip_previous</span>
          </button>
          <button
            className="w-5 h-5 flex items-center justify-center rounded hover:bg-white/10 text-[#666] hover:text-white transition-colors"
            onClick={() => apiClient.navigateDisplay("next")}
            title="다음"
          >
            <span className="material-symbols-outlined" style={{ fontSize: "13px" }}>skip_next</span>
          </button>
          <div className="w-px h-3 bg-white/10 mx-0.5" />
          <span className="text-[9px] font-bold tracking-[0.12em] text-[#444] uppercase hidden sm:block">
            Service Timeline
          </span>
          {items.length > 0 && startMs !== null && endMs !== null && (
            <>
              <span className="text-[#333] hidden sm:block">·</span>
              <span className="text-[9px] text-[#555] font-mono hidden sm:block">{formatClock(startMs)}</span>
              <span className="text-[#333] hidden sm:block">→</span>
              <span className="text-[9px] text-[#555] font-mono hidden sm:block">{formatClock(endMs)}</span>
            </>
          )}
          {items.length > 0 && (
            <>
              <span className="text-[#333] hidden sm:block">·</span>
              {editingTotal ? (
                <div className="flex items-center gap-0.5" onClick={(e) => e.stopPropagation()}>
                  <input
                    autoFocus
                    type="text"
                    value={totalEditValue}
                    onChange={(e) => setTotalEditValue(e.target.value)}
                    onBlur={() => commitTotalEdit(totalEditValue)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter") commitTotalEdit(totalEditValue);
                      if (e.key === "Escape") setEditingTotal(false);
                      e.stopPropagation();
                    }}
                    className="w-10 text-[9px] font-mono bg-transparent border-b border-[#4a9eff] text-white outline-none text-center"
                    placeholder="0"
                  />
                  <span className="text-[9px] text-[#4a9eff]">분</span>
                </div>
              ) : (
                <span
                  className="text-[9px] text-[#555] hidden sm:block cursor-pointer hover:text-[#4a9eff] transition-colors"
                  title="클릭하여 총 예배시간 편집 (단위: 분)"
                  onClick={() => {
                    setTotalEditValue(String(Math.round(totalSecs / 60)));
                    setEditingTotal(true);
                  }}
                >
                  est. {formatDuration(totalSecs)}
                  <span className="opacity-40 ml-0.5">({Math.round(totalSecs / 60)}분)</span>
                </span>
              )}
            </>
          )}
        </div>
        <div className="flex items-center gap-2">
          {autoEnabled && autoCountdown > 0 && (
            <span className="text-[9px] font-mono text-pro-accent tabular-nums">{autoCountdown}s</span>
          )}
          <button
            className={`text-[9px] font-bold px-1.5 py-0.5 rounded transition-all ${
              autoEnabled
                ? "bg-pro-accent/20 text-pro-accent"
                : "text-[#555] hover:text-[#888] hover:bg-white/5"
            }`}
            onClick={() => setAutoEnabled((v) => !v)}
            title="자동 진행 ON/OFF"
          >
            AUTO
          </button>
          {/* 서비스 시작/리셋 버튼 */}
          <button
            className={`text-[9px] font-bold px-1.5 py-0.5 rounded transition-all ${
              serviceStart !== null
                ? "bg-red-600/20 text-red-400 hover:bg-red-600/30"
                : "text-[#555] hover:text-[#888] hover:bg-white/5"
            }`}
            onClick={() => setServiceStart(serviceStart !== null ? null : Date.now())}
            title={serviceStart !== null ? "서비스 타이머 리셋" : "서비스 시작 (경과 시간 측정 시작)"}
          >
            {serviceStart !== null ? "■" : "▶"}
          </button>
          <span className="text-[9px] font-mono text-pro-accent tabular-nums">
            ● {elapsed}
          </span>
        </div>
      </div>

      {/* ── 타임라인 블록 ── */}
      {items.length === 0 ? (
        <div className="flex items-center justify-center flex-1 text-[#444] text-[9px]">
          예배 순서를 전송하면 타임라인이 표시됩니다
        </div>
      ) : (
        <div
          ref={containerRef}
          className="flex flex-1 items-stretch overflow-hidden relative"
        >
          {items.map((item, i) => {
            const secs = getEffectiveSecs(item.key);
            const widthPct = (secs / totalSecs) * 100;
            const isActive = i === currentIdx;
            const activeColor = ACTIVE_COLOR;
            const cumulativeSecs = items
              .slice(0, i)
              .reduce((s, it) => s + getEffectiveSecs(it.key), 0);
            const slotStartMs = (startMs ?? 0) + cumulativeSecs * 1000;
            const timeLabel = startMs !== null ? formatClock(slotStartMs) : formatDuration(cumulativeSecs);
            const isDividerHovered = hoverDivider === i;

            return (
              <div
                key={item.key || i}
                className="relative flex flex-col overflow-hidden cursor-pointer"
                style={{
                  width: `${widthPct}%`,
                  minWidth: "20px",
                  flexShrink: 0,
                  background: isActive ? `${activeColor}28` : "#181818",
                  borderTop: `2px solid ${isActive ? activeColor : "#2a2a2a"}`,
                  borderRight: i < items.length - 1 ? "1px solid #1c1c1c" : "none",
                  transition: "background 0.15s, border-top-color 0.15s",
                }}
                onClick={() => apiClient.jumpDisplay(i)}
                onDoubleClick={(e) => {
                  e.stopPropagation();
                  setEditingKey(item.key);
                  setEditValue(String(secs));
                }}
                title={`${item.title}${item.obj && item.obj !== "-" ? " · " + item.obj : ""} (${secs}s)`}
              >
                {/* 내용 */}
                {editingKey === item.key ? (
                  <div className="flex items-center justify-center h-full px-1">
                    <input
                      autoFocus
                      type="text"
                      value={editValue}
                      onChange={(e) => setEditValue(e.target.value)}
                      onBlur={() => commitEdit(item.key, editValue)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter") commitEdit(item.key, editValue);
                        if (e.key === "Escape") setEditingKey(null);
                        e.stopPropagation();
                      }}
                      onClick={(e) => e.stopPropagation()}
                      className="w-full text-center text-[10px] font-mono bg-transparent border-b border-[#4a9eff] text-white outline-none"
                      placeholder="초"
                    />
                  </div>
                ) : (
                  <div className="flex flex-col justify-between h-full px-1.5 pt-1 pb-0.5 overflow-hidden">
                    <div className="overflow-hidden">
                      <span
                        className="text-[10px] font-semibold truncate block leading-tight"
                        style={{ color: isActive ? activeColor : "#555" }}
                      >
                        {item.title}
                      </span>
                      {item.obj && item.obj !== "-" && (
                        <span
                          className="text-[8px] truncate block leading-tight"
                          style={{ color: isActive ? `${activeColor}99` : "#3a3a3a" }}
                        >
                          {item.obj}
                        </span>
                      )}
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-[8px] font-mono text-[#444] leading-none">{timeLabel}</span>
                      <span className="text-[8px] font-mono text-[#333] leading-none">
                        {formatDurationShort(secs)}
                      </span>
                    </div>
                  </div>
                )}

                {/* 드래그 경계선 핸들 */}
                {i < items.length - 1 && (
                  <div
                    className="absolute right-0 top-0 bottom-0 z-10 flex items-center justify-center"
                    style={{
                      width: "8px",
                      cursor: "col-resize",
                      transform: "translateX(50%)",
                    }}
                    onMouseEnter={() => setHoverDivider(i)}
                    onMouseLeave={() => setHoverDivider(null)}
                    onMouseDown={(e) => handleDividerMouseDown(e, i)}
                    onClick={(e) => e.stopPropagation()}
                  >
                    <div
                      className="rounded-full transition-all"
                      style={{
                        width: isDividerHovered ? "3px" : "1px",
                        height: isDividerHovered ? "60%" : "40%",
                        background: isDividerHovered ? "#fff" : "#333",
                      }}
                    />
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

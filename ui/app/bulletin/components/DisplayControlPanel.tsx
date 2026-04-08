"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { useRecoilState, useSetRecoilState } from "recoil";
import { displayItemsState, displayPanelOpenState } from "@/recoilState";
import { apiClient } from "@/lib/apiClient";
import { WorshipOrderItem, OBSStatus, StreamStatus } from "@/types";
import { useWS } from "@/components/WebSocketProvider";
import FeatureGate from "@/components/FeatureGate";

type TimerState = {
  enabled: boolean;
  countdown: number;
  speedFactor: number;
};

type ScheduleCountdown = {
  label: string;
  minutes: number;
  seconds: number;
} | null;

export default function DisplayControlPanel() {
  const [items, setItems] = useRecoilState(displayItemsState);
  const setPanelOpen = useSetRecoilState(displayPanelOpenState);
  const onClose = () => setPanelOpen(false);
  const [idx, setIdx] = useState(0);
  const [subPageIdx, setSubPageIdx] = useState(0);
  const [obsStatus, setObsStatus] = useState<OBSStatus>({ connected: false, currentScene: "" });
  const [timer, setTimer] = useState<TimerState>({ enabled: false, countdown: 0, speedFactor: 1.0 });
  const [expandedItems, setExpandedItems] = useState<Set<number>>(new Set());
  const [loadingMsg, setLoadingMsg] = useState("");
  const [schedCountdown, setSchedCountdown] = useState<ScheduleCountdown>(null);
  const [streamStatus, setStreamStatus] = useState<StreamStatus>({ active: false, reconnecting: false, timecode: "", bytesSent: 0 });
  const { subscribe } = useWS();
  const listRef = useRef<HTMLDivElement>(null);
  const itemsRef = useRef(items);
  itemsRef.current = items;
  const dragRef = useRef<{ from: number; wasDragging: boolean } | null>(null);
  const schedTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reorderLockRef = useRef(false);
  const reorderSuppressRef = useRef(false);
  const keyCounterRef = useRef(0);

  const ensureUniqueKeys = useCallback((rawItems: WorshipOrderItem[]) => {
    const seen = new Set<string>();
    return rawItems.map((item) => {
      let k = item.key;
      if (!k || seen.has(k)) {
        k = `_auto_${keyCounterRef.current++}`;
        return { ...item, key: k };
      }
      seen.add(k);
      return item;
    });
  }, []);

  useEffect(() => {
    return subscribe((msg) => {
      if (msg.type === "position" && typeof msg.idx === "number") {
        setIdx(msg.idx);
        if (typeof msg.subPageIdx === "number") setSubPageIdx(msg.subPageIdx);
      }
      if (msg.type === "timer_state") {
        setTimer({
          enabled: !!msg.enabled,
          countdown: msg.countdown || 0,
          speedFactor: msg.speedFactor || 1.0,
        });
        if (typeof msg.subPageIdx === "number") setSubPageIdx(msg.subPageIdx);
        if (typeof msg.idx === "number") setIdx(msg.idx);
      }
      if (msg.type === "order" && Array.isArray(msg.items)) {
        setLoadingMsg("");
        if (!reorderSuppressRef.current) {
          setItems(ensureUniqueKeys(msg.items as WorshipOrderItem[]));
          setExpandedItems(new Set());
          if (typeof msg.idx === "number") {
            setIdx(msg.idx);
            setSubPageIdx(0);
          } else {
            setIdx(0);
            setSubPageIdx(0);
          }
        }
      }
      if (msg.type === "display_loading") {
        if (msg.done) {
          setLoadingMsg("");
        } else {
          setLoadingMsg(msg.message || "준비 중...");
        }
      }
      if (msg.type === "schedule_countdown") {
        setSchedCountdown({
          label: msg.label,
          minutes: msg.minutes,
          seconds: msg.seconds,
        });
        if (schedTimeoutRef.current) clearTimeout(schedTimeoutRef.current);
        schedTimeoutRef.current = setTimeout(() => setSchedCountdown(null), 3000);
      }
      if (msg.type === "schedule_started") {
        setSchedCountdown(null);
        if (schedTimeoutRef.current) clearTimeout(schedTimeoutRef.current);
      }
    });
  }, [subscribe, setItems, ensureUniqueKeys]);

  useEffect(() => {
    const poll = () => {
      apiClient.getDisplayStatus()
        .then((data: any) => {
          if (data.obs) setObsStatus(data.obs);
          if (data.stream) setStreamStatus(data.stream);
          if (Array.isArray(data.items) && data.items.length > 0 && itemsRef.current.length === 0) {
            setItems(ensureUniqueKeys(data.items as WorshipOrderItem[]));
            if (typeof data.idx === "number") setIdx(data.idx);
          }
        })
        .catch((e) => console.error("display/status poll 에러:", e));
    };
    poll();
    const t = setInterval(poll, 5000);
    return () => clearInterval(t);
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    const el = listRef.current?.querySelector("[data-active='true']");
    el?.scrollIntoView({ block: "nearest", behavior: "smooth" });
  }, [idx, subPageIdx]);

  const handleNav = useCallback((dir: "prev" | "next") => {
    apiClient.navigateDisplay(dir);
  }, []);

  const handleJump = useCallback((index: number) => {
    setIdx(index);
    apiClient.jumpDisplay(index);
  }, []);

  const handleSectionJump = useCallback((itemIdx: number, subPageIdx: number) => {
    setIdx(itemIdx);
    setSubPageIdx(subPageIdx);
    apiClient.jumpDisplay(itemIdx, subPageIdx);
  }, []);

  const handleRemove = useCallback((index: number) => {
    apiClient.removeFromDisplay(index);
  }, []);

  const clearDragHighlight = useCallback(() => {
    listRef.current?.querySelectorAll("[data-dragover]").forEach((el) => el.removeAttribute("data-dragover"));
  }, []);

  const handleDragStart = useCallback((e: React.DragEvent, index: number) => {
    dragRef.current = { from: index, wasDragging: true };
    e.dataTransfer.effectAllowed = "move";
    (e.target as HTMLElement).style.opacity = "0.4";
  }, []);

  const handleDragEnd = useCallback((e: React.DragEvent) => {
    (e.target as HTMLElement).style.opacity = "1";
    clearDragHighlight();
    setTimeout(() => { dragRef.current = null; }, 50);
  }, [clearDragHighlight]);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
    const target = e.currentTarget as HTMLElement;
    if (!target.hasAttribute("data-dragover")) {
      clearDragHighlight();
      target.setAttribute("data-dragover", "true");
    }
  }, [clearDragHighlight]);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    (e.currentTarget as HTMLElement).removeAttribute("data-dragover");
  }, []);

  const handleDrop = useCallback((e: React.DragEvent, toIndex: number) => {
    e.preventDefault();
    clearDragHighlight();
    if (!dragRef.current) return;
    const fromIndex = dragRef.current.from;
    dragRef.current = null;
    if (fromIndex === toIndex) return;
    if (reorderLockRef.current) return;

    reorderLockRef.current = true;
    reorderSuppressRef.current = true;
    setItems((prev) => {
      const next = [...prev];
      const [moved] = next.splice(fromIndex, 1);
      next.splice(toIndex, 0, moved);
      return next;
    });
    setIdx((prev) => {
      if (prev === fromIndex) return toIndex;
      let newIdx = prev;
      if (fromIndex < prev) newIdx--;
      if (toIndex <= newIdx) newIdx++;
      return Math.max(0, Math.min(newIdx, items.length - 1));
    });

    apiClient.reorderDisplay(fromIndex, toIndex).finally(() => {
      reorderLockRef.current = false;
      setTimeout(() => { reorderSuppressRef.current = false; }, 1500);
    });
  }, [clearDragHighlight, items.length, setItems]);

  const toggleItemExpand = useCallback((index: number) => {
    setExpandedItems((prev) => {
      const next = new Set(prev);
      if (next.has(index)) next.delete(index);
      else next.add(index);
      return next;
    });
  }, []);

  const handleTimerToggle = useCallback(() => {
    apiClient.timerControl(timer.enabled ? "disable" : "enable");
  }, [timer.enabled]);

  const handleTimerRepeat = useCallback(() => {
    apiClient.timerControl("repeat");
  }, []);

  const handleTimerRestart = useCallback(() => {
    apiClient.timerControl("restart");
  }, []);

  const handleSpeedChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const factor = parseFloat(e.target.value);
    apiClient.timerControl("speed", factor);
  }, []);

  const handleStreamToggle = useCallback(() => {
    const action = streamStatus.active ? "stop" : "start";
    apiClient.streamControl(action);
  }, [streamStatus.active]);

  useEffect(() => {
    let blocked = false;
    let debounceTimer: ReturnType<typeof setTimeout>;
    const handler = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
      if (blocked) return;
      if (e.key === "ArrowRight" || e.key === "ArrowLeft") {
        blocked = true;
        handleNav(e.key === "ArrowRight" ? "next" : "prev");
        debounceTimer = setTimeout(() => { blocked = false; }, 300);
      }
    };
    window.addEventListener("keydown", handler);
    return () => {
      window.removeEventListener("keydown", handler);
      clearTimeout(debounceTimer);
    };
  }, [handleNav]);

  return (
    <div className="h-full bg-[#1a1a2e] text-gray-200 flex flex-col">
      {/* Header */}
      <div className="flex justify-between items-center px-4 py-3.5 border-b border-white/10">
        <div className="flex items-center gap-2">
          <span className={`w-2.5 h-2.5 rounded-full inline-block ${
            obsStatus.connected
              ? "bg-green-500 shadow-[0_0_6px_#4caf50]"
              : "bg-red-500 shadow-[0_0_6px_#f44336]"
          }`} />
          <span className="text-[13px] text-white/70">
            OBS {obsStatus.connected ? obsStatus.currentScene : "Disconnected"}
          </span>
          <FeatureGate feature="obs_control" fallback={null}>
            {streamStatus.active && (
              <span className="text-[10px] font-bold text-white bg-red-600 px-1.5 py-0.5 rounded animate-pulse">
                LIVE
              </span>
            )}
          </FeatureGate>
        </div>
        <div className="flex items-center gap-2">
          <FeatureGate
            feature="obs_control"
            fallback={
              <span className="text-[11px] text-white/35 px-2.5 py-1">Pro</span>
            }
          >
            <button
              className={`px-2.5 py-1 text-[11px] font-semibold border-none rounded cursor-pointer text-white ${
                streamStatus.active
                  ? "bg-[#424242] hover:bg-[#616161]"
                  : "bg-red-600 hover:bg-red-800"
              }`}
              onClick={handleStreamToggle}
            >
              {streamStatus.active ? "방송 종료" : "방송 시작"}
            </button>
          </FeatureGate>
          <button
            className="bg-transparent border-none text-white/50 text-lg cursor-pointer hover:text-white"
            onClick={onClose}
          >
            ✕
          </button>
        </div>
      </div>

      {/* Schedule Countdown */}
      {schedCountdown && (
        <div className="flex items-center justify-center gap-3 px-4 py-3 bg-red-900 text-white border-b border-white/10">
          <span className="text-sm font-semibold">{schedCountdown.label}</span>
          <span className="text-xl font-bold font-mono tracking-widest">
            {String(schedCountdown.minutes).padStart(2, "0")}:{String(schedCountdown.seconds).padStart(2, "0")}
          </span>
        </div>
      )}

      {/* Navigation */}
      <div className="flex items-center justify-center gap-4 px-4 py-3 border-b border-white/10">
        <button
          className="px-5 py-2 text-lg bg-[#2a2a4a] text-white border border-white/15 rounded-lg cursor-pointer hover:bg-[#3a3a5a]"
          onClick={() => handleNav("prev")}
        >
          ◀
        </button>
        <span className="text-base font-semibold min-w-[60px] text-center">
          {items.length > 0 ? `${idx + 1} / ${items.length}` : "-"}
        </span>
        <button
          className="px-5 py-2 text-lg bg-[#2a2a4a] text-white border border-white/15 rounded-lg cursor-pointer hover:bg-[#3a3a5a]"
          onClick={() => handleNav("next")}
        >
          ▶
        </button>
      </div>

      {/* Timer Controls */}
      <div className="px-4 py-2.5 border-b border-white/10">
        <div className="flex gap-1.5 mb-2">
          <button
            className={`flex-1 px-2 py-1.5 text-xs border rounded-lg cursor-pointer whitespace-nowrap ${
              timer.enabled
                ? "bg-blue-600 border-blue-400 text-white"
                : "bg-[#2a2a4a] border-white/15 text-gray-200 hover:bg-[#3a3a5a]"
            }`}
            onClick={handleTimerToggle}
          >
            {timer.enabled ? "⏸" : "▶"} 자동
          </button>
          <button
            className="flex-1 px-2 py-1.5 text-xs bg-[#2a2a4a] text-gray-200 border border-white/15 rounded-lg cursor-pointer hover:bg-[#3a3a5a]"
            onClick={handleTimerRepeat}
          >
            반복
          </button>
          <button
            className="flex-1 px-2 py-1.5 text-xs bg-[#2a2a4a] text-gray-200 border border-white/15 rounded-lg cursor-pointer hover:bg-[#3a3a5a]"
            onClick={handleTimerRestart}
          >
            처음
          </button>
        </div>
        {timer.enabled && timer.countdown > 0 && (
          <div className="text-center text-sm font-semibold text-yellow-300 py-1">
            다음까지 {timer.countdown}초
          </div>
        )}
        <div className="flex items-center gap-2 text-xs text-white/60">
          <span>속도</span>
          <input
            type="range"
            min="0.5"
            max="1.5"
            step="0.1"
            value={timer.speedFactor}
            onChange={handleSpeedChange}
            className="flex-1 accent-blue-500"
          />
          <span>{Math.round(timer.speedFactor * 100)}%</span>
        </div>
      </div>

      {/* Current Item */}
      {items.length > 0 && items[idx] && (
        <div className="px-4 py-2.5 text-[15px] font-semibold text-blue-300 border-b border-white/10">
          {items[idx].title}
          {items[idx].obj && items[idx].obj !== "-" && (
            <span className="font-normal text-[13px] text-white/50"> — {items[idx].obj}</span>
          )}
        </div>
      )}

      {/* Order List */}
      <div className="flex-1 overflow-y-auto py-2" ref={listRef}>
        {loadingMsg && (
          <div className="flex items-center gap-2.5 px-4 py-3.5 text-white/70 text-[13px]">
            <div className="w-4 h-4 border-2 border-white/20 border-t-blue-500 rounded-full animate-spin shrink-0" />
            <span>{loadingMsg}</span>
          </div>
        )}
        {items.map((item, i) => {
          const hasSections = item.sections && item.sections.length > 0;
          const isExpanded = expandedItems.has(i);
          const isActive = i === idx;

          return (
            <div key={item.key || i}>
              <div
                className={`flex items-center gap-2.5 px-4 py-2.5 cursor-pointer border-l-[3px] transition-all ${
                  isActive
                    ? "bg-blue-500/15 border-l-blue-500"
                    : "border-l-transparent hover:bg-white/5"
                } data-[dragover]:border-t-2 data-[dragover]:border-t-blue-500 data-[dragover]:bg-blue-500/10`}
                data-active={isActive ? "true" : undefined}
                onClick={() => { if (!dragRef.current?.wasDragging) handleJump(i); }}
                draggable
                onDragStart={(e) => handleDragStart(e, i)}
                onDragEnd={handleDragEnd}
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={(e) => handleDrop(e, i)}
              >
                <span className="text-xs text-white/40 min-w-[20px] text-right">{i + 1}</span>
                <span className="text-sm font-medium shrink-0">{item.title}</span>
                <span className="text-xs text-white/50 overflow-hidden text-ellipsis whitespace-nowrap flex-1">
                  {item.obj && item.obj !== "-" ? item.obj : ""}
                </span>
                {hasSections && (
                  <button
                    className="bg-transparent border-none text-white/50 text-[10px] cursor-pointer px-1 shrink-0 hover:text-white"
                    onClick={(e) => {
                      e.stopPropagation();
                      toggleItemExpand(i);
                    }}
                  >
                    {isExpanded ? "▲" : "▼"}
                  </button>
                )}
                <button
                  className="bg-transparent border-none text-white/25 text-[11px] cursor-pointer px-1.5 shrink-0 transition-colors hover:text-red-500"
                  onClick={(e) => {
                    e.stopPropagation();
                    handleRemove(i);
                  }}
                  title="삭제"
                >
                  ✕
                </button>
              </div>
              {hasSections && isExpanded && (
                <div className="flex flex-col gap-1 py-1 px-2 pl-6 bg-white/[0.03]">
                  {item.sections!.map((sec, si) => {
                    const isSectionActive = i === idx && sec.startPage === subPageIdx;
                    return (
                      <div
                        key={si}
                        data-active={isSectionActive ? "true" : undefined}
                        className={`flex gap-2 px-2 py-1.5 rounded-lg cursor-pointer transition-all border ${
                          isSectionActive
                            ? "bg-blue-600 border-blue-400"
                            : "bg-[#2a2a4a] border-white/[0.08] hover:bg-blue-600 hover:border-blue-400"
                        }`}
                        onClick={() => handleSectionJump(i, sec.startPage)}
                      >
                        <span className={`text-[11px] font-semibold whitespace-nowrap min-w-[32px] pt-px ${
                          isSectionActive ? "text-white" : "text-blue-300"
                        }`}>
                          {sec.label}
                        </span>
                        <pre className={`font-[inherit] text-[11px] leading-snug m-0 whitespace-pre-wrap ${
                          isSectionActive ? "text-white" : "text-white/60"
                        }`}>
                          {sec.text || ""}
                        </pre>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

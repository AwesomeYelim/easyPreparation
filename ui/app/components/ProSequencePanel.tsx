"use client";

import { useEffect, useState, useCallback, useRef, useLayoutEffect, useMemo } from "react";
import { useRecoilState, useRecoilValue, useSetRecoilState } from "recoil"; // eslint-disable-line @typescript-eslint/no-unused-vars
import { sequencePanelOpenState, panelDetachedState, panelNativeState, itemTimersState, inspectorTabState, includeThumbnailOnStreamState, includeTitleOnStreamState } from "@/recoilState";
import { apiClient, openDisplayWindow, panelApi } from "@/lib/apiClient";
import { useWS } from "@/components/WebSocketProvider";
import FeatureGate from "@/components/FeatureGate";
import { useDisplaySession } from "@/lib/useDisplaySession";


type ProSequencePanelProps = {
  /** true = 분리된 창(/remote)에서 단독 렌더링 */
  standalone?: boolean;
};

export default function ProSequencePanel({ standalone = false }: ProSequencePanelProps) {
  const seqOpen = useRecoilValue(sequencePanelOpenState);
  const detached = useRecoilValue(panelDetachedState);
  // 분리 기능은 데스크톱 앱에서만 (분리 창 안에서는 항상 붙이기·핀 표시)
  const canDetach = useRecoilValue(panelNativeState) || standalone;
  const [pinned, setPinned] = useState(false);
  const headerDragRef = useRef<{ x: number; y: number; id: number } | null>(null);
  const setInspectorTab = useSetRecoilState(inspectorTabState);
  const [expandedItems, setExpandedItems] = useState<Set<number>>(new Set());

  // 진행 상태·조작 — 폰 리모컨(MobileRemote)과 공통 (lib/useDisplaySession)
  const session = useDisplaySession({ pollMs: 5000, onOrder: () => setExpandedItems(new Set()) });
  const {
    items, setItems, idx, subIdx: subPageIdx, timerOn: timerEnabled,
    obsStatus, streamStatus, schedCountdown, loadingMsg, pageVisible,
  } = session;
  const includeThumbnail = useRecoilValue(includeThumbnailOnStreamState);
  const includeTitle = useRecoilValue(includeTitleOnStreamState);
  const [viewMode, setViewMode] = useState<"list" | "grid">("list");
  const gridRef = useRef<HTMLDivElement>(null);
  const [gridCellWidth, setGridCellWidth] = useState(120);
  const [cameraMode, setCameraMode] = useState(false);
  // obs.json scenes + presets 맵 (title 기반, 서버 저장)
  const [obsSceneMapping, setObsSceneMapping] = useState<Record<string, string>>({});
  const [obsCameraScene, setObsCameraScene] = useState("camera");
  const [obsDisplayScene, setObsDisplayScene] = useState("monitor");
  // 카메라에 저장된 프리셋 목록
  const [cameraPresets, setCameraPresets] = useState<{ token: string; name: string }[]>([]);

  // Per-item timer state: Recoil (shared with ProTimeline)
  const itemTimers = useRecoilValue(itemTimersState);

  const { subscribe } = useWS();
  const listRef = useRef<HTMLDivElement>(null);
  const dragRef = useRef<{ from: number; wasDragging: boolean } | null>(null);

  // --- OBS 장면 매핑·카메라 프리셋 (보일 때 1회) ---
  useEffect(() => {
    if (!pageVisible) return;
    // obs.json scenes + presets 매핑 1회 로드
    apiClient.getObsSceneMapping()
      .then((data: any) => {
        setObsSceneMapping(data.mapping ?? {});
        if (data.cameraScene) setObsCameraScene(data.cameraScene);
        if (data.displayScene) setObsDisplayScene(data.displayScene);
      })
      .catch(() => {});
    // 카메라 프리셋 목록 로드 (연결 안 돼있으면 조용히 실패)
    apiClient.ptzGetPresets()
      .then((res) => { if (res.ok && res.presets) setCameraPresets(res.presets); })
      .catch(() => {});
  }, [pageVisible]); // eslint-disable-line react-hooks/exhaustive-deps

  // 사이드바 자동 스크롤 없음 — 현재 항목이 바뀌어도(클릭/자연 진행/폴링 모두) 운영자의 스크롤 위치를 건드리지 않는다.

  // --- Navigation ---
  const { navigate: handleNav, jump, setStream, reorder } = session;

  const handleJump = useCallback((index: number) => {
    jump(index);
    if (!standalone) setInspectorTab("preview"); // 분리 창에서는 메인 창 탭을 건드리지 않음
  }, [jump, setInspectorTab, standalone]);

  const handleSectionJump = useCallback((itemIdx: number, subPage: number) => {
    jump(itemIdx, subPage);
    if (!standalone) setInspectorTab("preview");
  }, [jump, setInspectorTab, standalone]);

  const handleRemove = useCallback((index: number) => {
    apiClient.removeFromDisplay(index);
  }, []);


  // --- Camera mode toggle (live override) ---
  const handleCameraToggle = useCallback(() => {
    const newMode = !cameraMode;
    setCameraMode(newMode);
    apiClient.ptzSetSource(newMode ? "camera" : "slides");
  }, [cameraMode]);

  // --- Stream toggle ---
  const handleStreamToggle = useCallback(() => {
    setStream(!streamStatus.active, { includeThumbnail, includeTitle }).catch(() => {});
  }, [setStream, streamStatus.active, includeThumbnail, includeTitle]);

  // --- 테스트 방송 (썸네일 없이, 제목 "TEST ...") ---
  const handleTestStream = useCallback(() => {
    if (streamStatus.active) return;
    setStream(true, { includeThumbnail: false, isTest: true }).catch(() => {});
  }, [setStream, streamStatus.active]);

  // --- Drag & drop ---
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
    reorder(fromIndex, toIndex);
  }, [clearDragHighlight, reorder]);

  const toggleItemExpand = useCallback((index: number) => {
    setExpandedItems((prev) => {
      const next = new Set(prev);
      if (next.has(index)) next.delete(index);
      else next.add(index);
      return next;
    });
  }, []);

  // --- Grid cell width measurement ---
  useLayoutEffect(() => {
    if (viewMode !== "grid" || !gridRef.current) return;
    const el = gridRef.current;
    const obs = new ResizeObserver((entries) => {
      const w = entries[0]?.contentRect.width;
      if (w) setGridCellWidth((w - 4) / 2); // 2 cols, gap-1(4px)
    });
    obs.observe(el);
    return () => obs.disconnect();
  }, [viewMode]);

  // grid iframe key: changes when items list changes (force re-render on order change)
  const itemsHash = useMemo(() => items.map(it => it.key || it.title).join(","), [items]);

  // --- Keyboard arrow navigation ---
  useEffect(() => {
    let blocked = false;
    let debounceTimer: ReturnType<typeof setTimeout>;
    const handler = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      const tag = target?.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
      if (target?.isContentEditable) return;
      if (blocked) return;
      if (e.key === "ArrowRight" || e.key === "ArrowLeft" || e.key === "ArrowDown" || e.key === "ArrowUp" || e.key === " " || e.key === "Spacebar") {
        e.preventDefault();
        blocked = true;
        const isPrev = e.key === "ArrowLeft" || e.key === "ArrowUp";
        handleNav(isPrev ? "prev" : "next");
        debounceTimer = setTimeout(() => { blocked = false; }, 300);
      }
    };
    window.addEventListener("keydown", handler);
    return () => {
      window.removeEventListener("keydown", handler);
      clearTimeout(debounceTimer);
    };
  }, [handleNav]);

  // 분리 창 핀(항상 위) 상태 — 서버 저장값 + WS panel_dock
  useEffect(() => {
    if (!standalone) return;
    panelApi.getState().then((st) => setPinned(!!st.pinned)).catch(() => {});
    return subscribe((msg) => {
      if (msg.type === "panel_dock" && typeof msg.pinned === "boolean") setPinned(msg.pinned);
    });
  }, [standalone, subscribe]);

  // 제목줄 끌어서 떼기 — 일정 거리 이상 끌면 분리 창이 커서 아래에 뜨고 드래그를 이어받는다(데스크톱)
  const onTitlePointerDown = (e: React.PointerEvent<HTMLDivElement>) => {
    if (e.button !== 0 || (e.target as HTMLElement).closest("button")) return;
    headerDragRef.current = { x: e.clientX, y: e.clientY, id: e.pointerId };
    e.currentTarget.setPointerCapture(e.pointerId);
  };
  const onTitlePointerMove = (e: React.PointerEvent<HTMLDivElement>) => {
    const d = headerDragRef.current;
    if (!d || d.id !== e.pointerId) return;
    if (Math.hypot(e.clientX - d.x, e.clientY - d.y) < 24) return;
    headerDragRef.current = null;
    e.currentTarget.releasePointerCapture(e.pointerId);
    panelApi.detach({ drag: true }).catch(() => {});
  };
  const onTitlePointerUp = (e: React.PointerEvent<HTMLDivElement>) => {
    if (headerDragRef.current?.id === e.pointerId) headerDragRef.current = null;
  };

  // 메인 창: 분리돼 있으면 패널을 그리지 않는다. (훅은 위에서 계속 동작 → 화살표 키 진행 유지)
  if (!standalone && (!seqOpen || detached)) return null;

  const currentItem = items[idx];

  return (
    <div
      className="flex flex-col bg-pro-surface border-r border-pro-border overflow-hidden"
      data-testid="seq-panel"
      style={standalone ? { flex: 1, minHeight: 0 } : { gridColumn: "2", gridRow: "2" }}
    >
      {/* ── 독 제목줄 (OBS 독 방식): 끌어서 떼기 · 더블클릭으로 떼기/붙이기 ── */}
      <div
        className="flex items-center justify-between h-7 pl-3 pr-1.5 bg-pro-elevated/60 border-b border-pro-border select-none flex-shrink-0"
        style={standalone || !canDetach ? undefined : { cursor: "grab", touchAction: "none" }}
        data-testid="seq-panel-titlebar"
        onPointerDown={standalone || !canDetach ? undefined : onTitlePointerDown}
        onPointerMove={standalone || !canDetach ? undefined : onTitlePointerMove}
        onPointerUp={standalone || !canDetach ? undefined : onTitlePointerUp}
        onPointerCancel={standalone || !canDetach ? undefined : onTitlePointerUp}
        onDoubleClick={canDetach ? () => (standalone ? panelApi.dock() : panelApi.detach()).catch(() => {}) : undefined}
      >
        <span className="text-[11px] font-semibold text-pro-text-dim tracking-wide">예배 순서</span>
        <div className="flex items-center gap-0.5">
          {standalone && (
            <button
              onClick={() => { const next = !pinned; setPinned(next); panelApi.setPinned(next).catch(() => setPinned(!next)); }}
              title={pinned ? "항상 위 해제" : "모든 창 위에 고정"}
              data-testid="seq-panel-pin"
              className={`w-6 h-6 flex items-center justify-center rounded transition-colors ${
                pinned ? "text-[#4a9eff] bg-[#4a9eff]/15" : "text-pro-text-dim hover:text-pro-text hover:bg-pro-hover"
              }`}
            >
              <span className="material-symbols-outlined" style={{ fontSize: "14px", fontVariationSettings: `'FILL' ${pinned ? 1 : 0}` }}>push_pin</span>
            </button>
          )}
          {canDetach && (<button
            onClick={() => (standalone ? panelApi.dock() : panelApi.detach()).catch(() => {})}
            title={standalone ? "메인 창에 다시 붙이기" : "별도 창으로 떼기 (제목줄을 끌어도 됩니다)"}
            data-testid={standalone ? "seq-panel-dock" : "seq-panel-detach"}
            className="w-6 h-6 flex items-center justify-center rounded text-pro-text-dim hover:text-pro-text hover:bg-pro-hover transition-colors"
          >
            <span className="material-symbols-outlined" style={{ fontSize: "14px" }}>{standalone ? "close_fullscreen" : "open_in_new"}</span>
          </button>)}
        </div>
      </div>

      {/* ── Header: OBS dot + scene | LIVE badge | stream button ── */}
      <div className="flex items-center justify-between px-3 py-2 border-b border-pro-border flex-shrink-0">
        <div className="flex items-center gap-2 min-w-0">
          <span
            className={`w-2 h-2 rounded-full flex-shrink-0 ${
              obsStatus.connected
                ? "bg-green-500 shadow-[0_0_5px_#4caf50]"
                : "bg-red-500 shadow-[0_0_5px_#f44336]"
            }`}
          />
          <span className="text-pro-text-dim text-[11px] truncate">
            {obsStatus.connected ? obsStatus.currentScene || "OBS" : "OBS Disconnected"}
          </span>
          <FeatureGate feature="obs_control" fallback={null}>
            {streamStatus.active && (
              <span className="text-[9px] font-bold text-white bg-red-600 px-1.5 py-0.5 rounded animate-pulse flex-shrink-0">
                LIVE
              </span>
            )}
          </FeatureGate>
          {timerEnabled && (
            <span className="text-[9px] font-bold text-white bg-blue-600 px-1.5 py-0.5 rounded flex-shrink-0">
              AUTO
            </span>
          )}
        </div>
        <div className="flex items-center gap-1.5 flex-shrink-0">
          {/* 카메라/슬라이드 수동 전환 */}
          <button
            onClick={handleCameraToggle}
            title={cameraMode ? "슬라이드로 전환" : "카메라로 전환"}
            className={`w-6 h-6 flex items-center justify-center rounded transition-colors ${
              cameraMode
                ? "bg-[#4a9eff]/20 text-[#4a9eff] hover:bg-[#4a9eff]/30"
                : "hover:bg-pro-hover text-pro-text-dim hover:text-pro-text"
            }`}
          >
            <svg width="13" height="13" viewBox="0 0 16 16" fill="none">
              <rect x="1" y="4" width="9" height="8" rx="1.2" stroke="currentColor" strokeWidth="1.4"/>
              <path d="M10 7L15 5V11L10 9V7Z" stroke="currentColor" strokeWidth="1.4" strokeLinejoin="round"/>
            </svg>
          </button>
          {/* Display 새 창 열기 */}
          <button
            onClick={() => openDisplayWindow(true)}
            title="Display 새 창으로 열기"
            className="w-6 h-6 flex items-center justify-center rounded hover:bg-pro-hover text-pro-text-dim hover:text-pro-text transition-colors"
          >
            <svg width="13" height="13" viewBox="0 0 16 16" fill="none">
              <rect x="1" y="2" width="14" height="10" rx="1.5" stroke="currentColor" strokeWidth="1.4"/>
              <path d="M5 14H11M8 12V14" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round"/>
            </svg>
          </button>
          <FeatureGate
            feature="obs_control"
            fallback={
              <span className="text-[10px] text-pro-text-dim/40 px-2">Pro</span>
            }
          >
            {!streamStatus.active && (
              <button
                className="px-2 py-1 text-[10px] font-semibold rounded cursor-pointer flex-shrink-0 bg-pro-elevated text-pro-text-dim border border-pro-border hover:bg-pro-hover hover:text-pro-text"
                onClick={handleTestStream}
                title="썸네일 없이 'TEST' 제목으로 방송 시작 (테스트용)"
              >
                테스트 방송
              </button>
            )}
            <button
              className={`px-2 py-1 text-[10px] font-semibold rounded cursor-pointer flex-shrink-0 ${
                streamStatus.active
                  ? "bg-pro-elevated text-pro-text-dim hover:bg-pro-hover"
                  : "bg-red-600 text-white hover:bg-red-700"
              }`}
              onClick={handleStreamToggle}
            >
              {streamStatus.active ? "방송 종료" : "방송 시작"}
            </button>
          </FeatureGate>
        </div>
      </div>

      {/* ── Schedule countdown bar ── */}
      {schedCountdown && (
        <div className="flex items-center justify-center gap-2 px-3 py-2 bg-red-900/80 border-b border-pro-border flex-shrink-0">
          <span className="text-[12px] font-semibold text-white">{schedCountdown.label}</span>
          <span className="text-base font-bold font-mono tracking-widest text-white">
            {String(schedCountdown.minutes).padStart(2, "0")}:{String(schedCountdown.seconds).padStart(2, "0")}
          </span>
        </div>
      )}

      {/* ── Current item display ── */}
      {currentItem && (
        <div className="px-3 py-2 border-b border-pro-border flex-shrink-0">
          <div className="text-[13px] font-semibold text-pro-accent truncate">
            {currentItem.title}
            {currentItem.obj && currentItem.obj !== "-" && (
              <span className="font-normal text-[11px] text-pro-text-dim"> — {currentItem.obj}</span>
            )}
          </div>
        </div>
      )}

      {/* ── Navigation row + view toggle ── */}
      <div className="flex items-center gap-1.5 px-3 py-2 border-b border-pro-border flex-shrink-0">
        <button
          className="flex-1 py-1.5 text-xs font-semibold bg-pro-elevated text-pro-text border border-pro-border rounded cursor-pointer hover:bg-pro-hover transition-colors"
          onClick={() => handleNav("prev")}
        >
          ◀ PREV
        </button>
        <span className="text-[11px] text-pro-text-dim tabular-nums min-w-[44px] text-center flex-shrink-0">
          {items.length > 0 ? `${idx + 1} / ${items.length}` : "-"}
        </span>
        <button
          className="flex-1 py-1.5 text-xs font-semibold bg-pro-elevated text-pro-text border border-pro-border rounded cursor-pointer hover:bg-pro-hover transition-colors"
          onClick={() => handleNav("next")}
        >
          NEXT ▶
        </button>
        <button
          className={`p-1.5 rounded border transition-colors flex-shrink-0 ${
            viewMode === "grid"
              ? "bg-pro-accent/20 border-pro-accent text-pro-accent"
              : "bg-pro-elevated border-pro-border text-pro-text-dim hover:bg-pro-hover"
          }`}
          onClick={() => setViewMode((v) => (v === "list" ? "grid" : "list"))}
          title={viewMode === "list" ? "그리드 보기" : "목록 보기"}
        >
          <span className="material-symbols-outlined" style={{ fontSize: "13px", lineHeight: 1 }}>
            {viewMode === "list" ? "grid_view" : "view_list"}
          </span>
        </button>
      </div>


      {/* ── Grid view ── */}
      {viewMode === "grid" && (
        <div ref={gridRef} className="flex-1 overflow-y-auto">
          {items.length === 0 ? (
            <div className="flex items-center justify-center h-full text-pro-text-dim text-[12px] px-4 text-center">
              예배 순서가 없습니다
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-1 p-2">
              {items.map((item, i) => {
                const isActive = i === idx;
                const isPdfOnly = item.info === "pdf_only";
                const cellHeight = Math.round(gridCellWidth * 9 / 16);
                const scale = gridCellWidth / 1920;
                return (
                  <div
                    key={item.key || i}
                    className="relative rounded overflow-hidden cursor-pointer border-2 transition-all hover:opacity-90"
                    style={{
                      height: `${cellHeight}px`,
                      borderColor: isActive
                        ? "#3B82F6"
                        : isPdfOnly
                        ? "#92400e"
                        : "#2a2a2a",
                    }}
                    onClick={() => handleJump(i)}
                    title={`${i + 1}. ${item.title}${item.obj && item.obj !== "-" ? ` — ${item.obj}` : ""}`}
                  >
                    <iframe
                      key={`${i}-${itemsHash}`}
                      src={`/display/preview?index=${i}`}
                      scrolling="no"
                      style={{
                        position: "absolute",
                        top: 0,
                        left: 0,
                        width: "1920px",
                        height: "1080px",
                        transform: `scale(${scale})`,
                        transformOrigin: "top left",
                        border: "none",
                        pointerEvents: "none",
                      }}
                    />
                    {/* Label overlay */}
                    <div className="absolute bottom-0 left-0 right-0 bg-black/70 px-1 py-0.5 flex items-center gap-1">
                      <span className="text-[7px] text-white/60 flex-shrink-0">{i + 1}</span>
                      <span className="text-[7px] text-white truncate">{item.title}</span>
                      {isPdfOnly && (
                        <span className="text-[6px] font-bold text-[#f59e0b] flex-shrink-0">PDF</span>
                      )}
                    </div>
                    {/* Active indicator */}
                    {isActive && (
                      <div className="absolute inset-0 ring-2 ring-inset ring-pro-accent pointer-events-none rounded" />
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}

      {/* ── Order list ── */}
      {viewMode === "list" && (
      <div className="flex-1 overflow-y-auto py-1" ref={listRef}>
        {loadingMsg && (
          <div className="flex items-center gap-2 px-3 py-3 text-pro-text-dim text-[12px]">
            <div className="w-3.5 h-3.5 border-2 border-pro-border border-t-pro-accent rounded-full animate-spin flex-shrink-0" />
            <span>{loadingMsg}</span>
          </div>
        )}

        {items.length === 0 && !loadingMsg && (
          <div className="flex flex-col items-center justify-center h-full gap-4 px-4 py-8 text-center">
            <span
              className="material-symbols-outlined text-pro-text-dim opacity-30"
              style={{ fontSize: "32px" }}
            >
              queue_music
            </span>
            <div className="text-[13px] font-semibold text-pro-text">예배 순서를 시작해보세요</div>
            <div className="flex flex-col gap-2 w-full mt-2">
              {[
                { step: "1", label: "상단 탭에서 예배 타입 선택" },
                { step: "2", label: "예배 순서 편집 후 저장" },
                { step: "3", label: "Display 전송 버튼 클릭" },
              ].map(({ step, label }) => (
                <div key={step} className="flex items-center gap-2 bg-white/5 rounded-lg px-3 py-2">
                  <span className="w-5 h-5 rounded-full bg-[#4a9eff]/20 text-[#4a9eff] text-[10px] font-bold flex items-center justify-center flex-shrink-0">
                    {step}
                  </span>
                  <span className="text-[11px] text-pro-text-dim text-left">{label}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {items.map((item, i) => {
          const hasSections = item.sections && item.sections.length > 0;
          const isExpanded = expandedItems.has(i);
          const isActive = i === idx;
          const isPdfOnly = item.info === "pdf_only";

          return (
            <div key={item.key || i}>
              <div
                className={`flex items-center gap-1.5 px-3 py-1 cursor-pointer border-l-2 transition-all ${
                  isPdfOnly
                    ? "opacity-45 border-l-transparent hover:bg-pro-hover hover:opacity-60"
                    : isActive
                    ? "bg-pro-accent/10 border-l-pro-accent"
                    : "border-l-transparent hover:bg-pro-hover"
                } data-[dragover]:border-t-2 data-[dragover]:border-t-pro-accent data-[dragover]:bg-pro-accent/5`}
                data-testid="seq-item"
                data-active={isActive ? "true" : undefined}
                onClick={() => { if (!dragRef.current?.wasDragging) handleJump(i); }}
                draggable
                onDragStart={(e) => handleDragStart(e, i)}
                onDragEnd={handleDragEnd}
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={(e) => handleDrop(e, i)}
              >
                {/* Index */}
                <span className="text-[10px] text-pro-text-dim/50 min-w-[18px] text-right flex-shrink-0">
                  {i + 1}
                </span>

                {/* Title + Obj (flex-1 min-w-0으로 묶어 패널 너비 초과 방지) */}
                <div className="flex-1 flex items-center gap-1.5 min-w-0 overflow-hidden">
                  <span
                    className={`text-[12px] font-medium truncate flex-shrink min-w-0 ${isPdfOnly ? "text-pro-text-dim" : "text-pro-text"}`}
                    title={item.title}
                  >
                    {item.title}
                  </span>

                  {/* PDF 전용 배지 */}
                  {isPdfOnly && (
                    <span className="text-[8px] font-semibold bg-[#2a1a00] text-[#f59e0b] px-1 py-0.5 rounded flex-shrink-0">
                      PDF
                    </span>
                  )}

                  {/* Obj */}
                  <span className="text-[11px] text-pro-text-dim truncate flex-1 min-w-0">
                    {item.obj && item.obj !== "-" ? item.obj : ""}
                  </span>
                </div>

                {/* OBS 씬 매핑 배지 (obs.json scenes 기반) */}
                {(() => {
                  const mapped = obsSceneMapping[item.title];
                  const isCamera = mapped === obsCameraScene;
                  const isDisplay = mapped === obsDisplayScene;
                  return (
                    <button
                      className={`w-5 h-5 flex items-center justify-center rounded border cursor-pointer transition-all flex-shrink-0 ${
                        isCamera
                          ? "bg-blue-900/50 text-[#4a9eff] border-blue-700/60"
                          : isDisplay
                          ? "bg-emerald-900/40 text-emerald-400 border-emerald-700/50"
                          : "bg-transparent text-pro-text-muted/20 border-transparent hover:border-pro-border hover:text-pro-text-muted/60"
                      }`}
                      title={`씬 매핑: ${mapped || "없음"} (클릭: 카메라→슬라이드→없음 순환)`}
                      onClick={(e) => {
                        e.stopPropagation();
                        // 순환: 없음 → cameraScene → displayScene → 없음
                        const next = !mapped
                          ? obsCameraScene
                          : mapped === obsCameraScene
                          ? obsDisplayScene
                          : "";
                        setObsSceneMapping((prev) => {
                          const updated = { ...prev };
                          if (next) updated[item.title] = next;
                          else delete updated[item.title];
                          return updated;
                        });
                        apiClient.setObsSceneMapping(item.title, next).catch(() => {});
                      }}
                    >
                      {isCamera ? (
                        <svg width="11" height="11" viewBox="0 0 16 16" fill="none">
                          <rect x="1" y="4" width="9" height="8" rx="1.2" stroke="currentColor" strokeWidth="1.4"/>
                          <path d="M10 7L15 5V11L10 9V7Z" stroke="currentColor" strokeWidth="1.4" strokeLinejoin="round"/>
                        </svg>
                      ) : isDisplay ? (
                        <svg width="11" height="11" viewBox="0 0 16 16" fill="none">
                          <rect x="1" y="2" width="14" height="10" rx="1.2" stroke="currentColor" strokeWidth="1.4"/>
                          <path d="M5 12L8 15L11 12" stroke="currentColor" strokeWidth="1.4" strokeLinejoin="round"/>
                        </svg>
                      ) : (
                        <svg width="11" height="11" viewBox="0 0 16 16" fill="none">
                          <circle cx="8" cy="8" r="2.5" stroke="currentColor" strokeWidth="1.4"/>
                        </svg>
                      )}
                    </button>
                  );
                })()}

                {/* PTZ 프리셋 셀렉트 (카메라 씬 항목만 — monitor 씬은 숨김) */}
                {obsSceneMapping[item.title] !== obsDisplayScene && (() => {
                  // 프리셋은 항목(셀) 자체의 ptzPreset 필드 — 예배 타입·같은 제목 항목 간 독립
                  const currentToken = String(item.ptzPreset || "");
                  const savePreset = (val: number) => {
                    setItems((prev) => prev.map((it, j) => (j === i ? { ...it, ptzPreset: val > 0 ? val : undefined } : it)));
                    apiClient.patchDisplayItemPreset(i, val).catch(() => {});
                  };
                  // 카메라 프리셋 있으면 셀렉트, 없으면 P· 사이클 버튼
                  if (cameraPresets.length > 0) {
                    return (
                      <select
                        className={`text-[9px] rounded border cursor-pointer flex-shrink-0 leading-none font-mono bg-transparent outline-none transition-all px-0.5 py-0.5 appearance-none ${
                          currentToken
                            ? "text-[#4a9eff] border-blue-700/60 bg-blue-900/30"
                            : "text-pro-text-muted/30 border-transparent hover:border-pro-border hover:text-pro-text-muted/60"
                        }`}
                        style={{ maxWidth: "52px" }}
                        title="PTZ 프리셋"
                        value={currentToken}
                        onClick={(e) => e.stopPropagation()}
                        onChange={(e) => {
                          e.stopPropagation();
                          const val = parseInt(e.target.value) || 0;
                          savePreset(val);
                          // 선택 즉시 카메라 이동 (미리보기)
                          if (val > 0) apiClient.ptzGoto(val).catch(() => {});
                        }}
                      >
                        <option value="">P·</option>
                        {cameraPresets.map((p) => (
                          <option key={p.token} value={p.token}>
                            {p.name || `P${p.token}`}
                          </option>
                        ))}
                      </select>
                    );
                  }
                  // 카메라 미연결 — 숫자 사이클 버튼 fallback
                  const preset = parseInt(currentToken) || 0;
                  return (
                    <button
                      className={`text-[9px] px-1 py-0.5 rounded border cursor-pointer transition-all flex-shrink-0 leading-none font-mono ${
                        preset > 0
                          ? "bg-blue-900/50 text-[#4a9eff] border-blue-700/60"
                          : "bg-transparent text-pro-text-muted/20 border-transparent hover:border-pro-border hover:text-pro-text-muted/60"
                      }`}
                      title="PTZ 프리셋 (없음→P1→P2→P3→없음, 카메라 미연결)"
                      onClick={(e) => {
                        e.stopPropagation();
                        const next = (preset + 1) % 4;
                        savePreset(next);
                      }}
                    >
                      {preset > 0 ? `P${preset}` : "P·"}
                    </button>
                  );
                })()}

                {/* Expand/collapse sections */}
                {hasSections && (
                  <button
                    className="bg-transparent border-none text-pro-text-dim/50 text-[9px] cursor-pointer flex-shrink-0 hover:text-pro-text transition-colors w-6 h-6 flex items-center justify-center"
                    onClick={(e) => {
                      e.stopPropagation();
                      toggleItemExpand(i);
                    }}
                  >
                    {isExpanded ? "▲" : "▼"}
                  </button>
                )}

                {/* Remove button */}
                <button
                  className="bg-transparent border-none text-pro-text-dim/25 text-[10px] cursor-pointer flex-shrink-0 hover:text-red-500 transition-colors w-6 h-6 flex items-center justify-center"
                  onClick={(e) => {
                    e.stopPropagation();
                    handleRemove(i);
                  }}
                  title="삭제"
                >
                  ✕
                </button>
              </div>

              {/* Expanded sections */}
              {hasSections && isExpanded && (
                <div className="flex flex-col gap-1 py-1 px-2 pl-7 bg-pro-elevated/30">
                  {item.sections!.map((sec, si) => {
                    const isSectionActive = i === idx && sec.startPage === subPageIdx;
                    return (
                      <div
                        key={si}
                        data-active={isSectionActive ? "true" : undefined}
                        className={`flex gap-2 px-2 py-1.5 rounded cursor-pointer transition-all border ${
                          isSectionActive
                            ? "bg-pro-accent border-pro-accent/60 text-white"
                            : "bg-pro-elevated border-pro-border hover:bg-pro-hover"
                        }`}
                        onClick={() => handleSectionJump(i, sec.startPage)}
                      >
                        <span className={`text-[10px] font-semibold whitespace-nowrap min-w-[28px] pt-px ${
                          isSectionActive ? "text-white" : "text-pro-accent"
                        }`}>
                          {sec.label}
                        </span>
                        <pre className={`font-[inherit] text-[10px] leading-snug m-0 whitespace-pre-wrap ${
                          isSectionActive ? "text-white" : "text-pro-text-dim"
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
      )}
    </div>
  );
}

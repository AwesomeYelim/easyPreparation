"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { useRecoilState, useSetRecoilState } from "recoil";
import { displayItemsState, displayPanelOpenState } from "@/recoilState";
import { apiClient } from "@/lib/apiClient";
import { WorshipOrderItem, OBSStatus, StreamStatus } from "@/types";
import { useWS } from "@/components/WebSocketProvider";
import s from "./DisplayControlPanel.module.scss";

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
  const dragRef = useRef<{ from: number } | null>(null);
  const [dragOver, setDragOver] = useState<number | null>(null);

  // WS: position, timer_state, order, display_loading 동기화
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
        setItems(msg.items as WorshipOrderItem[]);
        setExpandedItems(new Set());
        if (typeof msg.idx === "number") {
          setIdx(msg.idx);
          setSubPageIdx(0);
        } else {
          setIdx(0);
          setSubPageIdx(0);
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
      }
      if (msg.type === "schedule_started") {
        setSchedCountdown(null);
      }
    });
  }, [subscribe, setItems]);

  // 마운트 시 현재 상태 fetch + OBS 폴링 (5초)
  useEffect(() => {
    const poll = () => {
      apiClient.getDisplayStatus()
        .then((data: any) => {
          if (data.obs) setObsStatus(data.obs);
          if (data.stream) setStreamStatus(data.stream);
          if (Array.isArray(data.items) && data.items.length > 0 && itemsRef.current.length === 0) {
            setItems(data.items as WorshipOrderItem[]);
            if (typeof data.idx === "number") setIdx(data.idx);
          }
        })
        .catch((e) => console.error("display/status poll 에러:", e));
    };
    poll();
    const t = setInterval(poll, 5000);
    return () => clearInterval(t);
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // 활성 항목/섹션 자동 스크롤
  useEffect(() => {
    const el = listRef.current?.querySelector(`.${s.order_section_item}.${s.active}`)
      || listRef.current?.querySelector(`.${s.order_item}.${s.active}`);
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

  const handleDragStart = useCallback((e: React.DragEvent, index: number) => {
    dragRef.current = { from: index };
    e.dataTransfer.effectAllowed = "move";
    (e.target as HTMLElement).style.opacity = "0.4";
  }, []);

  const handleDragEnd = useCallback((e: React.DragEvent) => {
    (e.target as HTMLElement).style.opacity = "1";
    dragRef.current = null;
    setDragOver(null);
  }, []);

  const handleDragOver = useCallback((e: React.DragEvent, index: number) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
    setDragOver(index);
  }, []);

  const handleDrop = useCallback((e: React.DragEvent, toIndex: number) => {
    e.preventDefault();
    setDragOver(null);
    if (!dragRef.current) return;
    const fromIndex = dragRef.current.from;
    dragRef.current = null;
    if (fromIndex === toIndex) return;
    apiClient.reorderDisplay(fromIndex, toIndex);
  }, []);

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
    apiClient.streamControl(action).then(() => {
      // 상태 폴링에서 자동 갱신됨
    });
  }, [streamStatus.active]);

  // 키보드 ← → 제어 (디바운스)
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
    <div className={s.display_control_panel}>
      <div className={s.dcp_header}>
        <div className={s.dcp_title_row}>
          <span className={`${s.obs_dot} ${obsStatus.connected ? s.connected : s.disconnected}`} />
          <span className={s.dcp_title}>
            OBS {obsStatus.connected ? obsStatus.currentScene : "Disconnected"}
          </span>
          {streamStatus.active && (
            <span className={s.live_badge}>LIVE</span>
          )}
        </div>
        <div className={s.dcp_header_actions}>
          <button
            className={`${s.stream_btn} ${streamStatus.active ? s.stop : s.start}`}
            onClick={handleStreamToggle}
          >
            {streamStatus.active ? "방송 종료" : "방송 시작"}
          </button>
          <button className={s.dcp_close} onClick={onClose}>✕</button>
        </div>
      </div>

      {schedCountdown && (
        <div className={s.dcp_schedule_countdown}>
          <span className={s.sched_label}>{schedCountdown.label}</span>
          <span className={s.sched_time}>
            {String(schedCountdown.minutes).padStart(2, "0")}:{String(schedCountdown.seconds).padStart(2, "0")}
          </span>
        </div>
      )}

      <div className={s.dcp_nav}>
        <button onClick={() => handleNav("prev")}>◀</button>
        <span className={s.dcp_pos}>
          {items.length > 0 ? `${idx + 1} / ${items.length}` : "-"}
        </span>
        <button onClick={() => handleNav("next")}>▶</button>
      </div>

      {/* 타이머 제어 */}
      <div className={s.dcp_timer}>
        <div className={s.dcp_timer_row}>
          <button
            className={`${s.dcp_timer_btn} ${timer.enabled ? s.active : ""}`}
            onClick={handleTimerToggle}
            title="자동 넘김 ON/OFF"
          >
            {timer.enabled ? "⏸" : "▶"} 자동
          </button>
          <button className={s.dcp_timer_btn} onClick={handleTimerRepeat} title="현재 슬라이드 반복">
            반복
          </button>
          <button className={s.dcp_timer_btn} onClick={handleTimerRestart} title="처음으로">
            처음
          </button>
        </div>
        {timer.enabled && timer.countdown > 0 && (
          <div className={s.dcp_countdown}>
            다음까지 {timer.countdown}초
          </div>
        )}
        <div className={s.dcp_speed}>
          <span>속도</span>
          <input
            type="range"
            min="0.5"
            max="1.5"
            step="0.1"
            value={timer.speedFactor}
            onChange={handleSpeedChange}
          />
          <span>{Math.round(timer.speedFactor * 100)}%</span>
        </div>
      </div>

      {items.length > 0 && items[idx] && (
        <div className={s.dcp_current}>
          {items[idx].title}
          {items[idx].obj && items[idx].obj !== "-" && (
            <span className={s.dcp_current_obj}> — {items[idx].obj}</span>
          )}
        </div>
      )}

      <div className={s.dcp_order_list} ref={listRef}>
        {loadingMsg && (
          <div className={s.dcp_loading}>
            <div className={s.dcp_loading_spinner} />
            <span>{loadingMsg}</span>
          </div>
        )}
        {items.map((item, i) => {
          const hasSections = item.sections && item.sections.length > 0;
          const isExpanded = expandedItems.has(i);

          return (
            <div key={item.key || i}>
              <div
                className={`${s.order_item} ${i === idx ? s.active : ""}${dragOver === i ? ` ${s.drag_over}` : ""}`}
                onClick={() => handleJump(i)}
                draggable
                onDragStart={(e) => handleDragStart(e, i)}
                onDragEnd={handleDragEnd}
                onDragOver={(e) => handleDragOver(e, i)}
                onDrop={(e) => handleDrop(e, i)}
              >
                <span className={s.order_num}>{i + 1}</span>
                <span className={s.order_title_text}>{item.title}</span>
                <span className={s.order_obj}>
                  {item.obj && item.obj !== "-" ? item.obj : ""}
                </span>
                {hasSections && (
                  <button
                    className={s.order_toggle}
                    onClick={(e) => {
                      e.stopPropagation();
                      toggleItemExpand(i);
                    }}
                  >
                    {isExpanded ? "▲" : "▼"}
                  </button>
                )}
                <button
                  className={s.order_remove}
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
                <div className={s.order_sections}>
                  {item.sections!.map((sec, si) => {
                    const isActive = i === idx && sec.startPage === subPageIdx;
                    return (
                      <div
                        key={si}
                        className={`${s.order_section_item}${isActive ? ` ${s.active}` : ""}`}
                        onClick={() => handleSectionJump(i, sec.startPage)}
                      >
                        <span className={s.order_section_label}>{sec.label}</span>
                        <pre className={s.order_section_text}>{sec.text || ""}</pre>
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

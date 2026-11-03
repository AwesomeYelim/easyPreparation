"use client";

/**
 * useDisplaySession — Display 진행 상태와 조작을 한곳에서.
 * PC 순서 패널(ProSequencePanel, 메인·분리 창)과 폰 리모컨(MobileRemote)이 함께 쓴다.
 *
 * - 상태: 순서·현재 위치·서브페이지(총 개수 포함)·자동 타이머·OBS/방송 상태·예약 카운트다운·준비 중 메시지
 * - 갱신: 서버 WS(order/position/timer_state/navigate/display_loading/schedule_*) + display/status 폴링
 *         (페이지가 숨겨져 있으면 폴링을 쉰다 — 미리 띄워 둔 분리 창)
 * - 조작: 이전/다음, 항목·섹션 점프, 타이머, 방송 시작/종료, 순서 바꾸기(낙관적 반영 + 서버 브로드캐스트 잠시 무시)
 *
 * 순서·위치는 Recoil(displayItemsState 등)에 두어 같은 창의 타임라인·인스펙터와 공유한다.
 */

import { useCallback, useEffect, useRef, useState } from "react";
import { useRecoilState } from "recoil";
import { displayItemsState, displayPositionState, displaySubPageState } from "@/recoilState";
import { apiClient } from "@/lib/apiClient";
import { useWS } from "@/components/WebSocketProvider";
import { usePageVisible } from "@/lib/usePageVisible";
import { OBSStatus, StreamStatus, WorshipOrderItem } from "@/types";

export type ScheduleCountdown = { label: string; minutes: number; seconds: number } | null;

type Options = {
  /** display/status 폴링 주기 (ms) */
  pollMs?: number;
  /** 서버가 새 순서를 보냈을 때 (펼침 상태 초기화 등) */
  onOrder?: () => void;
  /** WS 로 현재 항목이 바뀌었을 때 (폰: 화면 가운데로 스크롤) */
  onIdxChange?: (idx: number) => void;
};

const EMPTY_STREAM: StreamStatus = { active: false, reconnecting: false, timecode: "", bytesSent: 0 };

export function useDisplaySession({ pollMs = 5000, onOrder, onIdxChange }: Options = {}) {
  const { subscribe, isOpen } = useWS();
  const pageVisible = usePageVisible();

  const [items, setItems] = useRecoilState(displayItemsState);
  const [idx, setIdx] = useRecoilState(displayPositionState);
  const [subIdx, setSubIdx] = useRecoilState(displaySubPageState);
  const [subTotal, setSubTotal] = useState(0);
  const [timerOn, setTimerOn] = useState(false);
  const [obsStatus, setObsStatus] = useState<OBSStatus>({ connected: false, currentScene: "" });
  const [streamStatus, setStreamStatus] = useState<StreamStatus>(EMPTY_STREAM);
  const [streamBusy, setStreamBusy] = useState(false);
  const [schedCountdown, setSchedCountdown] = useState<ScheduleCountdown>(null);
  const [loadingMsg, setLoadingMsg] = useState("");

  const idxRef = useRef(idx);
  idxRef.current = idx;
  const itemsLenRef = useRef(items.length);
  itemsLenRef.current = items.length;
  const streamBusyRef = useRef(false);
  streamBusyRef.current = streamBusy;
  const reorderLockRef = useRef(false);
  const reorderSuppressRef = useRef(false);
  const schedTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const keyCounterRef = useRef(0);
  // 콜백은 ref 로 — 호출하는 쪽이 매 렌더 새 함수를 넘겨도 WS 구독을 다시 걸지 않는다
  const onOrderRef = useRef(onOrder);
  onOrderRef.current = onOrder;
  const onIdxChangeRef = useRef(onIdxChange);
  onIdxChangeRef.current = onIdxChange;

  // 서버 순서에 key 가 없거나 중복이면 임시 key 부여 (React key·드래그 안정성)
  const ensureUniqueKeys = useCallback((raw: WorshipOrderItem[]) => {
    const seen = new Set<string>();
    return raw.map((item) => {
      const k = item.key;
      if (!k || seen.has(k)) return { ...item, key: `_auto_${keyCounterRef.current++}` };
      seen.add(k);
      return item;
    });
  }, []);

  const moveIdx = useCallback((next: number) => {
    if (next !== idxRef.current) onIdxChangeRef.current?.(next);
    setIdx(next);
  }, [setIdx]);

  // ── WS ──
  useEffect(() => {
    return subscribe((msg) => {
      switch (msg.type) {
        case "position":
          if (typeof msg.idx === "number") {
            if (msg.idx !== idxRef.current) setSubTotal(0);
            moveIdx(msg.idx);
            if (typeof msg.subPageIdx === "number") setSubIdx(msg.subPageIdx);
            if (typeof msg.subPageTotal === "number") setSubTotal(msg.subPageTotal);
          }
          break;
        case "navigate":
          if (msg.direction === "jump" && typeof msg.idx === "number") {
            moveIdx(msg.idx);
            setSubIdx(0);
            setSubTotal(0);
          }
          break;
        case "timer_state":
          if (typeof msg.idx === "number") moveIdx(msg.idx);
          if (typeof msg.subPageIdx === "number") setSubIdx(msg.subPageIdx);
          if (typeof msg.enabled === "boolean") setTimerOn(msg.enabled);
          break;
        case "order":
          if (!Array.isArray(msg.items)) break;
          setLoadingMsg("");
          if (reorderSuppressRef.current) break; // 방금 바꾼 순서를 서버의 이전 상태가 덮어쓰지 않게
          setItems(ensureUniqueKeys(msg.items as WorshipOrderItem[]));
          onOrderRef.current?.();
          if (typeof msg.idx === "number") {
            moveIdx(msg.idx);
            // append/remove처럼 재생 위치가 유지되는 브로드캐스트는 subPageIdx를 함께 보냄 — 없을 때만 리셋
            setSubIdx(typeof msg.subPageIdx === "number" ? msg.subPageIdx : 0);
          } else {
            moveIdx(0);
            setSubIdx(0);
          }
          break;
        case "display_loading":
          setLoadingMsg(msg.done ? "" : msg.message || "준비 중...");
          break;
        case "schedule_countdown":
          setSchedCountdown({ label: msg.label, minutes: msg.minutes, seconds: msg.seconds });
          if (schedTimeoutRef.current) clearTimeout(schedTimeoutRef.current);
          schedTimeoutRef.current = setTimeout(() => setSchedCountdown(null), 3000);
          break;
        case "schedule_started":
          setSchedCountdown(null);
          if (schedTimeoutRef.current) clearTimeout(schedTimeoutRef.current);
          break;
      }
    });
  }, [subscribe, setItems, setSubIdx, ensureUniqueKeys, moveIdx]);

  // ── 폴링 (보일 때만) ──
  useEffect(() => {
    if (!pageVisible) return;
    const poll = () =>
      apiClient.getDisplayStatus()
        .then((d: any) => {
          if (d.obs) setObsStatus(d.obs);
          if (d.stream && !streamBusyRef.current) setStreamStatus(d.stream);
          if (typeof d.timerEnabled === "boolean") setTimerOn(d.timerEnabled);
          if (Array.isArray(d.items) && d.items.length > 0 && !reorderSuppressRef.current) {
            setItems(ensureUniqueKeys(d.items as WorshipOrderItem[]));
            if (typeof d.idx === "number") setIdx(d.idx);
            // 새로고침 시 subPageIdx가 0으로 유실되던 문제 — 서버 상태를 그대로 복원
            if (typeof d.subPageIdx === "number") setSubIdx(d.subPageIdx);
          }
        })
        .catch((e: unknown) => console.error("display/status poll 에러:", e));
    poll();
    const t = setInterval(poll, pollMs);
    return () => clearInterval(t);
  }, [pageVisible, pollMs, setItems, setIdx, setSubIdx, ensureUniqueKeys]);

  // ── 조작 ──
  const navigate = useCallback((dir: "prev" | "next") => apiClient.navigateDisplay(dir), []);

  const jump = useCallback((index: number, subPage?: number) => {
    setIdx(index);
    if (subPage !== undefined) setSubIdx(subPage);
    return apiClient.jumpDisplay(index, subPage);
  }, [setIdx, setSubIdx]);

  const toggleTimer = useCallback(() => {
    setTimerOn((v) => !v); // WS 도착 전 즉시 반영
    return apiClient.timerControl("toggle");
  }, []);

  /** 방송 시작/종료. 결과를 돌려주므로 화면별로 토스트 등을 띄운다. */
  const setStream = useCallback(
    async (start: boolean, opts?: { includeThumbnail?: boolean; includeTitle?: boolean; isTest?: boolean }) => {
      setStreamBusy(true);
      try {
        const data: any = await apiClient.streamControl(start ? "start" : "stop", opts?.includeThumbnail, opts?.isTest, opts?.includeTitle);
        const ok = start ? !!(data?.ok || data?.active) : true;
        if (ok) setStreamStatus((s) => ({ ...s, active: start }));
        return { ok, error: data?.error as string | undefined };
      } finally {
        setStreamBusy(false);
      }
    },
    [],
  );

  /** 순서 바꾸기 — 화면에 먼저 반영하고, 서버 응답 직후 잠시 들어오는 이전 순서 브로드캐스트는 무시 */
  const reorder = useCallback((from: number, to: number) => {
    if (from === to || reorderLockRef.current) return Promise.resolve();
    reorderLockRef.current = true;
    reorderSuppressRef.current = true;
    setItems((prev) => {
      const next = [...prev];
      const [moved] = next.splice(from, 1);
      next.splice(to, 0, moved);
      return next;
    });
    setIdx((prev) => {
      if (prev === from) return to;
      let n = prev;
      if (from < prev) n--;
      if (to <= n) n++;
      return Math.max(0, Math.min(n, itemsLenRef.current - 1));
    });
    return apiClient.reorderDisplay(from, to).finally(() => {
      reorderLockRef.current = false;
      setTimeout(() => { reorderSuppressRef.current = false; }, 1500);
    });
  }, [setItems, setIdx]);

  return {
    connected: isOpen,
    pageVisible,
    items, setItems, idx, setIdx, subIdx, setSubIdx, subTotal,
    timerOn, obsStatus, streamStatus, streamBusy, schedCountdown, loadingMsg,
    navigate, jump, toggleTimer, setStream, reorder,
  };
}

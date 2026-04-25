"use client";

import { useEffect, useRef } from "react";
import toast from "react-hot-toast";
import { useWS } from "@/components/WebSocketProvider";
import { translateError, isImportantSuccess, PROGRESS_ERROR, PROGRESS_DONE } from "@/lib/errorMessages";

/**
 * WebSocket progress/done 메시지를 받아 Toast로 표시하는 브릿지 컴포넌트.
 * - code === -1 (에러): toast.error (중복 방지: 같은 메시지 2초 내 재발생 시 skip)
 * - code === 1 (완료): 중요 완료 메시지만 toast.success
 * - done 타입: toast.success
 * - code === 0 (진행중): 무시 (진행바가 있음)
 */
export default function WebSocketToastBridge() {
  const { subscribe } = useWS();
  // 중복 방지: 최근 표시한 메시지 key → 타임스탬프
  const recentToasts = useRef<Map<string, number>>(new Map());

  useEffect(() => {
    const DEDUP_MS = 2000; // 2초 내 같은 메시지 중복 방지

    const isDuplicate = (key: string): boolean => {
      const last = recentToasts.current.get(key);
      const now = Date.now();
      if (last && now - last < DEDUP_MS) return true;
      recentToasts.current.set(key, now);
      // 오래된 항목 정리 (메모리 누수 방지)
      if (recentToasts.current.size > 50) {
        const cutoff = now - 10000;
        recentToasts.current.forEach((ts, k) => {
          if (ts < cutoff) recentToasts.current.delete(k);
        });
      }
      return false;
    };

    const unsubscribe = subscribe((msg) => {
      if (msg.type === "progress") {
        if (msg.code === PROGRESS_ERROR) {
          const translated = translateError(msg.message || "알 수 없는 오류가 발생했습니다");
          const key = `error:${translated}`;
          if (!isDuplicate(key)) {
            toast.error(translated, { id: key, duration: 5000 });
          }
        } else if (msg.code === PROGRESS_DONE) {
          const message = msg.message || "";
          if (isImportantSuccess(message)) {
            const key = `success:${message}`;
            if (!isDuplicate(key)) {
              toast.success(message, { id: key, duration: 3000 });
            }
          }
        }
        // code === 0 (진행중): 무시
      } else if (msg.type === "done") {
        // 작업 완료 알림 (submitBulletin 등)
        const target = msg.target || "";
        const fileName = msg.fileName || "";
        const label = fileName
          ? `${fileName} 생성 완료`
          : target
          ? `${target} 완료`
          : "작업 완료";
        const key = `done:${target}:${fileName}`;
        if (!isDuplicate(key)) {
          toast.success(label, { id: key, duration: 3000 });
        }
      }
    });

    return unsubscribe;
  }, [subscribe]);

  return null;
}

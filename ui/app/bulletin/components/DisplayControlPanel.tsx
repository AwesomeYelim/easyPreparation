"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { apiClient } from "@/lib/apiClient";
import { WorshipOrderItem, OBSStatus } from "@/types";
import { useWS } from "@/components/WebSocketProvider";

type Props = {
  items: WorshipOrderItem[];
  onClose: () => void;
};

export default function DisplayControlPanel({ items, onClose }: Props) {
  const [idx, setIdx] = useState(0);
  const [obsStatus, setObsStatus] = useState<OBSStatus>({ connected: false, currentScene: "" });
  const { subscribe } = useWS();
  const listRef = useRef<HTMLDivElement>(null);

  // WS position 메시지 → idx 동기화 (display HTML이 보고)
  useEffect(() => {
    return subscribe((msg) => {
      if (msg.type === "position" && typeof msg.idx === "number") {
        setIdx(msg.idx);
      }
    });
  }, [subscribe]);

  // OBS 상태 폴링 (5초) — idx는 WS position으로만 동기화
  useEffect(() => {
    const poll = () => {
      apiClient.getDisplayStatus()
        .then((data: any) => {
          if (data.obs) setObsStatus(data.obs);
        })
        .catch(() => {});
    };
    poll();
    const timer = setInterval(poll, 5000);
    return () => clearInterval(timer);
  }, []);

  // 활성 항목 자동 스크롤
  useEffect(() => {
    const el = listRef.current?.querySelector(".order_item.active");
    el?.scrollIntoView({ block: "nearest", behavior: "smooth" });
  }, [idx]);

  const handleNav = useCallback((dir: "prev" | "next") => {
    apiClient.navigateDisplay(dir);
  }, []);

  const handleJump = useCallback((index: number) => {
    setIdx(index);
    apiClient.jumpDisplay(index);
  }, []);

  // 키보드 ← → 제어
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
      if (e.key === "ArrowRight") handleNav("next");
      else if (e.key === "ArrowLeft") handleNav("prev");
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [handleNav]);

  return (
    <div className="display_control_panel">
      <div className="dcp_header">
        <div className="dcp_title_row">
          <span className={`obs_dot ${obsStatus.connected ? "connected" : "disconnected"}`} />
          <span className="dcp_title">
            OBS {obsStatus.connected ? obsStatus.currentScene : "Disconnected"}
          </span>
        </div>
        <button className="dcp_close" onClick={onClose}>✕</button>
      </div>

      <div className="dcp_nav">
        <button onClick={() => handleNav("prev")}>◀</button>
        <span className="dcp_pos">
          {items.length > 0 ? `${idx + 1} / ${items.length}` : "-"}
        </span>
        <button onClick={() => handleNav("next")}>▶</button>
      </div>

      {items.length > 0 && items[idx] && (
        <div className="dcp_current">
          {items[idx].title}
          {items[idx].obj && items[idx].obj !== "-" && (
            <span className="dcp_current_obj"> — {items[idx].obj}</span>
          )}
        </div>
      )}

      <div className="dcp_order_list" ref={listRef}>
        {items.map((item, i) => (
          <div
            key={item.key || i}
            className={`order_item ${i === idx ? "active" : ""}`}
            onClick={() => handleJump(i)}
          >
            <span className="order_num">{i + 1}</span>
            <span className="order_title_text">{item.title}</span>
            <span className="order_obj">
              {item.obj && item.obj !== "-" ? item.obj : ""}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

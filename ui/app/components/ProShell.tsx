"use client";
import { useEffect, useState } from "react";
import { useRecoilState } from "recoil";
import { sequencePanelOpenState, inspectorOpenState, panelDetachedState, panelNativeState } from "@/recoilState";
import { useSetRecoilState } from "recoil";
import { useWS } from "@/components/WebSocketProvider";
import { panelApi } from "@/lib/apiClient";

export default function ProShell({ children }: { children: React.ReactNode }) {
  const [seqOpen, setSeqOpen] = useRecoilState(sequencePanelOpenState);
  const [inspOpen, setInspOpen] = useRecoilState(inspectorOpenState);
  const [detached, setDetached] = useRecoilState(panelDetachedState);
  const setNative = useSetRecoilState(panelNativeState);
  const [isMobile, setIsMobile] = useState(false);
  // 분리 창을 이 창의 패널 자리로 끌고 오는 중 → 붙을 자리 강조 (OBS 독처럼)
  const [dockHint, setDockHint] = useState(false);
  const { subscribe } = useWS();

  useEffect(() => {
    const close = () => {
      const mobile = window.innerWidth < 768;
      if (mobile) {
        setInspOpen(false);
        setSeqOpen(false);
      }
      setIsMobile(mobile);
    };
    close();
    window.addEventListener("resize", close);
    return () => window.removeEventListener("resize", close);
  }, [setInspOpen, setSeqOpen]);

  // 순서 패널 분리 상태: 서버 저장값으로 시작 → WS panel_dock 으로 갱신
  useEffect(() => {
    panelApi.getState().then((st) => {
      setNative(!!st.native);
      // 브라우저 모드에서는 (데스크톱 앱이 남긴) 분리 상태를 무시하고 항상 붙여서 보여 준다
      setDetached(!!st.native && !!st.detached);
    }).catch(() => {});
    return subscribe((msg) => {
      if (msg.type === "panel_dock" && typeof msg.detached === "boolean") {
        setDetached(msg.detached);
        setDockHint(false);
      }
      if (msg.type === "panel_dock_hint") setDockHint(!!msg.active);
    });
  }, [subscribe, setDetached, setNative]);

  const showSeq = seqOpen && (!detached || dockHint);

  return (
    <div
      className="pro-shell bg-pro-bg text-pro-text"
      style={{
        display: "grid",
        gridTemplateColumns: `48px ${showSeq ? "280px" : "0px"} 1fr ${inspOpen ? "280px" : "0px"}`,
        gridTemplateRows: `44px 1fr ${isMobile ? "0px" : "90px"}`,
        height: "100vh",
        overflow: "hidden",
        transition: "grid-template-columns 140ms ease-out",
      }}
    >
      {children}
      {detached && dockHint && (
        <div
          data-testid="seq-dock-hint"
          className="m-1.5 rounded-md border-2 border-dashed border-pro-accent bg-pro-accent/10 pointer-events-none"
          style={{ gridColumn: "2", gridRow: "2" }}
        />
      )}
    </div>
  );
}

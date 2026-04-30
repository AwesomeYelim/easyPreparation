"use client";
import { useEffect, useState } from "react";
import { useRecoilState } from "recoil";
import { sequencePanelOpenState, inspectorOpenState } from "@/recoilState";

export default function ProShell({ children }: { children: React.ReactNode }) {
  const [seqOpen, setSeqOpen] = useRecoilState(sequencePanelOpenState);
  const [inspOpen, setInspOpen] = useRecoilState(inspectorOpenState);
  const [isMobile, setIsMobile] = useState(false);

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

  return (
    <div
      className="pro-shell bg-pro-bg text-pro-text"
      style={{
        display: "grid",
        gridTemplateColumns: `48px ${seqOpen ? "280px" : "0px"} 1fr ${inspOpen ? "280px" : "0px"}`,
        gridTemplateRows: `44px 1fr ${isMobile ? "0px" : "90px"}`,
        height: "100vh",
        overflow: "hidden",
      }}
    >
      {children}
    </div>
  );
}

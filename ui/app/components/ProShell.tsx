"use client";
import { useRecoilValue } from "recoil";
import { sequencePanelOpenState, inspectorOpenState } from "@/recoilState";

export default function ProShell({ children }: { children: React.ReactNode }) {
  const seqOpen = useRecoilValue(sequencePanelOpenState);
  const inspOpen = useRecoilValue(inspectorOpenState);

  return (
    <div
      className="pro-shell bg-pro-bg text-pro-text"
      style={{
        display: "grid",
        gridTemplateColumns: `48px ${seqOpen ? "280px" : "0px"} 1fr ${inspOpen ? "280px" : "0px"}`,
        gridTemplateRows: "44px 1fr 90px",
        height: "100vh",
        overflow: "hidden",
      }}
    >
      {children}
    </div>
  );
}

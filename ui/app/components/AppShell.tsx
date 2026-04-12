"use client";

import { useRecoilValue } from "recoil";
import { displayPanelOpenState } from "@/recoilState";
import LeftSidebar from "./LeftSidebar";
import TopHeader from "./TopHeader";

export default function AppShell({ children }: { children: React.ReactNode }) {
  const panelOpen = useRecoilValue(displayPanelOpenState);

  return (
    <div className="min-h-screen min-w-[1024px] bg-surface">
      <LeftSidebar />
      <div
        className="transition-all duration-300 ease-in-out"
        style={{
          marginLeft: "256px", // w-64
          marginRight: panelOpen ? "320px" : "0px",
        }}
      >
        <TopHeader />
        <main className="p-6 lg:p-8">{children}</main>
      </div>
    </div>
  );
}

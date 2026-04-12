"use client";

import { useState } from "react";
import LyricsManager from "@/lyrics/components/LyricsManager";
import HymnSearch from "@/lyrics/components/HymnSearch";

export default function Lyrics() {
  const [tab, setTab] = useState<"free" | "hymnal">("free");

  return (
    <div className="w-full flex flex-col">
      <h1 className="text-3xl font-black tracking-tight text-primary mb-6">Find Your Worship</h1>

      {/* 탭 */}
      <div className="flex gap-1 mb-0">
        <button
          className={`px-5 py-2.5 text-sm font-bold rounded-t-lg border border-b-0 transition-all ${
            tab === "free"
              ? "bg-white text-secondary border-outline/40 shadow-sm"
              : "bg-surface-low text-on-surface-variant border-transparent hover:text-secondary hover:bg-white/60"
          }`}
          onClick={() => setTab("free")}
        >
          자유 곡
        </button>
        <button
          className={`px-5 py-2.5 text-sm font-bold rounded-t-lg border border-b-0 transition-all ${
            tab === "hymnal"
              ? "bg-white text-secondary border-outline/40 shadow-sm"
              : "bg-surface-low text-on-surface-variant border-transparent hover:text-secondary hover:bg-white/60"
          }`}
          onClick={() => setTab("hymnal")}
        >
          찬송가 검색
        </button>
      </div>

      <div className="border border-outline/40 rounded-b-xl rounded-tr-xl bg-white shadow-sm min-h-[500px]">
        {tab === "free" ? <LyricsManager /> : <HymnSearch />}
      </div>
    </div>
  );
}

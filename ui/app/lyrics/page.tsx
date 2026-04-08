"use client";

import { useState } from "react";
import LyricsManager from "@/lyrics/components/LyricsManager";
import HymnSearch from "@/lyrics/components/HymnSearch";

export default function Lyrics() {
  const [tab, setTab] = useState<"free" | "hymnal">("free");

  return (
    <div className="w-full flex flex-col">
      {/* 히어로 섹션 */}
      <section className="pb-6 border-b border-outline/40">
        <span className="text-secondary font-semibold tracking-widest text-[10px] uppercase block mb-1">
          Lyrics Retrieval Engine
        </span>
        <h2 className="text-4xl font-black text-primary tracking-tight">
          Find Your Worship
        </h2>
      </section>

      {/* 탭 */}
      <div className="flex gap-1 mt-6 mb-0">
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

"use client";

import { useState } from "react";
import LyricsManager from "@/lyrics/components/LyricsManager";
import HymnSearch from "@/lyrics/components/HymnSearch";
import CustomSongManager from "@/lyrics/components/CustomSongManager";

type Tab = "free" | "hymnal" | "my_songs";

export default function Lyrics() {
  const [tab, setTab] = useState<Tab>("free");

  const tabs: { key: Tab; label: string }[] = [
    { key: "free",     label: "자유 곡" },
    { key: "hymnal",   label: "찬송가 검색" },
    { key: "my_songs", label: "내 찬양" },
  ];

  return (
    <div className="w-full flex flex-col">
      <h1 className="text-xl font-black tracking-tight text-pro-text mb-4">Find Your Worship</h1>

      {/* 탭 */}
      <div className="flex gap-1 mb-0">
        {tabs.map(({ key, label }) => (
          <button
            key={key}
            className={`px-5 py-2.5 text-sm font-bold rounded-t-lg border border-b-0 transition-all ${
              tab === key
                ? "bg-pro-elevated text-electric-blue border-pro-border"
                : "bg-pro-surface text-pro-text-muted border-transparent hover:text-electric-blue hover:bg-pro-hover"
            }`}
            onClick={() => setTab(key)}
          >
            {label}
          </button>
        ))}
      </div>

      <div className="border border-pro-border rounded-b-xl rounded-tr-xl bg-pro-elevated min-h-[500px]">
        {tab === "free" ? (
          <LyricsManager />
        ) : tab === "hymnal" ? (
          <HymnSearch />
        ) : (
          <CustomSongManager />
        )}
      </div>
    </div>
  );
}

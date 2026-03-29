"use client";

import { useState } from "react";
import LyricsManager from "@/lyrics/components/LyricsManager";
import HymnSearch from "@/lyrics/components/HymnSearch";
import "./lyrics-tabs.scss";

export default function Lyrics() {
  const [tab, setTab] = useState<"free" | "hymnal">("free");

  return (
    <div className="lyrics_page">
      <div className="lyrics_tabs">
        <button
          className={`lyrics_tab${tab === "free" ? " active" : ""}`}
          onClick={() => setTab("free")}
        >
          자유 곡
        </button>
        <button
          className={`lyrics_tab${tab === "hymnal" ? " active" : ""}`}
          onClick={() => setTab("hymnal")}
        >
          찬송가 검색
        </button>
      </div>
      {tab === "free" ? <LyricsManager /> : <HymnSearch />}
    </div>
  );
}
